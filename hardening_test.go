package bisibility

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func collectPager[T any](t *testing.T, pager *Pager[T]) []T {
	t.Helper()
	var items []T
	for pager.Next() {
		items = append(items, pager.Item())
	}
	if err := pager.Err(); err != nil {
		t.Fatalf("pager error: %v", err)
	}
	return items
}

type pagerResponse struct {
	prefix PublicIDPrefix
	tag    string
	fields map[string]any
}

func (response pagerResponse) item(cursor string) (map[string]any, any) {
	id := "second"
	if response.prefix != "" {
		id = strictID(response.prefix)
	}
	next := any(nil)
	if cursor != "next" {
		next = "next"
		if response.prefix == "" {
			id = "first"
		}
	}
	item := map[string]any{"id": id}
	for key, value := range response.fields {
		item[key] = value
	}
	return item, next
}

type pagerCase struct {
	name     string
	path     string
	response pagerResponse
	iterate  func(*testing.T, *Client, context.Context) int
}

func pagerTestResponses(cases []pagerCase) map[string]pagerResponse {
	responses := make(map[string]pagerResponse, len(cases))
	for _, testCase := range cases {
		responses[testCase.path] = testCase.response
	}
	return responses
}

func pagerTestServer(t *testing.T, responses map[string]pagerResponse) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response, ok := responses[r.URL.Path]
		if !ok {
			t.Errorf("unexpected pager path %q", r.URL.Path)
			return
		}
		if response.tag != "" && r.URL.Query().Get("filter[tag]") != response.tag {
			t.Errorf("keyword tag filter was not preserved")
		}
		item, next := response.item(r.URL.Query().Get("cursor"))
		writeJSON(t, w, http.StatusOK, map[string]any{
			"data": []map[string]any{item},
			"meta": map[string]any{"next_cursor": next},
		})
	})
}

