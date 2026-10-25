package bisibility

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPromptTrackerResourceRoundTrips(t *testing.T) {
	const project = "prj_a00000000000000000000000"
	const topic = "ait_a00000000000000000000000"
	const prompt = "aip_a00000000000000000000000"
	const schedule = "ais_a00000000000000000000000"
	const run = "air_a00000000000000000000000"
	const previous = "air_b00000000000000000000000"
	name := "Renamed topic"
	pagination := &PaginationOptions{Cursor: "opaque+/=", Limit: 2}
	plan := PromptTrackerRunInput{PromptTrackerPreviewInput: PromptTrackerPreviewInput{PromptIDs: []string{prompt}, Configurations: []PromptTrackerSourceConfiguration{}}, Consent: true}
	listQuery := "cursor=opaque%2B%2F%3D&limit=2"
	listResponse := map[string]any{"data": []any{}, "meta": map[string]any{"next_cursor": "next-page"}}
	item := func(id string) any { return map[string]any{"data": map[string]any{"id": id}} }
	tests := []struct {
		name, method, path, query, token string
		body                             map[string]any
		response                         any
		call                             func(context.Context, *Client) (any, error)
	}{
		{"list topics", "GET", "topics", listQuery, "next-page", nil, listResponse,
			func(ctx context.Context, c *Client) (any, error) {
				return c.ListPromptTrackerTopics(ctx, project, pagination)
			}},
		{"create topic", "POST", "topics", "", topic, map[string]any{"name": "Topic"}, item(topic),
			func(ctx context.Context, c *Client) (any, error) {
				return c.CreatePromptTrackerTopic(ctx, project, PromptTrackerTopicInput{Name: "Topic"})
			}},
		{"update topic", "PATCH", "topics/" + topic, "", topic, map[string]any{"name": name}, item(topic),
			func(ctx context.Context, c *Client) (any, error) {
				return c.UpdatePromptTrackerTopic(ctx, project, topic, PromptTrackerTopicPatch{Name: &name})
			}},
		{"archive topic", "DELETE", "topics/" + topic, "", topic, nil, item(topic),
			func(ctx context.Context, c *Client) (any, error) {
				return c.ArchivePromptTrackerTopic(ctx, project, topic)
			}},
		{"list prompts", "GET", "prompts", listQuery, "next-page", nil, listResponse,
			func(ctx context.Context, c *Client) (any, error) {
				return c.ListPromptTrackerPrompts(ctx, project, pagination)
			}},
		{"create prompt", "POST", "prompts", "", prompt, map[string]any{"text": "Where is Example?", "topic_id": topic}, item(prompt),
			func(ctx context.Context, c *Client) (any, error) {
				return c.CreatePromptTrackerPrompt(ctx, project, PromptTrackerPromptInput{Text: "Where is Example?", TopicID: strPtr(topic)})
			}},
		{"update prompt", "PATCH", "prompts/" + prompt, "", prompt, map[string]any{"text": name}, item(prompt),
			func(ctx context.Context, c *Client) (any, error) {
				return c.UpdatePromptTrackerPrompt(ctx, project, prompt, PromptTrackerPromptPatch{Text: &name})
			}},
		{"archive prompt", "DELETE", "prompts/" + prompt, "", prompt, nil, item(prompt),
			func(ctx context.Context, c *Client) (any, error) {
				return c.ArchivePromptTrackerPrompt(ctx, project, prompt)
			}},
		{"list schedules", "GET", "schedules", listQuery, "next-page", nil, listResponse,
			func(ctx context.Context, c *Client) (any, error) {
				return c.ListPromptTrackerSchedules(ctx, project, pagination)
			}},
		{"create schedule", "POST", "schedules", "", schedule, map[string]any{"name": "Weekly", "cron": "0 9 * * 1", "timezone": "UTC", "consent": true}, item(schedule),
			func(ctx context.Context, c *Client) (any, error) {
				return c.CreatePromptTrackerSchedule(ctx, project, PromptTrackerScheduleInput{Name: "Weekly", Cron: "0 9 * * 1", Timezone: "UTC", Configuration: plan, Consent: true})
			}},
		{"archive schedule", "DELETE", "schedules/" + schedule, "", schedule, nil, item(schedule),
			func(ctx context.Context, c *Client) (any, error) {
				return c.ArchivePromptTrackerSchedule(ctx, project, schedule)
			}},
		{"preview run", "POST", "runs/preview", "", "1.5", map[string]any{"prompt_ids": []any{prompt}}, map[string]any{"data": map[string]any{"estimated_cost_cents": 1.5, "credential_connection_id": "conn_a00000000000000000000000"}},
			func(ctx context.Context, c *Client) (any, error) {
				return c.PreviewPromptTrackerRun(ctx, project, plan.PromptTrackerPreviewInput)
			}},
		{"list runs", "GET", "runs", listQuery, "next-page", nil, listResponse,
			func(ctx context.Context, c *Client) (any, error) {
				return c.ListPromptTrackerRuns(ctx, project, pagination)
			}},
		{"get run", "GET", "runs/" + run, "", run, nil, item(run),
			func(ctx context.Context, c *Client) (any, error) { return c.GetPromptTrackerRun(ctx, project, run) }},
		{"cancel run", "POST", "runs/" + run + "/cancel", "", run, nil, item(run),
			func(ctx context.Context, c *Client) (any, error) { return c.CancelPromptTrackerRun(ctx, project, run) }},
		{"retry run", "POST", "runs/" + run + "/retry", "", run, map[string]any{"consent": true}, item(run),
			func(ctx context.Context, c *Client) (any, error) {
				return c.RetryPromptTrackerRun(ctx, project, run, plan, WithIdempotencyKey("retry-key"))
			}},
		{"history", "GET", "history", listQuery, "next-page", nil, listResponse,
			func(ctx context.Context, c *Client) (any, error) {
				return c.GetPromptTrackerHistory(ctx, project, pagination)
			}},
		{"trends", "GET", "trends", listQuery + "&previous_run_id=" + previous + "&run_id=" + run, "different_engine", nil, map[string]any{"data": map[string]any{"comparable": false, "reason": "different_engine"}},
			func(ctx context.Context, c *Client) (any, error) {
				return c.GetPromptTrackerTrends(ctx, project, &PromptTrackerTrendOptions{PromptTrackerEvidenceOptions: PromptTrackerEvidenceOptions{RunID: run, Limit: 2, Cursor: "opaque+/="}, PreviousRunID: previous})
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				assertEqual(t, r.Method, tt.method)
				assertEqual(t, r.URL.Path, "/api/v1/projects/"+project+"/prompt-tracker/"+tt.path)
				assertEqual(t, r.URL.RawQuery, tt.query)
				assertEqual(t, r.Header.Get("Authorization"), "Bearer "+testAPIKey)
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}
				if tt.body == nil {
					assertEqual(t, string(body), "")
				} else {
					var decoded map[string]any
					if err := json.Unmarshal(body, &decoded); err != nil {
						t.Fatal(err)
					}
					for key, want := range tt.body {
						gotJSON, _ := json.Marshal(decoded[key])
						wantJSON, _ := json.Marshal(want)
						assertEqual(t, string(gotJSON), string(wantJSON))
					}
				}
				writeJSON(t, w, http.StatusOK, tt.response)
			}))
			defer server.Close()
			got, err := tt.call(context.Background(), newTestClient(t, server.URL+"/api/v1"))
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(decoded), tt.token) {
				t.Fatalf("response lost %q: %s", tt.token, decoded)
			}
			assertEqual(t, requests, 1)
		})
	}
}
