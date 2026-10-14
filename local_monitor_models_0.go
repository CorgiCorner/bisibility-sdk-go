package bisibility

// Fields derived from the actual Local retained-read OpenAPI source.

type LocalMonitorRunResultResponseDataProtocol struct {
	Geometry           string   `json:"geometry"`
	Matcher            string   `json:"matcher"`
	NormalizerVersions []string `json:"normalizer_versions"`
	MetricDefinition   string   `json:"metric_definition"`
}

type LocalMonitorRunResultResponseDataFreshnessPolicy struct {
	Basis                 *string  `json:"basis"`
	FreshForMS            *float64 `json:"fresh_for_ms"`
	FreshnessReason       *string  `json:"freshness_reason"`
	PolicyEpoch           *string  `json:"policy_epoch"`
	ConfigurationRevision *string  `json:"configuration_revision"`
	Semantics             string   `json:"semantics"`
}

type LocalMonitorRunResultResponseDataPointsItemReceipt struct {
	TargetID               string  `json:"target_id"`
	Status                 string  `json:"status"`
	Version                int     `json:"version"`
	Acceptance             *string `json:"acceptance"`
	KnownAcceptedTaskCount *int    `json:"known_accepted_task_count"`
	Collected              bool    `json:"collected"`
	CancellationRequested  bool    `json:"cancellation_requested"`
	FinancialState         *string `json:"financial_state"`
	OwnerRetained          bool    `json:"owner_retained"`
	Resolution             string  `json:"resolution"`
	AccountingPending      *bool   `json:"accounting_pending"`
	ReconciliationReason   *string `json:"reconciliation_reason"`
}

type LocalMonitorRunResultResponseDataPointsItem struct {
	ID               string                                             `json:"id"`
	KeywordID        string                                             `json:"keyword_id"`
	Keyword          string                                             `json:"keyword"`
	Ordinal          int                                                `json:"ordinal"`
	Row              int                                                `json:"row"`
	Column           int                                                `json:"column"`
	Latitude         string                                             `json:"latitude"`
	Longitude        string                                             `json:"longitude"`
	Surface          string                                             `json:"surface"`
	Cutoff           int                                                `json:"cutoff"`
	Zoom             *float64                                           `json:"zoom"`
	Language         string                                             `json:"language"`
	Device           string                                             `json:"device"`
	OperatingSystem  *string                                            `json:"operating_system"`
	ObservedAt       *string                                            `json:"observed_at"`
	RecordedAt       *string                                            `json:"recorded_at"`
	Position         *float64                                           `json:"position"`
	ObservedPosition *float64                                           `json:"observed_position"`
	Evidence         string                                             `json:"evidence"`
	Reason           *string                                            `json:"reason"`
	Freshness        string                                             `json:"freshness"`
	FreshnessReason  *string                                            `json:"freshness_reason"`
	IdentityMatch    string                                             `json:"identity_match"`
	ScopeClass       string                                             `json:"scope_class"`
	Receipt          LocalMonitorRunResultResponseDataPointsItemReceipt `json:"receipt"`
}

