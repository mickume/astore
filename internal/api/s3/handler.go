package s3

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/candlekeep/zot-artifact-store/internal/models"
	"github.com/candlekeep/zot-artifact-store/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	godigest "github.com/opencontainers/go-digest"
	"zotregistry.io/zot/pkg/log"
)

// Handler handles S3-compatible API requests
type Handler struct {
	metadataStore *storage.MetadataStore
	logger        log.Logger
	dataDir       string
	// TODO: Add Zot storage controller when needed for integration
}

// NewHandler creates a new S3 API handler
func NewHandler(metadataStore *storage.MetadataStore, dataDir string, logger log.Logger) *Handler {
	return &Handler{
		metadataStore: metadataStore,
		logger:        logger,
		dataDir:       dataDir,
	}
}

// RegisterRoutes registers S3 API routes
func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Bucket operations
	router.HandleFunc("/s3", h.ListBuckets).Methods("GET")
	router.HandleFunc("/s3/{bucket}", h.CreateBucket).Methods("PUT")
	router.HandleFunc("/s3/{bucket}", h.DeleteBucket).Methods("DELETE")
	router.HandleFunc("/s3/{bucket}", h.ListObjects).Methods("GET")

	// Object operations
	router.HandleFunc("/s3/{bucket}/{key:.*}", h.PutObject).Methods("PUT").Queries()
	router.HandleFunc("/s3/{bucket}/{key:.*}", h.GetObject).Methods("GET")
	router.HandleFunc("/s3/{bucket}/{key:.*}", h.HeadObject).Methods("HEAD")
	router.HandleFunc("/s3/{bucket}/{key:.*}", h.DeleteObject).Methods("DELETE")

	// Multipart upload operations
	router.HandleFunc("/s3/{bucket}/{key:.*}", h.InitiateMultipartUpload).Methods("POST").Queries("uploads", "")
	router.HandleFunc("/s3/{bucket}/{key:.*}", h.UploadPart).Methods("PUT").Queries("uploadId", "{uploadId}", "partNumber", "{partNumber}")
	router.HandleFunc("/s3/{bucket}/{key:.*}", h.CompleteMultipartUpload).Methods("POST").Queries("uploadId", "{uploadId}")
	router.HandleFunc("/s3/{bucket}/{key:.*}", h.AbortMultipartUpload).Methods("DELETE").Queries("uploadId", "{uploadId}")
}

// === Bucket Operations ===

// ListBuckets lists all buckets
func (h *Handler) ListBuckets(w http.ResponseWriter, r *http.Request) {
	buckets, err := h.metadataStore.ListBuckets()
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list buckets")
		http.Error(w, "Failed to list buckets", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"buckets": buckets,
	})
}

// CreateBucket creates a new bucket
func (h *Handler) CreateBucket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucket"]

	bucket := &models.Bucket{
		Name:        bucketName,
		Versioning:  false,
		ObjectCount: 0,
		TotalSize:   0,
	}

	if err := h.metadataStore.CreateBucket(bucket); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, "Bucket already exists", http.StatusConflict)
			return
		}
		h.logger.Error().Err(err).Str("bucket", bucketName).Msg("failed to create bucket")
		http.Error(w, "Failed to create bucket", http.StatusInternalServerError)
		return
	}

	h.logger.Info().Str("bucket", bucketName).Msg("bucket created")
	w.WriteHeader(http.StatusOK)
}

// DeleteBucket deletes a bucket
func (h *Handler) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucket"]

	// Check if bucket has objects
	artifacts, err := h.metadataStore.ListArtifacts(bucketName, "", 1)
	if err != nil {
		h.logger.Error().Err(err).Str("bucket", bucketName).Msg("failed to check bucket contents")
		http.Error(w, "Failed to check bucket", http.StatusInternalServerError)
		return
	}

	if len(artifacts) > 0 {
		http.Error(w, "Bucket not empty", http.StatusConflict)
		return
	}

	if err := h.metadataStore.DeleteBucket(bucketName); err != nil {
		h.logger.Error().Err(err).Str("bucket", bucketName).Msg("failed to delete bucket")
		http.Error(w, "Failed to delete bucket", http.StatusInternalServerError)
		return
	}

	h.logger.Info().Str("bucket", bucketName).Msg("bucket deleted")
	w.WriteHeader(http.StatusNoContent)
}

