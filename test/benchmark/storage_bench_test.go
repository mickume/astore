package benchmark_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/candlekeep/zot-artifact-store/internal/storage"
	"zotregistry.io/zot/pkg/log"
)

// BenchmarkMetadataStore benchmarks metadata store operations
func BenchmarkMetadataStore(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "bench-metadata-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := log.NewLogger("error", "")
	store, err := storage.NewMetadataStore(tmpDir + "/metadata.db")
	if err != nil {
		b.Fatal(err)
	}

	b.Run("CreateBucket", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bucket := fmt.Sprintf("bench-bucket-%d", i)
			store.CreateBucket(context.Background(), bucket)
		}
	})

	// Setup test bucket for other benchmarks
	testBucket := "test-bucket"
	store.CreateBucket(context.Background(), testBucket)

	b.Run("StoreObjectMetadata", func(b *testing.B) {
		metadata := &storage.ObjectMetadata{
			Bucket:      testBucket,
			Key:         "test-key",
			Size:        1024,
			ContentType: "application/octet-stream",
			ETag:        "abc123",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			metadata.Key = fmt.Sprintf("bench-key-%d", i)
			store.StoreObjectMetadata(context.Background(), metadata)
		}
	})

	b.Run("GetObjectMetadata", func(b *testing.B) {
		// Setup test data
		metadata := &storage.ObjectMetadata{
			Bucket:      testBucket,
			Key:         "lookup-key",
			Size:        1024,
			ContentType: "application/octet-stream",
			ETag:        "abc123",
		}
		store.StoreObjectMetadata(context.Background(), metadata)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			store.GetObjectMetadata(context.Background(), testBucket, "lookup-key")
		}
	})

	b.Run("ListObjects", func(b *testing.B) {
		// Setup test data
		for i := 0; i < 100; i++ {
			metadata := &storage.ObjectMetadata{
				Bucket:      testBucket,
				Key:         fmt.Sprintf("list-key-%d", i),
				Size:        1024,
				ContentType: "application/octet-stream",
				ETag:        fmt.Sprintf("etag-%d", i),
			}
			store.StoreObjectMetadata(context.Background(), metadata)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			store.ListObjects(context.Background(), testBucket, "", 100)
		}
	})

	b.Run("DeleteObjectMetadata", func(b *testing.B) {
		// Setup test data
		for i := 0; i < b.N; i++ {
			metadata := &storage.ObjectMetadata{
				Bucket:      testBucket,
				Key:         fmt.Sprintf("delete-key-%d", i),
				Size:        1024,
				ContentType: "application/octet-stream",
				ETag:        fmt.Sprintf("etag-%d", i),
			}
			store.StoreObjectMetadata(context.Background(), metadata)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			store.DeleteObjectMetadata(context.Background(), testBucket, fmt.Sprintf("delete-key-%d", i))
		}
	})

	_ = logger
}

// BenchmarkFileStorage benchmarks file storage operations
func BenchmarkFileStorage(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "bench-storage-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := log.NewLogger("error", "")
	fstorage := storage.NewFileStorage(tmpDir, logger)

	bucket := "bench-bucket"
	fstorage.CreateBucket(context.Background(), bucket)

	sizes := []int64{
		1024,           // 1 KB
		10 * 1024,      // 10 KB
		100 * 1024,     // 100 KB
		1024 * 1024,    // 1 MB
		10 * 1024 * 1024, // 10 MB
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("WriteObject_%dKB", size/1024), func(b *testing.B) {
			data := bytes.Repeat([]byte("x"), int(size))
			reader := bytes.NewReader(data)

			b.SetBytes(size)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("write-bench-%d", i)
				reader.Reset(data)
				fstorage.WriteObject(context.Background(), bucket, key, reader, size)
			}
		})

		b.Run(fmt.Sprintf("ReadObject_%dKB", size/1024), func(b *testing.B) {
			// Setup test data
			data := bytes.Repeat([]byte("x"), int(size))
			key := "read-bench"
			fstorage.WriteObject(context.Background(), bucket, key, bytes.NewReader(data), size)

			b.SetBytes(size)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				reader, _ := fstorage.ReadObject(context.Background(), bucket, key)
				if reader != nil {
					io.Copy(io.Discard, reader)
					reader.Close()
				}
			}
		})
	}

	b.Run("DeleteObject", func(b *testing.B) {
		// Setup test data
		data := bytes.Repeat([]byte("x"), 1024)
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("delete-bench-%d", i)
			fstorage.WriteObject(context.Background(), bucket, key, bytes.NewReader(data), 1024)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("delete-bench-%d", i)
			fstorage.DeleteObject(context.Background(), bucket, key)
		}
	})
}

