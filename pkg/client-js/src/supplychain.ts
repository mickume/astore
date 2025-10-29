/**
 * Supply chain security operations
 */

import { Signature, SBOM, Attestation, VerificationResult } from './types';
import { Client } from './client';

/**
 * Supply chain operations handler
 */
export class SupplyChain {
  constructor(private client: Client) {}

  private get axios() {
    return this.client.getAxiosInstance();
  }

  /**
   * Sign an artifact
   */
  async signArtifact(bucket: string, key: string, privateKey: string): Promise<Signature> {
    const response = await this.axios.post(`/supplychain/sign/${bucket}/${key}`, {
      privateKey,
    });

    return {
      id: response.data.id,
      artifactDigest: response.data.artifactDigest,
      signature: response.data.signature,
      algorithm: response.data.algorithm,
      signedBy: response.data.signedBy,
      timestamp: new Date(response.data.timestamp),
    };
  }

  /**
   * Get artifact signatures
   */
  async getSignatures(bucket: string, key: string): Promise<Signature[]> {
    const response = await this.axios.get(`/supplychain/signatures/${bucket}/${key}`);

    return (response.data.signatures || []).map((sig: any) => ({
      id: sig.id,
      artifactDigest: sig.artifactDigest,
      signature: sig.signature,
      algorithm: sig.algorithm,
      signedBy: sig.signedBy,
      timestamp: new Date(sig.timestamp),
    }));
  }

  /**
   * Verify artifact signatures
   */
  async verifySignatures(
    bucket: string,
    key: string,
    publicKeys: string[]
  ): Promise<VerificationResult> {
    const response = await this.axios.post(`/supplychain/verify/${bucket}/${key}`, {
      publicKeys,
    });

    const signatures = (response.data.signatures || []).map((sig: any) => ({
      id: sig.id,
      artifactDigest: sig.artifactDigest,
      signature: sig.signature,
      algorithm: sig.algorithm,
      signedBy: sig.signedBy,
      timestamp: new Date(sig.timestamp),
    }));

    return {
      valid: response.data.valid,
      message: response.data.message,
      signatures,
    };
  }

  /**
   * Attach SBOM to artifact
   */
  async attachSBOM(bucket: string, key: string, format: string, content: string): Promise<SBOM> {
    const response = await this.axios.post(`/supplychain/sbom/${bucket}/${key}`, {
      format,
      content,
    });

    return {
      id: response.data.id,
      artifactDigest: response.data.artifactDigest,
      format: response.data.format,
      content: response.data.content,
      timestamp: new Date(response.data.timestamp),
    };
  }

  /**
   * Get artifact SBOM
   */
  async getSBOM(bucket: string, key: string): Promise<SBOM> {
    const response = await this.axios.get(`/supplychain/sbom/${bucket}/${key}`);

    return {
      id: response.data.id,
      artifactDigest: response.data.artifactDigest,
      format: response.data.format,
      content: response.data.content,
      timestamp: new Date(response.data.timestamp),
    };
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
    const response = await this.axios.post(`/supplychain/attestations/${bucket}/${key}`, {
      type,
      data,
    });

    return {
      id: response.data.id,
      artifactDigest: response.data.artifactDigest,
      type: response.data.type,
      data: response.data.data,
      timestamp: new Date(response.data.timestamp),
    };
  }

  /**
   * Get artifact attestations
   */
  async getAttestations(bucket: string, key: string): Promise<Attestation[]> {
    const response = await this.axios.get(`/supplychain/attestations/${bucket}/${key}`);

    return (response.data.attestations || []).map((att: any) => ({
      id: att.id,
      artifactDigest: att.artifactDigest,
      type: att.type,
      data: att.data,
      timestamp: new Date(att.timestamp),
    }));
  }
}
