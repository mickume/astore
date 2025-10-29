"""
Data models for Zot Artifact Store client
"""

from datetime import datetime
from typing import Dict, List, Optional, Any
from dataclasses import dataclass, field


@dataclass
class Bucket:
    """Bucket model"""

    name: str
    creation_date: datetime


@dataclass
class Object:
    """Object/Artifact model"""

    key: str
    size: int
    last_modified: datetime
    etag: str = ""
    content_type: str = ""
    metadata: Dict[str, str] = field(default_factory=dict)


@dataclass
class ListBucketsResult:
    """Result of listing buckets"""

    buckets: List[Bucket] = field(default_factory=list)


@dataclass
class ListObjectsResult:
    """Result of listing objects"""

    objects: List[Object] = field(default_factory=list)
    prefix: str = ""
    max_keys: int = 1000
    is_truncated: bool = False


@dataclass
class Signature:
    """Artifact signature model"""

    id: str
    artifact_digest: str
    signature: str
    algorithm: str
    signed_by: str
    timestamp: datetime


@dataclass
class SBOM:
    """Software Bill of Materials model"""

    id: str
    artifact_digest: str
    format: str
    content: str
    timestamp: datetime


@dataclass
class Attestation:
    """Artifact attestation model"""

    id: str
    artifact_digest: str
    type: str
    data: Dict[str, Any]
    timestamp: datetime


@dataclass
class VerificationResult:
    """Signature verification result"""

    valid: bool
    message: str
    signatures: List[Signature] = field(default_factory=list)


@dataclass
class CompletedPart:
    """Completed multipart upload part"""

    part_number: int
    etag: str


@dataclass
class MultipartUpload:
    """Multipart upload session"""

    upload_id: str
    bucket: str
    key: str
