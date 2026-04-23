import sys
from pathlib import Path

import pytest

ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from zapfetch.v2.client import ZapfetchClient
from zapfetch.v2.client_async import AsyncZapfetchClient


@pytest.fixture(autouse=True)
def clear_zapfetch_api_key_env(monkeypatch):
    monkeypatch.delenv("ZAPFETCH_API_KEY", raising=False)
    yield


def test_cloud_requires_api_key():
    with pytest.raises(ValueError):
        ZapfetchClient(api_url="https://api.zapfetch.com")


def test_self_host_allows_missing_api_key():
    client = ZapfetchClient(api_url="http://localhost:3000")
    assert client.http_client.api_key is None


def test_async_cloud_requires_api_key():
    with pytest.raises(ValueError):
        AsyncZapfetchClient(api_url="https://api.zapfetch.com")


@pytest.mark.asyncio
async def test_async_self_host_allows_missing_api_key():
    client = AsyncZapfetchClient(api_url="http://localhost:3000")
    try:
        assert client.http_client.api_key is None
        await client.async_http_client.close()
    finally:
        # Ensure the underlying HTTPX client is closed even if assertions fail
        if not client.async_http_client._client.is_closed:
            await client.async_http_client.close()
