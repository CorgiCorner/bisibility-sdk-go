package bisibility

// RankHistoryReportObservationConfiguredScope preserves the version 2 server report fields.
type RankHistoryReportObservationConfiguredScope struct {
	Device           string  `json:"device"`
	Engine           string  `json:"engine"`
	IdentityRevision *string `json:"identity_revision"`
	Language         string  `json:"language"`
	Location         string  `json:"location"`
	PolicyRevision   *string `json:"policy_revision"`
	RankScale        string  `json:"rank_scale"`
	ResultKind       string  `json:"result_kind"`
	Surface          string  `json:"surface"`
	TargetRevision   *string `json:"target_revision"`
}

// RankHistoryReportObservation preserves the version 2 server report fields.
type RankHistoryReportObservation struct {
	CheckedAt            string                                       `json:"checked_at"`
	CheckRecordedAt      string                                       `json:"check_recorded_at"`
	ObservedAt           *string                                      `json:"observed_at"`
	RecordedAt           string                                       `json:"recorded_at"`
	RequestedAt          *string                                      `json:"requested_at"`
	ComparisonDigest     *string                                      `json:"comparison_digest"`
	CompleteThrough      *int                                         `json:"complete_through"`
	Completeness         string                                       `json:"completeness"`
	ConfiguredScope      *RankHistoryReportObservationConfiguredScope `json:"configured_scope"`
	EffectiveScope       *RankHistoryReportObservationConfiguredScope `json:"effective_scope"`
	Evidence             string                                       `json:"evidence"`
	ExpectedURLAtCheck   *string                                      `json:"expected_url_at_check"`
	Freshness            string                                       `json:"freshness"`
	FreshnessAsOf        string                                       `json:"freshness_as_of"`
	FreshnessReason      *string                                      `json:"freshness_reason"`
	ID                   string                                       `json:"id"`
	KeywordID            string                                       `json:"keyword_id"`
	Keyword              string                                       `json:"keyword"`
	NormalizationVersion *string                                      `json:"normalization_version"`
	Position             *float64                                     `json:"position"`
	Provider             *string                                      `json:"provider"`
	RankingURL           *string                                      `json:"ranking_url"`
	Reason               *string                                      `json:"reason"`
	ReportedPosition     *float64                                     `json:"reported_position"`
	RequestedDepth       *int                                         `json:"requested_depth"`
	NormalizedItems      []RankHistoryReportNormalizedItem            `json:"normalized_items"`
	NormalizedItemCount  int                                          `json:"normalized_item_count"`
}

// RankHistoryReportAttempt preserves the version 2 server report fields.
type RankHistoryReportAttempt struct {
	CheckedAt                    string  `json:"checked_at"`
	CheckRecordedAt              string  `json:"check_recorded_at"`
	ObservedAt                   *string `json:"observed_at"`
	RecordedAt                   string  `json:"recorded_at"`
	RequestedAt                  *string `json:"requested_at"`
	DeferredReason               *string `json:"deferred_reason"`
	ErrorCode                    *string `json:"error_code"`
	FinishedAt                   *string `json:"finished_at"`
	ID                           string  `json:"id"`
	KeywordID                    string  `json:"keyword_id"`
	Keyword                      string  `json:"keyword"`
	Kind                         string  `json:"kind"`
	Provider                     *string `json:"provider"`
	ReportedProviderAttemptCount int     `json:"reported_provider_attempt_count"`
	StartedAt                    *string `json:"started_at"`
	Status                       string  `json:"status"`
}

// RankHistoryReportMetadataWindow preserves the version 2 server report fields.
type RankHistoryReportMetadataWindow struct {
	Basis string `json:"basis"`
	From  string `json:"from"`
	To    string `json:"to"`
}

// RankHistoryReportMetadataSummaryMetricCounts preserves the version 2 server report fields.
type RankHistoryReportMetadataSummaryMetricCounts struct {
	Expected         int  `json:"expected"`
	Measured         int  `json:"measured"`
	MetricEligible   int  `json:"metric_eligible"`
	Ranked           int  `json:"ranked"`
	KnownAbsent      int  `json:"known_absent"`
	Unknown          int  `json:"unknown"`
	NeverChecked     int  `json:"never_checked"`
	Failed           int  `json:"failed"`
	Stale            int  `json:"stale"`
	Unsupported      int  `json:"unsupported"`
	FreshDeterminate *int `json:"fresh_determinate"`
	UnknownFreshness int  `json:"unknown_freshness"`
	Comparable       int  `json:"comparable"`
}

