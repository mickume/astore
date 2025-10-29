"""
Tests for supply chain security operations
"""

import pytest
import responses
from datetime import datetime


class TestSupplyChainOperations:
    """Test supply chain security operations"""

    def test_sign_artifact(self, client, mock_responses):
        """Given: Artifact and private key
        When: Signing artifact
        Then: Should return signature"""
        mock_responses.add(
            responses.POST,
            "https://test.example.com/supplychain/sign/mybucket/myfile.tar.gz",
            json={
                "id": "sig123",
                "artifactDigest": "sha256:abc123",
                "signature": "signature-data",
                "algorithm": "RSA-SHA256",
                "signedBy": "test-signer",
                "timestamp": "2024-01-15T10:30:00Z",
            },
            status=200,
        )

        signature = client.sign_artifact(
            "mybucket", "myfile.tar.gz", "private-key-pem"
        )

        assert signature.id == "sig123"
        assert signature.artifact_digest == "sha256:abc123"
        assert signature.algorithm == "RSA-SHA256"
        assert signature.signed_by == "test-signer"

    def test_get_signatures(self, client, mock_responses):
        """Given: Signed artifact
        When: Getting signatures
        Then: Should return list of signatures"""
        mock_responses.add(
            responses.GET,
            "https://test.example.com/supplychain/signatures/mybucket/myfile.tar.gz",
            json={
                "signatures": [
                    {
                        "id": "sig1",
                        "artifactDigest": "sha256:abc123",
                        "signature": "sig-data-1",
                        "algorithm": "RSA-SHA256",
                        "signedBy": "signer1",
                        "timestamp": "2024-01-15T10:30:00Z",
                    },
                    {
                        "id": "sig2",
                        "artifactDigest": "sha256:abc123",
                        "signature": "sig-data-2",
                        "algorithm": "RSA-SHA256",
                        "signedBy": "signer2",
                        "timestamp": "2024-01-15T11:00:00Z",
                    },
                ]
            },
            status=200,
        )

        signatures = client.get_signatures("mybucket", "myfile.tar.gz")

        assert len(signatures) == 2
        assert signatures[0].id == "sig1"
        assert signatures[1].id == "sig2"

    def test_verify_signatures(self, client, mock_responses):
        """Given: Signed artifact and public key
        When: Verifying signatures
        Then: Should return verification result"""
        mock_responses.add(
            responses.POST,
            "https://test.example.com/supplychain/verify/mybucket/myfile.tar.gz",
            json={
                "valid": True,
                "message": "All signatures valid",
                "signatures": [
                    {
                        "id": "sig1",
                        "artifactDigest": "sha256:abc123",
                        "signature": "sig-data",
                        "algorithm": "RSA-SHA256",
                        "signedBy": "signer1",
                        "timestamp": "2024-01-15T10:30:00Z",
                    }
                ],
            },
            status=200,
        )

        result = client.verify_signatures(
            "mybucket", "myfile.tar.gz", ["public-key-pem"]
        )

        assert result.valid is True
        assert result.message == "All signatures valid"
        assert len(result.signatures) == 1

    def test_verify_signatures_failure(self, client, mock_responses):
        """Given: Artifact with invalid signature
        When: Verifying signatures
        Then: Should return failed verification result"""
        mock_responses.add(
            responses.POST,
            "https://test.example.com/supplychain/verify/mybucket/myfile.tar.gz",
            json={
                "valid": False,
                "message": "Signature verification failed",
                "signatures": [],
            },
            status=200,
        )

        result = client.verify_signatures(
            "mybucket", "myfile.tar.gz", ["public-key-pem"]
        )

        assert result.valid is False
        assert "failed" in result.message.lower()

    def test_attach_sbom(self, client, mock_responses):
        """Given: Artifact and SBOM content
        When: Attaching SBOM
        Then: Should attach SBOM and return SBOM object"""
        mock_responses.add(
            responses.POST,
            "https://test.example.com/supplychain/sbom/mybucket/myfile.tar.gz",
            json={
                "id": "sbom123",
                "artifactDigest": "sha256:abc123",
                "format": "spdx",
                "content": '{"spdxVersion": "2.3"}',
                "timestamp": "2024-01-15T10:30:00Z",
            },
            status=200,
        )

        sbom = client.attach_sbom(
            "mybucket", "myfile.tar.gz", "spdx", '{"spdxVersion": "2.3"}'
        )

        assert sbom.id == "sbom123"
        assert sbom.format == "spdx"
        assert sbom.artifact_digest == "sha256:abc123"

    def test_get_sbom(self, client, mock_responses):
        """Given: Artifact with SBOM
        When: Getting SBOM
        Then: Should return SBOM content"""
        mock_responses.add(
            responses.GET,
            "https://test.example.com/supplychain/sbom/mybucket/myfile.tar.gz",
            json={
                "id": "sbom123",
                "artifactDigest": "sha256:abc123",
                "format": "spdx",
                "content": '{"spdxVersion": "2.3", "packages": []}',
                "timestamp": "2024-01-15T10:30:00Z",
            },
            status=200,
        )

        sbom = client.get_sbom("mybucket", "myfile.tar.gz")

        assert sbom.id == "sbom123"
        assert sbom.format == "spdx"
        assert "spdxVersion" in sbom.content

    def test_add_attestation(self, client, mock_responses):
        """Given: Artifact and attestation data
        When: Adding attestation
        Then: Should attach attestation and return attestation object"""
        mock_responses.add(
            responses.POST,
            "https://test.example.com/supplychain/attestations/mybucket/myfile.tar.gz",
            json={
                "id": "att123",
                "artifactDigest": "sha256:abc123",
                "type": "build",
                "data": {
                    "buildId": "12345",
                    "status": "success",
                    "tests": 142,
                },
                "timestamp": "2024-01-15T10:30:00Z",
            },
            status=200,
        )

        attestation = client.add_attestation(
            "mybucket",
            "myfile.tar.gz",
            "build",
            {"buildId": "12345", "status": "success", "tests": 142},
        )

        assert attestation.id == "att123"
        assert attestation.type == "build"
        assert attestation.data["buildId"] == "12345"
        assert attestation.data["status"] == "success"

    def test_get_attestations(self, client, mock_responses):
        """Given: Artifact with attestations
        When: Getting attestations
        Then: Should return list of attestations"""
        mock_responses.add(
            responses.GET,
            "https://test.example.com/supplychain/attestations/mybucket/myfile.tar.gz",
            json={
                "attestations": [
                    {
                        "id": "att1",
                        "artifactDigest": "sha256:abc123",
                        "type": "build",
                        "data": {"buildId": "123", "status": "success"},
                        "timestamp": "2024-01-15T10:30:00Z",
                    },
                    {
                        "id": "att2",
                        "artifactDigest": "sha256:abc123",
                        "type": "test",
                        "data": {"testsPassed": 142, "testsFailed": 0},
                        "timestamp": "2024-01-15T11:00:00Z",
                    },
                ]
            },
            status=200,
        )

        attestations = client.get_attestations("mybucket", "myfile.tar.gz")

        assert len(attestations) == 2
        assert attestations[0].type == "build"
        assert attestations[1].type == "test"
        assert attestations[0].data["buildId"] == "123"
        assert attestations[1].data["testsPassed"] == 142
