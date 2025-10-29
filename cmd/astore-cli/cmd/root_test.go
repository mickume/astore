package cmd

import (
	"testing"
)

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
		{5368709120, "5.0 GB"},
	}

	for _, tt := range tests {
		result := formatSize(tt.bytes)
		if result != tt.expected {
			t.Errorf("formatSize(%d) = %s; want %s", tt.bytes, result, tt.expected)
		}
	}
}

func TestGetBucketAndKey(t *testing.T) {
	tests := []struct {
		path        string
		wantBucket  string
		wantKey     string
		expectError bool
	}{
		{"bucket/key", "bucket", "key", false},
		{"/bucket/key", "bucket", "key", false},
		{"bucket/path/to/key", "bucket", "path/to/key", false},
		{"mybucket/mykey.tar.gz", "mybucket", "mykey.tar.gz", false},
		{"bucket", "", "", true}, // Missing key
		{"/bucket", "", "", true}, // Missing key
		{"", "", "", true},        // Empty path
	}

	for _, tt := range tests {
		bucket, key, err := getBucketAndKey(tt.path)

		if tt.expectError {
			if err == nil {
				t.Errorf("getBucketAndKey(%q) expected error but got none", tt.path)
			}
			continue
		}

		if err != nil {
			t.Errorf("getBucketAndKey(%q) unexpected error: %v", tt.path, err)
			continue
		}

		if bucket != tt.wantBucket {
			t.Errorf("getBucketAndKey(%q) bucket = %q; want %q", tt.path, bucket, tt.wantBucket)
		}

		if key != tt.wantKey {
			t.Errorf("getBucketAndKey(%q) key = %q; want %q", tt.path, key, tt.wantKey)
		}
	}
}

func TestGuessContentType(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"file.tar", "application/x-tar"},
		{"file.gz", "application/gzip"},
		{"file.gzip", "application/gzip"},
		{"file.tgz", "application/gzip"},
		{"file.zip", "application/zip"},
		{"file.json", "application/json"},
		{"file.xml", "application/xml"},
		{"file.txt", "text/plain"},
		{"file.md", "text/markdown"},
		{"file.bin", "application/octet-stream"},
		{"file", "application/octet-stream"},
	}

	for _, tt := range tests {
		result := guessContentType(tt.filename)
		if result != tt.expected {
			t.Errorf("guessContentType(%q) = %q; want %q", tt.filename, result, tt.expected)
		}
	}
}
