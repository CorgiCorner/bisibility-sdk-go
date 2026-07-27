package bisibility

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGetProjectOverview(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.Method, http.MethodGet)
		assertEqual(t, r.RequestURI, "/api/v1/projects/prj%2F%20one/overview?device=mobile&range=90d&tag=priority+tag")
		writeJSON(t, w, http.StatusOK, map[string]any{
			"project_id":                "prj/ one",
			"tracked_keyword_count":     0,
			"keywords_added_this_month": 0,
			"average_position":          nil,
			"average_position_delta":    nil,
			"top_3_count":               nil,
			"top_10_count":              0,
			"top_10_delta":              0,
			"top_100_count":             7,
			"visibility":                0.0,
			"visibility_delta":          -2.5,
			"position_distribution": []map[string]any{
				{"min": 1, "max": 3, "count": 0},
				{"min": 4, "max": 10, "count": nil},
			},
			"last_check_at": "2026-07-27T08:00:00Z",
			"next_check_at": nil,
		})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	overview, err := client.GetProjectOverview(context.Background(), "prj/ one", &ProjectOverviewOptions{
		Device: ProjectOverviewDeviceMobile,
		Range:  ProjectOverviewRange90Days,
		Tag:    "priority tag",
	})
	if err != nil {
		t.Fatalf("GetProjectOverview returned error: %v", err)
	}

	assertEqual(t, overview.ProjectID, "prj/ one")
	assertEqual(t, overview.TrackedKeywordCount, 0)
	assertEqual(t, overview.KeywordsAddedThisMonth, 0)
	if overview.AveragePosition != nil || overview.AveragePositionDelta != nil {
		t.Fatal("average positions = non-nil, want nil")
	}
	assertEqual(t, *overview.Top10Count, 0)
	assertEqual(t, *overview.Top10Delta, 0)
	assertEqual(t, *overview.Top100Count, 7)
	assertEqual(t, *overview.Visibility, 0.0)
	assertEqual(t, *overview.VisibilityDelta, -2.5)
	assertEqual(t, *overview.PositionDistribution[0].Count, 0)
	if overview.PositionDistribution[1].Count != nil {
		t.Fatalf("position_distribution[1].count = %v, want nil", *overview.PositionDistribution[1].Count)
	}
	assertTimePointerEqual(t, overview.LastCheckAt, "2026-07-27T08:00:00Z")
	if overview.NextCheckAt != nil {
		t.Fatalf("next_check_at = %v, want nil", overview.NextCheckAt)
	}
}

func TestGetProjectOverviewMapsForbidden(t *testing.T) {
	t.Parallel()

	problem := ProblemDetails{Type: "https://bisibility.dev/problems/forbidden", Title: "Forbidden", Status: http.StatusForbidden, Detail: "You cannot read this project overview."}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.Method, http.MethodGet)
		assertEqual(t, r.RequestURI, "/api/v1/projects/prj_1/overview")
		writeJSON(t, w, http.StatusForbidden, problem)
	}))
	defer server.Close()

	_, err := newTestClient(t, server.URL+"/api/v1").GetProjectOverview(context.Background(), "prj_1", nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %T, want APIError", err)
	}
	assertEqual(t, apiErr.StatusCode, http.StatusForbidden)
	assertEqual(t, apiErr.Problem.Detail, problem.Detail)
}

func TestProjectOverviewJSONTagsMatchContract(t *testing.T) {
	t.Parallel()

	assertStructFieldNamesEqual(t, reflect.TypeOf(ProjectOverviewOptions{}), []string{"Device", "Range", "Tag"})
	assertJSONTagsEqual(t, reflect.TypeOf(ProjectOverview{}), []string{
		"average_position", "average_position_delta", "keywords_added_this_month", "last_check_at", "next_check_at", "position_distribution", "project_id", "top_10_count", "top_10_delta", "top_100_count", "top_3_count", "tracked_keyword_count", "visibility", "visibility_delta",
	})
	assertJSONTagsEqual(t, reflect.TypeOf(ProjectOverviewPositionBucket{}), []string{"count", "max", "min"})
}

func assertStructFieldNamesEqual(t *testing.T, structType reflect.Type, want []string) {
	t.Helper()

	got := make([]string, structType.NumField())
	for index := range got {
		got[index] = structType.Field(index).Name
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s field names differ\ngot:  %v\nwant: %v", structType.Name(), got, want)
	}
}
