package zapfetch

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/zapfetch/sdks/go-sdk/option"
)

// newTestClient builds a Client that points at httptest.Server with a fast
// retry profile so test suites don't sleep for seconds.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	c, err := NewClient(
		option.WithAPIKey("zf-test"),
		option.WithAPIURL(srv.URL),
		option.WithMaxRetries(2),
		option.WithBackoffFactor(0.0), // instant retries
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c, srv
}

func TestScrape_SendsAuthAndUserAgent(t *testing.T) {
	var gotAuth, gotUA, gotContentType string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotUA = r.Header.Get("User-Agent")
		gotContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"url":"https://example.com"`) {
			t.Errorf("body missing url field: %s", string(body))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"markdown": "# ok"},
		})
	})

	doc, err := c.Scrape(context.Background(), "https://example.com", nil)
	if err != nil {
		t.Fatalf("Scrape: %v", err)
	}
	if doc.Markdown != "# ok" {
		t.Fatalf("Markdown = %q, want %q", doc.Markdown, "# ok")
	}
	if gotAuth != "Bearer zf-test" {
		t.Errorf("Authorization = %q, want Bearer zf-test", gotAuth)
	}
	if !strings.HasPrefix(gotUA, "zapfetch-go/") {
		t.Errorf("User-Agent = %q, want zapfetch-go/ prefix", gotUA)
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", gotContentType)
	}
}

func TestScrape_401ReturnsAuthenticationError(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "bad key"})
	})

	_, err := c.Scrape(context.Background(), "https://example.com", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var authErr *AuthenticationError
	if !errors.As(err, &authErr) {
		t.Fatalf("error type = %T, want *AuthenticationError", err)
	}
	if authErr.StatusCode != 401 {
		t.Errorf("StatusCode = %d, want 401", authErr.StatusCode)
	}
}

func TestScrape_429ReturnsRateLimitError(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "slow down"})
	})

	_, err := c.Scrape(context.Background(), "https://example.com", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var rateErr *RateLimitError
	if !errors.As(err, &rateErr) {
		t.Fatalf("error type = %T, want *RateLimitError", err)
	}
}

func TestScrape_Retries5xxThenSucceeds(t *testing.T) {
	var calls int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "flap"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"markdown": "# recovered"},
		})
	})

	doc, err := c.Scrape(context.Background(), "https://example.com", nil)
	if err != nil {
		t.Fatalf("Scrape: %v", err)
	}
	if doc.Markdown != "# recovered" {
		t.Fatalf("Markdown = %q, want recovered", doc.Markdown)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("call count = %d, want 3", got)
	}
}

func TestScrape_400NoRetry(t *testing.T) {
	var calls int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "bad"})
	})

	_, err := c.Scrape(context.Background(), "https://example.com", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 400 {
		t.Fatalf("err = %v, want *APIError with status 400", err)
	}
	// 400 must not retry.
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("call count = %d, want 1 (no retry on 400)", got)
	}
}

func TestBatchScrape_SetsIdempotencyKeyHeader(t *testing.T) {
	var gotHeader string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("x-idempotency-key")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "job-1", "url": "http://example",
		})
	})

	_, err := c.StartBatchScrape(context.Background(), []string{"https://example.com"}, &BatchScrapeOptions{
		IdempotencyKey: String("abc-123"),
	})
	if err != nil {
		t.Fatalf("StartBatchScrape: %v", err)
	}
	if gotHeader != "abc-123" {
		t.Fatalf("idempotency header = %q, want abc-123", gotHeader)
	}
}

func TestWithHeader_AppliesExtraHeader(t *testing.T) {
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Request-Source")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"markdown": "ok"},
		})
	}))
	t.Cleanup(srv.Close)

	c, err := NewClient(
		option.WithAPIKey("zf-x"),
		option.WithAPIURL(srv.URL),
		option.WithHeader("X-Request-Source", "tests"),
		option.WithBackoffFactor(0.0),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if _, err := c.Scrape(context.Background(), "https://example.com", nil); err != nil {
		t.Fatalf("Scrape: %v", err)
	}
	if gotHeader != "tests" {
		t.Fatalf("X-Request-Source = %q, want tests", gotHeader)
	}
}

func TestScrape_ContextCancelStopsRetries(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	c, err := NewClient(
		option.WithAPIKey("zf-x"),
		option.WithAPIURL(srv.URL),
		option.WithMaxRetries(5),
		option.WithBackoffFactor(0.0),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	_, err = c.Scrape(ctx, "https://example.com", nil)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	// Either context.Canceled surfaces directly, or an APIError wraps the last
	// response — both are acceptable; what matters is we didn't hang forever.
	if !errors.Is(err, context.Canceled) {
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("err = %v (%T), want context.Canceled or *APIError", err, err)
		}
	}
}
