// Discover URLs on a site using the map endpoint.
//
// Usage:
//
//	ZAPFETCH_API_KEY=zf-... go run . https://example.com pricing
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

	opts := &zapfetch.MapOptions{
		Limit:             zapfetch.Int(50),
		IncludeSubdomains: zapfetch.Bool(true),
	}
	if len(os.Args) > 2 {
		opts.Search = zapfetch.String(os.Args[2])
	}

	ctx := context.Background()
	data, err := client.Map(ctx, target, opts)
	if err != nil {
		log.Fatalf("map failed: %v", err)
	}

	fmt.Printf("%d links discovered on %s\n", len(data.Links), target)
	for i, link := range data.Links {
		fmt.Printf("[%d] %s\n", i+1, link.URL)
	}
}
