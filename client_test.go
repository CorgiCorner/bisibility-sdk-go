package bisibility

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

const testAPIKey = "bsk_live_1234567890abcdef"

type capturedRequest struct {
	Body          string
	Header        http.Header
	Method        string
	Path          string
	Query         url.Values
	RawQuery      string
	ContentLength int64
}

type discoveryTestCase struct {
	name      string
	call      func(context.Context, *Client) (any, error)
	response  any
	path      string
	want      func(t *testing.T, got any)
	text      string
	textReply bool
}

type publicCostTestCase struct {
	name     string
	call     func(context.Context, *Client) (any, error)
	path     string
	query    string
	response any
	want     func(t *testing.T, got any)
}

type protectedMethodTestCase struct {
	name             string
	call             func(context.Context, *Client) (any, error)
	method           string
	path             string
	query            string
	body             string
	response         any
	status           int
	idempotencyKey   string
	wantContentType  bool
	wantNoBody       bool
	requestHeaderKey string
	requestHeaderVal string
	want             func(t *testing.T, got any)
}

func TestDiscoveryMethods(t *testing.T) {
	t.Parallel()

	health := healthFixture()
	capability := Capability{
		Name:        "addKeywords",
		OperationID: "addKeywords",
		Description: "Add one or more keywords",
		InputSchema: JSONValue{"type": "object"},
	}
	openapi := OpenAPIDocument{
		OpenAPI: "3.1.0",
		Info:    JSONValue{"title": "Bisibility"},
		Paths:   JSONValue{},
	}

	tests := []discoveryTestCase{
		{
			name:     "health",
			call:     func(ctx context.Context, c *Client) (any, error) { return c.GetHealth(ctx) },
			response: health,
			path:     "/api/v1/health",
			want: func(t *testing.T, got any) {
				t.Helper()
				if got.(*HealthResponse).Status != "ok" {
					t.Fatalf("status = %q, want ok", got.(*HealthResponse).Status)
				}
			},
		},
		{
			name:     "openapi",
			call:     func(ctx context.Context, c *Client) (any, error) { return c.GetOpenAPI(ctx) },
			response: openapi,
			path:     "/api/v1/openapi.json",
			want: func(t *testing.T, got any) {
				t.Helper()
				if got.(*OpenAPIDocument).OpenAPI != "3.1.0" {
					t.Fatalf("openapi = %q, want 3.1.0", got.(*OpenAPIDocument).OpenAPI)
				}
			},
		},
		{
			name:     "capabilities",
			call:     func(ctx context.Context, c *Client) (any, error) { return c.GetCapabilities(ctx) },
			response: DataResponse[[]Capability]{Data: []Capability{capability}},
			path:     "/api/v1/capabilities",
			want: func(t *testing.T, got any) {
				t.Helper()
				if got.(*DataResponse[[]Capability]).Data[0].OperationID != "addKeywords" {
					t.Fatalf("operation id mismatch")
				}
			},
		},
		{
			name: "llms",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.GetLLMSText(ctx)
			},
			path:      "/api/v1/llms.txt",
			text:      "# Bisibility API v1",
			textReply: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				if got.(string) != "# Bisibility API v1" {
					t.Fatalf("text = %q", got.(string))
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			runDiscoveryTestCase(t, tt)
		})
	}
}

func runDiscoveryTestCase(t *testing.T, tt discoveryTestCase) {
	t.Helper()
	var captured capturedRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = captureRequest(t, r)
		assertEqual(t, r.Header.Get("Authorization"), "")
		assertEqual(t, r.URL.Path, tt.path)
		if tt.textReply {
			w.Header().Set(contentTypeHeader, "text/plain; charset=utf-8")
			_, _ = w.Write([]byte(tt.text))
			return
		}
		writeJSON(t, w, http.StatusOK, tt.response)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	got, err := tt.call(context.Background(), client)
	if err != nil {
		t.Fatalf("call returned error: %v", err)
	}
	tt.want(t, got)
	assertEqual(t, captured.Method, http.MethodGet)
}

func TestPublicCostMethods(t *testing.T) {
	t.Parallel()

	tests := []publicCostTestCase{
		{
			name:     "provider rates",
			call:     func(ctx context.Context, c *Client) (any, error) { return c.GetProviderRates(ctx) },
			path:     "/api/v1/provider-rates",
			response: map[string]any{"data": []any{flatProviderRateJSON(), planProviderRateJSON()}},
			want: func(t *testing.T, got any) {
				t.Helper()
				rates := got.(*DataResponse[[]ProviderRate])
				flat := rates.Data[0]
				assertEqual(t, flat.ProviderID, ProviderIDDataForSEO)
				assertEqual(t, flat.PricingModel, PricingModelFlat)
				assertEqual(t, flat.CheckedAt, "2026-07-03")
				assertEqual(t, flat.SourceURL, "https://dataforseo.com/apis/serp-api/pricing")
				assertEqual(t, flat.Options[0].Key, "standard")
				assertEqual(t, flat.Options[0].ShortLabel, "Standard")
				assertEqual(t, flat.Options[0].UnitCostCents, 0.06)
				assertEqual(t, flat.Options[0].UnitCostUSD, 0.0006)
				assertEqual(t, len(flat.Plans), 0)
				plan := rates.Data[1]
				assertEqual(t, plan.ProviderID, ProviderIDSerpAPI)
				assertEqual(t, plan.PricingModel, PricingModelPlan)
				assertEqual(t, plan.Plans[0].PlanKey, "developer")
				assertEqual(t, plan.Plans[0].IncludedChecks, 5000)
				assertEqual(t, plan.Plans[0].MonthlyPriceCents, 7500.0)
				assertEqual(t, plan.Plans[0].MonthlyPriceUSD, 75.0)
				assertEqual(t, len(plan.Options), 0)
			},
		},
		{
			name: "cost estimate flat option",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.GetCostEstimate(ctx, CostEstimateOptions{
					Devices:   1,
					Frequency: EstimateFrequencyDaily,
					Keywords:  248,
					Locations: 1,
					Option:    "standard",
					Provider:  ProviderIDDataForSEO,
				})
			},
			path:     "/api/v1/cost-estimate",
			query:    "devices=1&frequency=daily&keywords=248&locations=1&option=standard&provider=dataforseo",
			response: map[string]any{"data": flatCostEstimateJSON()},
			want: func(t *testing.T, got any) {
				t.Helper()
				estimate := got.(*DataResponse[CostEstimate]).Data
				assertEqual(t, estimate.ChecksPerRun, 248)
				assertEqual(t, estimate.MonthlyChecks, 7440)
				assertEqual(t, estimate.MonthlyCostCents, 446.4)
				assertEqual(t, estimate.MonthlyCostUSD, 4.464)
				assertEqual(t, estimate.EffectiveCostPerCheckCents, 0.06)
				assertEqual(t, estimate.PricingModel, PricingModelFlat)
				assertEqual(t, estimate.ProviderID, ProviderIDDataForSEO)
				assertEqual(t, estimate.RateCheckedAt, "2026-07-03")
				assertEqual(t, estimate.SelectedOption.Key, "standard")
				if estimate.SelectedPlan != nil {
					t.Fatalf("SelectedPlan = %v, want nil", estimate.SelectedPlan)
				}
			},
		},
		{
			name: "cost estimate plan defaults",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.GetCostEstimate(ctx, CostEstimateOptions{
					Frequency: EstimateFrequencyMonthly,
					Keywords:  5000,
					Provider:  ProviderIDSerpAPI,
				})
			},
			path:     "/api/v1/cost-estimate",
			query:    "frequency=monthly&keywords=5000&provider=serpapi",
			response: map[string]any{"data": planCostEstimateJSON()},
			want: func(t *testing.T, got any) {
				t.Helper()
				estimate := got.(*DataResponse[CostEstimate]).Data
				assertEqual(t, estimate.ExceedsLargestPlan, false)
				assertEqual(t, estimate.ExceedsSelectedPlan, false)
				assertEqual(t, estimate.MonthlyChecks, 5000)
				assertEqual(t, estimate.MonthlyCostCents, 7500.0)
				assertEqual(t, estimate.PricingModel, PricingModelPlan)
				assertEqual(t, estimate.SelectedPlan.PlanKey, "developer")
				if estimate.SelectedOption != nil {
					t.Fatalf("SelectedOption = %v, want nil", estimate.SelectedOption)
				}
			},
		},
		{
			name: "cost estimate always sends keywords",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.GetCostEstimate(ctx, CostEstimateOptions{})
			},
			path:     "/api/v1/cost-estimate",
			query:    "keywords=0",
			response: map[string]any{"data": flatCostEstimateJSON()},
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*DataResponse[CostEstimate]).Data.ProviderID, ProviderIDDataForSEO)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			runPublicCostTestCase(t, tt)
		})
	}
}

