// Batch-scrape a list of URLs and print titles + credit usage.
//
// Usage:
//
//	ZAPFETCH_API_KEY=zf-... go run .
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/zapfetch/sdks/go-sdk"
)

func main() {
	client, err := zapfetch.NewClient()
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	urls := []string{
		"https://example.com",
		"https://example.org",
		"https://example.net",
	}

	ctx := context.Background()
	job, err := client.BatchScrape(ctx, urls, &zapfetch.BatchScrapeOptions{
		ScrapeOptions: &zapfetch.ScrapeOptions{
			Formats:         []string{"markdown"},
			OnlyMainContent: zapfetch.Bool(true),
		},
		MaxConcurrency: zapfetch.Int(3),
	})
	if err != nil {
		log.Fatalf("batch scrape failed: %v", err)
	}

	fmt.Printf("batch status=%s completed=%d/%d\n", job.Status, job.Completed, job.Total)
	for i, doc := range job.Data {
		title, _ := doc.Metadata["title"].(string)
		sourceURL, _ := doc.Metadata["sourceURL"].(string)
		fmt.Printf("[%d] %s\n    %s\n", i+1, title, sourceURL)
	}
	if job.CreditsUsed != nil {
		fmt.Printf("credits used: %d\n", *job.CreditsUsed)
	}
}
