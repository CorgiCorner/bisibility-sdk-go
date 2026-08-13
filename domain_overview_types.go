package bisibility

import (
	"encoding/json"
	"time"
)

// DomainOverviewScope selects a registrable root domain or one subdomain.
type DomainOverviewScope string

const (
	DomainOverviewScopeRoot      DomainOverviewScope = "root"
	DomainOverviewScopeSubdomain DomainOverviewScope = "subdomain"
)

// DomainOverviewState describes whether the core and optional modules returned data.
type DomainOverviewState string

const (
	DomainOverviewStateOK      DomainOverviewState = "ok"
	DomainOverviewStatePartial DomainOverviewState = "partial"
	DomainOverviewStateNoData  DomainOverviewState = "no_data"
)

// DomainOverviewHistoryMode identifies how historical data is loaded.
type DomainOverviewHistoryMode string

const DomainOverviewHistoryModeLazy DomainOverviewHistoryMode = "lazy"

// DomainOverviewFailureReason is the machine-readable reason for a failed nested report module.
type DomainOverviewFailureReason string

const (
	DomainOverviewFailureBudgetExhausted     DomainOverviewFailureReason = "budget_exhausted"
	DomainOverviewFailureCostLimitExceeded   DomainOverviewFailureReason = "cost_limit_exceeded"
	DomainOverviewFailureInProgress          DomainOverviewFailureReason = "in_progress"
	DomainOverviewFailureLookupFailed        DomainOverviewFailureReason = "lookup_failed"
	DomainOverviewFailureNeedsReauth         DomainOverviewFailureReason = "needs_reauth"
	DomainOverviewFailureNoSource            DomainOverviewFailureReason = "no_source"
	DomainOverviewFailureRateLimited         DomainOverviewFailureReason = "rate_limited"
	DomainOverviewFailureSnapshotExpired     DomainOverviewFailureReason = "snapshot_expired"
	DomainOverviewFailureUnsupportedLocation DomainOverviewFailureReason = "unsupported_location"
)

// DomainOverviewKeywordIntent is the provider-classified search intent.
type DomainOverviewKeywordIntent string

const (
	DomainOverviewIntentInformational DomainOverviewKeywordIntent = "informational"
	DomainOverviewIntentNavigational  DomainOverviewKeywordIntent = "navigational"
	DomainOverviewIntentCommercial    DomainOverviewKeywordIntent = "commercial"
	DomainOverviewIntentTransactional DomainOverviewKeywordIntent = "transactional"
)

// AnalyzeDomainOverviewOptions controls a cache-aware estimate or domain overview analysis.
type AnalyzeDomainOverviewOptions struct {
	Target        string               `json:"target"`
	LocationCode  int                  `json:"location_code"`
	LanguageCode  string               `json:"language_code"`
	ScopeOverride *DomainOverviewScope `json:"scope_override,omitempty"`
	Fresh         *bool                `json:"fresh,omitempty"`
	MaxCostCents  *int                 `json:"max_cost_cents,omitempty"`
	EstimateOnly  *bool                `json:"estimate_only,omitempty"`
	KeywordLimit  *int                 `json:"keyword_limit,omitempty"`
	PageLimit     *int                 `json:"page_limit,omitempty"`
}

// LoadDomainOverviewHistoryOptions controls a historical index load.
type LoadDomainOverviewHistoryOptions struct {
	Target        string               `json:"target"`
	LocationCode  int                  `json:"location_code"`
	LanguageCode  string               `json:"language_code"`
	ScopeOverride *DomainOverviewScope `json:"scope_override,omitempty"`
	Fresh         *bool                `json:"fresh,omitempty"`
	MaxCostCents  int                  `json:"max_cost_cents"`
}

// LoadDomainOverviewKeywordsOptions controls one ranked-keyword page load.
type LoadDomainOverviewKeywordsOptions struct {
	Target        string               `json:"target"`
	LocationCode  int                  `json:"location_code"`
	LanguageCode  string               `json:"language_code"`
	ScopeOverride *DomainOverviewScope `json:"scope_override,omitempty"`
	Fresh         *bool                `json:"fresh,omitempty"`
	MaxCostCents  int                  `json:"max_cost_cents"`
	Limit         int                  `json:"limit"`
	Offset        int                  `json:"offset"`
}

// LoadDomainOverviewPagesOptions controls one relevant-page page load.
type LoadDomainOverviewPagesOptions = LoadDomainOverviewKeywordsOptions

