package bisibility

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// TestPaidGETDoesNotRetryWithoutIdempotencyKey covers contract-audit finding #1:
// the SDK must not auto-retry a potentially paid GET after a lost response,
// because the backend route is not idempotent and a second attempt would spend
// provider budget twice. A caller who accepts repetition opts in by supplying
// an Idempotency-Key, exactly as POST writes do.
func TestPaidGETDoesNotRetryWithoutIdempotencyKey(t *testing.T) {
	t.Run("non-estimate backlinks execute at most once", func(t *testing.T) {
		var calls atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			w.Header().Set("Retry-After", "0")
			writeJSON(t, w, http.StatusServiceUnavailable, map[string]any{"type": "busy", "status": 503})
		}))
		defer server.Close()
		c := newTestClient(t, server.URL+"/api/v1")
		_, err := c.AnalyzeBacklinks(context.Background(), "prj_a00000000000000000000000", AnalyzeBacklinksOptions{
			Target:       "https://example.com/",
			TargetScope:  BacklinkTargetScopeSite,
			Mode:         BacklinkModeOnePerDomain,
			MaxCostCents: 10,
			Fresh:        true,
		})
		if err == nil {
			t.Fatal("expected 503 error, got nil")
		}
		if got := calls.Load(); got != 1 {
			t.Fatalf("paid backlinks executed %d times; want 1", got)
		}
	})

	t.Run("estimate-only backlinks still retry 503 like other idempotent GETs", func(t *testing.T) {
		var calls atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if calls.Add(1) == 1 {
				w.Header().Set("Retry-After", "0")
				writeJSON(t, w, http.StatusServiceUnavailable, map[string]any{"type": "busy", "status": 503})
				return
			}
			writeJSON(t, w, http.StatusOK, map[string]any{"data": backlinksEstimateJSON()})
		}))
		defer server.Close()
		c := newTestClient(t, server.URL+"/api/v1")
		if _, err := c.AnalyzeBacklinks(context.Background(), "prj_a00000000000000000000000", AnalyzeBacklinksOptions{
			Target:       "https://example.com/",
			TargetScope:  BacklinkTargetScopeSite,
			Mode:         BacklinkModeOnePerDomain,
			EstimateOnly: true,
		}); err != nil {
			t.Fatal(err)
		}
		if got := calls.Load(); got != 2 {
			t.Fatalf("estimate backlinks executed %d times; want 2 (retry allowed)", got)
		}
	})

	t.Run("paid keyword research executes at most once", func(t *testing.T) {
		var calls atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			w.Header().Set("Retry-After", "0")
			writeJSON(t, w, http.StatusServiceUnavailable, map[string]any{"type": "busy", "status": 503})
		}))
		defer server.Close()
		c := newTestClient(t, server.URL+"/api/v1")
		_, _ = c.ResearchKeywords(context.Background(), "prj_a00000000000000000000000", ResearchKeywordsOptions{
			Seed:         "corgis",
			MaxCostCents: 10,
			Fresh:        true,
		})
		if got := calls.Load(); got != 1 {
			t.Fatalf("paid keyword research executed %d times; want 1", got)
		}
	})

	t.Run("fresh ranked suggestions execute at most once", func(t *testing.T) {
		var calls atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			w.Header().Set("Retry-After", "0")
			writeJSON(t, w, http.StatusServiceUnavailable, map[string]any{"type": "busy", "status": 503})
		}))
		defer server.Close()
		c := newTestClient(t, server.URL+"/api/v1")
		_, _ = c.ListRankedKeywordSuggestions(context.Background(), "prj_a00000000000000000000000", &ListRankedKeywordSuggestionsOptions{Fresh: true})
		if got := calls.Load(); got != 1 {
			t.Fatalf("fresh suggestions executed %d times; want 1", got)
		}
	})

	t.Run("idempotency key opts a paid GET back into retries", func(t *testing.T) {
		var calls atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Idempotency-Key") == "" {
				t.Error("expected the caller's idempotency key to reach the backend")
			}
			if calls.Add(1) == 1 {
				w.Header().Set("Retry-After", "0")
				writeJSON(t, w, http.StatusServiceUnavailable, map[string]any{"type": "busy", "status": 503})
				return
			}
			writeJSON(t, w, http.StatusOK, map[string]any{"data": backlinksSnapshotJSON()})
		}))
		defer server.Close()
		c := newTestClient(t, server.URL+"/api/v1")
		if _, err := c.AnalyzeBacklinks(context.Background(), "prj_a00000000000000000000000", AnalyzeBacklinksOptions{
			Target:       "https://example.com/",
			TargetScope:  BacklinkTargetScopeSite,
			Mode:         BacklinkModeOnePerDomain,
			MaxCostCents: 10,
			Fresh:        true,
		}, WithIdempotencyKey("stable")); err != nil {
			t.Fatal(err)
		}
		if got := calls.Load(); got != 2 {
			t.Fatalf("paid backlinks with idempotency key executed %d times; want 2", got)
		}
	})
}
