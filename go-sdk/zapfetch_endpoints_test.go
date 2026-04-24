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
	"time"

	"github.com/zapfetch/sdks/go-sdk/option"
)

// fastClient is like newTestClient but lets tests override maxRetries / backoff
// to keep polling tests fast. Reused here to avoid duplicating server setup.
func fastClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	c, err := NewClient(
		option.WithAPIKey("zf-test"),
		option.WithAPIURL(srv.URL),
		option.WithMaxRetries(1),
		option.WithBackoffFactor(0.0),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c, srv
}

// ============================================================
// CRAWL
// ============================================================

func TestStartCrawl_Success(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/crawl" {
			t.Errorf("method/path = %s %s, want POST /v2/crawl", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"url":"https://example.com"`) {
			t.Errorf("body missing url: %s", string(body))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "crawl-1", "url": "https://example.com"})
	})

	resp, err := c.StartCrawl(context.Background(), "https://example.com", nil)
	if err != nil {
		t.Fatalf("StartCrawl: %v", err)
	}
	if resp.ID != "crawl-1" {
		t.Fatalf("ID = %q, want crawl-1", resp.ID)
	}
}

func TestStartCrawl_EmptyURL(t *testing.T) {
	c, _ := fastClient(t, nil)
	_, err := c.StartCrawl(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected error on empty URL")
	}
}

func TestStartCrawl_WithOptions(t *testing.T) {
	var gotBody string
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "x"})
	})

	_, err := c.StartCrawl(context.Background(), "https://example.com", &CrawlOptions{
		Limit:             Int(50),
		CrawlEntireDomain: Bool(true),
	})
	if err != nil {
		t.Fatalf("StartCrawl: %v", err)
	}
	if !strings.Contains(gotBody, `"limit":50`) {
		t.Errorf("body missing limit: %s", gotBody)
	}
	if !strings.Contains(gotBody, `"crawlEntireDomain":true`) {
		t.Errorf("body missing crawlEntireDomain: %s", gotBody)
	}
}

func TestGetCrawlStatus_Success(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v2/crawl/job-abc" {
			t.Errorf("method/path = %s %s, want GET /v2/crawl/job-abc", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":        "job-abc",
			"status":    "completed",
			"total":     2,
			"completed": 2,
			"data":      []map[string]any{{"markdown": "# page1"}, {"markdown": "# page2"}},
		})
	})

	job, err := c.GetCrawlStatus(context.Background(), "job-abc")
	if err != nil {
		t.Fatalf("GetCrawlStatus: %v", err)
	}
	if !job.IsDone() || len(job.Data) != 2 {
		t.Fatalf("job not as expected: %+v", job)
	}
}

func TestGetCrawlStatus_EmptyJobID(t *testing.T) {
	c, _ := fastClient(t, nil)
	if _, err := c.GetCrawlStatus(context.Background(), ""); err == nil {
		t.Fatal("expected error on empty jobID")
	}
}

func TestCrawlWithPolling_StateMachineToCompletion(t *testing.T) {
	var calls int32
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v2/crawl":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "j1"})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/crawl/j1":
			n := atomic.AddInt32(&calls, 1)
			if n < 2 {
				_ = json.NewEncoder(w).Encode(map[string]any{"id": "j1", "status": "scraping"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":     "j1",
				"status": "completed",
				"total":  1, "completed": 1,
				"data": []map[string]any{{"markdown": "# done"}},
			})
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	})

	// 0s poll interval to avoid waiting in CI.
	job, err := c.CrawlWithPolling(context.Background(), "https://ex.com", nil, 0, 10)
	if err != nil {
		t.Fatalf("CrawlWithPolling: %v", err)
	}
	if !job.IsDone() || len(job.Data) != 1 {
		t.Fatalf("unexpected job: %+v", job)
	}
}

func TestCrawlWithPolling_Timeout(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "timeout-job"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "timeout-job", "status": "scraping"})
	})

	// 0s timeout triggers the deadline path immediately.
	_, err := c.CrawlWithPolling(context.Background(), "https://ex.com", nil, 0, 0)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if _, ok := err.(*JobTimeoutError); !ok {
		t.Fatalf("err type = %T, want *JobTimeoutError", err)
	}
}

func TestCrawl_AutoPollingDefaults(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "j"})
			return
		}
		// Return done on first poll so Crawl() returns quickly even with default interval.
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "j", "status": "completed"})
	})

	job, err := c.Crawl(context.Background(), "https://ex.com", nil)
	if err != nil {
		t.Fatalf("Crawl: %v", err)
	}
	if !job.IsDone() {
		t.Fatalf("job not done: %+v", job)
	}
}

func TestCancelCrawl_Success(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v2/crawl/x" {
			t.Errorf("%s %s, want DELETE /v2/crawl/x", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
	})
	resp, err := c.CancelCrawl(context.Background(), "x")
	if err != nil {
		t.Fatalf("CancelCrawl: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("resp: %+v", resp)
	}
}

func TestCancelCrawl_EmptyJobID(t *testing.T) {
	c, _ := fastClient(t, nil)
	if _, err := c.CancelCrawl(context.Background(), ""); err == nil {
		t.Fatal("expected error on empty jobID")
	}
}

func TestGetCrawlErrors_Success(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/crawl/x/errors" {
			t.Errorf("path = %s, want /v2/crawl/x/errors", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"errors": []any{}, "robotsBlocked": []any{}})
	})
	resp, err := c.GetCrawlErrors(context.Background(), "x")
	if err != nil {
		t.Fatalf("GetCrawlErrors: %v", err)
	}
	if _, ok := resp["errors"]; !ok {
		t.Fatalf("missing errors key: %+v", resp)
	}
}

func TestGetCrawlErrors_EmptyJobID(t *testing.T) {
	c, _ := fastClient(t, nil)
	if _, err := c.GetCrawlErrors(context.Background(), ""); err == nil {
		t.Fatal("expected error on empty jobID")
	}
}

// ============================================================
// BATCH SCRAPE
// ============================================================

func TestStartBatchScrape_EmptyURLs(t *testing.T) {
	c, _ := fastClient(t, nil)
	if _, err := c.StartBatchScrape(context.Background(), nil, nil); err == nil {
		t.Fatal("expected error on empty URLs")
	}
}

func TestStartBatchScrape_NestedScrapeOptionsFlattened(t *testing.T) {
	// StartBatchScrape has special handling: it flattens `ScrapeOptions` (which
	// marshals to `"options":{...}`) up to the top level, so the API sees
	// formats/onlyMainContent/etc directly on the request body.
	var gotBody string
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "b-nested"})
	})

	_, err := c.StartBatchScrape(context.Background(), []string{"https://a"}, &BatchScrapeOptions{
		ScrapeOptions: &ScrapeOptions{
			Formats:         []string{"markdown"},
			OnlyMainContent: Bool(true),
		},
		MaxConcurrency: Int(4),
	})
	if err != nil {
		t.Fatalf("StartBatchScrape: %v", err)
	}
	// After flatten: formats and onlyMainContent appear at top level (not under "options").
	if strings.Contains(gotBody, `"options":`) {
		t.Errorf("nested options not flattened: %s", gotBody)
	}
	for _, want := range []string{`"formats":["markdown"]`, `"onlyMainContent":true`, `"maxConcurrency":4`, `"urls":["https://a"]`} {
		if !strings.Contains(gotBody, want) {
			t.Errorf("body missing %s: %s", want, gotBody)
		}
	}
}

func TestGetBatchScrapeStatus_Success(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/batch/scrape/b1" {
			t.Errorf("path = %s, want /v2/batch/scrape/b1", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "b1", "status": "completed", "total": 1, "completed": 1,
			"data": []map[string]any{{"markdown": "# x"}},
		})
	})
	job, err := c.GetBatchScrapeStatus(context.Background(), "b1")
	if err != nil {
		t.Fatalf("GetBatchScrapeStatus: %v", err)
	}
	if !job.IsDone() {
		t.Fatalf("not done: %+v", job)
	}
}

func TestGetBatchScrapeStatus_EmptyJobID(t *testing.T) {
	c, _ := fastClient(t, nil)
	if _, err := c.GetBatchScrapeStatus(context.Background(), ""); err == nil {
		t.Fatal("expected error")
	}
}

func TestBatchScrapeWithPolling_StateMachine(t *testing.T) {
	var calls int32
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost:
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "b"})
		case r.Method == http.MethodGet:
			n := atomic.AddInt32(&calls, 1)
			if n < 2 {
				_ = json.NewEncoder(w).Encode(map[string]any{"id": "b", "status": "scraping"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "b", "status": "completed", "total": 1, "completed": 1,
				"data": []map[string]any{{"markdown": "# ok"}},
			})
		}
	})

	job, err := c.BatchScrapeWithPolling(context.Background(), []string{"https://a"}, nil, 0, 10)
	if err != nil {
		t.Fatalf("BatchScrapeWithPolling: %v", err)
	}
	if !job.IsDone() || len(job.Data) != 1 {
		t.Fatalf("unexpected: %+v", job)
	}
}

func TestBatchScrapeWithPolling_Timeout(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "to"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "to", "status": "scraping"})
	})
	_, err := c.BatchScrapeWithPolling(context.Background(), []string{"https://a"}, nil, 0, 0)
	if err == nil {
		t.Fatal("expected timeout")
	}
	if _, ok := err.(*JobTimeoutError); !ok {
		t.Fatalf("err = %T, want *JobTimeoutError", err)
	}
}

func TestBatchScrape_AutoPolling(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "b"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "b", "status": "completed"})
	})
	job, err := c.BatchScrape(context.Background(), []string{"https://a"}, nil)
	if err != nil || !job.IsDone() {
		t.Fatalf("BatchScrape: err=%v job=%+v", err, job)
	}
}

func TestCancelBatchScrape_Success(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v2/batch/scrape/b" {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
	})
	resp, err := c.CancelBatchScrape(context.Background(), "b")
	if err != nil {
		t.Fatalf("CancelBatchScrape: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("%+v", resp)
	}
}

func TestCancelBatchScrape_EmptyJobID(t *testing.T) {
	c, _ := fastClient(t, nil)
	if _, err := c.CancelBatchScrape(context.Background(), ""); err == nil {
		t.Fatal("expected error")
	}
}

// ============================================================
// MAP
// ============================================================

func TestMap_Success(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/map" || r.Method != http.MethodPost {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"links": []any{"https://a.com/1", "https://a.com/2"}},
		})
	})
	data, err := c.Map(context.Background(), "https://a.com", nil)
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	if len(data.Links) != 2 || data.Links[0].URL != "https://a.com/1" {
		t.Fatalf("links: %+v", data.Links)
	}
}

func TestMap_LinksAsObjects(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"links": []any{
				map[string]any{"url": "https://x.com", "title": "X"},
			}},
		})
	})
	data, err := c.Map(context.Background(), "https://x", nil)
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	if data.Links[0].Title != "X" {
		t.Fatalf("Title = %q", data.Links[0].Title)
	}
}

func TestMap_EmptyURL(t *testing.T) {
	c, _ := fastClient(t, nil)
	if _, err := c.Map(context.Background(), "", nil); err == nil {
		t.Fatal("expected error")
	}
}

// ============================================================
// SEARCH
// ============================================================

func TestSearch_Success(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/search" {
			t.Errorf("path = %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"query":"go lang"`) {
			t.Errorf("body: %s", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"web": []any{map[string]any{"title": "Go"}}},
		})
	})
	data, err := c.Search(context.Background(), "go lang", nil)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(data.Web) != 1 {
		t.Fatalf("web: %+v", data.Web)
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	c, _ := fastClient(t, nil)
	if _, err := c.Search(context.Background(), "", nil); err == nil {
		t.Fatal("expected error")
	}
}

