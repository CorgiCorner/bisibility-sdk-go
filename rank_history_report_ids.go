package bisibility

import "reflect"

// V2 variants have their own resource identities. Retained item IDs are opaque,
// while their run and keyword identities retain the normal public-ID guards.
var reportResponsePublicIDFields = map[reflect.Type]map[string]PublicIDPrefix{
	reflect.TypeOf(RankHistoryReportObservation{}):    {"id": PublicIDPrefixCheck, "keyword_id": PublicIDPrefixKeyword},
	reflect.TypeOf(RankHistoryReportAttempt{}):        {"id": PublicIDPrefixCheck, "keyword_id": PublicIDPrefixKeyword},
	reflect.TypeOf(RankHistoryReportRun{}):            {"id": PublicIDPrefixRun, "run_id": PublicIDPrefixRun},
	reflect.TypeOf(RankHistoryReportRunItem{}):        {"run_id": PublicIDPrefixRun, "keyword_id": PublicIDPrefixKeyword},
	reflect.TypeOf(RankHistoryReportNormalizedItem{}): {"check_id": PublicIDPrefixCheck, "keyword_id": PublicIDPrefixKeyword},
}
