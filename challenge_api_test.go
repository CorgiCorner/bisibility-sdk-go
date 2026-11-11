package bisibility

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func challengeFixtures(t *testing.T) map[string]json.RawMessage {
	t.Helper()
	b, err := os.ReadFile("testdata/challenge-api.json")
	if err != nil {
		t.Fatal(err)
	}
	var values map[string]json.RawMessage
	if err := json.Unmarshal(b, &values); err != nil {
		t.Fatal(err)
	}
	return values
}

func TestChallengeOperations(t *testing.T) {
	ctx := context.Background()
	project := "prj_a00000000000000000000000"
	report := "agr_a00000000000000000000000"
	token := "dpt_a00000000000000000000000"
	fixtures := challengeFixtures(t)
	for _, tt := range []struct {
		name, method, path, query, body string
		call                            func(*Client) (any, error)
	}{
		{"getDomainVerification", "GET", "/projects/" + project + "/domain-verification", "", "", func(c *Client) (any, error) { return c.GetDomainVerification(ctx, project) }},
		{"issueDomainVerification", "POST", "/projects/" + project + "/domain-verification", "", `{"method":"dns_txt"}`, func(c *Client) (any, error) {
			return c.IssueDomainVerification(ctx, project, DomainVerificationInput{Method: DomainVerificationMethodDNSTXT}, WithIdempotencyKey("challenge"))
		}},
		{"checkDomainVerification", "POST", "/projects/" + project + "/domain-verification/check", "", "", func(c *Client) (any, error) { return c.CheckDomainVerification(ctx, project) }},
		{"getSiteAuditRun", "GET", "/projects/" + project + "/site-audits/" + report + "/run", "", "", func(c *Client) (any, error) { return c.GetSiteAuditRun(ctx, project, report) }},
		{"cancelSiteAuditRun", "POST", "/projects/" + project + "/site-audits/" + report + "/cancel", "", "", func(c *Client) (any, error) { return c.CancelSiteAuditRun(ctx, project, report) }},
		{"resumeSiteAuditRun", "POST", "/projects/" + project + "/site-audits/" + report + "/resume", "", "", func(c *Client) (any, error) { return c.ResumeSiteAuditRun(ctx, project, report) }},
		{"listSiteAuditPages", "GET", "/projects/" + project + "/site-audits/" + report + "/pages", "limit=100&offset=15", "", func(c *Client) (any, error) {
			return c.ListSiteAuditPages(ctx, project, report, &ListSiteAuditPagesOptions{Limit: 100, Offset: 15})
		}},
		{"deleteMe", "DELETE", "/me", "", `{"confirm_email":"owner@example.com"}`, func(c *Client) (any, error) { return c.DeleteMe(ctx, DeleteMeInput{ConfirmEmail: "owner@example.com"}) }},
		{"listDevicePushTokens", "GET", "/me/push-tokens", "", "", func(c *Client) (any, error) { return c.ListDevicePushTokens(ctx) }},
		{"registerDevicePushToken", "POST", "/me/push-tokens", "", `{"token":"00000000000000000000000000000000","environment":"sandbox","bundle_id":"com.example.app"}`, func(c *Client) (any, error) {
			return c.RegisterDevicePushToken(ctx, RegisterDevicePushTokenInput{Token: "00000000000000000000000000000000", Environment: DevicePushEnvironmentSandbox, BundleID: "com.example.app"})
		}},
		{"deleteDevicePushToken", "DELETE", "/me/push-tokens/" + token, "", "", func(c *Client) (any, error) { return c.DeleteDevicePushToken(ctx, token) }},
		{"getSearchPerformanceTotals", "GET", "/projects/" + project + "/analytics/totals", "compare_previous=true&end_date=2026-11-10&start_date=2026-11-01", "", func(c *Client) (any, error) {
			return c.GetSearchPerformanceTotals(ctx, project, SearchPerformanceTotalsOptions{StartDate: "2026-11-01", EndDate: "2026-11-10", ComparePrevious: true})
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			c, err := NewClient(WithAPIKey(testPATKey), WithBaseURL("https://api.example.com/api/v1"), WithHTTPClient(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				assertEqual(t, r.Method, tt.method)
				assertEqual(t, r.URL.Path, "/api/v1"+tt.path)
				assertEqual(t, r.URL.RawQuery, tt.query)
				assertEqual(t, r.Header.Get("Authorization"), "Bearer "+testPATKey)
				if tt.name == "issueDomainVerification" {
					assertEqual(t, r.Header.Get("Idempotency-Key"), "challenge")
				}
				var body []byte
				if r.Body != nil {
					body, _ = io.ReadAll(r.Body)
				}
				if tt.body != "" {
					assertJSONEqual(t, string(body), tt.body)
				} else {
					assertEqual(t, string(body), "")
				}
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(string(fixtures[tt.name]))), Request: r}, nil
			})}))
			if err != nil {
				t.Fatal(err)
			}
			value, err := tt.call(c)
			if err != nil {
				t.Fatal(err)
			}
			assertEqual(t, calls, 1)
			b, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			assertJSONEqual(t, string(b), string(fixtures[tt.name]))
		})
	}
}

