"""
Tests for artifact operations
"""

import pytest
import io
import responses
from datetime import datetime
from astore_client.exceptions import NotFoundError, ConflictError


class TestBucketOperations:
    """Test bucket management operations"""

    def test_create_bucket(self, client, mock_responses):
        """Given: Client configured
        When: Creating a bucket
        Then: Should send PUT request to correct endpoint"""
        mock_responses.add(
            responses.PUT, "https://test.example.com/s3/mybucket", status=200
        )

        client.create_bucket("mybucket")

        assert len(mock_responses.calls) == 1
        assert mock_responses.calls[0].request.url == "https://test.example.com/s3/mybucket"

    def test_delete_bucket(self, client, mock_responses):
        """Given: Existing bucket
        When: Deleting the bucket
        Then: Should send DELETE request"""
        mock_responses.add(
            responses.DELETE, "https://test.example.com/s3/mybucket", status=204
        )

        client.delete_bucket("mybucket")

        assert len(mock_responses.calls) == 1
        assert mock_responses.calls[0].request.method == "DELETE"

    def test_list_buckets(self, client, mock_responses):
        """Given: Multiple buckets exist
        When: Listing buckets
        Then: Should return all buckets"""
        mock_responses.add(
            responses.GET,
            "https://test.example.com/s3",
            json={
                "buckets": [
                    {"name": "bucket1", "creationDate": "2024-01-15T10:30:00Z"},
                    {"name": "bucket2", "creationDate": "2024-01-16T14:20:00Z"},
                ]
            },
            status=200,
        )

        result = client.list_buckets()

        assert len(result.buckets) == 2
        assert result.buckets[0].name == "bucket1"
        assert result.buckets[1].name == "bucket2"


