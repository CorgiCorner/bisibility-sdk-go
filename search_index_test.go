package bisibility

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

const searchIndexProject = "prj_a00000000000000000000000"
const searchIndexRoot = "/api/v1/projects/" + searchIndexProject + "/search-index/"

func TestSearchIndexSitemaps(t *testing.T) {
	sitemap := json.RawMessage(`{"path":"https://example.com/sitemap.xml","type":null,"is_sitemaps_index":true,"is_pending":false,"errors":2,"warnings":3,"last_submitted":null,"last_downloaded":"2026-10-23T10:00:00Z","contents":[{"type":"web","submitted":12}]}`)
	runResourceMethodTestCase(t, resourceMethodTestCase{
		call: func(ctx context.Context, c *Client) (any, error) {
			return c.ListSearchIndexSitemaps(ctx, searchIndexProject)
		},
		method: http.MethodGet, path: searchIndexRoot + "sitemaps",
		response: map[string]any{"property": "sc-domain:example.com", "sitemaps": []any{sitemap}},
		want: func(t *testing.T, got any) {
			response := got.(*SearchIndexSitemaps)
			assertEqual(t, response.Property, "sc-domain:example.com")
			assertEqual(t, response.Sitemaps[0].Contents[0].Submitted, 12)
			assertResponseFieldsRoundTrip(t, sitemap, response.Sitemaps[0], "type", "last_submitted", "last_downloaded", "contents", "errors", "warnings", "is_pending", "is_sitemaps_index", "path")
		},
	})
	fullURL := "https://example.com/maps/sitemap.xml?kind=web&part=1:2"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.Method, http.MethodGet)
		assertEqual(t, r.URL.EscapedPath(), searchIndexRoot+"sitemaps/"+url.PathEscape(fullURL))
		assertEqual(t, r.URL.RawQuery, "")
		assertEqual(t, r.Header.Get("Authorization"), "Bearer "+testAPIKey)
		writeJSON(t, w, 200, map[string]any{"property": "sc-domain:example.com", "sitemap": sitemap})
	}))
	defer server.Close()
	response, err := newTestClient(t, server.URL+"/api/v1").GetSearchIndexSitemap(context.Background(), searchIndexProject, fullURL)
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, response.Sitemap.Path, "https://example.com/sitemap.xml")
	assertEqual(t, response.Sitemap.Type == nil, true)
}

func TestGetSearchIndexSitemapPathValidation(t *testing.T) {
	const prefix = "https://example.com/"
	for _, tc := range []struct {
		name string
		path string
	}{
		{"empty", ""},
		{"dot", "."},
		{"dot dot", ".."},
		{"relative path", "sitemap.xml"},
		{"root relative path", "/sitemap.xml"},
		{"FTP scheme", "ftp://example.com/sitemap.xml"},
		{"mailto scheme", "mailto:a@example.com"},
		{"missing host", "https://"},
		{"invalid URL escape", prefix + "%zz"},
		{"oversized ASCII", prefix + strings.Repeat("a", 2049-len(prefix))},
		{"oversized multibyte", prefix + strings.Repeat("é", 2049-len(prefix))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			client := newTestClient(t, "https://api.example.com/api/v1")
			client.maxRetries = 0
			client.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				calls++
				return nil, errors.New("unexpected request")
			})}
			result, err := client.GetSearchIndexSitemap(context.Background(), searchIndexProject, tc.path)
			assertEqual(t, calls, 0)
			assertEqual(t, result == nil, true)
			assertConfigurationError(t, err)
		})
	}
	for _, tc := range []struct {
		name string
		path string
	}{
		{"2048 characters ASCII", prefix + strings.Repeat("a", 2048-len(prefix))},
		{"2048 characters multibyte", prefix + strings.Repeat("é", 2048-len(prefix))},
		{"uppercase scheme", "HTTPS://example.com/sitemap.xml"},
		{"HTTP scheme", "http://example.com/sitemap.xml"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := tc.path
			runResourceMethodTestCase(t, resourceMethodTestCase{
				call: func(ctx context.Context, c *Client) (any, error) {
					return c.GetSearchIndexSitemap(ctx, searchIndexProject, path)
				},
				method: http.MethodGet, path: searchIndexRoot + "sitemaps/" + url.PathEscape(path),
				response: map[string]any{"sitemap": map[string]any{"path": path}},
				want: func(t *testing.T, got any) {
					assertEqual(t, got.(*SearchIndexSitemapResponse).Sitemap.Path, path)
				},
			})
		})
	}
}

