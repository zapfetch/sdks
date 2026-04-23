// Crawl a domain with depth and page-count limits, then summarise each result.
//
// Usage:
//
//	ZAPFETCH_API_KEY=zf-... go run . https://example.com
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/zapfetch/sdks/go-sdk"
)

func main() {
	client, err := zapfetch.NewClient()
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	target := "https://example.com"
	if len(os.Args) > 1 {
		target = os.Args[1]
	}

	ctx := context.Background()
	job, err := client.Crawl(ctx, target, &zapfetch.CrawlOptions{
		Limit:             zapfetch.Int(10),
		MaxDiscoveryDepth: zapfetch.Int(2),
		ScrapeOptions: &zapfetch.ScrapeOptions{
			Formats:         []string{"markdown"},
			OnlyMainContent: zapfetch.Bool(true),
		},
	})
	if err != nil {
		log.Fatalf("crawl failed: %v", err)
	}

	fmt.Printf("crawl status=%s completed=%d/%d\n", job.Status, job.Completed, job.Total)
	for i, doc := range job.Data {
		title := ""
		if t, ok := doc.Metadata["title"].(string); ok {
			title = t
		}
		fmt.Printf("[%d] %s\n    %d chars of markdown\n", i+1, title, len(doc.Markdown))
	}
}