func runPublicCostTestCase(t *testing.T, tt publicCostTestCase) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.Header.Get("Authorization"), "")
		assertEqual(t, r.Method, http.MethodGet)
		assertEqual(t, r.URL.Path, tt.path)
		assertEqual(t, r.URL.RawQuery, tt.query)
		writeJSON(t, w, http.StatusOK, tt.response)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	got, err := tt.call(context.Background(), client)
	if err != nil {
		t.Fatalf("call returned error: %v", err)
	}
	tt.want(t, got)
}

func TestAnonymousRequestStripsCallerAuthorizationHeader(t *testing.T) {
	t.Parallel()

	var captured capturedRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = captureRequest(t, r)
		writeJSON(t, w, http.StatusOK, map[string]any{"data": []any{flatProviderRateJSON()}})
	}))
	defer server.Close()

	client, err := NewClient(
		WithBaseURL(server.URL+"/api/v1"),
		WithDefaultHeader("Authorization", "Bearer x"),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	if _, err := client.GetProviderRates(context.Background()); err != nil {
		t.Fatalf("GetProviderRates returned error: %v", err)
	}
	if got, ok := captured.Header["Authorization"]; ok {
		t.Fatalf("Authorization header = %q, want none", got)
	}
}

func flatProviderRateJSON() map[string]any {
	return map[string]any{
		"checked_at": "2026-07-03",
		"label":      "DataForSEO",
		"notes":      "Pay-as-you-go; base price covers page 1.",
		"options": []any{
			map[string]any{
				"key":             "standard",
				"label":           "Standard queue",
				"short_label":     "Standard",
				"turnaround":      "~5 min",
				"unit_cost_cents": 0.06,
				"unit_cost_usd":   0.0006,
			},
		},
		"pricing_model": "flat",
		"provider_id":   "dataforseo",
		"source_url":    "https://dataforseo.com/apis/serp-api/pricing",
	}
}

func planProviderRateJSON() map[string]any {
	return map[string]any{
		"checked_at": "2026-07-04",
		"label":      "SerpAPI",
		"notes":      "Subscription plans; only successful searches count.",
		"plans": []any{
			map[string]any{
				"included_checks":     5000,
				"label":               "Developer",
				"monthly_price_cents": 7500,
				"monthly_price_usd":   75,
				"plan_key":            "developer",
			},
		},
		"pricing_model": "plan",
		"provider_id":   "serpapi",
		"source_url":    "https://serpapi.com/pricing",
	}
}

func flatCostEstimateJSON() map[string]any {
	return map[string]any{
		"checks_per_run":                 248,
		"effective_cost_per_check_cents": 0.06,
		"exceeds_largest_plan":           false,
		"exceeds_selected_plan":          false,
		"monthly_checks":                 7440,
		"monthly_cost_cents":             446.4,
		"monthly_cost_usd":               4.464,
		"pricing_model":                  "flat",
		"provider_id":                    "dataforseo",
		"rate_checked_at":                "2026-07-03",
		"rate_source_url":                "https://dataforseo.com/apis/serp-api/pricing",
		"selected_option": map[string]any{
			"key":             "standard",
			"label":           "Standard queue",
			"short_label":     "Standard",
			"turnaround":      "~5 min",
			"unit_cost_cents": 0.06,
			"unit_cost_usd":   0.0006,
		},
	}
}

func planCostEstimateJSON() map[string]any {
	return map[string]any{
		"checks_per_run":                 5000,
		"effective_cost_per_check_cents": 1.5,
		"exceeds_largest_plan":           false,
		"exceeds_selected_plan":          false,
		"monthly_checks":                 5000,
		"monthly_cost_cents":             7500,
		"monthly_cost_usd":               75,
		"pricing_model":                  "plan",
		"provider_id":                    "serpapi",
		"rate_checked_at":                "2026-07-04",
		"rate_source_url":                "https://serpapi.com/pricing",
		"selected_plan": map[string]any{
			"included_checks":     5000,
			"label":               "Developer",
			"monthly_price_cents": 7500,
			"monthly_price_usd":   75,
			"plan_key":            "developer",
		},
	}
}