func TestSearchIndexInspection(t *testing.T) {
	inspection := json.RawMessage(`{"verdict":"FUTURE_VERDICT","coverage_state":null,"robots_txt_state":"FUTURE_ROBOTS","indexing_state":null,"last_crawl_time":null,"last_crawl_at":"2026-10-23T10:00:00Z","page_fetch_state":null,"google_canonical":null,"user_canonical":null,"crawled_as":null,"referring_urls":[],"sitemap":["https://example.com/sitemap.xml"],"inspection_result_link":null,"rich_results_result":{"verdict":null,"detected_items":[{"rich_result_type":"Product","item_count":1,"issue_count":2,"issues":[{"severity":null,"count":2}]}]}}`)
	response := json.RawMessage(`{"property":"sc-domain:example.com","results":[{"url":"https://example.com/ok","inspection":` + string(inspection) + `,"error":null},{"url":"https://example.com/limited","inspection":null,"error":{"code":"provider_rate_limited","message":"wait","reset_at":"2026-10-24T10:00:00Z","scope":"minute"}},{"url":"https://example.com/failed","inspection":null,"error":{"code":"inspection_failed","message":"failed"}}]}`)
	runResourceMethodTestCase(t, resourceMethodTestCase{
		call: func(ctx context.Context, c *Client) (any, error) {
			return c.InspectSearchIndexURLs(ctx, searchIndexProject, SearchIndexInspectionInput{URLs: []string{"https://example.com/ok", "https://example.com/limited", "https://example.com/failed"}}, WithIdempotencyKey("inspect-1"))
		},
		method: http.MethodPost, path: searchIndexRoot + "inspections", body: `{"urls":["https://example.com/ok","https://example.com/limited","https://example.com/failed"]}`, idempotencyKey: "inspect-1", wantContentType: true, response: response,
		want: func(t *testing.T, got any) {
			result := got.(*SearchIndexInspections)
			assertEqual(t, result.Results[0].Error == nil, true)
			assertEqual(t, result.Results[1].Inspection == nil, true)
			limited := result.Results[1].Error
			assertEqual(t, limited.Code, SearchIndexInspectionErrorCodeProviderRateLimited)
			assertEqual(t, *limited.Scope, SearchIndexRateLimitScopeMinute)
			assertEqual(t, limited.ResetAt.UTC().Format(time.RFC3339), "2026-10-24T10:00:00Z")
			assertEqual(t, result.Results[2].Error.ResetAt == nil, true)
			assertEqual(t, result.Results[2].Error.Scope == nil, true)
			encoded, err := json.Marshal(result.Results[0].Inspection)
			if err != nil {
				t.Fatal(err)
			}
			assertJSONEqual(t, string(encoded), string(inspection))
		},
	})
	for _, scope := range []SearchIndexRateLimitScope{SearchIndexRateLimitScopeDaily, SearchIndexRateLimitScopeMinute, SearchIndexRateLimitScopeUnknown, "future_scope"} {
		data := `{"code":"future_code","message":"new","scope":"` + string(scope) + `"}`
		var got SearchIndexInspectionError
		if err := json.Unmarshal([]byte(data), &got); err != nil {
			t.Fatal(err)
		}
		assertEqual(t, *got.Scope, scope)
		assertEqual(t, string(got.Code), "future_code")
		assertResponseFieldsRoundTrip(t, json.RawMessage(data), got, "code", "scope")
	}
}