// BenchmarkConcurrentOperations benchmarks concurrent storage operations
func BenchmarkConcurrentOperations(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "bench-concurrent-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := log.NewLogger("error", "")
	metaStore, err := storage.NewMetadataStore(tmpDir + "/metadata.db")
	if err != nil {
		b.Fatal(err)
	}
	fileStore := storage.NewFileStorage(tmpDir+"/artifacts", logger)

	bucket := "concurrent-bucket"
	metaStore.CreateBucket(context.Background(), bucket)
	fileStore.CreateBucket(context.Background(), bucket)

	concurrencies := []int{1, 2, 4, 8, 16, 32}

	for _, concurrency := range concurrencies {
		b.Run(fmt.Sprintf("ConcurrentWrites_%d", concurrency), func(b *testing.B) {
			data := bytes.Repeat([]byte("x"), 1024)

			b.SetParallelism(concurrency)
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					key := fmt.Sprintf("concurrent-write-%d", i)
					fileStore.WriteObject(context.Background(), bucket, key, bytes.NewReader(data), 1024)
					i++
				}
			})
		})

		b.Run(fmt.Sprintf("ConcurrentReads_%d", concurrency), func(b *testing.B) {
			// Setup test data
			data := bytes.Repeat([]byte("x"), 1024)
			key := "concurrent-read"
			fileStore.WriteObject(context.Background(), bucket, key, bytes.NewReader(data), 1024)

			b.SetParallelism(concurrency)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					reader, _ := fileStore.ReadObject(context.Background(), bucket, key)
					if reader != nil {
						io.Copy(io.Discard, reader)
						reader.Close()
					}
				}
			})
		})
	}
}

// BenchmarkMultipartUpload benchmarks multipart upload operations
func BenchmarkMultipartUpload(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "bench-multipart-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := log.NewLogger("error", "")
	metaStore, err := storage.NewMetadataStore(tmpDir + "/metadata.db")
	if err != nil {
		b.Fatal(err)
	}

	bucket := "multipart-bucket"
	key := "large-object"

	b.Run("InitiateMultipartUpload", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			metaStore.InitiateMultipartUpload(context.Background(), bucket, key)
		}
	})

	b.Run("UploadPart", func(b *testing.B) {
		// Setup
		uploadID, _ := metaStore.InitiateMultipartUpload(context.Background(), bucket, key)
		partData := bytes.Repeat([]byte("x"), 5*1024*1024) // 5 MB parts

		b.SetBytes(int64(len(partData)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			part := &storage.MultipartPart{
				PartNumber: i + 1,
				Size:       int64(len(partData)),
				ETag:       fmt.Sprintf("etag-%d", i),
			}
			metaStore.StorePart(context.Background(), uploadID, part)
		}
	})

	b.Run("CompleteMultipartUpload", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			uploadID, _ := metaStore.InitiateMultipartUpload(context.Background(), bucket, key)

			// Add parts
			for j := 0; j < 5; j++ {
				part := &storage.MultipartPart{
					PartNumber: j + 1,
					Size:       5 * 1024 * 1024,
					ETag:       fmt.Sprintf("etag-%d", j),
				}
				metaStore.StorePart(context.Background(), uploadID, part)
			}

			// Complete upload
			metaStore.CompleteMultipartUpload(context.Background(), uploadID)
		}
	})
}
