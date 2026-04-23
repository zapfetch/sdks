package zapfetch

import (
	"encoding/json"
	"testing"
)

// The map endpoint returns each link either as a bare string or as an object;
// LinkResult.UnmarshalJSON bridges both shapes.
func TestLinkResult_UnmarshalBareString(t *testing.T) {
	var link LinkResult
	if err := json.Unmarshal([]byte(`"https://example.com/a"`), &link); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if link.URL != "https://example.com/a" {
		t.Fatalf("URL = %q, want %q", link.URL, "https://example.com/a")
	}
	if link.Title != "" || link.Description != "" {
		t.Errorf("Title/Description should be empty for bare string form")
	}
}

func TestLinkResult_UnmarshalObject(t *testing.T) {
	raw := `{"url":"https://example.com/b","title":"B","description":"desc"}`
	var link LinkResult
	if err := json.Unmarshal([]byte(raw), &link); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if link.URL != "https://example.com/b" || link.Title != "B" || link.Description != "desc" {
		t.Fatalf("got %+v", link)
	}
}

func TestLinkResult_UnmarshalArray(t *testing.T) {
	raw := `[
		"https://a.example.com",
		{"url":"https://b.example.com","title":"B"}
	]`
	var links []LinkResult
	if err := json.Unmarshal([]byte(raw), &links); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("len = %d, want 2", len(links))
	}
	if links[0].URL != "https://a.example.com" || links[0].Title != "" {
		t.Errorf("links[0] = %+v", links[0])
	}
	if links[1].URL != "https://b.example.com" || links[1].Title != "B" {
		t.Errorf("links[1] = %+v", links[1])
	}
}

func TestCrawlJob_IsDone(t *testing.T) {
	cases := []struct {
		status string
		want   bool
	}{
		{"completed", true},
		{"failed", true},
		{"cancelled", true},
		{"scraping", false},
		{"pending", false},
		{"", false},
	}
	for _, tc := range cases {
		t.Run(tc.status, func(t *testing.T) {
			c := &CrawlJob{Status: tc.status}
			if got := c.IsDone(); got != tc.want {
				t.Fatalf("IsDone(%q) = %v, want %v", tc.status, got, tc.want)
			}
		})
	}
}

func TestBatchScrapeJob_IsDone(t *testing.T) {
	b := &BatchScrapeJob{Status: "completed"}
	if !b.IsDone() {
		t.Error("expected IsDone() to be true for completed batch")
	}
}

func TestAgentStatusResponse_IsDone(t *testing.T) {
	a := &AgentStatusResponse{Status: "failed"}
	if !a.IsDone() {
		t.Error("expected IsDone() to be true for failed agent task")
	}
	a.Status = "running"
	if a.IsDone() {
		t.Error("expected IsDone() to be false for running agent task")
	}
}
