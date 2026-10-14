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

func reportFixture(t *testing.T) map[string]json.RawMessage {
	t.Helper()
	raw, err := os.ReadFile("testdata/rank-history-report.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture map[string]json.RawMessage
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func TestRankHistoryReportActualServerFields(t *testing.T) {
	for _, stream := range []string{"observations", "recorded", "attempts", "full_attempts"} {
		t.Run(stream, func(t *testing.T) {
			fixture := reportFixture(t)
			var direct RankHistoryExportResponse
			if err := json.Unmarshal(fixture[stream], &direct); err != nil {
				t.Fatalf("actual server fixture decode: %v", err)
			}
			selectedStream := stream
			ids := []string{"kw_abcdefghijklmnopqrstuvwx"}
			if stream == "full_attempts" {
				selectedStream = "attempts"
				ids = nil
			}
			if stream == "recorded" {
				selectedStream = "observations"
			}
			selectedBasis := "recorded_at"
			selectedLimit := 200
			if stream == "observations" {
				selectedBasis = "observed_at"
			}
			if stream == "attempts" {
				selectedLimit = 1
			}
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/capabilities" {
					writeJSON(t, w, 200, map[string][]string{"apiVersions": {"v1"}})
					return
				}
				calls++
				assertEqual(t, r.Method, "GET")
				query := r.URL.Query()
				assertEqual(t, query.Get("version"), "2")
				assertEqual(t, query.Get("from"), "1970-01-01T00:00:00Z")
				assertEqual(t, query.Get("to"), "2026-10-12T12:00:00Z")
				assertEqual(t, query.Get("basis"), selectedBasis)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(fixture[stream])
			}))
			defer server.Close()
			response, err := newTestClient(t, server.URL+"/api/v1").ExportRankHistory(context.Background(), "prj_abcdefghijklmnopqrstuvwx", &ExportRankHistoryOptions{Version: "2", Stream: selectedStream, Basis: selectedBasis, Range: "all", KeywordIDs: ids, From: "1970-01-01T00:00:00Z", To: "2026-10-12T12:00:00Z", Aggregation: "none", Limit: selectedLimit})
			if err != nil {
				t.Fatal(err)
			}
			if response.Meta.Report == nil || response.Data[0].Report == nil {
				t.Fatal("report projection missing")
			}
			encoded, err := json.Marshal(response)
			if err != nil {
				t.Fatal(err)
			}
			var original, roundtrip map[string]json.RawMessage
			if err = json.Unmarshal(fixture[stream], &original); err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(encoded, &roundtrip); err != nil {
				t.Fatal(err)
			}
			// Compare JSON values to ignore only whitespace and object-key ordering.
			for _, key := range []string{"data", "meta"} {
				var a, b interface{}
				if err = json.Unmarshal(original[key], &a); err != nil {
					t.Fatal(err)
				}
				if err = json.Unmarshal(roundtrip[key], &b); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(a, b) {
					t.Fatalf("%s lost report fields: got %s expected %s", key, roundtrip[key], original[key])
				}
			}
			assertEqual(t, calls, 1)
		})
	}
}

func TestRankHistoryReportActualSignedPages(t *testing.T) {
	fixture := reportFixture(t)
	var pages []json.RawMessage
	if err := json.Unmarshal(fixture["attempt_pages"], &pages); err != nil {
		t.Fatal(err)
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		assertEqual(t, query.Get("version"), "2")
		assertEqual(t, query.Get("stream"), "attempts")
		assertEqual(t, query.Get("range"), "all")
		if calls > 0 {
			var previous RankHistoryExportResponse
			if err := json.Unmarshal(pages[calls-1], &previous); err != nil {
				t.Fatal(err)
			}
			assertEqual(t, query.Get("cursor"), *previous.Meta.NextCursor)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(pages[calls])
		calls++
	}))
	defer server.Close()
	rows := collectPager(t, newTestClient(t, server.URL+"/api/v1").IterateRankHistory(context.Background(), "prj_abcdefghijklmnopqrstuvwx", &ExportRankHistoryOptions{Version: "2", Stream: "attempts", Range: "all", Basis: "recorded_at", Aggregation: "none", From: "1970-01-01T00:00:00Z", To: "2026-10-12T12:00:00Z", KeywordIDs: []string{"kw_abcdefghijklmnopqrstuvwx"}, Limit: 1}))
	assertEqual(t, len(rows), len(pages))
	assertEqual(t, calls, len(pages))
	for _, row := range rows {
		if row.Report == nil {
			t.Fatal("attempt variant lost")
		}
	}
}

