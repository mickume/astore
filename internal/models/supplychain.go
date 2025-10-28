package models

import (
	"time"
)

// Signature represents a cryptographic signature for an artifact
type Signature struct {
	ID          string    `json:"id"`
	ArtifactID  string    `json:"artifactId"` // bucket/key
	Algorithm   string    `json:"algorithm"`  // e.g., "RSA", "ECDSA"
	Signature   []byte    `json:"signature"`  // The actual signature bytes
	PublicKey   string    `json:"publicKey"`  // PEM-encoded public key
	SignedBy    string    `json:"signedBy"`   // User/system that created the signature
	SignedAt    time.Time `json:"signedAt"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// SBOM represents a Software Bill of Materials
type SBOM struct {
	ID          string    `json:"id"`
	ArtifactID  string    `json:"artifactId"` // bucket/key
	Format      SBOMFormat `json:"format"`     // SPDX, CycloneDX
	Version     string    `json:"version"`    // Format version
	Content     []byte    `json:"content"`    // SBOM document content
	ContentType string    `json:"contentType"` // application/json, application/xml, etc.
	Hash        string    `json:"hash"`       // SHA256 hash of content
	CreatedBy   string    `json:"createdBy"`  // User/system that created the SBOM
	CreatedAt   time.Time `json:"createdAt"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// SBOMFormat defines supported SBOM formats
type SBOMFormat string

const (
	SBOMFormatSPDX      SBOMFormat = "spdx"
	SBOMFormatCycloneDX SBOMFormat = "cyclonedx"
)

// Attestation represents a build, test, or deployment attestation
type Attestation struct {
	ID            string           `json:"id"`
	ArtifactID    string           `json:"artifactId"` // bucket/key
	Type          AttestationType  `json:"type"`       // build, test, deploy, scan
	Predicate     map[string]interface{} `json:"predicate"` // Attestation-specific data
	PredicateType string           `json:"predicateType"` // e.g., "https://slsa.dev/provenance/v0.2"
	Signature     []byte           `json:"signature,omitempty"` // Optional signature
	CreatedBy     string           `json:"createdBy"`
	CreatedAt     time.Time        `json:"createdAt"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// AttestationType defines types of attestations
type AttestationType string

const (
	AttestationTypeBuild      AttestationType = "build"
	AttestationTypeTest       AttestationType = "test"
	AttestationTypeDeploy     AttestationType = "deploy"
	AttestationTypeScan       AttestationType = "scan"
	AttestationTypeProvenance AttestationType = "provenance"
)

// VerificationResult represents the result of signature verification
type VerificationResult struct {
	Verified    bool      `json:"verified"`
	SignatureID string    `json:"signatureId"`
	SignedBy    string    `json:"signedBy"`
	SignedAt    time.Time `json:"signedAt"`
	Error       string    `json:"error,omitempty"`
}
