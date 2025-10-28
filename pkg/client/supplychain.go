package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/candlekeep/zot-artifact-store/internal/errors"
)

// Signature represents a cryptographic signature on an artifact
type Signature struct {
	ID          string    `json:"id"`
	Bucket      string    `json:"bucket"`
	Key         string    `json:"key"`
	Signature   string    `json:"signature"`
	PublicKey   string    `json:"publicKey"`
	Algorithm   string    `json:"algorithm"`
	SignedAt    time.Time `json:"signedAt"`
	SignedBy    string    `json:"signedBy,omitempty"`
	ArtifactSHA string    `json:"artifactSha"`
}

// SBOM represents a Software Bill of Materials
type SBOM struct {
	ID        string    `json:"id"`
	Bucket    string    `json:"bucket"`
	Key       string    `json:"key"`
	Format    string    `json:"format"` // spdx or cyclonedx
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

// Attestation represents a build/test/scan attestation
type Attestation struct {
	ID        string                 `json:"id"`
	Bucket    string                 `json:"bucket"`
	Key       string                 `json:"key"`
	Type      string                 `json:"type"` // build, test, scan, deploy
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"createdAt"`
	CreatedBy string                 `json:"createdBy,omitempty"`
}

// VerificationResult represents the result of signature verification
type VerificationResult struct {
	Valid      bool      `json:"valid"`
	Signatures []string  `json:"signatures,omitempty"`
	Message    string    `json:"message,omitempty"`
	VerifiedAt time.Time `json:"verifiedAt"`
}

// SignArtifact signs an artifact with the provided private key
func (c *Client) SignArtifact(ctx context.Context, bucket, key, privateKey string) (*Signature, error) {
	body := map[string]string{
		"privateKey": privateKey,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, errors.NewInternal("failed to marshal request: " + err.Error())
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	urlPath := fmt.Sprintf("/supplychain/sign/%s/%s", bucket, key)

	resp, err := c.doRequest(ctx, "POST", urlPath, bytes.NewReader(bodyBytes), headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var signature Signature
	if err := json.NewDecoder(resp.Body).Decode(&signature); err != nil {
		return nil, errors.NewInternal("failed to parse response: " + err.Error())
	}

	return &signature, nil
}

// GetSignatures retrieves all signatures for an artifact
func (c *Client) GetSignatures(ctx context.Context, bucket, key string) ([]Signature, error) {
	urlPath := fmt.Sprintf("/supplychain/signatures/%s/%s", bucket, key)

	resp, err := c.doRequest(ctx, "GET", urlPath, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Signatures []Signature `json:"signatures"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.NewInternal("failed to parse response: " + err.Error())
	}

	return result.Signatures, nil
}

// VerifySignatures verifies all signatures on an artifact
func (c *Client) VerifySignatures(ctx context.Context, bucket, key string, publicKeys []string) (*VerificationResult, error) {
	body := map[string]interface{}{
		"publicKeys": publicKeys,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, errors.NewInternal("failed to marshal request: " + err.Error())
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	urlPath := fmt.Sprintf("/supplychain/verify/%s/%s", bucket, key)

	resp, err := c.doRequest(ctx, "POST", urlPath, bytes.NewReader(bodyBytes), headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result VerificationResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.NewInternal("failed to parse response: " + err.Error())
	}

	return &result, nil
}

// AttachSBOM attaches a Software Bill of Materials to an artifact
func (c *Client) AttachSBOM(ctx context.Context, bucket, key, format, content string) (*SBOM, error) {
	body := map[string]string{
		"format":  format,
		"content": content,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, errors.NewInternal("failed to marshal request: " + err.Error())
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	urlPath := fmt.Sprintf("/supplychain/sbom/%s/%s", bucket, key)

	resp, err := c.doRequest(ctx, "POST", urlPath, bytes.NewReader(bodyBytes), headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var sbom SBOM
	if err := json.NewDecoder(resp.Body).Decode(&sbom); err != nil {
		return nil, errors.NewInternal("failed to parse response: " + err.Error())
	}

	return &sbom, nil
}

// GetSBOM retrieves the SBOM for an artifact
func (c *Client) GetSBOM(ctx context.Context, bucket, key string) (*SBOM, error) {
	urlPath := fmt.Sprintf("/supplychain/sbom/%s/%s", bucket, key)

	resp, err := c.doRequest(ctx, "GET", urlPath, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var sbom SBOM
	if err := json.NewDecoder(resp.Body).Decode(&sbom); err != nil {
		return nil, errors.NewInternal("failed to parse response: " + err.Error())
	}

	return &sbom, nil
}

// AddAttestation adds an attestation to an artifact
func (c *Client) AddAttestation(ctx context.Context, bucket, key, attType string, data map[string]interface{}) (*Attestation, error) {
	body := map[string]interface{}{
		"type": attType,
		"data": data,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, errors.NewInternal("failed to marshal request: " + err.Error())
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	urlPath := fmt.Sprintf("/supplychain/attestations/%s/%s", bucket, key)

	resp, err := c.doRequest(ctx, "POST", urlPath, bytes.NewReader(bodyBytes), headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var attestation Attestation
	if err := json.NewDecoder(resp.Body).Decode(&attestation); err != nil {
		return nil, errors.NewInternal("failed to parse response: " + err.Error())
	}

	return &attestation, nil
}

// GetAttestations retrieves all attestations for an artifact
func (c *Client) GetAttestations(ctx context.Context, bucket, key string) ([]Attestation, error) {
	urlPath := fmt.Sprintf("/supplychain/attestations/%s/%s", bucket, key)

	resp, err := c.doRequest(ctx, "GET", urlPath, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Attestations []Attestation `json:"attestations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.NewInternal("failed to parse response: " + err.Error())
	}

	return result.Attestations, nil
}
