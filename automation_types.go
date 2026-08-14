package bisibility

import "time"

// ListRankedKeywordSuggestionsOptions controls one offset-paginated ranked keyword lookup.
type ListRankedKeywordSuggestionsOptions struct {
	ConnectionID string
	Fresh        bool
	Limit        int
	Offset       int
}

// RankedKeywordSuggestion is one query for which the project domain already ranks.
type RankedKeywordSuggestion struct {
	AlreadyTracked   bool     `json:"already_tracked"`
	EstimatedTraffic *float64 `json:"estimated_traffic"`
	Keyword          string   `json:"keyword"`
	Position         *int     `json:"position"`
	SearchVolume     *float64 `json:"search_volume"`
}

// RankedKeywordConnection is an eligible project-owned DataForSEO connection.
type RankedKeywordConnection struct {
	ID       string     `json:"id"`
	Label    string     `json:"label"`
	Provider ProviderID `json:"provider"`
}

// RankedKeywordSuggestionsResponse is one cached or paid ranked keyword page.
type RankedKeywordSuggestionsResponse struct {
	Cached      bool                      `json:"cached"`
	Connections []RankedKeywordConnection `json:"connections"`
	CostCents   float64                   `json:"cost_cents"`
	FetchedAt   time.Time                 `json:"fetched_at"`
	Offset      int                       `json:"offset"`
	Rows        []RankedKeywordSuggestion `json:"rows"`
	TotalCount  *int                      `json:"total_count"`
}

// KeywordResearchMode selects the DataForSEO Labs research source or automatic cascade.
type KeywordResearchMode string

const (
	KeywordResearchModeAuto        KeywordResearchMode = "auto"
	KeywordResearchModeRelated     KeywordResearchMode = "related"
	KeywordResearchModeSuggestions KeywordResearchMode = "suggestions"
	KeywordResearchModeIdeas       KeywordResearchMode = "ideas"
)

// ResearchKeywordsOptions controls a paid or estimated single-seed keyword research lookup.
type ResearchKeywordsOptions struct {
	ConnectionID       string
	EstimateOnly       bool
	Fresh              bool
	IncludeClickstream bool
	MaxCostCents       int
	Mode               KeywordResearchMode
	ResultLimit        int
	Seed               string
}

// KeywordResearchSource identifies which DataForSEO Labs source returned a keyword.
type KeywordResearchSource string

const (
	KeywordResearchSourceRelated    KeywordResearchSource = "related"
	KeywordResearchSourceSuggestion KeywordResearchSource = "suggestion"
	KeywordResearchSourceIdea       KeywordResearchSource = "idea"
)

// KeywordResearchSourceStatus describes the outcome of one source in a research cascade.
type KeywordResearchSourceStatus string

const (
	KeywordResearchSourceStatusOK      KeywordResearchSourceStatus = "ok"
	KeywordResearchSourceStatusFailed  KeywordResearchSourceStatus = "failed"
	KeywordResearchSourceStatusSkipped KeywordResearchSourceStatus = "skipped"
)

// KeywordResearchSourceReason explains why a research source failed or was skipped.
type KeywordResearchSourceReason string

const (
	KeywordResearchSourceReasonBudgetExhausted      KeywordResearchSourceReason = "budget_exhausted"
	KeywordResearchSourceReasonCostLimit            KeywordResearchSourceReason = "cost_limit"
	KeywordResearchSourceReasonInProgress           KeywordResearchSourceReason = "in_progress"
	KeywordResearchSourceReasonNeedsReauth          KeywordResearchSourceReason = "needs_reauth"
	KeywordResearchSourceReasonNoSource             KeywordResearchSourceReason = "no_source"
	KeywordResearchSourceReasonPreviousSourceFailed KeywordResearchSourceReason = "previous_source_failed"
	KeywordResearchSourceReasonProviderError        KeywordResearchSourceReason = "provider_error"
	KeywordResearchSourceReasonRateLimited          KeywordResearchSourceReason = "rate_limited"
	KeywordResearchSourceReasonResultLimit          KeywordResearchSourceReason = "result_limit"
	KeywordResearchSourceReasonUnsupportedLocation  KeywordResearchSourceReason = "unsupported_location"
)

// KeywordIntent is the provider-classified search intent.
type KeywordIntent string

const (
	KeywordIntentInformational KeywordIntent = "informational"
	KeywordIntentCommercial    KeywordIntent = "commercial"
	KeywordIntentTransactional KeywordIntent = "transactional"
	KeywordIntentNavigational  KeywordIntent = "navigational"
	KeywordIntentUnknown       KeywordIntent = "unknown"
)

// KeywordMonthlyTrend is one month of provider search-volume history.
type KeywordMonthlyTrend struct {
	Month        int      `json:"month"`
	SearchVolume *float64 `json:"search_volume"`
	Year         int      `json:"year"`
}