func TestSearchIndexRetryAndErrors(t *testing.T) {
	for _, key := range []string{"", "inspect-retry"} {
		t.Run("key="+key, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				assertEqual(t, r.Header.Get("Idempotency-Key"), key)
				if calls == 1 {
					w.Header().Set("Retry-After", "0")
					writeJSON(t, w, 503, map[string]any{"title": "Unavailable", "status": 503})
					return
				}
				writeJSON(t, w, 200, map[string]any{"property": "sc-domain:example.com", "results": []any{}})
			}))
			defer server.Close()
			client := newTestClient(t, server.URL+"/api/v1")
			client.maxRetries = 1
			result, err := client.InspectSearchIndexURLs(context.Background(), searchIndexProject, SearchIndexInspectionInput{URLs: []string{"https://example.com/"}}, WithIdempotencyKey(key))
			if key == "" {
				var apiErr *APIError
				if result != nil || !errors.As(err, &apiErr) {
					t.Fatalf("expected APIError, got %v", err)
				}
				assertEqual(t, apiErr.StatusCode, 503)
				assertEqual(t, calls, 1)
			} else {
				if err != nil {
					t.Fatal(err)
				}
				assertEqual(t, calls, 2)
			}
		})
	}
	for _, status := range []int{400, 404, 422, 429, 502} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"type":"https://example.com/problem","title":"Search index error","detail":"connection unavailable","status":` + strconv.Itoa(status) + `}`))
		}))
		client := newTestClient(t, server.URL+"/api/v1")
		client.maxRetries = 0
		_, err := client.ListSearchIndexSitemaps(context.Background(), searchIndexProject)
		server.Close()
		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.Problem == nil {
			t.Fatalf("expected problem APIError, got %v", err)
		}
		assertEqual(t, apiErr.StatusCode, status)
		assertEqual(t, apiErr.Problem.Detail, "connection unavailable")
	}
}

func TestPageIndexingReportAndSettings(t *testing.T) {
	report := json.RawMessage(`{"property":"sc-domain:example.com","checked":1,"total":3,"complete":false,"remaining":2,"oldest_result_at":null,"newest_result_at":"2026-10-23T10:00:00Z","budget_mode":"future_mode","today_count":1,"last429_kind":null,"next_reset_at":"2026-10-24T07:00:00Z","discovery_complete":false,"discovered_at":null,"estimated_days_remaining":null,"groups":[{"group":"future_group","checked":0,"total":2,"complete":false,"remaining":2,"oldest_result_at":null,"newest_result_at":null,"budget_mode":"off","today_count":0,"last429_kind":"future_limit","next_reset_at":"2026-10-24T07:00:00Z","reasons":[{"coverage_state":null,"total":2,"oldest_result_at":null},{"coverage_state":"Indexed","total":1,"oldest_result_at":"2026-10-23T10:00:00.123456789Z"}]}]}`)
	settings := json.RawMessage(`{"property":"sc-domain:example.com","budget_mode":"fixed","inspection_daily_limit":2000,"available_modes":["off","fixed","future_mode"],"today_count":14,"last429_kind":null,"last429_at":null,"next_reset_at":"2026-10-24T07:00:00Z"}`)
	for _, tc := range []resourceMethodTestCase{
		{
			name: "report", method: http.MethodGet, path: searchIndexRoot + "report", response: report,
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.GetPageIndexingReport(ctx, searchIndexProject)
			},
			want: func(t *testing.T, got any) {
				result := got.(*PageIndexingReport)
				assertEqual(t, result.Complete, false)
				assertEqual(t, result.EstimatedDaysRemaining == nil, true)
				assertEqual(t, string(result.BudgetMode), "future_mode")
				assertEqual(t, string(result.Groups[0].Group), "future_group")
				assertEqual(t, result.Groups[0].Reasons[0].CoverageState == nil, true)
				assertEqual(t, result.Groups[0].Reasons[0].OldestResultAt == nil, true)
				reasonOldest := result.Groups[0].Reasons[1].OldestResultAt
				if reasonOldest == nil {
					t.Fatal("expected oldest stored result for reason")
				}
				assertEqual(t, reasonOldest.UTC().Format(time.RFC3339Nano), "2026-10-23T10:00:00.123456789Z")
				encoded, err := json.Marshal(result)
				if err != nil {
					t.Fatal(err)
				}
				assertJSONEqual(t, string(encoded), string(report))
			},
		},
		{
			name: "settings", method: http.MethodGet, path: searchIndexRoot + "settings", response: settings,
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.GetPageIndexingSettings(ctx, searchIndexProject)
			},
			want: func(t *testing.T, got any) {
				result := got.(*PageIndexingSettings)
				assertEqual(t, result.InspectionDailyLimit, 2000)
				assertEqual(t, result.BudgetMode, PageIndexingBudgetModeFixed)
				assertEqual(t, string(result.AvailableModes[2]), "future_mode")
				encoded, err := json.Marshal(result)
				if err != nil {
					t.Fatal(err)
				}
				assertJSONEqual(t, string(encoded), string(settings))
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) { runResourceMethodTestCase(t, tc) })
	}
	var populated PageIndexingReport
	if err := json.Unmarshal([]byte(`{"oldest_result_at":"2026-10-23T10:00:00Z","discovered_at":"2026-10-23T10:00:00Z","estimated_days_remaining":0}`), &populated); err != nil {
		t.Fatal(err)
	}
	assertEqual(t, populated.OldestResultAt != nil, true)
	assertEqual(t, populated.DiscoveredAt != nil, true)
	assertEqual(t, *populated.EstimatedDaysRemaining, 0)
}

func TestPageIndexingSettingsPartialPatch(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input PageIndexingSettingsPatch
		body  string
	}{
		{"mode only", PageIndexingSettingsPatch{BudgetMode: ptr(PageIndexingBudgetModeOff)}, `{"budget_mode":"off"}`},
		{"explicit zero", PageIndexingSettingsPatch{InspectionDailyLimit: ptr(0)}, `{"inspection_daily_limit":0}`},
		{"both", PageIndexingSettingsPatch{BudgetMode: ptr(PageIndexingBudgetModeAvailableQuota), InspectionDailyLimit: ptr(2000)}, `{"budget_mode":"available_quota","inspection_daily_limit":2000}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runResourceMethodTestCase(t, resourceMethodTestCase{
				call: func(ctx context.Context, c *Client) (any, error) {
					return c.UpdatePageIndexingSettings(ctx, searchIndexProject, tc.input, WithIdempotencyKey("settings-1"))
				},
				method: http.MethodPatch, path: searchIndexRoot + "settings", body: tc.body, wantContentType: true, idempotencyKey: "settings-1",
				response: json.RawMessage(`{"property":"sc-domain:example.com","budget_mode":"off","inspection_daily_limit":0,"available_modes":["off","fixed"],"today_count":0,"last429_kind":"daily","last429_at":"2026-10-23T10:00:00Z","next_reset_at":"2026-10-24T07:00:00Z"}`),
				want: func(t *testing.T, got any) {
					result := got.(*PageIndexingSettings)
					assertEqual(t, result.InspectionDailyLimit, 0)
					assertEqual(t, *result.Last429Kind, "daily")
					assertEqual(t, result.Last429At.UTC().Format(time.RFC3339), "2026-10-23T10:00:00Z")
				},
			})
		})
	}
}

