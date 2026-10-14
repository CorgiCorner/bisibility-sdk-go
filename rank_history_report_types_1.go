package bisibility

// RankHistoryReportMetadata preserves the version 2 server report fields.
type RankHistoryReportMetadata struct {
	Version                   int                                        `json:"version"`
	Aggregation               string                                     `json:"aggregation"`
	AsOf                      string                                     `json:"as_of"`
	Window                    RankHistoryReportMetadataWindow            `json:"window"`
	SummaryBasis              string                                     `json:"summary_basis"`
	DefinitionVersion         string                                     `json:"definition_version"`
	MetricID                  string                                     `json:"metric_id"`
	SelectionDigest           string                                     `json:"selection_digest"`
	TotalCount                int                                        `json:"total_count"`
	ExcludedUnknownTime       int                                        `json:"excluded_unknown_time"`
	UnresolvedMembershipCount int                                        `json:"unresolved_membership_count"`
	NormalizedItems           RankHistoryReportNormalizedItems           `json:"normalized_items"`
	Membership                string                                     `json:"membership"`
	Summary                   RankHistoryReportMetadataSummary           `json:"summary"`
	Resolutions               []RankHistoryReportMetadataResolutionsItem `json:"resolutions"`
}
