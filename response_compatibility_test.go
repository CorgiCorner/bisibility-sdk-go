package bisibility

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestApplicationResponseCompatibility(t *testing.T) {
	content, err := os.ReadFile("testdata/response-compatibility.json")
	if err != nil {
		t.Fatal(err)
	}
	type response struct {
		Status int             `json:"status"`
		Body   json.RawMessage `json:"body"`
	}
	var fixture struct {
		Full    response `json:"full_backlinks"`
		Partial response `json:"partial_backlinks"`
		Page    response `json:"page_backlinks"`
		Failed  response `json:"failed_summary"`
		Ranks   []struct {
			Check   json.RawMessage `json:"check"`
			Keyword json.RawMessage `json:"keyword"`
		} `json:"ranks"`
		FailedKeyword     json.RawMessage `json:"failed_keyword"`
		UnobservedKeyword json.RawMessage `json:"unobserved_keyword"`
	}
	if err := json.Unmarshal(content, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, source := range []response{fixture.Full, fixture.Partial, fixture.Page} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(source.Status)
			_, _ = w.Write(source.Body)
		}))
		result, err := newTestClient(t, server.URL+"/api/v1").AnalyzeBacklinks(context.Background(), "prj_a00000000000000000000000", AnalyzeBacklinksOptions{Target: "example.com"})
		server.Close()
		if err != nil {
			t.Fatal(err)
		}
		if result.Data.Snapshot == nil || result.Data.Estimate != nil {
			t.Fatal("expected successful snapshot")
		}
		snapshot := result.Data.Snapshot
		if snapshot.Summary.BacklinksTotal != 12 || snapshot.CostCents != 0.3 {
			t.Fatal("summary or cost evidence lost")
		}
		if snapshot.TargetScope == BacklinkTargetScopeSite && snapshot.HistoryUnavailable != (len(snapshot.History) == 0) {
			t.Fatal("history availability lost")
		}
		encoded, err := json.Marshal(snapshot)
		if err != nil {
			t.Fatal(err)
		}
		var roundtrip BacklinksSnapshot
		if err := json.Unmarshal(encoded, &roundtrip); err != nil {
			t.Fatal(err)
		}
		if roundtrip.HistoryUnavailable != snapshot.HistoryUnavailable || !reflect.DeepEqual(roundtrip.History, snapshot.History) {
			t.Fatal("round trip lost history evidence")
		}
	}
	for _, source := range fixture.Ranks {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.URL.Path == "/api/v1/rank-checks/check_a00000000000000000000000" {
				_, _ = w.Write(source.Check)
			} else {
				_, _ = w.Write(source.Keyword)
			}
		}))
		client := newTestClient(t, server.URL+"/api/v1")
		check, checkErr := client.GetRankCheckResult(context.Background(), "check_a00000000000000000000000")
		keyword, keywordErr := client.GetKeyword(context.Background(), "kw_a00000000000000000000000")
		server.Close()
		if checkErr != nil || keywordErr != nil {
			t.Fatalf("response decode failed: %v %v", checkErr, keywordErr)
		}
		if keyword.LatestCheck == nil || keyword.LatestSuccessfulCheck == nil {
			t.Fatal("latest checks lost")
		}
		if !reflect.DeepEqual(check.ObservationCompleteness, keyword.LatestCheck.ObservationCompleteness) || !reflect.DeepEqual(check.Position, keyword.LatestSuccessfulCheck.Position) {
			t.Fatal("coverage or position evidence changed")
		}
		assertResponseFieldsRoundTrip(t, source.Check, check, "observation_completeness", "position")
		assertResponseFieldsRoundTrip(t, source.Keyword, keyword, "latest_check", "latest_successful_check", "latest_position")
	}
	var failedKeyword Keyword
	var unobserved Keyword
	if err := json.Unmarshal(fixture.UnobservedKeyword, &unobserved); err != nil {
		t.Fatal(err)
	}
	if unobserved.LatestCheck != nil || unobserved.LatestSuccessfulCheck != nil || unobserved.LatestPosition != nil {
		t.Fatal("unobserved keyword acquired an observation")
	}
	assertResponseFieldsRoundTrip(t, fixture.UnobservedKeyword, unobserved, "latest_check", "latest_successful_check", "latest_position")
	if err := json.Unmarshal(fixture.FailedKeyword, &failedKeyword); err != nil {
		t.Fatal(err)
	}
	if failedKeyword.LatestCheck == nil || failedKeyword.LatestSuccessfulCheck == nil {
		t.Fatal("latest failed or successful check lost")
	}
	if failedKeyword.LatestPosition != nil || failedKeyword.LatestCheck.Status != "failed" || failedKeyword.LatestSuccessfulCheck.Position == nil || *failedKeyword.LatestSuccessfulCheck.Position != 6 {
		t.Fatal("new failure replaced successful observation")
	}
	if *failedKeyword.LatestSuccessfulCheck.ObservationCompleteness != ObservationTruncatedByStopOnMatch {
		t.Fatal("prior coverage lost")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(fixture.Failed.Status)
		_, _ = w.Write(fixture.Failed.Body)
	}))
	defer server.Close()
	result, err := newTestClient(t, server.URL+"/api/v1").AnalyzeBacklinks(context.Background(), "prj_a00000000000000000000000", AnalyzeBacklinksOptions{Target: "example.com"})
	var apiError *APIError
	if result != nil || !errors.As(err, &apiError) || apiError.StatusCode != 500 {
		t.Fatalf("failed summary became success: %v", err)
	}
	if apiError.Problem == nil {
		t.Fatal("problem evidence lost")
	}
	var details map[string]any
	if err := json.Unmarshal(apiError.Problem.Extensions["details"], &details); err != nil {
		t.Fatal(err)
	}
	if details["cost_cents"] != nil || details["known_summary_cost_cents"] != 0.2 {
		t.Fatal("unknown total or known summary cost changed")
	}
}

func assertResponseFieldsRoundTrip(t *testing.T, source json.RawMessage, value any, fields ...string) {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var before, after map[string]any
	if err := json.Unmarshal(source, &before); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &after); err != nil {
		t.Fatal(err)
	}
	for _, field := range fields {
		left, leftObject := before[field].(map[string]any)
		right, rightObject := after[field].(map[string]any)
		if leftObject && rightObject {
			leftDate, leftErr := time.Parse(time.RFC3339Nano, left["checked_at"].(string))
			rightDate, rightErr := time.Parse(time.RFC3339Nano, right["checked_at"].(string))
			if leftErr != nil || rightErr != nil || !leftDate.Equal(rightDate) {
				t.Fatal("check timestamp changed")
			}
			delete(left, "checked_at")
			delete(right, "checked_at")
		}
		if !reflect.DeepEqual(before[field], after[field]) {
			t.Fatalf("%s lost on round trip: %v vs %v", field, before[field], after[field])
		}
	}
}
