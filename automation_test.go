package bisibility

import (
	"context"
	"net/http"
	"testing"
)

func TestAutomationEndpointMethods(t *testing.T) {
	t.Parallel()

	tests := []resourceMethodTestCase{
		{
			name: "list ranked keyword suggestions",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ListRankedKeywordSuggestions(ctx, "prj_a00000000000000000000000", &ListRankedKeywordSuggestionsOptions{
					ConnectionID: "conn_a00000000000000000000000",
					Fresh:        true,
					Limit:        100,
					Offset:       100,
				})
			},
			method: http.MethodGet,
			path:   "/api/v1/projects/prj_a00000000000000000000000/ranked-keyword-suggestions",
			query:  "connection_id=conn_a00000000000000000000000&fresh=true&limit=100&offset=100",
			response: map[string]any{
				"cached":      true,
				"connections": []any{map[string]any{"id": "conn_a00000000000000000000000", "label": "DataForSEO", "provider": "dataforseo"}},
				"cost_cents":  2.0,
				"fetched_at":  "2026-07-22T10:00:00Z",
				"offset":      100,
				"rows": []any{map[string]any{
					"already_tracked":   true,
					"estimated_traffic": 61.2,
					"keyword":           "rank tracker api",
					"position":          4,
					"search_volume":     720,
				}},
				"total_count": 184,
			},
			want: func(t *testing.T, got any) {
				response := got.(*RankedKeywordSuggestionsResponse)
				assertEqual(t, response.Cached, true)
				assertEqual(t, response.Rows[0].AlreadyTracked, true)
				assertEqual(t, *response.Rows[0].Position, 4)
				assertEqual(t, *response.Rows[0].SearchVolume, 720.0)
				assertEqual(t, response.Connections[0].Provider, ProviderIDDataForSEO)
			},
		},
		{
			name: "research keywords maps partial source diagnostics",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ResearchKeywords(ctx, "prj_a00000000000000000000000", ResearchKeywordsOptions{
					ConnectionID:       "conn_a00000000000000000000000",
					Fresh:              true,
					IncludeClickstream: true,
					MaxCostCents:       5,
					Mode:               KeywordResearchModeAuto,
					ResultLimit:        300,
					Seed:               "rank tracker",
				})
			},
			method: http.MethodGet,
			path:   "/api/v1/projects/prj_a00000000000000000000000/keyword-research",
			query:  "connection_id=conn_a00000000000000000000000&fresh=true&include_clickstream=true&max_cost_cents=5&mode=auto&result_limit=300&seed=rank+tracker",
			response: map[string]any{
				"cached":      false,
				"connections": []any{map[string]any{"id": "conn_a00000000000000000000000", "label": "DataForSEO", "provider": "dataforseo"}},
				"cost_cents":  2.0,
				"fetched_at":  "2026-07-22T10:00:00Z",
				"provider":    "DataForSEO",
				"rows": []any{map[string]any{
					"already_tracked": true,
					"competition":     0.42,
					"cpc_cents":       189,
					"difficulty":      36.0,
					"intent":          "commercial",
					"keyword":         "best rank tracker",
					"monthly_trend": []any{map[string]any{
						"month": 7, "search_volume": 720.0, "year": 2026,
					}},
					"search_volume": 720.0,
					"source":        "related",
				}},
				"sources": []any{
					map[string]any{
						"cached": false, "cost_cents": 2.0, "returned": 1, "source": "related", "status": "ok",
					},
					map[string]any{
						"cached": false, "cost_cents": 0, "reason": "budget_exhausted", "returned": 0, "source": "suggestion", "status": "failed",
					},
					map[string]any{
						"cached": false, "cost_cents": 0, "reason": "previous_source_failed", "returned": 0, "source": "idea", "status": "skipped",
					},
				},
				"total_count": 1,
			},
			want: func(t *testing.T, got any) {
				response := got.(*KeywordResearchResponse)
				assertEqual(t, response.Cached, false)
				assertEqual(t, response.Rows[0].AlreadyTracked, true)
				assertEqual(t, response.Rows[0].Source, KeywordResearchSourceRelated)
				assertEqual(t, *response.Rows[0].Intent, KeywordIntentCommercial)
				assertEqual(t, *response.Rows[0].CPCCents, 189)
				assertEqual(t, response.Sources[0].Cached, false)
				assertEqual(t, response.Sources[0].Status, KeywordResearchSourceStatusOK)
				assertEqual(t, response.Sources[1].Status, KeywordResearchSourceStatusFailed)
				assertEqual(t, *response.Sources[1].Reason, KeywordResearchSourceReasonBudgetExhausted)
				assertEqual(t, response.Sources[2].Status, KeywordResearchSourceStatusSkipped)
				assertEqual(t, *response.Sources[2].Reason, KeywordResearchSourceReasonPreviousSourceFailed)
				assertEqual(t, response.Connections[0].Provider, ProviderIDDataForSEO)
			},
		},
		{
			name: "research keywords maps estimate response",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ResearchKeywords(ctx, "prj_a00000000000000000000000", ResearchKeywordsOptions{
					EstimateOnly: true,
					Mode:         KeywordResearchModeAuto,
					ResultLimit:  100,
					Seed:         "rank tracker",
				})
			},
			method: http.MethodGet,
			path:   "/api/v1/projects/prj_a00000000000000000000000/keyword-research",
			query:  "estimate_only=true&mode=auto&result_limit=100&seed=rank+tracker",
			response: map[string]any{
				"cached":      false,
				"connections": []any{map[string]any{"id": "conn_a00000000000000000000000", "label": "DataForSEO", "provider": "dataforseo"}},
				"cost_cents":  3.0,
				"estimate":    true,
				"fetched_at":  "2026-07-22T10:00:00Z",
				"provider":    "DataForSEO",
				"rows":        []any{},
				"sources": []any{
					map[string]any{
						"cached": true, "cost_cents": 0, "returned": 0, "source": "related", "status": "ok",
					},
					map[string]any{
						"cached": false, "cost_cents": 1.0, "returned": 0, "source": "suggestion", "status": "ok",
					},
					map[string]any{
						"cached": false, "cost_cents": 2.0, "returned": 0, "source": "idea", "status": "ok",
					},
				},
				"total_count": 0,
			},
			want: func(t *testing.T, got any) {
				response := got.(*KeywordResearchResponse)
				assertEqual(t, *response.Estimate, true)
				assertEqual(t, response.CostCents, 3.0)
				assertEqual(t, len(response.Rows), 0)
				assertEqual(t, response.Sources[0].Cached, true)
				assertEqual(t, response.Sources[1].Status, KeywordResearchSourceStatusOK)
			},
		},
		{
			name: "get keyword metrics preserves nullable fields",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.GetKeywordMetrics(ctx, "prj_a00000000000000000000000", GetKeywordMetricsInput{
					ConnectionID:       "conn_a00000000000000000000000",
					Fresh:              true,
					IncludeClickstream: true,
					Keywords:           []string{"rank tracker", "seo api"},
					MaxCostCents:       5,
				})
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects/prj_a00000000000000000000000/keyword-metrics",
			body:            `{"connection_id":"conn_a00000000000000000000000","fresh":true,"include_clickstream":true,"keywords":["rank tracker","seo api"],"max_cost_cents":5}`,
			wantContentType: true,
			response: map[string]any{
				"cached_count":  1,
				"connections":   []any{map[string]any{"id": "conn_a00000000000000000000000", "label": "DataForSEO", "provider": "dataforseo"}},
				"cost_cents":    1.0,
				"fetched_at":    "2026-07-22T10:00:00Z",
				"fetched_count": 1,
				"provider":      "DataForSEO",
				"rows": []any{map[string]any{
					"competition": nil,
					"cpc_cents":   75,
					"difficulty":  nil,
					"intent":      nil,
					"keyword":     "seo api",
					"monthly_trend": []any{map[string]any{
						"month": 7, "search_volume": nil, "year": 2026,
					}},
					"search_volume": 110.0,
				}},
				"total_count": 2,
			},
			want: func(t *testing.T, got any) {
				response := got.(*KeywordMetricsResponse)
				assertEqual(t, response.CachedCount, 1)
				assertEqual(t, response.FetchedCount, 1)
				assertEqual(t, response.Rows[0].Keyword, "seo api")
				assertEqual(t, response.Rows[0].Difficulty == nil, true)
				assertEqual(t, response.Rows[0].Intent == nil, true)
				assertEqual(t, response.Rows[0].MonthlyTrend[0].SearchVolume == nil, true)
				assertEqual(t, *response.Rows[0].CPCCents, 75)
			},
		},
		{
			name: "get keyword metrics maps estimate response",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.GetKeywordMetrics(ctx, "prj_a00000000000000000000000", GetKeywordMetricsInput{
					EstimateOnly: true,
					Keywords:     []string{"rank tracker", "seo api"},
				})
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects/prj_a00000000000000000000000/keyword-metrics",
			body:            `{"estimate_only":true,"keywords":["rank tracker","seo api"]}`,
			wantContentType: true,
			response: map[string]any{
				"cached_count":           1,
				"connections":            []any{map[string]any{"id": "conn_a00000000000000000000000", "label": "DataForSEO", "provider": "dataforseo"}},
				"cost_cents":             0,
				"estimate":               true,
				"estimated_cost_cents":   1.01,
				"fetched_at":             "2026-07-22T10:00:00Z",
				"fetched_count":          0,
				"fetched_count_estimate": 1,
				"provider":               "DataForSEO",
				"rows":                   []any{},
				"total_count":            2,
			},
			want: func(t *testing.T, got any) {
				response := got.(*KeywordMetricsResponse)
				assertEqual(t, *response.Estimate, true)
				assertEqual(t, *response.EstimatedCostCents, 1.01)
				assertEqual(t, *response.FetchedCountEstimate, 1)
				assertEqual(t, response.CachedCount, 1)
				assertEqual(t, len(response.Rows), 0)
			},
		},
		{
			name: "search locations",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.SearchLocations(ctx, SearchLocationsOptions{Country: "US", Limit: 20, Query: "Austin"})
			},
			method: http.MethodGet,
			path:   "/api/v1/locations/search",
			query:  "country=US&limit=20&q=Austin",
			response: map[string]any{
				"data": []any{map[string]any{
					"city_name":      "Austin",
					"country_code":   "US",
					"display_name":   "Austin, Texas, United States",
					"hl":             "en",
					"kind":           "city",
					"language_code":  "en",
					"language_label": "English",
					"location_key":   "US/Texas/Austin",
					"region_code":    "TX",
					"region_name":    "Texas",
				}},
				"meta": map[string]any{"next_cursor": nil},
			},
			want: func(t *testing.T, got any) {
				response := got.(*LocationSuggestionsResponse)
				assertEqual(t, response.Data[0].Kind, LocationKindCity)
				assertEqual(t, response.Data[0].LocationKey, "US/Texas/Austin")
				assertEqual(t, response.Data[0].LanguageCode, "en")
			},
		},
		{
			name: "update team member role",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.UpdateTeamMemberRole(ctx, "prj_a00000000000000000000000", "mbr_a00000000000000000000000", UpdateTeamMemberRoleInput{Role: TeamRoleAdmin})
			},
			method:          http.MethodPatch,
			path:            "/api/v1/projects/prj_a00000000000000000000000/team/members/mbr_a00000000000000000000000",
			body:            `{"role":"admin"}`,
			response:        TeamMemberRoleResult{ID: "mbr_a00000000000000000000000", Role: TeamRoleAdmin},
			wantContentType: true,
			want: func(t *testing.T, got any) {
				assertEqual(t, got.(*TeamMemberRoleResult).Role, TeamRoleAdmin)
			},
		},
		{
			name: "remove team member",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.RemoveTeamMember(ctx, "prj_a00000000000000000000000", "mbr_a00000000000000000000000")
			},
			method:   http.MethodDelete,
			path:     "/api/v1/projects/prj_a00000000000000000000000/team/members/mbr_a00000000000000000000000",
			response: TeamMemberMutationResult{ID: "mbr_a00000000000000000000000"},
			want: func(t *testing.T, got any) {
				assertEqual(t, got.(*TeamMemberMutationResult).ID, "mbr_a00000000000000000000000")
			},
		},
		{
			name: "resend team invite",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ResendTeamInvite(ctx, "prj_a00000000000000000000000", "inv_a00000000000000000000000")
			},
			method: http.MethodPost,
			path:   "/api/v1/projects/prj_a00000000000000000000000/team/invites/inv_a00000000000000000000000/resend",
			response: map[string]any{
				"expires_at":  "2026-07-29T10:00:00Z",
				"id":          "inv_a00000000000000000000000",
				"invite_link": "https://app.test/invite/token",
			},
			want: func(t *testing.T, got any) {
				assertEqual(t, got.(*TeamInviteResendResult).InviteLink, "https://app.test/invite/token")
			},
		},
		{
			name: "list traffic snapshots",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ListTrafficSnapshots(ctx, "prj_a00000000000000000000000", ListTrafficSnapshotsOptions{
					EndDate:   "2026-06-30",
					Limit:     50,
					Offset:    20,
					Paths:     []string{"/blog", "/pricing"},
					StartDate: "2026-06-01",
				})
			},
			method: http.MethodGet,
			path:   "/api/v1/projects/prj_a00000000000000000000000/analytics/traffic-snapshots",
			query:  "end_date=2026-06-30&limit=50&offset=20&path=%2Fblog&path=%2Fpricing&start_date=2026-06-01",
			response: map[string]any{
				"offset": 20,
				"rows": []any{map[string]any{
					"bounce_rate":            0.42,
					"created_at":             "2026-07-01T00:00:00Z",
					"date":                   "2026-06-30",
					"engagement_rate":        0.58,
					"key_events":             3.0,
					"path":                   "/pricing",
					"provider":               "ga4",
					"scroll_depth":           0.71,
					"sessions":               42,
					"updated_at":             "2026-07-01T00:00:00Z",
					"visit_duration_seconds": 91.5,
					"visitors":               35,
					"window_days":            1,
				}},
				"total_count": 70,
			},
			want: func(t *testing.T, got any) {
				response := got.(*PageTrafficSnapshotsResponse)
				assertEqual(t, response.TotalCount, 70)
				assertEqual(t, response.Rows[0].Provider, "ga4")
				assertEqual(t, *response.Rows[0].Visitors, 35)
			},
		},
		{
			name: "list search performance query stats",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ListSearchPerformanceQueryStats(ctx, "prj_a00000000000000000000000", ListSearchPerformanceQueryStatsOptions{
					ConnectionID: "conn_b00000000000000000000000",
					EndDate:      "2026-06-30",
					Limit:        100,
					Query:        "rank tracker",
					StartDate:    "2026-06-01",
				})
			},
			method: http.MethodGet,
			path:   "/api/v1/projects/prj_a00000000000000000000000/analytics/query-stats",
			query:  "connection_id=conn_b00000000000000000000000&end_date=2026-06-30&limit=100&query=rank+tracker&start_date=2026-06-01",
			response: map[string]any{
				"connection": map[string]any{"id": "conn_b00000000000000000000000", "label": "Search Console", "provider": "gsc"},
				"rows": []any{map[string]any{
					"clicks":      12,
					"ctr":         0.15,
					"impressions": 80,
					"page":        "/pricing",
					"position":    4.2,
					"query":       "rank tracker",
				}},
			},
			want: func(t *testing.T, got any) {
				response := got.(*SearchPerformanceQueryStatsResponse)
				assertEqual(t, response.Connection.Provider, "gsc")
				assertEqual(t, response.Rows[0].Position, 4.2)
				assertEqual(t, *response.Rows[0].Page, "/pricing")
			},
		},
		{
			name: "sync project traffic",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.SyncProjectTraffic(ctx, "prj_a00000000000000000000000", WithIdempotencyKey("sync_1"))
			},
			method:         http.MethodPost,
			path:           "/api/v1/projects/prj_a00000000000000000000000/analytics/sync",
			idempotencyKey: "sync_1",
			response: map[string]any{
				"connections":       1,
				"keyword_snapshots": 3,
				"page_snapshots":    4,
				"project_id":        "prj_a00000000000000000000000",
				"runs": []any{map[string]any{
					"connection_id": "conn_c00000000000000000000000",
					"provider":      "ga4",
					"rows_fetched":  7,
					"rows_matched":  7,
					"rows_upserted": 7,
					"status":        "succeeded_with_data",
					"truncated":     false,
				}},
				"skipped": []any{map[string]any{"provider": "gsc", "reason": "no_capability"}},
			},
			want: func(t *testing.T, got any) {
				response := got.(*TrafficSyncSummary)
				assertEqual(t, response.Runs[0].Status, TrafficSyncRunSucceededWithData)
				assertEqual(t, response.Skipped[0].Reason, TrafficSyncSkipNoCapability)
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

func TestMirroredOpenAPIEnums(t *testing.T) {
	t.Parallel()

	assertStringSlicesEqual(t, []string{
		string(RankCheckFrequencyPaused),
		string(RankCheckFrequencyManual),
		string(RankCheckFrequencyDaily),
		string(RankCheckFrequencyWeekly),
		string(RankCheckFrequencyMonthly),
		string(RankCheckFrequencyCustomCron),
	}, []string{"paused", "manual", "daily", "weekly", "monthly", "custom_cron"})

	assertStringSlicesEqual(t, []string{
		string(ProjectWriteModeActive),
		string(ProjectWriteModeMigrationHold),
		string(ProjectWriteModeMigrated),
	}, []string{"active", "migration_hold", "migrated"})

	assertStringSlicesEqual(t, []string{
		string(ProviderIDDataForSEO),
		string(ProviderIDSerpAPI),
		string(ProviderIDGSC),
		string(ProviderIDGA4),
		string(ProviderIDPlausible),
		string(ProviderIDAhrefs),
		string(ProviderIDSemrush),
	}, []string{"dataforseo", "serpapi", "gsc", "ga4", "plausible", "ahrefs", "semrush"})

	assertStringSlicesEqual(t, []string{
		string(ProviderStatusConnected),
		string(ProviderStatusNeedsReauth),
		string(ProviderStatusReady),
		string(ProviderStatusPlanned),
		string(ProviderStatusOptional),
	}, []string{"connected", "needs_reauth", "ready", "planned", "optional"})

	assertStringSlicesEqual(t, []string{
		string(RankHistoryExportFormatCSV),
		string(RankHistoryExportFormatJSON),
	}, []string{"csv", "json"})
	assertStringSlicesEqual(t, []string{
		string(RankHistoryExportRange30Days),
		string(RankHistoryExportRange90Days),
		string(RankHistoryExportRangeAll),
	}, []string{"30", "90", "all"})
	assertStringSlicesEqual(t, []string{
		string(RankHistoryGranularityDaily),
		string(RankHistoryGranularityWeekly),
	}, []string{"daily", "weekly"})
	assertStringSlicesEqual(t, []string{
		string(SitemapMonitorStatusActive),
		string(SitemapMonitorStatusDisabled),
		string(SitemapMonitorStatusPending),
	}, []string{"active", "disabled", "pending"})
	assertStringSlicesEqual(t, []string{
		string(KeywordResearchModeAuto),
		string(KeywordResearchModeRelated),
		string(KeywordResearchModeSuggestions),
		string(KeywordResearchModeIdeas),
	}, []string{"auto", "related", "suggestions", "ideas"})
	assertStringSlicesEqual(t, []string{
		string(KeywordResearchSourceRelated),
		string(KeywordResearchSourceSuggestion),
		string(KeywordResearchSourceIdea),
	}, []string{"related", "suggestion", "idea"})
	assertStringSlicesEqual(t, []string{
		string(KeywordIntentInformational),
		string(KeywordIntentCommercial),
		string(KeywordIntentTransactional),
		string(KeywordIntentNavigational),
		string(KeywordIntentUnknown),
	}, []string{"informational", "commercial", "transactional", "navigational", "unknown"})
}
