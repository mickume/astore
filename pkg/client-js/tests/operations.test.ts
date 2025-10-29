/**
 * Tests for operations
 */

import { Client } from '../src/client';
import { Config } from '../src/types';
import axios from 'axios';

// Mock axios
jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

describe('Operations', () => {
  let client: Client;
  let mockAxiosInstance: any;

  beforeEach(() => {
    jest.clearAllMocks();

    mockAxiosInstance = {
      get: jest.fn(),
      post: jest.fn(),
      put: jest.fn(),
      delete: jest.fn(),
      head: jest.fn(),
      defaults: {
        headers: {
          common: {},
        },
      },
      interceptors: {
        response: {
          use: jest.fn(),
        },
      },
    };

    mockedAxios.create.mockReturnValue(mockAxiosInstance);

    const config: Config = {
      baseURL: 'https://test.example.com',
      token: 'test-token',
    };
    client = new Client(config);
  });

  describe('Bucket Operations', () => {
    it('should create bucket', async () => {
      // Given: Bucket name
      mockAxiosInstance.put.mockResolvedValue({ data: {} });

      // When: Creating bucket
      await client.createBucket('mybucket');

      // Then: Should call PUT /s3/mybucket
      expect(mockAxiosInstance.put).toHaveBeenCalledWith('/s3/mybucket');
    });

    it('should delete bucket', async () => {
      // Given: Existing bucket
      mockAxiosInstance.delete.mockResolvedValue({ data: {} });

      // When: Deleting bucket
      await client.deleteBucket('mybucket');

      // Then: Should call DELETE /s3/mybucket
      expect(mockAxiosInstance.delete).toHaveBeenCalledWith('/s3/mybucket');
    });

    it('should list buckets', async () => {
      // Given: Multiple buckets exist
      mockAxiosInstance.get.mockResolvedValue({
        data: {
          buckets: [
            { name: 'bucket1', creationDate: '2024-01-15T10:30:00Z' },
            { name: 'bucket2', creationDate: '2024-01-16T14:20:00Z' },
          ],
        },
      });

      // When: Listing buckets
      const result = await client.listBuckets();

      // Then: Should return all buckets
      expect(result.buckets).toHaveLength(2);
      expect(result.buckets[0].name).toBe('bucket1');
      expect(result.buckets[1].name).toBe('bucket2');
      expect(result.buckets[0].creationDate).toBeInstanceOf(Date);
    });
  });

  describe('Object Operations', () => {
    it('should upload object', async () => {
      // Given: Binary data to upload
      mockAxiosInstance.put.mockResolvedValue({ data: {} });
      const data = Buffer.from('test data');

      // When: Uploading artifact
      await client.upload('mybucket', 'myfile.tar.gz', data);

      // Then: Should call PUT with data
      expect(mockAxiosInstance.put).toHaveBeenCalledWith(
        '/s3/mybucket/myfile.tar.gz',
        data,
        expect.any(Object)
      );
    });

    it('should upload with metadata', async () => {
      // Given: Artifact with metadata
      mockAxiosInstance.put.mockResolvedValue({ data: {} });
      const data = Buffer.from('test data');

      // When: Uploading with metadata
      await client.upload('mybucket', 'myfile.tar.gz', data, {
        metadata: { version: '1.0.0', author: 'test' },
      });

      // Then: Should include metadata headers
      expect(mockAxiosInstance.put).toHaveBeenCalledWith(
        '/s3/mybucket/myfile.tar.gz',
        data,
        expect.objectContaining({
          headers: expect.objectContaining({
            'X-Amz-Meta-version': '1.0.0',
            'X-Amz-Meta-author': 'test',
          }),
        })
      );
    });

    it('should download object', async () => {
      // Given: Artifact exists
      mockAxiosInstance.get.mockResolvedValue({
        data: Buffer.from('downloaded data'),
      });

      // When: Downloading artifact
      const result = await client.download('mybucket', 'myfile.tar.gz');

      // Then: Should return data
      expect(result).toBeInstanceOf(Buffer);
      expect(result.toString()).toBe('downloaded data');
    });

    it('should download with range', async () => {
      // Given: Artifact exists
      mockAxiosInstance.get.mockResolvedValue({
        data: Buffer.from('partial'),
      });

      // When: Downloading with range
      await client.download('mybucket', 'myfile.tar.gz', {
        range: 'bytes=0-6',
      });

      // Then: Should include Range header
      expect(mockAxiosInstance.get).toHaveBeenCalledWith(
        '/s3/mybucket/myfile.tar.gz',
        expect.objectContaining({
          headers: expect.objectContaining({
            Range: 'bytes=0-6',
          }),
        })
      );
    });

    it('should get object metadata', async () => {
      // Given: Artifact exists
      mockAxiosInstance.head.mockResolvedValue({
        headers: {
          'content-length': '1024',
          'content-type': 'application/gzip',
          etag: '"abc123"',
          'last-modified': 'Mon, 15 Jan 2024 10:30:00 GMT',
          'x-amz-meta-version': '1.0.0',
        },
      });

      // When: Getting metadata
      const obj = await client.getObjectMetadata('mybucket', 'myfile.tar.gz');

      // Then: Should return metadata
      expect(obj.key).toBe('myfile.tar.gz');
      expect(obj.size).toBe(1024);
      expect(obj.contentType).toBe('application/gzip');
      expect(obj.etag).toBe('abc123');
      expect(obj.metadata?.version).toBe('1.0.0');
    });

    it('should delete object', async () => {
      // Given: Artifact exists
      mockAxiosInstance.delete.mockResolvedValue({ data: {} });

      // When: Deleting artifact
      await client.deleteObject('mybucket', 'myfile.tar.gz');

      // Then: Should call DELETE
      expect(mockAxiosInstance.delete).toHaveBeenCalledWith('/s3/mybucket/myfile.tar.gz');
    });

    it('should list objects', async () => {
      // Given: Objects in bucket
      mockAxiosInstance.get.mockResolvedValue({
        data: {
          contents: [
            {
              key: 'file1.tar.gz',
              size: 1024,
              lastModified: '2024-01-15T10:30:00Z',
              etag: 'abc',
              contentType: 'application/gzip',
            },
            {
              key: 'file2.tar.gz',
              size: 2048,
              lastModified: '2024-01-16T14:20:00Z',
              etag: 'def',
              contentType: 'application/gzip',
            },
          ],
          prefix: '',
          maxKeys: 1000,
          isTruncated: false,
        },
      });

      // When: Listing objects
      const result = await client.listObjects('mybucket');

      // Then: Should return objects
      expect(result.objects).toHaveLength(2);
      expect(result.objects[0].key).toBe('file1.tar.gz');
      expect(result.objects[0].size).toBe(1024);
      expect(result.objects[1].key).toBe('file2.tar.gz');
    });

    it('should list objects with prefix', async () => {
      // Given: Objects with prefix
      mockAxiosInstance.get.mockResolvedValue({
        data: {
          contents: [],
          prefix: 'app/',
          maxKeys: 1000,
          isTruncated: false,
        },
      });

      // When: Listing with prefix
      const result = await client.listObjects('mybucket', { prefix: 'app/' });

      // Then: Should include prefix in params
      expect(mockAxiosInstance.get).toHaveBeenCalledWith(
        '/s3/mybucket',
        expect.objectContaining({
          params: expect.objectContaining({
            prefix: 'app/',
          }),
        })
      );
      expect(result.prefix).toBe('app/');
    });

    it('should copy object', async () => {
      // Given: Source artifact exists
      mockAxiosInstance.put.mockResolvedValue({ data: {} });

      // When: Copying artifact
      await client.copyObject('srcbucket', 'srcfile.tar.gz', 'destbucket', 'destfile.tar.gz');

      // Then: Should include copy source header
      expect(mockAxiosInstance.put).toHaveBeenCalledWith(
        '/s3/destbucket/destfile.tar.gz',
        null,
        expect.objectContaining({
          headers: expect.objectContaining({
            'X-Amz-Copy-Source': '/srcbucket/srcfile.tar.gz',
          }),
        })
      );
    });
  });

  describe('Multipart Upload', () => {
    it('should initiate multipart upload', async () => {
      // Given: Large file to upload
      mockAxiosInstance.post.mockResolvedValue({
        data: { uploadId: 'upload123' },
      });

      // When: Initiating multipart upload
      const upload = await client.initiateMultipartUpload('mybucket', 'largefile.tar.gz');

      // Then: Should return upload ID
      expect(upload.uploadId).toBe('upload123');
      expect(upload.bucket).toBe('mybucket');
      expect(upload.key).toBe('largefile.tar.gz');
    });

    it('should upload part', async () => {
      // Given: Initiated multipart upload
      mockAxiosInstance.put.mockResolvedValue({
        headers: { etag: '"part-etag-1"' },
      });

      const data = Buffer.from('part data');

      // When: Uploading part
      const etag = await client.uploadPart('mybucket', 'largefile.tar.gz', 'upload123', 1, data);

      // Then: Should return ETag
      expect(etag).toBe('part-etag-1');
    });

    it('should complete multipart upload', async () => {
      // Given: All parts uploaded
      mockAxiosInstance.post.mockResolvedValue({ data: {} });

      const parts = [
        { partNumber: 1, etag: 'etag1' },
        { partNumber: 2, etag: 'etag2' },
      ];

      // When: Completing upload
      await client.completeMultipartUpload('mybucket', 'largefile.tar.gz', 'upload123', parts);

      // Then: Should send parts list
      expect(mockAxiosInstance.post).toHaveBeenCalledWith(
        '/s3/mybucket/largefile.tar.gz',
        expect.objectContaining({
          parts: expect.arrayContaining([
            { partNumber: 1, etag: 'etag1' },
            { partNumber: 2, etag: 'etag2' },
          ]),
        }),
        expect.objectContaining({
          params: { uploadId: 'upload123' },
        })
      );
    });

    it('should abort multipart upload', async () => {
      // Given: Initiated multipart upload
      mockAxiosInstance.delete.mockResolvedValue({ data: {} });

      // When: Aborting upload
      await client.abortMultipartUpload('mybucket', 'largefile.tar.gz', 'upload123');

      // Then: Should call DELETE
      expect(mockAxiosInstance.delete).toHaveBeenCalledWith(
        '/s3/mybucket/largefile.tar.gz',
        expect.objectContaining({
          params: { uploadId: 'upload123' },
        })
      );
    });
  });
});
