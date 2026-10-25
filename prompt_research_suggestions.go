package bisibility

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"unicode/utf8"
)

func promptResearchResourcePath(projectID, resource string) (string, error) {
	return projectResourcePath(projectID, "prompt-research/"+resource), nil
}

func validateSuggestionSnapshot(input PromptResearchSuggestionsPreviewInput) error {
	fail := func(message string) error { return &ConfigurationError{Message: message} }
	if input.CredentialConnectionID != "" && !regexp.MustCompile(`^conn_[a-z][a-z0-9]{23}$`).MatchString(input.CredentialConnectionID) {
		return fail("Expected a conn_ connection ID.")
	}
	seen := map[string]bool{}
	for _, row := range input.InputSnapshot.Competitors {
		if !regexp.MustCompile(`^cmp_[a-z][a-z0-9]{23}$`).MatchString(row.ID) {
			return fail("Expected a cmp_ competitor ID.")
		}
		if seen[row.ID] {
			return fail("Reviewed competitors must have unique public IDs.")
		}
		seen[row.ID] = true
	}
	value := map[string]any{"context": map[string]string{
		"business": input.InputSnapshot.Context.Business, "audience": input.InputSnapshot.Context.Audience,
		"products": input.InputSnapshot.Context.Products, "goals": input.InputSnapshot.Context.Goals,
		"agentRules": input.InputSnapshot.Context.AgentRules,
	}, "competitors": input.InputSnapshot.Competitors}
	var encoded bytes.Buffer
	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return err
	}
	if utf8.RuneCountInString(strings.TrimSuffix(encoded.String(), "\n")) > 5000 {
		return fail("Reviewed input snapshot exceeds 5000 serialized characters.")
	}
	return nil
}

// PromptResearchSuggestionsPreview reviews an exact snapshot without paid provider work.
func (c *Client) PromptResearchSuggestionsPreview(ctx context.Context, projectID string, input PromptResearchSuggestionsPreviewInput, options ...RequestOption) (*DataResponse[PromptResearchSuggestionsPreview], error) {
	path, err := promptResearchResourcePath(projectID, "suggestions/preview")
	if err != nil {
		return nil, err
	}
	if err := validateSuggestionSnapshot(input); err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[DataResponse[PromptResearchSuggestionsPreview]](c, ctx, "POST", path, config)
}

// PromptResearchSuggestionsGenerate uses the frozen forecast and explicit consent with a stable UUID.
func (c *Client) PromptResearchSuggestionsGenerate(ctx context.Context, projectID string, input PromptResearchSuggestionsGenerateInput, options ...RequestOption) (*DataResponse[PromptResearchSuggestionsGeneration], error) {
	path, err := promptResearchResourcePath(projectID, "suggestions/generate")
	if err != nil {
		return nil, err
	}
	if err := validateSuggestionSnapshot(input.Preview.PromptResearchSuggestionsPreviewInput); err != nil {
		return nil, err
	}
	if !input.Consent {
		return nil, &ConfigurationError{Message: "Explicit model generation consent is required."}
	}
	config := newRequestConfig(options...)
	key := config.idempotencyKey
	if key == "" {
		key = config.headers.Get("Idempotency-Key")
	}
	if !regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`).MatchString(key) {
		return nil, &ConfigurationError{Message: "A stable UUID Idempotency-Key is required."}
	}
	config.body = input
	return requestJSON[DataResponse[PromptResearchSuggestionsGeneration]](c, ctx, "POST", path, config)
}

// SuggestPromptResearchPrompts calls the project-scoped Prompt Research contract.
func (c *Client) SuggestPromptResearchPrompts(ctx context.Context, projectID string, options ...RequestOption) (*DataResponse[PromptResearchSuggestions], error) {
	path, err := promptResearchResourcePath(projectID, "suggestions")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	config.body = map[string]any{}
	return requestJSON[DataResponse[PromptResearchSuggestions]](c, ctx, "POST", path, config)
}

// AcceptPromptResearchSuggestions calls the project-scoped Prompt Research contract.
func (c *Client) AcceptPromptResearchSuggestions(ctx context.Context, projectID string, input PromptResearchAcceptanceInput, options ...RequestOption) (*DataResponse[PromptResearchAcceptance], error) {
	path, err := promptResearchResourcePath(projectID, "suggestions/accept")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if err := validatePromptResearchAcceptance(input); err != nil {
		return nil, err
	}
	config.body = input
	return requestJSON[DataResponse[PromptResearchAcceptance]](c, ctx, "POST", path, config)
}

func validatePromptResearchAcceptance(input PromptResearchAcceptanceInput) error {
	for _, draft := range input.Drafts {
		if draft.Popularity != nil {
			return &ConfigurationError{Message: "Caller popularity claims are not accepted."}
		}
		if err := validatePromptTrackerProvenance(draft.GenerationReference, draft.ProviderDatasetReference); err != nil {
			return err
		}
		if (draft.Provenance == "model_generated_hypothesis" && draft.GenerationReference == nil) || (draft.Provenance == "provider_dataset" && draft.ProviderDatasetReference == nil) || (len(draft.EvidenceIDs) > 0 && draft.GenerationReference == nil && draft.ProviderDatasetReference == nil) {
			return &ConfigurationError{Message: "Evidence claims require a trusted provenance reference."}
		}
	}
	return nil
}

func validatePromptTrackerProvenance(generation *PromptTrackerGenerationReference, dataset *PromptTrackerProviderDatasetReference) error {
	if generation != nil && dataset != nil {
		return &ConfigurationError{Message: "Choose one trusted provenance reference."}
	}
	if generation != nil && (!regexp.MustCompile(`^asg_[a-z][a-z0-9]{23}$`).MatchString(generation.GenerationID) || !regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`).MatchString(generation.DraftID)) {
		return &ConfigurationError{Message: "Expected a trusted generation ID and draft UUID."}
	}
	if dataset != nil && (!regexp.MustCompile(`^agr_[a-z][a-z0-9]{23}$`).MatchString(dataset.ReportID) || dataset.RowIndex < 0) {
		return &ConfigurationError{Message: "Expected a trusted dataset report and row index."}
	}
	return nil
}
