# ZapFetch Go SDK

Go SDK for the [ZapFetch](https://zapfetch.com) v2 web scraping API.

## Requirements

- **Go:** 1.23 or later

## Installation

```bash
go get github.com/zapfetch/sdks/go-sdk
```

## API Key Setup

Get your API key from the [ZapFetch Dashboard](https://zapfetch.com/dashboard) and set it as an environment variable:

```bash
export ZAPFETCH_API_KEY="zf-your-api-key-here"
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/zapfetch/sdks/go-sdk"
	"github.com/zapfetch/sdks/go-sdk/option"
)

func main() {
	// Create a client (reads ZAPFETCH_API_KEY from environment)
	client, err := zapfetch.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// Or provide the API key directly
	client, err = zapfetch.NewClient(
		option.WithAPIKey("zf-your-api-key"),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Scrape a single page
	doc, err := client.Scrape(ctx, "https://example.com", &zapfetch.ScrapeOptions{
		Formats: []string{"markdown"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(doc.Markdown)
}
```

## Configuration

```go
client, err := zapfetch.NewClient(
	option.WithAPIKey("zf-your-api-key"),           // API key (or set ZAPFETCH_API_KEY)
	option.WithAPIURL("https://api.zapfetch.com"),  // Override base URL (or set ZAPFETCH_API_URL)
	option.WithMaxRetries(3),                        // Max retry attempts (default: 3)
	option.WithBackoffFactor(0.5),                   // Exponential backoff factor (default: 0.5)
	option.WithTimeout(5 * time.Minute),             // HTTP timeout (default: 5 minutes)
	option.WithHTTPClient(customHTTPClient),         // Custom *http.Client
	option.WithHeader("X-Custom", "value"),          // Extra header on every request
)
```

The default base URL is **`https://api.zapfetch.com`**. Override with `option.WithAPIURL()`
or the `ZAPFETCH_API_URL` environment variable (useful for self-hosted or staging environments).

## Examples

Runnable examples live in [`examples/`](./examples):

| Example | Demonstrates |
|---------|--------------|
| [`examples/scrape`](./examples/scrape) | Single-page scrape with formats and geo options |
| [`examples/crawl`](./examples/crawl) | Polling crawl with depth limits |
| [`examples/search`](./examples/search) | Search + inline scraping |
| [`examples/map`](./examples/map) | URL discovery on a domain |
| [`examples/batch_scrape`](./examples/batch_scrape) | Concurrent scrape across many URLs |
| [`examples/agent`](./examples/agent) | Structured extraction with a JSON schema |

Run any example:

```bash
cd examples/scrape
ZAPFETCH_API_KEY=zf-your-key go run .
```

## API Reference

### Scrape

Scrape a single URL and get its content.

```go
// Basic scrape
doc, err := client.Scrape(ctx, "https://example.com", nil)

// With options
doc, err := client.Scrape(ctx, "https://example.com", &zapfetch.ScrapeOptions{
	Formats:         []string{"markdown", "html"},
	OnlyMainContent: zapfetch.Bool(true),
	WaitFor:         zapfetch.Int(5000),
	Location:        &zapfetch.LocationConfig{Country: "US"},
})
```

#### Interactive Browser

Execute code in a scrape-bound browser session:

```go
resp, err := client.Interact(ctx, scrapeJobID, "document.title", &zapfetch.InteractParams{
	Language: "node",
	Timeout:  zapfetch.Int(30),
})

// Stop the browser session
deleteResp, err := client.StopInteractiveBrowser(ctx, scrapeJobID)
```

### Crawl

Crawl a website and get content from multiple pages.

```go
// Auto-polling: starts the crawl and waits for completion
job, err := client.Crawl(ctx, "https://example.com", &zapfetch.CrawlOptions{
	Limit:             zapfetch.Int(50),
	MaxDiscoveryDepth: zapfetch.Int(3),
	ScrapeOptions: &zapfetch.ScrapeOptions{
		Formats: []string{"markdown"},
	},
})

// Or manage polling manually
resp, err := client.StartCrawl(ctx, "https://example.com", &zapfetch.CrawlOptions{
	Limit: zapfetch.Int(50),
})

// Check status
status, err := client.GetCrawlStatus(ctx, resp.ID)

// Cancel
_, err = client.CancelCrawl(ctx, resp.ID)

// Get errors
errors, err := client.GetCrawlErrors(ctx, resp.ID)
```

### Batch Scrape

Scrape multiple URLs in a single batch job.

```go
urls := []string{
	"https://example.com/page1",
	"https://example.com/page2",
	"https://example.com/page3",
}

// Auto-polling
job, err := client.BatchScrape(ctx, urls, &zapfetch.BatchScrapeOptions{
	ScrapeOptions: &zapfetch.ScrapeOptions{
		Formats: []string{"markdown"},
	},
})

// Or manage manually
resp, err := client.StartBatchScrape(ctx, urls, nil)
status, err := client.GetBatchScrapeStatus(ctx, resp.ID)
_, err = client.CancelBatchScrape(ctx, resp.ID)
```

### Map

Discover URLs on a website.

```go
mapData, err := client.Map(ctx, "https://example.com", &zapfetch.MapOptions{
	Search:            zapfetch.String("pricing"),
	IncludeSubdomains: zapfetch.Bool(true),
	Limit:             zapfetch.Int(100),
})
```

### Search

Search the web and optionally scrape results inline.

```go
results, err := client.Search(ctx, "zapfetch web scraping", &zapfetch.SearchOptions{
	Limit: zapfetch.Int(5),
	ScrapeOptions: &zapfetch.ScrapeOptions{
		Formats: []string{"markdown"},
	},
})
```

### Agent

Run an AI-powered agent to extract structured data.

```go
// Auto-polling
status, err := client.Agent(ctx, &zapfetch.AgentOptions{
	Prompt: "Find all pricing plans and their features",
	URLs:   []string{"https://example.com/pricing"},
	Schema: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"plans": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name":  map[string]interface{}{"type": "string"},
						"price": map[string]interface{}{"type": "string"},
					},
				},
			},
		},
	},
})

// Or manage manually
resp, err := client.StartAgent(ctx, &zapfetch.AgentOptions{
	Prompt: "Extract product information",
})
status, err := client.GetAgentStatus(ctx, resp.ID)
_, err = client.CancelAgent(ctx, resp.ID)
```

### Browser

Create and manage standalone browser sessions.

```go
session, err := client.Browser(ctx, &zapfetch.BrowserOptions{
	TTL:           zapfetch.Int(300),
	StreamWebView: zapfetch.Bool(true),
})

result, err := client.BrowserExecute(ctx, session.ID, "echo 'hello'", &zapfetch.BrowserExecuteParams{
	Language: "bash",
	Timeout:  zapfetch.Int(30),
})

list, err := client.ListBrowsers(ctx, "active")
_, err = client.DeleteBrowser(ctx, session.ID)
```

### Usage & Metrics

```go
// Check concurrency
concurrency, err := client.GetConcurrency(ctx)
fmt.Printf("Using %d of %d\n", concurrency.Concurrency, concurrency.MaxConcurrency)

// Check credit usage
credits, err := client.GetCreditUsage(ctx)
fmt.Printf("Remaining: %d of %d\n", credits.RemainingCredits, credits.PlanCredits)
```

## Error Handling

The SDK uses typed errors for different failure scenarios. `APIError` is the
base type; `AuthenticationError`, `RateLimitError`, and `JobTimeoutError` embed
it for errors.As discrimination.

```go
doc, err := client.Scrape(ctx, "https://example.com", nil)
if err != nil {
	var authErr *zapfetch.AuthenticationError
	var rateErr *zapfetch.RateLimitError
	var timeoutErr *zapfetch.JobTimeoutError
	var apiErr *zapfetch.APIError

	switch {
	case errors.As(err, &authErr):
		fmt.Println("Invalid API key:", authErr.Message)
	case errors.As(err, &rateErr):
		fmt.Println("Rate limited:", rateErr.Message)
	case errors.As(err, &timeoutErr):
		fmt.Printf("Job %s timed out after %ds\n", timeoutErr.JobID, timeoutErr.TimeoutSeconds)
	case errors.As(err, &apiErr):
		fmt.Printf("API error (HTTP %d): %s\n", apiErr.StatusCode, apiErr.Message)
	default:
		fmt.Println("Unexpected error:", err)
	}
}
```

### Retry Logic

The SDK automatically retries transient failures:

- **Retried:** 408, 409, 5xx, and connection failures
- **Not retried:** 401, 429, and other 4xx errors
- **Backoff:** Exponential, with configurable factor

## Context Support

All methods accept a `context.Context` for cancellation and deadline control:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

doc, err := client.Scrape(ctx, "https://example.com", nil)
```

## Pointer Helpers

Convenience functions for optional pointer fields:

```go
zapfetch.Bool(true)     // *bool
zapfetch.Int(50)        // *int
zapfetch.Int64(1000)    // *int64
zapfetch.String("test") // *string
zapfetch.Float64(0.5)   // *float64
```

## Releases

This SDK lives in a monorepo subdirectory, so Go module releases follow the
[nested module tagging](https://go.dev/ref/mod#vcs-version) convention. Tags
**must** be prefixed with the module subdirectory path:

```
go-sdk/v1.0.0
```

A bare `v1.0.0` tag is not resolvable by the Go module proxy from this repo.

### Consuming a specific version

```bash
go get github.com/zapfetch/sdks/go-sdk@v1.0.0
```

Users pin via the semantic version suffix only; Go's toolchain translates that
to the full `go-sdk/v1.0.0` tag under the hood.

## License

MIT