func TestProtectedMethods(t *testing.T) {
	t.Parallel()

	createdAPIKey := CreatedAPIKey{
		APIKey:      apiKeyFixture("key_new"),
		MaskedValue: "bsk_live_12345678******cdef",
		Token:       testAPIKey,
	}
	keyword := keywordFixture("kw_1")
	check := rankCheckFixture("check_1")
	since := mustTime("2026-01-01T00:00:00Z")
	until := mustTime("2026-01-31T00:00:00Z")
	targetURL := "https://example.com/page"
	newKeywordText := "new text"
	tags := []string{"API"}
	frequency := RankCheckFrequencyDaily

	tests := []protectedMethodTestCase{
		{
			name:     "list projects",
			call:     func(ctx context.Context, c *Client) (any, error) { return c.ListProjects(ctx) },
			method:   http.MethodGet,
			path:     "/api/v1/projects",
			response: listEnvelope([]any{projectJSON("prj_1")}, ""),
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*ListResponse[Project]).Data[0].ID, "prj_1")
				assertEqual(t, got.(*ListResponse[Project]).Data[0].WriteMode, ProjectWriteModeActive)
			},
		},
		{
			name:     "projects alias",
			call:     func(ctx context.Context, c *Client) (any, error) { return c.Projects(ctx) },
			method:   http.MethodGet,
			path:     "/api/v1/projects",
			response: listResponse(projectFixture("prj_1")),
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*ListResponse[Project]).Data[0].Name, "Example")
			},
		},
		{
			name:     "get project escapes id",
			call:     func(ctx context.Context, c *Client) (any, error) { return c.GetProject(ctx, "prj spaced") },
			method:   http.MethodGet,
			path:     "/api/v1/projects/prj%20spaced",
			response: projectFixture("prj spaced"),
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*Project).ID, "prj spaced")
			},
		},
		{
			name: "update project",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.UpdateProject(ctx, "prj_1", UpdateProjectInput{
					Domain: strPtr("renamed.example"),
					Name:   strPtr("Renamed"),
				}, WithIdempotencyKey("idem_project"))
			},
			method:          http.MethodPatch,
			path:            "/api/v1/projects/prj_1",
			body:            `{"domain":"renamed.example","name":"Renamed"}`,
			response:        projectJSON("prj_1"),
			idempotencyKey:  "idem_project",
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*Project).ID, "prj_1")
				assertEqual(t, got.(*Project).WriteMode, ProjectWriteModeActive)
			},
		},
		{
			name: "update project name only",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.UpdateProject(ctx, "prj_1", UpdateProjectInput{Name: strPtr("Renamed")})
			},
			method:          http.MethodPatch,
			path:            "/api/v1/projects/prj_1",
			body:            `{"name":"Renamed"}`,
			response:        projectJSON("prj_1"),
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*Project).Name, "Example")
			},
		},
		{
			name: "delete project escapes id",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.DeleteProject(ctx, "prj spaced")
			},
			method:   http.MethodDelete,
			path:     "/api/v1/projects/prj%20spaced",
			response: projectJSON("prj spaced"),
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*Project).ID, "prj spaced")
			},
		},
		{
			name: "update project defaults",
			call: func(ctx context.Context, c *Client) (any, error) {
				autoSchedule := true
				jitter := 30
				return c.UpdateProjectDefaults(ctx, "prj_1", ProjectDefaultsPatch{
					AutoSchedule:  &autoSchedule,
					City:          strPtr("Austin"),
					Country:       "United States",
					Device:        DeviceMobile,
					Frequency:     RankCheckFrequencyDaily,
					JitterMinutes: &jitter,
					Timezone:      "America/Chicago",
				})
			},
			method:          http.MethodPatch,
			path:            "/api/v1/projects/prj_1/defaults",
			body:            `{"auto_schedule":true,"city":"Austin","country":"United States","device":"mobile","frequency":"daily","jitter_minutes":30,"timezone":"America/Chicago"}`,
			response:        projectDefaultsJSON("prj_1"),
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				defaults := got.(*ProjectDefaults)
				assertEqual(t, defaults.ProjectID, "prj_1")
				assertEqual(t, defaults.Country, "United States")
				assertEqual(t, *defaults.City, "Austin")
				assertEqual(t, defaults.Device, DeviceMobile)
				assertEqual(t, defaults.Frequency, RankCheckFrequencyDaily)
				assertEqual(t, defaults.LocationKey, "US/Texas/Austin")
				assertEqual(t, defaults.JitterMinutes, 30)
				if defaults.LastCheckedAt != nil {
					t.Fatalf("last_checked_at = %v, want nil", defaults.LastCheckedAt)
				}
			},
		},
		{
			name: "update project defaults by location key",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.UpdateProjectDefaults(ctx, "prj_1", ProjectDefaultsPatch{
					Frequency:   RankCheckFrequencyWeekly,
					LocationKey: "US/Texas/Austin",
				})
			},
			method:          http.MethodPatch,
			path:            "/api/v1/projects/prj_1/defaults",
			body:            `{"frequency":"weekly","location_key":"US/Texas/Austin"}`,
			response:        projectDefaultsJSON("prj_1"),
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*ProjectDefaults).LocationKey, "US/Texas/Austin")
			},
		},
		{
			name: "list api keys",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ListAPIKeys(ctx, &PaginationOptions{Cursor: "cursor 1", Limit: 10})
			},
			method: http.MethodGet,
			path:   "/api/v1/api-keys",
			query:  "cursor=cursor+1&limit=10",
			response: ListResponse[APIKey]{
				Data: []APIKey{apiKeyFixture("key_1")},
				Meta: ListMeta{NextCursor: strPtr("cursor_2")},
			},
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, *got.(*ListResponse[APIKey]).Meta.NextCursor, "cursor_2")
			},
		},
		{
			name: "create api key",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.CreateAPIKey(ctx, CreateAPIKeyInput{Name: "CI"}, WithIdempotencyKey("idem_1"))
			},
			method:          http.MethodPost,
			path:            "/api/v1/api-keys",
			body:            `{"name":"CI"}`,
			response:        createdAPIKey,
			status:          http.StatusCreated,
			idempotencyKey:  "idem_1",
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*CreatedAPIKey).Token, testAPIKey)
			},
		},
		{
			name:     "revoke api key",
			call:     func(ctx context.Context, c *Client) (any, error) { return c.RevokeAPIKey(ctx, "key_1") },
			method:   http.MethodDelete,
			path:     "/api/v1/api-keys/key_1",
			response: apiKeyFixture("key_1"),
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*APIKey).ID, "key_1")
			},
		},
		{
			name: "list keywords with filters",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ListKeywords(ctx, "prj_1", &ListKeywordsOptions{
					Country:    "United States",
					Cursor:     "cursor_1",
					Device:     DeviceDesktop,
					Intent:     "transactional",
					Limit:      25,
					PositionGT: 3,
					PositionLT: 10,
					Search:     "rank tracker",
					Sort:       "-updated_at",
					Tag:        "Product",
					Topic:      "Rank tracking",
				})
			},
			method:   http.MethodGet,
			path:     "/api/v1/projects/prj_1/keywords",
			query:    "cursor=cursor_1&filter%5Bcountry%5D=United+States&filter%5Bdevice%5D=desktop&filter%5Bintent%5D=transactional&filter%5Bposition_gt%5D=3&filter%5Bposition_lt%5D=10&filter%5Btag%5D=Product&filter%5Btopic%5D=Rank+tracking&limit=25&search=rank+tracker&sort=-updated_at",
			response: listResponse(keyword),
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*ListResponse[Keyword]).Data[0].Text, "rank tracker")
			},
		},
		{
			name: "keywords list alias",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.KeywordsList(ctx, "prj_1", &ListKeywordsOptions{Limit: 5})
			},
			method:   http.MethodGet,
			path:     "/api/v1/projects/prj_1/keywords",
			query:    "limit=5",
			response: listResponse(keyword),
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*ListResponse[Keyword]).Data[0].ID, "kw_1")
			},
		},
		{
			name: "create keywords",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.CreateKeywords(ctx, "prj_1", CreateKeywordsInput{Keywords: []CreateKeywordInput{{
					Keyword:   "rank tracker",
					Schedule:  &KeywordScheduleInput{CronExpression: nil, Frequency: RankCheckFrequencyDaily},
					Tags:      []string{"Product"},
					TargetURL: &targetURL,
				}}}, WithIdempotencyKey("idem_keywords"))
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects/prj_1/keywords",
			body:            `{"keywords":[{"keyword":"rank tracker","schedule":{"cronExpression":null,"frequency":"daily"},"tags":["Product"],"target_url":"https://example.com/page"}]}`,
			response:        createKeywordsResponse(keyword),
			status:          http.StatusCreated,
			idempotencyKey:  "idem_keywords",
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*CreateKeywordsResponse).Created, 1)
			},
		},
		{
			name: "keywords create alias",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.KeywordsCreate(ctx, "prj_1", CreateKeywordsInput{Keywords: []CreateKeywordInput{{Keyword: "rank tracker"}}})
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects/prj_1/keywords",
			body:            `{"keywords":[{"keyword":"rank tracker"}]}`,
			response:        createKeywordsResponse(keyword),
			status:          http.StatusCreated,
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*CreateKeywordsResponse).Results[0].Status, "created")
			},
		},
		{
			name: "add keywords array body",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.AddKeywords(ctx, "prj_1", []CreateKeywordInput{{Keyword: "rank tracker", Tags: []string{"Product"}}})
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects/prj_1/keywords",
			body:            `[{"keyword":"rank tracker","tags":["Product"]}]`,
			response:        createKeywordsResponse(keyword),
			status:          http.StatusCreated,
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*CreateKeywordsResponse).Skipped, 0)
			},
		},
		{
			name: "create keywords with location and intent fields",
			call: func(ctx context.Context, c *Client) (any, error) {
				intent := "commercial"
				topic := "tooling"
				return c.CreateKeywords(ctx, "prj_1", CreateKeywordsInput{Keywords: []CreateKeywordInput{{
					Keyword:     "rank tracker",
					City:        "Austin",
					Country:     "United States",
					LocationKey: "US/Texas/Austin",
					Intent:      &intent,
					Topic:       &topic,
				}}})
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects/prj_1/keywords",
			body:            `{"keywords":[{"keyword":"rank tracker","city":"Austin","country":"United States","location_key":"US/Texas/Austin","intent":"commercial","topic":"tooling"}]}`,
			response:        createKeywordsResponseJSON(),
			status:          http.StatusCreated,
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				resp := got.(*CreateKeywordsResponse)
				assertEqual(t, resp.Created, 1)
				assertEqual(t, resp.Results[0].Warning, "City not found; tracking at country level.")
				assertEqual(t, len(resp.Warnings), 1)
				assertEqual(t, resp.Warnings[0], "City not found; tracking at country level.")
				assertEqual(t, *resp.Results[0].Keyword.Intent, "commercial")
				assertEqual(t, *resp.Results[0].Keyword.Topic, "tooling")
			},
		},
		{
			name:     "get keyword",
			call:     func(ctx context.Context, c *Client) (any, error) { return c.GetKeyword(ctx, "kw_1") },
			method:   http.MethodGet,
			path:     "/api/v1/keywords/kw_1",
			response: keywordJSON("kw_1"),
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*Keyword).ID, "kw_1")
				assertEqual(t, *got.(*Keyword).Intent, "commercial")
				if got.(*Keyword).Topic != nil {
					t.Fatalf("topic = %v, want nil", *got.(*Keyword).Topic)
				}
			},
		},
		{
			name: "update keyword location intent and topic",
			call: func(ctx context.Context, c *Client) (any, error) {
				city := "Austin"
				locationKey := "US/Texas/Austin"
				return c.UpdateKeyword(ctx, "kw_1", UpdateKeywordInput{
					City:        &city,
					Intent:      NullString(),
					LocationKey: &locationKey,
					Topic:       StringValue("tooling"),
				})
			},
			method:          http.MethodPatch,
			path:            "/api/v1/keywords/kw_1",
			body:            `{"city":"Austin","intent":null,"location_key":"US/Texas/Austin","topic":"tooling"}`,
			response:        keywordJSON("kw_1"),
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*Keyword).ID, "kw_1")
			},
		},
		{
			name: "update keyword",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.UpdateKeyword(ctx, "kw_1", UpdateKeywordInput{Keyword: &newKeywordText, Tags: tags})
			},
			method:          http.MethodPatch,
			path:            "/api/v1/keywords/kw_1",
			body:            `{"keyword":"new text","tags":["API"]}`,
			response:        keywordFixture("kw_1"),
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*Keyword).ID, "kw_1")
			},
		},
		{
			name: "set keyword target url to null",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.SetKeywordTargetURL(ctx, "kw_1", nil)
			},
			method:          http.MethodPatch,
			path:            "/api/v1/keywords/kw_1",
			body:            `{"target_url":null}`,
			response:        keywordFixture("kw_1"),
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*Keyword).ID, "kw_1")
			},
		},
		{
			name:     "delete keyword",
			call:     func(ctx context.Context, c *Client) (any, error) { return c.DeleteKeyword(ctx, "kw_1") },
			method:   http.MethodDelete,
			path:     "/api/v1/keywords/kw_1",
			response: keyword,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*Keyword).Text, "rank tracker")
			},
		},
		{
			name: "bulk update keywords",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.BulkUpdateKeywords(ctx, KeywordBulkInput{KeywordIDs: []string{"kw_1"}, Operation: KeywordBulkOperationAddTags, Tags: []string{"Product"}})
			},
			method:          http.MethodPost,
			path:            "/api/v1/keywords/bulk",
			body:            `{"keyword_ids":["kw_1"],"operation":"add_tags","tags":["Product"]}`,
			response:        KeywordBulkResponse{Operation: KeywordBulkOperationAddTags, Results: []KeywordBulkItemResult{{KeywordID: "kw_1", Status: "updated"}}},
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*KeywordBulkResponse).Operation, KeywordBulkOperationAddTags)
			},
		},
		{
			name: "bulk set target url null",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.BulkUpdateKeywords(ctx, KeywordBulkInput{KeywordIDs: []string{"kw_1"}, Operation: KeywordBulkOperationSetTargetURL, TargetURL: NullString()})
			},
			method:          http.MethodPost,
			path:            "/api/v1/keywords/bulk",
			body:            `{"keyword_ids":["kw_1"],"operation":"set_target_url","target_url":null}`,
			response:        KeywordBulkResponse{Operation: KeywordBulkOperationSetTargetURL, Results: []KeywordBulkItemResult{{KeywordID: "kw_1", Status: "updated"}}},
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*KeywordBulkResponse).Results[0].Status, "updated")
			},
		},
		{
			name: "bulk set frequency",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.BulkUpdateKeywords(ctx, KeywordBulkInput{KeywordIDs: []string{"kw_1"}, Operation: KeywordBulkOperationSetFrequency, Frequency: &frequency})
			},
			method:          http.MethodPost,
			path:            "/api/v1/keywords/bulk",
			body:            `{"frequency":"daily","keyword_ids":["kw_1"],"operation":"set_frequency"}`,
			response:        KeywordBulkResponse{Operation: KeywordBulkOperationSetFrequency, Results: []KeywordBulkItemResult{{KeywordID: "kw_1", Status: "updated"}}},
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*KeywordBulkResponse).Operation, KeywordBulkOperationSetFrequency)
			},
		},
		{
			name: "list rank checks",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ListRankChecks(ctx, "kw_1", &ListRankChecksOptions{
					Cursor: "cursor_1",
					Limit:  5,
					Since:  since,
					Status: RankCheckStatusFailed,
					Until:  until,
				})
			},
			method:   http.MethodGet,
			path:     "/api/v1/keywords/kw_1/rank-checks",
			query:    "cursor=cursor_1&limit=5&since=2026-01-01T00%3A00%3A00Z&status=failed&until=2026-01-31T00%3A00%3A00Z",
			response: listResponse(check),
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*ListResponse[RankCheck]).Data[0].ID, "check_1")
			},
		},
		{
			name: "rank history alias",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.RankHistory(ctx, "kw_1", &ListRankChecksOptions{Limit: 2})
			},
			method:   http.MethodGet,
			path:     "/api/v1/keywords/kw_1/rank-checks",
			query:    "limit=2",
			response: listResponse(check),
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*ListResponse[RankCheck]).Data[0].Status, "completed")
			},
		},
		{
			name: "run rank check",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.RunRankCheck(ctx, "kw_1", &RunRankCheckInput{ProviderID: "dataforseo"})
			},
			method:          http.MethodPost,
			path:            "/api/v1/keywords/kw_1/checks",
			body:            `{"provider_id":"dataforseo"}`,
			response:        check,
			status:          http.StatusCreated,
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*RankCheck).Provider, "dataforseo")
			},
		},
		{
			name:            "run check alias without body",
			call:            func(ctx context.Context, c *Client) (any, error) { return c.RunCheck(ctx, "kw_1", nil) },
			method:          http.MethodPost,
			path:            "/api/v1/keywords/kw_1/checks",
			response:        check,
			status:          http.StatusCreated,
			wantNoBody:      true,
			wantContentType: false,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*RankCheck).ID, "check_1")
			},
		},
		{
			name: "run rank check async",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.RunRankCheck(ctx, "kw_1", &RunRankCheckInput{Async: true})
			},
			method:          http.MethodPost,
			path:            "/api/v1/keywords/kw_1/checks",
			query:           "async=true",
			response:        runningRankCheckJSON("check_2"),
			status:          http.StatusAccepted,
			wantNoBody:      true,
			wantContentType: false,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*RankCheck).Status, string(RankCheckStatusRunning))
			},
		},
		{
			name: "run rank check async with provider",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.RunRankCheck(ctx, "kw_1", &RunRankCheckInput{Async: true, ProviderID: "dataforseo"})
			},
			method:          http.MethodPost,
			path:            "/api/v1/keywords/kw_1/checks",
			query:           "async=true",
			body:            `{"provider_id":"dataforseo"}`,
			response:        runningRankCheckJSON("check_2"),
			status:          http.StatusAccepted,
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*RankCheck).ID, "check_2")
			},
		},
		{
			name:     "get rank check result",
			call:     func(ctx context.Context, c *Client) (any, error) { return c.GetRankCheckResult(ctx, "check_1") },
			method:   http.MethodGet,
			path:     "/api/v1/rank-checks/check_1",
			response: check,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*RankCheck).KeywordID, "kw_1")
			},
		},
		{
			name: "get failed rank check with attempts",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.GetRankCheckResult(ctx, "check_3")
			},
			method:   http.MethodGet,
			path:     "/api/v1/rank-checks/check_3",
			response: failedRankCheckJSON("check_3"),
			want: func(t *testing.T, got any) {
				t.Helper()
				result := got.(*RankCheck)
				assertEqual(t, result.Status, string(RankCheckStatusFailed))
				assertEqual(t, len(result.Attempts), 2)
				assertEqual(t, result.Attempts[0].Provider, "dataforseo")
				assertEqual(t, result.Attempts[0].Message, "Rate limited.")
				assertEqual(t, result.Attempts[1].Provider, "serpapi")
				assertEqual(t, *result.Error, "All providers failed.")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			runProtectedMethodTestCase(t, tt)
		})
	}
}

