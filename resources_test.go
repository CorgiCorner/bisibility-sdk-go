package bisibility

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type resourceMethodTestCase struct {
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

func TestNewEndpointMethods(t *testing.T) {
	t.Parallel()

	enabled := true
	disabled := false
	primary := true
	notPrimary := false
	priority := 0
	threshold := 10
	changePct := 12.5
	cost := 0.06

	tests := []resourceMethodTestCase{
		{
			name: "list alert rules",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ListAlertRules(ctx, "prj_a00000000000000000000000", &PaginationOptions{Cursor: "cursor 1", Limit: 2})
			},
			method:   http.MethodGet,
			path:     "/api/v1/projects/prj_a00000000000000000000000/alert-rules",
			query:    "cursor=cursor+1&limit=2",
			response: listEnvelope([]any{alertRuleJSON("alr_a00000000000000000000000")}, "cursor_2"),
			want: func(t *testing.T, got any) {
				t.Helper()
				rules := got.(*ListResponse[AlertRule])
				assertEqual(t, rules.Data[0].ID, "alr_a00000000000000000000000")
				assertEqual(t, rules.Data[0].ChangePct.Float64(), 12.5)
				assertEqual(t, *rules.Meta.NextCursor, "cursor_2")
			},
		},
		{
			name: "create alert rule",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.CreateAlertRule(ctx, "prj_a00000000000000000000000", CreateAlertRuleInput{
					Channels:          []AlertChannel{AlertChannelEmail, AlertChannelWebhook},
					ConditionType:     AlertConditionTypeThreshold,
					Enabled:           &disabled,
					Name:              "Dropped out",
					RecipientIDs:      []string{"usr_a00000000000000000000000"},
					TargetIDs:         []string{"kw_a00000000000000000000000"},
					TargetType:        AlertTargetTypeKeyword,
					ThresholdPosition: &threshold,
				}, WithIdempotencyKey("idem_alert"))
			},
			method: http.MethodPost,
			path:   "/api/v1/projects/prj_a00000000000000000000000/alert-rules",
			body: `{
				"channels":["email","webhook"],
				"condition_type":"threshold",
				"enabled":false,
				"name":"Dropped out",
				"recipient_ids":["usr_a00000000000000000000000"],
				"target_ids":["kw_a00000000000000000000000"],
				"target_type":"keyword",
				"threshold_position":10
			}`,
			response:        alertRuleJSON("alr_a00000000000000000000000"),
			status:          http.StatusCreated,
			idempotencyKey:  "idem_alert",
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*AlertRule).ID, "alr_a00000000000000000000000")
			},
		},
		{
			name: "update alert rule",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.UpdateAlertRule(ctx, "alr_a00000000000000000000000", UpdateAlertRuleInput{
					ChangePct:     &changePct,
					ConditionType: AlertConditionTypeChangePct,
					Enabled:       &enabled,
					Name:          "Changed",
					TargetType:    AlertTargetTypeAll,
				})
			},
			method: http.MethodPatch,
			path:   "/api/v1/alert-rules/alr_a00000000000000000000000",
			body: `{
				"change_pct":12.5,
				"condition_type":"change_pct",
				"enabled":true,
				"name":"Changed",
				"target_type":"all"
			}`,
			response:        alertRuleJSON("alr_a00000000000000000000000"),
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*AlertRule).Name, "Rank changed")
			},
		},
		{
			name: "delete alert rule",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.DeleteAlertRule(ctx, "alr_a00000000000000000000000")
			},
			method:   http.MethodDelete,
			path:     "/api/v1/alert-rules/alr_a00000000000000000000000",
			response: AlertRuleDeleteResult{Deleted: true},
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*AlertRuleDeleteResult).Deleted, true)
			},
		},
		{
			name: "list triggered alerts",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ListTriggeredAlerts(ctx, "prj_a00000000000000000000000", &PaginationOptions{Limit: 5})
			},
			method:   http.MethodGet,
			path:     "/api/v1/projects/prj_a00000000000000000000000/triggered-alerts",
			query:    "limit=5",
			response: listResponse(triggeredAlertFixture("al_a00000000000000000000000")),
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*ListResponse[TriggeredAlert]).Data[0].Severity, AlertSeverityWarning)
			},
		},
		{
			name: "mute triggered alert",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.MuteTriggeredAlert(ctx, "prj_a00000000000000000000000", "al_a00000000000000000000000", WithIdempotencyKey("idem_mute"))
			},
			method:         http.MethodPost,
			path:           "/api/v1/projects/prj_a00000000000000000000000/triggered-alerts/al_a00000000000000000000000/mute",
			response:       map[string]any{"muted": true, "snoozed_until": "2026-07-23T10:00:00Z"},
			idempotencyKey: "idem_mute",
			want: func(t *testing.T, got any) {
				t.Helper()
				result := got.(*TriggeredAlertMuteResult)
				assertEqual(t, result.Muted, true)
				assertEqual(t, result.SnoozedUntil.UTC().Format(time.RFC3339), "2026-07-23T10:00:00Z")
			},
		},
		{
			name: "mark project alerts read",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.MarkProjectAlertsRead(ctx, "prj_a00000000000000000000000", WithIdempotencyKey("idem_read"))
			},
			method:         http.MethodPost,
			path:           "/api/v1/projects/prj_a00000000000000000000000/triggered-alerts/mark-read",
			response:       TriggeredAlertsReadResult{Updated: 7},
			idempotencyKey: "idem_read",
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*TriggeredAlertsReadResult).Updated, 7)
			},
		},
		{
			name: "list team members",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ListTeamMembers(ctx, "prj_a00000000000000000000000", &PaginationOptions{Limit: 25})
			},
			method:   http.MethodGet,
			path:     "/api/v1/projects/prj_a00000000000000000000000/team/members",
			query:    "limit=25",
			response: listResponse(teamMemberFixture("mbr_a00000000000000000000000")),
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*ListResponse[TeamMember]).Data[0].RoleValue, TeamRoleOwner)
			},
		},
		{
			name: "list team invites",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ListTeamInvites(ctx, "prj_a00000000000000000000000", &PaginationOptions{Cursor: "cursor_1"})
			},
			method:   http.MethodGet,
			path:     "/api/v1/projects/prj_a00000000000000000000000/team/invites",
			query:    "cursor=cursor_1",
			response: listResponse(teamInviteFixture("inv_a00000000000000000000000")),
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*ListResponse[TeamInvite]).Data[0].RoleValue, TeamRoleMember)
			},
		},
		{
			name: "list sitemap monitors",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ListSitemapMonitors(ctx, "prj_a00000000000000000000000")
			},
			method: http.MethodGet,
			path:   "/api/v1/projects/prj_a00000000000000000000000/sitemap-monitors",
			response: listResponse(map[string]any{
				"enabled":     true,
				"id":          "prj_a00000000000000000000000",
				"project_id":  "prj_a00000000000000000000000",
				"sitemap_url": "https://example.com/sitemap.xml",
				"status":      "active",
				"latest_snapshot": map[string]any{
					"fetched_at":  "2026-07-22T09:00:00Z",
					"sitemap_url": "https://example.com/sitemap.xml",
					"url_count":   42,
				},
			}),
			want: func(t *testing.T, got any) {
				t.Helper()
				monitor := got.(*ListResponse[SitemapMonitor]).Data[0]
				assertEqual(t, monitor.Status, SitemapMonitorStatusActive)
				assertEqual(t, monitor.LatestSnapshot.URLCount, 42)
			},
		},
		{
			name: "update sitemap monitor includes false",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.UpdateSitemapMonitor(ctx, "prj_a00000000000000000000000", "prj_a00000000000000000000000", UpdateSitemapMonitorInput{Enabled: false})
			},
			method:          http.MethodPatch,
			path:            "/api/v1/projects/prj_a00000000000000000000000/sitemap-monitors/prj_a00000000000000000000000",
			body:            `{"enabled":false}`,
			response:        map[string]any{"enabled": false, "id": "prj_a00000000000000000000000", "latest_snapshot": nil, "project_id": "prj_a00000000000000000000000", "sitemap_url": nil, "status": "disabled"},
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				monitor := got.(*SitemapMonitor)
				assertEqual(t, monitor.Enabled, false)
				assertEqual(t, monitor.Status, SitemapMonitorStatusDisabled)
				if monitor.LatestSnapshot != nil || monitor.SitemapURL != nil {
					t.Fatal("nullable sitemap fields were not preserved")
				}
			},
		},
		{
			name: "create team invite",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.CreateTeamInvite(ctx, "prj_a00000000000000000000000", CreateTeamInviteInput{
					Email: "new@example.com",
					Role:  TeamRoleViewer,
				})
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects/prj_a00000000000000000000000/team/invites",
			body:            `{"email":"new@example.com","role":"viewer"}`,
			response:        createdTeamInviteFixture("inv_b00000000000000000000000"),
			status:          http.StatusCreated,
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*CreatedTeamInvite).InviteLink, "https://app.test/invite/raw")
			},
		},
		{
			name: "revoke team invite top-level",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.RevokeTeamInvite(ctx, "inv_a00000000000000000000000")
			},
			method:   http.MethodDelete,
			path:     "/api/v1/team/invites/inv_a00000000000000000000000",
			response: RevokeTeamInviteResult{ID: "inv_a00000000000000000000000"},
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*RevokeTeamInviteResult).ID, "inv_a00000000000000000000000")
			},
		},
		{
			name: "revoke team invite project route",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.RevokeProjectTeamInvite(ctx, "prj_a00000000000000000000000", "inv_a00000000000000000000000")
			},
			method:   http.MethodDelete,
			path:     "/api/v1/projects/prj_a00000000000000000000000/team/invites/inv_a00000000000000000000000",
			response: RevokeTeamInviteResult{ID: "inv_a00000000000000000000000"},
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*RevokeTeamInviteResult).ID, "inv_a00000000000000000000000")
			},
		},
		{
			name: "list providers",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ListProviders(ctx, "prj_a00000000000000000000000", &PaginationOptions{Limit: 10})
			},
			method:   http.MethodGet,
			path:     "/api/v1/projects/prj_a00000000000000000000000/providers",
			query:    "limit=10",
			response: listResponse(providerFixture(ProviderIDDataForSEO)),
			want: func(t *testing.T, got any) {
				t.Helper()
				providers := got.(*ListResponse[Provider])
				assertEqual(t, providers.Data[0].ID, ProviderIDDataForSEO)
				assertEqual(t, *providers.Data[0].Enabled, true)
			},
		},
		{
			name: "connect provider",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConnectProvider(ctx, "prj_a00000000000000000000000", ProviderIDDataForSEO, ConnectProviderInput{
					CostPerCheck: &cost,
					Enabled:      &enabled,
					Login:        "login",
					Primary:      &primary,
					Priority:     &priority,
					Secret:       "secret",
				})
			},
			method: http.MethodPost,
			path:   "/api/v1/projects/prj_a00000000000000000000000/providers/dataforseo/connect",
			body: `{
				"cost_per_check":0.06,
				"enabled":true,
				"login":"login",
				"primary":true,
				"priority":0,
				"secret":"secret"
			}`,
			response:        providerConnectionJSON(ProviderIDDataForSEO),
			status:          http.StatusCreated,
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				conn := got.(*ProviderConnection)
				assertEqual(t, conn.Provider, ProviderIDDataForSEO)
				assertEqual(t, conn.CostPerCheckCents.Float64(), 0.06)
			},
		},
		{
			name: "test provider connection",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.TestProviderConnection(ctx, "prj_a00000000000000000000000", ProviderIDSerpAPI, TestProviderConnectionInput{
					Credentials: &ProviderCredentialsInput{APIKey: "api_key"},
				})
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects/prj_a00000000000000000000000/providers/serpapi/test",
			body:            `{"credentials":{"api_key":"api_key"}}`,
			response:        ProviderTestResult{Balance: &cost, Message: "Ready", OK: true},
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*ProviderTestResult).OK, true)
			},
		},
		{
			name: "test provider connection with endpoint",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.TestProviderConnection(ctx, "prj_a00000000000000000000000", ProviderIDPlausible, TestProviderConnectionInput{
					Credentials: &ProviderCredentialsInput{APIKey: "api_key", Endpoint: "https://plausible.example"},
				})
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects/prj_a00000000000000000000000/providers/plausible/test",
			body:            `{"credentials":{"api_key":"api_key","endpoint":"https://plausible.example"}}`,
			response:        ProviderTestResult{Message: "Ready", OK: true},
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*ProviderTestResult).OK, true)
			},
		},
		{
			name: "update provider settings",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.UpdateProviderSettings(ctx, "prj_a00000000000000000000000", ProviderIDDataForSEO, ProviderSettingsInput{
					Enabled:  &enabled,
					Primary:  &notPrimary,
					Priority: &priority,
				})
			},
			method:          http.MethodPatch,
			path:            "/api/v1/projects/prj_a00000000000000000000000/providers/dataforseo",
			body:            `{"enabled":true,"primary":false,"priority":0}`,
			response:        providerConnectionJSON(ProviderIDDataForSEO),
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*ProviderConnection).Priority, 0)
			},
		},
		{
			name: "set provider enabled includes false",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.SetProviderEnabled(ctx, "prj_a00000000000000000000000", ProviderIDDataForSEO, false)
			},
			method:          http.MethodPatch,
			path:            "/api/v1/projects/prj_a00000000000000000000000/providers/dataforseo",
			body:            `{"enabled":false}`,
			response:        providerConnectionJSON(ProviderIDDataForSEO),
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*ProviderConnection).Enabled, true)
			},
		},
		{
			name: "set provider priority includes zero",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.SetProviderPriority(ctx, "prj_a00000000000000000000000", ProviderIDDataForSEO, 0)
			},
			method:          http.MethodPatch,
			path:            "/api/v1/projects/prj_a00000000000000000000000/providers/dataforseo",
			body:            `{"priority":0}`,
			response:        providerConnectionJSON(ProviderIDDataForSEO),
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*ProviderConnection).Priority, 0)
			},
		},
		{
			name: "set primary provider includes false",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.SetPrimaryProvider(ctx, "prj_a00000000000000000000000", ProviderIDDataForSEO, false)
			},
			method:          http.MethodPatch,
			path:            "/api/v1/projects/prj_a00000000000000000000000/providers/dataforseo",
			body:            `{"primary":false}`,
			response:        providerConnectionJSON(ProviderIDDataForSEO),
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*ProviderConnection).IsPrimary, true)
			},
		},
		{
			name: "disconnect provider",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.DisconnectProvider(ctx, "prj_a00000000000000000000000", ProviderIDDataForSEO)
			},
			method:   http.MethodDelete,
			path:     "/api/v1/projects/prj_a00000000000000000000000/providers/dataforseo",
			response: ProviderDisconnectResult{OK: true},
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*ProviderDisconnectResult).OK, true)
			},
		},
		{
			name: "list saved views",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ListSavedViews(ctx, "prj_a00000000000000000000000", &PaginationOptions{Limit: 3})
			},
			method:   http.MethodGet,
			path:     "/api/v1/projects/prj_a00000000000000000000000/saved-views",
			query:    "limit=3",
			response: listResponse(savedViewFixture("viw_a00000000000000000000000")),
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*ListResponse[SavedView]).Data[0].Config.Filters.Position[0], SavedViewPositionTop10)
			},
		},
		{
			name: "create saved view",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.CreateSavedView(ctx, "prj_a00000000000000000000000", CreateSavedViewInput{
					Config: savedViewConfigFixture(),
					Name:   "Winners",
				})
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects/prj_a00000000000000000000000/saved-views",
			body:            `{"config":{"filters":{"change":"up","country":"us","device":"desktop","position":["top10"],"tags":["Product"],"vol_max":50},"search":"rank"},"name":"Winners"}`,
			response:        savedViewFixture("viw_a00000000000000000000000"),
			status:          http.StatusCreated,
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*SavedView).Name, "Winners")
			},
		},
		{
			name: "delete saved view top-level",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.DeleteSavedView(ctx, "viw_a00000000000000000000000")
			},
			method:   http.MethodDelete,
			path:     "/api/v1/saved-views/viw_a00000000000000000000000",
			response: SavedViewDeleteResult{Deleted: true},
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*SavedViewDeleteResult).Deleted, true)
			},
		},
		{
			name: "delete saved view project route",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.DeleteProjectSavedView(ctx, "prj_a00000000000000000000000", "viw_a00000000000000000000000")
			},
			method:   http.MethodDelete,
			path:     "/api/v1/projects/prj_a00000000000000000000000/saved-views/viw_a00000000000000000000000",
			response: SavedViewDeleteResult{Deleted: true},
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*SavedViewDeleteResult).Deleted, true)
			},
		},
		{
			name: "list competitors",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ListCompetitors(ctx, "prj_a00000000000000000000000", &PaginationOptions{Limit: 5})
			},
			method:   http.MethodGet,
			path:     "/api/v1/projects/prj_a00000000000000000000000/competitors",
			query:    "limit=5",
			response: competitorsResponseFixture(),
			want: func(t *testing.T, got any) {
				t.Helper()
				resp := got.(*ListCompetitorsResponse)
				assertEqual(t, resp.Data[0].Domain, "rankzly.io")
				assertEqual(t, resp.Meta.Markets[0].Shares[0].ShareOfVoice, 65)
				assertEqual(t, resp.Meta.Suggestions[0].Overlap, 4)
			},
		},
		{
			name: "add competitor",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.AddCompetitor(ctx, "prj_a00000000000000000000000", AddCompetitorInput{Domain: "rankzly.io", Label: "Rankzly"})
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects/prj_a00000000000000000000000/competitors",
			body:            `{"domain":"rankzly.io","label":"Rankzly"}`,
			response:        competitorFixture("cmp_a00000000000000000000000"),
			status:          http.StatusCreated,
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*Competitor).Domain, "rankzly.io")
			},
		},
		{
			name: "remove competitor top-level",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.RemoveCompetitor(ctx, "cmp_a00000000000000000000000")
			},
			method:   http.MethodDelete,
			path:     "/api/v1/competitors/cmp_a00000000000000000000000",
			response: CompetitorRemoveResult{Removed: true},
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*CompetitorRemoveResult).Removed, true)
			},
		},
		{
			name: "remove competitor project route",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.RemoveProjectCompetitor(ctx, "prj_a00000000000000000000000", "cmp_a00000000000000000000000")
			},
			method:   http.MethodDelete,
			path:     "/api/v1/projects/prj_a00000000000000000000000/competitors/cmp_a00000000000000000000000",
			response: CompetitorRemoveResult{Removed: true},
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*CompetitorRemoveResult).Removed, true)
			},
		},
		{
			name: "get notification preferences",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.GetNotificationPreferences(ctx, "prj_a00000000000000000000000")
			},
			method:   http.MethodGet,
			path:     "/api/v1/projects/prj_a00000000000000000000000/notification-preferences",
			response: notificationPreferencesFixture(),
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*NotificationPreferences).EmailVerification, "verified")
			},
		},
		{
			name: "update notification preferences includes false",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.UpdateNotificationPreferences(ctx, "prj_a00000000000000000000000", UpdateNotificationPreferencesInput{
					AlertEmail: &disabled,
					AlertInApp: &enabled,
					CheckEmail: &disabled,
				})
			},
			method:          http.MethodPatch,
			path:            "/api/v1/projects/prj_a00000000000000000000000/notification-preferences",
			body:            `{"alert_email":false,"alert_in_app":true,"check_email":false}`,
			response:        updatedNotificationPreferencesFixture(),
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*UpdatedNotificationPreferences).AlertEmail, false)
			},
		},
		{
			name: "list migration tokens",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ListMigrationTokens(ctx, "prj_a00000000000000000000000")
			},
			method:   http.MethodGet,
			path:     "/api/v1/projects/prj_a00000000000000000000000/migration-tokens",
			response: migrationTokensResponseFixture(),
			want: func(t *testing.T, got any) {
				t.Helper()
				resp := got.(*ListMigrationTokensResponse)
				assertEqual(t, resp.Data[0].Scope, MigrationScopeFull)
				assertEqual(t, resp.Meta.ImportJob.State, CloudImportStateIdle)
				if !json.Valid(resp.Meta.ImportJob.Counts) {
					t.Fatalf("counts is invalid JSON: %q", string(resp.Meta.ImportJob.Counts))
				}
			},
		},
		{
			name: "mint migration token",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.MintMigrationToken(ctx, "prj_a00000000000000000000000", MintMigrationTokenInput{Scope: MigrationScopeKeywords})
			},
			method:          http.MethodPost,
			path:            "/api/v1/projects/prj_a00000000000000000000000/migration-tokens",
			body:            `{"scope":"keywords"}`,
			response:        issuedMigrationTokenFixture("ferry_a00000000000000000000000"),
			status:          http.StatusCreated,
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*IssuedMigrationToken).Token, "mig_raw")
			},
		},
		{
			name: "revoke migration token top-level",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.RevokeMigrationToken(ctx, "ferry_a00000000000000000000000")
			},
			method:   http.MethodDelete,
			path:     "/api/v1/migration-tokens/ferry_a00000000000000000000000",
			response: revokedMigrationTokenFixture("ferry_a00000000000000000000000"),
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*RevokedMigrationToken).ID, "ferry_a00000000000000000000000")
			},
		},
		{
			name: "revoke migration token project route",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.RevokeProjectMigrationToken(ctx, "prj_a00000000000000000000000", "ferry_a00000000000000000000000")
			},
			method:   http.MethodDelete,
			path:     "/api/v1/projects/prj_a00000000000000000000000/migration-tokens/ferry_a00000000000000000000000",
			response: revokedMigrationTokenFixture("ferry_a00000000000000000000000"),
			want: func(t *testing.T, got any) {
				t.Helper()
				assertEqual(t, got.(*RevokedMigrationToken).ID, "ferry_a00000000000000000000000")
			},
		},
		{
			name: "create signal",
			call: func(ctx context.Context, c *Client) (any, error) {
				happened := mustTime("2026-07-04T19:30:00Z")
				return c.CreateSignal(ctx, CreateSignalInput{
					HappenedAt: &happened,
					KeywordID:  "kw_a00000000000000000000000",
					Payload:    JSONValue{"version": "1.2.3"},
					Severity:   SignalSeverityWarning,
					Source:     SignalSourceDeploy,
					Type:       "deploy.completed",
					URL:        "https://example.com/releases/1",
				}, WithIdempotencyKey("idem_signal"))
			},
			method: http.MethodPost,
			path:   "/api/v1/signals",
			body: `{
				"happened_at":"2026-07-04T19:30:00Z",
				"keyword_id":"kw_a00000000000000000000000",
				"payload":{"version":"1.2.3"},
				"severity":"warning",
				"source":"deploy",
				"type":"deploy.completed",
				"url":"https://example.com/releases/1"
			}`,
			response:        signalJSON("sig_a00000000000000000000000"),
			status:          http.StatusCreated,
			idempotencyKey:  "idem_signal",
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				signal := got.(*Signal)
				assertEqual(t, signal.ID, "sig_a00000000000000000000000")
				assertEqual(t, signal.PublicID, "sig_a00000000000000000000000")
				assertEqual(t, signal.ProjectID, "prj_a00000000000000000000000")
				assertEqual(t, *signal.KeywordID, "kw_a00000000000000000000000")
				assertEqual(t, signal.Severity, SignalSeverityWarning)
				assertEqual(t, signal.Source, SignalSourceDeploy)
				assertEqual(t, signal.Type, "deploy.completed")
				assertEqual(t, signal.Payload["version"], any("1.2.3"))
				assertEqual(t, signal.HappenedAt, mustTime("2026-07-04T19:30:00Z"))
				assertEqual(t, signal.CreatedAt, mustTime("2026-07-04T19:31:00Z"))
			},
		},
		{
			name: "create signal omits optional fields",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.CreateSignal(ctx, CreateSignalInput{
					Source: SignalSourceAPI,
					Type:   "api.changed",
				})
			},
			method:          http.MethodPost,
			path:            "/api/v1/signals",
			body:            `{"source":"api","type":"api.changed"}`,
			response:        minimalSignalJSON("sig_b00000000000000000000000"),
			status:          http.StatusCreated,
			wantContentType: true,
			want: func(t *testing.T, got any) {
				t.Helper()
				signal := got.(*Signal)
				assertEqual(t, signal.ID, "sig_b00000000000000000000000")
				assertEqual(t, signal.Severity, SignalSeverityInfo)
				if signal.KeywordID != nil {
					t.Fatalf("KeywordID = %v, want nil", *signal.KeywordID)
				}
				if signal.URL != nil {
					t.Fatalf("URL = %v, want nil", *signal.URL)
				}
				if signal.Payload != nil {
					t.Fatalf("Payload = %v, want nil", signal.Payload)
				}
			},
		},
		{
			name: "list signals with filters",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ListSignals(ctx, "prj_a00000000000000000000000", &ListSignalsOptions{
					Cursor: "cursor 1",
					From:   mustTime("2026-07-01T00:00:00Z"),
					Limit:  1,
					Source: SignalSourceDeploy,
					To:     mustTime("2026-07-05T00:00:00Z"),
					Type:   "deploy.completed",
				})
			},
			method:   http.MethodGet,
			path:     "/api/v1/projects/prj_a00000000000000000000000/signals",
			query:    "cursor=cursor+1&from=2026-07-01T00%3A00%3A00Z&limit=1&source=deploy&to=2026-07-05T00%3A00%3A00Z&type=deploy.completed",
			response: listEnvelope([]any{signalJSON("sig_b00000000000000000000000")}, "cursor_2"),
			want: func(t *testing.T, got any) {
				t.Helper()
				signals := got.(*ListResponse[Signal])
				assertEqual(t, signals.Data[0].PublicID, "sig_b00000000000000000000000")
				assertEqual(t, signals.Data[0].Source, SignalSourceDeploy)
				assertEqual(t, *signals.Meta.NextCursor, "cursor_2")
			},
		},
		{
			name: "list signals without filters",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ListSignals(ctx, "prj_a00000000000000000000000", nil)
			},
			method: http.MethodGet,
			path:   "/api/v1/projects/prj_a00000000000000000000000/signals",
			response: map[string]any{
				"data": []any{signalJSON("sig_a00000000000000000000000")},
				"meta": map[string]any{"next_cursor": nil},
			},
			want: func(t *testing.T, got any) {
				t.Helper()
				signals := got.(*ListResponse[Signal])
				assertEqual(t, signals.Data[0].ID, "sig_a00000000000000000000000")
				if signals.Meta.NextCursor != nil {
					t.Fatalf("NextCursor = %v, want nil", *signals.Meta.NextCursor)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			runResourceMethodTestCase(t, tt)
		})
	}
}

