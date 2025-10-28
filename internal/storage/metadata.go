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
	bucketsBucket        = []byte("buckets")
	artifactsBucket      = []byte("artifacts")
	multipartBucket      = []byte("multipart_uploads")
	uploadProgressBucket = []byte("upload_progress")
	policiesBucket       = []byte("policies")
	auditLogsBucket      = []byte("audit_logs")
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
		for _, bucket := range [][]byte{bucketsBucket, artifactsBucket, multipartBucket, uploadProgressBucket, policiesBucket, auditLogsBucket} {
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

// === Policy Operations ===

// StorePolicy stores a policy
func (s *MetadataStore) StorePolicy(policy *models.Policy) error {
	policy.UpdatedAt = time.Now()
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = policy.UpdatedAt
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(policiesBucket)
		data, err := json.Marshal(policy)
		if err != nil {
			return fmt.Errorf("failed to marshal policy: %w", err)
		}
		return b.Put([]byte(policy.ID), data)
	})
}

// GetPolicy retrieves a policy by ID
func (s *MetadataStore) GetPolicy(id string) (*models.Policy, error) {
	var policy models.Policy
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(policiesBucket)
		data := b.Get([]byte(id))
		if data == nil {
			return fmt.Errorf("policy %s not found", id)
		}
		return json.Unmarshal(data, &policy)
	})
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// ListPolicies lists all policies
func (s *MetadataStore) ListPolicies() ([]*models.Policy, error) {
	var policies []*models.Policy

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(policiesBucket)
		return b.ForEach(func(k, v []byte) error {
			var policy models.Policy
			if err := json.Unmarshal(v, &policy); err != nil {
				return err
			}
			policies = append(policies, &policy)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}
	return policies, nil
}

// DeletePolicy deletes a policy
func (s *MetadataStore) DeletePolicy(id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(policiesBucket)
		return b.Delete([]byte(id))
	})
}

// === Audit Log Operations ===

// StoreAuditLog stores an audit log entry
func (s *MetadataStore) StoreAuditLog(log *models.AuditLog) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(auditLogsBucket)
		data, err := json.Marshal(log)
		if err != nil {
			return fmt.Errorf("failed to marshal audit log: %w", err)
		}
		// Use timestamp + ID as key for chronological ordering
		key := []byte(fmt.Sprintf("%d_%s", log.Timestamp.Unix(), log.ID))
		return b.Put(key, data)
	})
}

// ListAuditLogs retrieves audit logs with optional filtering
func (s *MetadataStore) ListAuditLogs(userID string, resource string, startTime, endTime time.Time, limit int) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(auditLogsBucket)
		c := b.Cursor()

		// Iterate in reverse chronological order
		count := 0
		for k, v := c.Last(); k != nil && (limit == 0 || count < limit); k, v = c.Prev() {
			var log models.AuditLog
			if err := json.Unmarshal(v, &log); err != nil {
				continue
			}

			// Apply filters
			if userID != "" && log.UserID != userID {
				continue
			}
			if resource != "" && log.Resource != resource {
				continue
			}
			if !startTime.IsZero() && log.Timestamp.Before(startTime) {
				continue
			}
			if !endTime.IsZero() && log.Timestamp.After(endTime) {
				continue
			}

			logs = append(logs, &log)
			count++
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return logs, nil
}

// === Helper Functions ===

func artifactKey(bucket, key string) string {
	return bucket + "/" + key
}
