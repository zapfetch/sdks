"""
Zapfetch v1 API (Legacy)

This module provides the legacy v1 API for backward compatibility.

Usage:
    from zapfetch.v1 import V1ZapfetchApp
    app = V1ZapfetchApp(api_key="your-api-key")
    result = app.scrape_url("https://example.com")
"""

from .client import V1ZapfetchApp, AsyncV1ZapfetchApp, V1JsonConfig, V1ScrapeOptions, V1ChangeTrackingOptions

__all__ = ['V1ZapfetchApp', 'AsyncV1ZapfetchApp', 'V1JsonConfig', 'V1ScrapeOptions', 'V1ChangeTrackingOptions']