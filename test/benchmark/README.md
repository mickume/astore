# Performance Benchmarks

Comprehensive performance benchmarks for the Zot Artifact Store.

## Running Benchmarks

### Run All Benchmarks

```bash
go test -bench=. -benchmem ./test/benchmark/...
```

### Run Specific Benchmark

```bash
# Storage benchmarks only
go test -bench=BenchmarkMetadataStore -benchmem ./test/benchmark/

# API benchmarks only
go test -bench=BenchmarkS3APIHandlers -benchmem ./test/benchmark/

# Concurrent operations
go test -bench=BenchmarkConcurrent -benchmem ./test/benchmark/
```

### Run with CPU Profiling

```bash
go test -bench=. -cpuprofile=cpu.prof ./test/benchmark/
go tool pprof cpu.prof
```

### Run with Memory Profiling

```bash
go test -bench=. -memprofile=mem.prof -benchmem ./test/benchmark/
go tool pprof mem.prof
```

### Run with Custom Parameters

```bash
# Run for longer duration
go test -bench=. -benchtime=10s ./test/benchmark/

# Run with specific iteration count
go test -bench=. -benchtime=1000x ./test/benchmark/

# Save results for comparison
go test -bench=. -benchmem ./test/benchmark/ | tee bench-baseline.txt
```

## Benchmark Categories

### 1. Metadata Store Benchmarks

- **CreateBucket**: Bucket creation performance
- **StoreObjectMetadata**: Object metadata write performance
- **GetObjectMetadata**: Object metadata read performance
- **ListObjects**: Object listing performance with 100 objects
- **DeleteObjectMetadata**: Object metadata deletion performance

### 2. File Storage Benchmarks

Tests storage operations with various object sizes:
- 1 KB (small objects)
- 10 KB (medium objects)
- 100 KB (large objects)
- 1 MB (very large objects)
- 10 MB (extra large objects)

Operations tested:
- **WriteObject**: Write performance
- **ReadObject**: Read performance
- **DeleteObject**: Deletion performance

### 3. Concurrent Operations

Tests performance under concurrent load:
- 1, 2, 4, 8, 16, 32 concurrent operations
- **ConcurrentWrites**: Parallel write operations
- **ConcurrentReads**: Parallel read operations

### 4. Multipart Upload Benchmarks

- **InitiateMultipartUpload**: Multipart upload initialization
- **UploadPart**: Part upload performance (5 MB parts)
- **CompleteMultipartUpload**: Multipart completion (5 parts)

### 5. API Handler Benchmarks

HTTP endpoint performance testing:
- **CreateBucket**: Bucket creation via HTTP
- **PutObject**: Object upload via HTTP (1KB to 1MB)
- **GetObject**: Object download via HTTP
- **HeadObject**: Metadata retrieval
- **ListObjects**: Bucket listing (100 objects)
- **DeleteObject**: Object deletion

### 6. Concurrent API Requests

HTTP endpoint performance under load:
- 1, 2, 4, 8, 16, 32, 64 concurrent requests
- **ConcurrentGETs**: Parallel GET requests
- **ConcurrentPUTs**: Parallel PUT requests
- **ConcurrentHEADs**: Parallel HEAD requests

### 7. End-to-End Workflow

Complete artifact lifecycle:
1. Create bucket
2. Upload object
3. Get metadata
4. Download object
5. List objects
6. Delete object
7. Delete bucket

### 8. Memory Usage

Memory allocation patterns:
- **LargeNumberOfSmallObjects**: Many 1 KB objects
- **SmallNumberOfLargeObjects**: Few 10 MB objects
- **MetadataOperations**: Metadata storage patterns

## Interpreting Results

### Benchmark Output Format

```
BenchmarkMetadataStore/CreateBucket-8         50000    25000 ns/op    1024 B/op    10 allocs/op
```

- **50000**: Number of iterations run
- **25000 ns/op**: Nanoseconds per operation
- **1024 B/op**: Bytes allocated per operation
- **10 allocs/op**: Number of allocations per operation

### Performance Goals

#### Metadata Operations
- CreateBucket: < 100 µs/op
- StoreObjectMetadata: < 500 µs/op
- GetObjectMetadata: < 100 µs/op
- ListObjects (100 items): < 5 ms/op

