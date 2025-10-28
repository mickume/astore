package storage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/candlekeep/zot-artifact-store/internal/models"
	bolt "go.etcd.io/bbolt"
)

var (
	// Bucket names for BoltDB
	bucketsBucket       = []byte("buckets")
	artifactsBucket     = []byte("artifacts")
	multipartBucket     = []byte("multipart_uploads")
	uploadProgressBucket = []byte("upload_progress")
)

// MetadataStore manages artifact and bucket metadata using BoltDB
type MetadataStore struct {
	db *bolt.DB
}

// NewMetadataStore creates a new metadata store
func NewMetadataStore(dbPath string) (*MetadataStore, error) {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("failed to open metadata database: %w", err)
	}

	// Create buckets if they don't exist
	err = db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range [][]byte{bucketsBucket, artifactsBucket, multipartBucket, uploadProgressBucket} {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", bucket, err)
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	return &MetadataStore{db: db}, nil
}

// Close closes the metadata store
func (s *MetadataStore) Close() error {
	return s.db.Close()
}

// === Bucket Operations ===

// CreateBucket creates a new bucket
func (s *MetadataStore) CreateBucket(bucket *models.Bucket) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketsBucket)
		existing := b.Get([]byte(bucket.Name))
		if existing != nil {
			return fmt.Errorf("bucket %s already exists", bucket.Name)
		}

		bucket.CreatedAt = time.Now()
		bucket.UpdatedAt = bucket.CreatedAt
		data, err := json.Marshal(bucket)
		if err != nil {
			return fmt.Errorf("failed to marshal bucket: %w", err)
		}

		return b.Put([]byte(bucket.Name), data)
	})
}

// GetBucket retrieves a bucket by name
func (s *MetadataStore) GetBucket(name string) (*models.Bucket, error) {
	var bucket models.Bucket
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketsBucket)
		data := b.Get([]byte(name))
		if data == nil {
			return fmt.Errorf("bucket %s not found", name)
		}

		return json.Unmarshal(data, &bucket)
	})
	if err != nil {
		return nil, err
	}
	return &bucket, nil
}

// ListBuckets lists all buckets
func (s *MetadataStore) ListBuckets() ([]*models.Bucket, error) {
	var buckets []*models.Bucket
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketsBucket)
		return b.ForEach(func(k, v []byte) error {
			var bucket models.Bucket
			if err := json.Unmarshal(v, &bucket); err != nil {
				return err
			}
			buckets = append(buckets, &bucket)
			return nil
		})
	})
	return buckets, err
}

// DeleteBucket deletes a bucket
func (s *MetadataStore) DeleteBucket(name string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketsBucket)
		return b.Delete([]byte(name))
	})
}

// UpdateBucket updates bucket metadata
func (s *MetadataStore) UpdateBucket(bucket *models.Bucket) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketsBucket)
		bucket.UpdatedAt = time.Now()
		data, err := json.Marshal(bucket)
		if err != nil {
			return fmt.Errorf("failed to marshal bucket: %w", err)
		}
		return b.Put([]byte(bucket.Name), data)
	})
}

// === Artifact Operations ===

// StoreArtifact stores artifact metadata
func (s *MetadataStore) StoreArtifact(artifact *models.Artifact) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(artifactsBucket)
		key := artifactKey(artifact.Bucket, artifact.Key)

		artifact.UpdatedAt = time.Now()
		if artifact.CreatedAt.IsZero() {
			artifact.CreatedAt = artifact.UpdatedAt
		}

		data, err := json.Marshal(artifact)
		if err != nil {
			return fmt.Errorf("failed to marshal artifact: %w", err)
		}

		return b.Put([]byte(key), data)
	})
}

// GetArtifact retrieves artifact metadata
func (s *MetadataStore) GetArtifact(bucket, key string) (*models.Artifact, error) {
	var artifact models.Artifact
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(artifactsBucket)
		data := b.Get([]byte(artifactKey(bucket, key)))
		if data == nil {
			return fmt.Errorf("artifact %s/%s not found", bucket, key)
		}

		return json.Unmarshal(data, &artifact)
	})
	if err != nil {
		return nil, err
	}
	return &artifact, nil
}

// ListArtifacts lists artifacts in a bucket with optional prefix
func (s *MetadataStore) ListArtifacts(bucket, prefix string, maxKeys int) ([]*models.Artifact, error) {
	var artifacts []*models.Artifact
	count := 0

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(artifactsBucket)
		c := b.Cursor()

		searchPrefix := []byte(bucket + "/")
		if prefix != "" {
			searchPrefix = []byte(artifactKey(bucket, prefix))
		}

		for k, v := c.Seek(searchPrefix); k != nil && count < maxKeys; k, v = c.Next() {
			// Check if key matches bucket prefix
			keyStr := string(k)
			if len(keyStr) < len(bucket)+1 || keyStr[:len(bucket)+1] != bucket+"/" {
				break
			}

			var artifact models.Artifact
			if err := json.Unmarshal(v, &artifact); err != nil {
				return err
			}

			artifacts = append(artifacts, &artifact)
			count++
		}

		return nil
	})

	return artifacts, err
}

// DeleteArtifact deletes artifact metadata
func (s *MetadataStore) DeleteArtifact(bucket, key string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(artifactsBucket)
		return b.Delete([]byte(artifactKey(bucket, key)))
	})
}

// === Multipart Upload Operations ===

// CreateMultipartUpload creates a new multipart upload
func (s *MetadataStore) CreateMultipartUpload(upload *models.MultipartUpload) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(multipartBucket)
		upload.InitiatedAt = time.Now()
		data, err := json.Marshal(upload)
		if err != nil {
			return fmt.Errorf("failed to marshal multipart upload: %w", err)
		}
		return b.Put([]byte(upload.UploadID), data)
	})
}

// GetMultipartUpload retrieves a multipart upload
func (s *MetadataStore) GetMultipartUpload(uploadID string) (*models.MultipartUpload, error) {
	var upload models.MultipartUpload
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(multipartBucket)
		data := b.Get([]byte(uploadID))
		if data == nil {
			return fmt.Errorf("multipart upload %s not found", uploadID)
		}
		return json.Unmarshal(data, &upload)
	})
	if err != nil {
		return nil, err
	}
	return &upload, nil
}

// UpdateMultipartUpload updates multipart upload state
func (s *MetadataStore) UpdateMultipartUpload(upload *models.MultipartUpload) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(multipartBucket)
		data, err := json.Marshal(upload)
		if err != nil {
			return fmt.Errorf("failed to marshal multipart upload: %w", err)
		}
		return b.Put([]byte(upload.UploadID), data)
	})
}

// DeleteMultipartUpload deletes a multipart upload
func (s *MetadataStore) DeleteMultipartUpload(uploadID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(multipartBucket)
		return b.Delete([]byte(uploadID))
	})
}

// === Helper Functions ===

func artifactKey(bucket, key string) string {
	return bucket + "/" + key
}
