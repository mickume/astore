package reliability

import (
	"context"
	"fmt"
	"io"
	"time"

	"zotregistry.io/zot/pkg/log"
)

// ProgressTracker tracks upload/download progress
type ProgressTracker struct {
	TotalBytes      int64
	CompletedBytes  int64
	StartTime       time.Time
	LastUpdateTime  time.Time
	PartialProgress map[int]int64 // For multipart uploads: partNumber -> bytes
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(totalBytes int64) *ProgressTracker {
	return &ProgressTracker{
		TotalBytes:      totalBytes,
		CompletedBytes:  0,
		StartTime:       time.Now(),
		LastUpdateTime:  time.Now(),
		PartialProgress: make(map[int]int64),
	}
}

// Update updates the progress
func (p *ProgressTracker) Update(bytes int64) {
	p.CompletedBytes += bytes
	p.LastUpdateTime = time.Now()
}

// UpdatePart updates progress for a specific part
func (p *ProgressTracker) UpdatePart(partNumber int, bytes int64) {
	p.PartialProgress[partNumber] = bytes
	p.LastUpdateTime = time.Now()

	// Recalculate total completed bytes
	total := int64(0)
	for _, partBytes := range p.PartialProgress {
		total += partBytes
	}
	p.CompletedBytes = total
}

// GetProgress returns the current progress percentage
func (p *ProgressTracker) GetProgress() float64 {
	if p.TotalBytes == 0 {
		return 0
	}
	return float64(p.CompletedBytes) / float64(p.TotalBytes) * 100
}

// GetSpeed returns the current transfer speed in bytes per second
func (p *ProgressTracker) GetSpeed() float64 {
	elapsed := time.Since(p.StartTime).Seconds()
	if elapsed == 0 {
		return 0
	}
	return float64(p.CompletedBytes) / elapsed
}

// GetETA returns the estimated time to completion
func (p *ProgressTracker) GetETA() time.Duration {
	speed := p.GetSpeed()
	if speed == 0 {
		return 0
	}
	remaining := p.TotalBytes - p.CompletedBytes
	return time.Duration(float64(remaining)/speed) * time.Second
}

// ResumableUpload handles resumable upload operations
type ResumableUpload struct {
	UploadID  string
	Bucket    string
	Key       string
	TotalSize int64
	Parts     map[int]*UploadPart
	tracker   *ProgressTracker
	logger    log.Logger
}

// UploadPart represents a single part of a multipart upload
type UploadPart struct {
	PartNumber int
	Size       int64
	Offset     int64
	ETag       string
	Completed  bool
}

// NewResumableUpload creates a new resumable upload session
func NewResumableUpload(uploadID, bucket, key string, totalSize int64, logger log.Logger) *ResumableUpload {
	return &ResumableUpload{
		UploadID:  uploadID,
		Bucket:    bucket,
		Key:       key,
		TotalSize: totalSize,
		Parts:     make(map[int]*UploadPart),
		tracker:   NewProgressTracker(totalSize),
		logger:    logger,
	}
}

// AddPart adds or updates a part in the upload
func (r *ResumableUpload) AddPart(partNumber int, size, offset int64, etag string, completed bool) {
	r.Parts[partNumber] = &UploadPart{
		PartNumber: partNumber,
		Size:       size,
		Offset:     offset,
		ETag:       etag,
		Completed:  completed,
	}

	if completed {
		r.tracker.UpdatePart(partNumber, size)
	}
}

// GetNextPart returns the next incomplete part number
func (r *ResumableUpload) GetNextPart() (int, *UploadPart, bool) {
	for partNum, part := range r.Parts {
		if !part.Completed {
			return partNum, part, true
		}
	}
	return 0, nil, false
}

// GetProgress returns upload progress information
func (r *ResumableUpload) GetProgress() map[string]interface{} {
	completedParts := 0
	for _, part := range r.Parts {
		if part.Completed {
			completedParts++
		}
	}

	return map[string]interface{}{
		"uploadId":       r.UploadID,
		"totalSize":      r.TotalSize,
		"completedBytes": r.tracker.CompletedBytes,
		"totalParts":     len(r.Parts),
		"completedParts": completedParts,
		"percentage":     r.tracker.GetProgress(),
		"speed":          r.tracker.GetSpeed(),
		"eta":            r.tracker.GetETA(),
	}
}

// IsComplete checks if all parts are completed
func (r *ResumableUpload) IsComplete() bool {
	if len(r.Parts) == 0 {
		return false
	}

	for _, part := range r.Parts {
		if !part.Completed {
			return false
		}
	}
	return true
}

// ResumableDownload handles resumable download operations with range requests
type ResumableDownload struct {
	Bucket         string
	Key            string
	TotalSize      int64
	DownloadedSize int64
	tracker        *ProgressTracker
	logger         log.Logger
}

// NewResumableDownload creates a new resumable download session
func NewResumableDownload(bucket, key string, totalSize int64, logger log.Logger) *ResumableDownload {
	return &ResumableDownload{
		Bucket:         bucket,
		Key:            key,
		TotalSize:      totalSize,
		DownloadedSize: 0,
		tracker:        NewProgressTracker(totalSize),
		logger:         logger,
	}
}

// GetRangeHeader returns the HTTP Range header for resuming download
func (r *ResumableDownload) GetRangeHeader() string {
	if r.DownloadedSize == 0 {
		return ""
	}
	// Request from last downloaded byte to end
	return fmt.Sprintf("bytes=%d-", r.DownloadedSize)
}

// UpdateProgress updates the downloaded bytes
func (r *ResumableDownload) UpdateProgress(bytes int64) {
	r.DownloadedSize += bytes
	r.tracker.Update(bytes)
}

// GetProgress returns download progress information
func (r *ResumableDownload) GetProgress() map[string]interface{} {
	return map[string]interface{}{
		"totalSize":      r.TotalSize,
		"downloadedSize": r.DownloadedSize,
		"percentage":     r.tracker.GetProgress(),
		"speed":          r.tracker.GetSpeed(),
		"eta":            r.tracker.GetETA(),
	}
}

// IsComplete checks if download is complete
func (r *ResumableDownload) IsComplete() bool {
	return r.DownloadedSize >= r.TotalSize
}

// PartialRetryReader wraps an io.Reader to track progress and support resume
type PartialRetryReader struct {
	reader   io.Reader
	tracker  *ProgressTracker
	onUpdate func(bytes int64)
}

// NewPartialRetryReader creates a new partial retry reader
func NewPartialRetryReader(reader io.Reader, totalSize int64, onUpdate func(bytes int64)) *PartialRetryReader {
	return &PartialRetryReader{
		reader:   reader,
		tracker:  NewProgressTracker(totalSize),
		onUpdate: onUpdate,
	}
}

// Read implements io.Reader
func (p *PartialRetryReader) Read(buf []byte) (int, error) {
	n, err := p.reader.Read(buf)
	if n > 0 {
		p.tracker.Update(int64(n))
		if p.onUpdate != nil {
			p.onUpdate(int64(n))
		}
	}
	return n, err
}

// GetProgress returns current progress
func (p *PartialRetryReader) GetProgress() *ProgressTracker {
	return p.tracker
}

// PartialRetryUploader handles upload with partial retry capability
type PartialRetryUploader struct {
	retryer *Retryer
	breaker *CircuitBreaker
	logger  log.Logger
}

// NewPartialRetryUploader creates a new partial retry uploader
func NewPartialRetryUploader(retryer *Retryer, breaker *CircuitBreaker, logger log.Logger) *PartialRetryUploader {
	return &PartialRetryUploader{
		retryer: retryer,
		breaker: breaker,
		logger:  logger,
	}
}

// UploadWithRetry uploads data with retry and circuit breaker
func (u *PartialRetryUploader) UploadWithRetry(
	ctx context.Context,
	uploadFn func(ctx context.Context, offset int64, data io.Reader) error,
	data io.Reader,
	totalSize int64,
) error {
	tracker := NewProgressTracker(totalSize)

	return u.breaker.Execute(ctx, func(ctx context.Context) error {
		return u.retryer.Do(ctx, func(ctx context.Context) error {
			// Wrap reader to track progress
			reader := NewPartialRetryReader(data, totalSize, func(bytes int64) {
				tracker.Update(bytes)
			})

			// Attempt upload from current offset
			return uploadFn(ctx, tracker.CompletedBytes, reader)
		})
	})
}
