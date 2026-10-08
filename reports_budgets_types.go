package bisibility

import (
	"encoding/json"
	"fmt"
)

// StoredBacklinksReport describes the API response.
type StoredBacklinksReport struct {
	BacklinksSnapshot
	FreshUntil string `json:"fresh_until"`
	SavedAt    string `json:"saved_at"`
	Stale      bool   `json:"stale"`
	State      string `json:"state"`
}

// StoredDomainOverviewReport describes the API response.
type StoredDomainOverviewReport struct {
	Cached                   bool                              `json:"cached"`
	CostCents                float64                           `json:"cost_cents"`
	CountryCode              *string                           `json:"country_code"`
	DataState                string                            `json:"data_state"`
	FetchedAt                string                            `json:"fetched_at"`
	FreshUntil               string                            `json:"fresh_until"`
	History                  *[]HistoricalOverviewRow          `json:"history"`
	Keywords                 *DomainOverviewRankedKeywordsPage `json:"keywords"`
	LanguageCode             string                            `json:"language_code"`
	LocationCode             int                               `json:"location_code"`
	Overview                 *DomainRankMetrics                `json:"overview"`
	Pages                    *DomainOverviewRelevantPagesPage  `json:"pages"`
	Partial                  bool                              `json:"partial"`
	PreviousFetchedAt        *string                           `json:"previous_fetched_at"`
	PreviousOverview         *DomainRankMetrics                `json:"previous_overview"`
	PreviousSourceSnapshotAt *string                           `json:"previous_source_snapshot_at"`
	Provider                 string                            `json:"provider"`
	SavedAt                  string                            `json:"saved_at"`
	Scope                    string                            `json:"scope"`
	SourceSnapshotAt         *string                           `json:"source_snapshot_at"`
	Stale                    bool                              `json:"stale"`
	State                    string                            `json:"state"`
	Target                   string                            `json:"target"`
}

// StoredKeywordResearchReport describes the API response.
type StoredKeywordResearchReport struct {
	Cached             bool    `json:"cached"`
	CostCents          float64 `json:"cost_cents"`
	CountryCode        string  `json:"country_code"`
	FetchedAt          string  `json:"fetched_at"`
	FreshUntil         string  `json:"fresh_until"`
	IncludeClickstream bool    `json:"include_clickstream"`
	LanguageCode       string  `json:"language_code"`
	Mode               string  `json:"mode"`
	Partial            bool    `json:"partial"`
	Provider           string  `json:"provider"`
	RequestKey         string  `json:"request_key"`
	ResultLimit        int     `json:"result_limit"`
	Rows               []struct {
		AlreadySaved   bool     `json:"already_saved"`
		AlreadyTracked bool     `json:"already_tracked"`
		Competition    *float64 `json:"competition"`
		CPCCents       *int     `json:"cpc_cents"`
		Difficulty     *int     `json:"difficulty"`
		Intent         *string  `json:"intent"`
		Keyword        string   `json:"keyword"`
		MonthlyTrend   []struct {
			Month        int      `json:"month"`
			SearchVolume *float64 `json:"search_volume"`
			Year         int      `json:"year"`
		} `json:"monthly_trend"`
		SearchVolume *float64 `json:"search_volume"`
		Source       string   `json:"source"`
	} `json:"rows"`
	SavedAt string `json:"saved_at"`
	Seed    string `json:"seed"`
	Sources []struct {
		Cached    bool    `json:"cached"`
		CostCents float64 `json:"cost_cents"`
		Reason    string  `json:"reason,omitempty"`
		Returned  int     `json:"returned"`
		Source    string  `json:"source"`
		Status    string  `json:"status"`
	} `json:"sources"`
	Stale bool   `json:"stale"`
	State string `json:"state"`
}

// StoredResearchReportSummary describes the API response.
type StoredResearchReportSummary struct {
	CountryCode        string `json:"country_code,omitempty"`
	FreshUntil         string `json:"fresh_until"`
	IncludeClickstream bool   `json:"include_clickstream,omitempty"`
	IncludeSubdomains  bool   `json:"include_subdomains,omitempty"`
	Kind               string `json:"kind"`
	LanguageCode       string `json:"language_code,omitempty"`
	LocationCode       int    `json:"location_code,omitempty"`
	Mode               string `json:"mode,omitempty"`
	ResultLimit        int    `json:"result_limit,omitempty"`
	SavedAt            string `json:"saved_at"`
	Seed               string `json:"seed,omitempty"`
	State              string `json:"state"`
	Target             string `json:"target,omitempty"`
	TargetScope        string `json:"target_scope,omitempty"`
}

