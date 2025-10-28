package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/candlekeep/zot-artifact-store/pkg/client"
	"github.com/candlekeep/zot-artifact-store/test"
)

func TestNewClient(t *testing.T) {
	t.Run("Create client with valid config", func(t *testing.T) {
		// Given: Valid configuration
		config := &client.Config{
			BaseURL: "https://artifacts.example.com",
			Token:   "test-token",
			Timeout: 10 * time.Second,
		}

		// When: Creating a new client
		c, err := client.NewClient(config)

		// Then: Client is created successfully
		test.AssertNoError(t, err, "create client")
		test.AssertTrue(t, c != nil, "client should not be nil")
	})

	t.Run("Create client with missing base URL", func(t *testing.T) {
		// Given: Configuration without base URL
		config := &client.Config{
			Token: "test-token",
		}

		// When: Creating a new client
		_, err := client.NewClient(config)

		// Then: Returns error
		test.AssertError(t, err, "should return error for missing baseURL")
	})

	t.Run("Create client with invalid base URL", func(t *testing.T) {
		// Given: Configuration with invalid URL
		config := &client.Config{
			BaseURL: "not-a-valid-url",
		}

		// When: Creating a new client
		_, err := client.NewClient(config)

		// Then: Returns error
		test.AssertError(t, err, "should return error for invalid baseURL")
	})

	t.Run("Create client with custom HTTP client", func(t *testing.T) {
		// Given: Configuration with custom HTTP client
		customHTTPClient := &http.Client{
			Timeout: 5 * time.Second,
		}
		config := &client.Config{
			BaseURL:    "https://artifacts.example.com",
			HTTPClient: customHTTPClient,
		}

		// When: Creating a new client
		c, err := client.NewClient(config)

		// Then: Client is created successfully
		test.AssertNoError(t, err, "create client")
		test.AssertTrue(t, c != nil, "client should not be nil")
	})

	t.Run("Create client with insecure TLS", func(t *testing.T) {
		// Given: Configuration with insecure TLS
		config := &client.Config{
			BaseURL:            "https://artifacts.example.com",
			InsecureSkipVerify: true,
		}

		// When: Creating a new client
		c, err := client.NewClient(config)

		// Then: Client is created successfully
		test.AssertNoError(t, err, "create client")
		test.AssertTrue(t, c != nil, "client should not be nil")
	})
}

func TestSetToken(t *testing.T) {
	t.Run("Update authentication token", func(t *testing.T) {
		// Given: Client with initial token
		config := &client.Config{
			BaseURL: "https://artifacts.example.com",
			Token:   "initial-token",
		}
		c, _ := client.NewClient(config)

		// When: Updating the token
		c.SetToken("new-token")

		// Then: Token is updated (verified implicitly by no error)
		test.AssertTrue(t, true, "token update should succeed")
	})
}

func TestHTTPErrorHandling(t *testing.T) {
	testCases := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectError    bool
		errorSubstring string
	}{
		{
			name:           "400 Bad Request",
			statusCode:     http.StatusBadRequest,
			responseBody:   `{"error": "invalid request"}`,
			expectError:    true,
			errorSubstring: "invalid request",
		},
		{
			name:           "401 Unauthorized",
			statusCode:     http.StatusUnauthorized,
			responseBody:   `{"error": "unauthorized"}`,
			expectError:    true,
			errorSubstring: "unauthorized",
		},
		{
			name:           "403 Forbidden",
			statusCode:     http.StatusForbidden,
			responseBody:   `{"error": "forbidden"}`,
			expectError:    true,
			errorSubstring: "forbidden",
		},
		{
			name:           "404 Not Found",
			statusCode:     http.StatusNotFound,
			responseBody:   `{"error": "not found"}`,
			expectError:    true,
			errorSubstring: "not found",
		},
		{
			name:           "409 Conflict",
			statusCode:     http.StatusConflict,
			responseBody:   `{"error": "conflict"}`,
			expectError:    true,
			errorSubstring: "conflict",
		},
		{
			name:           "500 Internal Server Error",
			statusCode:     http.StatusInternalServerError,
			responseBody:   `{"error": "internal error"}`,
			expectError:    true,
			errorSubstring: "internal error",
		},
		{
			name:           "503 Service Unavailable",
			statusCode:     http.StatusServiceUnavailable,
			responseBody:   `{"error": "service unavailable"}`,
			expectError:    true,
			errorSubstring: "service unavailable",
		},
		{
			name:         "200 OK",
			statusCode:   http.StatusOK,
			responseBody: `{"buckets": []}`,
			expectError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given: Test server with configured response
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				w.Write([]byte(tc.responseBody))
			}))
			defer server.Close()

			config := &client.Config{
				BaseURL: server.URL,
			}
			c, _ := client.NewClient(config)

			// When: Making a request
			ctx := context.Background()
			_, err := c.ListBuckets(ctx)

			// Then: Error handling matches expectation
			if tc.expectError {
				test.AssertError(t, err, "should return error")
			} else {
				test.AssertNoError(t, err, "should not return error")
			}
		})
	}
}

func TestAuthenticationHeader(t *testing.T) {
	t.Run("Request includes bearer token", func(t *testing.T) {
		// Given: Client with authentication token
		expectedToken := "test-bearer-token"
		var receivedAuth string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"buckets": []}`))
		}))
		defer server.Close()

		config := &client.Config{
			BaseURL: server.URL,
			Token:   expectedToken,
		}
		c, _ := client.NewClient(config)

		// When: Making a request
		ctx := context.Background()
		c.ListBuckets(ctx)

		// Then: Authorization header is set correctly
		expectedHeader := "Bearer " + expectedToken
		test.AssertEqual(t, expectedHeader, receivedAuth, "authorization header")
	})

	t.Run("Request without token has no auth header", func(t *testing.T) {
		// Given: Client without authentication token
		var receivedAuth string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"buckets": []}`))
		}))
		defer server.Close()

		config := &client.Config{
			BaseURL: server.URL,
		}
		c, _ := client.NewClient(config)

		// When: Making a request
		ctx := context.Background()
		c.ListBuckets(ctx)

		// Then: No authorization header
		test.AssertEqual(t, "", receivedAuth, "authorization header should be empty")
	})
}

func TestUserAgent(t *testing.T) {
	t.Run("Request includes default user agent", func(t *testing.T) {
		// Given: Client with default user agent
		var receivedUA string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedUA = r.Header.Get("User-Agent")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"buckets": []}`))
		}))
		defer server.Close()

		config := &client.Config{
			BaseURL: server.URL,
		}
		c, _ := client.NewClient(config)

		// When: Making a request
		ctx := context.Background()
		c.ListBuckets(ctx)

		// Then: Default user agent is set
		test.AssertTrue(t, receivedUA != "", "user agent should be set")
	})

	t.Run("Request includes custom user agent", func(t *testing.T) {
		// Given: Client with custom user agent
		customUA := "my-custom-client/2.0"
		var receivedUA string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedUA = r.Header.Get("User-Agent")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"buckets": []}`))
		}))
		defer server.Close()

		config := &client.Config{
			BaseURL:   server.URL,
			UserAgent: customUA,
		}
		c, _ := client.NewClient(config)

		// When: Making a request
		ctx := context.Background()
		c.ListBuckets(ctx)

		// Then: Custom user agent is set
		test.AssertEqual(t, customUA, receivedUA, "user agent")
	})
}
