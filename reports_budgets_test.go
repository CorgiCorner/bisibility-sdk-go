package bisibility

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
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

const researchWorkspaceFixtures = `{"report": {"id": "agr_a00000000000000000000000", "kind": "external_review", "title": "Review", "created_at": "2026-10-07T00:00:00Z", "body": {"CamelCase": {"keyword_id": "producer-defined"}}, "provenance": {"SourceUrl": "https://example.org/source"}}, "summary": {"id": "agr_a00000000000000000000000", "kind": "external_review", "title": "Review", "created_at": "2026-10-07T00:00:00Z"}, "context": {"business": "Example", "audience": "Developers", "products": "API", "goals": "Quality", "agent_rules": "Use sources", "updated_at": null}, "site": {"id": "agr_a00000000000000000000000", "created_at": "2026-10-07T00:00:00Z", "cached": false, "result": {"version": 1, "target": "https://example.com", "started_at": "2026-10-07T00:00:00Z", "completed_at": "2026-10-07T00:00:00Z", "state": "complete", "stop_reason": "finished", "limits": {"max_pages": 10, "max_requests": 20, "max_duration_ms": 1000, "max_page_bytes": 1048576}, "requests": 1, "pages": [{"url": "https://example.com", "final_url": "https://example.com", "status": 200, "response_time_ms": 1, "title": "Example", "description": null, "canonical": null, "headings": [{"level": 1, "text": "Example"}], "h1_count": 1, "indexable": true, "robots": null, "internal_link_count": 0, "external_link_count": 0, "internal_links": [], "image_count": 0, "missing_alt_count": 0, "issues": [{"code": "description_missing", "message": "Missing description", "severity": "warning"}]}], "summary": {"pages": 1, "errors": 0, "warnings": 1, "indexable": 1}, "limitations": ["Bounded crawl"]}}, "estimate": {"ok": true, "estimate": true, "estimated_cost_cents": 1.25, "evidence": "observed_dataset"}, "ai": {"ok": true, "estimate": false, "cached": false, "report_id": "agr_a00000000000000000000000", "cost_cents": 1.25, "result": {"evidence": "synthetic_prompt_test", "rows": [{"prompt": "Example", "model": "gpt-4.1-mini", "answer": "Example API", "observed_at": null, "brand_mentioned": true, "domain_cited": true, "citations": [{"title": "Example", "url": "https://example.com", "target_domain": true}], "content_truncated": false}], "total_available": null, "truncated": false, "fetched_at": "2026-10-07T00:00:00Z", "cost_cents": 1.25, "cost_status": "confirmed", "failure": null}}}`

func TestResearchWorkspaceOperations(t *testing.T) {
	var fixtures map[string]json.RawMessage
	if err := json.Unmarshal([]byte(researchWorkspaceFixtures), &fixtures); err != nil {
		t.Fatal(err)
	}
	projectID := "prj_a00000000000000000000000"
	reportID := "agr_a00000000000000000000000"
	for _, test := range []struct {
		name, method, path, fixture string
		call                        func(*Client) (any, error)
	}{
		{"context get", "GET", "context", "context", func(c *Client) (any, error) { return c.GetProjectContext(context.Background(), projectID) }},
		{"context update", "PATCH", "context", "context", func(c *Client) (any, error) {
			return c.UpdateProjectContext(context.Background(), projectID, ProjectContextInput{Business: "Example"})
		}},
		{"report list", "GET", "agent-reports", "summary", func(c *Client) (any, error) {
			return c.ListAgentReports(context.Background(), projectID, &ListAgentReportsOptions{Kind: "external_review", Limit: 10})
		}},
		{"report create", "POST", "agent-reports", "report", func(c *Client) (any, error) {
			return c.CreateAgentReport(context.Background(), projectID, CreateAgentReportInput{Kind: "external_review", Title: "Review", Body: map[string]any{"CamelCase": true}})
		}},
		{"report get", "GET", "agent-reports/" + reportID, "report", func(c *Client) (any, error) { return c.GetAgentReport(context.Background(), projectID, reportID) }},
		{"audit list", "GET", "site-audits", "summary", func(c *Client) (any, error) { return c.ListSiteAudits(context.Background(), projectID) }},
		{"audit run", "POST", "site-audits", "site", func(c *Client) (any, error) {
			return c.RunSiteAudit(context.Background(), projectID, &RunSiteAuditOptions{MaxPages: 2})
		}},
		{"audit get", "GET", "site-audits/" + reportID, "site", func(c *Client) (any, error) { return c.GetSiteAudit(context.Background(), projectID, reportID) }},
		{"visibility", "POST", "ai-visibility", "estimate", func(c *Client) (any, error) {
			return c.AnalyzeAIVisibility(context.Background(), projectID, AnalyzeAIVisibilityOptions{AIResearchInput: AIResearchInput{Brand: "Example", Domain: "example.com", MaxCostCents: 0, EstimateOnly: true}})
		}},
		{"prompt", "POST", "prompt-explorer", "ai", func(c *Client) (any, error) {
			return c.CompareAIPrompts(context.Background(), projectID, CompareAIPromptsOptions{AIResearchInput: AIResearchInput{Brand: "Example", Domain: "example.com", MaxCostCents: 2}, Prompt: "Example", Models: []string{"gpt-4.1-mini"}})
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			client, err := NewClient(WithBaseURL("https://api.example.com/api/v1"), WithAPIKey(testAPIKey), WithHTTPClient(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				assertEqual(t, r.Method, test.method)
				assertEqual(t, r.URL.Path, "/api/v1/projects/"+projectID+"/"+test.path)
				if test.method != "GET" {
					body, err := io.ReadAll(r.Body)
					if err != nil {
						t.Fatal(err)
					}
					var value map[string]any
					if err := json.Unmarshal(body, &value); err != nil {
						t.Fatal(err)
					}
					if strings.HasPrefix(test.path, "ai-") || test.path == "prompt-explorer" {
						if _, present := value["max_cost_cents"]; !present {
							t.Fatal("missing explicit cost cap")
						}
					}
					if test.name == "report create" {
						if value["body"].(map[string]any)["CamelCase"] != true {
							t.Fatal("producer JSON changed")
						}
					}
				}
				data := string(fixtures[test.fixture])
				if strings.HasSuffix(test.name, "list") {
					data = "[" + data + "]"
				}
				body := `{"data":` + data + `}`
				if test.name == "report list" {
					body = `{"data":` + data + `,"meta":{"next_cursor":null}}`
				}
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
			})}))
			if err != nil {
				t.Fatal(err)
			}
			result, err := test.call(client)
			if err != nil {
				t.Fatal(err)
			}
			if result == nil {
				t.Fatal("nil result")
			}
			assertEqual(t, calls, 1)
			switch value := result.(type) {
			case *DataResponse[SiteAuditReport]:
				assertEqual(t, value.Data.Result.Pages[0].Headings[0].Text, "Example")
			case *DataResponse[AIAnalysisOutcome]:
				if value.Data.Estimate {
					if value.Data.EstimatedCostCents == nil {
						t.Fatal("missing estimate")
					}
				} else {
					assertEqual(t, value.Data.Result.Rows[0].Citations[0].URL, "https://example.com")
				}
			case *DataResponse[AgentReportResource]:
				if _, ok := value.Data.Body["CamelCase"]; !ok {
					t.Fatal("producer JSON changed")
				}
			}
		})
	}
}

