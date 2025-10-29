# Zot Artifact Store JavaScript/TypeScript Client

A TypeScript client library for interacting with the Zot Artifact Store, providing support for artifact management, supply chain security, and RBAC.

## Installation

```bash
npm install @astore/client
```

Or with Yarn:

```bash
yarn add @astore/client
```

## Quick Start

```typescript
import { Client } from '@astore/client';

// Create client
const client = new Client({
  baseURL: 'https://artifacts.example.com',
  token: 'your-bearer-token'
});

// Upload an artifact
const data = Buffer.from('artifact content');
await client.upload('releases', 'myapp-1.0.0.tar.gz', data, {
  contentType: 'application/gzip',
  metadata: { version: '1.0.0', env: 'production' }
});

// Download an artifact
const downloaded = await client.download('releases', 'myapp-1.0.0.tar.gz');

// List artifacts
const result = await client.listObjects('releases', { prefix: 'myapp/' });
for (const obj of result.objects) {
  console.log(`${obj.key} (${obj.size} bytes)`);
}
```

## Features

- **Artifact Management**: Upload, download, list, and delete artifacts
- **Bucket Operations**: Create, list, and delete buckets
- **Multipart Upload**: Support for large file uploads
- **Supply Chain Security**: Signing, verification, SBOM, and attestations
- **Progress Tracking**: Monitor upload/download progress
- **Custom Metadata**: Attach custom metadata to artifacts
- **Range Requests**: Download specific byte ranges
- **Authentication**: Bearer token authentication
- **TypeScript**: Full TypeScript support with type definitions
- **Error Handling**: Comprehensive exception hierarchy

## Configuration

Create a client with configuration:

```typescript
import { Client, Config } from '@astore/client';

const config: Config = {
  baseURL: 'https://artifacts.example.com',
  token: 'your-token',                    // Optional
  timeout: 60000,                         // Request timeout in ms (default: 60000)
  insecureSkipVerify: false,              // Skip TLS verification (default: false)
  userAgent: 'my-app/1.0'                // Custom User-Agent
};

const client = new Client(config);
```

### Environment Variables

You can also use environment variables:

```bash
export ASTORE_SERVER=https://artifacts.example.com
export ASTORE_TOKEN=your-token
```

```typescript
const client = new Client({
  baseURL: process.env.ASTORE_SERVER!,
  token: process.env.ASTORE_TOKEN
});
```

## Usage Examples

### Upload with Progress Tracking

```typescript
import * as fs from 'fs';

const data = fs.readFileSync('large-file.tar.gz');

await client.upload('releases', 'large-file.tar.gz', data, {
  progressCallback: (bytesTransferred) => {
    console.log(`Uploaded: ${bytesTransferred} bytes`);
  }
});
```

### Download with Range Request

```typescript
// Download first 1KB
const partial = await client.download('releases', 'myapp-1.0.0.tar.gz', {
  range: 'bytes=0-1023'
});
```

### Multipart Upload

```typescript
import * as fs from 'fs';

// For large files (>5MB recommended)
const fileData = fs.readFileSync('large-app.tar.gz');
const partSize = 5 * 1024 * 1024; // 5MB

const upload = await client.initiateMultipartUpload(
  'releases',
  'large-app.tar.gz',
  {
    contentType: 'application/gzip',
    metadata: { version: '2.0.0' }
  }
);

const parts: CompletedPart[] = [];
let partNumber = 1;

for (let start = 0; start < fileData.length; start += partSize) {
  const end = Math.min(start + partSize, fileData.length);
  const partData = fileData.slice(start, end);

  const etag = await client.uploadPart(
    upload.bucket,
    upload.key,
    upload.uploadId,
    partNumber,
    partData
  );

  parts.push({ partNumber, etag });
  partNumber++;
}

// Complete upload
await client.completeMultipartUpload(
  upload.bucket,
  upload.key,
  upload.uploadId,
  parts
);
```

### Supply Chain Operations