func TestRankHistoryReportUnknownFieldKeepsStringCID(t *testing.T) {
	fixture := reportFixture(t)
	var envelope struct {
		Data []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(fixture["observations"], &envelope); err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(envelope.Data[0], &fields); err != nil {
		t.Fatal(err)
	}
	fields["cid"] = json.RawMessage(`"18446744073709551615"`)
	fields["evidence"] = json.RawMessage(`"future_evidence"`)
	raw, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	var record RankHistoryReportRecord
	if err := json.Unmarshal(raw, &record); err != nil {
		t.Fatal(err)
	}
	assertEqual(t, string(record.Extensions["cid"]), `"18446744073709551615"`)
	assertEqual(t, record.Observation.Evidence, "future_evidence")
	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	fields = nil
	if err = json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	assertEqual(t, string(fields["cid"]), `"18446744073709551615"`)
}

func TestRankHistoryReportActualCSV(t *testing.T) {
	fixture := reportFixture(t)
	for _, name := range []string{"csv_observations", "csv_attempts"} {
		var expected string
		if err := json.Unmarshal(fixture[name], &expected); err != nil {
			t.Fatal(err)
		}
		stream, basis := "observations", "observed_at"
		ids := []string{"kw_abcdefghijklmnopqrstuvwx"}
		if name == "csv_attempts" {
			stream = "attempts"
			basis = "recorded_at"
			ids = nil
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertEqual(t, r.Method, "GET")
			assertEqual(t, r.URL.Query().Get("version"), "2")
			assertEqual(t, r.URL.Query().Get("format"), "csv")
			assertEqual(t, r.URL.Query().Get("cursor"), "")
			w.Header().Set("Content-Type", "text/csv")
			_, _ = w.Write([]byte(expected))
		}))
		result, err := newTestClient(t, server.URL+"/api/v1").ExportRankHistory(context.Background(), "prj_abcdefghijklmnopqrstuvwx", &ExportRankHistoryOptions{Version: "2", Format: RankHistoryExportFormatCSV, Stream: stream, Basis: basis, Range: "all", KeywordIDs: ids, From: "1970-01-01T00:00:00Z", To: "2026-10-12T12:00:00Z", Aggregation: "none"})
		server.Close()
		if err != nil {
			t.Fatal(err)
		}
		assertEqual(t, result.CSV, expected)
	}
}

func TestRankHistoryReportRejectsNumericCID(t *testing.T) {
	fixture := reportFixture(t)
	original := string(fixture["observations"])
	invalid := strings.Replace(original, `"18446744073709551615"`, `18446744073709551615`, 1)
	if invalid == original {
		t.Fatal("actual CID fixture missing")
	}
	var response RankHistoryExportResponse
	if err := json.Unmarshal([]byte(invalid), &response); err == nil {
		t.Fatal("numeric CID accepted")
	}
}

func TestRankHistoryFutureOutputMetadata(t *testing.T) {
	raw := string(reportFixture(t)["attempts"])
	for _, pair := range [][2]string{{`"basis": "recorded_at"`, `"basis": "future_time_basis"`}, {`"current_definition_as_of_registered_baseline"`, `"future_summary_basis"`}, {`"retained_only"`, `"future_retention"`}} {
		if !strings.Contains(raw, pair[0]) {
			t.Fatal("fixture metadata missing")
		}
		raw = strings.ReplaceAll(raw, pair[0], pair[1])
	}
	var response RankHistoryExportResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"future_time_basis", "future_summary_basis", "future_retention"} {
		if !strings.Contains(string(encoded), value) {
			t.Fatalf("lost %s", value)
		}
	}
	invalid := strings.Replace(raw, `"version": 2`, `"version": 3`, 1)
	if invalid == raw {
		t.Fatal("version discriminator missing")
	}
	if err := json.Unmarshal([]byte(invalid), &response); err == nil {
		t.Fatal("unsupported version accepted")
	}
}

func TestRankHistoryReportRejectsExplicitNullKind(t *testing.T) {
	fixture := reportFixture(t)
	var page map[string]json.RawMessage
	if err := json.Unmarshal(fixture["observations"], &page); err != nil {
		t.Fatal(err)
	}
	var records []map[string]json.RawMessage
	if err := json.Unmarshal(page["data"], &records); err != nil {
		t.Fatal(err)
	}
	records[0]["kind"] = json.RawMessage(`null`)
	raw, err := json.Marshal(records[0])
	if err != nil {
		t.Fatal(err)
	}
	var record RankHistoryReportRecord
	if json.Unmarshal(raw, &record) == nil {
		t.Fatal("explicit null kind was accepted as an observation")
	}
}
