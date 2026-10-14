package bisibility

// RankHistoryReportNormalizedItem contains retained normalized values, never raw provider payloads.
type RankHistoryReportNormalizedItem struct {
	ID               string  `json:"id"`
	CheckID          string  `json:"check_id"`
	KeywordID        string  `json:"keyword_id"`
	CheckedAt        string  `json:"checked_at"`
	CheckRecordedAt  string  `json:"check_recorded_at"`
	ObservedAt       *string `json:"observed_at"`
	RecordedAt       string  `json:"recorded_at"`
	RequestedAt      *string `json:"requested_at"`
	SourceObservedAt *string `json:"source_observed_at"`
	BlockPosition    *int    `json:"block_position"`
	BusinessName     *string `json:"business_name"`
	CID              *string `json:"cid"`
	Domain           *string `json:"domain"`
	MapsURL          *string `json:"maps_url"`
	Ordinal          int     `json:"ordinal"`
	PlaceID          *string `json:"place_id"`
	PositionInBlock  *int    `json:"position_in_block"`
	RankAbsolute     *int    `json:"rank_absolute"`
	RankGroup        *int    `json:"rank_group"`
	RankScale        *string `json:"rank_scale"`
	Promotion        *string `json:"promotion"`
	SourcePage       *int    `json:"source_page"`
	SourceBlockID    *string `json:"source_block_id"`
	SourceOrdinal    *int    `json:"source_ordinal"`
	ResultKind       string  `json:"result_kind"`
	Title            *string `json:"title"`
	URL              *string `json:"url"`
}

type RankHistoryReportNormalizedItems struct {
	TotalCount   int    `json:"total_count"`
	Completeness string `json:"completeness"`
	IsTruncated  bool   `json:"is_truncated"`
	Capacity     int    `json:"capacity"`
	Paging       string `json:"paging"`
}
