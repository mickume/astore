package s3

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/candlekeep/zot-artifact-store/internal/models"
)

// saveToFile saves data from reader to a file at the specified path
func (h *Handler) saveToFile(path string, reader io.Reader) (int64, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, err
	}

	// Create the file
	file, err := os.Create(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// Copy data to file
	size, err := io.Copy(file, reader)
	if err != nil {
		// Clean up on error
		os.Remove(path)
		return 0, err
	}

	return size, nil
}

// openFile opens a file for reading
func (h *Handler) openFile(path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// deleteFile deletes a file
func (h *Handler) deleteFile(path string) error {
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// handleRangeRequest handles HTTP range requests for resumable downloads
func (h *Handler) handleRangeRequest(w http.ResponseWriter, r *http.Request, artifact *models.Artifact, rangeHeader string) {
	file, err := h.openFile(artifact.StoragePath)
	if err != nil {
		h.logger.Error().Err(err).Str("path", artifact.StoragePath).Msg("failed to open file for range request")
		http.Error(w, "Failed to read object", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Parse range header (e.g., "bytes=0-1023")
	rangeHeader = strings.TrimPrefix(rangeHeader, "bytes=")
	rangeParts := strings.Split(rangeHeader, "-")
	if len(rangeParts) != 2 {
		http.Error(w, "Invalid range header", http.StatusBadRequest)
		return
	}

	var start, end int64
	if rangeParts[0] != "" {
		fmt.Sscanf(rangeParts[0], "%d", &start)
	}
	if rangeParts[1] != "" {
		fmt.Sscanf(rangeParts[1], "%d", &end)
	} else {
		end = artifact.Size - 1
	}

	// Validate range
	if start < 0 || end >= artifact.Size || start > end {
		http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", artifact.Size))
		return
	}

	length := end - start + 1

	// Seek to start position
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(start, io.SeekStart); err != nil {
			http.Error(w, "Failed to seek", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", artifact.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(length, 10))
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, artifact.Size))
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, artifact.MD5))
	w.Header().Set("Accept-Ranges", "bytes")
	w.WriteHeader(http.StatusPartialContent)

	// Copy only the requested range
	io.CopyN(w, file, length)

	h.logger.Info().
		Str("bucket", artifact.Bucket).
		Str("key", artifact.Key).
		Int64("start", start).
		Int64("end", end).
		Msg("range request served")
}
