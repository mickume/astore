"""
Pytest fixtures for astore_client tests
"""

import pytest
import responses
from astore_client import Client, Config


@pytest.fixture
def config():
    """Basic client configuration"""
    return Config(base_url="https://test.example.com", token="test-token")


@pytest.fixture
def client(config):
    """Client instance for testing"""
    return Client(config)


@pytest.fixture
def mock_responses():
    """Setup responses mock"""
    with responses.RequestsMock() as rsps:
        yield rsps
