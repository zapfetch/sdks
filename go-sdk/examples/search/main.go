// Search the web and scrape each result inline.
//
// Usage:
//
//	ZAPFETCH_API_KEY=zf-... go run . "web scraping tools"
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/zapfetch/sdks/go-sdk"
)

func main() {
	client, err := zapfetch.NewClient()
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	query := "web scraping tools"
	if len(os.Args) > 1 {
		query = strings.Join(os.Args[1:], " ")
	}

	ctx := context.Background()
	results, err := client.Search(ctx, query, &zapfetch.SearchOptions{
		Limit: zapfetch.Int(5),
		ScrapeOptions: &zapfetch.ScrapeOptions{
			Formats:         []string{"markdown"},
			OnlyMainContent: zapfetch.Bool(true),
		},
	})
	if err != nil {
		log.Fatalf("search failed: %v", err)
	}

	fmt.Printf("query=%q web=%d news=%d images=%d\n", query, len(results.Web), len(results.News), len(results.Images))
	for i, r := range results.Web {
		title, _ := r["title"].(string)
		url, _ := r["url"].(string)
		fmt.Printf("[%d] %s\n    %s\n", i+1, title, url)
	}
}
