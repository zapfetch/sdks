## CHANGELOG

## [0.1.0]

### Added

- Initial ZapFetch Rust SDK release.
- Default API URL `https://api.zapfetch.com`.
- `Client::from_env()` reads `ZAPFETCH_API_KEY` / `ZAPFETCH_API_URL`.
- `Client::with_http_client()` for custom `reqwest::Client` (proxy, timeout, TLS).
- Six runnable examples: `scrape`, `crawl`, `search`, `map`, `batch_scrape`, `agent`.
