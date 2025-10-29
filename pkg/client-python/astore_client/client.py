"""
Zot Artifact Store Client

Main client class for interacting with the Zot Artifact Store API.
"""

import requests
from typing import Optional, Dict, BinaryIO, Callable
from urllib.parse import urljoin
import json

from .exceptions import raise_for_status
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
from .operations import Operations
from .supplychain import SupplyChain


class Config:
    """Client configuration"""

    def __init__(
        self,
        base_url: str,
        token: Optional[str] = None,
        timeout: int = 60,
        insecure_skip_verify: bool = False,
        user_agent: str = "astore-python/1.0.0",
    ):
        """
        Initialize client configuration

        Args:
            base_url: Artifact store server URL
            token: Bearer authentication token (optional)
            timeout: Request timeout in seconds (default: 60)
            insecure_skip_verify: Skip TLS certificate verification (default: False)
            user_agent: Custom User-Agent header (default: astore-python/1.0.0)
        """
        if not base_url:
            raise ValueError("base_url is required")

        self.base_url = base_url.rstrip("/")
        self.token = token
        self.timeout = timeout
        self.insecure_skip_verify = insecure_skip_verify
        self.user_agent = user_agent


class Client:
    """
    Zot Artifact Store Client

    Example:
        >>> config = Config(
        ...     base_url="https://artifacts.example.com",
        ...     token="your-token"
        ... )
        >>> client = Client(config)
        >>> client.upload("mybucket", "myfile.tar.gz", data, len(data))
    """

    def __init__(self, config: Config):
        """
        Initialize client

        Args:
            config: Client configuration
        """
        self.config = config
        self.session = requests.Session()

        # Configure session
        self.session.headers.update({"User-Agent": config.user_agent})

        if config.token:
            self.session.headers.update({"Authorization": f"Bearer {config.token}"})

        # Configure TLS verification
        if config.insecure_skip_verify:
            self.session.verify = False
            import urllib3

            urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

        # Initialize operation mixins
        self._operations = Operations(self)
        self._supplychain = SupplyChain(self)

    def set_token(self, token: str):
        """
        Update authentication token

        Args:
            token: New bearer token
        """
        self.config.token = token
        self.session.headers.update({"Authorization": f"Bearer {token}"})

    def _url(self, path: str) -> str:
        """Build full URL from path"""
        return urljoin(self.config.base_url, path.lstrip("/"))

    def _request(
        self,
        method: str,
        url: str,
        headers: Optional[Dict[str, str]] = None,
        **kwargs,
    ) -> requests.Response:
        """
        Make HTTP request

        Args:
            method: HTTP method (GET, POST, PUT, DELETE, etc.)
            url: Full URL
            headers: Additional headers
            **kwargs: Additional arguments for requests

        Returns:
            Response object

        Raises:
            ArtifactStoreError: On HTTP error
        """
        if headers:
            request_headers = self.session.headers.copy()
            request_headers.update(headers)
            kwargs["headers"] = request_headers

        kwargs.setdefault("timeout", self.config.timeout)

        response = self.session.request(method, url, **kwargs)
        raise_for_status(response)
        return response

    # Bucket operations
    def create_bucket(self, bucket: str) -> None:
        """
        Create a new bucket

        Args:
            bucket: Bucket name

        Raises:
            ConflictError: If bucket already exists
            ArtifactStoreError: On other errors
        """
        return self._operations.create_bucket(bucket)

    def delete_bucket(self, bucket: str) -> None:
        """
        Delete a bucket

        Args:
            bucket: Bucket name

        Raises:
            NotFoundError: If bucket doesn't exist
            ArtifactStoreError: On other errors
        """
        return self._operations.delete_bucket(bucket)

    def list_buckets(self) -> ListBucketsResult:
        """
        List all buckets

        Returns:
            ListBucketsResult containing bucket list

        Raises:
            ArtifactStoreError: On error
        """
        return self._operations.list_buckets()

    # Object operations
    def upload(
        self,
        bucket: str,
        key: str,
        data: BinaryIO,
        size: int,
        content_type: str = "application/octet-stream",
        metadata: Optional[Dict[str, str]] = None,
        progress_callback: Optional[Callable[[int], None]] = None,
    ) -> None:
        """
        Upload an artifact

        Args:
            bucket: Bucket name
            key: Object key
            data: Binary data to upload
            size: Size of data in bytes
            content_type: Content type (default: application/octet-stream)
            metadata: Custom metadata (optional)
            progress_callback: Progress callback function(bytes_transferred)

        Raises:
            NotFoundError: If bucket doesn't exist
            ArtifactStoreError: On other errors
        """
        return self._operations.upload(
            bucket, key, data, size, content_type, metadata, progress_callback
        )

    def download(
        self,
        bucket: str,
        key: str,
        writer: BinaryIO,
        byte_range: Optional[str] = None,
        progress_callback: Optional[Callable[[int], None]] = None,
    ) -> None:
        """
        Download an artifact

        Args:
            bucket: Bucket name
            key: Object key
            writer: Writer to write data to
            byte_range: Byte range (e.g., "bytes=0-1023")
            progress_callback: Progress callback function(bytes_transferred)

        Raises:
            NotFoundError: If object doesn't exist
            ArtifactStoreError: On other errors
        """
        return self._operations.download(bucket, key, writer, byte_range, progress_callback)

    def get_object_metadata(self, bucket: str, key: str) -> Object:
        """
        Get object metadata

        Args:
            bucket: Bucket name
            key: Object key

        Returns:
            Object metadata

        Raises:
            NotFoundError: If object doesn't exist
            ArtifactStoreError: On other errors
        """
        return self._operations.get_object_metadata(bucket, key)

    def delete_object(self, bucket: str, key: str) -> None:
        """
        Delete an object

        Args:
            bucket: Bucket name
            key: Object key

        Raises:
            NotFoundError: If object doesn't exist
            ArtifactStoreError: On other errors
        """
        return self._operations.delete_object(bucket, key)

    def list_objects(
        self, bucket: str, prefix: str = "", max_keys: int = 1000
    ) -> ListObjectsResult:
        """
        List objects in a bucket

        Args:
            bucket: Bucket name
            prefix: Key prefix filter (default: "")
            max_keys: Maximum number of keys to return (default: 1000)

        Returns:
            ListObjectsResult containing object list

        Raises:
            NotFoundError: If bucket doesn't exist
            ArtifactStoreError: On other errors
        """
        return self._operations.list_objects(bucket, prefix, max_keys)

    def copy_object(
        self, source_bucket: str, source_key: str, dest_bucket: str, dest_key: str
    ) -> None:
        """
        Copy an object

        Args:
            source_bucket: Source bucket name
            source_key: Source object key
            dest_bucket: Destination bucket name
            dest_key: Destination object key

        Raises:
            NotFoundError: If source object doesn't exist
            ArtifactStoreError: On other errors
        """
        return self._operations.copy_object(
            source_bucket, source_key, dest_bucket, dest_key
        )

    # Multipart upload operations
    def initiate_multipart_upload(
        self,
        bucket: str,
        key: str,
        content_type: str = "application/octet-stream",
        metadata: Optional[Dict[str, str]] = None,
    ) -> MultipartUpload:
        """
        Initiate multipart upload

        Args:
            bucket: Bucket name
            key: Object key
            content_type: Content type (default: application/octet-stream)
            metadata: Custom metadata (optional)

        Returns:
            MultipartUpload object

        Raises:
            ArtifactStoreError: On error
        """
        return self._operations.initiate_multipart_upload(
            bucket, key, content_type, metadata
        )

    def upload_part(
        self,
        bucket: str,
        key: str,
        upload_id: str,
        part_number: int,
        data: BinaryIO,
        size: int,
    ) -> str:
        """
        Upload a part in multipart upload

        Args:
            bucket: Bucket name
            key: Object key
            upload_id: Upload ID from initiate_multipart_upload
            part_number: Part number (1-based)
            data: Part data
            size: Part size in bytes

        Returns:
            ETag of uploaded part

        Raises:
            ArtifactStoreError: On error
        """
        return self._operations.upload_part(
            bucket, key, upload_id, part_number, data, size
        )

    def complete_multipart_upload(
        self, bucket: str, key: str, upload_id: str, parts: list[CompletedPart]
    ) -> None:
        """
        Complete multipart upload

        Args:
            bucket: Bucket name
            key: Object key
            upload_id: Upload ID from initiate_multipart_upload
            parts: List of completed parts

        Raises:
            ArtifactStoreError: On error
        """
        return self._operations.complete_multipart_upload(
            bucket, key, upload_id, parts
        )

    def abort_multipart_upload(self, bucket: str, key: str, upload_id: str) -> None:
        """
        Abort multipart upload

        Args:
            bucket: Bucket name
            key: Object key
            upload_id: Upload ID from initiate_multipart_upload

        Raises:
            ArtifactStoreError: On error
        """
        return self._operations.abort_multipart_upload(bucket, key, upload_id)

    # Supply chain operations
    def sign_artifact(self, bucket: str, key: str, private_key: str) -> Signature:
        """
        Sign an artifact

        Args:
            bucket: Bucket name
            key: Object key
            private_key: PEM-encoded private key

        Returns:
            Signature object

        Raises:
            ArtifactStoreError: On error
        """
        return self._supplychain.sign_artifact(bucket, key, private_key)

    def get_signatures(self, bucket: str, key: str) -> list[Signature]:
        """
        Get artifact signatures

        Args:
            bucket: Bucket name
            key: Object key

        Returns:
            List of signatures

        Raises:
            ArtifactStoreError: On error
        """
        return self._supplychain.get_signatures(bucket, key)

    def verify_signatures(
        self, bucket: str, key: str, public_keys: list[str]
    ) -> VerificationResult:
        """
        Verify artifact signatures

        Args:
            bucket: Bucket name
            key: Object key
            public_keys: List of PEM-encoded public keys

        Returns:
            VerificationResult

        Raises:
            ArtifactStoreError: On error
        """
        return self._supplychain.verify_signatures(bucket, key, public_keys)

    def attach_sbom(self, bucket: str, key: str, format: str, content: str) -> SBOM:
        """
        Attach SBOM to artifact

        Args:
            bucket: Bucket name
            key: Object key
            format: SBOM format (e.g., "spdx", "cyclonedx")
            content: SBOM content

        Returns:
            SBOM object

        Raises:
            ArtifactStoreError: On error
        """
        return self._supplychain.attach_sbom(bucket, key, format, content)

    def get_sbom(self, bucket: str, key: str) -> SBOM:
        """
        Get artifact SBOM

        Args:
            bucket: Bucket name
            key: Object key

        Returns:
            SBOM object

        Raises:
            NotFoundError: If SBOM doesn't exist
            ArtifactStoreError: On other errors
        """
        return self._supplychain.get_sbom(bucket, key)

    def add_attestation(
        self, bucket: str, key: str, attestation_type: str, data: Dict
    ) -> Attestation:
        """
        Add attestation to artifact

        Args:
            bucket: Bucket name
            key: Object key
            attestation_type: Attestation type (e.g., "build", "test", "scan")
            data: Attestation data

        Returns:
            Attestation object

        Raises:
            ArtifactStoreError: On error
        """
        return self._supplychain.add_attestation(bucket, key, attestation_type, data)

    def get_attestations(self, bucket: str, key: str) -> list[Attestation]:
        """
        Get artifact attestations

        Args:
            bucket: Bucket name
            key: Object key

        Returns:
            List of attestations

        Raises:
            ArtifactStoreError: On error
        """
        return self._supplychain.get_attestations(bucket, key)
