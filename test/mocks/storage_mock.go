package mocks

import (
	"io"

	godigest "github.com/opencontainers/go-digest"
	ispec "github.com/opencontainers/image-spec/specs-go/v1"
	"zotregistry.io/zot/pkg/storage"
)

// MockedImageStore is a mock implementation of storage.ImageStore for testing
type MockedImageStore struct {
	// Add fields for tracking method calls if needed
	storage.ImageStore
}

// NewMockedImageStore creates a new mock image store
func NewMockedImageStore() *MockedImageStore {
	return &MockedImageStore{}
}

// MockedStoreController is a mock implementation of storage.StoreController
type MockedStoreController struct {
	DefaultStore *MockedImageStore
	SubStores    map[string]*MockedImageStore
}

// NewMockedStoreController creates a new mock store controller
func NewMockedStoreController() *MockedStoreController {
	return &MockedStoreController{
		DefaultStore: NewMockedImageStore(),
		SubStores:    make(map[string]*MockedImageStore),
	}
}

// GetImageStore returns the default mock image store
func (m *MockedStoreController) GetImageStore(name string) storage.ImageStore {
	if store, ok := m.SubStores[name]; ok {
		return store
	}
	return m.DefaultStore
}

// GetDefaultImageStore returns the default image store
func (m *MockedStoreController) GetDefaultImageStore() storage.ImageStore {
	return m.DefaultStore
}

// MockedBlobUpload is a mock implementation of storage.BlobUpload
type MockedBlobUpload struct {
	storage.BlobUpload
}

// Write implements io.Writer
func (m *MockedBlobUpload) Write(p []byte) (int, error) {
	return len(p), nil
}

// Close closes the upload
func (m *MockedBlobUpload) Close() error {
	return nil
}

// GetCurrentSize returns current upload size
func (m *MockedBlobUpload) GetCurrentSize() int64 {
	return 0
}

// GetDigest returns the digest
func (m *MockedBlobUpload) GetDigest() godigest.Digest {
	return godigest.FromString("mock")
}

// NewBlobUpload creates a new mock blob upload
func NewBlobUpload() *MockedBlobUpload {
	return &MockedBlobUpload{}
}

// PutBlobChunk mocks putting a blob chunk
func (m *MockedImageStore) PutBlobChunk(repo, upload string, from, to int64, body io.Reader) (int64, error) {
	return to - from, nil
}

// NewBlobUpload mocks creating a new blob upload
func (m *MockedImageStore) NewBlobUpload(repo string) (string, error) {
	return "mock-upload-id", nil
}

// FinishBlobUpload mocks finishing a blob upload
func (m *MockedImageStore) FinishBlobUpload(repo, uuid string, body io.Reader, digest godigest.Digest) error {
	return nil
}

// FullBlobUpload mocks a full blob upload
func (m *MockedImageStore) FullBlobUpload(repo string, body io.Reader, digest godigest.Digest) (string, int64, error) {
	return digest.String(), 0, nil
}

// GetBlob mocks getting a blob
func (m *MockedImageStore) GetBlob(repo string, digest godigest.Digest, mediaType string) (io.ReadCloser, int64, error) {
	return io.NopCloser(io.LimitReader(nil, 0)), 0, nil
}

// PutImageManifest mocks putting an image manifest
func (m *MockedImageStore) PutImageManifest(repo, reference, mediaType string, body []byte) (godigest.Digest, error) {
	return godigest.FromBytes(body), nil
}

// GetImageManifest mocks getting an image manifest
func (m *MockedImageStore) GetImageManifest(repo, reference string) ([]byte, godigest.Digest, string, error) {
	return []byte("{}"), godigest.FromString("mock"), ispec.MediaTypeImageManifest, nil
}

// DeleteImageManifest mocks deleting an image manifest
func (m *MockedImageStore) DeleteImageManifest(repo, reference string, detectCollision bool) error {
	return nil
}

// DeleteBlob mocks deleting a blob
func (m *MockedImageStore) DeleteBlob(repo string, digest godigest.Digest) error {
	return nil
}
