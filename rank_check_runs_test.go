package bisibility

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

const rankRunProject = "prj_aaaaaaaaaaaaaaaaaaaaaaaa"
const rankRunID = "rcr_aaaaaaaaaaaaaaaaaaaaaaaa"

// Synthetic response matching the finite public projection, without any check row.
const rankRunJSON = `{"id":"rcr_aaaaaaaaaaaaaaaaaaaaaaaa","project_id":"prj_aaaaaaaaaaaaaaaaaaaaaaaa","status":"blocked","outcome":null,"terminal":false,"polling_state":"blocked","trigger":"scheduled","counts_basis":"planned_or_unlaunched","counts":{"requested":1,"total":1,"completed":0,"failed":0,"deferred":0,"skipped":0,"cancelled":0},"keyword_count":1,"target_count":1,"planned_for":"2026-10-12T12:00:00.000Z","launched_at":null,"started_at":null,"finished_at":null,"blocked_reason":"budget_exhausted","parent_run_id":null,"cost_cents":null,"usage":{"actual":null,"estimated":null,"unit":"operations"},"native_estimate":{"provider_id":null,"unit":null,"quantity":null,"unknown_targets":1}}`

func TestRankCheckRunReadsUseGETAndPreserveEvidence(t *testing.T) {
	cursor := "opaque+cursor/="
	for _, list := range []bool{false, true} {
		t.Run(map[bool]string{false: "detail", true: "list"}[list], func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				path := "/api/v1/projects/" + rankRunProject + "/rank-check-runs"
				if !list {
					path += "/" + rankRunID
				}
				if r.Method != http.MethodGet || r.URL.Path != path {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				if r.Header.Get("Authorization") != "Bearer bsb_key_test_fixture" || r.Header.Get("X-Bisibility-Source") == "" {
					t.Error("shared authentication or source header lost")
				}
				w.Header().Set("Content-Type", "application/json")
				if list {
					query := r.URL.Query()
					if query.Get("segment") != "planned" || query.Get("status") != "planned,blocked" || query.Get("limit") != "25" || query.Get("cursor") != cursor {
						t.Errorf("filters changed: %v", query)
					}
					w.Write([]byte(`{"data":[` + rankRunJSON + `],"meta":{"next_cursor":"next+opaque/="}}`))
				} else {
					w.Write([]byte(rankRunJSON))
				}
			}))
			defer server.Close()
			client, err := NewClient(WithBaseURL(server.URL+"/api/v1"), WithAPIKey("bsb_key_test_fixture"))
			if err != nil {
				t.Fatal(err)
			}
			var result *RankCheckRunRead
			if list {
				page, err := client.ListRankCheckRuns(context.Background(), rankRunProject, &ListRankCheckRunsOptions{Segment: "planned", Status: "planned,blocked", Limit: 25, Cursor: cursor})
				if err != nil {
					t.Fatal(err)
				}
				if len(page.Data) != 1 || page.Meta.NextCursor == nil || *page.Meta.NextCursor != "next+opaque/=" {
					t.Fatalf("pagination lost: %#v", page)
				}
				result = &page.Data[0]
			} else {
				result, err = client.GetRankCheckRun(context.Background(), rankRunProject, rankRunID)
				if err != nil {
					t.Fatal(err)
				}
			}
			if result.CostCents != nil || result.Usage == nil || result.Usage.Actual != nil || result.NativeEstimate == nil || result.NativeEstimate.Quantity != nil || result.Terminal || result.PollingState != "blocked" {
				t.Errorf("unknown or blocked evidence changed: %#v", result)
			}
			raw, err := json.Marshal(result)
			if err != nil {
				t.Fatal(err)
			}
			var actual, expected any
			json.Unmarshal(raw, &actual)
			json.Unmarshal([]byte(rankRunJSON), &expected)
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("public projection changed: %s", raw)
			}
			if calls != 1 {
				t.Errorf("unexpected HTTP calls: %d", calls)
			}
		})
	}
}

