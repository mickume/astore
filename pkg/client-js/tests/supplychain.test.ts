/**
 * Tests for supply chain operations
 */

import { Client } from '../src/client';
import { Config } from '../src/types';
import axios from 'axios';

// Mock axios
jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

describe('Supply Chain Operations', () => {
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

  it('should sign artifact', async () => {
    // Given: Artifact and private key
    mockAxiosInstance.post.mockResolvedValue({
      data: {
        id: 'sig123',
        artifactDigest: 'sha256:abc123',
        signature: 'signature-data',
        algorithm: 'RSA-SHA256',
        signedBy: 'test-signer',
        timestamp: '2024-01-15T10:30:00Z',
      },
    });

    // When: Signing artifact
    const signature = await client.signArtifact('mybucket', 'myfile.tar.gz', 'private-key-pem');

    // Then: Should return signature
    expect(signature.id).toBe('sig123');
    expect(signature.artifactDigest).toBe('sha256:abc123');
    expect(signature.algorithm).toBe('RSA-SHA256');
    expect(signature.signedBy).toBe('test-signer');
    expect(signature.timestamp).toBeInstanceOf(Date);
  });

  it('should get signatures', async () => {
    // Given: Signed artifact
    mockAxiosInstance.get.mockResolvedValue({
      data: {
        signatures: [
          {
            id: 'sig1',
            artifactDigest: 'sha256:abc123',
            signature: 'sig-data-1',
            algorithm: 'RSA-SHA256',
            signedBy: 'signer1',
            timestamp: '2024-01-15T10:30:00Z',
          },
          {
            id: 'sig2',
            artifactDigest: 'sha256:abc123',
            signature: 'sig-data-2',
            algorithm: 'RSA-SHA256',
            signedBy: 'signer2',
            timestamp: '2024-01-15T11:00:00Z',
          },
        ],
      },
    });

    // When: Getting signatures
    const signatures = await client.getSignatures('mybucket', 'myfile.tar.gz');

    // Then: Should return list of signatures
    expect(signatures).toHaveLength(2);
    expect(signatures[0].id).toBe('sig1');
    expect(signatures[1].id).toBe('sig2');
  });

  it('should verify signatures', async () => {
    // Given: Signed artifact and public key
    mockAxiosInstance.post.mockResolvedValue({
      data: {
        valid: true,
        message: 'All signatures valid',
        signatures: [
          {
            id: 'sig1',
            artifactDigest: 'sha256:abc123',
            signature: 'sig-data',
            algorithm: 'RSA-SHA256',
            signedBy: 'signer1',
            timestamp: '2024-01-15T10:30:00Z',
          },
        ],
      },
    });

    // When: Verifying signatures
    const result = await client.verifySignatures('mybucket', 'myfile.tar.gz', ['public-key-pem']);

    // Then: Should return verification result
    expect(result.valid).toBe(true);
    expect(result.message).toBe('All signatures valid');
    expect(result.signatures).toHaveLength(1);
  });

  it('should verify signatures failure', async () => {
    // Given: Artifact with invalid signature
    mockAxiosInstance.post.mockResolvedValue({
      data: {
        valid: false,
        message: 'Signature verification failed',
        signatures: [],
      },
    });

    // When: Verifying signatures
    const result = await client.verifySignatures('mybucket', 'myfile.tar.gz', ['public-key-pem']);

    // Then: Should return failed result
    expect(result.valid).toBe(false);
    expect(result.message).toContain('failed');
  });

  it('should attach SBOM', async () => {
    // Given: Artifact and SBOM content
    mockAxiosInstance.post.mockResolvedValue({
      data: {
        id: 'sbom123',
        artifactDigest: 'sha256:abc123',
        format: 'spdx',
        content: '{"spdxVersion": "2.3"}',
        timestamp: '2024-01-15T10:30:00Z',
      },
    });

    // When: Attaching SBOM
    const sbom = await client.attachSBOM(
      'mybucket',
      'myfile.tar.gz',
      'spdx',
      '{"spdxVersion": "2.3"}'
    );

    // Then: Should return SBOM
    expect(sbom.id).toBe('sbom123');
    expect(sbom.format).toBe('spdx');
    expect(sbom.artifactDigest).toBe('sha256:abc123');
    expect(sbom.timestamp).toBeInstanceOf(Date);
  });

  it('should get SBOM', async () => {
    // Given: Artifact with SBOM
    mockAxiosInstance.get.mockResolvedValue({
      data: {
        id: 'sbom123',
        artifactDigest: 'sha256:abc123',
        format: 'spdx',
        content: '{"spdxVersion": "2.3", "packages": []}',
        timestamp: '2024-01-15T10:30:00Z',
      },
    });

    // When: Getting SBOM
    const sbom = await client.getSBOM('mybucket', 'myfile.tar.gz');

    // Then: Should return SBOM
    expect(sbom.id).toBe('sbom123');
    expect(sbom.format).toBe('spdx');
    expect(sbom.content).toContain('spdxVersion');
  });

  it('should add attestation', async () => {
    // Given: Artifact and attestation data
    mockAxiosInstance.post.mockResolvedValue({
      data: {
        id: 'att123',
        artifactDigest: 'sha256:abc123',
        type: 'build',
        data: {
          buildId: '12345',
          status: 'success',
          tests: 142,
        },
        timestamp: '2024-01-15T10:30:00Z',
      },
    });

    // When: Adding attestation
    const attestation = await client.addAttestation('mybucket', 'myfile.tar.gz', 'build', {
      buildId: '12345',
      status: 'success',
      tests: 142,
    });

    // Then: Should return attestation
    expect(attestation.id).toBe('att123');
    expect(attestation.type).toBe('build');
    expect(attestation.data.buildId).toBe('12345');
    expect(attestation.data.status).toBe('success');
  });

  it('should get attestations', async () => {
    // Given: Artifact with attestations
    mockAxiosInstance.get.mockResolvedValue({
      data: {
        attestations: [
          {
            id: 'att1',
            artifactDigest: 'sha256:abc123',
            type: 'build',
            data: { buildId: '123', status: 'success' },
            timestamp: '2024-01-15T10:30:00Z',
          },
          {
            id: 'att2',
            artifactDigest: 'sha256:abc123',
            type: 'test',
            data: { testsPassed: 142, testsFailed: 0 },
            timestamp: '2024-01-15T11:00:00Z',
          },
        ],
      },
    });

    // When: Getting attestations
    const attestations = await client.getAttestations('mybucket', 'myfile.tar.gz');

    // Then: Should return attestations
    expect(attestations).toHaveLength(2);
    expect(attestations[0].type).toBe('build');
    expect(attestations[1].type).toBe('test');
    expect(attestations[0].data.buildId).toBe('123');
    expect(attestations[1].data.testsPassed).toBe(142);
  });
});
