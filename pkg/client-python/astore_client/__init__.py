"""
Zot Artifact Store Python Client SDK

A Python client library for interacting with the Zot Artifact Store.
Provides support for artifact management, supply chain security, and RBAC.
"""

from .client import Client, Config
from .exceptions import (
    ArtifactStoreError,
    BadRequestError,
    UnauthorizedError,
    ForbiddenError,
    NotFoundError,
    ConflictError,
    InternalServerError,
    ServiceUnavailableError,
)
from .models import (
    Bucket,
    Object,
    ListBucketsResult,
    ListObjectsResult,
    Signature,
    SBOM,
    Attestation,
    VerificationResult,
    MultipartUpload,
    CompletedPart,
)

__version__ = "1.0.0"
__all__ = [
    "Client",
    "Config",
    "ArtifactStoreError",
    "BadRequestError",
    "UnauthorizedError",
    "ForbiddenError",
    "NotFoundError",
    "ConflictError",
    "InternalServerError",
    "ServiceUnavailableError",
    "Bucket",
    "Object",
    "ListBucketsResult",
    "ListObjectsResult",
    "Signature",
    "SBOM",
    "Attestation",
    "VerificationResult",
    "MultipartUpload",
    "CompletedPart",
]
