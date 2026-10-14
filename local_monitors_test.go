package bisibility

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

func localFixture(t *testing.T) map[string]json.RawMessage {
	t.Helper()
	raw, err := os.ReadFile("testdata/local-monitor-transport.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture map[string]json.RawMessage
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	return fixture
}

const localProject = "prj_abcdefghijklmnopqrstuvwx"
const localMonitor = "lmn_abcdefghijklmnopqrstuvwx"
const localRun = "lmr_abcdefghijklmnopqrstuvwx"
const localAt = "2026-10-12T12:00:00.000Z"

func TestLocalActualServerGETRoundtrip(t *testing.T) {
	f := localFixture(t)
	var result LocalMonitorRunResultResponse
	if err := json.Unmarshal(f["result"], &result); err != nil {
		t.Fatal(err)
	}
	point := result.Data.Points[0].ID
	cursor := "opaque+cursor/="
	tests := []struct {
		name, key, path string
		run             func(*Client) (any, error)
	}{
		{"history", "history", "/runs", func(c *Client) (any, error) {
			return c.ListLocalMonitorRuns(context.Background(), localProject, localMonitor, &LocalMonitorHistoryOptions{AsOf: localAt, From: "2026-10-12T00:00:00.000Z", To: localAt, Cursor: &cursor})
		}},
		{"result", "result", "/runs/" + localRun + "/results", func(c *Client) (any, error) {
			return c.GetLocalMonitorRunResult(context.Background(), localProject, localMonitor, localRun, &LocalMonitorReadOptions{AsOf: localAt})
		}},
		{"listings", "listings", "/runs/" + localRun + "/points/" + point + "/listings", func(c *Client) (any, error) {
			return c.ListLocalMonitorPointListings(context.Background(), localProject, localMonitor, localRun, point, &LocalMonitorReadOptions{AsOf: localAt})
		}},
		{"point_history", "point_history", "/runs/" + localRun + "/points/" + point + "/history", func(c *Client) (any, error) {
			return c.ListLocalMonitorPointHistory(context.Background(), localProject, localMonitor, localRun, point, &LocalMonitorHistoryOptions{AsOf: localAt, Cursor: &cursor})
		}},
		{"export", "export_json", "/runs/" + localRun + "/export", func(c *Client) (any, error) {
			r, e := c.ExportLocalMonitorRun(context.Background(), localProject, localMonitor, localRun, &LocalMonitorExportOptions{AsOf: localAt})
			if e != nil {
				return nil, e
			}
			return r.Report, nil
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/capabilities") {
					w.Write([]byte(`{"data":[],"apiVersions":["v1"]}`))
					return
				}
				calls++
				if r.Method != "GET" || r.URL.Path != "/api/v1/projects/"+localProject+"/local-monitors/"+localMonitor+tt.path {
					t.Errorf("wrong method/path: %s %s", r.Method, r.URL.Path)
				}
				if r.URL.Query().Get("as_of") != localAt {
					t.Error("as_of lost")
				}
				if tt.key == "history" || tt.key == "point_history" {
					if r.URL.Query().Get("cursor") != cursor {
						t.Error("cursor changed")
					}
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write(f[tt.key])
			}))
			defer server.Close()
			c, err := NewClient(WithBaseURL(server.URL+"/api/v1"), WithAPIKey("bsb_key_test_fixture"))
			if err != nil {
				t.Fatal(err)
			}
			value, err := tt.run(c)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			var actual, expected any
			json.Unmarshal(raw, &actual)
			json.Unmarshal(f[tt.key], &expected)
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("retained contract lost: %s", raw)
			}
			if calls != 1 {
				t.Fatalf("calls=%d", calls)
			}
		})
	}
}

func TestLocalUnknownStaleFutureAndCID(t *testing.T) {
	f := localFixture(t)
	for _, key := range []string{"result_unknown", "result_stale"} {
		var r LocalMonitorRunResultResponse
		if err := json.Unmarshal(f[key], &r); err != nil {
			t.Fatal(err)
		}
		if r.Data.FreshnessPolicy == nil || r.Data.FreshnessPolicy.PolicyEpoch == nil || *r.Data.FreshnessPolicy.PolicyEpoch != "8" {
			t.Error("policy lost")
		}
	}
	var gridRaw map[string]any
	json.Unmarshal(f["result"], &gridRaw)
	gridRaw["data"].(map[string]any)["grids"].([]any)[0].(map[string]any)["size"] = 9
	gridBody, _ := json.Marshal(gridRaw)
	var futureGrid LocalMonitorRunResultResponse
	if err := json.Unmarshal(gridBody, &futureGrid); err != nil {
		t.Fatal(err)
	}
	if futureGrid.Data.Grids[0].Size != 9 {
		t.Error("future numeric metadata lost")
	}
	raw := strings.ReplaceAll(string(f["result"]), `"manual_ttl"`, `"future_policy"`)
	raw = strings.ReplaceAll(raw, `"maps_organic"`, `"finder_group"`)
	var r LocalMonitorRunResultResponse
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		t.Fatal(err)
	}
	var listings map[string]any
	json.Unmarshal(f["listings"], &listings)
	first := listings["data"].([]any)[0].(map[string]any)
	first["cid"] = "18446744073709551615"
	body, _ := json.Marshal(listings)
	var valid LocalMonitorPointListingsResponse
	if err := json.Unmarshal(body, &valid); err != nil {
		t.Fatal(err)
	}
	if valid.Data[0].Cid == nil || *valid.Data[0].Cid != "18446744073709551615" {
		t.Error("opaque CID changed")
	}
	first["cid"] = 1234567890123456789
	body, _ = json.Marshal(listings)
	if json.Unmarshal(body, &valid) == nil {
		t.Error("numeric CID accepted")
	}
}