// DomainRankMetrics contains nullable provider estimates and organic position buckets.
type DomainRankMetrics struct {
	Count                     *int     `json:"count"`
	EstimatedTrafficCostCents *float64 `json:"estimated_traffic_cost_cents"`
	ETV                       *float64 `json:"etv"`
	IsDown                    int      `json:"is_down"`
	IsLost                    int      `json:"is_lost"`
	IsNew                     int      `json:"is_new"`
	IsUp                      int      `json:"is_up"`
	Pos1                      int      `json:"pos1"`
	Pos2To3                   int      `json:"pos2_3"`
	Pos4To10                  int      `json:"pos4_10"`
	Pos11To20                 int      `json:"pos11_20"`
	Pos21To30                 int      `json:"pos21_30"`
	Pos31To40                 int      `json:"pos31_40"`
	Pos41To50                 int      `json:"pos41_50"`
	Pos51To60                 int      `json:"pos51_60"`
	Pos61To70                 int      `json:"pos61_70"`
	Pos71To80                 int      `json:"pos71_80"`
	Pos81To90                 int      `json:"pos81_90"`
	Pos91To100                int      `json:"pos91_100"`
}

// HistoricalOverviewRow is one monthly point in the historical rank overview.
type HistoricalOverviewRow struct {
	Metrics DomainRankMetrics `json:"metrics"`
	Month   int               `json:"month"`
	Year    int               `json:"year"`
}

// DomainOverviewRankedKeyword is one organic keyword row.
type DomainOverviewRankedKeyword struct {
	Keyword           string                       `json:"keyword"`
	Position          *float64                     `json:"position"`
	SearchVolume      *float64                     `json:"search_volume"`
	EstimatedTraffic  *float64                     `json:"estimated_traffic"`
	CPCCents          *float64                     `json:"cpc_cents"`
	Difficulty        *float64                     `json:"difficulty"`
	Intent            *DomainOverviewKeywordIntent `json:"intent"`
	RankingURL        *string                      `json:"ranking_url"`
	SERPFeatures      []string                     `json:"serp_features"`
	RankAbsoluteDelta *float64                     `json:"rank_absolute_delta"`
	RankAbsolute      *float64                     `json:"rank_absolute"`
}

// DomainOverviewRankedKeywordsPage is one provider page of organic keywords.
type DomainOverviewRankedKeywordsPage struct {
	Rows       []DomainOverviewRankedKeyword `json:"rows"`
	TotalCount *int                          `json:"total_count"`
	CostCents  float64                       `json:"cost_cents"`
}

// DomainOverviewRelevantPage is one provider-ranked page for the target.
type DomainOverviewRelevantPage struct {
	ETV                *float64 `json:"etv"`
	ETVDeltaPct        *float64 `json:"etv_delta_pct"`
	KeywordCount       *int     `json:"keyword_count"`
	Path               string   `json:"path"`
	TopKeyword         *string  `json:"top_keyword"`
	TopKeywordPosition *float64 `json:"top_keyword_position"`
}

// DomainOverviewRelevantPagesPage is one provider page of target URLs.
type DomainOverviewRelevantPagesPage struct {
	Rows       []DomainOverviewRelevantPage `json:"rows"`
	TotalCount int                          `json:"total_count"`
	CostCents  float64                      `json:"cost_cents"`
}

// DomainOverviewModuleOutcome is a successful or failed module embedded in an analysis report.
// Success fields and failure fields remain pointer-valued when absent from the corresponding JSON
// variant; CostCents and OK are present in both variants.
type DomainOverviewModuleOutcome[T any] struct {
	OK        bool                         `json:"ok"`
	Cached    *bool                        `json:"cached,omitempty"`
	CostCents float64                      `json:"cost_cents"`
	Data      *T                           `json:"data,omitempty"`
	FetchedAt *time.Time                   `json:"fetched_at,omitempty"`
	Reason    *DomainOverviewFailureReason `json:"reason,omitempty"`
	ResetAt   *float64                     `json:"reset_at,omitempty"`
}

// DomainOverviewEstimate is the cache-aware estimate variant returned by AnalyzeDomainOverview.
type DomainOverviewEstimate struct {
	Cached                        bool                      `json:"cached"`
	Estimate                      bool                      `json:"estimate"`
	EstimatedCostCents            float64                   `json:"estimated_cost_cents"`
	FreshEstimatedCostCents       float64                   `json:"fresh_estimated_cost_cents"`
	HistoryEstimatedCostCents     float64                   `json:"history_estimated_cost_cents"`
	HistoryMode                   DomainOverviewHistoryMode `json:"history_mode"`
	KeywordPageEstimatedCostCents float64                   `json:"keyword_page_estimated_cost_cents"`
	LanguageCode                  string                    `json:"language_code"`
	LocationCode                  int                       `json:"location_code"`
	PagePageEstimatedCostCents    float64                   `json:"page_page_estimated_cost_cents"`
	Provider                      string                    `json:"provider"`
	Scope                         DomainOverviewScope       `json:"scope"`
	Target                        string                    `json:"target"`
}

