package option

import (
	"net/http"
	"testing"
	"time"
)

func apply(opts ...RequestOption) *RequestConfig {
	c := &RequestConfig{}
	for _, o := range opts {
		o(c)
	}
	return c
}

func TestWithAPIKey(t *testing.T) {
	c := apply(WithAPIKey("zf-key"))
	if c.APIKey != "zf-key" {
		t.Fatalf("APIKey = %q, want zf-key", c.APIKey)
	}
}

func TestWithAPIURL(t *testing.T) {
	c := apply(WithAPIURL("https://custom.zapfetch.com"))
	if c.APIURL != "https://custom.zapfetch.com" {
		t.Fatalf("APIURL = %q", c.APIURL)
	}
}

func TestWithMaxRetries(t *testing.T) {
	c := apply(WithMaxRetries(7))
	if c.MaxRetries != 7 {
		t.Fatalf("MaxRetries = %d, want 7", c.MaxRetries)
	}
}

func TestWithBackoffFactor(t *testing.T) {
	c := apply(WithBackoffFactor(1.5))
	if c.BackoffFactor != 1.5 {
		t.Fatalf("BackoffFactor = %v, want 1.5", c.BackoffFactor)
	}
}

func TestWithTimeout_NoExistingClient(t *testing.T) {
	c := apply(WithTimeout(30 * time.Second))
	if c.HTTPClient == nil {
		t.Fatal("HTTPClient should be created")
	}
	if c.HTTPClient.Timeout != 30*time.Second {
		t.Fatalf("Timeout = %v, want 30s", c.HTTPClient.Timeout)
	}
}

func TestWithTimeout_ExistingClient(t *testing.T) {
	existing := &http.Client{Timeout: time.Minute}
	c := apply(WithHTTPClient(existing), WithTimeout(10*time.Second))
	if c.HTTPClient != existing {
		t.Fatal("WithTimeout should mutate the existing HTTPClient, not replace it")
	}
	if c.HTTPClient.Timeout != 10*time.Second {
		t.Fatalf("Timeout = %v, want 10s", c.HTTPClient.Timeout)
	}
}

func TestWithHeader_Multiple(t *testing.T) {
	c := apply(
		WithHeader("X-A", "a"),
		WithHeader("X-B", "b"),
	)
	if got := c.ExtraHeaders["X-A"]; got != "a" {
		t.Errorf("X-A = %q", got)
	}
	if got := c.ExtraHeaders["X-B"]; got != "b" {
		t.Errorf("X-B = %q", got)
	}
}

func TestWithHeader_Overwrite(t *testing.T) {
	c := apply(
		WithHeader("X-A", "v1"),
		WithHeader("X-A", "v2"),
	)
	if got := c.ExtraHeaders["X-A"]; got != "v2" {
		t.Errorf("X-A = %q, want v2 (last write wins)", got)
	}
}