func runProtectedMethodTestCase(t *testing.T, tt protectedMethodTestCase) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertProtectedRequest(t, captureRequest(t, r), tt)
		writeJSON(t, w, statusOrOK(tt.status), tt.response)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	got, err := tt.call(context.Background(), client)
	if err != nil {
		t.Fatalf("call returned error: %v", err)
	}
	tt.want(t, got)
}

func assertProtectedRequest(t *testing.T, captured capturedRequest, tt protectedMethodTestCase) {
	t.Helper()
	assertEqual(t, captured.Method, tt.method)
	assertEqual(t, captured.Path, tt.path)
	assertEqual(t, captured.RawQuery, tt.query)
	assertEqual(t, captured.Header.Get("Authorization"), "Bearer "+testAPIKey)
	assertEqual(t, captured.Header.Get("X-Client"), "sdk-test")
	assertEqual(t, captured.Header.Get("Idempotency-Key"), tt.idempotencyKey)
	if tt.requestHeaderKey != "" {
		assertEqual(t, captured.Header.Get(tt.requestHeaderKey), tt.requestHeaderVal)
	}
	if tt.wantNoBody {
		assertEqual(t, captured.Body, "")
		assertEqual(t, captured.Header.Get(contentTypeHeader), "")
		return
	}
	if tt.body != "" {
		assertJSONEqual(t, captured.Body, tt.body)
	}
	if tt.wantContentType {
		assertEqual(t, captured.Header.Get(contentTypeHeader), "application/json")
	}
}

