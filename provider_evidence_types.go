package bisibility

// ResultCompleteness describes retained response evidence. Unknown future values
// remain readable. Partial or unresolved evidence cannot establish full-depth absence.
type ResultCompleteness string

const (
	ResultCompletenessComplete   ResultCompleteness = "complete"
	ResultCompletenessPartial    ResultCompleteness = "partial"
	ResultCompletenessUnresolved ResultCompleteness = "unresolved"
)

// ProviderEvidenceRecovery preserves nullable recovery facts and original RFC 3339 timestamps.
// State and reason are open strings supplied by the server.
type ProviderEvidenceRecovery struct {
	State         *string `json:"state"`
	LastAttemptAt *string `json:"last_attempt_at"`
	NextAttemptAt *string `json:"next_attempt_at"`
	Reason        *string `json:"reason"`
}

// ProviderEvidenceProjection contains retained provider facts. Nil fields mean
// unknown, including cost and returned depth. Status strings may evolve independently.
type ProviderEvidenceProjection struct {
	Acceptance        *string                   `json:"acceptance"`
	Fulfillment       *string                   `json:"fulfillment"`
	ReturnedDepth     *int                      `json:"returned_depth"`
	Accounting        *string                   `json:"accounting"`
	Result            *string                   `json:"result"`
	Cost              *string                   `json:"cost"`
	OriginalCheckedAt *string                   `json:"original_checked_at"`
	Recovery          *ProviderEvidenceRecovery `json:"recovery"`
}

// RankCheckRunEvidenceProjection associates retained evidence with its public check ID.
type RankCheckRunEvidenceProjection struct {
	ProviderEvidenceProjection
	CheckID string `json:"check_id"`
}