func TestUpdatePageIndexingSettingsRejectsEmptyPatch(t *testing.T) {
	calls := 0
	client := newTestClient(t, "https://api.example.com/api/v1")
	client.maxRetries = 0
	client.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("unexpected request")
	})}
	result, err := client.UpdatePageIndexingSettings(context.Background(), searchIndexProject, PageIndexingSettingsPatch{})
	assertEqual(t, calls, 0)
	assertEqual(t, result == nil, true)
	assertConfigurationError(t, err)
}

func TestPageIndexingURLs(t *testing.T) {
	changedSince := time.Date(2026, time.October, 23, 12, 0, 0, 123456789, time.FixedZone("offset", 2*60*60))
	nullable := json.RawMessage(`{"url":"https://example.com/a","verdict":null,"coverage_state":null,"last_crawl_at":null,"canonical_ok":null,"checked_at":null,"created_at":"2026-10-23T10:00:00Z","updated_at":"2026-10-23T10:00:00Z","source":"future_source","sitemap_url":null,"google_canonical":null,"user_canonical":null,"indexing_state":null,"robots_txt_state":null,"page_fetch_state":null,"crawled_as":null,"referring_urls":[],"sitemaps":[],"rich_results_summary":null,"inspection_result_link":null,"inspection_error":null,"inspection_failed_at":null,"group":"future_group"}`)
	var populated map[string]any
	if err := json.Unmarshal(nullable, &populated); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"verdict", "coverage_state", "sitemap_url", "google_canonical", "user_canonical", "indexing_state", "robots_txt_state", "page_fetch_state", "crawled_as", "inspection_result_link", "inspection_error"} {
		populated[key] = "future_value"
	}
	for _, key := range []string{"last_crawl_at", "checked_at", "inspection_failed_at"} {
		populated[key] = "2026-10-23T10:00:00Z"
	}
	populated["canonical_ok"] = false
	populated["source"] = "manual"
	populated["group"] = "inspection_failed"
	populated["referring_urls"] = []string{"https://example.com/referrer"}
	populated["sitemaps"] = []string{"https://example.com/sitemap.xml"}
	populated["rich_results_summary"] = json.RawMessage(`{"verdict":"future_verdict","detected_items":[{"rich_result_type":null,"item_count":1,"issue_count":1,"issues":[{"severity":"future_severity","count":1}]}]}`)
	populated["history"] = json.RawMessage(`{"complete":false,"changes":[{"at":"2026-10-23T10:00:00Z","previous":{"verdict":null,"coverage_state":null,"google_canonical":null,"indexing_state":null,"robots_txt_state":null,"page_fetch_state":null},"next":{"verdict":"future_verdict","coverage_state":"Indexed","google_canonical":"https://example.com/a","indexing_state":"future_indexing","robots_txt_state":"future_robots","page_fetch_state":"future_fetch"}}]}`)
	full, err := json.Marshal(populated)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		input *ListPageIndexingURLsOptions
		query string
		row   json.RawMessage
	}{
		{"defaults", nil, "", nullable},
		{"changed since", &ListPageIndexingURLsOptions{ChangedSince: &changedSince}, "changed_since=2026-10-23T10%3A00%3A00.123456789Z", nullable},
		{"nil changed since omitted", &ListPageIndexingURLsOptions{ChangedSince: nil}, "", nullable},
		{"filters and history", &ListPageIndexingURLsOptions{Group: PageIndexingGroupKindNotIndexed, CoverageState: "Unspecified", Sitemap: "https://example.com/sitemap.xml?type=a:b", Source: PageIndexingSourceManual, Search: "/a b", Cursor: "opaque+/=", Limit: 100, History: true}, "coverage_state=Unspecified&cursor=opaque%2B%2F%3D&group=not_indexed&history=true&limit=100&search=%2Fa+b&sitemap=https%3A%2F%2Fexample.com%2Fsitemap.xml%3Ftype%3Da%3Ab&source=manual", full},
		{"false history omitted", &ListPageIndexingURLsOptions{History: false}, "", nullable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := map[string]any{"data": []any{tc.row}, "meta": map[string]any{"next_cursor": nil}}
			runResourceMethodTestCase(t, resourceMethodTestCase{
				call: func(ctx context.Context, c *Client) (any, error) {
					return c.ListPageIndexingURLs(ctx, searchIndexProject, tc.input)
				},
				method: http.MethodGet, path: searchIndexRoot + "urls", query: tc.query, response: body,
				want: func(t *testing.T, got any) {
					result := got.(*PageIndexingURLs)
					assertEqual(t, result.Meta.NextCursor == nil, true)
					row := result.Data[0]
					encoded, err := json.Marshal(row)
					if err != nil {
						t.Fatal(err)
					}
					assertJSONEqual(t, string(encoded), string(tc.row))
					if tc.input != nil && tc.input.History {
						assertEqual(t, *row.CanonicalOK, false)
						assertEqual(t, row.Group, PageIndexingGroupKindInspectionFailed)
						assertEqual(t, row.History.Complete, false)
						assertEqual(t, row.History.Changes[0].Previous.Verdict == nil, true)
						assertEqual(t, *row.History.Changes[0].Next.Verdict, "future_verdict")
					} else {
						assertEqual(t, row.History == nil, true)
					}
				},
			})
		})
	}
}

