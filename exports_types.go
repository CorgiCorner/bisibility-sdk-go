package bisibility

import (
	"encoding/json"
	"fmt"
	"time"
)

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
	Version     string
	Stream      string
	Basis       string
	From        string
	To          string
	Aggregation string
	Cursor      string
	Format      RankHistoryExportFormat
	Granularity RankHistoryGranularity
	KeywordIDs  []string
	Limit       int
	Range       RankHistoryExportRange
}

// RankHistoryExportRow is one exported rank-check result.
type RankHistoryExportRow struct {
	// Report is populated only for opt-in version 2. The legacy fields remain available.
	Report               *RankHistoryReportRecord `json:"-"`
	Provider             *string                  `json:"provider"`
	RequestedDepth       *int                     `json:"requested_depth"`
	NormalizationVersion *string                  `json:"normalization_version"`
	CheckedAt            time.Time                `json:"checked_at"`
	ID                   string                   `json:"id"`
	Keyword              string                   `json:"keyword"`
	KeywordID            string                   `json:"keyword_id"`
	Position             *int                     `json:"position"`
	PreviousPosition     *int                     `json:"previous_position"`
	RankingURL           *string                  `json:"ranking_url"`
}

// RankHistoryExportResponse contains either a JSON page or a complete CSV document.
// Format identifies which representation is populated.
type RankHistoryExportResponse struct {
	CSV    string                  `json:"-"`
	Data   []RankHistoryExportRow  `json:"data"`
	Format RankHistoryExportFormat `json:"-"`
	Meta   ListMeta                `json:"meta"`
}

func (r *RankHistoryExportRow) UnmarshalJSON(data []byte) error {
	type legacy RankHistoryExportRow
	var value legacy
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	var shape map[string]json.RawMessage
	if err := json.Unmarshal(data, &shape); err != nil {
		return err
	}
	if _, observation := shape["evidence"]; observation || shape["kind"] != nil {
		var record RankHistoryReportRecord
		if err := json.Unmarshal(data, &record); err != nil {
			return err
		}
		value.Report = &record
	}
	*r = RankHistoryExportRow(value)
	return nil
}

func (r RankHistoryExportRow) MarshalJSON() ([]byte, error) {
	if r.Report != nil {
		return json.Marshal(r.Report)
	}
	type legacy RankHistoryExportRow
	return json.Marshal(legacy(r))
}

func (r *RankHistoryExportResponse) UnmarshalJSON(raw []byte) error {
	type response RankHistoryExportResponse
	var value response
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	if value.Meta.Report != nil && !validReportContract(raw, "RankHistoryReportResponse") {
		return fmt.Errorf("invalid required rank history report fields")
	}
	*r = RankHistoryExportResponse(value)
	return nil
}