// ListObjects lists objects in a bucket
func (h *Handler) ListObjects(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucket"]

	// Parse query parameters
	prefix := r.URL.Query().Get("prefix")
	maxKeysStr := r.URL.Query().Get("max-keys")
	maxKeys := 1000 // default
	if maxKeysStr != "" {
		if parsed, err := strconv.Atoi(maxKeysStr); err == nil && parsed > 0 {
			maxKeys = parsed
		}
	}

	artifacts, err := h.metadataStore.ListArtifacts(bucketName, prefix, maxKeys)
	if err != nil {
		h.logger.Error().Err(err).Str("bucket", bucketName).Msg("failed to list objects")
		http.Error(w, "Failed to list objects", http.StatusInternalServerError)
		return
	}

	// Convert to S3 format
	objects := make([]models.Object, 0, len(artifacts))
	for _, artifact := range artifacts {
		objects = append(objects, models.Object{
			Key:          artifact.Key,
			Size:         artifact.Size,
			LastModified: artifact.UpdatedAt,
			ETag:         artifact.MD5,
		})
	}

	result := models.ListObjectsResult{
		Bucket:      bucketName,
		Prefix:      prefix,
		MaxKeys:     maxKeys,
		IsTruncated: len(objects) >= maxKeys,
		Objects:     objects,
	}

	h.writeJSON(w, http.StatusOK, result)
}

// === Object Operations ===

// PutObject uploads an object
func (h *Handler) PutObject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucket"]
	key := vars["key"]

	// Check if bucket exists
	bucket, err := h.metadataStore.GetBucket(bucketName)
	if err != nil {
		http.Error(w, "Bucket not found", http.StatusNotFound)
		return
	}

	// Read and store the object
	hash := md5.New()
	tempFile := filepath.Join(h.dataDir, bucketName, key)

	// Calculate digest while reading
	teeReader := io.TeeReader(r.Body, hash)
	size, err := h.saveToFile(tempFile, teeReader)
	if err != nil {
		h.logger.Error().Err(err).Str("bucket", bucketName).Str("key", key).Msg("failed to save object")
		http.Error(w, "Failed to save object", http.StatusInternalServerError)
		return
	}

	md5Sum := hex.EncodeToString(hash.Sum(nil))
	digest := godigest.NewDigestFromHex("sha256", md5Sum) // Simplified for now

	// Store metadata
	artifact := &models.Artifact{
		Bucket:      bucketName,
		Key:         key,
		Digest:      digest,
		Size:        size,
		ContentType: r.Header.Get("Content-Type"),
		MD5:         md5Sum,
		StoragePath: tempFile,
		Metadata:    extractMetadata(r.Header),
	}

	if err := h.metadataStore.StoreArtifact(artifact); err != nil {
		h.logger.Error().Err(err).Str("bucket", bucketName).Str("key", key).Msg("failed to store artifact metadata")
		http.Error(w, "Failed to store metadata", http.StatusInternalServerError)
		return
	}

	// Update bucket statistics
	bucket.ObjectCount++
	bucket.TotalSize += size
	h.metadataStore.UpdateBucket(bucket)

	h.logger.Info().Str("bucket", bucketName).Str("key", key).Int64("size", size).Msg("object uploaded")

	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, md5Sum))
	w.WriteHeader(http.StatusOK)
}

// GetObject downloads an object
func (h *Handler) GetObject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucket"]
	key := vars["key"]

	artifact, err := h.metadataStore.GetArtifact(bucketName, key)
	if err != nil {
		http.Error(w, "Object not found", http.StatusNotFound)
		return
	}

	// Handle range requests
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		h.handleRangeRequest(w, r, artifact, rangeHeader)
		return
	}

	// Full object download
	file, err := h.openFile(artifact.StoragePath)
	if err != nil {
		h.logger.Error().Err(err).Str("bucket", bucketName).Str("key", key).Msg("failed to open object")
		http.Error(w, "Failed to read object", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", artifact.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(artifact.Size, 10))
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, artifact.MD5))
	w.Header().Set("Last-Modified", artifact.UpdatedAt.Format(http.TimeFormat))

	// Set custom metadata headers
	for k, v := range artifact.Metadata {
		w.Header().Set("X-Amz-Meta-"+k, v)
	}

	w.WriteHeader(http.StatusOK)
	io.Copy(w, file)

	h.logger.Info().Str("bucket", bucketName).Str("key", key).Msg("object downloaded")
}

