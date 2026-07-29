package bisibility

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestMatchProjectKeywords(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.Method, http.MethodPost)
		assertEqual(t, r.RequestURI, "/api/v1/projects/prj_a00000000000000000000000/keyword-matches")
		assertEqual(t, r.Header.Get("Content-Type"), "application/json")
		var input KeywordMatchRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if !reflect.DeepEqual(input.Texts, []string{" Headless CMS ", "SEO tool"}) {
			t.Fatalf("texts = %v, want request texts", input.Texts)
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"data": []map[string]any{
				{
					"keyword_id": "kw_a00000000000000000000000", "matched_text": "headless cms", "text": " Headless CMS ", "latest_position": 3, "previous_position": 7,
					"ranking_url": "https://example.com/headless-cms",
					"market":      map[string]any{"location": "Austin", "location_key": "US/Texas/Austin", "country_code": "US", "device": "mobile"},
				},
				{
					"keyword_id": "kw_b00000000000000000000000", "matched_text": "seo tool", "text": "SEO Tool", "latest_position": nil, "previous_position": 0,
					"ranking_url": nil,
					"market":      map[string]any{"location": "United States", "location_key": "US", "country_code": "US", "device": "desktop"},
				},
			},
			"meta": map[string]any{"truncated_texts": []string{"headless cms"}},
		})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	response, err := client.MatchProjectKeywords(context.Background(), "prj_a00000000000000000000000", KeywordMatchRequest{Texts: []string{" Headless CMS ", "SEO tool"}})
	if err != nil {
		t.Fatalf("MatchProjectKeywords returned error: %v", err)
	}

	assertEqual(t, len(response.Data), 2)
	assertEqual(t, response.Data[0].KeywordID, "kw_a00000000000000000000000")
	assertEqual(t, response.Data[0].MatchedText, "headless cms")
	assertEqual(t, response.Data[0].Text, " Headless CMS ")
	assertEqual(t, *response.Data[0].LatestPosition, 3)
	assertEqual(t, *response.Data[0].PreviousPosition, 7)
	assertEqual(t, *response.Data[0].RankingURL, "https://example.com/headless-cms")
	assertEqual(t, response.Data[0].Market.Location, "Austin")
	assertEqual(t, response.Data[0].Market.LocationKey, "US/Texas/Austin")
	assertEqual(t, response.Data[0].Market.CountryCode, "US")
	assertEqual(t, response.Data[0].Market.Device, DeviceMobile)
	if response.Data[1].LatestPosition != nil {
		t.Fatalf("latest_position = %v, want nil", *response.Data[1].LatestPosition)
	}
	if response.Data[1].RankingURL != nil {
		t.Fatalf("ranking_url = %q, want nil", *response.Data[1].RankingURL)
	}
	assertEqual(t, *response.Data[1].PreviousPosition, 0)
	if !reflect.DeepEqual(response.Meta.TruncatedTexts, []string{"headless cms"}) {
		t.Fatalf("truncated_texts = %v, want [headless cms]", response.Meta.TruncatedTexts)
	}
}

func TestMatchProjectKeywordsMapsForbidden(t *testing.T) {
	t.Parallel()

	problem := ProblemDetails{Type: "https://bisibility.dev/problems/forbidden", Title: "Forbidden", Status: http.StatusForbidden, Detail: "You cannot match this project."}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.Method, http.MethodPost)
		assertEqual(t, r.RequestURI, "/api/v1/projects/prj_a00000000000000000000000/keyword-matches")
		writeJSON(t, w, http.StatusForbidden, problem)
	}))
	defer server.Close()

	_, err := newTestClient(t, server.URL+"/api/v1").MatchProjectKeywords(context.Background(), "prj_a00000000000000000000000", KeywordMatchRequest{Texts: []string{"headless cms"}})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %T, want APIError", err)
	}
	assertEqual(t, apiErr.StatusCode, http.StatusForbidden)
	assertEqual(t, apiErr.Problem.Detail, problem.Detail)
}

func TestKeywordMatchJSONTagsMatchContract(t *testing.T) {
	t.Parallel()

	assertJSONTagsEqual(t, reflect.TypeOf(KeywordMatchRequest{}), []string{"texts"})
	assertJSONTagsEqual(t, reflect.TypeOf(KeywordMatchResponse{}), []string{"data", "meta"})
	assertJSONTagsEqual(t, reflect.TypeOf(KeywordMatchMeta{}), []string{"truncated_texts"})
	assertJSONTagsEqual(t, reflect.TypeOf(KeywordMatch{}), []string{"keyword_id", "latest_position", "market", "matched_text", "previous_position", "ranking_url", "text"})
	assertJSONTagsEqual(t, reflect.TypeOf(KeywordMatchMarket{}), []string{"country_code", "device", "location", "location_key"})
}

func TestKeywordMatchRankingURLJSONRoundTrip(t *testing.T) {
	t.Parallel()

	rankingURL := "https://example.com/headless-cms"
	encoded, err := json.Marshal(KeywordMatch{RankingURL: &rankingURL})
	if err != nil {
		t.Fatalf("marshal match with ranking URL: %v", err)
	}

	var decoded KeywordMatch
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal match with ranking URL: %v", err)
	}
	if decoded.RankingURL == nil {
		t.Fatal("ranking_url = nil, want URL")
	}
	assertEqual(t, *decoded.RankingURL, rankingURL)

	encoded, err = json.Marshal(KeywordMatch{RankingURL: nil})
	if err != nil {
		t.Fatalf("marshal match with null ranking URL: %v", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("unmarshal match fields: %v", err)
	}
	rankingURLJSON, ok := fields["ranking_url"]
	if !ok {
		t.Fatal("ranking_url is missing from serialized match")
	}
	assertEqual(t, string(rankingURLJSON), "null")
}