func TestResearchWorkspaceRejectsMalformedContracts(t *testing.T) {
	client, err := NewClient(WithBaseURL("https://api.example.com/api/v1"), WithAPIKey(testAPIKey), WithHTTPClient(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		t.Fatal("invalid input reached transport")
		return nil, nil
	})}))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetAgentReport(context.Background(), "prj_a00000000000000000000000", "kw_a00000000000000000000000"); err == nil {
		t.Fatal("wrong report prefix accepted")
	}
	if _, err := client.RunSiteAudit(context.Background(), "prj_a00000000000000000000000", &RunSiteAuditOptions{MaxPages: 16}); err == nil {
		t.Fatal("unbounded crawl accepted")
	}
	if _, err := client.CreateAgentReport(context.Background(), "prj_a00000000000000000000000", CreateAgentReportInput{Kind: "site_audit", Title: "Report", Body: map[string]any{}}); err == nil {
		t.Fatal("reserved producer accepted")
	}
	var outcome AIAnalysisOutcome
	if err := json.Unmarshal([]byte(`{"ok":true,"estimate":false,"cached":false}`), &outcome); err == nil {
		t.Fatal("incomplete report accepted")
	}
}

func TestAgentReportPagerPreservesFilters(t *testing.T) {
	cursor := base64.RawURLEncoding.EncodeToString([]byte(`{"v":3,"public_id":"agr_a00000000000000000000000","t":"2026-10-07T00:00:00Z"}`))
	calls := 0
	client, err := NewClient(WithBaseURL("https://api.example.com/api/v1"), WithAPIKey(testAPIKey), WithHTTPClient(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		assertEqual(t, r.URL.Query().Get("kind"), "external_review")
		assertEqual(t, r.URL.Query().Get("limit"), "1")
		next := `null`
		if calls == 1 {
			assertEqual(t, r.URL.Query().Get("cursor"), "")
			next = `"` + cursor + `"`
		} else {
			assertEqual(t, r.URL.Query().Get("cursor"), cursor)
		}
		body := `{"data":[{"id":"agr_a00000000000000000000000","title":"Review","kind":"external_review","created_at":"2026-10-07T00:00:00Z"}],"meta":{"next_cursor":` + next + `}}`
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})}))
	if err != nil {
		t.Fatal(err)
	}
	pager := client.IterateAgentReports(context.Background(), "prj_a00000000000000000000000", &ListAgentReportsOptions{Kind: "external_review", Limit: 1})
	count := 0
	for pager.Next() {
		count++
		assertEqual(t, pager.Item().Kind, "external_review")
	}
	if err := pager.Err(); err != nil {
		t.Fatal(err)
	}
	assertEqual(t, count, 2)
	assertEqual(t, calls, 2)
}

func TestAIAnalysisRequiresBooleanEstimate(t *testing.T) {
	var fixtures map[string]json.RawMessage
	if err := json.Unmarshal([]byte(researchWorkspaceFixtures), &fixtures); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name, fixture, discriminator string
		valid                        bool
	}{
		{"null", "ai", "null", false},
		{"missing", "ai", "", false},
		{"string", "ai", `"false"`, false},
		{"number", "ai", "0", false},
		{"report false", "ai", "false", true},
		{"estimate true", "estimate", "true", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(fixtures[test.fixture], &fields); err != nil {
				t.Fatal(err)
			}
			if test.discriminator == "" {
				delete(fields, "estimate")
			} else {
				fields["estimate"] = json.RawMessage(test.discriminator)
			}
			data, err := json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}
			var outcome AIAnalysisOutcome
			err = json.Unmarshal(data, &outcome)
			if test.valid && err != nil {
				t.Fatalf("valid discriminator rejected: %v", err)
			}
			if !test.valid && err == nil {
				t.Fatal("invalid discriminator accepted")
			}
		})
	}
}
