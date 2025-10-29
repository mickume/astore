//go:build ignore
// +build ignore

package benchmark_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/candlekeep/zot-artifact-store/internal/api/s3"
	"github.com/candlekeep/zot-artifact-store/internal/storage"
	"github.com/gorilla/mux"
	"zotregistry.io/zot/pkg/log"
)

// setupBenchServer creates a test server for benchmarking
func setupBenchServer(b *testing.B) (*httptest.Server, string, func()) {
	tmpDir, err := os.MkdirTemp("", "bench-api-*")
	if err != nil {
		b.Fatal(err)
	}

	logger := log.NewLogger("error", "")
	metadataStore, err := storage.NewMetadataStore(tmpDir + "/metadata.db")
	if err != nil {
		b.Fatal(err)
	}

	router := mux.NewRouter()
	handler := s3.NewHandler(metadataStore, tmpDir+"/artifacts", logger)
	handler.RegisterRoutes(router)

	server := httptest.NewServer(router)

	cleanup := func() {
		server.Close()
		os.RemoveAll(tmpDir)
	}

	return server, tmpDir, cleanup
}

// BenchmarkS3APIHandlers benchmarks S3 API HTTP handlers
func BenchmarkS3APIHandlers(b *testing.B) {
	server, _, cleanup := setupBenchServer(b)
	defer cleanup()

	client := &http.Client{}
	baseURL := server.URL
	bucket := "bench-bucket"

	// Create bucket first
	req, _ := http.NewRequest("PUT", baseURL+"/s3/"+bucket, nil)
	client.Do(req)

	b.Run("CreateBucket", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bucketName := fmt.Sprintf("bench-bucket-%d", i)
			req, _ := http.NewRequest("PUT", baseURL+"/s3/"+bucketName, nil)
			resp, _ := client.Do(req)
			if resp != nil {
				resp.Body.Close()
			}
		}
	})

	sizes := []int64{
		1024,        // 1 KB
		10 * 1024,   // 10 KB
		100 * 1024,  // 100 KB
		1024 * 1024, // 1 MB
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("PutObject_%dKB", size/1024), func(b *testing.B) {
			data := bytes.Repeat([]byte("x"), int(size))

			b.SetBytes(size)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("bench-object-%d", i)
				req, _ := http.NewRequest("PUT", baseURL+"/s3/"+bucket+"/"+key, bytes.NewReader(data))
				req.Header.Set("Content-Type", "application/octet-stream")
				resp, _ := client.Do(req)
				if resp != nil {
					resp.Body.Close()
				}
			}
		})

		b.Run(fmt.Sprintf("GetObject_%dKB", size/1024), func(b *testing.B) {
			// Setup test data
			data := bytes.Repeat([]byte("x"), int(size))
			key := "get-bench-object"
			req, _ := http.NewRequest("PUT", baseURL+"/s3/"+bucket+"/"+key, bytes.NewReader(data))
			resp, _ := client.Do(req)
			if resp != nil {
				resp.Body.Close()
			}

			b.SetBytes(size)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				req, _ := http.NewRequest("GET", baseURL+"/s3/"+bucket+"/"+key, nil)
				resp, _ := client.Do(req)
				if resp != nil {
					resp.Body.Close()
				}
			}
		})
	}

	b.Run("HeadObject", func(b *testing.B) {
		// Setup test data
		data := bytes.Repeat([]byte("x"), 1024)
		key := "head-bench-object"
		req, _ := http.NewRequest("PUT", baseURL+"/s3/"+bucket+"/"+key, bytes.NewReader(data))
		resp, _ := client.Do(req)
		if resp != nil {
			resp.Body.Close()
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest("HEAD", baseURL+"/s3/"+bucket+"/"+key, nil)
			resp, _ := client.Do(req)
			if resp != nil {
				resp.Body.Close()
			}
		}
	})

	b.Run("ListObjects", func(b *testing.B) {
		// Setup test data - create 100 objects
		data := bytes.Repeat([]byte("x"), 1024)
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("list-object-%d", i)
			req, _ := http.NewRequest("PUT", baseURL+"/s3/"+bucket+"/"+key, bytes.NewReader(data))
			resp, _ := client.Do(req)
			if resp != nil {
				resp.Body.Close()
			}
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest("GET", baseURL+"/s3/"+bucket, nil)
			resp, _ := client.Do(req)
			if resp != nil {
				resp.Body.Close()
			}
		}
	})

	b.Run("DeleteObject", func(b *testing.B) {
		// Setup test data
		data := bytes.Repeat([]byte("x"), 1024)
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("delete-object-%d", i)
			req, _ := http.NewRequest("PUT", baseURL+"/s3/"+bucket+"/"+key, bytes.NewReader(data))
			resp, _ := client.Do(req)
			if resp != nil {
				resp.Body.Close()
			}
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("delete-object-%d", i)
			req, _ := http.NewRequest("DELETE", baseURL+"/s3/"+bucket+"/"+key, nil)
			resp, _ := client.Do(req)
			if resp != nil {
				resp.Body.Close()
			}
		}
	})
}