class TestObjectOperations:
    """Test object/artifact operations"""

    def test_upload_object(self, client, mock_responses):
        """Given: Binary data to upload
        When: Uploading artifact
        Then: Should send PUT request with data"""
        mock_responses.add(
            responses.PUT,
            "https://test.example.com/s3/mybucket/myfile.tar.gz",
            status=200,
        )

        data = io.BytesIO(b"test data")
        client.upload("mybucket", "myfile.tar.gz", data, 9)

        assert len(mock_responses.calls) == 1
        assert mock_responses.calls[0].request.body == b"test data"

    def test_upload_with_metadata(self, client, mock_responses):
        """Given: Artifact with custom metadata
        When: Uploading
        Then: Should include metadata in headers"""
        mock_responses.add(
            responses.PUT,
            "https://test.example.com/s3/mybucket/myfile.tar.gz",
            status=200,
        )

        data = io.BytesIO(b"test data")
        metadata = {"version": "1.0.0", "author": "test"}

        client.upload("mybucket", "myfile.tar.gz", data, 9, metadata=metadata)

        headers = mock_responses.calls[0].request.headers
        assert "X-Amz-Meta-version" in headers
        assert headers["X-Amz-Meta-version"] == "1.0.0"
        assert "X-Amz-Meta-author" in headers
        assert headers["X-Amz-Meta-author"] == "test"

    def test_download_object(self, client, mock_responses):
        """Given: Artifact exists
        When: Downloading artifact
        Then: Should download content"""
        mock_responses.add(
            responses.GET,
            "https://test.example.com/s3/mybucket/myfile.tar.gz",
            body=b"downloaded data",
            headers={"Content-Length": "15"},
            status=200,
        )

        output = io.BytesIO()
        client.download("mybucket", "myfile.tar.gz", output)

        assert output.getvalue() == b"downloaded data"

    def test_download_with_range(self, client, mock_responses):
        """Given: Artifact exists
        When: Downloading with byte range
        Then: Should include Range header"""
        mock_responses.add(
            responses.GET,
            "https://test.example.com/s3/mybucket/myfile.tar.gz",
            body=b"partial",
            headers={"Content-Length": "7"},
            status=206,
        )

        output = io.BytesIO()
        client.download("mybucket", "myfile.tar.gz", output, byte_range="bytes=0-6")

        headers = mock_responses.calls[0].request.headers
        assert "Range" in headers
        assert headers["Range"] == "bytes=0-6"

    def test_get_object_metadata(self, client, mock_responses):
        """Given: Artifact exists
        When: Getting metadata
        Then: Should return object metadata"""
        mock_responses.add(
            responses.HEAD,
            "https://test.example.com/s3/mybucket/myfile.tar.gz",
            headers={
                "Content-Length": "1024",
                "Content-Type": "application/gzip",
                "ETag": '"abc123"',
                "Last-Modified": "Mon, 15 Jan 2024 10:30:00 GMT",
                "X-Amz-Meta-version": "1.0.0",
            },
            status=200,
        )

        obj = client.get_object_metadata("mybucket", "myfile.tar.gz")

        assert obj.key == "myfile.tar.gz"
        assert obj.size == 1024
        assert obj.content_type == "application/gzip"
        assert obj.etag == "abc123"
        assert obj.metadata["version"] == "1.0.0"

    def test_delete_object(self, client, mock_responses):
        """Given: Artifact exists
        When: Deleting artifact
        Then: Should send DELETE request"""
        mock_responses.add(
            responses.DELETE,
            "https://test.example.com/s3/mybucket/myfile.tar.gz",
            status=204,
        )

        client.delete_object("mybucket", "myfile.tar.gz")

        assert len(mock_responses.calls) == 1
        assert mock_responses.calls[0].request.method == "DELETE"

    def test_list_objects(self, client, mock_responses):
        """Given: Objects in bucket
        When: Listing objects
        Then: Should return object list"""
        mock_responses.add(
            responses.GET,
            "https://test.example.com/s3/mybucket",
            json={
                "contents": [
                    {
                        "key": "file1.tar.gz",
                        "size": 1024,
                        "lastModified": "2024-01-15T10:30:00Z",
                        "etag": "abc",
                        "contentType": "application/gzip",
                    },
                    {
                        "key": "file2.tar.gz",
                        "size": 2048,
                        "lastModified": "2024-01-16T14:20:00Z",
                        "etag": "def",
                        "contentType": "application/gzip",
                    },
                ],
                "prefix": "",
                "maxKeys": 1000,
                "isTruncated": False,
            },
            status=200,
        )

        result = client.list_objects("mybucket")

        assert len(result.objects) == 2
        assert result.objects[0].key == "file1.tar.gz"
        assert result.objects[0].size == 1024
        assert result.objects[1].key == "file2.tar.gz"

    def test_list_objects_with_prefix(self, client, mock_responses):
        """Given: Objects with different prefixes
        When: Listing with prefix filter
        Then: Should include prefix in request"""
        mock_responses.add(
            responses.GET,
            "https://test.example.com/s3/mybucket?prefix=app%2F&max-keys=1000",
            json={"contents": [], "prefix": "app/", "maxKeys": 1000, "isTruncated": False},
            status=200,
        )

        result = client.list_objects("mybucket", prefix="app/")

        assert result.prefix == "app/"

    def test_copy_object(self, client, mock_responses):
        """Given: Source artifact exists
        When: Copying to destination
        Then: Should send PUT request with copy source header"""
        mock_responses.add(
            responses.PUT,
            "https://test.example.com/s3/destbucket/destfile.tar.gz",
            status=200,
        )

        client.copy_object("srcbucket", "srcfile.tar.gz", "destbucket", "destfile.tar.gz")

        headers = mock_responses.calls[0].request.headers
        assert "X-Amz-Copy-Source" in headers
        assert headers["X-Amz-Copy-Source"] == "/srcbucket/srcfile.tar.gz"


