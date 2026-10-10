package bisibility

import (
	"encoding/json"
	"time"
)

// BacklinkStatus describes whether a backlink is active, newly discovered, or lost.
type BacklinkStatus string

const (
	BacklinkStatusActive BacklinkStatus = "active"
	BacklinkStatusNew    BacklinkStatus = "new"
	BacklinkStatusLost   BacklinkStatus = "lost"
)

// BacklinkTargetScope selects a whole site or one page as the backlink target.
type BacklinkTargetScope string

const (
	BacklinkTargetScopeSite BacklinkTargetScope = "site"
	BacklinkTargetScopePage BacklinkTargetScope = "page"
)

// BacklinkMode controls provider-side row grouping over the full backlink corpus.
type BacklinkMode string

const (
	BacklinkModeAsIs         BacklinkMode = "as_is"
	BacklinkModeOnePerDomain BacklinkMode = "one_per_domain"
)

// BacklinksSummary contains provider totals for a backlinks snapshot.
type BacklinksSummary struct {
	BacklinksTotal        int     `json:"backlinks_total"`
	BrokenBacklinks       int     `json:"broken_backlinks"`
	BrokenPages           int     `json:"broken_pages"`
	DofollowPct           float64 `json:"dofollow_pct"`
	DomainRank            int     `json:"domain_rank"`
	LostBacklinks         int     `json:"lost_backlinks"`
	LostReferringDomains  int     `json:"lost_referring_domains"`
	NewBacklinks          int     `json:"new_backlinks"`
	NewReferringDomains   int     `json:"new_referring_domains"`
	ReferringDomainsTotal int     `json:"referring_domains_total"`
	ReferringPages        int     `json:"referring_pages"`
	SpamScore             float64 `json:"spam_score"`
}

// BacklinksHistoryMonth contains backlink gains and losses for one calendar month.
type BacklinksHistoryMonth struct {
	LostLinks int    `json:"lost_links"`
	Month     string `json:"month"`
	NewLinks  int    `json:"new_links"`
}

// BacklinkRow describes one backlink returned in a snapshot.
type BacklinkRow struct {
	Anchor          string         `json:"anchor"`
	DomainAuthority int            `json:"domain_authority"`
	FirstSeen       string         `json:"first_seen"`
	Flags           []string       `json:"flags"`
	LinksCount      int            `json:"links_count"`
	LostAt          *string        `json:"lost_at"`
	SourceDomain    string         `json:"source_domain"`
	SourceURL       string         `json:"source_url"`
	SpamScore       float64        `json:"spam_score"`
	Status          BacklinkStatus `json:"status"`
	TargetURL       string         `json:"target_url"`
}

// BacklinksEstimate is the free, cost-only dry run returned for EstimateOnly requests.
// It never carries summary, history, or row data, so an estimate can never be mistaken
// for an empty backlink profile.
type BacklinksEstimate struct {
	Cached             bool                `json:"cached"`
	CachedUntil        *time.Time          `json:"cached_until"`
	CostCents          float64             `json:"cost_cents"`
	Estimate           bool                `json:"estimate"`
	EstimatedCostCents float64             `json:"estimated_cost_cents"`
	IncludeSubdomains  bool                `json:"include_subdomains"`
	Provider           string              `json:"provider"`
	Target             string              `json:"target"`
	TargetScope        BacklinkTargetScope `json:"target_scope"`
}

// BacklinksSnapshot is one cached or paid backlink analysis result.
type BacklinksSnapshot struct {
	Cached          bool                    `json:"cached"`
	CachedUntil     time.Time               `json:"cached_until"`
	CostCents       float64                 `json:"cost_cents"`
	FetchedAt       time.Time               `json:"fetched_at"`
	FetchedRowCount int                     `json:"fetched_row_count"`
	History         []BacklinksHistoryMonth `json:"history"`
	// True only with empty history after an optional history failure.
	HistoryUnavailable bool                `json:"history_unavailable"`
	IncludeSubdomains  bool                `json:"include_subdomains"`
	Provider           string              `json:"provider"`
	Rows               []BacklinkRow       `json:"rows"`
	Summary            BacklinksSummary    `json:"summary"`
	Target             string              `json:"target"`
	TargetScope        BacklinkTargetScope `json:"target_scope"`
	TotalRowsAvailable int                 `json:"total_rows_available"`
}

// BacklinksResult is the discriminated estimate/snapshot response data returned by
// AnalyzeBacklinks. Exactly one pointer is set after successful decoding.
type BacklinksResult struct {
	Estimate *BacklinksEstimate
	Snapshot *BacklinksSnapshot
}

// UnmarshalJSON decodes an AnalyzeBacklinks result using estimate=true as its discriminator.
func (result *BacklinksResult) UnmarshalJSON(data []byte) error {
	var discriminator struct {
		Estimate bool `json:"estimate"`
	}
	if err := json.Unmarshal(data, &discriminator); err != nil {
		return err
	}
	if discriminator.Estimate {
		var estimate BacklinksEstimate
		if err := json.Unmarshal(data, &estimate); err != nil {
			return err
		}
		result.Estimate = &estimate
		result.Snapshot = nil
		return nil
	}
	var snapshot BacklinksSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return err
	}
	result.Estimate = nil
	result.Snapshot = &snapshot
	return nil
}

// BacklinksResponse wraps an estimate or a snapshot in the public API data envelope.
type BacklinksResponse = DataResponse[BacklinksResult]

// BacklinksSnapshotResponse wraps a backlinks snapshot in the public API data envelope.
// LoadMoreBacklinkRows always returns a snapshot, never an estimate.
type BacklinksSnapshotResponse = DataResponse[BacklinksSnapshot]

// AnalyzeBacklinksOptions controls a paid or estimated backlink analysis.
type AnalyzeBacklinksOptions struct {
	Target            string
	TargetScope       BacklinkTargetScope
	IncludeSubdomains bool
	ResultLimit       int
	Mode              BacklinkMode
	EstimateOnly      bool
	Fresh             bool
	MaxCostCents      int
}

// LoadMoreBacklinkRowsOptions selects rows to append to an unexpired snapshot.
type LoadMoreBacklinkRowsOptions struct {
	Target            string              `json:"target"`
	TargetScope       BacklinkTargetScope `json:"target_scope"`
	IncludeSubdomains bool                `json:"include_subdomains"`
	Limit             int                 `json:"limit"`
}
