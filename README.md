# ZapFetch SDKs

Official SDKs for the [ZapFetch](https://zapfetch.com) v2 web scraping API — scrape pages, crawl sites, search the web, map URLs, batch-scrape, and run LLM-powered extraction agents, all returned in an LLM-ready format.

## Packages

| Language | Registry | Install |
|---|---|---|
| **Go** | [pkg.go.dev](https://pkg.go.dev/github.com/zapfetch/sdks/go-sdk) | `go get github.com/zapfetch/sdks/go-sdk` |
| **Rust** | [![crates.io](https://img.shields.io/crates/v/zapfetch?label=zapfetch&color=orange)](https://crates.io/crates/zapfetch) | `cargo add zapfetch` |
| **Python** | [![PyPI](https://img.shields.io/pypi/v/zapfetch-py?label=zapfetch-py&color=blue)](https://pypi.org/project/zapfetch-py/) | `pip install zapfetch-py` |
| **JavaScript / TypeScript** | [![npm](https://img.shields.io/npm/v/@zapfetchdev/sdk?label=@zapfetchdev/sdk&color=red)](https://www.npmjs.com/package/@zapfetchdev/sdk) | `pnpm add @zapfetchdev/sdk` |
| **Java** | [![Maven Central](https://img.shields.io/maven-central/v/com.zapfetch/zapfetch-java?label=com.zapfetch:zapfetch-java&color=green)](https://central.sonatype.com/artifact/com.zapfetch/zapfetch-java) | `implementation("com.zapfetch:zapfetch-java:0.1.0")` |

All SDKs share the same configuration surface — set `ZAPFETCH_API_KEY` in the environment (get one at [zapfetch.com](https://zapfetch.com)) and you're ready to scrape.

## Quick Start

### Go

```go
package main

import (
    "context"
    "fmt"

    "github.com/zapfetch/sdks/go-sdk"
)

func main() {
    client, _ := zapfetch.NewClient() // reads ZAPFETCH_API_KEY
    doc, _ := client.Scrape(context.Background(), "https://example.com", nil)
    fmt.Println(doc.Markdown)
}
```

### Rust

```rust
use zapfetch::Client;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let client = Client::from_env()?; // reads ZAPFETCH_API_KEY
    let doc = client.scrape("https://example.com", None).await?;
    println!("{}", doc.markdown.unwrap_or_default());
    Ok(())
}
```

### Python

```python
from zapfetch import Zapfetch

client = Zapfetch(api_key="zf-...")  # or Zapfetch() to read from env
doc = client.scrape("https://example.com")
print(doc.markdown)
```

### JavaScript / TypeScript

```ts
import Zapfetch from "@zapfetchdev/sdk";

const client = new Zapfetch({ apiKey: process.env.ZAPFETCH_API_KEY });
const doc = await client.scrape("https://example.com");
console.log(doc.markdown);
```

### Java

```java
import com.zapfetch.client.ZapfetchClient;
import com.zapfetch.models.Document;

var client = ZapfetchClient.builder()
    .apiKey(System.getenv("ZAPFETCH_API_KEY"))
    .build();
Document doc = client.scrape("https://example.com");
System.out.println(doc.getMarkdown());
```

## Features

Every SDK exposes the same endpoints:

| Endpoint | What it does |
|---|---|
| `scrape` | Single-page scrape with markdown / HTML / links / screenshot formats |
| `crawl` | Multi-page crawl with depth + concurrency limits |
| `batch_scrape` | Concurrent scrape across a URL list |
| `search` | Web search with optional inline scraping |
| `map` | URL discovery on a domain |
| `extract` | JSON-schema-constrained structured extraction |
| `agent` | LLM-powered autonomous browsing with schemas |

Async / streaming APIs and WebSocket watchers for long-running jobs are available in all SDKs.

## Configuration

| Env var | Purpose | Default |
|---|---|---|
| `ZAPFETCH_API_KEY` | API key (required) | — |
| `ZAPFETCH_API_URL` | Base URL (override for self-hosted) | `https://api.zapfetch.com` |

Each SDK also accepts these values as explicit constructor arguments — see the per-SDK README for the full option surface.

## Per-SDK documentation

| SDK | Source | README |
|---|---|---|
| Go | [`go-sdk/`](./go-sdk) | [README](./go-sdk/README.md) |
| Rust | [`rust-sdk/`](./rust-sdk) | [README](./rust-sdk/README.md) |
| Python | [`python-sdk/`](./python-sdk) | [README](./python-sdk/README.md) |
| JavaScript | [`js-sdk/`](./js-sdk) | [README](./js-sdk/README.md) |
| Java | [`java-sdk/`](./java-sdk) | [README](./java-sdk/README.md) |

## Versioning

All SDKs follow [SemVer](https://semver.org/) and share the same major version track. A release of `zapfetch@0.2.0` on any language implies corresponding `0.2.0` releases on every SDK within a short window, unless only one platform is affected.

Release tags are prefixed with the SDK name to coexist in this monorepo:

- `go-sdk/v0.1.0`
- `rust-sdk/v0.1.0`
- `python-sdk/v0.1.0`
- `js-sdk/v0.1.0`
- `java-sdk/v0.1.0`

## License

MIT — see [LICENSE](./LICENSE).
