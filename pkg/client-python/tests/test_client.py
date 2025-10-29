"""
Tests for Client class
"""

import pytest
from astore_client import Client, Config
from astore_client.exceptions import ArtifactStoreError


class TestClientConfiguration:
    """Test client configuration and initialization"""

    def test_create_client_with_valid_config(self):
        """Given: Valid configuration
        When: Creating a client
        Then: Client should be created successfully"""
        config = Config(base_url="https://test.example.com")
        client = Client(config)

        assert client.config.base_url == "https://test.example.com"
        assert client.config.timeout == 60
        assert client.session is not None

    def test_create_client_with_missing_base_url(self):
        """Given: Configuration without base_url
        When: Creating config
        Then: Should raise ValueError"""
        with pytest.raises(ValueError, match="base_url is required"):
            Config(base_url="")

    def test_create_client_with_token(self):
        """Given: Configuration with authentication token
        When: Creating a client
        Then: Authorization header should be set"""
        config = Config(base_url="https://test.example.com", token="my-token")
        client = Client(config)

        assert "Authorization" in client.session.headers
        assert client.session.headers["Authorization"] == "Bearer my-token"

    def test_create_client_with_custom_timeout(self):
        """Given: Configuration with custom timeout
        When: Creating a client
        Then: Timeout should be set correctly"""
        config = Config(base_url="https://test.example.com", timeout=120)
        client = Client(config)

        assert client.config.timeout == 120

    def test_create_client_with_insecure_skip_verify(self):
        """Given: Configuration with insecure_skip_verify
        When: Creating a client
        Then: TLS verification should be disabled"""
        config = Config(
            base_url="https://test.example.com", insecure_skip_verify=True
        )
        client = Client(config)

        assert client.session.verify is False

    def test_create_client_with_custom_user_agent(self):
        """Given: Configuration with custom user agent
        When: Creating a client
        Then: User-Agent header should be set"""
        config = Config(
            base_url="https://test.example.com", user_agent="my-app/1.0"
        )
        client = Client(config)

        assert client.session.headers["User-Agent"] == "my-app/1.0"

    def test_base_url_trailing_slash_removed(self):
        """Given: Base URL with trailing slash
        When: Creating config
        Then: Trailing slash should be removed"""
        config = Config(base_url="https://test.example.com/")
        assert config.base_url == "https://test.example.com"


class TestSetToken:
    """Test token management"""

    def test_update_authentication_token(self, client):
        """Given: Client with initial token
        When: Updating token
        Then: Authorization header should be updated"""
        client.set_token("new-token")

        assert client.config.token == "new-token"
        assert client.session.headers["Authorization"] == "Bearer new-token"


class TestURLBuilding:
    """Test URL building"""

    def test_url_building_with_path(self, client):
        """Given: Client with base URL
        When: Building URL with path
        Then: Full URL should be constructed correctly"""
        url = client._url("/s3/mybucket")
        assert url == "https://test.example.com/s3/mybucket"

    def test_url_building_with_leading_slash(self, client):
        """Given: Path with leading slash
        When: Building URL
        Then: Should handle leading slash correctly"""
        url = client._url("/path")
        assert url == "https://test.example.com/path"
