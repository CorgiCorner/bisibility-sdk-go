package bisibility

import "time"

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

// BacklinksSnapshot is one cached, paid, or estimated backlink analysis result.
type BacklinksSnapshot struct {
	Cached             bool                    `json:"cached"`
	CachedUntil        time.Time               `json:"cached_until"`
	CostCents          float64                 `json:"cost_cents"`
	Estimate           *bool                   `json:"estimate,omitempty"`
	EstimatedCostCents *float64                `json:"estimated_cost_cents,omitempty"`
	FetchedAt          time.Time               `json:"fetched_at"`
	FetchedRowCount    int                     `json:"fetched_row_count"`
	History            []BacklinksHistoryMonth `json:"history"`
	IncludeSubdomains  bool                    `json:"include_subdomains"`
	Provider           string                  `json:"provider"`
	Rows               []BacklinkRow           `json:"rows"`
	Summary            BacklinksSummary        `json:"summary"`
	Target             string                  `json:"target"`
	TargetScope        BacklinkTargetScope     `json:"target_scope"`
	TotalRowsAvailable int                     `json:"total_rows_available"`
}

// BacklinksSnapshotResponse wraps a backlinks snapshot in the public API data envelope.
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