func TestRankCheckRunReadsValidateBeforeHTTP(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(500) }))
	defer server.Close()
	client, err := NewClient(WithBaseURL(server.URL+"/api/v1"), WithAPIKey("bsb_key_test_fixture"), WithMaxRetries(0))
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []*ListRankCheckRunsOptions{{Limit: -1}, {Limit: 201}, {Segment: "future_segment"}, {Status: "blocked,invalid"}} {
		if _, err := client.ListRankCheckRuns(context.Background(), rankRunProject, input); err == nil {
			t.Errorf("invalid filters accepted: %#v", input)
		}
	}
	if _, err := client.ListRankCheckRuns(context.Background(), rankRunID, nil); err == nil {
		t.Error("foreign project ID accepted")
	}
	for _, id := range []string{"", rankRunProject, "rcr_invalid"} {
		if _, err := client.GetRankCheckRun(context.Background(), rankRunProject, id); err == nil {
			t.Errorf("invalid run ID accepted: %q", id)
		}
	}
	if calls != 0 {
		t.Errorf("invalid input made %d HTTP requests", calls)
	}
}

func TestRankCheckRunReadContract(t *testing.T) {
	for _, mutation := range []string{"missing", "null_count", "negative_count", "foreign_id", "foreign_parent", "future_status", "fractional_cost"} {
		t.Run(mutation, func(t *testing.T) {
			var fields map[string]any
			json.Unmarshal([]byte(rankRunJSON), &fields)
			switch mutation {
			case "missing":
				delete(fields, "terminal")
			case "null_count":
				fields["counts"].(map[string]any)["requested"] = nil
			case "negative_count":
				fields["counts"].(map[string]any)["requested"] = -1
			case "foreign_id":
				fields["id"] = rankRunProject
			case "foreign_parent":
				fields["parent_run_id"] = rankRunProject
			case "future_status":
				fields["status"] = "future_status"
				fields["polling_state"] = "future_state"
			case "fractional_cost":
				fields["cost_cents"] = 0.125
				fields["native_estimate"].(map[string]any)["quantity"] = 1.25
			}
			raw, _ := json.Marshal(fields)
			var run RankCheckRunRead
			err := json.Unmarshal(raw, &run)
			if mutation == "future_status" || mutation == "fractional_cost" {
				if err != nil {
					t.Fatal(err)
				}
				if mutation == "fractional_cost" && (run.CostCents == nil || *run.CostCents != 0.125 || run.NativeEstimate.Quantity == nil || *run.NativeEstimate.Quantity != 1.25) {
					t.Error("fractional evidence lost")
				}
			} else if err == nil {
				t.Error("malformed run contract accepted")
			}
		})
	}
}

func TestRankCheckRunReadAuthAndCancellation(t *testing.T) {
	for _, status := range []int{401, 403, 404} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				w.Write([]byte(`{"error":{"code":"denied","message":"Denied"}}`))
			}))
			defer server.Close()
			client, _ := NewClient(WithBaseURL(server.URL+"/api/v1"), WithAPIKey("bsb_key_test_fixture"))
			if _, err := client.GetRankCheckRun(context.Background(), rankRunProject, rankRunID); err == nil {
				t.Error("HTTP refusal accepted")
			}
			if calls != 1 {
				t.Errorf("refusal retried: %d", calls)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			_, err := client.GetRankCheckRun(ctx, rankRunProject, rankRunID)
			if !errors.Is(err, context.Canceled) || calls != 1 {
				t.Errorf("cancellation lost: %v, calls %d", err, calls)
			}
		})
	}
}

func TestRankCheckRunListRejectsMalformedPagination(t *testing.T) {
	for _, body := range []string{`{"data":[],"meta":{}}`, `{"data":null,"meta":{"next_cursor":null}}`, `{"data":[],"meta":{"next_cursor":1}}`} {
		var page RankCheckRunReadList
		if err := json.Unmarshal([]byte(body), &page); err == nil {
			t.Errorf("malformed pagination accepted: %s", body)
		}
	}
	var page RankCheckRunReadList
	if err := json.Unmarshal([]byte(`{"data":[],"meta":{"next_cursor":null}}`), &page); err != nil || page.Meta.NextCursor != nil {
		t.Errorf("empty list failed: %v", err)
	}
	var run RankCheckRunRead
	terminal := strings.ReplaceAll(rankRunJSON, `"status":"blocked"`, `"status":"cancelled"`)
	terminal = strings.ReplaceAll(terminal, `"terminal":false`, `"terminal":true`)
	terminal = strings.ReplaceAll(terminal, `"polling_state":"blocked"`, `"polling_state":"terminal"`)
	if err := json.Unmarshal([]byte(terminal), &run); err != nil {
		t.Fatal(err)
	}
	if !run.Terminal || run.Status != "cancelled" || run.PollingState != "terminal" {
		t.Error("terminal run without a check was lost")
	}
}