// StoredResearchReportsResponse describes the API response.
type StoredResearchReportsResponse struct {
	Data []StoredResearchReportSummary `json:"data"`
	Meta struct {
		FreshnessDays int `json:"freshness_days"`
	} `json:"meta"`
}

// ProviderBudgets describes the API response.
type ProviderBudgets struct {
	ConnectionID     string `json:"connection_id"`
	CredentialSource string `json:"credential_source"`
	Credits          struct {
		App *struct {
			AmountPerMonth int    `json:"amount_per_month"`
			Unit           string `json:"unit"`
		} `json:"app"`
		Programmatic *struct {
			AmountPerMonth int    `json:"amount_per_month"`
			Unit           string `json:"unit"`
		} `json:"programmatic"`
	} `json:"credits"`
	Own struct {
		App *struct {
			AmountPerMonth int    `json:"amount_per_month"`
			Unit           string `json:"unit"`
		} `json:"app"`
		Programmatic *struct {
			AmountPerMonth int    `json:"amount_per_month"`
			Unit           string `json:"unit"`
		} `json:"programmatic"`
	} `json:"own"`
	Provider string `json:"provider"`
	Source   string `json:"source"`
}

// StoredResearchReport holds exactly one decoded stored report.
type StoredResearchReport struct {
	Backlinks       *StoredBacklinksReport
	DomainOverview  *StoredDomainOverviewReport
	KeywordResearch *StoredKeywordResearchReport
}

func (r *StoredResearchReport) UnmarshalJSON(data []byte) error {
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(data, &keys); err != nil {
		return err
	}
	*r = StoredResearchReport{}
	if _, ok := keys["data_state"]; ok {
		r.DomainOverview = new(StoredDomainOverviewReport)
		return json.Unmarshal(data, r.DomainOverview)
	}
	if _, ok := keys["request_key"]; ok {
		r.KeywordResearch = new(StoredKeywordResearchReport)
		return json.Unmarshal(data, r.KeywordResearch)
	}
	if _, ok := keys["summary"]; ok {
		r.Backlinks = new(StoredBacklinksReport)
		return json.Unmarshal(data, r.Backlinks)
	}
	return fmt.Errorf("unknown stored research report shape")
}

type StoredResearchReportResponse = DataResponse[StoredResearchReport]
type StoredResearchReportKind string

const (
	StoredReportBacklinks       StoredResearchReportKind = "backlinks"
	StoredReportDomainOverview  StoredResearchReportKind = "domain_overview"
	StoredReportKeywordResearch StoredResearchReportKind = "keyword_research"
)

// Pointer flags preserve an explicitly false value and an explicitly zero location code.
type StoredResearchReportOptions struct {
	Target             string
	TargetScope        string
	Mode               string
	IncludeSubdomains  *bool
	Seed               string
	IncludeClickstream *bool
	ResultLimit        int
	ConnectionID       string
	LanguageCode       string
	LocationCode       *int
}
type ProviderMonthlyBudget struct {
	AmountPerMonth int    `json:"amount_per_month"`
	Unit           string `json:"unit"`
}

// NullableProviderBudget distinguishes omission from an explicit budget deletion.
type NullableProviderBudget struct {
	set   bool
	value *ProviderMonthlyBudget
}

func ProviderBudgetValue(value ProviderMonthlyBudget) NullableProviderBudget {
	return NullableProviderBudget{set: true, value: &value}
}
func ClearProviderBudget() NullableProviderBudget { return NullableProviderBudget{set: true} }

type ProviderBudgetPatch struct {
	App          NullableProviderBudget
	Programmatic NullableProviderBudget
}

func (p ProviderBudgetPatch) MarshalJSON() ([]byte, error) {
	result := map[string]*ProviderMonthlyBudget{}
	if p.App.set {
		result["app"] = p.App.value
	}
	if p.Programmatic.set {
		result["programmatic"] = p.Programmatic.value
	}
	return json.Marshal(result)
}

// Nil surfaces are omitted. ClearProviderBudget clears one surface explicitly.
type ProviderBudgetsUpdate struct {
	Own     *ProviderBudgetPatch `json:"own,omitempty"`
	Credits *ProviderBudgetPatch `json:"credits,omitempty"`
}