func runResourceMethodTestCase(t *testing.T, tt resourceMethodTestCase) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertResourceRequest(t, captureRequest(t, r), tt)
		writeJSON(t, w, statusOrOK(tt.status), tt.response)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	got, err := tt.call(context.Background(), client)
	if err != nil {
		t.Fatalf("call returned error: %v", err)
	}
	tt.want(t, got)
}

func assertResourceRequest(t *testing.T, captured capturedRequest, tt resourceMethodTestCase) {
	t.Helper()
	assertEqual(t, captured.Method, tt.method)
	assertEqual(t, captured.Path, tt.path)
	assertEqual(t, captured.RawQuery, tt.query)
	assertEqual(t, captured.Header.Get("Authorization"), "Bearer "+testAPIKey)
	assertEqual(t, captured.Header.Get("X-Client"), "sdk-test")
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

func listEnvelope(items []any, nextCursor string) map[string]any {
	return map[string]any{
		"data": items,
		"meta": map[string]any{"next_cursor": nextCursor},
	}
}

func alertRuleJSON(id string) map[string]any {
	return map[string]any{
		"change_pct":         "12.5",
		"channel":            "Email",
		"channels":           []string{"email"},
		"condition":          "position changes by 12.5%",
		"condition_type":     "change_pct",
		"enabled":            true,
		"fires":              "2 this week",
		"id":                 id,
		"name":               "Rank changed",
		"period":             "Each check",
		"recipient_ids":      []string{"usr_a00000000000000000000000"},
		"scope":              "All keywords",
		"severity":           "warning",
		"status":             "active",
		"target_ids":         []string{"kw_a00000000000000000000000"},
		"target_type":        "keyword",
		"threshold_position": nil,
		"top_n":              nil,
	}
}

func triggeredAlertFixture(id string) TriggeredAlert {
	return TriggeredAlert{
		Action:   "Review the keyword",
		CTAs:     []string{"Open keyword"},
		Current:  "#7",
		Headline: "rankzly.io overtook you",
		ID:       id,
		Keyword:  "rank tracker",
		Previous: "#4",
		Rule:     "Competitor overtook us",
		Severity: AlertSeverityWarning,
		Unread:   true,
		When:     "5h ago",
	}
}

func teamMemberFixture(id string) TeamMember {
	return TeamMember{
		Color:     "accent",
		Email:     "owner@example.com",
		ID:        id,
		Initials:  "OE",
		Name:      "Owner Example",
		Role:      "Owner",
		RoleValue: TeamRoleOwner,
	}
}

func teamInviteFixture(id string) TeamInvite {
	return TeamInvite{
		Email:        "new@example.com",
		ExpiresLabel: "expires in 7d",
		ID:           id,
		InvitedLabel: "invited just now",
		Role:         "Editor",
		RoleValue:    TeamRoleMember,
	}
}

func createdTeamInviteFixture(id string) CreatedTeamInvite {
	return CreatedTeamInvite{
		ExpiresAt:  mustTime("2026-01-08T00:00:00Z"),
		ID:         id,
		InviteLink: "https://app.test/invite/raw",
	}
}

func providerFixture(id ProviderID) Provider {
	return Provider{
		CategoryID:    "serp",
		CategoryTitle: "SERP providers",
		Description:   "Rank-data provider.",
		Drawer: ProviderDrawer{
			Activities:       []ProviderMetaRow{{Label: "Last used", Value: "Never"}},
			CostHelp:         "Provider billing remains direct.",
			CredentialFields: []ProviderCredentialField{{Label: "Login", Name: "login", Placeholder: "DATAFORSEO_LOGIN"}},
			Defaults: ProviderDrawerDefaults{
				CostPerCheck: 0.06,
				Depth:        "Top 100",
				Device:       "Desktop",
				Enabled:      boolPtr(true),
				Language:     "English",
				Location:     "United States",
				Login:        "",
				Primary:      true,
				Priority:     intPtr(0),
				Secret:       "",
			},
			EnvHint:            "Use env vars.",
			PrimaryToggleLabel: "Set primary",
		},
		Enabled:  boolPtr(true),
		Icon:     "database",
		ID:       id,
		Meta:     []ProviderMetaRow{{Label: "State", Value: "Enabled"}},
		Name:     "DataForSEO",
		Primary:  boolPtr(true),
		Priority: intPtr(0),
		Status:   ProviderStatusConnected,
		Tint:     "var(--accent)",
	}
}

func providerConnectionJSON(id ProviderID) map[string]any {
	return map[string]any{
		"cost_per_check_cents": "0.06",
		"created_at":           "2026-01-01T00:00:00Z",
		"enabled":              true,
		"id":                   "conn_a00000000000000000000000",
		"is_primary":           true,
		"kind":                 "serp",
		"last_used_at":         nil,
		"priority":             0,
		"project_id":           "prj_a00000000000000000000000",
		"provider":             id,
		"status":               "connected",
		"updated_at":           "2026-01-02T00:00:00Z",
	}
}

func savedViewConfigFixture() SavedViewConfig {
	return SavedViewConfig{
		Filters: SavedViewFilters{
			Change:   "up",
			Country:  "us",
			Device:   "desktop",
			Position: []SavedViewPositionBucket{SavedViewPositionTop10},
			Tags:     []string{"Product"},
			VolMax:   50,
		},
		Search: "rank",
	}
}

func savedViewFixture(id string) SavedView {
	return SavedView{
		Config:      savedViewConfigFixture(),
		CreatedAt:   mustTime("2026-01-01T00:00:00Z"),
		CreatedByID: strPtr("usr_a00000000000000000000000"),
		ID:          id,
		Name:        "Winners",
	}
}

func competitorsResponseFixture() ListCompetitorsResponse {
	gap := -3
	rankOwn := 4
	rankCompetitor := 1
	return ListCompetitorsResponse{
		Data: []ManagedCompetitor{{
			Domain:   "rankzly.io",
			ID:       "cmp_a00000000000000000000000",
			Initials: "RI",
			Label:    "Rankzly",
		}},
		Meta: CompetitorsMeta{
			Markets: []CompetitorMarket{{
				CheckedKeywordCount: 1,
				Columns: []CompetitorColumn{
					{Domain: "example.com", Kind: "You", Label: "example.com"},
					{Domain: "rankzly.io", ID: "cmp_a00000000000000000000000", Kind: "Managed", Label: "Rankzly"},
				},
				CompetitorCount:     1,
				Country:             "United States",
				Device:              "Desktop",
				Engine:              "Google",
				HasRankData:         true,
				Key:                 "United States::Desktop::Google",
				Rows:                []CompetitorHeadToHeadRow{{Gap: &gap, Keyword: "rank tracker", Ranks: map[string]*int{"example.com": &rankOwn, "rankzly.io": &rankCompetitor}}},
				Shares:              []CompetitorShare{{Domain: "example.com", Initials: "EC", Kind: "You", Label: "example.com", ShareOfVoice: 65, SharedKeywords: 1}},
				SharedKeywordCount:  1,
				TrackedKeywordCount: 1,
			}},
			Suggestions: []SuggestedCompetitor{{Domain: "serp.example", Initials: "SE", Overlap: 4}},
		},
	}
}

func competitorFixture(id string) Competitor {
	return Competitor{Domain: "rankzly.io", ID: id, Label: strPtr("Rankzly")}
}

func notificationPreferencesFixture() NotificationPreferences {
	return NotificationPreferences{
		AlertEmail:        true,
		AlertInApp:        true,
		AlertSlack:        false,
		AlertWebhook:      false,
		CheckEmail:        false,
		CheckInApp:        true,
		Email:             "owner@example.com",
		EmailVerification: "verified",
		ImportEmail:       true,
		ImportInApp:       true,
		InviteEmail:       true,
		InviteInApp:       true,
		ProjectID:         "prj_a00000000000000000000000",
		SlackAvailable:    false,
		WebhookAvailable:  false,
	}
}

func updatedNotificationPreferencesFixture() UpdatedNotificationPreferences {
	return UpdatedNotificationPreferences{
		AlertEmail:   false,
		AlertInApp:   true,
		AlertSlack:   false,
		AlertWebhook: false,
		CheckEmail:   false,
		CheckInApp:   true,
		ImportEmail:  true,
		ImportInApp:  true,
		InviteEmail:  true,
		InviteInApp:  true,
		ProjectID:    "prj_a00000000000000000000000",
	}
}

func migrationTokensResponseFixture() ListMigrationTokensResponse {
	createdAt := mustTime("2026-01-01T00:00:00Z")
	return ListMigrationTokensResponse{
		Data: []MigrationToken{{
			CreatedAt: createdAt,
			CreatedBy: MigrationTokenCreatedBy{
				Email: "owner@example.com",
				Name:  "Owner Example",
			},
			ExpiresAt: mustTime("2026-01-01T01:00:00Z"),
			ID:        "ferry_a00000000000000000000000",
			Scope:     MigrationScopeFull,
			SingleUse: true,
		}},
		Meta: MigrationTokensMeta{
			ImportJob: CloudImportJob{
				Counts:    json.RawMessage(`{"keywords":2}`),
				CreatedAt: &createdAt,
				Progress:  0,
				State:     CloudImportStateIdle,
			},
		},
	}
}

func issuedMigrationTokenFixture(id string) IssuedMigrationToken {
	createdAt := mustTime("2026-01-01T00:00:00Z")
	return IssuedMigrationToken{
		CreatedAt: createdAt,
		ExpiresAt: mustTime("2026-01-01T01:00:00Z"),
		ID:        id,
		ImportJob: CloudImportJob{
			Counts:    json.RawMessage(`null`),
			CreatedAt: &createdAt,
			Progress:  0,
			State:     CloudImportStateIdle,
		},
		Scope:     MigrationScopeKeywords,
		SingleUse: true,
		Token:     "mig_raw",
	}
}

func revokedMigrationTokenFixture(id string) RevokedMigrationToken {
	return RevokedMigrationToken{ID: id, RevokedAt: mustTime("2026-01-01T00:30:00Z")}
}

func signalJSON(id string) map[string]any {
	return map[string]any{
		"created_at":  "2026-07-04T19:31:00Z",
		"happened_at": "2026-07-04T19:30:00Z",
		"id":          id,
		"keyword_id":  "kw_a00000000000000000000000",
		"payload":     map[string]any{"version": "1.2.3"},
		"project_id":  "prj_a00000000000000000000000",
		"public_id":   id,
		"severity":    "warning",
		"source":      "deploy",
		"type":        "deploy.completed",
		"url":         "https://example.com/releases/1",
	}
}

func minimalSignalJSON(id string) map[string]any {
	return map[string]any{
		"created_at":  "2026-07-04T19:31:00Z",
		"happened_at": "2026-07-04T19:31:00Z",
		"id":          id,
		"keyword_id":  nil,
		"payload":     nil,
		"project_id":  "prj_a00000000000000000000000",
		"public_id":   id,
		"severity":    "info",
		"source":      "api",
		"type":        "api.changed",
		"url":         nil,
	}
}

func boolPtr(value bool) *bool {
	return &value
}

func intPtr(value int) *int {
	return &value
}