#### File Storage Operations
- WriteObject (1 KB): < 1 ms/op
- WriteObject (1 MB): < 50 ms/op
- ReadObject (1 KB): < 500 µs/op
- ReadObject (1 MB): < 30 ms/op

#### API Handlers
- PutObject (1 KB): < 5 ms/op
- GetObject (1 KB): < 3 ms/op
- HeadObject: < 2 ms/op
- ListObjects: < 10 ms/op

#### Throughput Goals
- Concurrent Reads (8 threads): > 1000 ops/sec
- Concurrent Writes (8 threads): > 500 ops/sec
- End-to-End Workflow: > 100 workflows/sec

## Comparing Results

### Save Baseline

```bash
go test -bench=. -benchmem ./test/benchmark/ | tee bench-baseline.txt
```

### Run Comparison

```bash
# After making changes
go test -bench=. -benchmem ./test/benchmark/ | tee bench-new.txt

# Compare results
benchcmp bench-baseline.txt bench-new.txt
```

Or use `benchstat` for statistical comparison:

```bash
# Install benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# Compare results
benchstat bench-baseline.txt bench-new.txt
```

## CI/CD Integration

Benchmarks can be integrated into CI/CD pipelines:

```yaml
# GitHub Actions example
- name: Run Benchmarks
  run: |
    go test -bench=. -benchmem ./test/benchmark/ | tee bench-results.txt

- name: Check Performance Regression
  run: |
    benchstat bench-baseline.txt bench-results.txt
```

## Profiling

### CPU Profiling

```bash
# Generate CPU profile
go test -bench=BenchmarkS3APIHandlers -cpuprofile=cpu.prof ./test/benchmark/

# Analyze with pprof
go tool pprof cpu.prof

# Commands in pprof:
# - top10: Show top 10 functions by CPU time
# - list <function>: Show line-by-line profile
# - web: Generate SVG call graph
```

### Memory Profiling

```bash
# Generate memory profile
go test -bench=BenchmarkMemoryUsage -memprofile=mem.prof -benchmem ./test/benchmark/

# Analyze with pprof
go tool pprof -alloc_space mem.prof

# Commands in pprof:
# - top10: Show top 10 allocators
# - list <function>: Show allocations by line
# - web: Generate SVG visualization
```

### Block Profiling

```bash
# Generate block profile
go test -bench=BenchmarkConcurrent -blockprofile=block.prof ./test/benchmark/

# Analyze blocking operations
go tool pprof block.prof
```

## Performance Tuning Tips

### Storage Layer
1. Use larger buffer sizes for large objects
2. Enable OS-level page cache for frequently accessed objects
3. Consider SSD for metadata database
4. Tune BoltDB bucket size and fill percentage

### API Layer
1. Enable HTTP/2 for better multiplexing
2. Use connection pooling
3. Enable gzip compression for responses
4. Tune max concurrent requests

### Concurrency
1. Use worker pools for bounded concurrency
2. Tune GOMAXPROCS based on CPU count
3. Monitor goroutine count
4. Use sync.Pool for frequently allocated objects

### Database
1. Batch metadata operations when possible
2. Use transactions for related operations
3. Tune BoltDB page size
4. Enable mmap for read-heavy workloads

## Continuous Monitoring

Set up continuous benchmark tracking:

```bash
# Run benchmarks and store results with timestamp
DATE=$(date +%Y%m%d-%H%M%S)
go test -bench=. -benchmem ./test/benchmark/ > bench-$DATE.txt

# Track results over time
git add bench-$DATE.txt
git commit -m "Benchmark results $DATE"
```

## Troubleshooting

### Inconsistent Results

```bash
# Run benchmarks multiple times
go test -bench=. -benchmem -count=5 ./test/benchmark/

# Use benchstat for statistical analysis
benchstat bench-run1.txt bench-run2.txt bench-run3.txt
```

### High Memory Usage

```bash
# Run with memory profiling
go test -bench=BenchmarkMemoryUsage -memprofile=mem.prof -benchmem ./test/benchmark/

# Check for memory leaks
go tool pprof -alloc_space mem.prof
```

### CPU Bottlenecks

```bash
# Run with CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./test/benchmark/

# Identify hot spots
go tool pprof -top cpu.prof
```