// BenchmarkConcurrentAPIRequests benchmarks concurrent HTTP requests
func BenchmarkConcurrentAPIRequests(b *testing.B) {
	server, _, cleanup := setupBenchServer(b)
	defer cleanup()

	baseURL := server.URL
	bucket := "concurrent-bucket"

	// Create bucket
	req, _ := http.NewRequest("PUT", baseURL+"/s3/"+bucket, nil)
	client := &http.Client{}
	resp, _ := client.Do(req)
	if resp != nil {
		resp.Body.Close()
	}

	// Setup test object
	data := bytes.Repeat([]byte("x"), 1024)
	key := "concurrent-object"
	req, _ = http.NewRequest("PUT", baseURL+"/s3/"+bucket+"/"+key, bytes.NewReader(data))
	resp, _ = client.Do(req)
	if resp != nil {
		resp.Body.Close()
	}

	concurrencies := []int{1, 2, 4, 8, 16, 32, 64}

	for _, concurrency := range concurrencies {
		b.Run(fmt.Sprintf("ConcurrentGETs_%d", concurrency), func(b *testing.B) {
			b.SetParallelism(concurrency)
			b.RunParallel(func(pb *testing.PB) {
				client := &http.Client{}
				for pb.Next() {
					req, _ := http.NewRequest("GET", baseURL+"/s3/"+bucket+"/"+key, nil)
					resp, _ := client.Do(req)
					if resp != nil {
						resp.Body.Close()
					}
				}
			})
		})

		b.Run(fmt.Sprintf("ConcurrentPUTs_%d", concurrency), func(b *testing.B) {
			b.SetParallelism(concurrency)
			b.RunParallel(func(pb *testing.PB) {
				client := &http.Client{}
				data := bytes.Repeat([]byte("x"), 1024)
				i := 0
				for pb.Next() {
					key := fmt.Sprintf("concurrent-put-%d", i)
					req, _ := http.NewRequest("PUT", baseURL+"/s3/"+bucket+"/"+key, bytes.NewReader(data))
					resp, _ := client.Do(req)
					if resp != nil {
						resp.Body.Close()
					}
					i++
				}
			})
		})

		b.Run(fmt.Sprintf("ConcurrentHEADs_%d", concurrency), func(b *testing.B) {
			b.SetParallelism(concurrency)
			b.RunParallel(func(pb *testing.PB) {
				client := &http.Client{}
				for pb.Next() {
					req, _ := http.NewRequest("HEAD", baseURL+"/s3/"+bucket+"/"+key, nil)
					resp, _ := client.Do(req)
					if resp != nil {
						resp.Body.Close()
					}
				}
			})
		})
	}
}

// BenchmarkEndToEndWorkflow benchmarks complete artifact lifecycle
func BenchmarkEndToEndWorkflow(b *testing.B) {
	server, _, cleanup := setupBenchServer(b)
	defer cleanup()

	client := &http.Client{}
	baseURL := server.URL

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bucket := fmt.Sprintf("workflow-bucket-%d", i)
		key := "workflow-object"
		data := bytes.Repeat([]byte("x"), 10*1024) // 10 KB

		// 1. Create bucket
		req, _ := http.NewRequest("PUT", baseURL+"/s3/"+bucket, nil)
		resp, _ := client.Do(req)
		if resp != nil {
			resp.Body.Close()
		}

		// 2. Upload object
		req, _ = http.NewRequest("PUT", baseURL+"/s3/"+bucket+"/"+key, bytes.NewReader(data))
		resp, _ = client.Do(req)
		if resp != nil {
			resp.Body.Close()
		}

		// 3. Get object metadata
		req, _ = http.NewRequest("HEAD", baseURL+"/s3/"+bucket+"/"+key, nil)
		resp, _ = client.Do(req)
		if resp != nil {
			resp.Body.Close()
		}

		// 4. Download object
		req, _ = http.NewRequest("GET", baseURL+"/s3/"+bucket+"/"+key, nil)
		resp, _ = client.Do(req)
		if resp != nil {
			resp.Body.Close()
		}

		// 5. List objects
		req, _ = http.NewRequest("GET", baseURL+"/s3/"+bucket, nil)
		resp, _ = client.Do(req)
		if resp != nil {
			resp.Body.Close()
		}

		// 6. Delete object
		req, _ = http.NewRequest("DELETE", baseURL+"/s3/"+bucket+"/"+key, nil)
		resp, _ = client.Do(req)
		if resp != nil {
			resp.Body.Close()
		}

		// 7. Delete bucket
		req, _ = http.NewRequest("DELETE", baseURL+"/s3/"+bucket, nil)
		resp, _ = client.Do(req)
		if resp != nil {
			resp.Body.Close()
		}
	}
}

// BenchmarkMemoryUsage benchmarks memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "bench-memory-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := log.NewLogger("error", "")
	metadataStore, err := storage.NewMetadataStore(tmpDir + "/metadata.db")
	if err != nil {
		b.Fatal(err)
	}
	fileStore := storage.NewFileStorage(tmpDir+"/artifacts", logger)

	bucket := "memory-bucket"
	metadataStore.CreateBucket(context.Background(), bucket)
	fileStore.CreateBucket(context.Background(), bucket)

	b.Run("LargeNumberOfSmallObjects", func(b *testing.B) {
		data := bytes.Repeat([]byte("x"), 1024) // 1 KB

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("small-object-%d", i)
			fileStore.WriteObject(context.Background(), bucket, key, bytes.NewReader(data), 1024)
		}
	})

	b.Run("SmallNumberOfLargeObjects", func(b *testing.B) {
		data := bytes.Repeat([]byte("x"), 10*1024*1024) // 10 MB

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("large-object-%d", i)
			fileStore.WriteObject(context.Background(), bucket, key, bytes.NewReader(data), int64(len(data)))
		}
	})

	b.Run("MetadataOperations", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			metadata := &storage.ObjectMetadata{
				Bucket:      bucket,
				Key:         fmt.Sprintf("metadata-bench-%d", i),
				Size:        1024,
				ContentType: "application/octet-stream",
				ETag:        fmt.Sprintf("etag-%d", i),
			}
			metadataStore.StoreObjectMetadata(context.Background(), metadata)
		}
	})
}
