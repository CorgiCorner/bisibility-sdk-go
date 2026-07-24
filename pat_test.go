package bisibility

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testPATKey = "bsp_live_1234567890abcdef"

type patMethodTestCase struct {
	name            string
	call            func(context.Context, *Client) (any, error)
	method          string
	path            string
	query           string
	body            string
	response        any
	status          int
	idempotencyKey  string
	wantContentType bool
	want            func(t *testing.T, got any)
}

func TestPersonalAccessTokenMethods(t *testing.T) {
	t.Parallel()

	enabled := true
	expiresInDays := 30
	description := "Deploy notifications"

	tests := []patMethodTestCase{
		{
			name:     "get me",
			call:     func(ctx context.Context, c *Client) (any, error) { return c.GetMe(ctx) },
			method:   http.MethodGet,
			path:     "/api/v1/me",
			response: meJSON("Owner Example"),
			want: func(t *testing.T, got any) {
				t.Helper()
				me := got.(*Me)
				assertEqual(t, me.ID, "usr_1")
				assertEqual(t, me.Email, "owner@example.com")
				assertEqual(t, me.Projects[0].ID, "prj_1")
				assertEqual(t, me.Projects[0].Domain, "example.com")
				assertEqual(t, me.Projects[0].Role, TeamRoleOwner)
			},
		},
		{
			name: "update me",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.UpdateMe(ctx, UpdateMeInput{Name: "Renamed Owner"}, WithIdempotencyKey("idem_me"))
			},
			method:          http.MethodPatch,
			path:            "/api/v1/me",
			body:            `{"name":"Renamed Owner"}`,
			response:        meJSON("Renamed Owner"),
			idempotencyKey:  "idem_me",
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*Me).Name, "Renamed Owner")
			},
		},
		{
			name:     "list my tokens",
			call:     func(ctx context.Context, c *Client) (any, error) { return c.ListMyTokens(ctx) },
			method:   http.MethodGet,
			path:     "/api/v1/me/tokens",
			response: map[string]any{"data": []any{personalAccessTokenJSON("pat_1")}, "meta": map[string]any{"next_cursor": nil}},
			want: func(t *testing.T, got any) {
				t.Helper()
				tokens := got.(*ListResponse[PersonalAccessToken])
				assertEqual(t, tokens.Data[0].ID, "pat_1")
				assertEqual(t, tokens.Data[0].Prefix, "bsp_live_12345678")
				assertEqual(t, tokens.Data[0].Scope, TokenScopeWrite)
				if tokens.Data[0].RevokedAt != nil {
					t.Fatalf("revoked_at = %v, want nil", tokens.Data[0].RevokedAt)
				}
				if tokens.Meta.NextCursor != nil {
					t.Fatalf("NextCursor = %v, want nil", *tokens.Meta.NextCursor)
				}
			},
		},
		{
			name: "create my token",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.CreateMyToken(ctx, CreateMyTokenInput{
					ExpiresInDays: &expiresInDays,
					Name:          "CI",
					Scope:         TokenScopeWrite,
				}, WithIdempotencyKey("idem_token"))
			},
			method:          http.MethodPost,
			path:            "/api/v1/me/tokens",
			body:            `{"expires_in_days":30,"name":"CI","scope":"write"}`,
			response:        createdPersonalAccessTokenJSON("pat_new"),
			status:          http.StatusCreated,
			idempotencyKey:  "idem_token",
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				created := got.(*CreatedPersonalAccessToken)
				assertEqual(t, created.ID, "pat_new")
				assertEqual(t, created.MaskedValue, "bsp_live_12345678******cdef")
				assertEqual(t, created.Token, testPATKey)
			},
		},
		{
			name: "create my token with defaults",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.CreateMyToken(ctx, CreateMyTokenInput{Name: "CI"})
			},
			method:          http.MethodPost,
			path:            "/api/v1/me/tokens",
			body:            `{"name":"CI"}`,
			response:        createdPersonalAccessTokenJSON("pat_new"),
			status:          http.StatusCreated,
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*CreatedPersonalAccessToken).Token, testPATKey)
			},
		},
		{
			name: "revoke my token escapes id",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.RevokeMyToken(ctx, "pat spaced")
			},
			method:   http.MethodDelete,
			path:     "/api/v1/me/tokens/pat%20spaced",
			response: revokedPersonalAccessTokenJSON("pat spaced"),
			want: func(t *testing.T, got any) {
				t.Helper()
				token := got.(*PersonalAccessToken)
				assertEqual(t, token.ID, "pat spaced")
				if token.RevokedAt == nil {
					t.Fatal("revoked_at = nil, want value")
				}
			},
		},
		{
			name: "revoke my current token",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.RevokeMyToken(ctx, "current")
			},
			method:   http.MethodDelete,
			path:     "/api/v1/me/tokens/current",
			response: revokedPersonalAccessTokenJSON("pat_1"),
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*PersonalAccessToken).ID, "pat_1")
			},
		},
		{
			name: "create project",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.CreateProject(ctx, CreateProjectInput{
					Defaults: &ProjectDefaultsPatch{
						Frequency:   RankCheckFrequencyDaily,
						LocationKey: "US/Texas/Austin",
					},
					Domain:        "example.com",
					Name:          "Example",
					TrackingScope: TrackingScopeCity,
				}, WithIdempotencyKey("idem_create_project"))
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects",
			body:            `{"defaults":{"frequency":"daily","location_key":"US/Texas/Austin"},"domain":"example.com","name":"Example","tracking_scope":"city"}`,
			response:        projectJSON("prj_new"),
			status:          http.StatusCreated,
			idempotencyKey:  "idem_create_project",
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*Project).ID, "prj_new")
				assertEqual(t, got.(*Project).WriteMode, ProjectWriteModeActive)
			},
		},
		{
			name: "create project with defaults omitted",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.CreateProject(ctx, CreateProjectInput{Domain: "example.com", Name: "Example"})
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects",
			body:            `{"domain":"example.com","name":"Example"}`,
			response:        projectJSON("prj_new"),
			status:          http.StatusCreated,
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*Project).Name, "Example")
			},
		},
		{
			name: "list project api keys",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ListProjectAPIKeys(ctx, "prj spaced", &PaginationOptions{Cursor: "cursor 1", Limit: 10})
			},
			method:   http.MethodGet,
			path:     "/api/v1/projects/prj%20spaced/api-keys",
			query:    "cursor=cursor+1&limit=10",
			response: listEnvelope([]any{apiKeyJSON("key_1")}, "cursor_2"),
			want: func(t *testing.T, got any) {
				t.Helper()
				keys := got.(*ListResponse[APIKey])
				assertEqual(t, keys.Data[0].ID, "key_1")
				assertEqual(t, keys.Data[0].Prefix, "bsk_live_12345678")
				assertEqual(t, *keys.Meta.NextCursor, "cursor_2")
			},
		},
		{
			name: "create project api key",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.CreateProjectAPIKey(ctx, "prj_1", CreateAPIKeyInput{Name: "CI"}, WithIdempotencyKey("idem_key"))
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects/prj_1/api-keys",
			body:            `{"name":"CI"}`,
			response:        createdProjectAPIKeyJSON("key_new"),
			status:          http.StatusCreated,
			idempotencyKey:  "idem_key",
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				created := got.(*CreatedAPIKey)
				assertEqual(t, created.ID, "key_new")
				assertEqual(t, created.MaskedValue, "bsk_live_12345678******cdef")
				assertEqual(t, created.Token, testAPIKey)
			},
		},
		{
			name: "list webhooks",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ListWebhooks(ctx, "prj spaced", &PaginationOptions{Limit: 5})
			},
			method:   http.MethodGet,
			path:     "/api/v1/projects/prj%20spaced/webhooks",
			query:    "limit=5",
			response: listEnvelope([]any{webhookJSON("wh_1")}, "cursor_2"),
			want: func(t *testing.T, got any) {
				t.Helper()
				webhooks := got.(*ListResponse[Webhook])
				assertEqual(t, webhooks.Data[0].ID, "wh_1")
				assertEqual(t, webhooks.Data[0].URL, "https://example.com/hooks/bisibility")
				assertEqual(t, webhooks.Data[0].Enabled, true)
				assertEqual(t, *webhooks.Meta.NextCursor, "cursor_2")
			},
		},
		{
			name: "create webhook",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.CreateWebhook(ctx, "prj_1", CreateWebhookInput{
					Description: description,
					Enabled:     &enabled,
					HMACSecret:  "super-secret-hmac-key",
					URL:         "https://example.com/hooks/bisibility",
				}, WithIdempotencyKey("idem_webhook"))
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects/prj_1/webhooks",
			body:            `{"description":"Deploy notifications","enabled":true,"hmac_secret":"super-secret-hmac-key","url":"https://example.com/hooks/bisibility"}`,
			response:        webhookJSON("wh_new"),
			status:          http.StatusCreated,
			idempotencyKey:  "idem_webhook",
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*Webhook).ID, "wh_new")
			},
		},
		{
			name: "update webhook",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.UpdateWebhook(ctx, "prj spaced", "wh spaced", UpdateWebhookInput{
					Description: &description,
					Enabled:     &enabled,
					HMACSecret:  "rotated-secret-hmac-key",
					URL:         "https://example.com/hooks/renamed",
				})
			},
			method:          http.MethodPatch,
			path:            "/api/v1/projects/prj%20spaced/webhooks/wh%20spaced",
			body:            `{"description":"Deploy notifications","enabled":true,"hmac_secret":"rotated-secret-hmac-key","url":"https://example.com/hooks/renamed"}`,
			response:        webhookJSON("wh spaced"),
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*Webhook).ID, "wh spaced")
			},
		},
		{
			name: "delete webhook",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.DeleteWebhook(ctx, "prj_1", "wh_1")
			},
			method:   http.MethodDelete,
			path:     "/api/v1/projects/prj_1/webhooks/wh_1",
			response: webhookJSON("wh_1"),
			want: func(t *testing.T, got any) {
				t.Helper()
				webhook := got.(*Webhook)
				assertEqual(t, webhook.ID, "wh_1")
				assertEqual(t, *webhook.Description, "Deploy notifications")
				if webhook.LastDeliveryAt == nil {
					t.Fatal("last_delivery_at = nil, want value")
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			runPATMethodTestCase(t, tt)
		})
	}
}