// RankHistoryReportMetadataSummaryMetricContributionsItem preserves the version 2 server report fields.
type RankHistoryReportMetadataSummaryMetricContributionsItem struct {
	TargetID string   `json:"target_id"`
	Weight   float64  `json:"weight"`
	Value    *float64 `json:"value"`
	Delta    *float64 `json:"delta"`
}

// RankHistoryReportMetadataSummaryMetricExclusionsItem preserves the version 2 server report fields.
type RankHistoryReportMetadataSummaryMetricExclusionsItem struct {
	TargetID string `json:"target_id"`
	Reason   string `json:"reason"`
}

// RankHistoryReportMetadataSummaryMetricWeightProvenance preserves the version 2 server report fields.
type RankHistoryReportMetadataSummaryMetricWeightProvenance struct {
	Source *string `json:"source"`
	Period *string `json:"period"`
	Market *string `json:"market"`
}

// RankHistoryReportMetadataSummaryMetric preserves the version 2 server report fields.
type RankHistoryReportMetadataSummaryMetric struct {
	MetricID           string                                                    `json:"metric_id"`
	DefinitionVersion  string                                                    `json:"definition_version"`
	AsOf               string                                                    `json:"as_of"`
	Value              *float64                                                  `json:"value"`
	Delta              *float64                                                  `json:"delta"`
	Counts             RankHistoryReportMetadataSummaryMetricCounts              `json:"counts"`
	DataCoverage       *float64                                                  `json:"data_coverage"`
	ObservedCoverage   *float64                                                  `json:"observed_coverage"`
	Freshness          string                                                    `json:"freshness"`
	FreshnessReason    *string                                                   `json:"freshness_reason"`
	ImputedWeightShare *float64                                                  `json:"imputed_weight_share"`
	Contributions      []RankHistoryReportMetadataSummaryMetricContributionsItem `json:"contributions"`
	Exclusions         []RankHistoryReportMetadataSummaryMetricExclusionsItem    `json:"exclusions"`
	Added              *[]string                                                 `json:"added"`
	Removed            *[]string                                                 `json:"removed"`
	CompositionReason  *string                                                   `json:"composition_reason"`
	WeightProvenance   RankHistoryReportMetadataSummaryMetricWeightProvenance    `json:"weight_provenance"`
	IncludedCount      int                                                       `json:"included_count"`
	EligibleTotal      int                                                       `json:"eligible_total"`
	IsTruncated        bool                                                      `json:"is_truncated"`
}

// RankHistoryReportMetadataSummaryMovementCounts preserves the version 2 server report fields.
type RankHistoryReportMetadataSummaryMovementCounts struct {
	Improved  int `json:"improved"`
	Declined  int `json:"declined"`
	Entered   int `json:"entered"`
	Exited    int `json:"exited"`
	First     int `json:"first"`
	Moved     int `json:"moved"`
	ReEntered int `json:"re_entered"`
	Steady    int `json:"steady"`
	Unknown   int `json:"unknown"`
}

// RankHistoryReportMetadataSummaryRetainedHistory preserves the version 2 server report fields.
type RankHistoryReportMetadataSummaryRetainedHistory struct {
	Checks       int    `json:"checks"`
	Completeness string `json:"completeness"`
	Targets      int    `json:"targets"`
}

// RankHistoryReportMetadataSummaryTimeBasisBaseline preserves the version 2 server report fields.
type RankHistoryReportMetadataSummaryTimeBasisBaseline struct {
	Before    string  `json:"before"`
	NotBefore *string `json:"not_before"`
}

// RankHistoryReportMetadataSummaryTimeBasis preserves the version 2 server report fields.
type RankHistoryReportMetadataSummaryTimeBasis struct {
	AsOf             string                                             `json:"as_of"`
	Baseline         *RankHistoryReportMetadataSummaryTimeBasisBaseline `json:"baseline"`
	CurrentNotBefore *string                                            `json:"current_not_before"`
}

// RankHistoryReportMetadataSummary preserves the version 2 server report fields.
type RankHistoryReportMetadataSummary struct {
	Comparison      string                                          `json:"comparison"`
	Metric          RankHistoryReportMetadataSummaryMetric          `json:"metric"`
	MovementCounts  *RankHistoryReportMetadataSummaryMovementCounts `json:"movement_counts"`
	RetainedHistory RankHistoryReportMetadataSummaryRetainedHistory `json:"retained_history"`
	SelectionDigest string                                          `json:"selection_digest"`
	TimeBasis       RankHistoryReportMetadataSummaryTimeBasis       `json:"time_basis"`
}