```typescript
import * as fs from 'fs';

// Sign an artifact
const privateKey = fs.readFileSync('private.pem', 'utf8');
const signature = await client.signArtifact('releases', 'myapp-1.0.0.tar.gz', privateKey);
console.log(`Signed: ${signature.id}`);

// Verify signatures
const publicKey = fs.readFileSync('public.pem', 'utf8');
const result = await client.verifySignatures(
  'releases',
  'myapp-1.0.0.tar.gz',
  [publicKey]
);

if (result.valid) {
  console.log('✓ Signature verification passed');
} else {
  console.log(`✗ Verification failed: ${result.message}`);
}

// Attach SBOM
const sbomContent = fs.readFileSync('sbom.json', 'utf8');
const sbom = await client.attachSBOM(
  'releases',
  'myapp-1.0.0.tar.gz',
  'spdx',
  sbomContent
);

// Add build attestation
const attestation = await client.addAttestation(
  'releases',
  'myapp-1.0.0.tar.gz',
  'build',
  {
    buildId: '12345',
    status: 'success',
    tests: 142,
    coverage: '85.3%'
  }
);
```

## Error Handling

The client provides a comprehensive exception hierarchy:

```typescript
import {
  NotFoundError,
  UnauthorizedError,
  ForbiddenError,
  ArtifactStoreError
} from '@astore/client';

try {
  await client.download('releases', 'nonexistent.tar.gz');
} catch (error) {
  if (error instanceof NotFoundError) {
    console.log('Artifact not found');
  } else if (error instanceof UnauthorizedError) {
    console.log('Authentication failed');
  } else if (error instanceof ForbiddenError) {
    console.log('Permission denied');
  } else if (error instanceof ArtifactStoreError) {
    console.log(`Error: ${error.message} (status: ${error.statusCode})`);
  }
}
```

## Testing

Run tests:

```bash
npm test
```

Run tests with coverage:

```bash
npm run test:coverage
```

## API Reference

### Client Configuration

- `new Client(config: Config)` - Create client instance

### Bucket Operations

- `client.createBucket(bucket: string): Promise<void>` - Create a bucket
- `client.deleteBucket(bucket: string): Promise<void>` - Delete a bucket
- `client.listBuckets(): Promise<ListBucketsResult>` - List all buckets

### Object Operations

- `client.upload(bucket, key, data, options?): Promise<void>` - Upload artifact
- `client.download(bucket, key, options?): Promise<Buffer>` - Download artifact
- `client.getObjectMetadata(bucket, key): Promise<ArtifactObject>` - Get artifact metadata
- `client.deleteObject(bucket, key): Promise<void>` - Delete artifact
- `client.listObjects(bucket, options?): Promise<ListObjectsResult>` - List artifacts
- `client.copyObject(sourceBucket, sourceKey, destBucket, destKey): Promise<void>` - Copy artifact

### Multipart Upload

- `client.initiateMultipartUpload(bucket, key, options?): Promise<MultipartUpload>` - Start multipart upload
- `client.uploadPart(bucket, key, uploadId, partNumber, data): Promise<string>` - Upload part
- `client.completeMultipartUpload(bucket, key, uploadId, parts): Promise<void>` - Complete upload
- `client.abortMultipartUpload(bucket, key, uploadId): Promise<void>` - Abort upload

### Supply Chain Operations

- `client.signArtifact(bucket, key, privateKey): Promise<Signature>` - Sign artifact
- `client.getSignatures(bucket, key): Promise<Signature[]>` - Get signatures
- `client.verifySignatures(bucket, key, publicKeys): Promise<VerificationResult>` - Verify signatures
- `client.attachSBOM(bucket, key, format, content): Promise<SBOM>` - Attach SBOM
- `client.getSBOM(bucket, key): Promise<SBOM>` - Get SBOM
- `client.addAttestation(bucket, key, type, data): Promise<Attestation>` - Add attestation
- `client.getAttestations(bucket, key): Promise<Attestation[]>` - Get attestations

## Requirements

- Node.js >= 14.0.0
- TypeScript >= 5.0.0 (for TypeScript projects)

## License

Apache-2.0

## See Also

- [Zot Artifact Store Documentation](../../docs/)
- [Go Client SDK](../client/)
- [Python Client SDK](../client-python/)
- [CLI Tool](../../cmd/astore-cli/)