func runPATMethodTestCase(t *testing.T, tt patMethodTestCase) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertPATRequest(t, captureRequest(t, r), tt)
		writeJSON(t, w, statusOrOK(tt.status), tt.response)
	}))
	defer server.Close()

	client := newPATTestClient(t, server.URL+"/api/v1")
	got, err := tt.call(context.Background(), client)
	if err != nil {
		t.Fatalf("call returned error: %v", err)
	}
	tt.want(t, got)
}

func assertPATRequest(t *testing.T, captured capturedRequest, tt patMethodTestCase) {
	t.Helper()
	assertEqual(t, captured.Method, tt.method)
	assertEqual(t, captured.Path, tt.path)
	assertEqual(t, captured.RawQuery, tt.query)
	assertEqual(t, captured.Header.Get("Authorization"), "Bearer "+testPATKey)
	assertEqual(t, captured.Header.Get("X-Client"), "sdk-test")
	assertEqual(t, captured.Header.Get(projectHeader), "")
	assertEqual(t, captured.Header.Get("Idempotency-Key"), tt.idempotencyKey)
	if tt.body == "" {
		assertEqual(t, captured.Body, "")
		return
	}
	assertJSONEqual(t, captured.Body, tt.body)
	if tt.wantContentType {
		assertEqual(t, captured.Header.Get(contentTypeHeader), "application/json")
	}
}

