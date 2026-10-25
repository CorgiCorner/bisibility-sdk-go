package bisibility

import "reflect"

// Prompt Tracker and Prompt Research extend runtime validation without mutating the historical core registry.
var promptTrackerPublicIDPrefixes = map[PublicIDPrefix]struct{}{
	"ait": {}, "aip": {}, "apr": {}, "ais": {}, "air": {}, "asm": {}, "asg": {},
}
var promptTrackerResponsePublicIDFields = map[reflect.Type]map[string]PublicIDPrefix{
	reflect.TypeOf(PromptTrackerPromptRevision{}):           {"id": "apr"},
	reflect.TypeOf(PromptTrackerTopic{}):                    {"id": "ait"},
	reflect.TypeOf(PromptTrackerPrompt{}):                   {"id": "aip", "topic_id": "ait"},
	reflect.TypeOf(PromptTrackerSchedule{}):                 {"id": "ais"},
	reflect.TypeOf(PromptTrackerRun{}):                      {"id": "air"},
	reflect.TypeOf(PromptTrackerSample{}):                   {"id": "asm", "prompt_revision_id": "apr"},
	reflect.TypeOf(PromptTrackerPreview{}):                  {"credential_connection_id": PublicIDPrefixConn},
	reflect.TypeOf(PromptResearchSuggestionsGeneration{}):   {"generation_id": "asg"},
	reflect.TypeOf(PromptResearchSuggestionsPreview{}):      {"credential_connection_id": PublicIDPrefixConn},
	reflect.TypeOf(PromptResearchReviewedCompetitor{}):      {"id": "cmp"},
	reflect.TypeOf(PromptTrackerGenerationReference{}):      {"generation_id": "asg"},
	reflect.TypeOf(PromptTrackerProviderDatasetReference{}): {"report_id": "agr"},
}

func isRegisteredPublicIDPrefix(prefix PublicIDPrefix) bool {
	if _, ok := publicIDPrefixes[prefix]; ok {
		return true
	}
	_, ok := promptTrackerPublicIDPrefixes[prefix]
	return ok
}
