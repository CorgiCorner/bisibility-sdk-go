package bisibility

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestUpdateAITrackingScheduleIsPartialWithTopLevelConsent covers contract-audit
// finding #7: the PATCH must omit untouched fields and surface consent at the
// top level of the body, matching the API's scheduleInputSchema.partial() plus
// a top-level consent flag.
func TestUpdateAITrackingScheduleIsPartialWithTopLevelConsent(t *testing.T) {
	t.Run("rename-only patch sends just the new name and consent", func(t *testing.T) {
		var captured string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPatch {
				t.Errorf("method = %s, want PATCH", r.Method)
			}
			body := make([]byte, r.ContentLength)
			_, _ = r.Body.Read(body)
			captured = string(body)
			writeJSON(t, w, 200, map[string]any{"data": map[string]any{
				"id": "ais_a00000000000000000000000", "name": "Weekly refresh",
				"cron": "0 9 * * 1", "timezone": "UTC", "enabled": true,
				"configuration": map[string]any{"prompt_ids": []string{}, "configurations": []any{}},
				"next_run_at":   nil, "archived_at": nil,
				"created_at": "2026-10-10T00:00:00Z", "updated_at": "2026-10-10T00:00:00Z",
			}})
		}))
		defer server.Close()
		c := newTestClient(t, server.URL+"/api/v1")
		name := "Weekly refresh"
		if _, err := c.UpdateAITrackingSchedule(
			context.Background(),
			"prj_a00000000000000000000000",
			"ais_a00000000000000000000000",
			AITrackingSchedulePatch{Name: &name, Consent: true},
		); err != nil {
			t.Fatal(err)
		}
		var value map[string]json.RawMessage
		if err := json.Unmarshal([]byte(captured), &value); err != nil {
			t.Fatalf("captured body not JSON: %q", captured)
		}
		if _, found := value["cron"]; found {
			t.Error("rename-only patch leaked cron field")
		}
		if _, found := value["configuration"]; found {
			t.Error("rename-only patch leaked configuration field")
		}
		if _, found := value["enabled"]; found {
			t.Error("rename-only patch leaked enabled field")
		}
		if string(value["name"]) != `"Weekly refresh"` {
			t.Fatalf("name = %s, want \"Weekly refresh\"", value["name"])
		}
		if string(value["consent"]) != `true` {
			t.Fatalf("consent not at top level of body: %s", captured)
		}
	})

	t.Run("clearing next_run_at sends explicit null", func(t *testing.T) {
		var captured string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body := make([]byte, r.ContentLength)
			_, _ = r.Body.Read(body)
			captured = string(body)
			writeJSON(t, w, 200, map[string]any{"data": map[string]any{
				"id": "ais_a00000000000000000000000", "name": "n",
				"cron": "0 9 * * 1", "timezone": "UTC", "enabled": true,
				"configuration": map[string]any{"prompt_ids": []string{}, "configurations": []any{}},
				"next_run_at":   nil, "archived_at": nil,
				"created_at": "2026-10-10T00:00:00Z", "updated_at": "2026-10-10T00:00:00Z",
			}})
		}))
		defer server.Close()
		c := newTestClient(t, server.URL+"/api/v1")
		if _, err := c.UpdateAITrackingSchedule(
			context.Background(),
			"prj_a00000000000000000000000",
			"ais_a00000000000000000000000",
			AITrackingSchedulePatch{NextRunAt: NullString(), Consent: true},
		); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(captured, `"next_run_at":null`) {
			t.Fatalf("expected next_run_at:null in body, got %q", captured)
		}
	})

	t.Run("enabling requires reviewed configuration consent at top level", func(t *testing.T) {
		var captured string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body := make([]byte, r.ContentLength)
			_, _ = r.Body.Read(body)
			captured = string(body)
			writeJSON(t, w, 200, map[string]any{"data": map[string]any{
				"id": "ais_a00000000000000000000000", "name": "n",
				"cron": "0 9 * * 1", "timezone": "UTC", "enabled": true,
				"configuration": map[string]any{"prompt_ids": []string{}, "configurations": []any{}},
				"next_run_at":   nil, "archived_at": nil,
				"created_at": "2026-10-10T00:00:00Z", "updated_at": "2026-10-10T00:00:00Z",
			}})
		}))
		defer server.Close()
		c := newTestClient(t, server.URL+"/api/v1")
		enabled := true
		plan := &AITrackingRunInput{
			AITrackingPreviewInput: AITrackingPreviewInput{PromptIDs: []string{"aip_a00000000000000000000000"}},
			Consent:                true,
		}
		if _, err := c.UpdateAITrackingSchedule(
			context.Background(),
			"prj_a00000000000000000000000",
			"ais_a00000000000000000000000",
			AITrackingSchedulePatch{Enabled: &enabled, Configuration: plan, Consent: true},
		); err != nil {
			t.Fatal(err)
		}
		var value map[string]json.RawMessage
		if err := json.Unmarshal([]byte(captured), &value); err != nil {
			t.Fatalf("captured body not JSON: %q", captured)
		}
		if string(value["consent"]) != `true` {
			t.Fatalf("top-level consent missing: %s", captured)
		}
		if _, found := value["configuration"]; !found {
			t.Fatalf("configuration missing from patch: %s", captured)
		}
	})
}