func TestWithProjectIDSendsProjectHeaderOnEveryRequest(t *testing.T) {
	t.Parallel()

	calls := []struct {
		name string
		call func(context.Context, *Client) error
	}{
		{
			name: "get me",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.GetMe(ctx)
				return err
			},
		},
		{
			name: "list my tokens",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.ListMyTokens(ctx)
				return err
			},
		},
		{
			name: "list projects",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.ListProjects(ctx)
				return err
			},
		},
	}

	for _, tt := range calls {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var captured capturedRequest
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				captured = captureRequest(t, r)
				writeJSON(t, w, http.StatusOK, map[string]any{"data": []any{}, "meta": map[string]any{"next_cursor": nil}})
			}))
			defer server.Close()

			client := newPATTestClient(t, server.URL+"/api/v1", WithProjectID("prj_1"))
			if err := tt.call(context.Background(), client); err != nil {
				t.Fatalf("call returned error: %v", err)
			}
			if got := captured.Header.Get(projectHeader); got != "prj_1" {
				t.Fatalf("X-Bisibility-Project = %q, want prj_1", got)
			}
		})
	}
}

func TestWithRequestHeaderOverridesProjectHeader(t *testing.T) {
	t.Parallel()

	var captured capturedRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = captureRequest(t, r)
		writeJSON(t, w, http.StatusOK, map[string]any{"data": []any{}, "meta": map[string]any{"next_cursor": nil}})
	}))
	defer server.Close()

	client := newPATTestClient(t, server.URL+"/api/v1", WithProjectID("prj_1"))
	_, err := client.ListMyTokens(context.Background(), WithRequestHeader(projectHeader, "prj_2"))
	if err != nil {
		t.Fatalf("ListMyTokens returned error: %v", err)
	}
	if got := captured.Header[projectHeader]; len(got) != 1 || got[0] != "prj_2" {
		t.Fatalf("X-Bisibility-Project = %v, want [prj_2]", got)
	}
}

