package supplychain_test

import (
	"testing"

	"github.com/candlekeep/zot-artifact-store/internal/supplychain"
	"github.com/candlekeep/zot-artifact-store/test"
)

func TestSigning(t *testing.T) {
	t.Run("Generate key pair", func(t *testing.T) {
		// Given: Key generation parameters
		bits := 2048

		// When: Generating a key pair
		signer, privateKey, publicKey, err := supplychain.GenerateKeyPair(bits)

		// Then: Key pair is generated successfully
		test.AssertNoError(t, err, "key generation")
		test.AssertTrue(t, signer != nil, "signer created")
		test.AssertTrue(t, len(privateKey) > 0, "private key generated")
		test.AssertTrue(t, len(publicKey) > 0, "public key generated")
	})

	t.Run("Sign and verify artifact", func(t *testing.T) {
		// Given: A signer and test data
		signer, _, _, err := supplychain.GenerateKeyPair(2048)
		test.AssertNoError(t, err, "key generation")

		artifactID := "bucket/test-artifact"
		data := []byte("test artifact content")
		signedBy := "test@example.com"

		// When: Signing the artifact
		signature, err := signer.SignArtifact(artifactID, data, signedBy)

		// Then: Signature is created
		test.AssertNoError(t, err, "signing")
		test.AssertEqual(t, artifactID, signature.ArtifactID, "artifact ID")
		test.AssertEqual(t, "RSA-SHA256", signature.Algorithm, "algorithm")
		test.AssertEqual(t, signedBy, signature.SignedBy, "signed by")
		test.AssertTrue(t, len(signature.Signature) > 0, "signature data")

		// When: Verifying the signature
		result, err := supplychain.VerifySignature(signature, data)

		// Then: Signature is verified
		test.AssertNoError(t, err, "verification")
		test.AssertTrue(t, result.Verified, "signature verified")
		test.AssertEqual(t, signature.ID, result.SignatureID, "signature ID")
	})

	t.Run("Verify fails with wrong data", func(t *testing.T) {
		// Given: A signature and different data
		signer, _, _, _ := supplychain.GenerateKeyPair(2048)
		data := []byte("original data")
		signature, _ := signer.SignArtifact("bucket/key", data, "user")

		wrongData := []byte("tampered data")

		// When: Verifying with wrong data
		result, err := supplychain.VerifySignature(signature, wrongData)

		// Then: Verification fails
		test.AssertNoError(t, err, "verification call")
		test.AssertFalse(t, result.Verified, "signature should not verify")
		test.AssertTrue(t, len(result.Error) > 0, "error message present")
	})

	t.Run("Verify fails with invalid public key", func(t *testing.T) {
		// Given: A signature with invalid public key
		signer, _, _, _ := supplychain.GenerateKeyPair(2048)
		data := []byte("test data")
		signature, _ := signer.SignArtifact("bucket/key", data, "user")

		// Tamper with the public key
		signature.PublicKey = "invalid-key"

		// When: Verifying the signature
		result, err := supplychain.VerifySignature(signature, data)

		// Then: Verification fails
		test.AssertNoError(t, err, "verification call")
		test.AssertFalse(t, result.Verified, "signature should not verify")
		test.AssertTrue(t, len(result.Error) > 0, "error message present")
	})
}
