/**
 * Core artifact operations
 */

import { AxiosProgressEvent } from 'axios';
import {
  Bucket,
  ArtifactObject,
  ListBucketsResult,
  ListObjectsResult,
  UploadOptions,
  DownloadOptions,
  ListOptions,
  MultipartUpload,
  CompletedPart,
} from './types';
import { Client } from './client';

/**
 * Core operations handler
 */
export class Operations {
  constructor(private client: Client) {}

  private get axios() {
    return this.client.getAxiosInstance();
  }

  /**
   * Create a new bucket
   */
  async createBucket(bucket: string): Promise<void> {
    await this.axios.put(`/s3/${bucket}`);
  }

  /**
   * Delete a bucket
   */
  async deleteBucket(bucket: string): Promise<void> {
    await this.axios.delete(`/s3/${bucket}`);
  }

  /**
   * List all buckets
   */
  async listBuckets(): Promise<ListBucketsResult> {
    const response = await this.axios.get('/s3');
    const buckets: Bucket[] = (response.data.buckets || []).map((b: any) => ({
      name: b.name,
      creationDate: new Date(b.creationDate),
    }));
    return { buckets };
  }

  /**
   * Upload an artifact
   */
  async upload(
    bucket: string,
    key: string,
    data: Buffer | string,
    options?: UploadOptions
  ): Promise<void> {
    const headers: Record<string, string> = {
      'Content-Type': options?.contentType || 'application/octet-stream',
    };

    // Add metadata headers
    if (options?.metadata) {
      Object.entries(options.metadata).forEach(([k, v]) => {
        headers[`X-Amz-Meta-${k}`] = v;
      });
    }

    const config: any = { headers };

    // Add progress tracking
    if (options?.progressCallback) {
      config.onUploadProgress = (progressEvent: AxiosProgressEvent) => {
        if (progressEvent.loaded && options.progressCallback) {
          options.progressCallback(progressEvent.loaded);
        }
      };
    }

    await this.axios.put(`/s3/${bucket}/${key}`, data, config);
  }

  /**
   * Download an artifact
   */
  async download(
    bucket: string,
    key: string,
    options?: DownloadOptions
  ): Promise<Buffer> {
    const headers: Record<string, string> = {};

    if (options?.range) {
      headers['Range'] = options.range;
    }

    const config: any = {
      headers,
      responseType: 'arraybuffer',
    };

    // Add progress tracking
    if (options?.progressCallback) {
      config.onDownloadProgress = (progressEvent: AxiosProgressEvent) => {
        if (progressEvent.loaded && options.progressCallback) {
          options.progressCallback(progressEvent.loaded);
        }
      };
    }

    const response = await this.axios.get(`/s3/${bucket}/${key}`, config);
    return Buffer.from(response.data);
  }

  /**
   * Get object metadata
   */
  async getObjectMetadata(bucket: string, key: string): Promise<ArtifactObject> {
    const response = await this.axios.head(`/s3/${bucket}/${key}`);

    // Extract metadata from headers
    const metadata: Record<string, string> = {};
    Object.entries(response.headers).forEach(([headerName, headerValue]) => {
      if (headerName.toLowerCase().startsWith('x-amz-meta-')) {
        const keyName = headerName.substring(11); // Remove "x-amz-meta-" prefix
        metadata[keyName] = String(headerValue);
      }
    });

    const size = parseInt(response.headers['content-length'] || '0', 10);
    const lastModifiedStr = response.headers['last-modified'];
    const lastModified = lastModifiedStr ? new Date(lastModifiedStr) : new Date();

    return {
      key,
      size,
      lastModified,
      etag: response.headers['etag']?.replace(/"/g, ''),
      contentType: response.headers['content-type'],
      metadata,
    };
  }

  /**
   * Delete an object
   */
  async deleteObject(bucket: string, key: string): Promise<void> {
    await this.axios.delete(`/s3/${bucket}/${key}`);
  }

  /**
   * List objects in a bucket
   */
  async listObjects(bucket: string, options?: ListOptions): Promise<ListObjectsResult> {
    const params: Record<string, any> = {};

    if (options?.prefix) {
      params.prefix = options.prefix;
    }
    if (options?.maxKeys) {
      params['max-keys'] = options.maxKeys;
    }

    const response = await this.axios.get(`/s3/${bucket}`, { params });

    const objects: ArtifactObject[] = (response.data.contents || []).map((obj: any) => ({
      key: obj.key,
      size: obj.size || 0,
      lastModified: obj.lastModified ? new Date(obj.lastModified) : new Date(),
      etag: obj.etag || '',
      contentType: obj.contentType || '',
    }));

    return {
      objects,
      prefix: response.data.prefix || '',
      maxKeys: response.data.maxKeys || 1000,
      isTruncated: response.data.isTruncated || false,
    };
  }

  /**
   * Copy an object
   */
  async copyObject(
    sourceBucket: string,
    sourceKey: string,
    destBucket: string,
    destKey: string
  ): Promise<void> {
    const headers = {
      'X-Amz-Copy-Source': `/${sourceBucket}/${sourceKey}`,
    };

    await this.axios.put(`/s3/${destBucket}/${destKey}`, null, { headers });
  }

  /**
   * Initiate multipart upload
   */
  async initiateMultipartUpload(
    bucket: string,
    key: string,
    options?: UploadOptions
  ): Promise<MultipartUpload> {
    const headers: Record<string, string> = {
      'Content-Type': options?.contentType || 'application/octet-stream',
    };

    // Add metadata headers
    if (options?.metadata) {
      Object.entries(options.metadata).forEach(([k, v]) => {
        headers[`X-Amz-Meta-${k}`] = v;
      });
    }

    const response = await this.axios.post(`/s3/${bucket}/${key}`, null, {
      headers,
      params: { uploads: '' },
    });

    return {
      uploadId: response.data.uploadId,
      bucket,
      key,
    };
  }

  /**
   * Upload a part in multipart upload
   */
  async uploadPart(
    bucket: string,
    key: string,
    uploadId: string,
    partNumber: number,
    data: Buffer
  ): Promise<string> {
    const response = await this.axios.put(`/s3/${bucket}/${key}`, data, {
      params: {
        uploadId,
        partNumber,
      },
    });

    return response.headers['etag']?.replace(/"/g, '') || '';
  }

  /**
   * Complete multipart upload
   */
  async completeMultipartUpload(
    bucket: string,
    key: string,
    uploadId: string,
    parts: CompletedPart[]
  ): Promise<void> {
    const partsData = {
      parts: parts.map((p) => ({
        partNumber: p.partNumber,
        etag: p.etag,
      })),
    };

    await this.axios.post(`/s3/${bucket}/${key}`, partsData, {
      params: { uploadId },
    });
  }

  /**
   * Abort multipart upload
   */
  async abortMultipartUpload(bucket: string, key: string, uploadId: string): Promise<void> {
    await this.axios.delete(`/s3/${bucket}/${key}`, {
      params: { uploadId },
    });
  }
}
