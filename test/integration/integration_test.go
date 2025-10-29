//go:build integration
// +build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/candlekeep/zot-artifact-store/internal/api/s3"
	"github.com/candlekeep/zot-artifact-store/internal/extensions/metrics"
	metricsPkg "github.com/candlekeep/zot-artifact-store/internal/metrics"
	"github.com/candlekeep/zot-artifact-store/internal/storage"
	"github.com/candlekeep/zot-artifact-store/test"
	"github.com/gorilla/mux"
	"zotregistry.io/zot/pkg/log"
)

// TestEndToEndWorkflow tests a complete artifact lifecycle
func TestEndToEndWorkflow(t *testing.T) {
	// Given: A fully configured artifact store
	tmpDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	server, baseURL := setupTestServer(t, tmpDir)
	defer server.Close()

	bucket := "releases"
	key := "myapp-1.0.0.tar.gz"
	content := []byte("test artifact content")

	t.Run("Complete artifact lifecycle", func(t *testing.T) {
		// Step 1: Create bucket
		resp := makeRequest(t, "PUT", baseURL+"/s3/"+bucket, nil)
		test.AssertEqual(t, http.StatusOK, resp.StatusCode, "create bucket status")

		// Step 2: Upload artifact
		resp = makeRequest(t, "PUT", baseURL+"/s3/"+bucket+"/"+key, bytes.NewReader(content))
		test.AssertEqual(t, http.StatusOK, resp.StatusCode, "upload status")

		// Step 3: Download artifact
		resp = makeRequest(t, "GET", baseURL+"/s3/"+bucket+"/"+key, nil)
		test.AssertEqual(t, http.StatusOK, resp.StatusCode, "download status")

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		test.AssertEqual(t, string(content), string(body), "downloaded content")

		// Step 4: List artifacts
		resp = makeRequest(t, "GET", baseURL+"/s3/"+bucket, nil)
		test.AssertEqual(t, http.StatusOK, resp.StatusCode, "list objects status")

		// Step 5: Check health
		resp = makeRequest(t, "GET", baseURL+"/health", nil)
		test.AssertEqual(t, http.StatusOK, resp.StatusCode, "health check status")

		// Step 6: Check metrics
		resp = makeRequest(t, "GET", baseURL+"/metrics", nil)
		test.AssertEqual(t, http.StatusOK, resp.StatusCode, "metrics status")

		// Step 7: Delete artifact
		resp = makeRequest(t, "DELETE", baseURL+"/s3/"+bucket+"/"+key, nil)
		test.AssertEqual(t, http.StatusNoContent, resp.StatusCode, "delete status")

		// Step 8: Verify deletion
		resp = makeRequest(t, "GET", baseURL+"/s3/"+bucket+"/"+key, nil)
		test.AssertEqual(t, http.StatusNotFound, resp.StatusCode, "artifact should be deleted")
	})
}

// TestMultipartUploadWorkflow tests multipart upload end-to-end
func TestMultipartUploadWorkflow(t *testing.T) {
	// Given: A test server
	tmpDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	server, baseURL := setupTestServer(t, tmpDir)
	defer server.Close()

	bucket := "large-files"
	key := "large-artifact.bin"

	t.Run("Multipart upload lifecycle", func(t *testing.T) {
		// Step 1: Create bucket
		resp := makeRequest(t, "PUT", baseURL+"/s3/"+bucket, nil)
		test.AssertEqual(t, http.StatusOK, resp.StatusCode, "create bucket")

		// Step 2: Initiate multipart upload
		resp = makeRequest(t, "POST", baseURL+"/s3/"+bucket+"/"+key+"?uploads", nil)
		test.AssertEqual(t, http.StatusOK, resp.StatusCode, "initiate upload")

		var initResp map[string]string
		json.NewDecoder(resp.Body).Decode(&initResp)
		resp.Body.Close()

		uploadID := initResp["uploadId"]
		test.AssertTrue(t, uploadID != "", "upload ID received")

		// Step 3: Upload parts (simulate)
		part1 := []byte("part1data")
		url := fmt.Sprintf("%s/s3/%s/%s?uploadId=%s&partNumber=1", baseURL, bucket, key, uploadID)
		resp = makeRequest(t, "PUT", url, bytes.NewReader(part1))
		test.AssertEqual(t, http.StatusOK, resp.StatusCode, "upload part 1")

		// Step 4: Complete multipart upload
		completeReq := map[string]interface{}{
			"parts": []map[string]interface{}{
				{"partNumber": 1, "etag": "etag1"},
			},
		}
		completeData, _ := json.Marshal(completeReq)
		url = fmt.Sprintf("%s/s3/%s/%s?uploadId=%s", baseURL, bucket, key, uploadID)
		resp = makeRequest(t, "POST", url, bytes.NewReader(completeData))
		test.AssertEqual(t, http.StatusOK, resp.StatusCode, "complete upload")

		// Step 5: Verify artifact exists
		resp = makeRequest(t, "GET", baseURL+"/s3/"+bucket+"/"+key, nil)
		test.AssertEqual(t, http.StatusOK, resp.StatusCode, "artifact exists")
	})
}


