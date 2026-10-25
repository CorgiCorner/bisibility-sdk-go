package bisibility

import "encoding/json"

// PromptTrackerSourceConfiguration keeps observation engine and mechanism distinct.
type PromptTrackerSourceConfiguration struct {
	Provider   string         `json:"provider"`
	Endpoint   string         `json:"endpoint"`
	Engine     string         `json:"engine"`
	Source     string         `json:"source"`
	Model      *string        `json:"model"`
	Parameters map[string]any `json:"parameters"`
}
type PromptTrackerTopicInput struct {
	Paused      *bool   `json:"paused,omitempty"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}
type PromptTrackerTopic struct {
	PausedAt    *string `json:"paused_at"`
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	ArchivedAt  *string `json:"archived_at"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}
type PromptTrackerPromptInput struct {
	GenerationReference      *PromptTrackerGenerationReference      `json:"generation_reference,omitempty"`
	ProviderDatasetReference *PromptTrackerProviderDatasetReference `json:"provider_dataset_reference,omitempty"`
	Category                 string                                 `json:"category,omitempty"`
	Paused                   *bool                                  `json:"paused,omitempty"`
	Text                     string                                 `json:"text"`
	TopicID                  *string                                `json:"topic_id,omitempty"`
	Label                    *string                                `json:"label,omitempty"`
}
type PromptTrackerPrompt struct {
	PausedAt   *string                       `json:"paused_at"`
	Revisions  []PromptTrackerPromptRevision `json:"revisions"`
	ID         string                        `json:"id"`
	TopicID    *string                       `json:"topic_id"`
	Label      *string                       `json:"label"`
	ArchivedAt *string                       `json:"archived_at"`
	CreatedAt  string                        `json:"created_at"`
	UpdatedAt  string                        `json:"updated_at"`
}
type PromptTrackerPromptRevision struct {
	GenerationReference      *PromptTrackerGenerationReference      `json:"generation_reference,omitempty"`
	ProviderDatasetReference *PromptTrackerProviderDatasetReference `json:"provider_dataset_reference,omitempty"`
	ID                       string                                 `json:"id"`
	Ordinal                  int                                    `json:"ordinal"`
	Category                 string                                 `json:"category"`
	Text                     string                                 `json:"text"`
	TextHash                 string                                 `json:"text_hash"`
	CreatedAt                string                                 `json:"created_at"`
}
type PromptTrackerTopicPatch struct {
	Paused      *bool          `json:"paused,omitempty"`
	Name        *string        `json:"name,omitempty"`
	Description NullableString `json:"description,omitempty"`
}

// MarshalJSON omits unset nullable fields so a bare patch does not clear
// description by accident. ClearString serializes as an explicit null.
func (p PromptTrackerTopicPatch) MarshalJSON() ([]byte, error) {
	fields := map[string]any{}
	if p.Paused != nil {
		fields["paused"] = p.Paused
	}
	if p.Name != nil {
		fields["name"] = p.Name
	}
	if p.Description.IsSet() {
		fields["description"] = p.Description
	}
	return json.Marshal(fields)
}

type PromptTrackerPromptPatch struct {
	GenerationReference      *PromptTrackerGenerationReference      `json:"generation_reference,omitempty"`
	ProviderDatasetReference *PromptTrackerProviderDatasetReference `json:"provider_dataset_reference,omitempty"`
	Text                     *string                                `json:"text,omitempty"`
	TopicID                  NullableString                         `json:"topic_id,omitempty"`
	Label                    NullableString                         `json:"label,omitempty"`
	Category                 *string                                `json:"category,omitempty"`
	Paused                   *bool                                  `json:"paused,omitempty"`
}

// MarshalJSON omits unset nullable fields so a bare patch does not clear
// topic_id or label by accident. ClearString serializes as an explicit null.
func (p PromptTrackerPromptPatch) MarshalJSON() ([]byte, error) {
	fields := map[string]any{}
	if p.GenerationReference != nil {
		fields["generation_reference"] = p.GenerationReference
	}
	if p.ProviderDatasetReference != nil {
		fields["provider_dataset_reference"] = p.ProviderDatasetReference
	}
	if p.Text != nil {
		fields["text"] = p.Text
	}
	if p.TopicID.IsSet() {
		fields["topic_id"] = p.TopicID
	}
	if p.Label.IsSet() {
		fields["label"] = p.Label
	}
	if p.Category != nil {
		fields["category"] = p.Category
	}
	if p.Paused != nil {
		fields["paused"] = p.Paused
	}
	return json.Marshal(fields)
}

