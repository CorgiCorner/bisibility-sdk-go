package bisibility

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAITrackingExportScope(t *testing.T) {
	var result AITrackingExport
	if err := json.Unmarshal([]byte(`{"items":[],"run_id":"air_a00000000000000000000000","next_cursor":"resume","scope":{"complete":false,"max_pages":20,"loaded":1000,"resumed":true}}`), &result); err != nil {
		t.Fatal(err)
	}
	if result.Scope == nil || result.Scope.Complete || result.Scope.MaxPages != 20 || result.Scope.Loaded != 1000 || !result.Scope.Resumed {
		t.Fatalf("lost bounded export scope: %+v", result.Scope)
	}
	var evidence AITrackingEvidence
	if err := json.Unmarshal([]byte(`{"observed_at":null,"fetched_at":"2026-10-08T12:00:00Z"}`), &evidence); err != nil || evidence.ObservedAt != nil {
		t.Fatalf("fabricated unknown observation time: %+v %v", evidence, err)
	}
}

func TestAITrackingSuggestionGenerationContract(t *testing.T) {
	key := "11111111-1111-4111-8111-111111111111"
	input := AITrackingSuggestionsPreviewInput{
		Configuration: AITrackingSuggestionConfiguration{Provider: "dataforseo", Engine: "chat_gpt", Model: "gpt-5-mini", LanguageCode: "en", MaxOutputTokens: 1024, AdvisoryCostLimitCents: 50},
		InputSnapshot: AITrackingSuggestionSnapshot{Context: AITrackingReviewedContext{Business: "Example", Audience: "Teams", Products: "Analytics", Goals: "Visibility", AgentRules: "Review claims"}, Competitors: []AITrackingReviewedCompetitor{{ID: "cmp_a00000000000000000000000", Domain: "competitor.example.com"}}},
	}
	preview := AITrackingSuggestionsPreview{AITrackingSuggestionsPreviewInput: input, Version: 1, SnapshotHash: strings.Repeat("a", 64), EstimatedCostCents: 0.5, EstimateKind: "forecast", CredentialVersion: "v1", BudgetRevision: "b1", ConsentRevision: "c1", ExpiresAt: "2026-10-08T23:00:00Z", Limitations: []string{}}
	preview.CredentialConnectionID = "conn_a00000000000000000000000"
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if strings.HasSuffix(r.URL.Path, "/preview") {
			writeJSON(t, w, 200, map[string]any{"data": preview})
			return
		}
		if r.Header.Get("Idempotency-Key") != key {
			t.Error("missing stable UUID header")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if _, ok := body["idempotency_key"]; ok {
			t.Error("key leaked into generation body")
		}
		snapshot := body["preview"].(map[string]any)["input_snapshot"].(map[string]any)
		if snapshot["competitors"].([]any)[0].(map[string]any)["label"] != nil {
			t.Error("nullable label changed")
		}
		writeJSON(t, w, 200, map[string]any{"data": map[string]any{"generation_id": "asg_a00000000000000000000000", "drafts": []any{}, "cost_usd": nil, "cost_state": "unknown", "method": "model_generated_hypothesis", "limitations": []any{}}})
	}))
	defer server.Close()
	client := newTestClient(t, server.URL+"/api/v1")
	ctx := context.Background()
	project := "prj_a00000000000000000000000"
	if _, err := client.AITrackingSuggestionsPreview(ctx, project, input); err != nil {
		t.Fatal(err)
	}
	generation, err := client.AITrackingSuggestionsGenerate(ctx, project, AITrackingSuggestionsGenerateInput{Preview: preview, Consent: true}, WithIdempotencyKey(key))
	if err != nil || generation.Data.CostUSD != nil || generation.Data.Method != "model_generated_hypothesis" {
		t.Fatalf("lost model provenance/cost unknown: %v %v", generation, err)
	}
	if _, err := client.AITrackingSuggestionsGenerate(ctx, project, AITrackingSuggestionsGenerateInput{Preview: preview, Consent: true}); err == nil {
		t.Error("missing key accepted")
	}
	if _, err := client.AITrackingSuggestionsGenerate(ctx, project, AITrackingSuggestionsGenerateInput{Preview: preview}, WithIdempotencyKey(key)); err == nil {
		t.Error("missing consent accepted")
	}
	oversized := input
	oversized.InputSnapshot.Context.Business = strings.Repeat("x", 5000)
	if _, err := client.AITrackingSuggestionsPreview(ctx, project, oversized); err == nil {
		t.Error("oversized review accepted")
	}
	if calls != 2 {
		t.Fatalf("invalid requests reached transport: %d", calls)
	}
	reference := &AITrackingGenerationReference{GenerationID: "asg_a00000000000000000000000", DraftID: key}
	if err := validateAITrackingInput(AITrackingAcceptanceInput{Drafts: []AITrackingSuggestion{{Text: "Edited hypothesis?", Category: "neutral", GenerationReference: reference}}}); err != nil {
		t.Fatal(err)
	}
	if err := validateAITrackingInput(AITrackingAcceptanceInput{Drafts: []AITrackingSuggestion{{Text: "Example?", Category: "neutral", Provenance: "model_generated_hypothesis"}}}); err == nil {
		t.Error("untrusted generation claim accepted")
	}
}