// HeadObject returns object metadata
func (h *Handler) HeadObject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucket"]
	key := vars["key"]

	artifact, err := h.metadataStore.GetArtifact(bucketName, key)
	if err != nil {
		http.Error(w, "Object not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", artifact.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(artifact.Size, 10))
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, artifact.MD5))
	w.Header().Set("Last-Modified", artifact.UpdatedAt.Format(http.TimeFormat))

	for k, v := range artifact.Metadata {
		w.Header().Set("X-Amz-Meta-"+k, v)
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteObject deletes an object
func (h *Handler) DeleteObject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucket"]
	key := vars["key"]

	artifact, err := h.metadataStore.GetArtifact(bucketName, key)
	if err != nil {
		// S3 returns 204 even if object doesn't exist
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Delete file
	if err := h.deleteFile(artifact.StoragePath); err != nil {
		h.logger.Error().Err(err).Str("bucket", bucketName).Str("key", key).Msg("failed to delete object file")
	}

	// Delete metadata
	if err := h.metadataStore.DeleteArtifact(bucketName, key); err != nil {
		h.logger.Error().Err(err).Str("bucket", bucketName).Str("key", key).Msg("failed to delete artifact metadata")
		http.Error(w, "Failed to delete object", http.StatusInternalServerError)
		return
	}

	// Update bucket statistics
	bucket, _ := h.metadataStore.GetBucket(bucketName)
	if bucket != nil {
		bucket.ObjectCount--
		bucket.TotalSize -= artifact.Size
		h.metadataStore.UpdateBucket(bucket)
	}

	h.logger.Info().Str("bucket", bucketName).Str("key", key).Msg("object deleted")
	w.WriteHeader(http.StatusNoContent)
}

// === Multipart Upload Operations ===

// InitiateMultipartUpload initiates a multipart upload
func (h *Handler) InitiateMultipartUpload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucket"]
	objectKey := vars["key"]

	// Check if bucket exists
	if _, err := h.metadataStore.GetBucket(bucketName); err != nil {
		http.Error(w, "Bucket not found", http.StatusNotFound)
		return
	}

	uploadID := uuid.New().String()
	upload := &models.MultipartUpload{
		UploadID:    uploadID,
		Bucket:      bucketName,
		Key:         objectKey,
		ContentType: r.Header.Get("Content-Type"),
		Metadata:    extractMetadata(r.Header),
		Parts:       []models.MultipartPart{},
	}

	if err := h.metadataStore.CreateMultipartUpload(upload); err != nil {
		h.logger.Error().Err(err).Str("bucket", bucketName).Str("key", objectKey).Msg("failed to initiate multipart upload")
		http.Error(w, "Failed to initiate upload", http.StatusInternalServerError)
		return
	}

	h.logger.Info().Str("bucket", bucketName).Str("key", objectKey).Str("uploadId", uploadID).Msg("multipart upload initiated")

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"uploadId": uploadID,
		"bucket":   bucketName,
		"key":      objectKey,
	})
}

