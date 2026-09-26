package bisibility

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestGetProjectDefaults(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.Method, http.MethodGet)
		assertEqual(t, r.RequestURI, "/api/v1/projects/prj_a00000000000000000000000/defaults")
		assertEqual(t, r.Header.Get("X-Trace-ID"), "trace_defaults")

		writeJSON(t, w, http.StatusOK, map[string]any{
			"city":               "New York",
			"country":            "United States",
			"cron_expression":    "0 9 * * 1",
			"device":             "mobile",
			"frequency":          "weekly",
			"jitter_minutes":     15,
			"last_checked_at":    "2026-07-24T08:00:00Z",
			"location_key":       "US/New York/New York",
			"next_check_at":      "2026-07-31T08:00:00Z",
			"project_id":         "prj_a00000000000000000000000",
			"serp_depth":         50,
			"serp_stop_on_match": true,
			"source":             "explicit",
			"timezone":           "America/New_York",
			"updated_at":         "2026-07-24T09:00:00Z",
		})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	defaults, err := client.GetProjectDefaults(
		context.Background(),
		"prj_a00000000000000000000000",
		WithRequestHeader("X-Trace-ID", "trace_defaults"),
	)
	if err != nil {
		t.Fatalf("GetProjectDefaults returned error: %v", err)
	}

	assertEqual(t, defaults.ProjectID, "prj_a00000000000000000000000")
	assertEqual(t, *defaults.City, "New York")
	assertEqual(t, defaults.Country, "United States")
	assertEqual(t, *defaults.CronExpression, "0 9 * * 1")
	assertEqual(t, defaults.Device, DeviceMobile)
	assertEqual(t, defaults.Frequency, RankCheckFrequencyWeekly)
	assertEqual(t, defaults.JitterMinutes, 15)
	assertTimePointerEqual(t, defaults.LastCheckedAt, "2026-07-24T08:00:00Z")
	assertEqual(t, defaults.LocationKey, "US/New York/New York")
	assertTimePointerEqual(t, defaults.NextCheckAt, "2026-07-31T08:00:00Z")
	assertEqual(t, defaults.SerpDepth, 50)
	assertEqual(t, defaults.SerpStopOnMatch, true)
	assertEqual(t, defaults.Source, ProjectDefaultsSourceExplicit)
	assertEqual(t, defaults.Timezone, "America/New_York")
	assertTimePointerEqual(t, defaults.UpdatedAt, "2026-07-24T09:00:00Z")
}

func TestUpdateProjectDefaultsKeepsStoredSerpDepthWhenOmitted(t *testing.T) {
	t.Parallel()

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		assertEqual(t, r.Method, http.MethodPatch)
		assertEqual(t, r.RequestURI, "/api/v1/projects/prj_a00000000000000000000000/defaults")

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode defaults patch: %v", err)
		}
		if _, ok := body["serp_depth"]; ok {
			t.Fatal("defaults PATCH sent serp_depth for an omitted depth")
		}
		if _, ok := body["serp_stop_on_match"]; ok {
			t.Fatal("defaults PATCH sent serp_stop_on_match for an omitted value")
		}

		writeJSON(t, w, http.StatusOK, map[string]any{
			"country":            "United States",
			"device":             "desktop",
			"frequency":          "daily",
			"jitter_minutes":     60,
			"location_key":       "US/Texas/Austin",
			"project_id":         "prj_a00000000000000000000000",
			"serp_depth":         50,
			"serp_stop_on_match": true,
			"source":             "explicit",
			"timezone":           "UTC",
		})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	defaults, err := client.UpdateProjectDefaults(
		context.Background(),
		"prj_a00000000000000000000000",
		ProjectDefaultsPatch{Frequency: RankCheckFrequencyDaily, LocationKey: "US/Texas/Austin"},
	)
	if err != nil {
		t.Fatalf("UpdateProjectDefaults returned error: %v", err)
	}
	assertEqual(t, requests, 1)
	assertEqual(t, defaults.SerpDepth, 50)
	assertEqual(t, defaults.SerpStopOnMatch, true)
}

func TestUpdateProjectDefaultsSendsEverySupportedSerpDepth(t *testing.T) {
	t.Parallel()

	for _, depth := range SerpDepths {
		depth := depth
		t.Run(strconv.Itoa(depth), func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode defaults patch: %v", err)
				}
				assertEqual(t, body["serp_depth"], any(float64(depth)))
				writeJSON(t, w, http.StatusOK, map[string]any{
					"country":            "United States",
					"device":             "desktop",
					"frequency":          "daily",
					"jitter_minutes":     60,
					"location_key":       "US/Texas/Austin",
					"project_id":         "prj_a00000000000000000000000",
					"serp_depth":         depth,
					"serp_stop_on_match": true,
					"source":             "explicit",
					"timezone":           "UTC",
				})
			}))
			defer server.Close()

			client := newTestClient(t, server.URL+"/api/v1")
			defaults, err := client.UpdateProjectDefaults(
				context.Background(),
				"prj_a00000000000000000000000",
				ProjectDefaultsPatch{
					Frequency:   RankCheckFrequencyDaily,
					LocationKey: "US/Texas/Austin",
					SerpDepth:   &depth,
				},
			)
			if err != nil {
				t.Fatalf("UpdateProjectDefaults returned error: %v", err)
			}
			assertEqual(t, defaults.SerpDepth, depth)
		})
	}
}

func TestGetProjectDefaultsMapsForbidden(t *testing.T) {
	t.Parallel()

	problem := ProblemDetails{
		Type:   "https://bisibility.dev/problems/forbidden",
		Title:  "Forbidden",
		Status: http.StatusForbidden,
		Detail: "You cannot read defaults for this project.",
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.Method, http.MethodGet)
		assertEqual(t, r.RequestURI, "/api/v1/projects/prj_a00000000000000000000000/defaults")
		writeJSON(t, w, http.StatusForbidden, problem)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	_, err := client.GetProjectDefaults(context.Background(), "prj_a00000000000000000000000")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %T, want APIError", err)
	}
	assertEqual(t, apiErr.StatusCode, http.StatusForbidden)
	if apiErr.Problem == nil {
		t.Fatal("problem = nil, want decoded ProblemDetails")
	}
	assertEqual(t, apiErr.Problem.Type, problem.Type)
	assertEqual(t, apiErr.Problem.Title, problem.Title)
	assertEqual(t, apiErr.Problem.Status, problem.Status)
	assertEqual(t, apiErr.Problem.Detail, problem.Detail)
}

func assertTimePointerEqual(t *testing.T, got *time.Time, want string) {
	t.Helper()

	if got == nil {
		t.Fatalf("time = nil, want %s", want)
	}
	assertEqual(t, got.UTC().Format(time.RFC3339), want)
}