// ============================================================
// AGENT
// ============================================================

func TestStartAgent_NilOpts(t *testing.T) {
	c, _ := fastClient(t, nil)
	if _, err := c.StartAgent(context.Background(), nil); err == nil {
		t.Fatal("expected error on nil opts")
	}
}

func TestStartAgent_Success(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/agent" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "id": "a-1"})
	})
	resp, err := c.StartAgent(context.Background(), &AgentOptions{Prompt: "Test"})
	if err != nil {
		t.Fatalf("StartAgent: %v", err)
	}
	if resp.ID != "a-1" {
		t.Fatalf("ID = %q", resp.ID)
	}
}

func TestGetAgentStatus_Success(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true, "status": "completed", "data": "done",
		})
	})
	resp, err := c.GetAgentStatus(context.Background(), "a-1")
	if err != nil {
		t.Fatalf("GetAgentStatus: %v", err)
	}
	if !resp.IsDone() {
		t.Fatalf("not done: %+v", resp)
	}
}

func TestGetAgentStatus_EmptyJobID(t *testing.T) {
	c, _ := fastClient(t, nil)
	if _, err := c.GetAgentStatus(context.Background(), ""); err == nil {
		t.Fatal("expected error")
	}
}

func TestAgentWithPolling_MissingID(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		// Return success but no ID — triggers the "did not return a job ID" branch.
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
	})
	_, err := c.AgentWithPolling(context.Background(), &AgentOptions{Prompt: "x"}, 0, 10)
	if err == nil {
		t.Fatal("expected error on missing job ID")
	}
}

