package extensions_test

import (
	"testing"

	"github.com/candlekeep/zot-artifact-store/internal/extensions"
	"github.com/candlekeep/zot-artifact-store/internal/extensions/s3api"
	"github.com/candlekeep/zot-artifact-store/test"
	"zotregistry.io/zot/pkg/api/config"
)

// TestExtensionRegistry tests the extension registry functionality
// Following Given-When-Then pattern for AI-friendly tests
func TestExtensionRegistry(t *testing.T) {
	t.Run("Register extension successfully", func(t *testing.T) {
		// Given: A new extension registry
		logger := test.NewTestLogger(t)
		registry := extensions.NewRegistry(logger)

		// When: Registering an extension
		ext := s3api.NewS3APIExtension()
		err := registry.Register(ext)

		// Then: Extension is registered without error
		test.AssertNoError(t, err, "registering extension")
		retrievedExt, exists := registry.Get("s3api")
		test.AssertTrue(t, exists, "extension should exist")
		test.AssertEqual(t, ext, retrievedExt, "retrieved extension should match registered extension")
	})

	t.Run("Get all registered extensions", func(t *testing.T) {
		// Given: A registry with multiple extensions
		logger := test.NewTestLogger(t)
		registry := extensions.NewRegistry(logger)
		ext1 := s3api.NewS3APIExtension()
		registry.Register(ext1)

		// When: Getting all extensions
		allExts := registry.GetAll()

		// Then: All registered extensions are returned
		test.AssertEqual(t, 1, len(allExts), "should have one extension")
		_, exists := allExts["s3api"]
		test.AssertTrue(t, exists, "s3api extension should exist")
	})

	// TODO: Add setup and shutdown tests in Phase 2+ when storage controller is actually used
	// t.Run("Setup all enabled extensions", func(t *testing.T) { ... })
	// t.Run("Shutdown all extensions", func(t *testing.T) { ... })
}

// TestS3APIExtension tests the S3 API extension
func TestS3APIExtension(t *testing.T) {
	t.Run("Extension has correct name", func(t *testing.T) {
		// Given: A new S3 API extension
		ext := s3api.NewS3APIExtension()

		// When: Getting the extension name
		name := ext.Name()

		// Then: Name is correct
		test.AssertEqual(t, "s3api", name, "extension name")
	})

	t.Run("Extension is enabled by default", func(t *testing.T) {
		// Given: A new S3 API extension and config
		ext := s3api.NewS3APIExtension()
		cfg := config.New()

		// When: Checking if extension is enabled
		enabled := ext.IsEnabled(cfg)

		// Then: Extension is enabled
		test.AssertTrue(t, enabled, "extension should be enabled by default")
	})

	// TODO: Add setup test in Phase 2+ when storage controller is actually used
	// t.Run("Extension setup succeeds", func(t *testing.T) { ... })
}

// Benchmark tests for performance
func BenchmarkExtensionRegistration(b *testing.B) {
	logger := test.NewTestLogger(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry := extensions.NewRegistry(logger)
		ext := s3api.NewS3APIExtension()
		registry.Register(ext)
	}
}
