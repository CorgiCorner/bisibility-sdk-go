package bisibility

import (
	"context"
	"net/http"
	"testing"
)

func TestBacklinksEndpointMethods(t *testing.T) {
	t.Parallel()

	tests := []resourceMethodTestCase{
		{
			name: "analyze backlinks sends every query parameter",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.AnalyzeBacklinks(ctx, "prj_a00000000000000000000000", AnalyzeBacklinksOptions{
					EstimateOnly:      true,
					Fresh:             true,
					IncludeSubdomains: true,
					MaxCostCents:      9,
					Mode:              BacklinkModeOnePerDomain,
					ResultLimit:       300,
					Target:            "https://example.com/pricing",
					TargetScope:       BacklinkTargetScopePage,
				})
			},
			method: http.MethodGet,
			path:   "/api/v1/projects/prj_a00000000000000000000000/backlinks",
			query:  "estimate_only=true&fresh=true&include_subdomains=true&max_cost_cents=9&mode=one_per_domain&result_limit=300&target=https%3A%2F%2Fexample.com%2Fpricing&target_scope=page",
			response: map[string]any{
				"data": backlinksEstimateJSON(),
			},
			want: func(t *testing.T, got any) {
				result := got.(*BacklinksResponse).Data
				if result.Snapshot != nil {
					t.Fatal("an estimate-only response decoded as a snapshot")
				}
				if result.Estimate == nil {
					t.Fatal("backlinks estimate = nil, want a decoded estimate")
				}
				estimate := result.Estimate
				assertEqual(t, estimate.Estimate, true)
				assertEqual(t, estimate.Target, "example.com")
				assertEqual(t, estimate.TargetScope, BacklinkTargetScopeSite)
				assertEqual(t, estimate.IncludeSubdomains, true)
				assertEqual(t, estimate.Cached, false)
				assertEqual(t, estimate.Provider, "dataforseo")
				assertEqual(t, estimate.CostCents, 7.0)
				assertEqual(t, estimate.EstimatedCostCents, 7.0)
				if estimate.CachedUntil != nil {
					t.Fatal("cached_until = non-nil, want nil for an uncached estimate")
				}
			},
		},
		{
			name: "analyze backlinks decodes a cached estimate",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.AnalyzeBacklinks(ctx, "prj_a00000000000000000000000", AnalyzeBacklinksOptions{
					EstimateOnly: true,
					Target:       "example.com",
				})
			},
			method: http.MethodGet,
			path:   "/api/v1/projects/prj_a00000000000000000000000/backlinks",
			query:  "estimate_only=true&target=example.com",
			response: map[string]any{
				"data": cachedBacklinksEstimateJSON(),
			},
			want: func(t *testing.T, got any) {
				estimate := got.(*BacklinksResponse).Data.Estimate
				if estimate == nil {
					t.Fatal("backlinks estimate = nil, want a decoded estimate")
				}
				assertEqual(t, estimate.Cached, true)
				assertEqual(t, estimate.CostCents, 0.0)
				assertEqual(t, estimate.EstimatedCostCents, 7.0)
				if estimate.CachedUntil == nil {
					t.Fatal("cached_until = nil, want the cached snapshot expiry")
				}
				assertEqual(t, estimate.CachedUntil.UTC().Format("2006-01-02T15:04:05Z"), "2026-07-25T15:00:00Z")
			},
		},
		{
			name: "analyze backlinks decodes a paid snapshot",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.AnalyzeBacklinks(ctx, "prj_a00000000000000000000000", AnalyzeBacklinksOptions{
					Target: "example.com",
				})
			},
			method:   http.MethodGet,
			path:     "/api/v1/projects/prj_a00000000000000000000000/backlinks",
			query:    "target=example.com",
			response: map[string]any{"data": backlinksSnapshotJSON()},
			want: func(t *testing.T, got any) {
				result := got.(*BacklinksResponse).Data
				if result.Estimate != nil {
					t.Fatal("a paid response decoded as an estimate")
				}
				if result.Snapshot == nil {
					t.Fatal("backlinks snapshot = nil, want a decoded snapshot")
				}
				snapshot := result.Snapshot
				assertEqual(t, snapshot.Target, "example.com")
				assertEqual(t, snapshot.TargetScope, BacklinkTargetScopeSite)
				assertEqual(t, snapshot.Cached, false)
				assertEqual(t, snapshot.CostCents, 7.0)
				assertEqual(t, snapshot.FetchedAt.UTC().Format("2006-01-02T15:04:05Z"), "2026-07-24T15:00:00Z")
				assertEqual(t, snapshot.CachedUntil.UTC().Format("2006-01-02T15:04:05Z"), "2026-07-25T15:00:00Z")
				assertEqual(t, snapshot.Summary.BacklinksTotal, 1685)
				assertEqual(t, snapshot.History[0].Month, "2025-08")
				assertEqual(t, snapshot.Rows[0].Status, BacklinkStatusActive)
				assertEqual(t, snapshot.Rows[0].DomainAuthority, 91)
				assertEqual(t, snapshot.Rows[0].LostAt == nil, true)
				assertEqual(t, snapshot.FetchedRowCount, 100)
				assertEqual(t, snapshot.TotalRowsAvailable, 1685)
			},
		},
		{
			name: "load more backlink rows sends required body",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.LoadMoreBacklinkRows(ctx, "prj_a00000000000000000000000", LoadMoreBacklinkRowsOptions{
					IncludeSubdomains: false,
					Limit:             100,
					Target:            "https://example.com/pricing",
					TargetScope:       BacklinkTargetScopePage,
				})
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects/prj_a00000000000000000000000/backlinks/rows",
			body:            `{"target":"https://example.com/pricing","target_scope":"page","include_subdomains":false,"limit":100}`,
			wantContentType: true,
			response: map[string]any{
				"data": backlinksSnapshotJSON(),
			},
			want: func(t *testing.T, got any) {
				snapshot := got.(*BacklinksSnapshotResponse).Data
				assertEqual(t, len(snapshot.Rows), 1)
				assertEqual(t, snapshot.Rows[0].SourceDomain, "reddit.com")
				assertEqual(t, snapshot.Rows[0].Flags[1], "ugc")
				assertEqual(t, snapshot.Rows[0].FirstSeen, "2026-01-21")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			runResourceMethodTestCase(t, tt)
		})
	}
}

