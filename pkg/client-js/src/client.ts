/**
 * Zot Artifact Store Client
 */

import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import https from 'https';
import {
  Config,
  Bucket,
  ArtifactObject,
  ListBucketsResult,
  ListObjectsResult,
  UploadOptions,
  DownloadOptions,
  ListOptions,
  Signature,
  SBOM,
  Attestation,
  VerificationResult,
  MultipartUpload,
  CompletedPart,
} from './types';
import { raiseForStatus } from './exceptions';
import { Operations } from './operations';
import { SupplyChain } from './supplychain';

/**
 * Zot Artifact Store Client
 *
 * @example
 * ```typescript
 * const client = new Client({
 *   baseURL: 'https://artifacts.example.com',
 *   token: 'your-token'
 * });
 *
 * await client.upload('mybucket', 'myfile.tar.gz', buffer);
 * ```
 */
export class Client {
  private config: Config;
  private axiosInstance: AxiosInstance;
  private operations: Operations;
  private supplychain: SupplyChain;

  constructor(config: Config) {
    if (!config.baseURL) {
      throw new Error('baseURL is required');
    }

    this.config = {
      timeout: 60000,
      userAgent: '@astore/client/1.0.0',
      ...config,
      baseURL: config.baseURL.replace(/\/$/, ''), // Remove trailing slash
    };

    // Create axios instance
    const axiosConfig: AxiosRequestConfig = {
      baseURL: this.config.baseURL,
      timeout: this.config.timeout,
      headers: {
        'User-Agent': this.config.userAgent,
      },
    };

    // Add authentication if token provided
    if (this.config.token) {
      axiosConfig.headers = {
        ...axiosConfig.headers,
        Authorization: `Bearer ${this.config.token}`,
      };
    }

    // Configure TLS
    if (this.config.insecureSkipVerify) {
      axiosConfig.httpsAgent = new https.Agent({
        rejectUnauthorized: false,
      });
    }

    this.axiosInstance = axios.create(axiosConfig);

    // Add response interceptor for error handling
    this.axiosInstance.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response) {
          const message =
            error.response.data?.error ||
            error.response.data?.message ||
            error.message;
          raiseForStatus(error.response.status, message);
        }
        throw error;
      }
    );

    // Initialize operations
    this.operations = new Operations(this);
    this.supplychain = new SupplyChain(this);
  }

  /**
   * Update authentication token
   */
  setToken(token: string): void {
    this.config.token = token;
    this.axiosInstance.defaults.headers.common['Authorization'] = `Bearer ${token}`;
  }

  /**
   * Get axios instance for internal use
   * @internal
   */
  getAxiosInstance(): AxiosInstance {
    return this.axiosInstance;
  }

  // Bucket operations

  /**
   * Create a new bucket
   */
  async createBucket(bucket: string): Promise<void> {
    return this.operations.createBucket(bucket);
  }

  /**
   * Delete a bucket
   */
  async deleteBucket(bucket: string): Promise<void> {
    return this.operations.deleteBucket(bucket);
  }

  /**
   * List all buckets
   */
  async listBuckets(): Promise<ListBucketsResult> {
    return this.operations.listBuckets();
  }

  // Object operations

  /**
   * Upload an artifact
   */
  async upload(
    bucket: string,
    key: string,
    data: Buffer | string,
    options?: UploadOptions
  ): Promise<void> {
    return this.operations.upload(bucket, key, data, options);
  }

  /**
   * Download an artifact
   */
  async download(
    bucket: string,
    key: string,
    options?: DownloadOptions
  ): Promise<Buffer> {
    return this.operations.download(bucket, key, options);
  }

  /**
   * Get object metadata
   */
  async getObjectMetadata(bucket: string, key: string): Promise<ArtifactObject> {
    return this.operations.getObjectMetadata(bucket, key);
  }

  /**
   * Delete an object
   */
  async deleteObject(bucket: string, key: string): Promise<void> {
    return this.operations.deleteObject(bucket, key);
  }

  /**
   * List objects in a bucket
   */
  async listObjects(bucket: string, options?: ListOptions): Promise<ListObjectsResult> {
    return this.operations.listObjects(bucket, options);
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
    return this.operations.copyObject(sourceBucket, sourceKey, destBucket, destKey);
  }

  // Multipart upload operations

  /**
   * Initiate multipart upload
   */
  async initiateMultipartUpload(
    bucket: string,
    key: string,
    options?: UploadOptions
  ): Promise<MultipartUpload> {
    return this.operations.initiateMultipartUpload(bucket, key, options);
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
    return this.operations.uploadPart(bucket, key, uploadId, partNumber, data);
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
    return this.operations.completeMultipartUpload(bucket, key, uploadId, parts);
  }

  /**
   * Abort multipart upload
   */
  async abortMultipartUpload(
    bucket: string,
    key: string,
    uploadId: string
  ): Promise<void> {
    return this.operations.abortMultipartUpload(bucket, key, uploadId);
  }

  // Supply chain operations

  /**
   * Sign an artifact
   */
  async signArtifact(bucket: string, key: string, privateKey: string): Promise<Signature> {
    return this.supplychain.signArtifact(bucket, key, privateKey);
  }

  /**
   * Get artifact signatures
   */
  async getSignatures(bucket: string, key: string): Promise<Signature[]> {
    return this.supplychain.getSignatures(bucket, key);
  }

  /**
   * Verify artifact signatures
   */
  async verifySignatures(
    bucket: string,
    key: string,
    publicKeys: string[]
  ): Promise<VerificationResult> {
    return this.supplychain.verifySignatures(bucket, key, publicKeys);
  }

  /**
   * Attach SBOM to artifact
   */
  async attachSBOM(bucket: string, key: string, format: string, content: string): Promise<SBOM> {
    return this.supplychain.attachSBOM(bucket, key, format, content);
  }

  /**
   * Get artifact SBOM
   */
  async getSBOM(bucket: string, key: string): Promise<SBOM> {
    return this.supplychain.getSBOM(bucket, key);
  }

  /**
   * Add attestation to artifact
   */
  async addAttestation(
    bucket: string,
    key: string,
    type: string,
    data: Record<string, any>
  ): Promise<Attestation> {
    return this.supplychain.addAttestation(bucket, key, type, data);
  }

  /**
   * Get artifact attestations
   */
  async getAttestations(bucket: string, key: string): Promise<Attestation[]> {
    return this.supplychain.getAttestations(bucket, key);
  }
}