func (p ProviderBudgetsUpdate) validate() error {
	for _, source := range []struct {
		patch   *ProviderBudgetPatch
		credits bool
	}{{p.Own, false}, {p.Credits, true}} {
		if source.patch == nil {
			continue
		}
		for _, budget := range []NullableProviderBudget{source.patch.App, source.patch.Programmatic} {
			if !budget.set || budget.value == nil {
				continue
			}
			b := budget.value
			if b.AmountPerMonth < 1 || b.AmountPerMonth > 2147483647 || (b.Unit != "cents" && b.Unit != "units") || (source.credits && b.Unit != "cents") {
				return fmt.Errorf("invalid provider budget")
			}
		}
	}
	return nil
}

// AgentReportSummary describes the API response.
type AgentReportSummary struct {
	CreatedAt string `json:"created_at"`
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Title     string `json:"title"`
}

// AgentReportResource describes the API response.
type AgentReportResource struct {
	Body       map[string]any `json:"body"`
	CreatedAt  string         `json:"created_at"`
	ID         string         `json:"id"`
	Kind       string         `json:"kind"`
	Provenance map[string]any `json:"provenance"`
	Title      string         `json:"title"`
}

// ProjectContext describes the API response.
type ProjectContext struct {
	AgentRules string  `json:"agent_rules"`
	Audience   string  `json:"audience"`
	Business   string  `json:"business"`
	Goals      string  `json:"goals"`
	Products   string  `json:"products"`
	UpdatedAt  *string `json:"updated_at"`
}

// SiteAuditLimits describes the API response.
type SiteAuditLimits struct {
	MaxDurationMS float64 `json:"max_duration_ms"`
	MaxPageBytes  float64 `json:"max_page_bytes"`
	MaxPages      float64 `json:"max_pages"`
	MaxRequests   float64 `json:"max_requests"`
}

// SiteAuditHeading describes the API response.
type SiteAuditHeading struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
}

// SiteAuditIssue describes the API response.
type SiteAuditIssue struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

// SiteAuditPage describes the API response.
type SiteAuditPage struct {
	Canonical         *string            `json:"canonical"`
	Description       *string            `json:"description"`
	ExternalLinkCount float64            `json:"external_link_count"`
	FinalURL          string             `json:"final_url"`
	H1Count           float64            `json:"h1_count"`
	Headings          []SiteAuditHeading `json:"headings"`
	ImageCount        float64            `json:"image_count"`
	Indexable         bool               `json:"indexable"`
	InternalLinkCount float64            `json:"internal_link_count"`
	InternalLinks     []string           `json:"internal_links"`
	Issues            []SiteAuditIssue   `json:"issues"`
	MissingAltCount   float64            `json:"missing_alt_count"`
	ResponseTimeMS    float64            `json:"response_time_ms"`
	Robots            *string            `json:"robots"`
	Status            *int               `json:"status"`
	Title             *string            `json:"title"`
	URL               string             `json:"url"`
}

// SiteAuditSummary describes the API response.
type SiteAuditSummary struct {
	Errors    float64 `json:"errors"`
	Indexable float64 `json:"indexable"`
	Pages     float64 `json:"pages"`
	Warnings  float64 `json:"warnings"`
}

// SiteAuditResult describes the API response.
type SiteAuditResult struct {
	CompletedAt string           `json:"completed_at"`
	Limitations []string         `json:"limitations"`
	Limits      SiteAuditLimits  `json:"limits"`
	Pages       []SiteAuditPage  `json:"pages"`
	Requests    float64          `json:"requests"`
	StartedAt   string           `json:"started_at"`
	State       string           `json:"state"`
	StopReason  string           `json:"stop_reason"`
	Summary     SiteAuditSummary `json:"summary"`
	Target      string           `json:"target"`
	Version     int              `json:"version"`
}

// SiteAuditReport describes the API response.
type SiteAuditReport struct {
	Cached    bool            `json:"cached"`
	CreatedAt string          `json:"created_at"`
	ID        string          `json:"id"`
	Result    SiteAuditResult `json:"result"`
}

// AIAnalysisCitation describes the API response.
type AIAnalysisCitation struct {
	TargetDomain bool   `json:"target_domain"`
	Title        string `json:"title"`
	URL          string `json:"url"`
}