func TestPerRequestHeaderOverridesDefaultContentType(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != "application/vnd.bisibility+json" {
			t.Fatalf("Content-Type = %q", got)
		}
		writeJSON(t, w, http.StatusCreated, apiKeyFixture("key_new"))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	_, err := client.CreateAPIKey(
		context.Background(),
		CreateAPIKeyInput{Name: "CI"},
		WithRequestHeader("Content-Type", "application/vnd.bisibility+json"),
	)
	if err != nil {
		t.Fatalf("CreateAPIKey returned error: %v", err)
	}
}

func TestEmptySuccessfulJSONResponseReturnsNil(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	keyword, err := client.DeleteKeyword(context.Background(), "kw_1")
	if err != nil {
		t.Fatalf("DeleteKeyword returned error: %v", err)
	}
	if keyword != nil {
		t.Fatalf("keyword = %#v, want nil", keyword)
	}
}

func TestClientErrors(t *testing.T) {
	t.Parallel()

	t.Run("missing api key", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("server should not be called")
		}))
		defer server.Close()

		client, err := NewClient(WithBaseURL(server.URL + "/api/v1"))
		if err != nil {
			t.Fatalf("NewClient returned error: %v", err)
		}
		_, err = client.ListProjects(context.Background())
		assertConfigurationError(t, err)
	})

	t.Run("empty base url", func(t *testing.T) {
		t.Parallel()

		_, err := NewClient(WithBaseURL("   "))
		assertConfigurationError(t, err)
	})

	t.Run("relative base url", func(t *testing.T) {
		t.Parallel()

		_, err := NewClient(WithBaseURL("/api/v1"))
		assertConfigurationError(t, err)
	})

	t.Run("nil context", func(t *testing.T) {
		t.Parallel()

		client := newTestClient(t, "https://api.test/api/v1")
		_, err := client.ListProjects(nil)
		assertConfigurationError(t, err)
	})

	t.Run("update project rejects empty patch", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("server should not be called")
		}))
		defer server.Close()

		client := newTestClient(t, server.URL+"/api/v1")
		_, err := client.UpdateProject(context.Background(), "prj_1", UpdateProjectInput{})
		assertConfigurationError(t, err)
	})

	t.Run("update project defaults rejects invalid patches", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("server should not be called")
		}))
		defer server.Close()

		client := newTestClient(t, server.URL+"/api/v1")
		inputs := map[string]ProjectDefaultsPatch{
			"missing frequency": {Country: "United States", Device: DeviceDesktop},
			"country without device": {
				Country:   "United States",
				Frequency: RankCheckFrequencyDaily,
			},
			"device without country": {
				Device:    DeviceMobile,
				Frequency: RankCheckFrequencyDaily,
			},
		}
		for name, input := range inputs {
			_, err := client.UpdateProjectDefaults(context.Background(), "prj_1", input)
			t.Run(name, func(t *testing.T) {
				assertConfigurationError(t, err)
			})
		}
	})
}

