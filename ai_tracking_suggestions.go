package bisibility

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"unicode/utf8"
)

func validateSuggestionSnapshot(input AITrackingSuggestionsPreviewInput) error {
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

// AITrackingSuggestionsPreview reviews an exact snapshot without paid provider work.
func (c *Client) AITrackingSuggestionsPreview(ctx context.Context, projectID string, input AITrackingSuggestionsPreviewInput, options ...RequestOption) (*DataResponse[AITrackingSuggestionsPreview], error) {
	path, err := trackingResourcePath(projectID, "suggestions/preview", "", "")
	if err != nil {
		return nil, err
	}
	if err := validateSuggestionSnapshot(input); err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[DataResponse[AITrackingSuggestionsPreview]](c, ctx, "POST", path, config)
}

// AITrackingSuggestionsGenerate uses the frozen forecast and explicit consent with a stable UUID.
func (c *Client) AITrackingSuggestionsGenerate(ctx context.Context, projectID string, input AITrackingSuggestionsGenerateInput, options ...RequestOption) (*DataResponse[AITrackingSuggestionsGeneration], error) {
	path, err := trackingResourcePath(projectID, "suggestions/generate", "", "")
	if err != nil {
		return nil, err
	}
	if err := validateSuggestionSnapshot(input.Preview.AITrackingSuggestionsPreviewInput); err != nil {
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
	return requestJSON[DataResponse[AITrackingSuggestionsGeneration]](c, ctx, "POST", path, config)
}

func validateTrackingProvenance(generation *AITrackingGenerationReference, dataset *AITrackingProviderDatasetReference) error {
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