// AIAnalysisRow describes the API response.
type AIAnalysisRow struct {
	Answer           string               `json:"answer"`
	BrandMentioned   bool                 `json:"brand_mentioned"`
	Citations        []AIAnalysisCitation `json:"citations"`
	ContentTruncated bool                 `json:"content_truncated,omitempty"`
	DomainCited      bool                 `json:"domain_cited"`
	Model            string               `json:"model"`
	ObservedAt       *string              `json:"observed_at"`
	Prompt           string               `json:"prompt"`
}

// AIAnalysisResult describes the API response.
type AIAnalysisResult struct {
	CostCents      float64         `json:"cost_cents"`
	CostStatus     string          `json:"cost_status"`
	Evidence       string          `json:"evidence"`
	Failure        *string         `json:"failure"`
	FetchedAt      string          `json:"fetched_at"`
	Rows           []AIAnalysisRow `json:"rows"`
	TotalAvailable *int            `json:"total_available"`
	Truncated      bool            `json:"truncated"`
}

// ListAgentReportsOptions selects a report kind and cursor page.
type ListAgentReportsOptions struct {
	Cursor string `json:"cursor,omitempty"`
	Kind   string `json:"kind,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// CreateAgentReportInput saves external analysis while preserving producer JSON keys.
type CreateAgentReportInput struct {
	Kind       string         `json:"kind"`
	Title      string         `json:"title"`
	Body       map[string]any `json:"body"`
	Provenance map[string]any `json:"provenance,omitempty"`
}

// ProjectContextInput replaces the project's guidance fields.
type ProjectContextInput struct {
	Business   string `json:"business"`
	Audience   string `json:"audience"`
	Products   string `json:"products"`
	Goals      string `json:"goals"`
	AgentRules string `json:"agent_rules"`
}

// RunSiteAuditOptions sets the bounded crawl size.
type RunSiteAuditOptions struct {
	MaxPages int `json:"max_pages,omitempty"`
}

// AIResearchInput requires an explicit provider cost cap in cents.
type AIResearchInput struct {
	Brand        string `json:"brand"`
	Domain       string `json:"domain"`
	MaxCostCents int    `json:"max_cost_cents"`
	EstimateOnly bool   `json:"estimate_only,omitempty"`
	Fresh        bool   `json:"fresh,omitempty"`
}

// AnalyzeAIVisibilityOptions selects an observed provider dataset.
type AnalyzeAIVisibilityOptions struct {
	AIResearchInput
	LanguageCode string `json:"language_code,omitempty"`
	Limit        int    `json:"limit,omitempty"`
	LocationCode int64  `json:"location_code,omitempty"`
	Platform     string `json:"platform,omitempty"`
	TargetType   string `json:"target_type,omitempty"`
}

// CompareAIPromptsOptions selects a synthetic prompt and bounded model list.
type CompareAIPromptsOptions struct {
	AIResearchInput
	Prompt string   `json:"prompt"`
	Models []string `json:"models,omitempty"`
}

// AIAnalysisOutcome distinguishes an estimate from a saved analysis report.
type AIAnalysisOutcome struct {
	OK                 bool              `json:"ok"`
	Estimate           bool              `json:"estimate"`
	EstimatedCostCents *float64          `json:"estimated_cost_cents,omitempty"`
	Evidence           string            `json:"evidence,omitempty"`
	Cached             *bool             `json:"cached,omitempty"`
	ReportID           *string           `json:"report_id,omitempty"`
	CostCents          *float64          `json:"cost_cents,omitempty"`
	Result             *AIAnalysisResult `json:"result,omitempty"`
}

func (outcome *AIAnalysisOutcome) UnmarshalJSON(data []byte) error {
	type wire AIAnalysisOutcome
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	var estimate *bool
	if err := json.Unmarshal(fields["estimate"], &estimate); err != nil || estimate == nil || !value.OK {
		return fmt.Errorf("AI analysis must include ok:true and a boolean estimate discriminator")
	}
	if value.Estimate {
		if value.EstimatedCostCents == nil || (value.Evidence != "observed_dataset" && value.Evidence != "synthetic_prompt_test") {
			return fmt.Errorf("AI estimate is incomplete")
		}
	} else if value.Cached == nil || value.ReportID == nil || value.CostCents == nil || value.Result == nil {
		return fmt.Errorf("AI analysis report is incomplete")
	}
	*outcome = AIAnalysisOutcome(value)
	return nil
}
