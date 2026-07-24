package bisibility

import (
	"context"
	"net/http"
)

// ListRankedKeywordSuggestions lists one cached or paid page of ranked keyword suggestions.
func (c *Client) ListRankedKeywordSuggestions(ctx context.Context, projectID string, input *ListRankedKeywordSuggestionsOptions, options ...RequestOption) (*RankedKeywordSuggestionsResponse, error) {
	config := newRequestConfig(options...)
	if input != nil {
		addQuery(config.query, "connection_id", input.ConnectionID)
		if input.Fresh {
			config.query.Set("fresh", "true")
		}
		addIntQuery(config.query, "limit", input.Limit)
		addIntQuery(config.query, "offset", input.Offset)
	}
	return requestJSON[RankedKeywordSuggestionsResponse](c, ctx, http.MethodGet, projectResourcePath(projectID, "ranked-keyword-suggestions"), config)
}

// ResearchKeywords researches related keywords, suggestions, and ideas from one seed.
// The endpoint requires an API key with write scope because cache misses can spend provider budget.
func (c *Client) ResearchKeywords(ctx context.Context, projectID string, input ResearchKeywordsOptions, options ...RequestOption) (*KeywordResearchResponse, error) {
	config := newRequestConfig(options...)
	addQuery(config.query, "connection_id", input.ConnectionID)
	if input.EstimateOnly {
		config.query.Set("estimate_only", "true")
	}
	if input.Fresh {
		config.query.Set("fresh", "true")
	}
	if input.IncludeClickstream {
		config.query.Set("include_clickstream", "true")
	}
	addIntQuery(config.query, "max_cost_cents", input.MaxCostCents)
	addQuery(config.query, "mode", string(input.Mode))
	addIntQuery(config.query, "result_limit", input.ResultLimit)
	config.query.Set("seed", input.Seed)
	return requestJSON[KeywordResearchResponse](c, ctx, http.MethodGet, projectResourcePath(projectID, "keyword-research"), config)
}

// GetKeywordMetrics gets cached or paid provider metrics for up to 700 keywords.
// The endpoint requires an API key with write scope because cache misses can spend provider budget.
func (c *Client) GetKeywordMetrics(ctx context.Context, projectID string, input GetKeywordMetricsInput, options ...RequestOption) (*KeywordMetricsResponse, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[KeywordMetricsResponse](c, ctx, http.MethodPost, projectResourcePath(projectID, "keyword-metrics"), config)
}

// SearchLocations searches canonical location keys accepted by keyword methods.
func (c *Client) SearchLocations(ctx context.Context, input SearchLocationsOptions, options ...RequestOption) (*LocationSuggestionsResponse, error) {
	config := newRequestConfig(options...)
	addQuery(config.query, "country", input.Country)
	addIntQuery(config.query, "limit", input.Limit)
	config.query.Set("q", input.Query)
	return requestJSON[LocationSuggestionsResponse](c, ctx, http.MethodGet, "/locations/search", config)
}

// UpdateTeamMemberRole changes a non-owner project member's role.
func (c *Client) UpdateTeamMemberRole(ctx context.Context, projectID, memberID string, input UpdateTeamMemberRoleInput, options ...RequestOption) (*TeamMemberRoleResult, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[TeamMemberRoleResult](c, ctx, http.MethodPatch, projectMemberResourcePath(projectID, "team/members", memberID), config)
}

// RemoveTeamMember permanently removes a non-owner project member.
func (c *Client) RemoveTeamMember(ctx context.Context, projectID, memberID string, options ...RequestOption) (*TeamMemberMutationResult, error) {
	return requestJSON[TeamMemberMutationResult](c, ctx, http.MethodDelete, projectMemberResourcePath(projectID, "team/members", memberID), newRequestConfig(options...))
}

// ResendTeamInvite resends a pending project invite with a new token and expiration.
func (c *Client) ResendTeamInvite(ctx context.Context, projectID, inviteID string, options ...RequestOption) (*TeamInviteResendResult, error) {
	path := projectMemberResourcePath(projectID, teamInvitesResource, inviteID) + "/resend"
	return requestJSON[TeamInviteResendResult](c, ctx, http.MethodPost, path, newRequestConfig(options...))
}

// ListTrafficSnapshots lists stored page analytics snapshots for an inclusive date range.
func (c *Client) ListTrafficSnapshots(ctx context.Context, projectID string, input ListTrafficSnapshotsOptions, options ...RequestOption) (*PageTrafficSnapshotsResponse, error) {
	config := newRequestConfig(options...)
	config.query.Set("end_date", input.EndDate)
	addIntQuery(config.query, "limit", input.Limit)
	addIntQuery(config.query, "offset", input.Offset)
	for _, path := range input.Paths {
		config.query.Add("path", path)
	}
	config.query.Set("start_date", input.StartDate)
	return requestJSON[PageTrafficSnapshotsResponse](c, ctx, http.MethodGet, projectResourcePath(projectID, "analytics/traffic-snapshots"), config)
}

// ListSearchPerformanceQueryStats reads live query statistics from a project analytics connection.
func (c *Client) ListSearchPerformanceQueryStats(ctx context.Context, projectID string, input ListSearchPerformanceQueryStatsOptions, options ...RequestOption) (*SearchPerformanceQueryStatsResponse, error) {
	config := newRequestConfig(options...)
	addQuery(config.query, "connection_id", input.ConnectionID)
	config.query.Set("end_date", input.EndDate)
	addIntQuery(config.query, "limit", input.Limit)
	addQuery(config.query, "query", input.Query)
	config.query.Set("start_date", input.StartDate)
	return requestJSON[SearchPerformanceQueryStatsResponse](c, ctx, http.MethodGet, projectResourcePath(projectID, "analytics/query-stats"), config)
}

// SyncProjectTraffic runs the project's analytics traffic synchronization now.
// Pass WithIdempotencyKey to satisfy the endpoint's idempotency requirement.
func (c *Client) SyncProjectTraffic(ctx context.Context, projectID string, options ...RequestOption) (*TrafficSyncSummary, error) {
	return requestJSON[TrafficSyncSummary](c, ctx, http.MethodPost, projectResourcePath(projectID, "analytics/sync"), newRequestConfig(options...))
}
