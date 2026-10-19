package bisibility

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnalyticsConsentMethods(t *testing.T) {
	for _, analytics := range []bool{false, true} {
		t.Run(map[bool]string{false: "refuse", true: "accept"}[analytics], func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/api/v1/capabilities" {
					_, _ = w.Write([]byte(`{"apiVersions":["v1"],"data":[]}`))
					return
				}
				calls++
				if r.URL.Path != "/api/v1/me/analytics-consent" {
					t.Errorf("path=%s", r.URL.Path)
				}
				if r.Header.Get("Authorization") != "Bearer bsb_pat_live_fixture" {
					t.Error("missing PAT")
				}
				if r.Method == http.MethodGet {
					_, _ = w.Write([]byte(`{"analytics":false,"decided_at":null,"source":null,"version":2}`))
					return
				}
				if r.Method != http.MethodPut {
					t.Errorf("method=%s", r.Method)
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if len(body) != 1 || body["analytics"] != analytics {
					t.Errorf("body=%v", body)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"analytics": analytics, "decided_at": "2026-10-18T12:00:00Z", "source": "api", "version": 2})
			}))
			defer server.Close()
			client, err := NewClient(WithAPIKey("bsb_pat_live_fixture"), WithBaseURL(server.URL+"/api/v1"))
			if err != nil {
				t.Fatal(err)
			}
			pending, err := client.GetAnalyticsConsent(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if pending.Analytics || pending.DecidedAt != nil || pending.Source != nil || pending.Version != 2 {
				t.Fatalf("pending=%+v", pending)
			}
			saved, err := client.UpdateAnalyticsConsent(context.Background(), UpdateAnalyticsConsentInput{Analytics: analytics})
			if err != nil {
				t.Fatal(err)
			}
			if saved.Analytics != analytics || saved.DecidedAt == nil || saved.Source == nil || *saved.Source != AnalyticsConsentSourceAPI {
				t.Fatalf("saved=%+v", saved)
			}
			if calls != 2 {
				t.Fatalf("calls=%d", calls)
			}
		})
	}
}
func TestAnalyticsConsentMalformedResponse(t *testing.T) {
	for _, body := range []string{`{"analytics":"false","decided_at":null,"source":null,"version":2}`, `{"analytics":false,"decided_at":null,"source":"other","version":2}`, `{"analytics":false,"source":null,"version":2}`} {
		t.Run(body, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/api/v1/capabilities" {
					_, _ = w.Write([]byte(`{"apiVersions":["v1"],"data":[]}`))
					return
				}
				_, _ = w.Write([]byte(body))
			}))
			defer server.Close()
			client, _ := NewClient(WithAPIKey("bsb_pat_live_fixture"), WithBaseURL(server.URL+"/api/v1"))
			_, err := client.GetAnalyticsConsent(context.Background())
			var responseErr *ResponseError
			if !errors.As(err, &responseErr) {
				t.Fatalf("error=%v", err)
			}
		})
	}
}