// TestHealthAndMetrics tests observability endpoints
func TestHealthAndMetrics(t *testing.T) {
	// Given: A test server
	tmpDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	server, baseURL := setupTestServer(t, tmpDir)
	defer server.Close()

	t.Run("Health endpoints", func(t *testing.T) {
		// Health check
		resp := makeRequest(t, "GET", baseURL+"/health", nil)
		test.AssertEqual(t, http.StatusOK, resp.StatusCode, "health")

		var health map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&health)
		resp.Body.Close()
		test.AssertEqual(t, "healthy", health["status"], "overall status")

		// Readiness
		resp = makeRequest(t, "GET", baseURL+"/health/ready", nil)
		test.AssertEqual(t, http.StatusOK, resp.StatusCode, "readiness")

		// Liveness
		resp = makeRequest(t, "GET", baseURL+"/health/live", nil)
		test.AssertEqual(t, http.StatusOK, resp.StatusCode, "liveness")
	})

	t.Run("Metrics endpoint", func(t *testing.T) {
		// Metrics
		resp := makeRequest(t, "GET", baseURL+"/metrics", nil)
		test.AssertEqual(t, http.StatusOK, resp.StatusCode, "metrics")

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		test.AssertTrue(t, len(body) > 0, "metrics content")
	})
}

// Helper functions

func setupTestEnvironment(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "integration-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func setupTestServer(t *testing.T, tmpDir string) (*httptest.Server, string) {
	logger := log.NewLogger("error", "")

	// Create metadata store
	metadataStore, err := storage.NewMetadataStore(tmpDir + "/metadata.db")
	if err != nil {
		t.Fatalf("failed to create metadata store: %v", err)
	}

	// Create router
	router := mux.NewRouter()

	// Initialize S3 API handler
	s3Handler := s3.NewHandler(metadataStore, tmpDir+"/artifacts", logger)
	s3Handler.RegisterRoutes(router)

	// Initialize metrics extension
	metricsExt := metrics.NewMetricsExtension()
	metricsCollector := metricsPkg.NewPrometheusCollector()
	healthChecker := metricsPkg.NewHealthChecker(metadataStore, logger, "test-version")
	metricsHandler := metrics.NewHandler(metricsCollector, healthChecker, nil, logger)

	config := &metrics.Config{
		Prometheus: metrics.PrometheusCfg{
			Enabled: true,
			Path:    "/metrics",
		},
		Health: metrics.HealthCfg{
			Enabled:       true,
			HealthPath:    "/health",
			ReadinessPath: "/health/ready",
			LivenessPath:  "/health/live",
		},
	}
	metricsHandler.RegisterRoutes(router, config)

	// Create test server
	server := httptest.NewServer(router)

	// Store extensions for potential cleanup
	_ = s3Ext
	_ = metricsExt

	return server, server.URL
}

func makeRequest(t *testing.T, method, url string, body io.Reader) *http.Response {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	return resp
}