func TestResourcePagers(t *testing.T) {
	const projectID = "prj_a00000000000000000000000"
	const keywordID = "kw_a00000000000000000000000"
	cases := []pagerCase{
		{
			name:     "API keys",
			path:     "/api/v1/api-keys",
			response: pagerResponse{prefix: PublicIDPrefixKey},
			iterate: func(t *testing.T, client *Client, ctx context.Context) int {
				return len(collectPager(t, client.IterateAPIKeys(ctx, nil)))
			},
		},
		{
			name:     "project API keys",
			path:     "/api/v1/projects/" + projectID + "/api-keys",
			response: pagerResponse{prefix: PublicIDPrefixKey},
			iterate: func(t *testing.T, client *Client, ctx context.Context) int {
				return len(collectPager(t, client.IterateProjectAPIKeys(ctx, projectID, nil)))
			},
		},
		{
			name:     "webhooks",
			path:     "/api/v1/projects/" + projectID + "/webhooks",
			response: pagerResponse{prefix: PublicIDPrefixWebhook},
			iterate: func(t *testing.T, client *Client, ctx context.Context) int {
				return len(collectPager(t, client.IterateWebhooks(ctx, projectID, nil)))
			},
		},
		{
			name: "keywords",
			path: "/api/v1/projects/" + projectID + "/keywords",
			response: pagerResponse{
				prefix: PublicIDPrefixKeyword,
				tag:    "stable",
				fields: map[string]any{"project_id": strictID(PublicIDPrefixProject)},
			},
			iterate: func(t *testing.T, client *Client, ctx context.Context) int {
				return len(collectPager(t, client.IterateKeywords(ctx, projectID, &ListKeywordsOptions{Tag: "stable"})))
			},
		},
		{
			name: "rank checks",
			path: "/api/v1/keywords/" + keywordID + "/rank-checks",
			response: pagerResponse{
				prefix: PublicIDPrefixCheck,
				fields: map[string]any{"keyword_id": strictID(PublicIDPrefixKeyword)},
			},
			iterate: func(t *testing.T, client *Client, ctx context.Context) int {
				return len(collectPager(t, client.IterateRankChecks(ctx, keywordID, nil)))
			},
		},
		{
			name: "signals",
			path: "/api/v1/projects/" + projectID + "/signals",
			response: pagerResponse{
				prefix: PublicIDPrefixSignal,
				fields: map[string]any{
					"project_id": strictID(PublicIDPrefixProject),
					"public_id":  strictID(PublicIDPrefixSignal),
				},
			},
			iterate: func(t *testing.T, client *Client, ctx context.Context) int {
				return len(collectPager(t, client.IterateSignals(ctx, projectID, nil)))
			},
		},
		{
			name:     "alerts",
			path:     "/api/v1/projects/" + projectID + "/alert-rules",
			response: pagerResponse{prefix: PublicIDPrefixRule},
			iterate: func(t *testing.T, client *Client, ctx context.Context) int {
				return len(collectPager(t, client.IterateAlertRules(ctx, projectID, nil)))
			},
		},
		{
			name:     "triggered alerts",
			path:     "/api/v1/projects/" + projectID + "/triggered-alerts",
			response: pagerResponse{prefix: PublicIDPrefixAlert},
			iterate: func(t *testing.T, client *Client, ctx context.Context) int {
				return len(collectPager(t, client.IterateTriggeredAlerts(ctx, projectID, nil)))
			},
		},
		{
			name:     "members",
			path:     "/api/v1/projects/" + projectID + "/team/members",
			response: pagerResponse{prefix: PublicIDPrefixMember},
			iterate: func(t *testing.T, client *Client, ctx context.Context) int {
				return len(collectPager(t, client.IterateTeamMembers(ctx, projectID, nil)))
			},
		},
		{
			name:     "invites",
			path:     "/api/v1/projects/" + projectID + "/team/invites",
			response: pagerResponse{prefix: PublicIDPrefixInvite},
			iterate: func(t *testing.T, client *Client, ctx context.Context) int {
				return len(collectPager(t, client.IterateTeamInvites(ctx, projectID, nil)))
			},
		},
		{
			name:     "providers",
			path:     "/api/v1/projects/" + projectID + "/providers",
			response: pagerResponse{},
			iterate: func(t *testing.T, client *Client, ctx context.Context) int {
				return len(collectPager(t, client.IterateProviders(ctx, projectID, nil)))
			},
		},
		{
			name:     "saved keywords",
			path:     "/api/v1/projects/" + projectID + "/saved-keywords",
			response: pagerResponse{prefix: PublicIDPrefixSKW},
			iterate: func(t *testing.T, client *Client, ctx context.Context) int {
				return len(collectPager(t, client.IterateSavedKeywords(ctx, projectID, nil)))
			},
		},
		{
			name:     "views",
			path:     "/api/v1/projects/" + projectID + "/saved-views",
			response: pagerResponse{prefix: PublicIDPrefixView},
			iterate: func(t *testing.T, client *Client, ctx context.Context) int {
				return len(collectPager(t, client.IterateSavedViews(ctx, projectID, nil)))
			},
		},
		{
			name:     "competitors",
			path:     "/api/v1/projects/" + projectID + "/competitors",
			response: pagerResponse{prefix: PublicIDPrefixComp},
			iterate: func(t *testing.T, client *Client, ctx context.Context) int {
				return len(collectPager(t, client.IterateCompetitors(ctx, projectID, nil)))
			},
		},
		{
			name:     "migration tokens",
			path:     "/api/v1/projects/" + projectID + "/migration-tokens",
			response: pagerResponse{prefix: PublicIDPrefixMToken},
			iterate: func(t *testing.T, client *Client, ctx context.Context) int {
				return len(collectPager(t, client.IterateMigrationTokens(ctx, projectID, nil)))
			},
		},
	}
	server := httptest.NewServer(pagerTestServer(t, pagerTestResponses(cases)))
	defer server.Close()
	client := newTestClient(t, server.URL+"/api/v1")
	ctx := context.Background()

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := testCase.iterate(t, client, ctx); got != 2 {
				t.Fatalf("items = %d, want 2", got)
			}
		})
	}
}

func TestPagerSkipsEmptyPagesAndReportsErrors(t *testing.T) {
	calls := 0
	pager := newPager(context.Background(), "", func(context.Context, string) ([]int, *string, error) {
		calls++
		if calls == 1 {
			next := "next"
			return nil, &next, nil
		}
		return []int{42}, nil, nil
	})
	if !pager.Next() || pager.Item() != 42 || pager.Next() || pager.Err() != nil {
		t.Fatalf("unexpected pager state")
	}
	want := errors.New("page failed")
	failed := newPager(context.Background(), "", func(context.Context, string) ([]int, *string, error) {
		return nil, nil, want
	})
	if failed.Next() || !errors.Is(failed.Err(), want) {
		t.Fatalf("pager error = %v", failed.Err())
	}
}