type PromptTrackerPreviewInput struct {
	PromptIDs              []string                           `json:"prompt_ids"`
	Configurations         []PromptTrackerSourceConfiguration `json:"configurations"`
	CredentialConnectionID string                             `json:"credential_connection_id,omitempty"`
}
type PromptTrackerPreview struct {
	Configurations         []PromptTrackerSourceConfiguration `json:"configurations"`
	CredentialConnectionID string                             `json:"credential_connection_id"`
	CredentialVersion      string                             `json:"credential_version"`
	BudgetRevision         string                             `json:"budget_revision"`
	ConsentRevision        string                             `json:"consent_revision"`
	EstimatedCostCents     float64                            `json:"estimated_cost_cents"`
}

// PromptTrackerRunInput binds launch to reviewed credentials, budget, consent and deadline.
type PromptTrackerRunInput struct {
	PromptTrackerPreviewInput
	CredentialVersion string `json:"credential_version"`
	BudgetRevision    string `json:"budget_revision"`
	ConsentRevision   string `json:"consent_revision"`
	Deadline          string `json:"deadline"`
	Origin            string `json:"origin,omitempty"`
	EntrySource       string `json:"entry_source,omitempty"`
	Consent           bool   `json:"consent"`
	ScheduleID        string `json:"schedule_id,omitempty"`
	PlannedAt         string `json:"planned_at,omitempty"`
}
type PromptTrackerScheduleInput struct {
	Name          string                `json:"name"`
	Cron          string                `json:"cron"`
	Timezone      string                `json:"timezone"`
	Enabled       bool                  `json:"enabled"`
	Configuration PromptTrackerRunInput `json:"configuration"`
	NextRunAt     *string               `json:"next_run_at,omitempty"`
	// Consent is read from the top level of the request body by the API; a
	// create or an enable requires it, so it is distinct from the Consent
	// inside Configuration (which carries the run plan's own consent).
	Consent bool `json:"consent"`
}

// PromptTrackerSchedulePatch models a partial schedule update, matching the
// server's scheduleInputSchema.partial() plus a top-level consent flag. Fields
// left at their zero value are omitted from the payload so the server only
// mutates what the caller explicitly set. Consent is sent at the top level,
// not inside Configuration, because the API reads it from there to approve a
// reviewed configuration.
type PromptTrackerSchedulePatch struct {
	Name          *string                `json:"name,omitempty"`
	Cron          *string                `json:"cron,omitempty"`
	Timezone      *string                `json:"timezone,omitempty"`
	Enabled       *bool                  `json:"enabled,omitempty"`
	Configuration *PromptTrackerRunInput `json:"configuration,omitempty"`
	NextRunAt     NullableString         `json:"next_run_at,omitempty"`
	Consent       bool                   `json:"consent"`
}

// MarshalJSON omits unset partial fields and always emits consent at the top
// level of the body, matching the API's partial PATCH contract.
func (p PromptTrackerSchedulePatch) MarshalJSON() ([]byte, error) {
	fields := map[string]any{"consent": p.Consent}
	if p.Name != nil {
		fields["name"] = p.Name
	}
	if p.Cron != nil {
		fields["cron"] = p.Cron
	}
	if p.Timezone != nil {
		fields["timezone"] = p.Timezone
	}
	if p.Enabled != nil {
		fields["enabled"] = p.Enabled
	}
	if p.Configuration != nil {
		fields["configuration"] = p.Configuration
	}
	if p.NextRunAt.IsSet() {
		fields["next_run_at"] = p.NextRunAt
	}
	return json.Marshal(fields)
}

type PromptTrackerScheduleConfiguration struct {
	PromptIDs      []string                           `json:"prompt_ids"`
	Configurations []PromptTrackerSourceConfiguration `json:"configurations"`
}
type PromptTrackerSchedule struct {
	Name          string                             `json:"name"`
	Cron          string                             `json:"cron"`
	Timezone      string                             `json:"timezone"`
	Enabled       bool                               `json:"enabled"`
	Configuration PromptTrackerScheduleConfiguration `json:"configuration"`
	NextRunAt     *string                            `json:"next_run_at"`
	ID            string                             `json:"id"`
	ArchivedAt    *string                            `json:"archived_at"`
	CreatedAt     string                             `json:"created_at"`
	UpdatedAt     string                             `json:"updated_at"`
}
type PromptTrackerRun struct {
	UpdatedAt   string  `json:"updated_at"`
	SampleCount *int    `json:"sample_count"`
	ID          string  `json:"id"`
	State       string  `json:"state"`
	CreatedAt   string  `json:"created_at"`
	FinishedAt  *string `json:"finished_at"`
	PlannedAt   *string `json:"planned_at"`
}

