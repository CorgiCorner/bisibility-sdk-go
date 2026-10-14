package bisibility

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func defaultsReviewFixture(t *testing.T) map[string]json.RawMessage {
	t.Helper()
	data, err := os.ReadFile("testdata/defaults-review.json")
	if err != nil {
		t.Fatal(err)
	}
	var f map[string]json.RawMessage
	if json.Unmarshal(data, &f) != nil {
		t.Fatal("invalidfixture")
	}
	return f
}
func TestReviewedDefaultsPreviewAndConfirmedCommand(t *testing.T) {
	f := defaultsReviewFixture(t)
	var patch ProjectDefaultsPatch
	json.Unmarshal(f["patch"], &patch)
	var project string
	json.Unmarshal(f["project"], &project)
	var expected ProjectDefaultsPreview
	json.Unmarshal(f["preview"], &expected)
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/capabilities" {
			writeJSON(t, w, 200, map[string]any{"data": []Capability{}, "apiVersions": []string{"v1"}, "features": map[string]string{"reviewed_project_defaults": "v1"}})
			return
		}
		requests++
		if r.Method != "PATCH" {
			t.Errorf("unexpectedmethod %s", r.Method)
		}
		var input ProjectDefaultsPatch
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Error(err)
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("preview_only") == "true" {
			if input.Review != nil || r.Header.Get("Idempotency-Key") != "" {
				t.Error("previewcarriedrevieworcommandkey")
			}
			w.Write(f["preview"])
			return
		}
		if input.Review == nil || !input.Review.Confirmed || r.Header.Get("Idempotency-Key") != "synthetic-reviewed-key-123" {
			t.Error("review/commandkeylost")
		}
		w.Write(f["saved"])
	}))
	defer server.Close()
	client := newTestClient(t, server.URL+"/api/v1")
	preview, err := client.PreviewProjectDefaults(context.Background(), project, patch, WithRequestHeader("Idempotency-Key", "request-default-key"))
	if err != nil {
		t.Fatal(err)
	}
	if preview.ExpectedRevision != "9007199254740993123" {
		t.Fatal("revisionlostprecision")
	}
	patch.Review = &ProjectDefaultsReview{ExpectedSourceHash: preview.ExpectedSourceHash, ExpectedRevision: preview.ExpectedRevision, PreviewToken: preview.PreviewToken, IdempotencyKey: "synthetic-reviewed-key-123", Confirmed: true}
	if _, err := client.UpdateProjectDefaults(context.Background(), project, patch); err != nil {
		t.Fatal(err)
	}
	if requests != 2 {
		t.Fatalf("requests=%d", requests)
	}
}
func TestLegacyDefaultsServerReceivesNoPreviewPATCH(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != "GET" {
			t.Error("unsupportedserverreceivedPATCH")
		}
		writeJSON(t, w, 200, map[string]any{"data": []Capability{}})
	}))
	defer server.Close()
	_, err := newTestClient(t, server.URL+"/api/v1").PreviewProjectDefaults(context.Background(), "prj_aaaaaaaaaaaaaaaaaaaaaaaa", ProjectDefaultsPatch{Frequency: "manual"})
	if err == nil || calls != 1 {
		t.Fatal("previewrequiresadvertisedsupport")
	}
}
func TestDefaultsReviewRejectsConflictingOrMissingConsentBeforeHTTP(t *testing.T) {
	f := defaultsReviewFixture(t)
	var patch ProjectDefaultsPatch
	json.Unmarshal(f["patch"], &patch)
	var preview ProjectDefaultsPreview
	json.Unmarshal(f["preview"], &preview)
	for _, confirmed := range []bool{false, true} {
		patch.Review = &ProjectDefaultsReview{ExpectedSourceHash: preview.ExpectedSourceHash, ExpectedRevision: preview.ExpectedRevision, PreviewToken: preview.PreviewToken, IdempotencyKey: "synthetic-reviewed-key-123", Confirmed: confirmed}
		client := newTestClient(t, "https://example.invalid/api/v1")
		_, err := client.UpdateProjectDefaults(context.Background(), "prj_aaaaaaaaaaaaaaaaaaaaaaaa", patch, WithIdempotencyKey("different-key-456"))
		if err == nil {
			t.Fatal("conflictingreviewaccepted")
		}
	}
}

func TestDefaultsPreviewDoesNotInventMissingImpact(t *testing.T) {
	f := defaultsReviewFixture(t)
	var root map[string]json.RawMessage
	json.Unmarshal(f["preview"], &root)
	delete(root, "impact")
	raw, _ := json.Marshal(root)
	var preview ProjectDefaultsPreview
	if json.Unmarshal(raw, &preview) == nil {
		t.Fatal("missing impact became zero")
	}
}

func TestDefaultsPreviewNestedFieldsCannotBorrowFromOtherGroups(t *testing.T) {
	f := defaultsReviewFixture(t)
	var root map[string]json.RawMessage
	json.Unmarshal(f["preview"], &root)
	var draft map[string]json.RawMessage
	json.Unmarshal(root["draft"], &draft)
	delete(draft, "device")
	root["draft"], _ = json.Marshal(draft)
	raw, _ := json.Marshal(root)
	for i := 0; i < 100; i++ {
		var preview ProjectDefaultsPreview
		if json.Unmarshal(raw, &preview) == nil {
			t.Fatal("missing draft.device borrowed resolved.device")
		}
	}
}
func TestDefaultsReviewRejectsWhitespaceIdentity(t *testing.T) {
	review := &ProjectDefaultsReview{ExpectedSourceHash: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ExpectedRevision: "1", PreviewToken: "token", IdempotencyKey: " abcdefgh ", Confirmed: true}
	if _, err := reviewedDefaultsRequestConfig(review, newRequestConfig()); err == nil {
		t.Fatal("header/body keys differed")
	}
}