func TestAgentWithPolling_StateMachine(t *testing.T) {
	var calls int32
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "id": "a"})
			return
		}
		n := atomic.AddInt32(&calls, 1)
		if n < 2 {
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "scraping"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "completed", "data": "ok"})
	})
	resp, err := c.AgentWithPolling(context.Background(), &AgentOptions{Prompt: "p"}, 0, 10)
	if err != nil {
		t.Fatalf("AgentWithPolling: %v", err)
	}
	if !resp.IsDone() {
		t.Fatalf("not done: %+v", resp)
	}
}

func TestAgent_AutoPolling(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "id": "a"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "completed"})
	})
	resp, err := c.Agent(context.Background(), &AgentOptions{Prompt: "p"})
	if err != nil || !resp.IsDone() {
		t.Fatalf("Agent: %v, %+v", err, resp)
	}
}

func TestAgentWithPolling_Timeout(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "id": "a"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "scraping"})
	})
	_, err := c.AgentWithPolling(context.Background(), &AgentOptions{Prompt: "p"}, 0, 0)
	if err == nil {
		t.Fatal("expected timeout")
	}
	if _, ok := err.(*JobTimeoutError); !ok {
		t.Fatalf("err = %T, want *JobTimeoutError", err)
	}
}

func TestCancelAgent_Success(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v2/agent/a" {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
	})
	if _, err := c.CancelAgent(context.Background(), "a"); err != nil {
		t.Fatalf("CancelAgent: %v", err)
	}
}

