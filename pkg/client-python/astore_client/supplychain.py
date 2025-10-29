"""
Supply chain security operations for Zot Artifact Store Client
"""

from typing import Dict, List
from datetime import datetime
import json

from .models import Signature, SBOM, Attestation, VerificationResult


class SupplyChain:
    """Supply chain security operations mixin"""

    def __init__(self, client):
        self.client = client

    def sign_artifact(self, bucket: str, key: str, private_key: str) -> Signature:
        """Sign an artifact"""
        url = self.client._url(f"/supplychain/sign/{bucket}/{key}")

        payload = {"privateKey": private_key}

        response = self.client._request("POST", url, json=payload)

        data = response.json()
        return Signature(
            id=data["id"],
            artifact_digest=data["artifactDigest"],
            signature=data["signature"],
            algorithm=data["algorithm"],
            signed_by=data["signedBy"],
            timestamp=datetime.fromisoformat(data["timestamp"].replace("Z", "+00:00")),
        )

    def get_signatures(self, bucket: str, key: str) -> List[Signature]:
        """Get artifact signatures"""
        url = self.client._url(f"/supplychain/signatures/{bucket}/{key}")

        response = self.client._request("GET", url)

        data = response.json()
        signatures = []
        for sig in data.get("signatures", []):
            signatures.append(
                Signature(
                    id=sig["id"],
                    artifact_digest=sig["artifactDigest"],
                    signature=sig["signature"],
                    algorithm=sig["algorithm"],
                    signed_by=sig["signedBy"],
                    timestamp=datetime.fromisoformat(sig["timestamp"].replace("Z", "+00:00")),
                )
            )

        return signatures

    def verify_signatures(
        self, bucket: str, key: str, public_keys: List[str]
    ) -> VerificationResult:
        """Verify artifact signatures"""
        url = self.client._url(f"/supplychain/verify/{bucket}/{key}")

        payload = {"publicKeys": public_keys}

        response = self.client._request("POST", url, json=payload)

        data = response.json()

        signatures = []
        for sig in data.get("signatures", []):
            signatures.append(
                Signature(
                    id=sig["id"],
                    artifact_digest=sig["artifactDigest"],
                    signature=sig["signature"],
                    algorithm=sig["algorithm"],
                    signed_by=sig["signedBy"],
                    timestamp=datetime.fromisoformat(sig["timestamp"].replace("Z", "+00:00")),
                )
            )

        return VerificationResult(
            valid=data["valid"], message=data["message"], signatures=signatures
        )

    def attach_sbom(self, bucket: str, key: str, format: str, content: str) -> SBOM:
        """Attach SBOM to artifact"""
        url = self.client._url(f"/supplychain/sbom/{bucket}/{key}")

        payload = {"format": format, "content": content}

        response = self.client._request("POST", url, json=payload)

        data = response.json()
        return SBOM(
            id=data["id"],
            artifact_digest=data["artifactDigest"],
            format=data["format"],
            content=data["content"],
            timestamp=datetime.fromisoformat(data["timestamp"].replace("Z", "+00:00")),
        )

    def get_sbom(self, bucket: str, key: str) -> SBOM:
        """Get artifact SBOM"""
        url = self.client._url(f"/supplychain/sbom/{bucket}/{key}")

        response = self.client._request("GET", url)

        data = response.json()
        return SBOM(
            id=data["id"],
            artifact_digest=data["artifactDigest"],
            format=data["format"],
            content=data["content"],
            timestamp=datetime.fromisoformat(data["timestamp"].replace("Z", "+00:00")),
        )

    def add_attestation(
        self, bucket: str, key: str, attestation_type: str, data: Dict
    ) -> Attestation:
        """Add attestation to artifact"""
        url = self.client._url(f"/supplychain/attestations/{bucket}/{key}")

        payload = {"type": attestation_type, "data": data}

        response = self.client._request("POST", url, json=payload)

        resp_data = response.json()
        return Attestation(
            id=resp_data["id"],
            artifact_digest=resp_data["artifactDigest"],
            type=resp_data["type"],
            data=resp_data["data"],
            timestamp=datetime.fromisoformat(resp_data["timestamp"].replace("Z", "+00:00")),
        )

    def get_attestations(self, bucket: str, key: str) -> List[Attestation]:
        """Get artifact attestations"""
        url = self.client._url(f"/supplychain/attestations/{bucket}/{key}")

        response = self.client._request("GET", url)

        resp_data = response.json()
        attestations = []
        for att in resp_data.get("attestations", []):
            attestations.append(
                Attestation(
                    id=att["id"],
                    artifact_digest=att["artifactDigest"],
                    type=att["type"],
                    data=att["data"],
                    timestamp=datetime.fromisoformat(att["timestamp"].replace("Z", "+00:00")),
                )
            )

        return attestations