func assertConfigurationError(t *testing.T, err error) {
	t.Helper()
	var configErr *ConfigurationError
	if !errors.As(err, &configErr) {
		t.Fatalf("err = %T, want ConfigurationError", err)
	}
}

func TestAPIErrorPaths(t *testing.T) {
	t.Parallel()

	problem := ProblemDetails{
		Type:     "https://bisibility.dev/problems/not_found",
		Title:    "Not found",
		Status:   http.StatusNotFound,
		Detail:   "Keyword not found.",
		Instance: "urn:bisibility:api:v1:/api/v1/keywords/kw_missing",
		DocsURL:  "https://bisibility.com/docs/api/errors#not_found",
	}

	tests := []struct {
		name        string
		status      int
		contentType string
		body        string
		wantMessage string
		wantProblem bool
	}{
		{
			name:        "problem details",
			status:      http.StatusNotFound,
			contentType: "application/problem+json",
			body:        mustJSON(t, problem),
			wantMessage: "Keyword not found.",
			wantProblem: true,
		},
		{
			name:        "idempotency conflict problem without docs fields",
			status:      http.StatusConflict,
			contentType: "application/problem+json",
			body:        `{"detail":"A request with this Idempotency-Key is still in progress.","status":409,"title":"Idempotency key in progress","type":"https://bisibility.dev/problems/idempotency_in_progress"}`,
			wantMessage: "A request with this Idempotency-Key is still in progress.",
			wantProblem: true,
		},
		{
			name:        "plain text body",
			status:      http.StatusBadGateway,
			contentType: "text/plain; charset=utf-8",
			body:        "upstream unavailable",
			wantMessage: "upstream unavailable",
		},
		{
			name:        "json body without problem details",
			status:      http.StatusBadRequest,
			contentType: "application/json",
			body:        `{"error":"not a problem"}`,
			wantMessage: `{"error":"not a problem"}`,
		},
		{
			name:        "json body with bare type is problem details",
			status:      http.StatusBadRequest,
			contentType: "application/json",
			body:        `{"type":"validation"}`,
			wantMessage: `{"type":"validation"}`,
			wantProblem: true,
		},
		{
			name:        "json body with bare status is not problem details",
			status:      http.StatusBadRequest,
			contentType: "application/json",
			body:        `{"status":400}`,
			wantMessage: `{"status":400}`,
		},
		{
			name:        "malformed json body",
			status:      http.StatusInternalServerError,
			contentType: "application/problem+json",
			body:        `{`,
			wantMessage: `{`,
		},
		{
			name:        "empty body",
			status:      http.StatusInternalServerError,
			contentType: "application/json",
			body:        "",
			wantMessage: "Bisibility API request failed with status 500.",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				w.Header().Set("Retry-After", "10")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := newTestClient(t, server.URL+"/api/v1")
			_, err := client.GetKeyword(context.Background(), "kw_missing")
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("err = %T, want APIError", err)
			}
			if !IsAPIError(err) {
				t.Fatalf("IsAPIError returned false")
			}
			assertEqual(t, apiErr.Error(), tt.wantMessage)
			assertEqual(t, apiErr.StatusCode, tt.status)
			assertEqual(t, apiErr.Header.Get("Retry-After"), "10")
			if (apiErr.Problem != nil) != tt.wantProblem {
				t.Fatalf("problem present = %v, want %v", apiErr.Problem != nil, tt.wantProblem)
			}
		})
	}
}