func TestChallengeSavedViews(t *testing.T) {
	fixtures := challengeFixtures(t)
	for _, name := range []string{"savedV1", "savedV2"} {
		for _, shown := range []bool{false, true} {
			var raw map[string]any
			json.Unmarshal(fixtures[name], &raw)
			raw["definition"].(map[string]any)["filters"].(map[string]any)["ai_overview_shown"] = shown
			payload, _ := json.Marshal(raw)
			var view SavedView
			if err := json.Unmarshal(payload, &view); err != nil {
				t.Fatal(err)
			}
			if view.Definition == nil || view.Definition.Keywords == nil {
				t.Fatal("missing keyword definition")
			}
			assertEqual(t, view.Definition.Keywords.Filters.AIOverviewShown, shown)
			if name == "savedV2" {
				assertEqual(t, view.Definition.Keywords.ReportScope.View, SavedViewReportAIOverviews)
			}
			b, err := json.Marshal(view)
			if err != nil {
				t.Fatal(err)
			}
			assertJSONEqual(t, string(b), string(payload))
		}
	}
	var legacy SavedView
	if err := json.Unmarshal([]byte(`{"id":"viw_a00000000000000000000000","config":{"filters":{}},"name":"Legacy","created_at":"2026-11-11T12:00:00Z","created_by_id":null}`), &legacy); err != nil {
		t.Fatal(err)
	}
	if legacy.Definition != nil {
		t.Fatal("legacy response gained definition")
	}
	for _, name := range []string{"savedV1", "savedV2"} {
		var raw map[string]any
		json.Unmarshal(fixtures[name], &raw)
		delete(raw["definition"].(map[string]any)["filters"].(map[string]any), "ai_overview_shown")
		b, _ := json.Marshal(raw)
		var view SavedView
		if err := json.Unmarshal(b, &view); err != nil {
			t.Fatal(err)
		}
		if view.Definition.Keywords.Filters.AIOverviewShown {
			t.Fatal("missing additive field must default false")
		}
	}
}

