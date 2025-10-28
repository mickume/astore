package models

import (
	"time"

	godigest "github.com/opencontainers/go-digest"
)

// Artifact represents a stored binary artifact with metadata
type Artifact struct {
	// Identity
	Bucket    string          `json:"bucket"`
	Key       string          `json:"key"`
	Digest    godigest.Digest `json:"digest"`

	// Content metadata
	Size        int64     `json:"size"`
	ContentType string    `json:"contentType"`
	MD5         string    `json:"md5,omitempty"`

	// Timestamps
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Storage location
	StoragePath string    `json:"storagePath"`

	// User metadata (custom headers)
	Metadata    map[string]string `json:"metadata,omitempty"`

	// Upload tracking
	UploadID    string    `json:"uploadId,omitempty"`
	IsMultipart bool      `json:"isMultipart"`

	// Supply chain references (for Phase 4)
	Signatures  []string  `json:"signatures,omitempty"`
	SBOMRef     string    `json:"sbomRef,omitempty"`
	Attestations []string `json:"attestations,omitempty"`
}

// Bucket represents a storage bucket/namespace
type Bucket struct {
	Name        string            `json:"name"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`

	// Configuration
	Versioning  bool              `json:"versioning"`

	// Statistics
	ObjectCount int64             `json:"objectCount"`
	TotalSize   int64             `json:"totalSize"`

	// Metadata
	Tags        map[string]string `json:"tags,omitempty"`

	// RBAC references (for Phase 3)
	PolicyRef   string            `json:"policyRef,omitempty"`
}

// MultipartUpload tracks multipart upload state
type MultipartUpload struct {
	UploadID    string            `json:"uploadId"`
	Bucket      string            `json:"bucket"`
	Key         string            `json:"key"`
	InitiatedAt time.Time         `json:"initiatedAt"`
	Parts       []MultipartPart   `json:"parts"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	ContentType string            `json:"contentType"`
}

// MultipartPart represents a single part in a multipart upload
type MultipartPart struct {
	PartNumber int             `json:"partNumber"`
	ETag       string          `json:"etag"`
	Size       int64           `json:"size"`
	Digest     godigest.Digest `json:"digest"`
	UploadedAt time.Time       `json:"uploadedAt"`
}

// ListObjectsResult represents the result of listing objects
type ListObjectsResult struct {
	Bucket        string     `json:"bucket"`
	Prefix        string     `json:"prefix,omitempty"`
	Marker        string     `json:"marker,omitempty"`
	MaxKeys       int        `json:"maxKeys"`
	IsTruncated   bool       `json:"isTruncated"`
	NextMarker    string     `json:"nextMarker,omitempty"`
	Objects       []Object   `json:"objects"`
	CommonPrefixes []string  `json:"commonPrefixes,omitempty"`
}

// Object represents an object in list results
type Object struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
	ETag         string    `json:"etag"`
	StorageClass string    `json:"storageClass,omitempty"`
}

// UploadProgress tracks resumable upload progress
type UploadProgress struct {
	Bucket       string    `json:"bucket"`
	Key          string    `json:"key"`
	UploadID     string    `json:"uploadId"`
	TotalSize    int64     `json:"totalSize"`
	UploadedSize int64     `json:"uploadedSize"`
	StartedAt    time.Time `json:"startedAt"`
	LastActivity time.Time `json:"lastActivity"`
}
