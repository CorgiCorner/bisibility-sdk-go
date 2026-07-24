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

func TestResourcePagers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/projects/prj/keywords" && r.URL.Query().Get("filter[tag]") != "stable" {
			t.Errorf("keyword tag filter was not preserved")
		}
		next := any(nil)
		id := "second"
		if r.URL.Query().Get("cursor") != "next" {
			next = "next"
			id = "first"
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"data": []map[string]any{{"id": id}},
			"meta": map[string]any{"next_cursor": next},
		})
	}))
	defer server.Close()
	client := newTestClient(t, server.URL+"/api/v1")
	ctx := context.Background()

	if got := len(collectPager(t, client.IterateAPIKeys(ctx, nil))); got != 2 {
		t.Fatalf("API keys = %d", got)
	}
	if got := len(collectPager(t, client.IterateProjectAPIKeys(ctx, "prj", nil))); got != 2 {
		t.Fatalf("project API keys = %d", got)
	}
	if got := len(collectPager(t, client.IterateWebhooks(ctx, "prj", nil))); got != 2 {
		t.Fatalf("webhooks = %d", got)
	}
	if got := len(collectPager(t, client.IterateKeywords(ctx, "prj", &ListKeywordsOptions{Tag: "stable"}))); got != 2 {
		t.Fatalf("keywords = %d", got)
	}
	if got := len(collectPager(t, client.IterateRankChecks(ctx, "kw", nil))); got != 2 {
		t.Fatalf("rank checks = %d", got)
	}
	if got := len(collectPager(t, client.IterateSignals(ctx, "prj", nil))); got != 2 {
		t.Fatalf("signals = %d", got)
	}
	if got := len(collectPager(t, client.IterateAlertRules(ctx, "prj", nil))); got != 2 {
		t.Fatalf("alerts = %d", got)
	}
	if got := len(collectPager(t, client.IterateTriggeredAlerts(ctx, "prj", nil))); got != 2 {
		t.Fatalf("triggered alerts = %d", got)
	}
	if got := len(collectPager(t, client.IterateTeamMembers(ctx, "prj", nil))); got != 2 {
		t.Fatalf("members = %d", got)
	}
	if got := len(collectPager(t, client.IterateTeamInvites(ctx, "prj", nil))); got != 2 {
		t.Fatalf("invites = %d", got)
	}
	if got := len(collectPager(t, client.IterateProviders(ctx, "prj", nil))); got != 2 {
		t.Fatalf("providers = %d", got)
	}
	if got := len(collectPager(t, client.IterateSavedViews(ctx, "prj", nil))); got != 2 {
		t.Fatalf("views = %d", got)
	}
	if got := len(collectPager(t, client.IterateCompetitors(ctx, "prj", nil))); got != 2 {
		t.Fatalf("competitors = %d", got)
	}
	if got := len(collectPager(t, client.IterateMigrationTokens(ctx, "prj", nil))); got != 2 {
		t.Fatalf("migration tokens = %d", got)
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
	_, err := client.GetKeyword(context.Background(), "missing")
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