// KeywordMetrics contains nullable provider metrics shared by research and hydration rows.
type KeywordMetrics struct {
	Competition  *float64              `json:"competition"`
	CPCCents     *int                  `json:"cpc_cents"`
	Difficulty   *float64              `json:"difficulty"`
	Intent       *KeywordIntent        `json:"intent"`
	MonthlyTrend []KeywordMonthlyTrend `json:"monthly_trend"`
	SearchVolume *float64              `json:"search_volume"`
}

// KeywordResearchRow is one researched keyword and its nullable provider metrics.
type KeywordResearchRow struct {
	KeywordMetrics
	AlreadyTracked bool                  `json:"already_tracked"`
	Keyword        string                `json:"keyword"`
	Source         KeywordResearchSource `json:"source"`
}

// KeywordResearchSourceSummary describes one source used by a research lookup.
type KeywordResearchSourceSummary struct {
	Cached    bool                         `json:"cached"`
	CostCents float64                      `json:"cost_cents"`
	Reason    *KeywordResearchSourceReason `json:"reason,omitempty"`
	Returned  int                          `json:"returned"`
	Source    KeywordResearchSource        `json:"source"`
	Status    KeywordResearchSourceStatus  `json:"status"`
}

// KeywordResearchConnection is an eligible project-owned DataForSEO connection.
type KeywordResearchConnection = RankedKeywordConnection

// KeywordResearchResponse is one cached, paid, or estimated single-seed research result.
type KeywordResearchResponse struct {
	Cached      bool                           `json:"cached"`
	Connections []KeywordResearchConnection    `json:"connections"`
	CostCents   float64                        `json:"cost_cents"`
	Estimate    *bool                          `json:"estimate,omitempty"`
	FetchedAt   time.Time                      `json:"fetched_at"`
	Provider    string                         `json:"provider"`
	Rows        []KeywordResearchRow           `json:"rows"`
	Sources     []KeywordResearchSourceSummary `json:"sources"`
	TotalCount  int                            `json:"total_count"`
}

// GetKeywordMetricsInput requests or estimates provider metrics for up to 700 keywords.
type GetKeywordMetricsInput struct {
	ConnectionID       string   `json:"connection_id,omitempty"`
	EstimateOnly       bool     `json:"estimate_only,omitempty"`
	Fresh              bool     `json:"fresh,omitempty"`
	IncludeClickstream bool     `json:"include_clickstream,omitempty"`
	Keywords           []string `json:"keywords"`
	MaxCostCents       int      `json:"max_cost_cents,omitempty"`
}

// KeywordMetricsRow is one keyword with nullable provider metrics.
type KeywordMetricsRow struct {
	KeywordMetrics
	Keyword string `json:"keyword"`
}

// KeywordMetricsResponse reports cached and fetched metrics for the request.
type KeywordMetricsResponse struct {
	CachedCount          int                         `json:"cached_count"`
	Connections          []KeywordResearchConnection `json:"connections"`
	CostCents            float64                     `json:"cost_cents"`
	Estimate             *bool                       `json:"estimate,omitempty"`
	EstimatedCostCents   *float64                    `json:"estimated_cost_cents,omitempty"`
	FetchedAt            time.Time                   `json:"fetched_at"`
	FetchedCount         int                         `json:"fetched_count"`
	FetchedCountEstimate *int                        `json:"fetched_count_estimate,omitempty"`
	Provider             string                      `json:"provider"`
	Rows                 []KeywordMetricsRow         `json:"rows"`
	TotalCount           int                         `json:"total_count"`
}

// LocationKind identifies the level represented by a location suggestion.
type LocationKind string

const (
	LocationKindCountry LocationKind = "country"
	LocationKindRegion  LocationKind = "region"
	LocationKindCity    LocationKind = "city"
)

// SearchLocationsOptions filters canonical keyword locations.
type SearchLocationsOptions struct {
	Country string
	Limit   int
	Query   string
}

// LocationSuggestion is a canonical country, region, or city location.
type LocationSuggestion struct {
	CityName      *string      `json:"city_name"`
	CountryCode   string       `json:"country_code"`
	DisplayName   string       `json:"display_name"`
	HL            string       `json:"hl"`
	Kind          LocationKind `json:"kind"`
	LanguageCode  string       `json:"language_code"`
	LanguageLabel string       `json:"language_label"`
	LocationKey   string       `json:"location_key"`
	RegionCode    *string      `json:"region_code"`
	RegionName    *string      `json:"region_name"`
}

// LocationSuggestionsResponse contains canonical locations and compatibility pagination metadata.
type LocationSuggestionsResponse struct {
	Data []LocationSuggestion `json:"data"`
	Meta ListMeta             `json:"meta"`
}

// UpdateTeamMemberRoleInput changes a non-owner project member's role.
type UpdateTeamMemberRoleInput struct {
	Role TeamRoleValue `json:"role"`
}

// TeamMemberRoleResult is returned after changing a project member's role.
type TeamMemberRoleResult struct {
	ID   string        `json:"id"`
	Role TeamRoleValue `json:"role"`
}

