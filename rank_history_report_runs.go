package bisibility

// RankHistoryReportRunItem is a retained item without an attached RankCheck.
type RankHistoryReportRunItem struct {
	Kind            string  `json:"kind"`
	ID              string  `json:"id"`
	RunID           string  `json:"run_id"`
	KeywordID       string  `json:"keyword_id"`
	Keyword         string  `json:"keyword"`
	RecordedAt      string  `json:"recorded_at"`
	StartedAt       *string `json:"started_at"`
	FinishedAt      *string `json:"finished_at"`
	NotBefore       *string `json:"not_before"`
	Status          string  `json:"status"`
	BlockedReason   *string `json:"blocked_reason"`
	RunStatus       string  `json:"run_status"`
	RunOutcome      *string `json:"run_outcome"`
	RankCheckState  string  `json:"rank_check_state"`
	ScopeMembership string  `json:"scope_membership"`
}

type RankHistoryReportRunCounts struct {
	Requested int `json:"requested"`
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
	Deferred  int `json:"deferred"`
	Skipped   int `json:"skipped"`
	Cancelled int `json:"cancelled"`
}

// RankHistoryReportRun explicitly distinguishes retained item membership from project-only scope.
type RankHistoryReportRun struct {
	Kind            string                     `json:"kind"`
	ID              string                     `json:"id"`
	RunID           string                     `json:"run_id"`
	RecordedAt      string                     `json:"recorded_at"`
	StartedAt       *string                    `json:"started_at"`
	FinishedAt      *string                    `json:"finished_at"`
	PlannedFor      *string                    `json:"planned_for"`
	LaunchedAt      *string                    `json:"launched_at"`
	Status          string                     `json:"status"`
	Outcome         *string                    `json:"outcome"`
	Trigger         string                     `json:"trigger"`
	BlockedReason   *string                    `json:"blocked_reason"`
	SelectionKind   string                     `json:"selection_kind"`
	RankCheckState  string                     `json:"rank_check_state"`
	ScopeMembership string                     `json:"scope_membership"`
	Counts          RankHistoryReportRunCounts `json:"counts"`
}
