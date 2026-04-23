package zapfetch

import (
	"errors"
	"testing"

	"github.com/zapfetch/sdks/go-sdk/option"
)

func TestNewClient_APIKeyFromOption(t *testing.T) {
	t.Setenv("ZAPFETCH_API_KEY", "")
	c, err := NewClient(option.WithAPIKey("zf-explicit"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.http.apiKey != "zf-explicit" {
		t.Fatalf("apiKey = %q, want %q", c.http.apiKey, "zf-explicit")
	}
}

func TestNewClient_APIKeyFromEnv(t *testing.T) {
	t.Setenv("ZAPFETCH_API_KEY", "zf-from-env")
	c, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.http.apiKey != "zf-from-env" {
		t.Fatalf("apiKey = %q, want %q", c.http.apiKey, "zf-from-env")
	}
}

func TestNewClient_OptionOverridesEnv(t *testing.T) {
	t.Setenv("ZAPFETCH_API_KEY", "zf-env-loses")
	c, err := NewClient(option.WithAPIKey("zf-option-wins"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.http.apiKey != "zf-option-wins" {
		t.Fatalf("apiKey = %q, want %q", c.http.apiKey, "zf-option-wins")
	}
}

func TestNewClient_MissingAPIKey(t *testing.T) {
	t.Setenv("ZAPFETCH_API_KEY", "")
	_, err := NewClient()
	if err == nil {
		t.Fatal("expected error when API key is missing")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *APIError", err)
	}
}

func TestNewClient_DefaultAPIURL(t *testing.T) {
	t.Setenv("ZAPFETCH_API_URL", "")
	c, err := NewClient(option.WithAPIKey("zf-x"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.http.baseURL != defaultAPIURL {
		t.Fatalf("baseURL = %q, want %q", c.http.baseURL, defaultAPIURL)
	}
	if defaultAPIURL != "https://api.zapfetch.com" {
		t.Fatalf("defaultAPIURL = %q, want https://api.zapfetch.com", defaultAPIURL)
	}
}

func TestNewClient_APIURLFromEnv(t *testing.T) {
	t.Setenv("ZAPFETCH_API_URL", "https://staging.zapfetch.com/")
	c, err := NewClient(option.WithAPIKey("zf-x"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Trailing slash should be trimmed.
	if c.http.baseURL != "https://staging.zapfetch.com" {
		t.Fatalf("baseURL = %q, want trimmed staging URL", c.http.baseURL)
	}
}

func TestNewClient_OptionAPIURLOverridesEnv(t *testing.T) {
	t.Setenv("ZAPFETCH_API_URL", "https://env.example.com")
	c, err := NewClient(
		option.WithAPIKey("zf-x"),
		option.WithAPIURL("https://option.example.com"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.http.baseURL != "https://option.example.com" {
		t.Fatalf("baseURL = %q, want option override", c.http.baseURL)
	}
}

func TestPointerHelpers(t *testing.T) {
	if *Bool(true) != true {
		t.Error("Bool(true) did not round-trip")
	}
	if *Int(42) != 42 {
		t.Error("Int(42) did not round-trip")
	}
	if *Int64(1<<40) != 1<<40 {
		t.Error("Int64 did not round-trip")
	}
	if *String("hi") != "hi" {
		t.Error("String(\"hi\") did not round-trip")
	}
	if *Float64(0.5) != 0.5 {
		t.Error("Float64(0.5) did not round-trip")
	}
}