// UploadPart uploads a part of a multipart upload
func (h *Handler) UploadPart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucket"]
	uploadID := r.URL.Query().Get("uploadId")
	partNumberStr := r.URL.Query().Get("partNumber")

	partNumber, err := strconv.Atoi(partNumberStr)
	if err != nil || partNumber < 1 {
		http.Error(w, "Invalid part number", http.StatusBadRequest)
		return
	}

	upload, err := h.metadataStore.GetMultipartUpload(uploadID)
	if err != nil {
		http.Error(w, "Upload not found", http.StatusNotFound)
		return
	}

	// Store the part
	hash := md5.New()
	partPath := filepath.Join(h.dataDir, bucketName, ".multipart", uploadID, fmt.Sprintf("part-%d", partNumber))

	teeReader := io.TeeReader(r.Body, hash)
	size, err := h.saveToFile(partPath, teeReader)
	if err != nil {
		h.logger.Error().Err(err).Str("uploadId", uploadID).Int("partNumber", partNumber).Msg("failed to save part")
		http.Error(w, "Failed to save part", http.StatusInternalServerError)
		return
	}

	etag := hex.EncodeToString(hash.Sum(nil))

	// Update upload metadata
	part := models.MultipartPart{
		PartNumber: partNumber,
		ETag:       etag,
		Size:       size,
		Digest:     godigest.NewDigestFromHex("sha256", etag),
		UploadedAt: time.Now(),
	}

	// Add or update part
	found := false
	for i, p := range upload.Parts {
		if p.PartNumber == partNumber {
			upload.Parts[i] = part
			found = true
			break
		}
	}
	if !found {
		upload.Parts = append(upload.Parts, part)
	}

	if err := h.metadataStore.UpdateMultipartUpload(upload); err != nil {
		h.logger.Error().Err(err).Str("uploadId", uploadID).Msg("failed to update multipart upload")
		http.Error(w, "Failed to update upload", http.StatusInternalServerError)
		return
	}

	h.logger.Info().Str("uploadId", uploadID).Int("partNumber", partNumber).Int64("size", size).Msg("part uploaded")

	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, etag))
	w.WriteHeader(http.StatusOK)
}

// CompleteMultipartUpload completes a multipart upload
func (h *Handler) CompleteMultipartUpload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucket"]
	objectKey := vars["key"]
	uploadID := r.URL.Query().Get("uploadId")

	upload, err := h.metadataStore.GetMultipartUpload(uploadID)
	if err != nil {
		http.Error(w, "Upload not found", http.StatusNotFound)
		return
	}

	// Combine all parts into final object
	finalPath := filepath.Join(h.dataDir, bucketName, objectKey)
	totalSize := int64(0)

	// TODO: Implement actual part combining
	// For now, we'll use a simplified approach

	// Create artifact metadata
	artifact := &models.Artifact{
		Bucket:      bucketName,
		Key:         objectKey,
		Size:        totalSize,
		ContentType: upload.ContentType,
		StoragePath: finalPath,
		Metadata:    upload.Metadata,
		IsMultipart: true,
		UploadID:    uploadID,
	}

	if err := h.metadataStore.StoreArtifact(artifact); err != nil {
		h.logger.Error().Err(err).Str("uploadId", uploadID).Msg("failed to store artifact")
		http.Error(w, "Failed to complete upload", http.StatusInternalServerError)
		return
	}

	// Clean up multipart upload metadata
	if err := h.metadataStore.DeleteMultipartUpload(uploadID); err != nil {
		h.logger.Warn().Err(err).Str("uploadId", uploadID).Msg("failed to delete multipart upload metadata")
	}

	h.logger.Info().Str("bucket", bucketName).Str("key", objectKey).Str("uploadId", uploadID).Msg("multipart upload completed")

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"bucket": bucketName,
		"key":    objectKey,
		"etag":   artifact.MD5,
	})
}

// AbortMultipartUpload aborts a multipart upload
func (h *Handler) AbortMultipartUpload(w http.ResponseWriter, r *http.Request) {
	uploadID := r.URL.Query().Get("uploadId")

	// Delete multipart upload metadata
	if err := h.metadataStore.DeleteMultipartUpload(uploadID); err != nil {
		h.logger.Error().Err(err).Str("uploadId", uploadID).Msg("failed to abort multipart upload")
		http.Error(w, "Failed to abort upload", http.StatusInternalServerError)
		return
	}

	// TODO: Clean up part files

	h.logger.Info().Str("uploadId", uploadID).Msg("multipart upload aborted")
	w.WriteHeader(http.StatusNoContent)
}

// === Helper Functions ===

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func extractMetadata(headers http.Header) map[string]string {
	metadata := make(map[string]string)
	for key, values := range headers {
		if strings.HasPrefix(key, "X-Amz-Meta-") && len(values) > 0 {
			metaKey := strings.TrimPrefix(key, "X-Amz-Meta-")
			metadata[metaKey] = values[0]
		}
	}
	return metadata
}
