package client_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/candlekeep/zot-artifact-store/pkg/client"
	"github.com/candlekeep/zot-artifact-store/test"
)

func TestSupplyChainOperations(t *testing.T) {
	t.Run("Sign artifact", func(t *testing.T) {
		// Given: Test server
		var receivedBody []byte

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			test.AssertEqual(t, "POST", r.Method, "HTTP method")
			test.AssertEqual(t, "/supplychain/sign/test-bucket/test-key", r.URL.Path, "request path")
			receivedBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "sig-1",
				"bucket": "test-bucket",
				"key": "test-key",
				"signature": "base64-signature",
				"publicKey": "base64-public-key",
				"algorithm": "RSA-SHA256",
				"signedAt": "2024-01-01T00:00:00Z",
				"artifactSha": "sha256-hash"
			}`))
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Signing an artifact
		ctx := context.Background()
		sig, err := c.SignArtifact(ctx, "test-bucket", "test-key", "private-key-pem")

		// Then: Signature is created
		test.AssertNoError(t, err, "sign artifact")
		test.AssertTrue(t, sig != nil, "signature should not be nil")
		test.AssertEqual(t, "sig-1", sig.ID, "signature ID")
		test.AssertEqual(t, "RSA-SHA256", sig.Algorithm, "algorithm")
		test.AssertTrue(t, len(receivedBody) > 0, "should send request body")
	})

	t.Run("Get signatures", func(t *testing.T) {
		// Given: Test server with signatures
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			test.AssertEqual(t, "GET", r.Method, "HTTP method")
			test.AssertEqual(t, "/supplychain/signatures/test-bucket/test-key", r.URL.Path, "request path")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"signatures": [
					{
						"id": "sig-1",
						"bucket": "test-bucket",
						"key": "test-key",
						"signature": "signature1",
						"publicKey": "public-key-1",
						"algorithm": "RSA-SHA256",
						"signedAt": "2024-01-01T00:00:00Z",
						"artifactSha": "sha256-hash"
					}
				]
			}`))
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Getting signatures
		ctx := context.Background()
		sigs, err := c.GetSignatures(ctx, "test-bucket", "test-key")

		// Then: Returns signature list
		test.AssertNoError(t, err, "get signatures")
		test.AssertTrue(t, len(sigs) == 1, "should have 1 signature")
		test.AssertEqual(t, "sig-1", sigs[0].ID, "signature ID")
	})

	t.Run("Verify signatures", func(t *testing.T) {
		// Given: Test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			test.AssertEqual(t, "POST", r.Method, "HTTP method")
			test.AssertEqual(t, "/supplychain/verify/test-bucket/test-key", r.URL.Path, "request path")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"valid": true,
				"signatures": ["sig-1"],
				"message": "All signatures valid",
				"verifiedAt": "2024-01-01T00:00:00Z"
			}`))
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Verifying signatures
		ctx := context.Background()
		publicKeys := []string{"public-key-1", "public-key-2"}
		result, err := c.VerifySignatures(ctx, "test-bucket", "test-key", publicKeys)

		// Then: Returns verification result
		test.AssertNoError(t, err, "verify signatures")
		test.AssertTrue(t, result != nil, "result should not be nil")
		test.AssertTrue(t, result.Valid, "should be valid")
		test.AssertTrue(t, len(result.Signatures) == 1, "should have 1 verified signature")
	})

	t.Run("Attach SBOM", func(t *testing.T) {
		// Given: Test server
		sbomContent := "SPDX-2.3 document content"
		var receivedBody []byte

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			test.AssertEqual(t, "POST", r.Method, "HTTP method")
			test.AssertEqual(t, "/supplychain/sbom/test-bucket/test-key", r.URL.Path, "request path")
			receivedBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "sbom-1",
				"bucket": "test-bucket",
				"key": "test-key",
				"format": "spdx",
				"content": "` + sbomContent + `",
				"createdAt": "2024-01-01T00:00:00Z"
			}`))
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Attaching SBOM
		ctx := context.Background()
		sbom, err := c.AttachSBOM(ctx, "test-bucket", "test-key", "spdx", sbomContent)

		// Then: SBOM is attached
		test.AssertNoError(t, err, "attach SBOM")
		test.AssertTrue(t, sbom != nil, "SBOM should not be nil")
		test.AssertEqual(t, "sbom-1", sbom.ID, "SBOM ID")
		test.AssertEqual(t, "spdx", sbom.Format, "SBOM format")
		test.AssertTrue(t, len(receivedBody) > 0, "should send request body")
	})

	t.Run("Get SBOM", func(t *testing.T) {
		// Given: Test server with SBOM
		sbomContent := "SPDX-2.3 document content"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			test.AssertEqual(t, "GET", r.Method, "HTTP method")
			test.AssertEqual(t, "/supplychain/sbom/test-bucket/test-key", r.URL.Path, "request path")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "sbom-1",
				"bucket": "test-bucket",
				"key": "test-key",
				"format": "spdx",
				"content": "` + sbomContent + `",
				"createdAt": "2024-01-01T00:00:00Z"
			}`))
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Getting SBOM
		ctx := context.Background()
		sbom, err := c.GetSBOM(ctx, "test-bucket", "test-key")

		// Then: Returns SBOM
		test.AssertNoError(t, err, "get SBOM")
		test.AssertTrue(t, sbom != nil, "SBOM should not be nil")
		test.AssertEqual(t, "sbom-1", sbom.ID, "SBOM ID")
		test.AssertEqual(t, "spdx", sbom.Format, "SBOM format")
	})

	t.Run("Add attestation", func(t *testing.T) {
		// Given: Test server
		var receivedBody []byte

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			test.AssertEqual(t, "POST", r.Method, "HTTP method")
			test.AssertEqual(t, "/supplychain/attestations/test-bucket/test-key", r.URL.Path, "request path")
			receivedBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "att-1",
				"bucket": "test-bucket",
				"key": "test-key",
				"type": "build",
				"data": {"buildId": "123", "status": "success"},
				"createdAt": "2024-01-01T00:00:00Z",
				"createdBy": "ci-system"
			}`))
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Adding attestation
		ctx := context.Background()
		attData := map[string]interface{}{
			"buildId": "123",
			"status":  "success",
		}
		att, err := c.AddAttestation(ctx, "test-bucket", "test-key", "build", attData)

		// Then: Attestation is added
		test.AssertNoError(t, err, "add attestation")
		test.AssertTrue(t, att != nil, "attestation should not be nil")
		test.AssertEqual(t, "att-1", att.ID, "attestation ID")
		test.AssertEqual(t, "build", att.Type, "attestation type")
		test.AssertTrue(t, bytes.Contains(receivedBody, []byte("build")), "should contain type in body")
	})

	t.Run("Get attestations", func(t *testing.T) {
		// Given: Test server with attestations
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			test.AssertEqual(t, "GET", r.Method, "HTTP method")
			test.AssertEqual(t, "/supplychain/attestations/test-bucket/test-key", r.URL.Path, "request path")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"attestations": [
					{
						"id": "att-1",
						"bucket": "test-bucket",
						"key": "test-key",
						"type": "build",
						"data": {"buildId": "123"},
						"createdAt": "2024-01-01T00:00:00Z"
					},
					{
						"id": "att-2",
						"bucket": "test-bucket",
						"key": "test-key",
						"type": "test",
						"data": {"passed": true},
						"createdAt": "2024-01-01T01:00:00Z"
					}
				]
			}`))
		}))
		defer server.Close()

		c, _ := client.NewClient(&client.Config{BaseURL: server.URL})

		// When: Getting attestations
		ctx := context.Background()
		atts, err := c.GetAttestations(ctx, "test-bucket", "test-key")

		// Then: Returns attestation list
		test.AssertNoError(t, err, "get attestations")
		test.AssertTrue(t, len(atts) == 2, "should have 2 attestations")
		test.AssertEqual(t, "att-1", atts[0].ID, "first attestation ID")
		test.AssertEqual(t, "build", atts[0].Type, "first attestation type")
		test.AssertEqual(t, "att-2", atts[1].ID, "second attestation ID")
		test.AssertEqual(t, "test", atts[1].Type, "second attestation type")
	})
}
