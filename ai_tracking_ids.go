package bisibility

import "reflect"

// Tracking extends runtime validation without mutating the historical core registry.
var trackingPublicIDPrefixes = map[PublicIDPrefix]struct{}{
	"ait": {}, "aip": {}, "apr": {}, "ais": {}, "air": {}, "asm": {}, "asg": {},
}
var trackingResponsePublicIDFields = map[reflect.Type]map[string]PublicIDPrefix{
	reflect.TypeOf(AITrackingPromptRevision{}):           {"id": "apr"},
	reflect.TypeOf(AITrackingTopic{}):                    {"id": "ait"},
	reflect.TypeOf(AITrackingPrompt{}):                   {"id": "aip", "topic_id": "ait"},
	reflect.TypeOf(AITrackingSchedule{}):                 {"id": "ais"},
	reflect.TypeOf(AITrackingRun{}):                      {"id": "air"},
	reflect.TypeOf(AITrackingSample{}):                   {"id": "asm", "prompt_revision_id": "apr"},
	reflect.TypeOf(AITrackingPreview{}):                  {"credential_connection_id": PublicIDPrefixConn},
	reflect.TypeOf(AITrackingSuggestionsGeneration{}):    {"generation_id": "asg"},
	reflect.TypeOf(AITrackingSuggestionsPreview{}):       {"credential_connection_id": PublicIDPrefixConn},
	reflect.TypeOf(AITrackingReviewedCompetitor{}):       {"id": "cmp"},
	reflect.TypeOf(AITrackingGenerationReference{}):      {"generation_id": "asg"},
	reflect.TypeOf(AITrackingProviderDatasetReference{}): {"report_id": "agr"},
}

func isRegisteredPublicIDPrefix(prefix PublicIDPrefix) bool {
	if _, ok := publicIDPrefixes[prefix]; ok {
		return true
	}
	_, ok := trackingPublicIDPrefixes[prefix]
	return ok
}