// DomainOverviewReport is the report variant returned by AnalyzeDomainOverview.
type DomainOverviewReport struct {
	Cached                   bool                                                          `json:"cached"`
	CachedUntil              time.Time                                                     `json:"cached_until"`
	CostCents                float64                                                       `json:"cost_cents"`
	FetchedAt                time.Time                                                     `json:"fetched_at"`
	HistoryMode              DomainOverviewHistoryMode                                     `json:"history_mode"`
	Keywords                 DomainOverviewModuleOutcome[DomainOverviewRankedKeywordsPage] `json:"keywords"`
	LanguageCode             string                                                        `json:"language_code"`
	LocationCode             int                                                           `json:"location_code"`
	Overview                 *DomainRankMetrics                                            `json:"overview"`
	Pages                    DomainOverviewModuleOutcome[DomainOverviewRelevantPagesPage]  `json:"pages"`
	PreviousFetchedAt        *time.Time                                                    `json:"previous_fetched_at"`
	PreviousOverview         *DomainRankMetrics                                            `json:"previous_overview"`
	PreviousSourceSnapshotAt *time.Time                                                    `json:"previous_source_snapshot_at"`
	Provider                 string                                                        `json:"provider"`
	Scope                    DomainOverviewScope                                           `json:"scope"`
	SourceSnapshotAt         *time.Time                                                    `json:"source_snapshot_at"`
	State                    DomainOverviewState                                           `json:"state"`
	Target                   string                                                        `json:"target"`
}

// DomainOverviewAnalyzeResult is the discriminated estimate/report response data.
// Exactly one pointer is set after successful decoding.
type DomainOverviewAnalyzeResult struct {
	Estimate *DomainOverviewEstimate
	Report   *DomainOverviewReport
}

// UnmarshalJSON decodes an AnalyzeDomainOverview result using estimate=true as its discriminator.
func (result *DomainOverviewAnalyzeResult) UnmarshalJSON(data []byte) error {
	var discriminator struct {
		Estimate bool `json:"estimate"`
	}
	if err := json.Unmarshal(data, &discriminator); err != nil {
		return err
	}
	if discriminator.Estimate {
		var estimate DomainOverviewEstimate
		if err := json.Unmarshal(data, &estimate); err != nil {
			return err
		}
		result.Estimate = &estimate
		result.Report = nil
		return nil
	}
	var report DomainOverviewReport
	if err := json.Unmarshal(data, &report); err != nil {
		return err
	}
	result.Estimate = nil
	result.Report = &report
	return nil
}

// DomainOverviewModuleSuccess is the top-level success data for history and table operations.
type DomainOverviewModuleSuccess[T any] struct {
	Cached    bool      `json:"cached"`
	CostCents float64   `json:"cost_cents"`
	Data      T         `json:"data"`
	FetchedAt time.Time `json:"fetched_at"`
}

// DomainOverviewProblemErrors is the errors extension on a failed Domain Overview API problem.
// Decode APIError.Problem.Errors into this type to preserve charged cost and retry timing.
type DomainOverviewProblemErrors struct {
	CostCents float64                     `json:"cost_cents"`
	Reason    DomainOverviewFailureReason `json:"reason"`
	ResetAt   *float64                    `json:"reset_at,omitempty"`
}

// DomainOverviewAnalyzeResponse wraps an estimate or report in the public API data envelope.
type DomainOverviewAnalyzeResponse = DataResponse[DomainOverviewAnalyzeResult]

// DomainOverviewHistoryResponse wraps a historical index series in the public API data envelope.
type DomainOverviewHistoryResponse = DataResponse[DomainOverviewModuleSuccess[[]HistoricalOverviewRow]]

// DomainOverviewKeywordsResponse wraps a ranked-keyword page in the public API data envelope.
type DomainOverviewKeywordsResponse = DataResponse[DomainOverviewModuleSuccess[DomainOverviewRankedKeywordsPage]]

// DomainOverviewPagesResponse wraps a relevant-page page in the public API data envelope.
type DomainOverviewPagesResponse = DataResponse[DomainOverviewModuleSuccess[DomainOverviewRelevantPagesPage]]
