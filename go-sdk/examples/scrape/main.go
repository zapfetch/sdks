// Scrape a single page and print its markdown.
//
// Usage:
//
//	ZAPFETCH_API_KEY=zf-... go run .
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

	ctx := context.Background()
	targetURL := "https://example.com"
	if len(os.Args) > 1 {
		targetURL = os.Args[1]
	}

	doc, err := client.Scrape(ctx, targetURL, &zapfetch.ScrapeOptions{
		Formats:         []string{"markdown"},
		OnlyMainContent: zapfetch.Bool(true),
	})
	if err != nil {
		log.Fatalf("scrape failed: %v", err)
	}

	fmt.Printf("--- %s ---\n%s\n", targetURL, doc.Markdown)
}