func TestRetryPolicy(t *testing.T) {
	t.Run("GET retries 503 and honors Retry-After", func(t *testing.T) {
		var calls atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if calls.Add(1) == 1 {
				w.Header().Set("Retry-After", "0")
				writeJSON(t, w, http.StatusServiceUnavailable, map[string]any{"type": "busy"})
				return
			}
			writeJSON(t, w, http.StatusOK, map[string]any{"data": []any{}, "meta": map[string]any{"next_cursor": nil}})
		}))
		defer server.Close()
		client := newTestClient(t, server.URL+"/api/v1")
		if _, err := client.ListProjects(context.Background()); err != nil {
			t.Fatal(err)
		}
		if calls.Load() != 2 {
			t.Fatalf("calls = %d", calls.Load())
		}
	})

	t.Run("POST needs an idempotency key", func(t *testing.T) {
		var calls atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			w.Header().Set("Retry-After", "0")
			writeJSON(t, w, http.StatusServiceUnavailable, map[string]any{"type": "busy"})
		}))
		defer server.Close()
		client := newTestClient(t, server.URL+"/api/v1")
		_, _ = client.CreateAPIKey(context.Background(), CreateAPIKeyInput{Name: "CI"})
		if calls.Load() != 1 {
			t.Fatalf("calls without key = %d", calls.Load())
		}
		calls.Store(0)
		_, _ = client.CreateAPIKey(context.Background(), CreateAPIKeyInput{Name: "CI"}, WithIdempotencyKey("idem"))
		if calls.Load() != 3 {
			t.Fatalf("calls with key = %d", calls.Load())
		}
	})

	t.Run("cancellation interrupts retry sleep", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			cancel()
			writeJSON(t, w, http.StatusServiceUnavailable, map[string]any{"type": "busy"})
		}))
		defer server.Close()
		client := newTestClient(t, server.URL+"/api/v1")
		_, err := client.ListProjects(ctx)
		var networkErr *NetworkError
		if !errors.As(err, &networkErr) || !errors.Is(err, context.Canceled) {
			t.Fatalf("err = %v", err)
		}
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func TestNetworkRetryAndErrorContract(t *testing.T) {
	var calls atomic.Int32
	transport := roundTripFunc(func(*http.Request) (*http.Response, error) {
		if calls.Add(1) == 1 {
			return nil, fmt.Errorf("offline")
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"data":[],"meta":{"next_cursor":null}}`))}, nil
	})
	client, err := NewClient(WithAPIKey(testAPIKey), WithHTTPClient(&http.Client{Transport: transport}), WithMaxRetries(1))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListProjects(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d", calls.Load())
	}
	if _, err := NewClient(WithMaxRetries(-1)); err == nil {
		t.Fatal("expected invalid retries")
	}

	apiErr := &APIError{Header: http.Header{"Retry-After": []string{"120"}}, StatusCode: http.StatusTooManyRequests}
	if !apiErr.IsRateLimit() || apiErr.IsNotFound() {
		t.Fatal("rate limit helpers")
	}
	if delay, ok := apiErr.RetryAfter(); !ok || delay.String() != "1m0s" {
		t.Fatalf("RetryAfter = %v, %v", delay, ok)
	}
	var base BisibilityError
	if !errors.As(error(apiErr), &base) {
		t.Fatal("APIError does not implement BisibilityError")
	}
}

func TestAPIErrorRedactsSensitiveHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		for _, name := range []string{"Authorization", "Cookie", "Proxy-Authorization", "Set-Cookie", "X-Api-Key", "X-Auth-Token"} {
			w.Header().Set(name, "secret")
		}
		writeJSON(t, w, http.StatusNotFound, map[string]any{"type": "missing", "status": "wrong", "trace_id": "trace_1"})
	}))
	defer server.Close()
	client := newTestClient(t, server.URL+"/api/v1")
	_, err := client.GetKeyword(context.Background(), "kw_z00000000000000000000000")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatal(err)
	}
	for name := range apiErr.Header {
		if apiErr.Header.Get(name) == "secret" {
			t.Fatalf("header %s was not redacted", name)
		}
	}
	if !apiErr.IsNotFound() {
		t.Fatal("not-found helper")
	}
	if string(apiErr.Problem.Extensions["trace_id"]) != `"trace_1"` || apiErr.Problem.Status != 0 {
		t.Fatalf("problem extensions = %v, status = %d", apiErr.Problem.Extensions, apiErr.Problem.Status)
	}
}