func TestResponseAndNetworkErrors(t *testing.T) {
	t.Parallel()

	t.Run("invalid success json", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("not json"))
		}))
		defer server.Close()

		client := newTestClient(t, server.URL+"/api/v1")
		_, err := client.ListProjects(context.Background())
		var responseErr *ResponseError
		if !errors.As(err, &responseErr) {
			t.Fatalf("err = %T, want ResponseError", err)
		}
		assertEqual(t, responseErr.Body, "not json")
		assertEqual(t, responseErr.StatusCode, http.StatusOK)
	})

	t.Run("invalid success json keeps real status", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("not json"))
		}))
		defer server.Close()

		client := newTestClient(t, server.URL+"/api/v1")
		_, err := client.CreateAPIKey(context.Background(), CreateAPIKeyInput{Name: "CI"})
		var responseErr *ResponseError
		if !errors.As(err, &responseErr) {
			t.Fatalf("err = %T, want ResponseError", err)
		}
		assertEqual(t, responseErr.StatusCode, http.StatusCreated)
	})

	t.Run("network failure", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("server should be closed before request")
		}))
		baseURL := server.URL + "/api/v1"
		server.Close()

		client := newTestClient(t, baseURL)
		_, err := client.ListProjects(context.Background())
		var networkErr *NetworkError
		if !errors.As(err, &networkErr) {
			t.Fatalf("err = %T, want NetworkError", err)
		}
		if !strings.Contains(networkErr.URL, "/api/v1/projects") {
			t.Fatalf("networkErr.URL = %q", networkErr.URL)
		}
	})
}

func TestUserAgentHeader(t *testing.T) {
	t.Parallel()

	t.Run("default user agent", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.Header.Get("User-Agent"); got != "bisibility-sdk-go/"+Version {
				t.Fatalf("User-Agent = %q, want %q", got, "bisibility-sdk-go/"+Version)
			}
			if got := r.Header.Get("X-Bisibility-Client"); got != "bisibility-sdk-go/"+Version {
				t.Fatalf("X-Bisibility-Client = %q, want %q", got, "bisibility-sdk-go/"+Version)
			}
			writeJSON(t, w, http.StatusOK, healthFixture())
		}))
		defer server.Close()

		client := newTestClient(t, server.URL+"/api/v1")
		if _, err := client.GetHealth(context.Background()); err != nil {
			t.Fatalf("GetHealth returned error: %v", err)
		}
	})

	t.Run("default header override wins", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.Header.Get("User-Agent"); got != "custom-agent/1.0" {
				t.Fatalf("User-Agent = %q, want custom-agent/1.0", got)
			}
			writeJSON(t, w, http.StatusOK, healthFixture())
		}))
		defer server.Close()

		client, err := NewClient(
			WithBaseURL(server.URL+"/api/v1"),
			WithDefaultHeader("User-Agent", "custom-agent/1.0"),
		)
		if err != nil {
			t.Fatalf("NewClient returned error: %v", err)
		}
		if _, err := client.GetHealth(context.Background()); err != nil {
			t.Fatalf("GetHealth returned error: %v", err)
		}
	})
}

func TestDefaultHTTPClientTimeout(t *testing.T) {
	t.Parallel()

	client, err := NewClient(WithAPIKey(testAPIKey))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if client.httpClient == http.DefaultClient {
		t.Fatal("default httpClient must not be http.DefaultClient")
	}
	assertEqual(t, client.httpClient.Timeout, 30*time.Second)

	custom := &http.Client{Timeout: time.Second}
	client, err = NewClient(WithAPIKey(testAPIKey), WithHTTPClient(custom))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if client.httpClient != custom {
		t.Fatal("WithHTTPClient must override the default client")
	}
}

func newTestClient(t *testing.T, baseURL string) *Client {
	t.Helper()

	client, err := NewClient(
		WithAPIKey(testAPIKey),
		WithBaseURL(baseURL),
		WithDefaultHeader("X-Client", "sdk-test"),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	return client
}

func captureRequest(t *testing.T, r *http.Request) capturedRequest {
	t.Helper()

	body := make([]byte, r.ContentLength)
	if r.ContentLength > 0 {
		_, err := r.Body.Read(body)
		if err != nil && !errors.Is(err, http.ErrBodyReadAfterClose) && err.Error() != "EOF" {
			t.Fatalf("reading body: %v", err)
		}
	}

	return capturedRequest{
		Body:          string(body),
		Header:        r.Header.Clone(),
		Method:        r.Method,
		Path:          r.URL.EscapedPath(),
		Query:         r.URL.Query(),
		RawQuery:      r.URL.RawQuery,
		ContentLength: r.ContentLength,
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, status int, value any) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if value == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("encoding JSON: %v", err)
	}
}

func statusOrOK(status int) int {
	if status == 0 {
		return http.StatusOK
	}
	return status
}