func TestCancelAgent_EmptyJobID(t *testing.T) {
	c, _ := fastClient(t, nil)
	if _, err := c.CancelAgent(context.Background(), ""); err == nil {
		t.Fatal("expected error")
	}
}

// ============================================================
// BROWSER
// ============================================================

func TestBrowser_NilOpts(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/browser" {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "id": "s1"})
	})
	resp, err := c.Browser(context.Background(), nil)
	if err != nil {
		t.Fatalf("Browser: %v", err)
	}
	if resp.ID != "s1" {
		t.Fatalf("ID = %q", resp.ID)
	}
}

func TestBrowser_WithOpts(t *testing.T) {
	var gotBody string
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "s"})
	})
	_, err := c.Browser(context.Background(), &BrowserOptions{
		TTL:           Int(60),
		ActivityTTL:   Int(30),
		StreamWebView: Bool(true),
	})
	if err != nil {
		t.Fatalf("Browser: %v", err)
	}
	for _, want := range []string{`"ttl":60`, `"activityTtl":30`, `"streamWebView":true`} {
		if !strings.Contains(gotBody, want) {
			t.Errorf("body missing %s: %s", want, gotBody)
		}
	}
}

func TestBrowserExecute_EmptySessionID(t *testing.T) {
	c, _ := fastClient(t, nil)
	if _, err := c.BrowserExecute(context.Background(), "", "ls", nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestBrowserExecute_EmptyCode(t *testing.T) {
	c, _ := fastClient(t, nil)
	if _, err := c.BrowserExecute(context.Background(), "s", "", nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestBrowserExecute_Success(t *testing.T) {
	var gotBody string
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/browser/s1/execute" {
			t.Errorf("path = %s", r.URL.Path)
		}
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "stdout": "hi"})
	})
	resp, err := c.BrowserExecute(context.Background(), "s1", "echo hi", &BrowserExecuteParams{
		Language: "python",
		Timeout:  Int(10),
	})
	if err != nil {
		t.Fatalf("BrowserExecute: %v", err)
	}
	if resp.Stdout != "hi" {
		t.Fatalf("stdout = %q", resp.Stdout)
	}
	if !strings.Contains(gotBody, `"language":"python"`) || !strings.Contains(gotBody, `"timeout":10`) {
		t.Errorf("body missing overrides: %s", gotBody)
	}
}