type LocalMonitorRunResultResponseDataGridsItemCounts struct {
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

type LocalMonitorRunResultResponseDataGridsItem struct {
	KeywordID          string                                           `json:"keyword_id"`
	Keyword            string                                           `json:"keyword"`
	GridRevisionKey    string                                           `json:"grid_revision_key"`
	IdentityRevisionID string                                           `json:"identity_revision_id"`
	TopN               int                                              `json:"top_n"`
	Size               int                                              `json:"size"`
	PointIDs           []string                                         `json:"point_ids"`
	Expected           int                                              `json:"expected"`
	ObservedRate       *float64                                         `json:"observed_rate"`
	ObservedCoverage   *float64                                         `json:"observed_coverage"`
	DataCoverage       *float64                                         `json:"data_coverage"`
	SampleBounds       *[2]float64                                      `json:"sample_bounds"`
	Counts             LocalMonitorRunResultResponseDataGridsItemCounts `json:"counts"`
	Freshness          string                                           `json:"freshness"`
	Comparison         string                                           `json:"comparison"`
	Scope              string                                           `json:"scope"`
}

type LocalMonitorRunResultResponseDataReceiptCounts struct {
	Items                int `json:"items"`
	Complete             int `json:"complete"`
	AcceptanceUnresolved int `json:"acceptance_unresolved"`
	AccountingPending    int `json:"accounting_pending"`
	Unavailable          int `json:"unavailable"`
	RetainedOwners       int `json:"retained_owners"`
}

type LocalMonitorRunResultResponseDataReceipt struct {
	ID                    string                                               `json:"id"`
	Version               int                                                  `json:"version"`
	CancellationRequested bool                                                 `json:"cancellation_requested"`
	ExpectedCount         int                                                  `json:"expected_count"`
	Items                 []LocalMonitorRunResultResponseDataPointsItemReceipt `json:"items"`
	Counts                LocalMonitorRunResultResponseDataReceiptCounts       `json:"counts"`
}

type LocalMonitorRunResultResponseData struct {
	Version              int                                               `json:"version"`
	ProjectID            string                                            `json:"project_id"`
	MonitorID            string                                            `json:"monitor_id"`
	RunID                string                                            `json:"run_id"`
	AsOf                 string                                            `json:"as_of"`
	DefinitionRevisionID string                                            `json:"definition_revision_id"`
	IdentityRevisionID   string                                            `json:"identity_revision_id"`
	Subject              string                                            `json:"subject"`
	Protocol             LocalMonitorRunResultResponseDataProtocol         `json:"protocol"`
	FinancialBasis       string                                            `json:"financial_basis"`
	FreshnessPolicy      *LocalMonitorRunResultResponseDataFreshnessPolicy `json:"freshness_policy"`
	Points               []LocalMonitorRunResultResponseDataPointsItem     `json:"points"`
	Grids                []LocalMonitorRunResultResponseDataGridsItem      `json:"grids"`
	CreatedAt            string                                            `json:"created_at"`
	DefinitionRevision   int                                               `json:"definition_revision"`
	Receipt              LocalMonitorRunResultResponseDataReceipt          `json:"receipt"`
}

type LocalMonitorRunResultResponse struct {
	Data LocalMonitorRunResultResponseData `json:"data"`
}

type LocalMonitorRunHistoryResponseDataWindow struct {
	From      *string `json:"from"`
	To        string  `json:"to"`
	TimeBasis string  `json:"time_basis"`
}

type LocalMonitorRunHistoryResponseDataItemsItem struct {
	RunID              string `json:"run_id"`
	CreatedAt          string `json:"created_at"`
	DefinitionRevision int    `json:"definition_revision"`
	Mode               string `json:"mode"`
}

type LocalMonitorRunHistoryResponseData struct {
	Version       int                                           `json:"version"`
	ProjectID     string                                        `json:"project_id"`
	MonitorID     string                                        `json:"monitor_id"`
	AsOf          string                                        `json:"as_of"`
	Window        *LocalMonitorRunHistoryResponseDataWindow     `json:"window"`
	EligibleTotal *int                                          `json:"eligible_total"`
	NextCursor    *string                                       `json:"next_cursor"`
	Items         []LocalMonitorRunHistoryResponseDataItemsItem `json:"items"`
}

type LocalMonitorRunHistoryResponse struct {
	Data LocalMonitorRunHistoryResponseData `json:"data"`
}

type LocalMonitorPointHistoryResponseDataItemsItem struct {
	ID              string   `json:"id"`
	RunID           string   `json:"run_id"`
	CreatedAt       string   `json:"created_at"`
	ObservedAt      *string  `json:"observed_at"`
	Position        *float64 `json:"position"`
	Evidence        string   `json:"evidence"`
	Reason          *string  `json:"reason"`
	Freshness       *string  `json:"freshness"`
	FreshnessReason *string  `json:"freshness_reason"`
}

type LocalMonitorPointHistoryResponseData struct {
	Version                        int                                             `json:"version"`
	AsOf                           string                                          `json:"as_of"`
	TargetID                       string                                          `json:"target_id"`
	Scope                          string                                          `json:"scope"`
	Limit                          int                                             `json:"limit"`
	HasMore                        bool                                            `json:"has_more"`
	Window                         *LocalMonitorRunHistoryResponseDataWindow       `json:"window"`
	EligibleTotal                  *int                                            `json:"eligible_total"`
	NextCursor                     *string                                         `json:"next_cursor"`
	PolicyEpoch                    *string                                         `json:"policy_epoch"`
	FreshnessConfigurationRevision *string                                         `json:"freshness_configuration_revision"`
	Items                          []LocalMonitorPointHistoryResponseDataItemsItem `json:"items"`
}

type LocalMonitorPointHistoryResponse struct {
	Data LocalMonitorPointHistoryResponseData `json:"data"`
}