func backlinksEstimateJSON() map[string]any {
	return map[string]any{
		"cached":               false,
		"cached_until":         nil,
		"cost_cents":           7.0,
		"estimate":             true,
		"estimated_cost_cents": 7.0,
		"include_subdomains":   true,
		"provider":             "dataforseo",
		"target":               "example.com",
		"target_scope":         "site",
	}
}

func cachedBacklinksEstimateJSON() map[string]any {
	estimate := backlinksEstimateJSON()
	estimate["cached"] = true
	estimate["cached_until"] = "2026-07-25T15:00:00Z"
	estimate["cost_cents"] = 0.0
	return estimate
}

func backlinksSnapshotJSON() map[string]any {
	return map[string]any{
		"cached":               false,
		"cached_until":         "2026-07-25T15:00:00Z",
		"cost_cents":           7.0,
		"fetched_at":           "2026-07-24T15:00:00Z",
		"fetched_row_count":    100,
		"history":              []any{map[string]any{"lost_links": 8, "month": "2025-08", "new_links": 22}},
		"include_subdomains":   true,
		"provider":             "dataforseo",
		"rows":                 []any{map[string]any{"anchor": "example.com", "domain_authority": 91, "first_seen": "2026-01-21", "flags": []string{"nofollow", "ugc"}, "links_count": 6, "lost_at": nil, "source_domain": "reddit.com", "source_url": "https://reddit.com/r/example", "spam_score": 2.0, "status": "active", "target_url": "https://example.com/"}},
		"summary":              map[string]any{"backlinks_total": 1685, "broken_backlinks": 0, "broken_pages": 0, "dofollow_pct": 61.0, "domain_rank": 37, "lost_backlinks": 12, "lost_referring_domains": 1, "new_backlinks": 34, "new_referring_domains": 3, "referring_domains_total": 48, "referring_pages": 1422, "spam_score": 3.0},
		"target":               "example.com",
		"target_scope":         "site",
		"total_rows_available": 1685,
	}
}
