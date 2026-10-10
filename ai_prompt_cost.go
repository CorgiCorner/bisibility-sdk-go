package bisibility

import (
	"encoding/json"
	"regexp"
)

func (input CompareAIPromptsOptions) MarshalJSON() ([]byte, error) {
	type plain CompareAIPromptsOptions
	encoded, err := json.Marshal(plain(input))
	if err != nil || input.CostPolicy != "provider_actual_cost" {
		return encoded, err
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &body); err != nil {
		return nil, err
	}
	delete(body, "max_cost_cents")
	return json.Marshal(body)
}
func validatePromptCostPolicy(input CompareAIPromptsOptions) error {
	fail := func() error {
		return &ConfigurationError{Message: "Invalid prompt cost-policy consent, estimate identity, or locale options."}
	}
	if input.CountryISOCode != "" && (!input.WebSearch || !regexp.MustCompile(`^[A-Z]{2}$`).MatchString(input.CountryISOCode)) {
		return fail()
	}
	if input.MaxOutputTokens != 0 && (input.MaxOutputTokens < 16 || input.MaxOutputTokens > 4096) {
		return fail()
	}
	if input.ResponseLanguage != "" && !regexp.MustCompile(`^[a-z]{2,3}(?:-[a-zA-Z0-9]{2,4})?$`).MatchString(input.ResponseLanguage) {
		return fail()
	}
	if input.CostPolicy == "" || input.CostPolicy == "hard_cap" {
		if input.ActualCostAcknowledgement != "" || input.EstimatedCostLimitCents != nil || input.IdempotencyKey != "" || input.EstimateCredentialsRef != "" {
			return fail()
		}
		return nil
	}
	if input.CostPolicy != "provider_actual_cost" || input.MaxCostCents != 0 || input.ActualCostAcknowledgement != "non_guaranteed_estimate_v1" || input.EstimatedCostLimitCents == nil || *input.EstimatedCostLimitCents < 0 || *input.EstimatedCostLimitCents > 1000 {
		return fail()
	}
	if !input.EstimateOnly && (!regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`).MatchString(input.IdempotencyKey) || !regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(input.EstimateCredentialsRef)) {
		return fail()
	}
	return nil
}
