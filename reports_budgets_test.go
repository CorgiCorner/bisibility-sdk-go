package bisibility

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestReportsAndBudgets(t *testing.T) {
	requests := []capturedRequest{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		requests = append(requests, capturedRequest{Method: r.Method, Path: r.URL.Path, Query: r.URL.Query(), Body: string(data)})
		if r.Method == http.MethodPatch {
			writeJSON(t, w, 200, map[string]any{"provider": "dataforseo"})
			return
		}
		writeJSON(t, w, 200, map[string]any{"data": []any{}, "meta": map[string]any{"freshness_days": 30, "next_cursor": nil}})
	}))
	defer server.Close()
	c := newTestClient(t, server.URL+"/api/v1")
	ctx := context.Background()
	project := "prj_a00000000000000000000000"
	reports, err := c.ListStoredResearchReports(ctx, project)
	if err != nil || reports.Meta.FreshnessDays != 30 {
		t.Fatalf("reports: %v %v", reports, err)
	}
	if _, err = c.ListProviderBudgets(ctx, project); err != nil {
		t.Fatal(err)
	}
	if _, err = c.UpdateProviderBudgets(ctx, project, ProviderIDDataForSEO, ProviderBudgetsUpdate{Own: &ProviderBudgetPatch{App: ClearProviderBudget()}, Credits: &ProviderBudgetPatch{Programmatic: ProviderBudgetValue(ProviderMonthlyBudget{AmountPerMonth: 123, Unit: "cents"})}}); err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err = json.Unmarshal([]byte(requests[2].Body), &body); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"own": map[string]any{"app": nil}, "credits": map[string]any{"programmatic": map[string]any{"amount_per_month": float64(123), "unit": "cents"}}}
	if !reflect.DeepEqual(body, want) {
		t.Fatalf("body %v", body)
	}
	if requests[0].Method != "GET" || requests[1].Method != "GET" || requests[2].Method != "PATCH" {
		t.Fatal(requests)
	}
}
func TestMissingStoredReportNeverStartsProviderWork(t *testing.T) {
	count := 0
	no := false
	zero := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		if r.Method != "GET" || r.URL.Query().Get("include_subdomains") != "false" || r.URL.Query().Get("location_code") != "0" {
			t.Error(r.URL)
		}
		writeJSON(t, w, 404, map[string]any{"status": 404, "title": "Not found"})
	}))
	defer server.Close()
	c := newTestClient(t, server.URL+"/api/v1")
	_, err := c.GetStoredResearchReport(context.Background(), "prj_a00000000000000000000000", StoredReportBacklinks, StoredResearchReportOptions{Target: "example.com", IncludeSubdomains: &no, LocationCode: &zero})
	if err == nil || count != 1 {
		t.Fatalf("calls %d error %v", count, err)
	}
}
func TestStoredReportUnionAndBudgetValidation(t *testing.T) {
	for _, raw := range []string{`{"summary":{}}`, `{"data_state":"partial","state":"stale"}`, `{"request_key":"key","state":"fresh"}`} {
		var report StoredResearchReport
		if err := json.Unmarshal([]byte(raw), &report); err != nil {
			t.Fatal(err)
		}
		populated := 0
		if report.Backlinks != nil {
			populated++
		}
		if report.DomainOverview != nil {
			populated++
		}
		if report.KeywordResearch != nil {
			populated++
		}
		if populated != 1 {
			t.Fatal(report)
		}
	}
	var report StoredResearchReport
	if json.Unmarshal([]byte(`{}`), &report) == nil {
		t.Fatal("unknown shape accepted")
	}
	for _, b := range []ProviderMonthlyBudget{{AmountPerMonth: 0, Unit: "cents"}, {AmountPerMonth: 1, Unit: "units"}} {
		if (ProviderBudgetsUpdate{Credits: &ProviderBudgetPatch{App: ProviderBudgetValue(b)}}).validate() == nil {
			t.Fatal("invalid credit budget accepted")
		}
	}
}

func TestStoredReportFixtures(t *testing.T) {
	data, err := os.ReadFile("testdata/stored_reports.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures map[string]json.RawMessage
	if err = json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	for kind, data := range fixtures {
		var report StoredResearchReport
		if err = json.Unmarshal(data, &report); err != nil {
			t.Fatal(kind, err)
		}
		switch kind {
		case "backlinks":
			if report.Backlinks == nil || report.Backlinks.State != "fresh" {
				t.Fatal(report)
			}
		case "domain_overview":
			if report.DomainOverview == nil || report.DomainOverview.DataState != "no_data" {
				t.Fatal(report)
			}
		case "keyword_research":
			if report.KeywordResearch == nil || report.KeywordResearch.State != "fresh" {
				t.Fatal(report)
			}
		}
	}
}

func TestStoredKeywordIntentPreservesNull(t *testing.T) {
	var r StoredResearchReport
	if err := json.Unmarshal([]byte(`{"request_key":"saved","rows":[{"intent":null},{"intent":"commercial"}]}`), &r); err != nil {
		t.Fatal(err)
	}
	if r.KeywordResearch == nil || r.KeywordResearch.Rows[0].Intent != nil || r.KeywordResearch.Rows[1].Intent == nil || *r.KeywordResearch.Rows[1].Intent != "commercial" {
		t.Fatal(r)
	}
}
