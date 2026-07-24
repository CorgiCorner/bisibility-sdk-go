package bisibility

import "time"

// RankHistoryExportFormat selects the rank-history response representation.
type RankHistoryExportFormat string

const (
	RankHistoryExportFormatJSON RankHistoryExportFormat = "json"
	RankHistoryExportFormatCSV  RankHistoryExportFormat = "csv"
)

// RankHistoryExportRange selects the history window.
type RankHistoryExportRange string

const (
	RankHistoryExportRange30Days RankHistoryExportRange = "30"
	RankHistoryExportRange90Days RankHistoryExportRange = "90"
	RankHistoryExportRangeAll    RankHistoryExportRange = "all"
)

// RankHistoryGranularity controls export aggregation.
type RankHistoryGranularity string

const (
	RankHistoryGranularityDaily  RankHistoryGranularity = "daily"
	RankHistoryGranularityWeekly RankHistoryGranularity = "weekly"
)

// ExportRankHistoryOptions filters project rank history.
type ExportRankHistoryOptions struct {
	Cursor      string
	Format      RankHistoryExportFormat
	Granularity RankHistoryGranularity
	KeywordIDs  []string
	Limit       int
	Range       RankHistoryExportRange
}

// RankHistoryExportRow is one exported rank-check result.
type RankHistoryExportRow struct {
	CheckedAt        time.Time `json:"checked_at"`
	ID               string    `json:"id"`
	Keyword          string    `json:"keyword"`
	KeywordID        string    `json:"keyword_id"`
	Position         *int      `json:"position"`
	PreviousPosition *int      `json:"previous_position"`
	RankingURL       *string   `json:"ranking_url"`
}

// RankHistoryExportResponse contains either a JSON page or a complete CSV document.
// Format identifies which representation is populated.
type RankHistoryExportResponse struct {
	CSV    string                  `json:"-"`
	Data   []RankHistoryExportRow  `json:"data"`
	Format RankHistoryExportFormat `json:"-"`
	Meta   ListMeta                `json:"meta"`
}
