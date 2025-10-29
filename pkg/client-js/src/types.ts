/**
 * Type definitions for Zot Artifact Store Client
 */

/**
 * Client configuration
 */
export interface Config {
  /** Base URL of the artifact store */
  baseURL: string;
  /** Authentication token (optional) */
  token?: string;
  /** Request timeout in milliseconds (default: 60000) */
  timeout?: number;
  /** Skip TLS certificate verification (default: false) */
  insecureSkipVerify?: boolean;
  /** Custom User-Agent header */
  userAgent?: string;
}

/**
 * Bucket metadata
 */
export interface Bucket {
  name: string;
  creationDate: Date;
}

/**
 * Object/Artifact metadata
 */
export interface ArtifactObject {
  key: string;
  size: number;
  lastModified: Date;
  etag?: string;
  contentType?: string;
  metadata?: Record<string, string>;
}

/**
 * List buckets result
 */
export interface ListBucketsResult {
  buckets: Bucket[];
}

/**
 * List objects result
 */
export interface ListObjectsResult {
  objects: ArtifactObject[];
  prefix?: string;
  maxKeys?: number;
  isTruncated?: boolean;
}

/**
 * Upload options
 */
export interface UploadOptions {
  contentType?: string;
  metadata?: Record<string, string>;
  progressCallback?: (bytesTransferred: number) => void;
}

/**
 * Download options
 */
export interface DownloadOptions {
  range?: string;
  progressCallback?: (bytesTransferred: number) => void;
}

/**
 * List options
 */
export interface ListOptions {
  prefix?: string;
  maxKeys?: number;
}

/**
 * Artifact signature
 */
export interface Signature {
  id: string;
  artifactDigest: string;
  signature: string;
  algorithm: string;
  signedBy: string;
  timestamp: Date;
}

/**
 * Software Bill of Materials (SBOM)
 */
export interface SBOM {
  id: string;
  artifactDigest: string;
  format: string;
  content: string;
  timestamp: Date;
}

/**
 * Artifact attestation
 */
export interface Attestation {
  id: string;
  artifactDigest: string;
  type: string;
  data: Record<string, any>;
  timestamp: Date;
}

/**
 * Signature verification result
 */
export interface VerificationResult {
  valid: boolean;
  message: string;
  signatures: Signature[];
}

/**
 * Multipart upload session
 */
export interface MultipartUpload {
  uploadId: string;
  bucket: string;
  key: string;
}

/**
 * Completed multipart upload part
 */
export interface CompletedPart {
  partNumber: number;
  etag: string;
}