func TestChallengePageSpeed(t *testing.T) {
	ctx := context.Background()
	project := "prj_a00000000000000000000000"
	for _, tt := range []struct {
		method, suffix string
		call           func(*Client) (any, error)
	}{
		{"POST", "/connect", func(c *Client) (any, error) {
			return c.ConnectProvider(ctx, project, ProviderIDPageSpeed, ConnectProviderInput{})
		}},
		{"POST", "/test", func(c *Client) (any, error) {
			return c.TestProviderConnection(ctx, project, ProviderIDPageSpeed, TestProviderConnectionInput{})
		}},
		{"PATCH", "", func(c *Client) (any, error) {
			return c.UpdateProviderSettings(ctx, project, ProviderIDPageSpeed, ProviderSettingsInput{})
		}},
		{"DELETE", "", func(c *Client) (any, error) { return c.DisconnectProvider(ctx, project, ProviderIDPageSpeed) }},
		{"PATCH", "/budgets", func(c *Client) (any, error) {
			return c.UpdateProviderBudgets(ctx, project, ProviderIDPageSpeed, ProviderBudgetsUpdate{})
		}},
	} {
		t.Run(tt.method+tt.suffix, func(t *testing.T) {
			c, err := NewClient(WithAPIKey(testAPIKey), WithBaseURL("https://api.example.com/api/v1"), WithHTTPClient(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				assertEqual(t, r.Method, tt.method)
				assertEqual(t, r.URL.Path, "/api/v1/projects/"+project+"/providers/pagespeed"+tt.suffix)
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"id":"conn_a00000000000000000000000","project_id":"prj_a00000000000000000000000","provider":"pagespeed"}`)), Request: r}, nil
			})}))
			if err != nil {
				t.Fatal(err)
			}
			if _, err = tt.call(c); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestChallengeSavedViewEndpoints(t *testing.T) {
	for _, name := range []string{"savedV1", "savedV2"} {
		for _, method := range []string{"GET", "POST"} {
			t.Run(name+method, func(t *testing.T) {
				fixture := challengeFixtures(t)[name]
				c, err := NewClient(WithAPIKey(testAPIKey), WithBaseURL("https://api.example.com/api/v1"), WithHTTPClient(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
					assertEqual(t, r.Method, method)
					assertEqual(t, r.URL.Path, "/api/v1/projects/prj_a00000000000000000000000/saved-views")
					body := string(fixture)
					if method == "GET" {
						body = `{"data":[` + body + `],"meta":{"next_cursor":null}}`
					} else {
						b, _ := io.ReadAll(r.Body)
						assertJSONEqual(t, string(b), `{"name":"Example","config":{"filters":{"ai_overview_shown":false}}}`)
					}
					return &http.Response{StatusCode: map[string]int{"GET": 200, "POST": 201}[method], Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
				})}))
				if err != nil {
					t.Fatal(err)
				}
				var view *SavedView
				if method == "GET" {
					response, e := c.ListSavedViews(context.Background(), "prj_a00000000000000000000000", nil)
					if e != nil {
						t.Fatal(e)
					}
					view = &response.Data[0]
				} else {
					shown := false
					view, err = c.CreateSavedView(context.Background(), "prj_a00000000000000000000000", CreateSavedViewInput{Name: "Example", Config: SavedViewConfig{Filters: SavedViewFilters{AIOverviewShown: &shown}}})
					if err != nil {
						t.Fatal(err)
					}
				}
				if view.Definition == nil || view.Definition.Keywords == nil || view.Definition.Keywords.Filters.AIOverviewShown {
					t.Fatal("normalized false filter lost")
				}
				if name == "savedV2" {
					assertEqual(t, view.Definition.Keywords.ReportScope.View, SavedViewReportAIOverviews)
				}
			})
		}
	}
}

func TestChallengeInputBoundaries(t *testing.T) {
	c, err := NewClient(WithAPIKey(testPATKey), WithBaseURL("https://api.example.com/api/v1"), WithHTTPClient(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		t.Fatal("invalid input reached transport")
		return nil, nil
	})}))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	project := "prj_a00000000000000000000000"
	report := "agr_a00000000000000000000000"
	for _, call := range []func() error{
		func() error { _, e := c.GetDomainVerification(ctx, "invalid"); return e },
		func() error {
			_, e := c.IssueDomainVerification(ctx, project, DomainVerificationInput{Method: "invalid"})
			return e
		},
		func() error { _, e := c.GetSiteAuditRun(ctx, project, "kw_a00000000000000000000000"); return e },
		func() error { _, e := c.CancelSiteAuditRun(ctx, project, "invalid"); return e },
		func() error { _, e := c.ResumeSiteAuditRun(ctx, project, "invalid"); return e },
		func() error {
			_, e := c.ListSiteAuditPages(ctx, project, report, &ListSiteAuditPagesOptions{Offset: -1})
			return e
		},
		func() error {
			_, e := c.ListSiteAuditPages(ctx, project, report, &ListSiteAuditPagesOptions{Limit: 101})
			return e
		},
		func() error { _, e := c.DeleteDevicePushToken(ctx, "usr_a00000000000000000000000"); return e },
	} {
		if e := call(); e == nil {
			t.Fatal("invalid input accepted")
		}
	}
	var response DomainVerificationResponse
	if err := json.Unmarshal([]byte(`{"data":null}`), &response); err != nil || response.Data != nil {
		t.Fatalf("nullable proof: %v", err)
	}
}
