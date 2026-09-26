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
