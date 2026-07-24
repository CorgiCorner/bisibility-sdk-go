package bisibility

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestExportRankHistoryJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.Method, http.MethodGet)
		assertEqual(t, r.URL.EscapedPath(), "/api/v1/projects/prj%20one/exports/rank-history")
		assertEqual(t, r.URL.RawQuery, "cursor=cursor+1&format=json&granularity=weekly&keyword_id=kw+1&keyword_id=kw%2F2&limit=2&range=90")
		writeJSON(t, w, http.StatusOK, map[string]any{
			"data": []any{map[string]any{
				"checked_at":        "2026-07-22T10:00:00Z",
				"id":                "check_1",
				"keyword":           "rank tracker api",
				"keyword_id":        "kw 1",
				"position":          4,
				"previous_position": 7,
				"ranking_url":       "https://example.com/rank-tracker",
			}},
			"meta": map[string]any{"next_cursor": "cursor_2"},
		})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	response, err := client.ExportRankHistory(context.Background(), "prj one", &ExportRankHistoryOptions{
		Cursor:      "cursor 1",
		Format:      RankHistoryExportFormatJSON,
		Granularity: RankHistoryGranularityWeekly,
		KeywordIDs:  []string{"kw 1", "kw/2"},
		Limit:       2,
		Range:       RankHistoryExportRange90Days,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, response.Format, RankHistoryExportFormatJSON)
	assertEqual(t, *response.Data[0].Position, 4)
	assertEqual(t, response.Data[0].CheckedAt.UTC().Format(time.RFC3339), "2026-07-22T10:00:00Z")
	assertEqual(t, *response.Meta.NextCursor, "cursor_2")
	assertEqual(t, response.CSV, "")
}

func TestExportRankHistoryCSV(t *testing.T) {
	t.Parallel()

	const csv = "id,keyword_id,keyword,checked_at,position,previous_position,ranking_url\ncheck_1,kw_1,rank tracker api,2026-07-22T10:00:00Z,4,7,https://example.com\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.URL.RawQuery, "format=csv&keyword_id=kw_1&range=all")
		w.Header().Set(contentTypeHeader, "text/csv")
		_, _ = w.Write([]byte(csv))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	response, err := client.ExportRankHistory(context.Background(), "prj_1", &ExportRankHistoryOptions{
		Format:     RankHistoryExportFormatCSV,
		KeywordIDs: []string{"kw_1"},
		Range:      RankHistoryExportRangeAll,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, response.Format, RankHistoryExportFormatCSV)
	assertEqual(t, response.CSV, csv)
	assertEqual(t, len(response.Data), 0)
}

func TestIterateRankHistoryPreservesJSONFilters(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		assertEqual(t, query.Get("format"), "json")
		assertEqual(t, query.Get("granularity"), "daily")
		assertEqual(t, query.Get("limit"), "1")
		assertEqual(t, query.Get("range"), "all")
		assertStringSlicesEqual(t, query["keyword_id"], []string{"kw_1", "kw_2"})

		id := "check_1"
		var next any = "cursor_2"
		if query.Get("cursor") == "cursor_2" {
			id = "check_2"
			next = nil
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"data": []any{map[string]any{
				"checked_at":        "2026-07-22T10:00:00Z",
				"id":                id,
				"keyword":           "rank tracker",
				"keyword_id":        "kw_1",
				"position":          nil,
				"previous_position": nil,
				"ranking_url":       nil,
			}},
			"meta": map[string]any{"next_cursor": next},
		})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	rows := collectPager(t, client.IterateRankHistory(context.Background(), "prj_1", &ExportRankHistoryOptions{
		Format:      RankHistoryExportFormatCSV,
		Granularity: RankHistoryGranularityDaily,
		KeywordIDs:  []string{"kw_1", "kw_2"},
		Limit:       1,
		Range:       RankHistoryExportRangeAll,
	}))
	assertEqual(t, len(rows), 2)
	assertEqual(t, rows[1].ID, "check_2")
	if rows[0].Position != nil || rows[0].PreviousPosition != nil || rows[0].RankingURL != nil {
		t.Fatal("nullable export fields were not preserved")
	}
}
