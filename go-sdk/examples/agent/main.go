// Run an agent task that extracts structured pricing data.
//
// Usage:
//
//	ZAPFETCH_API_KEY=zf-... go run .
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/zapfetch/sdks/go-sdk"
)

func main() {
	client, err := zapfetch.NewClient()
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"plans": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name":     map[string]interface{}{"type": "string"},
						"price":    map[string]interface{}{"type": "string"},
						"features": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					},
					"required": []string{"name", "price"},
				},
			},
		},
		"required": []string{"plans"},
	}

	ctx := context.Background()
	status, err := client.Agent(ctx, &zapfetch.AgentOptions{
		URLs:       []string{"https://example.com/pricing"},
		Prompt:     "Extract each pricing plan with its monthly price and headline features.",
		Schema:     schema,
		MaxCredits: zapfetch.Int(50),
	})
	if err != nil {
		log.Fatalf("agent task failed: %v", err)
	}

	fmt.Printf("agent status=%s\n", status.Status)
	pretty, _ := json.MarshalIndent(status.Data, "", "  ")
	fmt.Println(string(pretty))
}