func assertJSONEqual(t *testing.T, got, want string) {
	t.Helper()

	var gotValue any
	if err := json.Unmarshal([]byte(got), &gotValue); err != nil {
		t.Fatalf("got invalid json %q: %v", got, err)
	}
	var wantValue any
	if err := json.Unmarshal([]byte(want), &wantValue); err != nil {
		t.Fatalf("want invalid json %q: %v", want, err)
	}
	if !reflect.DeepEqual(gotValue, wantValue) {
		t.Fatalf("json body = %s, want %s", got, want)
	}
}

func assertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return string(body)
}

func listResponse[T any](items ...T) ListResponse[T] {
	return ListResponse[T]{
		Data: items,
		Meta: ListMeta{},
	}
}

func healthFixture() HealthResponse {
	var health HealthResponse
	health.Status = "ok"
	health.CheckedAt = mustTime("2026-01-01T00:00:00Z")
	health.Providers.SERP = []string{"dataforseo"}
	health.Services.App = "ok"
	health.Services.Database = "ok"
	return health
}

func projectFixture(id string) Project {
	return Project{
		ID:        id,
		Name:      "Example",
		Domain:    "example.com",
		WriteMode: ProjectWriteModeActive,
		CreatedAt: mustTime("2026-01-01T00:00:00Z"),
		UpdatedAt: mustTime("2026-01-02T00:00:00Z"),
	}
}

// projectJSON mirrors the app's projectResource shape (lib/api/resources.ts).
func projectJSON(id string) map[string]any {
	return map[string]any{
		"created_at": "2026-01-01T00:00:00Z",
		"domain":     "example.com",
		"id":         id,
		"name":       "Example",
		"updated_at": "2026-01-02T00:00:00Z",
		"write_mode": "active",
	}
}

// projectDefaultsJSON mirrors the app's defaultsResource shape (lib/api/projects.ts).
func projectDefaultsJSON(projectID string) map[string]any {
	return map[string]any{
		"auto_schedule":   true,
		"city":            "Austin",
		"country":         "United States",
		"cron_expression": nil,
		"device":          "mobile",
		"frequency":       "daily",
		"jitter_minutes":  30,
		"last_checked_at": nil,
		"location_key":    "US/Texas/Austin",
		"next_check_at":   "2026-01-05T00:00:00Z",
		"project_id":      projectID,
		"timezone":        "America/Chicago",
		"updated_at":      "2026-01-04T00:00:00Z",
	}
}

func apiKeyFixture(id string) APIKey {
	return APIKey{
		ID:        id,
		Name:      "Production",
		Prefix:    "bsk_live_12345678",
		CreatedAt: mustTime("2026-01-01T00:00:00Z"),
	}
}

func keywordFixture(id string) Keyword {
	position := 4
	previous := 8
	target := "https://example.com/page"
	return Keyword{
		ID:               id,
		ProjectID:        "prj_1",
		Text:             "rank tracker",
		Country:          "United States",
		Location:         "United States",
		Device:           DeviceDesktop,
		TargetURL:        &target,
		RankingURL:       &target,
		LatestPosition:   &position,
		PreviousPosition: &previous,
		Tags:             []string{"Product"},
		CreatedAt:        mustTime("2026-01-03T00:00:00Z"),
		UpdatedAt:        mustTime("2026-01-04T00:00:00Z"),
	}
}

func createKeywordsResponse(keyword Keyword) CreateKeywordsResponse {
	return CreateKeywordsResponse{
		Created: 1,
		Skipped: 0,
		Results: []CreateKeywordResult{
			{Keyword: keyword, Status: "created"},
		},
	}
}

// keywordJSON mirrors the app's keywordResource shape (lib/api/resources.ts).
func keywordJSON(id string) map[string]any {
	return map[string]any{
		"country":           "United States",
		"created_at":        "2026-01-03T00:00:00Z",
		"device":            "desktop",
		"id":                id,
		"intent":            "commercial",
		"latest_position":   4,
		"location":          "United States",
		"previous_position": 8,
		"project_id":        "prj_1",
		"ranking_url":       "https://example.com/page",
		"schedule":          nil,
		"tags":              []string{"Product"},
		"target_url":        "https://example.com/page",
		"text":              "rank tracker",
		"topic":             nil,
		"updated_at":        "2026-01-04T00:00:00Z",
	}
}

// createKeywordsResponseJSON mirrors the app's keyword create response shape
// (lib/api/keyword-create.ts), including per-result and top-level warnings.
func createKeywordsResponseJSON() map[string]any {
	warning := "City not found; tracking at country level."
	keyword := keywordJSON("kw_1")
	keyword["topic"] = "tooling"
	return map[string]any{
		"created": 1,
		"results": []map[string]any{
			{"keyword": keyword, "status": "created", "warning": warning},
		},
		"skipped":  0,
		"warnings": []string{warning},
	}
}

// runningRankCheckJSON mirrors the app's 202 async rank check shape
// (lib/api/rank-checks.ts + rankCheckResource).
func runningRankCheckJSON(id string) map[string]any {
	return map[string]any{
		"attempts":          nil,
		"checked_at":        "2026-01-06T00:00:00Z",
		"cost_cents":        nil,
		"error":             nil,
		"id":                id,
		"keyword_id":        "kw_1",
		"position":          nil,
		"previous_position": nil,
		"provider":          "primary",
		"ranking_url":       nil,
		"status":            "running",
	}
}

// failedRankCheckJSON mirrors the app's failed rank check shape with provider
// fallback attempts (lib/api/resources.ts rankCheckResource).
func failedRankCheckJSON(id string) map[string]any {
	return map[string]any{
		"attempts": []map[string]any{
			{"message": "Rate limited.", "provider": "dataforseo"},
			{"message": "No results.", "provider": "serpapi"},
		},
		"checked_at":        "2026-01-06T00:00:00Z",
		"cost_cents":        nil,
		"error":             "All providers failed.",
		"id":                id,
		"keyword_id":        "kw_1",
		"position":          nil,
		"previous_position": nil,
		"provider":          "dataforseo",
		"ranking_url":       nil,
		"status":            "failed",
	}
}

func rankCheckFixture(id string) RankCheck {
	position := 4
	previous := 8
	cost := 0.06
	target := "https://example.com/page"
	return RankCheck{
		ID:               id,
		KeywordID:        "kw_1",
		CheckedAt:        mustTime("2026-01-06T00:00:00Z"),
		CostCents:        &cost,
		Position:         &position,
		PreviousPosition: &previous,
		Provider:         "dataforseo",
		RankingURL:       &target,
		Status:           "completed",
	}
}

func mustTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(fmt.Sprintf("parse time %q: %v", value, err))
	}
	return parsed
}

func strPtr(value string) *string {
	return &value
}
