package bisibility

import "time"

// SearchIndexSitemapContent is one submitted content category in a sitemap.
type SearchIndexSitemapContent struct {
	Type      string `json:"type"`
	Submitted int    `json:"submitted"`
}

// SearchIndexSitemap describes a sitemap as already known to Google.
type SearchIndexSitemap struct {
	Path            string                      `json:"path"`
	Type            *string                     `json:"type"`
	IsSitemapsIndex bool                        `json:"is_sitemaps_index"`
	IsPending       bool                        `json:"is_pending"`
	Errors          int                         `json:"errors"`
	Warnings        int                         `json:"warnings"`
	LastSubmitted   *string                     `json:"last_submitted"`
	LastDownloaded  *string                     `json:"last_downloaded"`
	Contents        []SearchIndexSitemapContent `json:"contents"`
}

// SearchIndexSitemaps contains sitemaps for the project's connected property.
type SearchIndexSitemaps struct {
	Property string               `json:"property"`
	Sitemaps []SearchIndexSitemap `json:"sitemaps"`
}

// SearchIndexSitemapResponse contains one sitemap for the connected property.
type SearchIndexSitemapResponse struct {
	Property string             `json:"property"`
	Sitemap  SearchIndexSitemap `json:"sitemap"`
}

// SearchIndexInspectionInput requests 1-10 HTTP(S) URLs within the connected property.
// The server normalizes URLs and counts duplicates after normalization once.
type SearchIndexInspectionInput struct {
	URLs []string `json:"urls"`
}

// SearchIndexInspectionErrorCode is an open per-URL failure code.
type SearchIndexInspectionErrorCode string

const (
	SearchIndexInspectionErrorCodeInspectionDailyLimit  SearchIndexInspectionErrorCode = "inspection_daily_limit"
	SearchIndexInspectionErrorCodeInspectionFailed      SearchIndexInspectionErrorCode = "inspection_failed"
	SearchIndexInspectionErrorCodeProviderAuthorization SearchIndexInspectionErrorCode = "provider_authorization"
	SearchIndexInspectionErrorCodeProviderRateLimited   SearchIndexInspectionErrorCode = "provider_rate_limited"
)

// SearchIndexRateLimitScope is an open provider rate-limit scope.
type SearchIndexRateLimitScope string

const (
	SearchIndexRateLimitScopeDaily   SearchIndexRateLimitScope = "daily"
	SearchIndexRateLimitScopeMinute  SearchIndexRateLimitScope = "minute"
	SearchIndexRateLimitScopeUnknown SearchIndexRateLimitScope = "unknown"
)

// SearchIndexInspectionError describes one failed URL. Provider rate limits
// include ResetAt and Scope and stop further Google calls in the batch.
type SearchIndexInspectionError struct {
	Code    SearchIndexInspectionErrorCode `json:"code"`
	Message string                         `json:"message"`
	ResetAt *time.Time                     `json:"reset_at,omitempty"`
	Scope   *SearchIndexRateLimitScope     `json:"scope,omitempty"`
}

// SearchIndexIssueCount aggregates rich-result issues by provider severity.
type SearchIndexIssueCount struct {
	Severity *string `json:"severity"`
	Count    int     `json:"count"`
}

// SearchIndexDetectedItem summarizes a detected rich-result category.
type SearchIndexDetectedItem struct {
	RichResultType *string                 `json:"rich_result_type"`
	ItemCount      int                     `json:"item_count"`
	IssueCount     int                     `json:"issue_count"`
	Issues         []SearchIndexIssueCount `json:"issues"`
}

// SearchIndexRichResults contains Google's known rich-result summary.
type SearchIndexRichResults struct {
	Verdict       *string                   `json:"verdict"`
	DetectedItems []SearchIndexDetectedItem `json:"detected_items"`
}

// SearchIndexInspection is Google's indexed state, not a live test.
// Values may be partial or delayed; unknown provider state strings are preserved.
type SearchIndexInspection struct {
	Verdict              *string                 `json:"verdict"`
	CoverageState        *string                 `json:"coverage_state"`
	RobotsTxtState       *string                 `json:"robots_txt_state"`
	IndexingState        *string                 `json:"indexing_state"`
	LastCrawlTime        *string                 `json:"last_crawl_time"`
	LastCrawlAt          *time.Time              `json:"last_crawl_at"`
	PageFetchState       *string                 `json:"page_fetch_state"`
	GoogleCanonical      *string                 `json:"google_canonical"`
	UserCanonical        *string                 `json:"user_canonical"`
	CrawledAs            *string                 `json:"crawled_as"`
	ReferringURLs        []string                `json:"referring_urls"`
	Sitemap              []string                `json:"sitemap"`
	InspectionResultLink *string                 `json:"inspection_result_link"`
	RichResultsResult    *SearchIndexRichResults `json:"rich_results_result"`
}

