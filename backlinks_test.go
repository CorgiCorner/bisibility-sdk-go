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
				return c.AnalyzeBacklinks(ctx, "proj_1", AnalyzeBacklinksOptions{
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
			path:   "/api/v1/projects/proj_1/backlinks",
			query:  "estimate_only=true&fresh=true&include_subdomains=true&max_cost_cents=9&mode=one_per_domain&result_limit=300&target=https%3A%2F%2Fexample.com%2Fpricing&target_scope=page",
			response: map[string]any{
				"data": backlinksSnapshotJSON(),
			},
			want: func(t *testing.T, got any) {
				response := got.(*BacklinksSnapshotResponse)
				snapshot := response.Data
				assertEqual(t, snapshot.Target, "example.com")
				assertEqual(t, snapshot.TargetScope, BacklinkTargetScopeSite)
				assertEqual(t, snapshot.Cached, false)
				assertEqual(t, snapshot.FetchedAt.UTC().Format("2006-01-02T15:04:05Z"), "2026-07-24T15:00:00Z")
				assertEqual(t, snapshot.CachedUntil.UTC().Format("2006-01-02T15:04:05Z"), "2026-07-25T15:00:00Z")
				assertEqual(t, *snapshot.Estimate, true)
				assertEqual(t, *snapshot.EstimatedCostCents, 5.0)
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
			name: "analyze backlinks omits zero-value optional parameters",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.AnalyzeBacklinks(ctx, "proj_1", AnalyzeBacklinksOptions{
					Target: "example.com",
				})
			},
			method:   http.MethodGet,
			path:     "/api/v1/projects/proj_1/backlinks",
			query:    "target=example.com",
			response: map[string]any{"data": backlinksSnapshotJSON()},
			want: func(t *testing.T, got any) {
				assertEqual(t, got.(*BacklinksSnapshotResponse).Data.Target, "example.com")
			},
		},
		{
			name: "load more backlink rows sends required body",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.LoadMoreBacklinkRows(ctx, "proj_1", LoadMoreBacklinkRowsOptions{
					IncludeSubdomains: false,
					Limit:             100,
					Target:            "https://example.com/pricing",
					TargetScope:       BacklinkTargetScopePage,
				})
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects/proj_1/backlinks/rows",
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

func backlinksSnapshotJSON() map[string]any {
	return map[string]any{
		"cached":               false,
		"cached_until":         "2026-07-25T15:00:00Z",
		"cost_cents":           5.0,
		"estimate":             true,
		"estimated_cost_cents": 5.0,
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