func TestDeleteBrowser_EmptySessionID(t *testing.T) {
	c, _ := fastClient(t, nil)
	if _, err := c.DeleteBrowser(context.Background(), ""); err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteBrowser_Success(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v2/browser/s1" {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
	})
	if _, err := c.DeleteBrowser(context.Background(), "s1"); err != nil {
		t.Fatalf("DeleteBrowser: %v", err)
	}
}

func TestListBrowsers_NoFilter(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/browser" || r.URL.RawQuery != "" {
			t.Errorf("path/query = %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "sessions": []any{}})
	})
	if _, err := c.ListBrowsers(context.Background(), ""); err != nil {
		t.Fatalf("ListBrowsers: %v", err)
	}
}

func TestListBrowsers_StatusFilter(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "status=active" {
			t.Errorf("query = %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
	})
	if _, err := c.ListBrowsers(context.Background(), "active"); err != nil {
		t.Fatalf("ListBrowsers: %v", err)
	}
}

// ============================================================
// INTERACT / STOP INTERACTIVE BROWSER
// ============================================================

func TestInteract_EmptyJobID(t *testing.T) {
	c, _ := fastClient(t, nil)
	if _, err := c.Interact(context.Background(), "", "x", nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestInteract_EmptyCode(t *testing.T) {
	c, _ := fastClient(t, nil)
	if _, err := c.Interact(context.Background(), "j", "", nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestInteract_Success(t *testing.T) {
	var gotBody string
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/scrape/j1/interact" {
			t.Errorf("path = %s", r.URL.Path)
		}
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "result": "done"})
	})
	resp, err := c.Interact(context.Background(), "j1", "await click()", &InteractParams{
		Language: "ts", Timeout: Int(30), Origin: "console",
	})
	if err != nil {
		t.Fatalf("Interact: %v", err)
	}
	if resp.Result != "done" {
		t.Fatalf("result = %q", resp.Result)
	}
	for _, want := range []string{`"language":"ts"`, `"timeout":30`, `"origin":"console"`} {
		if !strings.Contains(gotBody, want) {
			t.Errorf("body missing %s: %s", want, gotBody)
		}
	}
}

func TestStopInteractiveBrowser_EmptyJobID(t *testing.T) {
	c, _ := fastClient(t, nil)
	if _, err := c.StopInteractiveBrowser(context.Background(), ""); err == nil {
		t.Fatal("expected error")
	}
}

func TestStopInteractiveBrowser_Success(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v2/scrape/j1/interact" {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
	})
	if _, err := c.StopInteractiveBrowser(context.Background(), "j1"); err != nil {
		t.Fatalf("StopInteractiveBrowser: %v", err)
	}
}

// ============================================================
// USAGE & METRICS
// ============================================================

func TestGetConcurrency_Success(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/concurrency-check" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"concurrency": 3, "maxConcurrency": 10})
	})
	resp, err := c.GetConcurrency(context.Background())
	if err != nil {
		t.Fatalf("GetConcurrency: %v", err)
	}
	if resp.Concurrency != 3 || resp.MaxConcurrency != 10 {
		t.Fatalf("%+v", resp)
	}
}

func TestGetCreditUsage_Success(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/team/credit-usage" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"remainingCredits": 100, "planCredits": 1500,
		})
	})
	resp, err := c.GetCreditUsage(context.Background())
	if err != nil {
		t.Fatalf("GetCreditUsage: %v", err)
	}
	if resp.RemainingCredits != 100 {
		t.Fatalf("%+v", resp)
	}
}

// ============================================================
// PAGINATION (exercises paginateCrawl / paginateBatchScrape)
// ============================================================

