package bisibility

// PromptResearchSuggestionConfiguration records the reviewed model and market.
type PromptResearchSuggestionConfiguration struct {
	Provider               string  `json:"provider"`
	Engine                 string  `json:"engine"`
	Model                  string  `json:"model"`
	LanguageCode           string  `json:"language_code"`
	CountryISOCode         string  `json:"country_iso_code,omitempty"`
	MaxOutputTokens        int     `json:"max_output_tokens"`
	AdvisoryCostLimitCents float64 `json:"advisory_cost_limit_cents"`
}
type PromptResearchReviewedContext struct {
	Business   string `json:"business"`
	Audience   string `json:"audience"`
	Products   string `json:"products"`
	Goals      string `json:"goals"`
	AgentRules string `json:"agent_rules"`
}
type PromptResearchReviewedCompetitor struct {
	ID     string  `json:"id"`
	Label  *string `json:"label"`
	Domain string  `json:"domain"`
}
type PromptResearchSuggestionSnapshot struct {
	Context     PromptResearchReviewedContext      `json:"context"`
	Competitors []PromptResearchReviewedCompetitor `json:"competitors"`
}
type PromptResearchSuggestionsPreviewInput struct {
	Configuration          PromptResearchSuggestionConfiguration `json:"configuration"`
	InputSnapshot          PromptResearchSuggestionSnapshot      `json:"input_snapshot"`
	CredentialConnectionID string                                `json:"credential_connection_id,omitempty"`
}
type PromptResearchSuggestionsPreview struct {
	PromptResearchSuggestionsPreviewInput
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
type PromptResearchSuggestionsGenerateInput struct {
	Preview PromptResearchSuggestionsPreview `json:"preview"`
	Consent bool                             `json:"consent"`
}
type PromptResearchModelSuggestion struct {
	DraftID     string   `json:"draft_id"`
	Text        string   `json:"text"`
	Category    string   `json:"category"`
	Provenance  string   `json:"provenance"`
	EvidenceIDs []string `json:"evidence_ids"`
	Popularity  *float64 `json:"popularity"`
	Accepted    bool     `json:"accepted"`
}
type PromptResearchSuggestionsGeneration struct {
	GenerationID string                          `json:"generation_id"`
	Drafts       []PromptResearchModelSuggestion `json:"drafts"`
	CostUSD      *string                         `json:"cost_usd"`
	CostState    string                          `json:"cost_state"`
	Method       string                          `json:"method"`
	Limitations  []string                        `json:"limitations"`
}
type PromptTrackerGenerationReference struct {
	GenerationID string `json:"generation_id"`
	DraftID      string `json:"draft_id"`
}
type PromptTrackerProviderDatasetReference struct {
	ReportID string `json:"report_id"`
	RowIndex int    `json:"row_index"`
}
type PromptResearchSuggestion struct {
	GenerationReference      *PromptTrackerGenerationReference      `json:"generation_reference,omitempty"`
	ProviderDatasetReference *PromptTrackerProviderDatasetReference `json:"provider_dataset_reference,omitempty"`
	Text                     string                                 `json:"text"`
	Category                 string                                 `json:"category"`
	Provenance               string                                 `json:"provenance,omitempty"`
	EvidenceIDs              []string                               `json:"evidence_ids,omitempty"`
	Popularity               *float64                               `json:"popularity"`
	Accepted                 bool                                   `json:"accepted,omitempty"`
}
type PromptResearchSuggestions struct {
	InputSnapshot      PromptResearchSuggestionSnapshot `json:"input_snapshot"`
	ContextUpdatedAt   *string                          `json:"context_updated_at"`
	Drafts             []PromptResearchSuggestion       `json:"drafts"`
	Method             string                           `json:"method"`
	RequiresAcceptance bool                             `json:"requires_acceptance"`
	CostUSD            string                           `json:"cost_usd"`
	Limitations        []string                         `json:"limitations"`
}
type PromptResearchAcceptanceInput struct {
	Drafts []PromptResearchSuggestion `json:"drafts"`
}
type PromptResearchAcceptance struct {
	Prompts []PromptTrackerPrompt `json:"prompts"`
}
