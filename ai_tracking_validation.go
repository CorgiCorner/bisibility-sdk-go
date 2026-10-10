package bisibility

import "regexp"

func validateAITrackingInput(input any) error {
	valid := func(value, prefix string) bool {
		return regexp.MustCompile("^" + prefix + "_[a-z][a-z0-9]{23}$").MatchString(value)
	}
	switch row := input.(type) {
	case AITrackingPromptInput:
		if err := validateTrackingProvenance(row.GenerationReference, row.ProviderDatasetReference); err != nil {
			return err
		}
		if row.TopicID != nil && !valid(*row.TopicID, "ait") {
			return &ConfigurationError{Message: "Expected an ait_ topic ID."}
		}
	case AITrackingPromptPatch:
		if err := validateTrackingProvenance(row.GenerationReference, row.ProviderDatasetReference); err != nil {
			return err
		}
		if ptr := row.TopicID.Value(); row.TopicID.IsSet() && ptr != nil && !valid(*ptr, "ait") {
			return &ConfigurationError{Message: "Expected an ait_ topic ID."}
		}
	case AITrackingAcceptanceInput:
		for _, draft := range row.Drafts {
			if draft.Popularity != nil {
				return &ConfigurationError{Message: "Caller popularity claims are not accepted."}
			}
			if err := validateTrackingProvenance(draft.GenerationReference, draft.ProviderDatasetReference); err != nil {
				return err
			}
			if (draft.Provenance == "model_generated_hypothesis" && draft.GenerationReference == nil) || (draft.Provenance == "provider_dataset" && draft.ProviderDatasetReference == nil) || (len(draft.EvidenceIDs) > 0 && draft.GenerationReference == nil && draft.ProviderDatasetReference == nil) {
				return &ConfigurationError{Message: "Evidence claims require a trusted provenance reference."}
			}
		}

	case AITrackingPreviewInput:
		for _, id := range row.PromptIDs {
			if !valid(id, "aip") {
				return &ConfigurationError{Message: "Expected an aip_ prompt ID."}
			}
		}
		if row.CredentialConnectionID != "" && !valid(row.CredentialConnectionID, "conn") {
			return &ConfigurationError{Message: "Expected a conn_ connection ID."}
		}
	case AITrackingRunInput:
		if !row.Consent {
			return &ConfigurationError{Message: "Explicit tracking consent is required."}
		}
		if row.ScheduleID != "" && !valid(row.ScheduleID, "ais") {
			return &ConfigurationError{Message: "Expected an ais_ schedule ID."}
		}
		return validateAITrackingInput(row.AITrackingPreviewInput)
	case AITrackingScheduleInput:
		return validateAITrackingInput(row.Configuration)
	case AITrackingSchedulePatch:
		if row.Configuration != nil {
			return validateAITrackingInput(*row.Configuration)
		}
	}
	return nil
}