// SearchIndexInspectionRow contains a per-URL result or error.
type SearchIndexInspectionRow struct {
	URL        string                      `json:"url"`
	Inspection *SearchIndexInspection      `json:"inspection"`
	Error      *SearchIndexInspectionError `json:"error"`
}

// SearchIndexInspections contains results for the project's connected property.
type SearchIndexInspections struct {
	Property string                     `json:"property"`
	Results  []SearchIndexInspectionRow `json:"results"`
}

// PageIndexingBudgetMode is an open background inspection mode.
// AvailableQuota is offered only on self-hosted deployments.
type PageIndexingBudgetMode string

const (
	PageIndexingBudgetModeOff            PageIndexingBudgetMode = "off"
	PageIndexingBudgetModeFixed          PageIndexingBudgetMode = "fixed"
	PageIndexingBudgetModeAvailableQuota PageIndexingBudgetMode = "available_quota"
)

// PageIndexingGroupKind is an open stored result group.
type PageIndexingGroupKind string

const (
	PageIndexingGroupKindIndexed          PageIndexingGroupKind = "indexed"
	PageIndexingGroupKindNotIndexed       PageIndexingGroupKind = "not_indexed"
	PageIndexingGroupKindInspectionFailed PageIndexingGroupKind = "inspection_failed"
	PageIndexingGroupKindNotChecked       PageIndexingGroupKind = "not_checked"
)

// PageIndexingSource is an open candidate or inspection source.
type PageIndexingSource string

const (
	PageIndexingSourceKeywordTarget PageIndexingSource = "keyword_target"
	PageIndexingSourceSitemap       PageIndexingSource = "sitemap"
	PageIndexingSourceManual        PageIndexingSource = "manual"
)

// PageIndexingReason counts one Google coverage state and records its oldest stored result.
// Synthetic groups use a null CoverageState; OldestResultAt is null without a stored result.
type PageIndexingReason struct {
	CoverageState  *string    `json:"coverage_state"`
	Total          int        `json:"total"`
	OldestResultAt *time.Time `json:"oldest_result_at"`
}

// PageIndexingGroup describes completeness, freshness and usage for one group.
type PageIndexingGroup struct {
	Group          PageIndexingGroupKind  `json:"group"`
	Checked        int                    `json:"checked"`
	Total          int                    `json:"total"`
	Complete       bool                   `json:"complete"`
	Remaining      int                    `json:"remaining"`
	OldestResultAt *time.Time             `json:"oldest_result_at"`
	NewestResultAt *time.Time             `json:"newest_result_at"`
	BudgetMode     PageIndexingBudgetMode `json:"budget_mode"`
	TodayCount     int                    `json:"today_count"`
	Last429Kind    *string                `json:"last429_kind"`
	NextResetAt    time.Time              `json:"next_reset_at"`
	Reasons        []PageIndexingReason   `json:"reasons"`
}

// PageIndexingReport describes stored results for the connected property.
// Complete=false must not be treated as a full-site census.
type PageIndexingReport struct {
	Property               string                 `json:"property"`
	Checked                int                    `json:"checked"`
	Total                  int                    `json:"total"`
	Complete               bool                   `json:"complete"`
	Remaining              int                    `json:"remaining"`
	OldestResultAt         *time.Time             `json:"oldest_result_at"`
	NewestResultAt         *time.Time             `json:"newest_result_at"`
	BudgetMode             PageIndexingBudgetMode `json:"budget_mode"`
	TodayCount             int                    `json:"today_count"`
	Last429Kind            *string                `json:"last429_kind"`
	NextResetAt            time.Time              `json:"next_reset_at"`
	DiscoveryComplete      bool                   `json:"discovery_complete"`
	DiscoveredAt           *time.Time             `json:"discovered_at"`
	EstimatedDaysRemaining *int                   `json:"estimated_days_remaining"`
	Groups                 []PageIndexingGroup    `json:"groups"`
}

