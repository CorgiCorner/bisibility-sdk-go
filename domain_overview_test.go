package bisibility

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDomainOverviewEndpointMethods(t *testing.T) {
	t.Parallel()

	fresh := true
	estimateOnly := true
	scope := DomainOverviewScopeSubdomain
	maxCost := 7
	keywordLimit := 100
	pageLimit := 50

	tests := []resourceMethodTestCase{
		{
			name: "analyze domain overview sends complete body and decodes estimate",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.AnalyzeDomainOverview(ctx, "prj_a00000000000000000000000", AnalyzeDomainOverviewOptions{
					Target:        "blog.example.com",
					LocationCode:  2840,
					LanguageCode:  "en",
					ScopeOverride: &scope,
					Fresh:         &fresh,
					MaxCostCents:  &maxCost,
					EstimateOnly:  &estimateOnly,
					KeywordLimit:  &keywordLimit,
					PageLimit:     &pageLimit,
				})
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects/prj_a00000000000000000000000/domain-overview/analyze",
			body:            `{"target":"blog.example.com","location_code":2840,"language_code":"en","scope_override":"subdomain","fresh":true,"max_cost_cents":7,"estimate_only":true,"keyword_limit":100,"page_limit":50}`,
			wantContentType: true,
			response:        map[string]any{"data": domainOverviewEstimateJSON()},
			want: func(t *testing.T, got any) {
				result := got.(*DomainOverviewAnalyzeResponse).Data
				assertEqual(t, result.Report == nil, true)
				assertEqual(t, result.Estimate != nil, true)
				assertEqual(t, result.Estimate.EstimatedCostCents, 4.25)
				assertEqual(t, result.Estimate.HistoryMode, DomainOverviewHistoryModeLazy)
				assertEqual(t, result.Estimate.Scope, DomainOverviewScopeSubdomain)
			},
		},
		{
			name: "analyze domain overview omits optional input and decodes nested report",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.AnalyzeDomainOverview(ctx, "prj_a00000000000000000000000", AnalyzeDomainOverviewOptions{
					Target:       "example.com",
					LocationCode: 2840,
					LanguageCode: "en",
				})
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects/prj_a00000000000000000000000/domain-overview/analyze",
			body:            `{"target":"example.com","location_code":2840,"language_code":"en"}`,
			wantContentType: true,
			response:        map[string]any{"data": domainOverviewReportJSON()},
			want: func(t *testing.T, got any) {
				result := got.(*DomainOverviewAnalyzeResponse).Data
				assertEqual(t, result.Estimate == nil, true)
				assertEqual(t, result.Report != nil, true)
				report := result.Report
				assertEqual(t, report.State, DomainOverviewStatePartial)
				assertEqual(t, report.SourceSnapshotAt.UTC().Format("2006-01-02"), "2026-08-12")
				assertEqual(t, report.PreviousSourceSnapshotAt == nil, true)
				assertEqual(t, report.Overview.Count == nil, true)
				assertEqual(t, *report.Overview.ETV, 1413.3)
				assertEqual(t, report.Overview.Pos1, 4)
				assertEqual(t, report.Overview.Pos2To3, 8)
				assertEqual(t, report.Overview.Pos4To10, 21)
				assertEqual(t, report.Overview.Pos11To20, 34)
				assertEqual(t, report.Keywords.OK, true)
				assertEqual(t, report.Keywords.Data.Rows[0].Keyword, "rank tracker")
				assertEqual(t, *report.Keywords.Data.Rows[0].Intent, DomainOverviewIntentCommercial)
				assertEqual(t, report.Pages.OK, false)
				assertEqual(t, *report.Pages.Reason, DomainOverviewFailureRateLimited)
				assertEqual(t, *report.Pages.ResetAt, float64(1786579200000))
			},
		},
		{
			name: "load domain overview history sends explicit cache-only cap",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.LoadDomainOverviewHistory(ctx, "prj_a00000000000000000000000", LoadDomainOverviewHistoryOptions{
					Target:       "example.com",
					LocationCode: 2840,
					LanguageCode: "en",
					MaxCostCents: 0,
				})
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects/prj_a00000000000000000000000/domain-overview/history",
			body:            `{"target":"example.com","location_code":2840,"language_code":"en","max_cost_cents":0}`,
			wantContentType: true,
			response: map[string]any{"data": map[string]any{
				"cached": true, "cost_cents": 0.0, "fetched_at": "2026-08-12T09:30:00Z",
				"data": []any{map[string]any{"year": 2026, "month": 7, "metrics": domainRankMetricsJSON()}},
			}},
			want: func(t *testing.T, got any) {
				module := got.(*DomainOverviewHistoryResponse).Data
				assertEqual(t, module.Cached, true)
				assertEqual(t, module.Data[0].Month, 7)
				assertEqual(t, *module.Data[0].Metrics.EstimatedTrafficCostCents, 142.065)
			},
		},
		{
			name: "load domain overview keywords sends pagination body",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.LoadDomainOverviewKeywords(ctx, "prj_a00000000000000000000000", LoadDomainOverviewKeywordsOptions{
					Target:       "example.com",
					LocationCode: 2840,
					LanguageCode: "en",
					MaxCostCents: maxCost,
					Limit:        100,
					Offset:       200,
				})
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects/prj_a00000000000000000000000/domain-overview/keywords",
			body:            `{"target":"example.com","location_code":2840,"language_code":"en","max_cost_cents":7,"limit":100,"offset":200}`,
			wantContentType: true,
			response: map[string]any{"data": map[string]any{
				"cached": false, "cost_cents": 2.0, "fetched_at": "2026-08-12T09:31:00Z",
				"data": domainOverviewKeywordsPageJSON(),
			}},
			want: func(t *testing.T, got any) {
				module := got.(*DomainOverviewKeywordsResponse).Data
				assertEqual(t, module.Data.TotalCount != nil, true)
				assertEqual(t, *module.Data.TotalCount, 938)
				assertEqual(t, *module.Data.Rows[0].RankAbsoluteDelta, -2)
				assertEqual(t, module.Data.Rows[0].SERPFeatures[0], "featured_snippet")
			},
		},
		{
			name: "load domain overview pages sends complete body and decodes nullable row",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.LoadDomainOverviewPages(ctx, "prj_a00000000000000000000000", LoadDomainOverviewPagesOptions{
					Target:        "blog.example.com",
					LocationCode:  2826,
					LanguageCode:  "en",
					ScopeOverride: &scope,
					Fresh:         &fresh,
					MaxCostCents:  maxCost,
					Limit:         10,
					Offset:        0,
				})
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects/prj_a00000000000000000000000/domain-overview/pages",
			body:            `{"target":"blog.example.com","location_code":2826,"language_code":"en","scope_override":"subdomain","fresh":true,"max_cost_cents":7,"limit":10,"offset":0}`,
			wantContentType: true,
			response: map[string]any{"data": map[string]any{
				"cached": false, "cost_cents": 1.25, "fetched_at": "2026-08-12T09:32:00Z",
				"data": map[string]any{"cost_cents": 1.25, "total_count": 225, "rows": []any{map[string]any{
					"etv": nil, "etv_delta_pct": -12.5, "keyword_count": 42, "path": "/pricing",
					"top_keyword": "rank tracker", "top_keyword_position": nil,
				}}},
			}},
			want: func(t *testing.T, got any) {
				module := got.(*DomainOverviewPagesResponse).Data
				assertEqual(t, module.Data.TotalCount, 225)
				assertEqual(t, module.Data.Rows[0].ETV == nil, true)
				assertEqual(t, *module.Data.Rows[0].ETVDeltaPct, -12.5)
				assertEqual(t, module.Data.Rows[0].TopKeywordPosition == nil, true)
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

func TestDomainOverviewProblemErrorsDecodeChargedFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusTooManyRequests, map[string]any{
			"type":   "https://bisibility.com/problems/rate_limited",
			"title":  "Rate limited",
			"status": http.StatusTooManyRequests,
			"detail": "Provider rate limit reached.",
			"errors": map[string]any{
				"reason": "rate_limited", "cost_cents": 1.25, "reset_at": 1786579200000,
			},
		})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	maxCost := 2
	_, err := client.LoadDomainOverviewHistory(context.Background(), "prj_a00000000000000000000000", LoadDomainOverviewHistoryOptions{
		Target: "example.com", LocationCode: 2840, LanguageCode: "en", MaxCostCents: maxCost,
	})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T, want APIError", err)
	}
	var details DomainOverviewProblemErrors
	if err := json.Unmarshal(apiErr.Problem.Errors, &details); err != nil {
		t.Fatalf("decode Domain Overview problem errors: %v", err)
	}
	assertEqual(t, details.Reason, DomainOverviewFailureRateLimited)
	assertEqual(t, details.CostCents, 1.25)
	assertEqual(t, *details.ResetAt, float64(1786579200000))
}

func domainOverviewEstimateJSON() map[string]any {
	return map[string]any{
		"cached": false, "estimate": true, "estimated_cost_cents": 4.25,
		"fresh_estimated_cost_cents": 6.5, "history_estimated_cost_cents": 12.0,
		"history_mode": "lazy", "keyword_page_estimated_cost_cents": 2.0,
		"language_code": "en", "location_code": 2840,
		"page_page_estimated_cost_cents": 1.25, "provider": "dataforseo",
		"scope": "subdomain", "target": "blog.example.com",
	}
}

func domainOverviewReportJSON() map[string]any {
	return map[string]any{
		"cached": false, "cached_until": "2026-08-12T21:30:00Z", "cost_cents": 4.25,
		"fetched_at": "2026-08-12T09:30:00Z", "history_mode": "lazy",
		"keywords": map[string]any{
			"ok": true, "cached": false, "cost_cents": 2.0, "fetched_at": "2026-08-12T09:30:01Z",
			"data": domainOverviewKeywordsPageJSON(),
		},
		"language_code": "en", "location_code": 2840,
		"overview": domainRankMetricsJSON(),
		"pages": map[string]any{
			"ok": false, "cost_cents": 1.25, "reason": "rate_limited", "reset_at": int64(1786579200000),
		},
		"previous_fetched_at": "2026-08-05T09:30:00Z", "previous_overview": domainRankMetricsJSON(),
		"previous_source_snapshot_at": nil, "provider": "dataforseo", "scope": "root",
		"source_snapshot_at": "2026-08-12T00:00:00Z", "state": "partial", "target": "example.com",
	}
}

func domainRankMetricsJSON() map[string]any {
	return map[string]any{
		"count": nil, "estimated_traffic_cost_cents": 142.065, "etv": 1413.3,
		"is_down": 131, "is_lost": 360, "is_new": 640, "is_up": 153,
		"pos1": 4, "pos2_3": 8, "pos4_10": 21, "pos11_20": 34,
		"pos21_30": 45, "pos31_40": 56, "pos41_50": 67, "pos51_60": 78,
		"pos61_70": 89, "pos71_80": 90, "pos81_90": 91, "pos91_100": 92,
	}
}

func domainOverviewKeywordsPageJSON() map[string]any {
	return map[string]any{
		"cost_cents":  2.0,
		"total_count": 938,
		"rows": []any{map[string]any{
			"keyword": "rank tracker", "position": 3, "search_volume": 880,
			"estimated_traffic": 142.065, "cpc_cents": 75.5, "difficulty": 42.5,
			"intent": "commercial", "ranking_url": "https://example.com/rank-tracker",
			"serp_features": []string{"featured_snippet"}, "rank_absolute_delta": -2,
			"rank_absolute": 5,
		}},
	}
}
