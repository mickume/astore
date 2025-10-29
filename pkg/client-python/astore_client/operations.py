"""
Core artifact operations for Zot Artifact Store Client
"""

from typing import Optional, Dict, BinaryIO, Callable
from datetime import datetime
import json

from .models import (
    Bucket,
    Object,
    ListBucketsResult,
    ListObjectsResult,
    MultipartUpload,
    CompletedPart,
)


class Operations:
    """Core artifact operations mixin"""

    def __init__(self, client):
        self.client = client

    def create_bucket(self, bucket: str) -> None:
        """Create a new bucket"""
        url = self.client._url(f"/s3/{bucket}")
        self.client._request("PUT", url)

    def delete_bucket(self, bucket: str) -> None:
        """Delete a bucket"""
        url = self.client._url(f"/s3/{bucket}")
        self.client._request("DELETE", url)

    def list_buckets(self) -> ListBucketsResult:
        """List all buckets"""
        url = self.client._url("/s3")
        response = self.client._request("GET", url)

        data = response.json()
        buckets = []
        for b in data.get("buckets", []):
            buckets.append(
                Bucket(
                    name=b["name"],
                    creation_date=datetime.fromisoformat(b["creationDate"].replace("Z", "+00:00")),
                )
            )

        return ListBucketsResult(buckets=buckets)

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
        """Upload an artifact"""
        url = self.client._url(f"/s3/{bucket}/{key}")

        headers = {"Content-Type": content_type}

        # Add metadata headers
        if metadata:
            for k, v in metadata.items():
                headers[f"X-Amz-Meta-{k}"] = v

        # Wrap data with progress tracking if callback provided
        if progress_callback:
            data = ProgressReader(data, size, progress_callback)

        self.client._request("PUT", url, headers=headers, data=data)

    def download(
        self,
        bucket: str,
        key: str,
        writer: BinaryIO,
        byte_range: Optional[str] = None,
        progress_callback: Optional[Callable[[int], None]] = None,
    ) -> None:
        """Download an artifact"""
        url = self.client._url(f"/s3/{bucket}/{key}")

        headers = {}
        if byte_range:
            headers["Range"] = byte_range

        response = self.client._request("GET", url, headers=headers, stream=True)

        # Get total size from Content-Length header
        total_size = int(response.headers.get("Content-Length", 0))
        bytes_transferred = 0

        # Stream download with progress tracking
        chunk_size = 8192
        for chunk in response.iter_content(chunk_size=chunk_size):
            if chunk:
                writer.write(chunk)
                bytes_transferred += len(chunk)
                if progress_callback:
                    progress_callback(bytes_transferred)

    def get_object_metadata(self, bucket: str, key: str) -> Object:
        """Get object metadata"""
        url = self.client._url(f"/s3/{bucket}/{key}")
        response = self.client._request("HEAD", url)

        # Extract metadata from headers
        metadata = {}
        for header_name, header_value in response.headers.items():
            if header_name.lower().startswith("x-amz-meta-"):
                key_name = header_name[11:]  # Remove "x-amz-meta-" prefix
                metadata[key_name] = header_value

        size = int(response.headers.get("Content-Length", 0))
        last_modified_str = response.headers.get("Last-Modified", "")
        last_modified = datetime.strptime(
            last_modified_str, "%a, %d %b %Y %H:%M:%S %Z"
        ) if last_modified_str else datetime.now()

        return Object(
            key=key,
            size=size,
            last_modified=last_modified,
            etag=response.headers.get("ETag", "").strip('"'),
            content_type=response.headers.get("Content-Type", ""),
            metadata=metadata,
        )

    def delete_object(self, bucket: str, key: str) -> None:
        """Delete an object"""
        url = self.client._url(f"/s3/{bucket}/{key}")
        self.client._request("DELETE", url)

    def list_objects(
        self, bucket: str, prefix: str = "", max_keys: int = 1000
    ) -> ListObjectsResult:
        """List objects in a bucket"""
        url = self.client._url(f"/s3/{bucket}")

        params = {}
        if prefix:
            params["prefix"] = prefix
        if max_keys:
            params["max-keys"] = max_keys

        response = self.client._request("GET", url, params=params)

        data = response.json()
        objects = []
        for obj in data.get("contents", []):
            objects.append(
                Object(
                    key=obj["key"],
                    size=obj.get("size", 0),
                    last_modified=datetime.fromisoformat(obj.get("lastModified", "").replace("Z", "+00:00")) if obj.get("lastModified") else datetime.now(),
                    etag=obj.get("etag", ""),
                    content_type=obj.get("contentType", ""),
                )
            )

        return ListObjectsResult(
            objects=objects,
            prefix=data.get("prefix", ""),
            max_keys=data.get("maxKeys", 1000),
            is_truncated=data.get("isTruncated", False),
        )

    def copy_object(
        self, source_bucket: str, source_key: str, dest_bucket: str, dest_key: str
    ) -> None:
        """Copy an object"""
        url = self.client._url(f"/s3/{dest_bucket}/{dest_key}")

        headers = {"X-Amz-Copy-Source": f"/{source_bucket}/{source_key}"}

        self.client._request("PUT", url, headers=headers)

    def initiate_multipart_upload(
        self,
        bucket: str,
        key: str,
        content_type: str = "application/octet-stream",
        metadata: Optional[Dict[str, str]] = None,
    ) -> MultipartUpload:
        """Initiate multipart upload"""
        url = self.client._url(f"/s3/{bucket}/{key}")

        headers = {"Content-Type": content_type}

        # Add metadata headers
        if metadata:
            for k, v in metadata.items():
                headers[f"X-Amz-Meta-{k}"] = v

        response = self.client._request("POST", url, headers=headers, params={"uploads": ""})

        data = response.json()
        return MultipartUpload(
            upload_id=data["uploadId"], bucket=bucket, key=key
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
        """Upload a part in multipart upload"""
        url = self.client._url(f"/s3/{bucket}/{key}")

        params = {"uploadId": upload_id, "partNumber": part_number}

        response = self.client._request("PUT", url, params=params, data=data)

        return response.headers.get("ETag", "").strip('"')

    def complete_multipart_upload(
        self, bucket: str, key: str, upload_id: str, parts: list[CompletedPart]
    ) -> None:
        """Complete multipart upload"""
        url = self.client._url(f"/s3/{bucket}/{key}")

        params = {"uploadId": upload_id}

        # Build parts list
        parts_data = {"parts": [{"partNumber": p.part_number, "etag": p.etag} for p in parts]}

        self.client._request("POST", url, params=params, json=parts_data)

    def abort_multipart_upload(self, bucket: str, key: str, upload_id: str) -> None:
        """Abort multipart upload"""
        url = self.client._url(f"/s3/{bucket}/{key}")

        params = {"uploadId": upload_id}

        self.client._request("DELETE", url, params=params)


class ProgressReader:
    """Wrapper for BinaryIO that reports progress"""

    def __init__(self, data: BinaryIO, total_size: int, callback: Callable[[int], None]):
        self.data = data
        self.total_size = total_size
        self.callback = callback
        self.bytes_read = 0

    def read(self, size: int = -1) -> bytes:
        """Read data and report progress"""
        chunk = self.data.read(size)
        if chunk:
            self.bytes_read += len(chunk)
            self.callback(self.bytes_read)
        return chunk

    def __iter__(self):
        """Make iterable for requests library"""
        return self

    def __next__(self):
        """Read next chunk"""
        chunk = self.read(8192)
        if not chunk:
            raise StopIteration
        return chunk