// PageIndexingSettings contains background settings, supported modes and estimated usage.
type PageIndexingSettings struct {
	Property             string                   `json:"property"`
	BudgetMode           PageIndexingBudgetMode   `json:"budget_mode"`
	InspectionDailyLimit int                      `json:"inspection_daily_limit"`
	AvailableModes       []PageIndexingBudgetMode `json:"available_modes"`
	TodayCount           int                      `json:"today_count"`
	Last429Kind          *string                  `json:"last429_kind"`
	Last429At            *time.Time               `json:"last429_at"`
	NextResetAt          time.Time                `json:"next_reset_at"`
}

// PageIndexingSettingsPatch changes at least one setting. Nil omits a field;
// null is not accepted. InspectionDailyLimit accepts 0-2000 inclusive.
// BudgetMode accepts off, fixed, or available_quota (self-host only).
type PageIndexingSettingsPatch struct {
	BudgetMode           *PageIndexingBudgetMode `json:"budget_mode,omitempty"`
	InspectionDailyLimit *int                    `json:"inspection_daily_limit,omitempty"`
}

// ListPageIndexingURLsOptions filters a cursor page of stored URLs.
// CoverageState is an exact Google state; Unspecified matches a missing state
// in indexed/not_indexed groups. For synthetic groups, filter by Group alone.
type ListPageIndexingURLsOptions struct {
	Group         PageIndexingGroupKind
	CoverageState string
	Sitemap       string
	Source        PageIndexingSource
	Search        string
	Cursor        string
	Limit         int
	History       bool
	// ChangedSince includes URLs with a stored tracked change at or after this time,
	// applied before pagination.
	ChangedSince *time.Time
}

// PageIndexingTrackedState contains the fields recorded in change history.
type PageIndexingTrackedState struct {
	Verdict         *string `json:"verdict"`
	CoverageState   *string `json:"coverage_state"`
	GoogleCanonical *string `json:"google_canonical"`
	IndexingState   *string `json:"indexing_state"`
	RobotsTxtState  *string `json:"robots_txt_state"`
	PageFetchState  *string `json:"page_fetch_state"`
}

// PageIndexingChange records one stored state transition.
type PageIndexingChange struct {
	At       time.Time                `json:"at"`
	Previous PageIndexingTrackedState `json:"previous"`
	Next     PageIndexingTrackedState `json:"next"`
}

// PageIndexingHistory contains up to 100 recent changes and their completeness.
type PageIndexingHistory struct {
	Complete bool                 `json:"complete"`
	Changes  []PageIndexingChange `json:"changes"`
}

// PageIndexingRichResults contains the same rich-result summary as explicit inspections.
type PageIndexingRichResults = SearchIndexRichResults

// PageIndexingURL is one stored URL and Google's known indexed state.
// CheckedAt records the last result; InspectionFailedAt records a failed attempt.
// Results may be partial or delayed and are not a live test.
type PageIndexingURL struct {
	URL                  string                   `json:"url"`
	Verdict              *string                  `json:"verdict"`
	CoverageState        *string                  `json:"coverage_state"`
	LastCrawlAt          *time.Time               `json:"last_crawl_at"`
	CanonicalOK          *bool                    `json:"canonical_ok"`
	CheckedAt            *time.Time               `json:"checked_at"`
	CreatedAt            time.Time                `json:"created_at"`
	UpdatedAt            time.Time                `json:"updated_at"`
	Source               PageIndexingSource       `json:"source"`
	SitemapURL           *string                  `json:"sitemap_url"`
	GoogleCanonical      *string                  `json:"google_canonical"`
	UserCanonical        *string                  `json:"user_canonical"`
	IndexingState        *string                  `json:"indexing_state"`
	RobotsTxtState       *string                  `json:"robots_txt_state"`
	PageFetchState       *string                  `json:"page_fetch_state"`
	CrawledAs            *string                  `json:"crawled_as"`
	ReferringURLs        []string                 `json:"referring_urls"`
	Sitemaps             []string                 `json:"sitemaps"`
	RichResultsSummary   *PageIndexingRichResults `json:"rich_results_summary"`
	InspectionResultLink *string                  `json:"inspection_result_link"`
	InspectionError      *string                  `json:"inspection_error"`
	InspectionFailedAt   *time.Time               `json:"inspection_failed_at"`
	Group                PageIndexingGroupKind    `json:"group"`
	History              *PageIndexingHistory     `json:"history,omitempty"`
}

// PageIndexingURLs contains one cursor page of stored URLs.
type PageIndexingURLs = ListResponse[PageIndexingURL]