func TestWithProjectIDRejectsEmptyProjectID(t *testing.T) {
	t.Parallel()

	_, err := NewClient(WithAPIKey(testPATKey), WithProjectID("   "))
	var configErr *ConfigurationError
	if !errors.As(err, &configErr) {
		t.Fatalf("err = %T, want ConfigurationError", err)
	}
}

func newPATTestClient(t *testing.T, baseURL string, extra ...Option) *Client {
	t.Helper()

	options := append([]Option{
		WithAPIKey(testPATKey),
		WithBaseURL(baseURL),
		WithDefaultHeader("X-Client", "sdk-test"),
	}, extra...)
	client, err := NewClient(options...)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	return client
}

// meJSON mirrors the app's meResource shape (lib/api/me.ts).
func meJSON(name string) map[string]any {
	return map[string]any{
		"email": "owner@example.com",
		"id":    "usr_1",
		"name":  name,
		"projects": []map[string]any{
			{"domain": "example.com", "id": "prj_1", "name": "Example", "role": "owner"},
		},
	}
}

// personalAccessTokenJSON mirrors the app's token resource shape (lib/api/me-tokens.ts).
func personalAccessTokenJSON(id string) map[string]any {
	return map[string]any{
		"created_at":   "2026-01-01T00:00:00Z",
		"expires_at":   "2026-03-01T00:00:00Z",
		"id":           id,
		"last_used_at": nil,
		"name":         "CI",
		"prefix":       "bsp_live_12345678",
		"revoked_at":   nil,
		"scope":        "write",
	}
}

func createdPersonalAccessTokenJSON(id string) map[string]any {
	token := personalAccessTokenJSON(id)
	token["masked_value"] = "bsp_live_12345678******cdef"
	token["token"] = testPATKey
	return token
}

func revokedPersonalAccessTokenJSON(id string) map[string]any {
	token := personalAccessTokenJSON(id)
	token["revoked_at"] = "2026-01-07T00:00:00Z"
	return token
}

// apiKeyJSON mirrors the app's apiKeyResource shape (lib/api/resources.ts).
func apiKeyJSON(id string) map[string]any {
	return map[string]any{
		"created_at":   "2026-01-01T00:00:00Z",
		"id":           id,
		"last_used_at": nil,
		"name":         "Production",
		"prefix":       "bsk_live_12345678",
		"revoked_at":   nil,
	}
}

func createdProjectAPIKeyJSON(id string) map[string]any {
	key := apiKeyJSON(id)
	key["masked_value"] = "bsk_live_12345678******cdef"
	key["token"] = testAPIKey
	return key
}

// webhookJSON mirrors the app's webhookResource shape (lib/api/webhooks.ts).
// The hmac_secret is write-only and never appears in responses.
func webhookJSON(id string) map[string]any {
	return map[string]any{
		"created_at":       "2026-01-01T00:00:00Z",
		"description":      "Deploy notifications",
		"enabled":          true,
		"id":               id,
		"last_delivery_at": "2026-01-06T00:00:00Z",
		"updated_at":       "2026-01-02T00:00:00Z",
		"url":              "https://example.com/hooks/bisibility",
	}
}