func TestIteratePageIndexingURLs(t *testing.T) {
	changedSince := time.Date(2026, time.October, 23, 12, 0, 0, 123456789, time.FixedZone("offset", 2*60*60))
	filters := ListPageIndexingURLsOptions{ChangedSince: &changedSince, Cursor: "initial", Group: PageIndexingGroupKindIndexed, CoverageState: "Indexed", Sitemap: "https://example.com/sitemap.xml", Source: PageIndexingSourceSitemap, Search: "a", Limit: 1, History: true}
	expected := filters
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		assertEqual(t, r.Method, http.MethodGet)
		assertEqual(t, r.URL.Path, searchIndexRoot+"urls")
		assertEqual(t, r.Header.Get("X-Request-Test"), "pager")
		query := r.URL.Query()
		assertEqual(t, query.Get("group"), string(expected.Group))
		assertEqual(t, query.Get("coverage_state"), expected.CoverageState)
		assertEqual(t, query.Get("sitemap"), expected.Sitemap)
		assertEqual(t, query.Get("source"), string(expected.Source))
		assertEqual(t, query.Get("search"), expected.Search)
		assertEqual(t, query.Get("limit"), "1")
		assertEqual(t, query.Get("history"), "true")
		assertEqual(t, query.Get("changed_since"), "2026-10-23T10:00:00.123456789Z")
		switch calls {
		case 1:
			assertEqual(t, query.Get("cursor"), "initial")
			writeJSON(t, w, 200, listEnvelope([]any{map[string]any{"url": "https://example.com/a"}}, "next+/="))
		case 2:
			assertEqual(t, query.Get("cursor"), "next+/=")
			writeJSON(t, w, 200, listEnvelope([]any{}, "last"))
		case 3:
			assertEqual(t, query.Get("cursor"), "last")
			writeJSON(t, w, 200, map[string]any{"data": []any{map[string]any{"url": "https://example.com/b"}}, "meta": map[string]any{"next_cursor": nil}})
		default:
			t.Errorf("unexpected page %d", calls)
			w.WriteHeader(400)
		}
	}))
	defer server.Close()
	pager := newTestClient(t, server.URL+"/api/v1").IteratePageIndexingURLs(context.Background(), searchIndexProject, &filters, WithRequestHeader("X-Request-Test", "pager"))
	filters.Search = "changed after construction"
	filters.ChangedSince = nil
	var urls []string
	for pager.Next() {
		urls = append(urls, pager.Item().URL)
	}
	if pager.Err() != nil {
		t.Fatal(pager.Err())
	}
	assertEqual(t, len(urls), 2)
	assertEqual(t, urls[0], "https://example.com/a")
	assertEqual(t, urls[1], "https://example.com/b")
	assertEqual(t, calls, 3)
	assertEqual(t, filters.Cursor, "initial")
}

func TestIteratePageIndexingURLsFailure(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		assertEqual(t, r.URL.RawQuery, "")
		writeJSON(t, w, 422, map[string]any{"title": "Connection unavailable", "status": 422})
	}))
	defer server.Close()
	client := newTestClient(t, server.URL+"/api/v1")
	pager := client.IteratePageIndexingURLs(context.Background(), searchIndexProject, nil)
	assertEqual(t, pager.Next(), false)
	assertEqual(t, pager.Next(), false)
	var apiErr *APIError
	if !errors.As(pager.Err(), &apiErr) {
		t.Fatalf("expected APIError, got %v", pager.Err())
	}
	assertEqual(t, apiErr.StatusCode, 422)
	assertEqual(t, calls, 1)
}
