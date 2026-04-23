# ZapFetch Rust SDK

Rust SDK for the [ZapFetch](https://zapfetch.com) v2 web scraping API.
Scrape pages, crawl sites, search the web, map URLs, batch-scrape, and run
LLM-powered extraction agents — all returned in an LLM-ready format.

## Installation

```toml
[dependencies]
zapfetch = "0.1"
tokio = { version = "1", features = ["full"] }
```

## Quick Start

```rust
use zapfetch::Client;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Reads ZAPFETCH_API_KEY (required) and ZAPFETCH_API_URL (optional).
    let client = Client::from_env()?;

    let doc = client.scrape("https://example.com", None).await?;
    println!("{}", doc.markdown.unwrap_or_default());
    Ok(())
}
```

You can also construct a client explicitly:

```rust
use zapfetch::Client;

// Cloud (default: https://api.zapfetch.com)
let client = Client::new("zf-YOUR-API-KEY")?;

// Self-hosted
let client = Client::new_selfhosted("http://localhost:3000", Some("zf-key"))?;

// Bring your own reqwest::Client (custom timeout, no_proxy, etc.)
let http = reqwest::Client::builder()
    .timeout(std::time::Duration::from_secs(60))
    .no_proxy()
    .build()?;
let client = Client::with_http_client("https://api.zapfetch.com", Some("zf-key"), http)?;
```

## Examples

Runnable examples live in [`examples/`](./examples):

| Example | Demonstrates |
|---------|--------------|
| [`scrape`](./examples/scrape.rs) | Single-page scrape with format + main-content options |
| [`crawl`](./examples/crawl.rs) | Polling crawl with depth/limit constraints |
| [`search`](./examples/search.rs) | Search + inline scraping, WebResult/Document union |
| [`map`](./examples/map.rs) | URL discovery on a domain |
| [`batch_scrape`](./examples/batch_scrape.rs) | Concurrent scrape across many URLs |
| [`agent`](./examples/agent.rs) | Structured extraction with a typed `serde` schema |

Run any example:

```bash
ZAPFETCH_API_KEY=zf-your-key cargo run --example scrape -- https://example.com
```

## API Reference

### Scrape

```rust
use zapfetch::{Client, Format, ScrapeOptions};

let client = Client::from_env()?;
let doc = client
    .scrape(
        "https://example.com",
        ScrapeOptions {
            formats: Some(vec![Format::Markdown, Format::Html, Format::Links]),
            only_main_content: Some(true),
            ..Default::default()
        },
    )
    .await?;
```

### Structured extraction

```rust
use serde_json::json;

let schema = json!({
    "type": "object",
    "properties": {
        "title":       { "type": "string" },
        "description": { "type": "string" }
    }
});

let data = client
    .scrape_with_schema("https://example.com", schema, Some("Extract title and description"))
    .await?;
```

### Crawl

```rust
use zapfetch::{CrawlOptions, Format, ScrapeOptions};

let job = client
    .crawl(
        "https://example.com",
        CrawlOptions {
            limit: Some(50),
            max_discovery_depth: Some(3),
            scrape_options: Some(ScrapeOptions {
                formats: Some(vec![Format::Markdown]),
                ..Default::default()
            }),
            ..Default::default()
        },
    )
    .await?;

println!("crawled {} pages, credits used {:?}", job.data.len(), job.credits_used);
```

For fire-and-forget semantics use `start_crawl` + `get_crawl_status` + `cancel_crawl`.

### Search

```rust
let resp = client.search("zapfetch web scraping", None).await?;
```

### Map

```rust
use zapfetch::{MapOptions, SitemapMode};

let resp = client
    .map(
        "https://example.com",
        MapOptions {
            sitemap: Some(SitemapMode::Include),
            limit: Some(100),
            ..Default::default()
        },
    )
    .await?;
```

### Batch Scrape

```rust
use zapfetch::{BatchScrapeOptions, Format, ScrapeOptions};

let job = client
    .batch_scrape(
        vec!["https://example.com".into(), "https://example.org".into()],
        BatchScrapeOptions {
            options: Some(ScrapeOptions {
                formats: Some(vec![Format::Markdown]),
                ..Default::default()
            }),
            ..Default::default()
        },
    )
    .await?;
```

### Agent (LLM extraction)

```rust
use serde::Deserialize;
use serde_json::json;

#[derive(Debug, Deserialize)]
struct Company { name: String }

let schema = json!({
    "type": "object",
    "properties": { "name": { "type": "string" } },
    "required": ["name"]
});

let out: Option<Company> = client
    .agent_with_schema(
        vec!["https://example.com".into()],
        "Extract the company name",
        schema,
    )
    .await?;
```

### Scrape-bound interactive browsing (v2)

Replay a scrape job's browser context to run follow-up code:

```rust
use zapfetch::{Client, ScrapeExecuteLanguage, ScrapeExecuteOptions};

let run = client
    .interact(
        "550e8400-e29b-41d4-a716-446655440000",
        ScrapeExecuteOptions {
            code: Some("console.log(await page.url())".into()),
            language: Some(ScrapeExecuteLanguage::Node),
            timeout: Some(60),
            ..Default::default()
        },
    )
    .await?;

println!("{:?}", run.stdout);
client.stop_interaction("550e8400-e29b-41d4-a716-446655440000").await?;
```

## Error Handling

All fallible methods return `Result<T, zapfetch::Error>`. The error enum
discriminates transport, parse, and API failures so you can branch cleanly:

```rust
use zapfetch::Error;

match client.scrape("https://example.com", None).await {
    Ok(doc) => { /* ... */ }
    Err(Error::Api(action, body)) => {
        eprintln!("{} failed: {}", action, body.error);
    }
    Err(Error::HttpRequestFailed(action, status, _)) => {
        eprintln!("{} failed with HTTP {}", action, status);
    }
    Err(e) => eprintln!("unexpected: {e}"),
}
```

## Testing

The SDK ships with ~50 unit tests that use [`mockito`](https://docs.rs/mockito)
to verify request wiring without hitting the real API:

```bash
cargo test --lib
```

> **macOS + local HTTP proxy (Clash / Mihomo / etc.)**: system proxies
> intercept loopback requests to `mockito`'s ephemeral servers and return
> `502 Bad Gateway`. Export `NO_PROXY` before testing:
>
> ```bash
> NO_PROXY="127.0.0.1,localhost" cargo test --lib
> ```
>
> The SDK itself can bypass proxies via `Client::with_http_client(...)` with a
> `reqwest::Client::builder().no_proxy().build()`.

End-to-end tests in `tests/v2_e2e.rs` are `#[ignore]`-gated and require a live
API. Run them against a self-hosted instance or staging environment:

```bash
export API_URL=http://localhost:3000
export TEST_API_KEY=zf-your-key
cargo test --test v2_e2e -- --ignored
```

## License

MIT
