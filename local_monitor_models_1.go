package bisibility

// Fields derived from the actual Local retained-read OpenAPI source.

type LocalMonitorPointListingsResponseDataItem struct {
	ID            string  `json:"id"`
	Name          *string `json:"name"`
	RankGroup     *int    `json:"rank_group"`
	RankAbsolute  *int    `json:"rank_absolute"`
	Promotion     string  `json:"promotion"`
	ObservedAt    *string `json:"observed_at"`
	Matched       string  `json:"matched"`
	Cid           *string `json:"cid"`
	PlaceID       *string `json:"place_id"`
	SourcePage    *int    `json:"source_page"`
	SourceBlockID *string `json:"source_block_id"`
	SourceOrdinal *int    `json:"source_ordinal"`
	RankScale     *string `json:"rank_scale"`
}

type LocalMonitorPointListingsResponse struct {
	Data []LocalMonitorPointListingsResponseDataItem `json:"data"`
}

type LocalMonitorPointReportProtocol struct {
	Geometry           string   `json:"geometry"`
	Matcher            string   `json:"matcher"`
	NormalizerVersions []string `json:"normalizerVersions"`
	MetricDefinition   string   `json:"metricDefinition"`
}

type LocalMonitorPointReportFreshnessPolicy struct {
	Basis                 *string  `json:"basis"`
	FreshForMS            *float64 `json:"freshForMs"`
	FreshnessReason       *string  `json:"freshnessReason"`
	PolicyEpoch           *string  `json:"policyEpoch"`
	ConfigurationRevision *string  `json:"configurationRevision"`
	Semantics             string   `json:"semantics"`
}

type LocalMonitorPointReportPointsItemReceipt struct {
	TargetID               string  `json:"targetId"`
	Status                 string  `json:"status"`
	Version                int     `json:"version"`
	Acceptance             *string `json:"acceptance"`
	KnownAcceptedTaskCount *int    `json:"knownAcceptedTaskCount"`
	Collected              bool    `json:"collected"`
	CancellationRequested  bool    `json:"cancellationRequested"`
	FinancialState         *string `json:"financialState"`
	OwnerRetained          bool    `json:"ownerRetained"`
	Resolution             string  `json:"resolution"`
	AccountingPending      *bool   `json:"accountingPending"`
	ReconciliationReason   *string `json:"reconciliationReason"`
}

type LocalMonitorPointReportPointsItem struct {
	ID               string                                   `json:"id"`
	KeywordID        string                                   `json:"keywordId"`
	Keyword          string                                   `json:"keyword"`
	Ordinal          int                                      `json:"ordinal"`
	Row              int                                      `json:"row"`
	Column           int                                      `json:"column"`
	Latitude         string                                   `json:"latitude"`
	Longitude        string                                   `json:"longitude"`
	Surface          string                                   `json:"surface"`
	Cutoff           int                                      `json:"cutoff"`
	Zoom             *float64                                 `json:"zoom"`
	Language         string                                   `json:"language"`
	Device           string                                   `json:"device"`
	OperatingSystem  *string                                  `json:"operatingSystem"`
	ObservedAt       *string                                  `json:"observedAt"`
	RecordedAt       *string                                  `json:"recordedAt"`
	Position         *float64                                 `json:"position"`
	ObservedPosition *float64                                 `json:"observedPosition"`
	Evidence         string                                   `json:"evidence"`
	Reason           *string                                  `json:"reason"`
	Freshness        string                                   `json:"freshness"`
	FreshnessReason  *string                                  `json:"freshnessReason"`
	IdentityMatch    string                                   `json:"identityMatch"`
	ScopeClass       string                                   `json:"scopeClass"`
	Receipt          LocalMonitorPointReportPointsItemReceipt `json:"receipt"`
}

type LocalMonitorPointReportGridsItemCounts struct {
	Expected         int  `json:"expected"`
	Measured         int  `json:"measured"`
	MetricEligible   int  `json:"metricEligible"`
	Ranked           int  `json:"ranked"`
	KnownAbsent      int  `json:"knownAbsent"`
	Unknown          int  `json:"unknown"`
	NeverChecked     int  `json:"neverChecked"`
	Failed           int  `json:"failed"`
	Stale            int  `json:"stale"`
	Unsupported      int  `json:"unsupported"`
	FreshDeterminate *int `json:"freshDeterminate"`
	UnknownFreshness int  `json:"unknownFreshness"`
	Comparable       int  `json:"comparable"`
}

type LocalMonitorPointReportGridsItem struct {
	KeywordID          string                                 `json:"keywordId"`
	Keyword            string                                 `json:"keyword"`
	GridRevisionKey    string                                 `json:"gridRevisionKey"`
	IdentityRevisionID string                                 `json:"identityRevisionId"`
	TopN               int                                    `json:"topN"`
	Size               int                                    `json:"size"`
	PointIDs           []string                               `json:"pointIds"`
	Expected           int                                    `json:"expected"`
	ObservedRate       *float64                               `json:"observedRate"`
	ObservedCoverage   *float64                               `json:"observedCoverage"`
	DataCoverage       *float64                               `json:"dataCoverage"`
	SampleBounds       *[2]float64                            `json:"sampleBounds"`
	Counts             LocalMonitorPointReportGridsItemCounts `json:"counts"`
	Freshness          string                                 `json:"freshness"`
	Comparison         string                                 `json:"comparison"`
	Scope              string                                 `json:"scope"`
}

type LocalMonitorPointReport struct {
	Version              int                                     `json:"version"`
	ProjectID            string                                  `json:"projectId"`
	MonitorID            string                                  `json:"monitorId"`
	RunID                string                                  `json:"runId"`
	AsOf                 string                                  `json:"asOf"`
	DefinitionRevisionID string                                  `json:"definitionRevisionId"`
	IdentityRevisionID   string                                  `json:"identityRevisionId"`
	Subject              string                                  `json:"subject"`
	Protocol             LocalMonitorPointReportProtocol         `json:"protocol"`
	FinancialBasis       string                                  `json:"financialBasis"`
	FreshnessPolicy      *LocalMonitorPointReportFreshnessPolicy `json:"freshnessPolicy"`
	Points               []LocalMonitorPointReportPointsItem     `json:"points"`
	Grids                []LocalMonitorPointReportGridsItem      `json:"grids"`
	Report               string                                  `json:"report"`
	Scope                string                                  `json:"scope"`
	CoordinateProof      string                                  `json:"coordinateProof"`
	Comparison           string                                  `json:"comparison"`
}
