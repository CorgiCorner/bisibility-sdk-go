package bisibility

// AITrackingSuggestionConfiguration records the reviewed model and market.
type AITrackingSuggestionConfiguration struct {
	Provider               string  `json:"provider"`
	Engine                 string  `json:"engine"`
	Model                  string  `json:"model"`
	LanguageCode           string  `json:"language_code"`
	CountryISOCode         string  `json:"country_iso_code,omitempty"`
	MaxOutputTokens        int     `json:"max_output_tokens"`
	AdvisoryCostLimitCents float64 `json:"advisory_cost_limit_cents"`
}
type AITrackingReviewedContext struct {
	Business   string `json:"business"`
	Audience   string `json:"audience"`
	Products   string `json:"products"`
	Goals      string `json:"goals"`
	AgentRules string `json:"agent_rules"`
}
type AITrackingReviewedCompetitor struct {
	ID     string  `json:"id"`
	Label  *string `json:"label"`
	Domain string  `json:"domain"`
}
type AITrackingSuggestionSnapshot struct {
	Context     AITrackingReviewedContext      `json:"context"`
	Competitors []AITrackingReviewedCompetitor `json:"competitors"`
}
type AITrackingSuggestionsPreviewInput struct {
	Configuration          AITrackingSuggestionConfiguration `json:"configuration"`
	InputSnapshot          AITrackingSuggestionSnapshot      `json:"input_snapshot"`
	CredentialConnectionID string                            `json:"credential_connection_id,omitempty"`
}
type AITrackingSuggestionsPreview struct {
	AITrackingSuggestionsPreviewInput
	Version             int      `json:"version"`
	SnapshotHash        string   `json:"snapshot_hash"`
	EstimatedCostCents  float64  `json:"estimated_cost_cents"`
	EstimateKind        string   `json:"estimate_kind"`
	IsGuaranteedMaximum bool     `json:"is_guaranteed_maximum"`
	CredentialVersion   string   `json:"credential_version"`
	BudgetRevision      string   `json:"budget_revision"`
	ConsentRevision     string   `json:"consent_revision"`
	ExpiresAt           string   `json:"expires_at"`
	Limitations         []string `json:"limitations"`
}
type AITrackingSuggestionsGenerateInput struct {
	Preview AITrackingSuggestionsPreview `json:"preview"`
	Consent bool                         `json:"consent"`
}
type AITrackingModelSuggestion struct {
	DraftID     string   `json:"draft_id"`
	Text        string   `json:"text"`
	Category    string   `json:"category"`
	Provenance  string   `json:"provenance"`
	EvidenceIDs []string `json:"evidence_ids"`
	Popularity  *float64 `json:"popularity"`
	Accepted    bool     `json:"accepted"`
}
type AITrackingSuggestionsGeneration struct {
	GenerationID string                      `json:"generation_id"`
	Drafts       []AITrackingModelSuggestion `json:"drafts"`
	CostUSD      *string                     `json:"cost_usd"`
	CostState    string                      `json:"cost_state"`
	Method       string                      `json:"method"`
	Limitations  []string                    `json:"limitations"`
}
type AITrackingGenerationReference struct {
	GenerationID string `json:"generation_id"`
	DraftID      string `json:"draft_id"`
}
type AITrackingProviderDatasetReference struct {
	ReportID string `json:"report_id"`
	RowIndex int    `json:"row_index"`
}