func TestCrawlWithPolling_Pagination(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// POST /v2/crawl → return job ID
		// GET /v2/crawl/j1 → return done + next pointing to /page2
		// GET /page2 → return next page data, no next
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v2/crawl":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "j1"})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/crawl/j1":
			next := "http://" + r.Host + "/page2"
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "j1", "status": "completed",
				"data": []map[string]any{{"markdown": "# 1"}},
				"next": next,
			})
		case r.URL.Path == "/page2":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "completed",
				"data":   []map[string]any{{"markdown": "# 2"}, {"markdown": "# 3"}},
			})
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)

	c, err := NewClient(
		option.WithAPIKey("zf-x"),
		option.WithAPIURL(srv.URL),
		option.WithBackoffFactor(0.0),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	job, err := c.CrawlWithPolling(context.Background(), "https://ex.com", nil, 0, 10)
	if err != nil {
		t.Fatalf("CrawlWithPolling: %v", err)
	}
	if len(job.Data) != 3 {
		t.Fatalf("pagination merge failed, data len = %d", len(job.Data))
	}
}

func TestBatchScrapeWithPolling_Pagination(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost:
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "b1"})
		case r.URL.Path == "/v2/batch/scrape/b1":
			next := "http://" + r.Host + "/bpage2"
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "b1", "status": "completed",
				"data": []map[string]any{{"markdown": "a"}},
				"next": next,
			})
		case r.URL.Path == "/bpage2":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "completed",
				"data":   []map[string]any{{"markdown": "b"}},
			})
		}
	}))
	t.Cleanup(srv.Close)

	c, err := NewClient(
		option.WithAPIKey("zf-x"),
		option.WithAPIURL(srv.URL),
		option.WithBackoffFactor(0.0),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	job, err := c.BatchScrapeWithPolling(context.Background(), []string{"https://a"}, nil, 0, 10)
	if err != nil {
		t.Fatalf("BatchScrapeWithPolling: %v", err)
	}
	if len(job.Data) != 2 {
		t.Fatalf("pagination merge failed, len = %d", len(job.Data))
	}
}

// ============================================================
// ERROR ENVELOPE PATHS (extractError + extractDataAs edge)
// ============================================================

func TestExtractError_MalformedJSON(t *testing.T) {
	// Triggers the extractError "not valid JSON" branch at http_client.go:174.
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("<<not-json>>"))
	})
	_, err := c.Scrape(context.Background(), "https://x", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 400 {
		t.Fatalf("err = %v (%T), want *APIError 400", err, err)
	}
	if !strings.Contains(apiErr.Message, "HTTP 400") {
		t.Fatalf("msg = %q", apiErr.Message)
	}
}

func TestExtractError_MessageFieldFallback(t *testing.T) {
	// extractError should try "error" first, then fall back to "message".
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "forbidden",
			"code":    "FORBIDDEN",
		})
	})
	_, err := c.Scrape(context.Background(), "https://x", nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v (%T)", err, err)
	}
	if apiErr.Message != "forbidden" || apiErr.ErrorCode != "FORBIDDEN" {
		t.Fatalf("apiErr = %+v", apiErr)
	}
}

func TestScrape_DataAtTopLevelFallback(t *testing.T) {
	// extractDataAs has a branch: if response has no "data" key, unmarshal raw as T.
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		// Response without "data" envelope.
		_ = json.NewEncoder(w).Encode(map[string]any{"markdown": "# top-level"})
	})
	doc, err := c.Scrape(context.Background(), "https://x", nil)
	if err != nil {
		t.Fatalf("Scrape: %v", err)
	}
	if doc.Markdown != "# top-level" {
		t.Fatalf("Markdown = %q", doc.Markdown)
	}
}

// ============================================================
// JobTimeoutError message formatting
// ============================================================

func TestJobTimeoutError_ErrorString(t *testing.T) {
	e := &JobTimeoutError{
		APIError:       APIError{Message: "x"},
		JobID:          "abc",
		TimeoutSeconds: 42,
	}
	got := e.Error()
	if !strings.Contains(got, "abc") || !strings.Contains(got, "42") {
		t.Fatalf("Error() = %q, want to contain 'abc' and '42'", got)
	}
}

// ============================================================
// CONTEXT CANCELLATION (poll path)
// ============================================================

func TestCrawlWithPolling_ContextCancel(t *testing.T) {
	c, _ := fastClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "j"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "j", "status": "scraping"})
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := c.CrawlWithPolling(ctx, "https://ex.com", nil, 0, 30)
	if err == nil {
		t.Fatal("expected context cancellation")
	}
}