// Nullable decimal USD values and unknown state are preserved without invented zeroes.
type PromptTrackerEvidence struct {
	AnswerText      *string                 `json:"answer_text"`
	Raw             any                     `json:"raw"`
	AnswerTruncated bool                    `json:"answer_truncated"`
	RawTruncated    bool                    `json:"raw_truncated"`
	SearchResults   []PromptTrackerCitation `json:"search_results"`
	RequestedLocale *string                 `json:"requested_locale"`
	EffectiveLocale *string                 `json:"effective_locale"`
	LocaleMechanism *string                 `json:"locale_mechanism"`
	RequestedModel  *string                 `json:"requested_model"`
	ActualModel     *string                 `json:"actual_model"`
	ProviderStatus  *string                 `json:"provider_status"`
	ObservedAt      *string                 `json:"observed_at"`
	FetchedAt       string                  `json:"fetched_at"`
	RecordedSource  string                  `json:"recorded_source"`
}
type PromptTrackerCitation struct {
	URL      string  `json:"url"`
	Title    *string `json:"title"`
	Position int     `json:"position"`
}
type PromptTrackerSample struct {
	ID               string                  `json:"id"`
	Measurement      string                  `json:"measurement"`
	Source           string                  `json:"source"`
	Engine           string                  `json:"engine"`
	Prompt           string                  `json:"prompt"`
	PromptRevisionID string                  `json:"prompt_revision_id"`
	Evidence         *PromptTrackerEvidence  `json:"evidence"`
	Citations        []PromptTrackerCitation `json:"citations"`
	CostUSD          *string                 `json:"cost_usd"`
	CostState        string                  `json:"cost_state"`
}
type PromptTrackerExport struct {
	Items      []PromptTrackerSample     `json:"items"`
	RunID      string                    `json:"run_id"`
	NextCursor *string                   `json:"next_cursor"`
	Scope      *PromptTrackerExportScope `json:"scope,omitempty"`
}
type PromptTrackerExportScope struct {
	Complete bool `json:"complete"`
	MaxPages int  `json:"max_pages"`
	Loaded   int  `json:"loaded"`
	Resumed  bool `json:"resumed"`
}
type PromptTrackerExportResponse struct {
	Data   *PromptTrackerExport `json:"data,omitempty"`
	CSV    string               `json:"-"`
	Format string               `json:"-"`
}
type PromptTrackerEvidenceOptions struct {
	Format string
	RunID  string
	Limit  int
	Cursor string
}
type PromptTrackerTrendOptions struct {
	PromptTrackerEvidenceOptions
	PreviousRunID string
}
type PromptTrackerDenominator struct {
	Expected    int      `json:"expected"`
	Observed    int      `json:"observed"`
	Eligible    int      `json:"eligible"`
	Mentioned   int      `json:"mentioned"`
	AbsentAIO   int      `json:"absent_aio"`
	Partial     int      `json:"partial"`
	Failed      int      `json:"failed"`
	Unknown     int      `json:"unknown"`
	Missing     int      `json:"missing"`
	Coverage    float64  `json:"coverage"`
	MentionRate *float64 `json:"mention_rate"`
}
type PromptTrackerTrends struct {
	Comparable         bool                      `json:"comparable"`
	Reason             *string                   `json:"reason"`
	CurrentRunID       string                    `json:"current_run_id"`
	PreviousRunID      *string                   `json:"previous_run_id"`
	Current            PromptTrackerDenominator  `json:"current"`
	Previous           *PromptTrackerDenominator `json:"previous"`
	Delta              *float64                  `json:"delta"`
	NextCursor         *string                   `json:"next_cursor"`
	PreviousNextCursor *string                   `json:"previous_next_cursor"`
	// Baseline, Category and Strata preserve the stratification the server
	// emits whenever a previous run was supplied. The top-level fields echo
	// the neutral stratum so a response that is decoded and re-encoded does
	// not lose the stratification context.
	Baseline string                 `json:"baseline,omitempty"`
	Category string                 `json:"category,omitempty"`
	Strata   []PromptTrackerStratum `json:"strata,omitempty"`
}

// PromptTrackerStratum reports a single comparison stratum returned by the API.
type PromptTrackerStratum struct {
	Category   string                    `json:"category"`
	Current    PromptTrackerDenominator  `json:"current"`
	Previous   *PromptTrackerDenominator `json:"previous"`
	Comparable bool                      `json:"comparable"`
	Reason     *string                   `json:"reason"`
	Delta      *float64                  `json:"delta"`
}