func TestLocalRequiredContractRejectsMalformed(t *testing.T) {
	f := localFixture(t)
	for _, mutation := range []string{"missing", "version", "foreign", "bounds"} {
		t.Run(mutation, func(t *testing.T) {
			var raw map[string]any
			json.Unmarshal(f["result"], &raw)
			data := raw["data"].(map[string]any)
			switch mutation {
			case "missing":
				delete(data, "points")
			case "version":
				data["version"] = 2
			case "foreign":
				data["project_id"] = localRun
			case "bounds":
				points := data["points"].([]any)
				many := make([]any, 101)
				for i := range many {
					many[i] = points[0]
				}
				data["points"] = many
			}
			body, _ := json.Marshal(raw)
			var r LocalMonitorRunResultResponse
			if json.Unmarshal(body, &r) == nil {
				t.Error("invalid contract accepted")
			}
		})
	}
}

func TestLocalInvalidInputsNeverFetch(t *testing.T) {
	calls := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(500) }))
	defer s.Close()
	c, _ := NewClient(WithBaseURL(s.URL+"/api/v1"), WithAPIKey("bsb_key_test_fixture"))
	empty := ""
	long := strings.Repeat("x", 8001)
	for _, input := range []*LocalMonitorHistoryOptions{{Cursor: &empty}, {Cursor: &long}, {AsOf: "2026-10-12T12:00:00Z"}, {AsOf: "2026-02-30T12:00:00.000Z"}, {From: localAt, To: "2026-10-12T00:00:00.000Z"}} {
		if _, err := c.ListLocalMonitorRuns(context.Background(), localProject, localMonitor, input); err == nil {
			t.Error("invalid input accepted")
		}
	}
	if _, err := c.GetLocalMonitorRunResult(context.Background(), localRun, localMonitor, localRun, nil); err == nil {
		t.Error("foreign ID accepted")
	}
	if _, err := c.ListLocalMonitorPointListings(context.Background(), localProject, localMonitor, localRun, "", nil); err == nil {
		t.Error("empty target accepted")
	}
	if _, err := c.ExportLocalMonitorRun(context.Background(), localProject, localMonitor, localRun, &LocalMonitorExportOptions{Version: "2"}); err == nil {
		t.Error("unknown export version accepted")
	}
	if calls != 0 {
		t.Fatalf("%d unexpected HTTP calls", calls)
	}
}

func TestLocalCSVRetainsServerText(t *testing.T) {
	f := localFixture(t)
	var csv string
	json.Unmarshal(f["export_csv"], &csv)
	calls := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/capabilities") {
			w.Write([]byte(`{"data":[],"apiVersions":["v1"]}`))
			return
		}
		calls++
		if r.Method != "GET" || r.URL.Query().Get("version") != "1" || r.URL.Query().Get("format") != "csv" {
			t.Error("wrong CSV request")
		}
		w.Write([]byte(csv))
	}))
	defer s.Close()
	c, _ := NewClient(WithBaseURL(s.URL+"/api/v1"), WithAPIKey("bsb_key_test_fixture"))
	r, e := c.ExportLocalMonitorRun(context.Background(), localProject, localMonitor, localRun, &LocalMonitorExportOptions{Format: "csv"})
	if e != nil {
		t.Fatal(e)
	}
	if r.CSV != csv || r.Report != nil || calls != 1 {
		t.Error("CSV changed")
	}
}

func TestLocalAuthRefusalNeverRetries(t *testing.T) {
	for _, status := range []int{401, 403} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			calls := 0
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/capabilities") {
					w.Write([]byte(`{"data":[],"apiVersions":["v1"]}`))
					return
				}
				calls++
				if r.Method != "GET" {
					t.Error("write request")
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				w.Write([]byte(`{"error":{"code":"forbidden","message":"Denied"}}`))
			}))
			defer s.Close()
			c, _ := NewClient(WithBaseURL(s.URL+"/api/v1"), WithAPIKey("bsb_key_test_fixture"))
			if _, err := c.GetLocalMonitorRunResult(context.Background(), localProject, localMonitor, localRun, nil); err == nil {
				t.Error("auth refusal accepted")
			}
			if calls != 1 {
				t.Errorf("calls=%d", calls)
			}
		})
	}
}

func TestLocalListingRetainsFinderGroupRankScale(t *testing.T) {
	f := localFixture(t)
	raw := strings.ReplaceAll(string(f["listings"]), `"maps_organic"`, `"finder_group"`)
	var listings LocalMonitorPointListingsResponse
	if err := json.Unmarshal([]byte(raw), &listings); err != nil {
		t.Fatal(err)
	}
	if listings.Data[0].RankScale == nil || *listings.Data[0].RankScale != "finder_group" {
		t.Fatal("Finder rank scale was lost")
	}
}
