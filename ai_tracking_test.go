package bisibility

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAITrackingTransport(t *testing.T) {
	project := "prj_a00000000000000000000000"
	prompt := "aip_a00000000000000000000000"
	run := "air_a00000000000000000000000"
	count := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		if r.Method == "POST" {
			if r.Header.Get("Idempotency-Key") != "stable" {
				t.Error("missing retry identity")
			}
			body, _ := io.ReadAll(r.Body)
			var value map[string]any
			_ = json.Unmarshal(body, &value)
			if _, ok := value["idempotency_key"]; ok {
				t.Error("key leaked into launch body")
			}
			writeJSON(t, w, 201, map[string]any{"data": map[string]any{"id": run, "state": "planned"}})
		} else {
			if r.URL.Query().Get("cursor") != "opaque+/=" {
				t.Error("cursor changed")
			}
			writeJSON(t, w, 200, map[string]any{"data": []any{map[string]any{"id": "asm_a00000000000000000000000", "prompt_revision_id": "apr_a00000000000000000000000", "measurement": "unknown", "cost_usd": nil, "cost_state": "unknown", "evidence": nil}}, "meta": map[string]any{"next_cursor": "opaque+/="}})
		}
	}))
	defer server.Close()
	c := newTestClient(t, server.URL+"/api/v1")
	ctx := context.Background()
	plan := AITrackingRunInput{AITrackingPreviewInput: AITrackingPreviewInput{PromptIDs: []string{prompt}}, Consent: true}
	if _, err := c.CreateAITrackingRun(ctx, project, plan, WithIdempotencyKey("stable")); err != nil {
		t.Fatal(err)
	}
	result, err := c.ListAITrackingSamples(ctx, project, run, &PaginationOptions{Cursor: "opaque+/="})
	if err != nil || result.Data[0].CostUSD != nil || result.Data[0].Evidence != nil || result.Data[0].Measurement != "unknown" {
		t.Fatalf("lost unknown state: %v %v", result, err)
	}
	if _, err := c.CreateAITrackingRun(ctx, project, plan); err == nil {
		t.Error("missing identity accepted")
	}
	if _, err := c.GetAITrackingRun(ctx, project, prompt); err == nil {
		t.Error("wrong namespace accepted")
	}
	if count != 2 {
		t.Fatalf("invalid inputs reached transport: %d", count)
	}
}

func TestActualCostPromptJSON(t *testing.T) {
	zero := 0
	input := CompareAIPromptsOptions{AIResearchInput: AIResearchInput{Brand: "Example", Domain: "example.com", EstimateOnly: true}, Prompt: "Example?", Models: []string{"gpt-5-mini"}, CostPolicy: "provider_actual_cost", ActualCostAcknowledgement: "non_guaranteed_estimate_v1", EstimatedCostLimitCents: &zero}
	if err := validatePromptCostPolicy(input); err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	_ = json.Unmarshal(body, &value)
	if _, found := value["max_cost_cents"]; found {
		t.Fatal("actual cost serialized a guaranteed cap")
	}
	input.EstimateOnly = false
	if err := validatePromptCostPolicy(input); err == nil {
		t.Fatal("execution accepted missing estimate identity")
	}
}