// RankHistoryReportMetadataResolutionsItemBaseline preserves the version 2 server report fields.
type RankHistoryReportMetadataResolutionsItemBaseline struct {
	ComparisonDigest     *string  `json:"comparison_digest"`
	Evidence             string   `json:"evidence"`
	Freshness            string   `json:"freshness"`
	FreshnessReason      *string  `json:"freshness_reason"`
	ID                   string   `json:"id"`
	NormalizationVersion *string  `json:"normalization_version"`
	ObservedAt           *string  `json:"observed_at"`
	Position             *float64 `json:"position"`
	Provider             *string  `json:"provider"`
	RankingURL           *string  `json:"ranking_url"`
	Reason               *string  `json:"reason"`
	RecordedAt           string   `json:"recorded_at"`
	ReportedPosition     *float64 `json:"reported_position"`
	RequestedDepth       *int     `json:"requested_depth"`
}

// RankHistoryReportMetadataResolutionsItemComparisonBoundary preserves the version 2 server report fields.
type RankHistoryReportMetadataResolutionsItemComparisonBoundary struct {
	Point  *RankHistoryReportMetadataResolutionsItemBaseline `json:"point"`
	Reason string                                            `json:"reason"`
}

// RankHistoryReportMetadataResolutionsItemLatestAttempt preserves the version 2 server report fields.
type RankHistoryReportMetadataResolutionsItemLatestAttempt struct {
	ID                    string  `json:"id"`
	MeasurementRecordedAt string  `json:"measurement_recorded_at"`
	ObservedAt            *string `json:"observed_at"`
	RecordedAt            *string `json:"recorded_at"`
	Status                string  `json:"status"`
}

// RankHistoryReportMetadataResolutionsItemLastKnown preserves the version 2 server report fields.
type RankHistoryReportMetadataResolutionsItemLastKnown struct {
	Baseline                    *RankHistoryReportMetadataResolutionsItemBaseline           `json:"baseline"`
	Comparison                  string                                                      `json:"comparison"`
	ComparisonBoundary          *RankHistoryReportMetadataResolutionsItemComparisonBoundary `json:"comparison_boundary"`
	ComparisonWindow            *RankHistoryReportMetadataSummaryTimeBasisBaseline          `json:"comparison_window"`
	ComparisonUnavailableReason *string                                                     `json:"comparison_unavailable_reason"`
	Current                     *RankHistoryReportMetadataResolutionsItemBaseline           `json:"current"`
	Event                       *string                                                     `json:"event"`
	KeywordID                   string                                                      `json:"keyword_id"`
	LastSuccess                 *RankHistoryReportMetadataResolutionsItemBaseline           `json:"last_success"`
	LatestAttempt               *RankHistoryReportMetadataResolutionsItemLatestAttempt      `json:"latest_attempt"`
	Movement                    *float64                                                    `json:"movement"`
}

// RankHistoryReportMetadataResolutionsItem preserves the version 2 server report fields.
type RankHistoryReportMetadataResolutionsItem struct {
	Baseline                    *RankHistoryReportMetadataResolutionsItemBaseline           `json:"baseline"`
	Comparison                  string                                                      `json:"comparison"`
	ComparisonBoundary          *RankHistoryReportMetadataResolutionsItemComparisonBoundary `json:"comparison_boundary"`
	ComparisonWindow            *RankHistoryReportMetadataSummaryTimeBasisBaseline          `json:"comparison_window"`
	ComparisonUnavailableReason *string                                                     `json:"comparison_unavailable_reason"`
	Current                     *RankHistoryReportMetadataResolutionsItemBaseline           `json:"current"`
	Event                       *string                                                     `json:"event"`
	KeywordID                   string                                                      `json:"keyword_id"`
	LastSuccess                 *RankHistoryReportMetadataResolutionsItemBaseline           `json:"last_success"`
	LatestAttempt               *RankHistoryReportMetadataResolutionsItemLatestAttempt      `json:"latest_attempt"`
	Movement                    *float64                                                    `json:"movement"`
	CurrentScopeState           string                                                      `json:"current_scope_state"`
	LastKnown                   RankHistoryReportMetadataResolutionsItemLastKnown           `json:"last_known"`
}
