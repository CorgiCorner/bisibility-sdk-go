package bisibility

import "regexp"

func validatePromptTrackerInput(input any) error {
	valid := func(value, prefix string) bool {
		return regexp.MustCompile("^" + prefix + "_[a-z][a-z0-9]{23}$").MatchString(value)
	}
	switch row := input.(type) {
	case PromptTrackerPromptInput:
		if err := validatePromptTrackerProvenance(row.GenerationReference, row.ProviderDatasetReference); err != nil {
			return err
		}
		if row.TopicID != nil && !valid(*row.TopicID, "ait") {
			return &ConfigurationError{Message: "Expected an ait_ topic ID."}
		}
	case PromptTrackerPromptPatch:
		if err := validatePromptTrackerProvenance(row.GenerationReference, row.ProviderDatasetReference); err != nil {
			return err
		}
		if ptr := row.TopicID.Value(); row.TopicID.IsSet() && ptr != nil && !valid(*ptr, "ait") {
			return &ConfigurationError{Message: "Expected an ait_ topic ID."}
		}
	case PromptTrackerPreviewInput:
		for _, id := range row.PromptIDs {
			if !valid(id, "aip") {
				return &ConfigurationError{Message: "Expected an aip_ prompt ID."}
			}
		}
		if row.CredentialConnectionID != "" && !valid(row.CredentialConnectionID, "conn") {
			return &ConfigurationError{Message: "Expected a conn_ connection ID."}
		}
	case PromptTrackerRunInput:
		if !row.Consent {
			return &ConfigurationError{Message: "Explicit tracking consent is required."}
		}
		if row.ScheduleID != "" && !valid(row.ScheduleID, "ais") {
			return &ConfigurationError{Message: "Expected an ais_ schedule ID."}
		}
		return validatePromptTrackerInput(row.PromptTrackerPreviewInput)
	case PromptTrackerScheduleInput:
		return validatePromptTrackerInput(row.Configuration)
	case PromptTrackerSchedulePatch:
		if row.Configuration != nil {
			return validatePromptTrackerInput(*row.Configuration)
		}
	}
	return nil
}