// TeamMemberMutationResult is returned after removing a project member.
type TeamMemberMutationResult struct {
	ID string `json:"id"`
}

// TeamInviteResendResult is returned after resending a project invite.
type TeamInviteResendResult struct {
	ExpiresAt  time.Time `json:"expires_at"`
	ID         string    `json:"id"`
	InviteLink string    `json:"invite_link"`
}

// ListTrafficSnapshotsOptions filters stored page traffic snapshots by date and path.
type ListTrafficSnapshotsOptions struct {
	EndDate   string
	Limit     int
	Offset    int
	Paths     []string
	StartDate string
}

// PageTrafficSnapshot is one stored page analytics observation.
type PageTrafficSnapshot struct {
	BounceRate           *float64  `json:"bounce_rate"`
	CreatedAt            time.Time `json:"created_at"`
	Date                 string    `json:"date"`
	EngagementRate       *float64  `json:"engagement_rate"`
	KeyEvents            *float64  `json:"key_events"`
	Path                 string    `json:"path"`
	Provider             string    `json:"provider"`
	ScrollDepth          *float64  `json:"scroll_depth"`
	Sessions             int       `json:"sessions"`
	UpdatedAt            time.Time `json:"updated_at"`
	VisitDurationSeconds *float64  `json:"visit_duration_seconds"`
	Visitors             *int      `json:"visitors"`
	WindowDays           int       `json:"window_days"`
}

// PageTrafficSnapshotsResponse is one offset-paginated page of traffic snapshots.
type PageTrafficSnapshotsResponse struct {
	Offset     int                   `json:"offset"`
	Rows       []PageTrafficSnapshot `json:"rows"`
	TotalCount int                   `json:"total_count"`
}

// ListSearchPerformanceQueryStatsOptions controls a live connected-provider query.
type ListSearchPerformanceQueryStatsOptions struct {
	ConnectionID string
	EndDate      string
	Limit        int
	Query        string
	StartDate    string
}

// AnalyticsConnection identifies the selected project-owned analytics connection.
type AnalyticsConnection struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Provider string `json:"provider"`
}

// SearchPerformanceQueryStat is one live search-performance query row.
type SearchPerformanceQueryStat struct {
	Clicks      int     `json:"clicks"`
	CTR         float64 `json:"ctr"`
	Impressions int     `json:"impressions"`
	Page        *string `json:"page"`
	Position    float64 `json:"position"`
	Query       string  `json:"query"`
}

// SearchPerformanceQueryStatsResponse contains the selected connection and its rows.
type SearchPerformanceQueryStatsResponse struct {
	Connection AnalyticsConnection          `json:"connection"`
	Rows       []SearchPerformanceQueryStat `json:"rows"`
}

// TrafficSyncRunStatus reports one analytics connection sync outcome.
type TrafficSyncRunStatus string

const (
	TrafficSyncRunSucceededWithData TrafficSyncRunStatus = "succeeded_with_data"
	TrafficSyncRunSucceededEmpty    TrafficSyncRunStatus = "succeeded_empty"
	TrafficSyncRunDeferredRateLimit TrafficSyncRunStatus = "deferred_rate_limit"
	TrafficSyncRunFailed            TrafficSyncRunStatus = "failed"
	TrafficSyncRunNotApplicable     TrafficSyncRunStatus = "not_applicable"
)

// TrafficSyncRun is one provider connection's sync result.
type TrafficSyncRun struct {
	ConnectionID string               `json:"connection_id"`
	Error        *string              `json:"error,omitempty"`
	ErrorClass   *string              `json:"error_class,omitempty"`
	Provider     string               `json:"provider"`
	RowsFetched  int                  `json:"rows_fetched"`
	RowsMatched  int                  `json:"rows_matched"`
	RowsUpserted int                  `json:"rows_upserted"`
	Status       TrafficSyncRunStatus `json:"status"`
	Truncated    bool                 `json:"truncated"`
}

// TrafficSyncSkipReason explains why a provider did not run.
type TrafficSyncSkipReason string

const (
	TrafficSyncSkipNoCapability TrafficSyncSkipReason = "no_capability"
	TrafficSyncSkipRateLimited  TrafficSyncSkipReason = "rate_limited"
)

// TrafficSyncSkipped is one provider skipped during a project sync.
type TrafficSyncSkipped struct {
	Provider string                `json:"provider"`
	Reason   TrafficSyncSkipReason `json:"reason"`
}

// TrafficSyncSummary summarizes an on-demand project traffic sync.
type TrafficSyncSummary struct {
	Connections      int                  `json:"connections"`
	KeywordSnapshots int                  `json:"keyword_snapshots"`
	PageSnapshots    int                  `json:"page_snapshots"`
	ProjectID        string               `json:"project_id"`
	Runs             []TrafficSyncRun     `json:"runs"`
	Skipped          []TrafficSyncSkipped `json:"skipped"`
}
