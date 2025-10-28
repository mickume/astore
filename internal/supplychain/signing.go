package supplychain

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/candlekeep/zot-artifact-store/internal/models"
	"github.com/google/uuid"
)

// Signer provides artifact signing functionality
type Signer struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewSigner creates a new signer with a private key
func NewSigner(privateKeyPEM string) (*Signer, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &Signer{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}, nil
}

// GenerateKeyPair generates a new RSA key pair
func GenerateKeyPair(bits int) (*Signer, string, string, error) {
	if bits == 0 {
		bits = 2048
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate key: %w", err)
	}

	// Encode private key to PEM
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Encode public key to PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	signer := &Signer{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}

	return signer, string(privateKeyPEM), string(publicKeyPEM), nil
}

// SignArtifact signs artifact data and returns a signature model
func (s *Signer) SignArtifact(artifactID string, data []byte, signedBy string) (*models.Signature, error) {
	// Calculate hash
	hash := sha256.Sum256(data)

	// Sign the hash
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	// Encode public key to PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(s.publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return &models.Signature{
		ID:         uuid.New().String(),
		ArtifactID: artifactID,
		Algorithm:  "RSA-SHA256",
		Signature:  signature,
		PublicKey:  string(publicKeyPEM),
		SignedBy:   signedBy,
		SignedAt:   time.Now(),
	}, nil
}

// VerifySignature verifies a signature against artifact data
func VerifySignature(signature *models.Signature, data []byte) (*models.VerificationResult, error) {
	// Parse public key
	block, _ := pem.Decode([]byte(signature.PublicKey))
	if block == nil {
		return &models.VerificationResult{
			Verified: false,
			Error:    "failed to decode public key PEM",
		}, nil
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return &models.VerificationResult{
			Verified: false,
			Error:    fmt.Sprintf("failed to parse public key: %v", err),
		}, nil
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return &models.VerificationResult{
			Verified: false,
			Error:    "public key is not RSA",
		}, nil
	}

	// Calculate hash of data
	hash := sha256.Sum256(data)

	// Verify signature
	err = rsa.VerifyPKCS1v15(rsaPublicKey, crypto.SHA256, hash[:], signature.Signature)
	if err != nil {
		return &models.VerificationResult{
			Verified:    false,
			SignatureID: signature.ID,
			SignedBy:    signature.SignedBy,
			SignedAt:    signature.SignedAt,
			Error:       fmt.Sprintf("signature verification failed: %v", err),
		}, nil
	}

	return &models.VerificationResult{
		Verified:    true,
		SignatureID: signature.ID,
		SignedBy:    signature.SignedBy,
		SignedAt:    signature.SignedAt,
	}, nil
}