func TestRankCheckRunMetadataRejectsForeignRunID(t *testing.T) {
	for _, id := range []string{rankRunProject, rankRunID} {
		t.Run(id, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"id":"check_aaaaaaaaaaaaaaaaaaaaaaaa","keyword_id":"kw_aaaaaaaaaaaaaaaaaaaaaaaa","checked_at":"2026-10-12T12:00:00Z","run":{"id":"` + id + `","started_at":null,"finished_at":null,"status":"blocked","trigger":"scheduled"}}`))
			}))
			defer server.Close()
			client, _ := NewClient(WithBaseURL(server.URL+"/api/v1"), WithAPIKey("bsb_key_test_fixture"))
			_, err := client.GetRankCheckResult(context.Background(), "check_aaaaaaaaaaaaaaaaaaaaaaaa")
			if id == rankRunProject && err == nil {
				t.Error("foreign nested run ID accepted")
			}
			if id == rankRunID && err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRankCheckRunReadsRejectForeignScope(t *testing.T) {
	foreignProject := "prj_bbbbbbbbbbbbbbbbbbbbbbbb"
	foreignRun := "rcr_bbbbbbbbbbbbbbbbbbbbbbbb"
	for _, test := range []struct {
		name string
		list bool
		body string
	}{
		{"detail project", false, strings.ReplaceAll(rankRunJSON, rankRunProject, foreignProject)},
		{"detail run", false, strings.ReplaceAll(rankRunJSON, rankRunID, foreignRun)},
		{"list first project", true, `{"data":[` + strings.ReplaceAll(rankRunJSON, rankRunProject, foreignProject) + `],"meta":{"next_cursor":null}}`},
		{"list second project", true, `{"data":[` + rankRunJSON + `,` + strings.ReplaceAll(rankRunJSON, rankRunProject, foreignProject) + `],"meta":{"next_cursor":null}}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(test.body))
			}))
			defer server.Close()
			client, _ := NewClient(WithBaseURL(server.URL+"/api/v1"), WithAPIKey("bsb_key_test_fixture"))
			var err error
			path := "/api/v1/projects/" + rankRunProject + "/rank-check-runs"
			if test.list {
				_, err = client.ListRankCheckRuns(context.Background(), rankRunProject, nil)
			} else {
				path += "/" + rankRunID
				_, err = client.GetRankCheckRun(context.Background(), rankRunProject, rankRunID)
			}
			if err == nil {
				t.Fatal("valid foreign public IDs accepted outside requested scope")
			}
			var responseError *ResponseError
			if !errors.As(err, &responseError) || responseError.StatusCode != http.StatusOK || responseError.Method != http.MethodGet || responseError.Body != test.body || responseError.URL != server.URL+path {
				t.Errorf("scope contract error lost response evidence: %v", err)
			}
			if calls != 1 {
				t.Errorf("scope error retried: %d", calls)
			}
		})
	}
}

func TestRankCheckRunRouteIdentifierGuard(t *testing.T) {
	path := "/projects/" + rankRunProject + "/rank-check-runs"
	if err := validateRoutePublicIDs(path); err != nil {
		t.Fatal(err)
	}
	if err := validateRoutePublicIDs(path + "/" + rankRunID); err != nil {
		t.Fatal(err)
	}
	if err := validateRoutePublicIDs(path + "/" + rankRunProject); err == nil {
		t.Error("foreign run namespace accepted by shared route guard")
	}
}

func TestRankCheckRunRetainsUnknownExtension(t *testing.T) {
	raw := strings.TrimSuffix(rankRunJSON, "}") + `,"future_evidence":{"rank_scale":"finder_group","certainty":null}}`
	var run RankCheckRunRead
	if err := json.Unmarshal([]byte(raw), &run); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(run)
	if err != nil {
		t.Fatal(err)
	}
	var actual, expected any
	json.Unmarshal(encoded, &actual)
	json.Unmarshal([]byte(raw), &expected)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("future extension was lost: %s", encoded)
	}
}
