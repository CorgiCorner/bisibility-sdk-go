package bisibility

import (
	"encoding/json"
	"testing"
)

// TestAITrackingTrendsPreservesStratification covers contract-audit finding
// #8: decoding then re-encoding a trends response must preserve baseline,
// strata, and each stratum's category.
func TestAITrackingTrendsPreservesStratification(t *testing.T) {
	body := []byte(`{
      "current_run_id": "air_a00000000000000000000000",
      "previous_run_id": "air_b00000000000000000000000",
      "baseline": "neutral",
      "category": "neutral",
      "comparable": true,
      "reason": null,
      "delta": 0.125,
      "current": {"expected": 10, "observed": 9, "eligible": 8, "mentioned": 4, "absent_aio": 0, "partial": 1, "failed": 0, "unknown": 0, "missing": 1, "coverage": 0.9, "mention_rate": 0.5},
      "previous": {"expected": 10, "observed": 10, "eligible": 10, "mentioned": 3, "absent_aio": 0, "partial": 0, "failed": 0, "unknown": 0, "missing": 0, "coverage": 1.0, "mention_rate": 0.3},
      "strata": [
        {"category": "neutral", "current": {"expected": 5, "observed": 5, "eligible": 5, "mentioned": 2, "absent_aio": 0, "partial": 0, "failed": 0, "unknown": 0, "missing": 0, "coverage": 1.0, "mention_rate": 0.4}, "previous": {"expected": 5, "observed": 5, "eligible": 5, "mentioned": 1, "absent_aio": 0, "partial": 0, "failed": 0, "unknown": 0, "missing": 0, "coverage": 1.0, "mention_rate": 0.2}, "comparable": true, "reason": null, "delta": 0.2},
        {"category": "comparative", "current": {"expected": 5, "observed": 4, "eligible": 3, "mentioned": 1, "absent_aio": 0, "partial": 1, "failed": 0, "unknown": 0, "missing": 1, "coverage": 0.8, "mention_rate": 0.3333333333333333}, "previous": {"expected": 5, "observed": 5, "eligible": 5, "mentioned": 2, "absent_aio": 0, "partial": 0, "failed": 0, "unknown": 0, "missing": 0, "coverage": 1.0, "mention_rate": 0.4}, "comparable": false, "reason": "At least 90% complete coverage is required in both periods", "delta": null}
      ],
      "next_cursor": null,
      "previous_next_cursor": null
    }`)
	var decoded AITrackingTrends
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Baseline != "neutral" {
		t.Fatalf("baseline = %q, want neutral", decoded.Baseline)
	}
	if decoded.Category != "neutral" {
		t.Fatalf("category = %q, want neutral", decoded.Category)
	}
	if len(decoded.Strata) != 2 {
		t.Fatalf("strata len = %d, want 2", len(decoded.Strata))
	}
	if decoded.Strata[1].Category != "comparative" {
		t.Fatalf("strata[1].category = %q, want comparative", decoded.Strata[1].Category)
	}
	if decoded.Strata[1].Comparable {
		t.Fatal("strata[1].comparable = true, want false")
	}

	reencoded, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip map[string]json.RawMessage
	if err := json.Unmarshal(reencoded, &roundTrip); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"baseline", "category", "strata"} {
		if _, found := roundTrip[key]; !found {
			t.Fatalf("re-encoded trends missing %q: %s", key, reencoded)
		}
	}
	var strata []map[string]json.RawMessage
	if err := json.Unmarshal(roundTrip["strata"], &strata); err != nil {
		t.Fatal(err)
	}
	if len(strata) != 2 {
		t.Fatalf("re-encoded strata len = %d, want 2", len(strata))
	}
	if string(strata[0]["category"]) != `"neutral"` {
		t.Fatalf("re-encoded strata[0].category = %s, want \"neutral\"", strata[0]["category"])
	}
}