class TestMultipartUpload:
    """Test multipart upload operations"""

    def test_initiate_multipart_upload(self, client, mock_responses):
        """Given: Large file to upload
        When: Initiating multipart upload
        Then: Should return upload ID"""
        mock_responses.add(
            responses.POST,
            "https://test.example.com/s3/mybucket/largefile.tar.gz?uploads=",
            json={"uploadId": "upload123"},
            status=200,
        )

        upload = client.initiate_multipart_upload("mybucket", "largefile.tar.gz")

        assert upload.upload_id == "upload123"
        assert upload.bucket == "mybucket"
        assert upload.key == "largefile.tar.gz"

    def test_upload_part(self, client, mock_responses):
        """Given: Initiated multipart upload
        When: Uploading a part
        Then: Should return ETag"""
        mock_responses.add(
            responses.PUT,
            "https://test.example.com/s3/mybucket/largefile.tar.gz?uploadId=upload123&partNumber=1",
            headers={"ETag": '"part-etag-1"'},
            status=200,
        )

        data = io.BytesIO(b"part data")
        etag = client.upload_part("mybucket", "largefile.tar.gz", "upload123", 1, data, 9)

        assert etag == "part-etag-1"

    def test_complete_multipart_upload(self, client, mock_responses):
        """Given: All parts uploaded
        When: Completing multipart upload
        Then: Should send parts list"""
        from astore_client.models import CompletedPart

        mock_responses.add(
            responses.POST,
            "https://test.example.com/s3/mybucket/largefile.tar.gz?uploadId=upload123",
            status=200,
        )

        parts = [
            CompletedPart(part_number=1, etag="etag1"),
            CompletedPart(part_number=2, etag="etag2"),
        ]

        client.complete_multipart_upload("mybucket", "largefile.tar.gz", "upload123", parts)

        assert len(mock_responses.calls) == 1

    def test_abort_multipart_upload(self, client, mock_responses):
        """Given: Initiated multipart upload
        When: Aborting upload
        Then: Should send DELETE request"""
        mock_responses.add(
            responses.DELETE,
            "https://test.example.com/s3/mybucket/largefile.tar.gz?uploadId=upload123",
            status=204,
        )

        client.abort_multipart_upload("mybucket", "largefile.tar.gz", "upload123")

        assert len(mock_responses.calls) == 1


class TestErrorHandling:
    """Test HTTP error handling"""

    def test_404_not_found(self, client, mock_responses):
        """Given: Non-existent artifact
        When: Trying to download
        Then: Should raise NotFoundError"""
        mock_responses.add(
            responses.GET,
            "https://test.example.com/s3/mybucket/nonexistent.tar.gz",
            json={"error": "Artifact not found"},
            status=404,
        )

        with pytest.raises(NotFoundError, match="Artifact not found"):
            output = io.BytesIO()
            client.download("mybucket", "nonexistent.tar.gz", output)

    def test_409_conflict(self, client, mock_responses):
        """Given: Bucket already exists
        When: Creating bucket
        Then: Should raise ConflictError"""
        mock_responses.add(
            responses.PUT,
            "https://test.example.com/s3/mybucket",
            json={"error": "Bucket already exists"},
            status=409,
        )

        with pytest.raises(ConflictError, match="Bucket already exists"):
            client.create_bucket("mybucket")


class TestProgressCallbacks:
    """Test progress tracking"""

    def test_upload_progress_callback(self, client, mock_responses):
        """Given: Upload with progress callback
        When: Uploading data
        Then: Callback should be called with progress"""
        mock_responses.add(
            responses.PUT,
            "https://test.example.com/s3/mybucket/myfile.tar.gz",
            status=200,
        )

        progress_values = []

        def progress_callback(bytes_transferred):
            progress_values.append(bytes_transferred)

        data = io.BytesIO(b"test data")
        client.upload(
            "mybucket", "myfile.tar.gz", data, 9, progress_callback=progress_callback
        )

        assert len(progress_values) > 0
        assert progress_values[-1] == 9

    def test_download_progress_callback(self, client, mock_responses):
        """Given: Download with progress callback
        When: Downloading data
        Then: Callback should be called with progress"""
        mock_responses.add(
            responses.GET,
            "https://test.example.com/s3/mybucket/myfile.tar.gz",
            body=b"downloaded data",
            headers={"Content-Length": "15"},
            status=200,
        )

        progress_values = []

        def progress_callback(bytes_transferred):
            progress_values.append(bytes_transferred)

        output = io.BytesIO()
        client.download(
            "mybucket", "myfile.tar.gz", output, progress_callback=progress_callback
        )

        assert len(progress_values) > 0
        assert progress_values[-1] == 15
