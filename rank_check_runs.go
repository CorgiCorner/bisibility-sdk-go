package bisibility

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

//go:embed rank_check_run_contract.json
var rankRunContractJSON []byte

// ListRankCheckRunsOptions filters project runs, including unlaunched outcomes.
// Status accepts comma-separated execution statuses. Cursor is opaque.
type ListRankCheckRunsOptions struct {
	Segment string
	Status  string
	Cursor  string
	Limit   int
}

// RankCheckRunCounts retains server counts without requiring completed check rows.
type RankCheckRunCounts struct {
	Requested int `json:"requested"`
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
	Deferred  int `json:"deferred"`
	Skipped   int `json:"skipped"`
	Cancelled int `json:"cancelled"`
}

// RankCheckRunUsage preserves unknown quantities as nil.
type RankCheckRunUsage struct {
	Actual    *int   `json:"actual"`
	Estimated *int   `json:"estimated"`
	Unit      string `json:"unit"`
}

// RankCheckRunNativeEstimate preserves fractional provider quantities.
type RankCheckRunNativeEstimate struct {
	ProviderID     *string  `json:"provider_id"`
	Unit           *string  `json:"unit"`
	Quantity       *float64 `json:"quantity"`
	UnknownTargets int      `json:"unknown_targets"`
}

// RankCheckRunRead is the finite public run projection. Execution and polling
// states remain strings so that future server states can still be read.
// Nullable timestamps retain the server's original RFC 3339 text.
type RankCheckRunRead struct {
	// Extensions retains additional top-level fields from future server versions.
	Extensions     map[string]json.RawMessage  `json:"-"`
	ID             string                      `json:"id"`
	ProjectID      string                      `json:"project_id"`
	Status         string                      `json:"status"`
	Outcome        *string                     `json:"outcome"`
	Terminal       bool                        `json:"terminal"`
	PollingState   string                      `json:"polling_state"`
	Trigger        string                      `json:"trigger"`
	CountsBasis    string                      `json:"counts_basis"`
	Counts         RankCheckRunCounts          `json:"counts"`
	KeywordCount   int                         `json:"keyword_count"`
	TargetCount    int                         `json:"target_count"`
	PlannedFor     *string                     `json:"planned_for"`
	LaunchedAt     *string                     `json:"launched_at"`
	StartedAt      *string                     `json:"started_at"`
	FinishedAt     *string                     `json:"finished_at"`
	BlockedReason  *string                     `json:"blocked_reason"`
	ParentRunID    *string                     `json:"parent_run_id"`
	CostCents      *float64                    `json:"cost_cents"`
	Usage          *RankCheckRunUsage          `json:"usage"`
	NativeEstimate *RankCheckRunNativeEstimate `json:"native_estimate"`
}

// RankCheckRunReadList contains a page of public run projections.
type RankCheckRunReadList struct {
	Data []RankCheckRunRead `json:"data"`
	Meta ListMeta           `json:"meta"`
}

func unmarshalRankRunContract(raw []byte, definition string, into any) error {
	var definitions map[string]reportSchema
	if err := json.Unmarshal(rankRunContractJSON, &definitions); err != nil {
		return err
	}
	schema, ok := definitions[definition]
	if !ok || !validReportValue(raw, schema, definitions) {
		return fmt.Errorf("invalid required rank-check run contract: %s", definition)
	}
	return json.Unmarshal(raw, into)
}

func (r *RankCheckRunRead) UnmarshalJSON(raw []byte) error {
	type wire RankCheckRunRead
	var decoded wire
	if err := unmarshalRankRunContract(raw, "RankCheckRunRead", &decoded); err != nil {
		return err
	}
	known, err := json.Marshal(decoded)
	if err != nil {
		return err
	}
	extensions, err := reportExtensions(raw, known)
	if err != nil {
		return err
	}
	decoded.Extensions = extensions
	*r = RankCheckRunRead(decoded)
	return nil
}

func (r *RankCheckRunReadList) UnmarshalJSON(raw []byte) error {
	type wire RankCheckRunReadList
	return unmarshalRankRunContract(raw, "RankCheckRunReadList", (*wire)(r))
}

// ListRankCheckRuns reads project runs without launching a measurement.
func (c *Client) ListRankCheckRuns(ctx context.Context, projectID string, input *ListRankCheckRunsOptions, options ...RequestOption) (*RankCheckRunReadList, error) {
	if err := requirePublicID(projectID, "projectID", PublicIDPrefixProject); err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if input != nil {
		if input.Segment != "" && input.Segment != "history" && input.Segment != "planned" {
			return nil, &ConfigurationError{Message: "segment must be history or planned."}
		}
		if input.Limit < 0 || input.Limit > 200 {
			return nil, &ConfigurationError{Message: "limit must be between 1 and 200, or zero to use the server default."}
		}
		if input.Status != "" {
			for _, status := range strings.Split(input.Status, ",") {
				switch strings.TrimSpace(status) {
				case "planned", "blocked", "queued", "running", "cancelling", "completed", "cancelled":
				default:
					return nil, &ConfigurationError{Message: "status contains an unsupported execution status."}
				}
			}
		}
		addQuery(config.query, "segment", input.Segment)
		addQuery(config.query, "status", input.Status)
		addQuery(config.query, "cursor", input.Cursor)
		if input.Limit != 0 {
			config.query.Set("limit", strconv.Itoa(input.Limit))
		}
	}
	return requestJSONWithValidation[RankCheckRunReadList](c, ctx, http.MethodGet, projectResourcePath(projectID, "rank-check-runs"), config, func(page *RankCheckRunReadList) error {
		for _, run := range page.Data {
			if run.ProjectID != projectID {
				return fmt.Errorf("rank-check run project_id does not match the requested project")
			}
		}
		return nil
	})
}

// GetRankCheckRun reads an existing run, including blocked or cancelled runs
// without a rank-check result.
func (c *Client) GetRankCheckRun(ctx context.Context, projectID, runID string, options ...RequestOption) (*RankCheckRunRead, error) {
	if err := requirePublicID(projectID, "projectID", PublicIDPrefixProject); err != nil {
		return nil, err
	}
	if err := requirePublicID(runID, "runID", PublicIDPrefixRun); err != nil {
		return nil, err
	}
	return requestJSONWithValidation[RankCheckRunRead](c, ctx, http.MethodGet, projectResourcePath(projectID, "rank-check-runs/"+url.PathEscape(runID)), newRequestConfig(options...), func(run *RankCheckRunRead) error {
		if run.ProjectID != projectID || run.ID != runID {
			return fmt.Errorf("rank-check run IDs do not match the requested project and run")
		}
		return nil
	})
}

func (r RankCheckRunRead) MarshalJSON() ([]byte, error) {
	type wire RankCheckRunRead
	known, err := json.Marshal(wire(r))
	if err != nil {
		return nil, err
	}
	return mergeReportExtensions(known, r.Extensions)
}
