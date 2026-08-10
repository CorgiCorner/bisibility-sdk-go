package bisibility

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

const (
	teamInvitesResource     = "team/invites"
	savedKeywordsResource   = "saved-keywords"
	savedViewsResource      = "saved-views"
	migrationTokensResource = "migration-tokens"
)

func projectResourcePath(projectID, resource string) string {
	return "/projects/" + url.PathEscape(projectID) + "/" + resource
}

func projectMemberResourcePath(projectID, resource, resourceID string) string {
	return projectResourcePath(projectID, resource) + "/" + url.PathEscape(resourceID)
}

// ListAlertRules lists alert rules for a project.
func (c *Client) ListAlertRules(ctx context.Context, projectID string, pagination *PaginationOptions, options ...RequestOption) (*ListResponse[AlertRule], error) {
	config := newRequestConfig(options...)
	addPagination(config.query, pagination)
	return requestJSON[ListResponse[AlertRule]](c, ctx, http.MethodGet, projectResourcePath(projectID, "alert-rules"), config)
}

// CreateAlertRule creates an alert rule for a project.
func (c *Client) CreateAlertRule(ctx context.Context, projectID string, input CreateAlertRuleInput, options ...RequestOption) (*AlertRule, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[AlertRule](c, ctx, http.MethodPost, projectResourcePath(projectID, "alert-rules"), config)
}

// UpdateAlertRule updates an alert rule by ID.
func (c *Client) UpdateAlertRule(ctx context.Context, ruleID string, input UpdateAlertRuleInput, options ...RequestOption) (*AlertRule, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[AlertRule](c, ctx, http.MethodPatch, "/alert-rules/"+url.PathEscape(ruleID), config)
}

// DeleteAlertRule deletes an alert rule by ID.
func (c *Client) DeleteAlertRule(ctx context.Context, ruleID string, options ...RequestOption) (*AlertRuleDeleteResult, error) {
	return requestJSON[AlertRuleDeleteResult](c, ctx, http.MethodDelete, "/alert-rules/"+url.PathEscape(ruleID), newRequestConfig(options...))
}

// ListTriggeredAlerts lists triggered alert events for a project.
func (c *Client) ListTriggeredAlerts(ctx context.Context, projectID string, pagination *PaginationOptions, options ...RequestOption) (*ListResponse[TriggeredAlert], error) {
	config := newRequestConfig(options...)
	addPagination(config.query, pagination)
	return requestJSON[ListResponse[TriggeredAlert]](c, ctx, http.MethodGet, projectResourcePath(projectID, "triggered-alerts"), config)
}

// MuteTriggeredAlert mutes one triggered alert for 24 hours.
func (c *Client) MuteTriggeredAlert(ctx context.Context, projectID, alertID string, options ...RequestOption) (*TriggeredAlertMuteResult, error) {
	path := projectMemberResourcePath(projectID, "triggered-alerts", alertID) + "/mute"
	return requestJSON[TriggeredAlertMuteResult](c, ctx, http.MethodPost, path, newRequestConfig(options...))
}

// MarkProjectAlertsRead marks every firing alert in a project as read.
func (c *Client) MarkProjectAlertsRead(ctx context.Context, projectID string, options ...RequestOption) (*TriggeredAlertsReadResult, error) {
	return requestJSON[TriggeredAlertsReadResult](c, ctx, http.MethodPost, projectResourcePath(projectID, "triggered-alerts/mark-read"), newRequestConfig(options...))
}

// ListTeamMembers lists team members for a project.
func (c *Client) ListTeamMembers(ctx context.Context, projectID string, pagination *PaginationOptions, options ...RequestOption) (*ListResponse[TeamMember], error) {
	config := newRequestConfig(options...)
	addPagination(config.query, pagination)
	return requestJSON[ListResponse[TeamMember]](c, ctx, http.MethodGet, projectResourcePath(projectID, "team/members"), config)
}

// ListTeamInvites lists pending team invites for a project.
func (c *Client) ListTeamInvites(ctx context.Context, projectID string, pagination *PaginationOptions, options ...RequestOption) (*ListResponse[TeamInvite], error) {
	config := newRequestConfig(options...)
	addPagination(config.query, pagination)
	return requestJSON[ListResponse[TeamInvite]](c, ctx, http.MethodGet, projectResourcePath(projectID, teamInvitesResource), config)
}

// ListSitemapMonitors lists the project sitemap monitor and latest snapshot.
func (c *Client) ListSitemapMonitors(ctx context.Context, projectID string, options ...RequestOption) (*ListResponse[SitemapMonitor], error) {
	return requestJSON[ListResponse[SitemapMonitor]](c, ctx, http.MethodGet, projectResourcePath(projectID, "sitemap-monitors"), newRequestConfig(options...))
}

// UpdateSitemapMonitor enables or disables a project sitemap monitor.
func (c *Client) UpdateSitemapMonitor(ctx context.Context, projectID, monitorID string, input UpdateSitemapMonitorInput, options ...RequestOption) (*SitemapMonitor, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[SitemapMonitor](c, ctx, http.MethodPatch, projectMemberResourcePath(projectID, "sitemap-monitors", monitorID), config)
}

// CreateTeamInvite creates a team invite for a project.
func (c *Client) CreateTeamInvite(ctx context.Context, projectID string, input CreateTeamInviteInput, options ...RequestOption) (*CreatedTeamInvite, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[CreatedTeamInvite](c, ctx, http.MethodPost, projectResourcePath(projectID, teamInvitesResource), config)
}

// RevokeTeamInvite revokes a team invite by ID using the top-level route.
func (c *Client) RevokeTeamInvite(ctx context.Context, inviteID string, options ...RequestOption) (*RevokeTeamInviteResult, error) {
	return requestJSON[RevokeTeamInviteResult](c, ctx, http.MethodDelete, "/team/invites/"+url.PathEscape(inviteID), newRequestConfig(options...))
}

// RevokeProjectTeamInvite revokes a team invite by project and invite ID.
func (c *Client) RevokeProjectTeamInvite(ctx context.Context, projectID, inviteID string, options ...RequestOption) (*RevokeTeamInviteResult, error) {
	return requestJSON[RevokeTeamInviteResult](c, ctx, http.MethodDelete, projectMemberResourcePath(projectID, teamInvitesResource, inviteID), newRequestConfig(options...))
}

// ListProviders lists available and connected providers for a project.
func (c *Client) ListProviders(ctx context.Context, projectID string, pagination *PaginationOptions, options ...RequestOption) (*ListResponse[Provider], error) {
	config := newRequestConfig(options...)
	addPagination(config.query, pagination)
	return requestJSON[ListResponse[Provider]](c, ctx, http.MethodGet, projectResourcePath(projectID, "providers"), config)
}

// ConnectProvider connects or updates credentials for a project provider.
func (c *Client) ConnectProvider(ctx context.Context, projectID string, providerID ProviderID, input ConnectProviderInput, options ...RequestOption) (*ProviderConnection, error) {
	config := newRequestConfig(options...)
	connectInput := input
	connectInput.Primary = nil
	connectInput.Priority = nil
	config.body = connectInput
	path := projectMemberResourcePath(projectID, "providers", string(providerID)) + "/connect"
	connection, err := requestJSON[ProviderConnection](c, ctx, http.MethodPost, path, config)
	if err != nil {
		return nil, err
	}

	priority, shouldSyncPriority := connectProviderPriority(input)
	if !shouldSyncPriority {
		return connection, nil
	}
	updated, err := c.syncConnectedProviderPriority(ctx, projectID, providerID, priority, options...)
	if err != nil {
		return connection, &ProviderPrioritySyncError{Cause: err}
	}
	if updated == nil {
		return connection, nil
	}
	return updated, nil
}

func connectProviderPriority(input ConnectProviderInput) (int, bool) {
	if input.Primary != nil && *input.Primary {
		return 0, true
	}
	if input.Priority != nil {
		return *input.Priority, true
	}
	return 0, false
}

func (c *Client) syncConnectedProviderPriority(ctx context.Context, projectID string, providerID ProviderID, priority int, options ...RequestOption) (*ProviderConnection, error) {
	config := newRequestConfig(options...)
	config.body = ProviderSettingsInput{Priority: &priority}
	config.headers.Del("Idempotency-Key")
	config.idempotencyKey = ""
	config.omitIdempotencyKey = true
	return requestJSON[ProviderConnection](c, ctx, http.MethodPatch, projectMemberResourcePath(projectID, "providers", string(providerID)), config)
}

// TestProviderConnection tests credentials or a stored connection for a project provider.
func (c *Client) TestProviderConnection(ctx context.Context, projectID string, providerID ProviderID, input TestProviderConnectionInput, options ...RequestOption) (*ProviderTestResult, error) {
	config := newRequestConfig(options...)
	config.body = input
	path := projectMemberResourcePath(projectID, "providers", string(providerID)) + "/test"
	return requestJSON[ProviderTestResult](c, ctx, http.MethodPost, path, config)
}

// UpdateProviderSettings updates enabled, primary, or priority settings for a provider.
func (c *Client) UpdateProviderSettings(ctx context.Context, projectID string, providerID ProviderID, input ProviderSettingsInput, options ...RequestOption) (*ProviderConnection, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[ProviderConnection](c, ctx, http.MethodPatch, projectMemberResourcePath(projectID, "providers", string(providerID)), config)
}

// SetProviderEnabled enables or disables a provider connection.
func (c *Client) SetProviderEnabled(ctx context.Context, projectID string, providerID ProviderID, enabled bool, options ...RequestOption) (*ProviderConnection, error) {
	return c.UpdateProviderSettings(ctx, projectID, providerID, ProviderSettingsInput{Enabled: &enabled}, options...)
}

// SetProviderPriority sets a provider fallback priority.
func (c *Client) SetProviderPriority(ctx context.Context, projectID string, providerID ProviderID, priority int, options ...RequestOption) (*ProviderConnection, error) {
	return c.UpdateProviderSettings(ctx, projectID, providerID, ProviderSettingsInput{Priority: &priority}, options...)
}

// SetPrimaryProvider marks or unmarks a provider as primary.
func (c *Client) SetPrimaryProvider(ctx context.Context, projectID string, providerID ProviderID, primary bool, options ...RequestOption) (*ProviderConnection, error) {
	return c.UpdateProviderSettings(ctx, projectID, providerID, ProviderSettingsInput{Primary: &primary}, options...)
}

// DisconnectProvider disconnects a provider from a project.
func (c *Client) DisconnectProvider(ctx context.Context, projectID string, providerID ProviderID, options ...RequestOption) (*ProviderDisconnectResult, error) {
	return requestJSON[ProviderDisconnectResult](c, ctx, http.MethodDelete, projectMemberResourcePath(projectID, "providers", string(providerID)), newRequestConfig(options...))
}

// ListSavedViews lists keyword saved views for a project.
func (c *Client) ListSavedViews(ctx context.Context, projectID string, pagination *PaginationOptions, options ...RequestOption) (*ListResponse[SavedView], error) {
	config := newRequestConfig(options...)
	addPagination(config.query, pagination)
	return requestJSON[ListResponse[SavedView]](c, ctx, http.MethodGet, projectResourcePath(projectID, savedViewsResource), config)
}

// CreateSavedView creates a keyword saved view for a project.
func (c *Client) CreateSavedView(ctx context.Context, projectID string, input CreateSavedViewInput, options ...RequestOption) (*SavedView, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[SavedView](c, ctx, http.MethodPost, projectResourcePath(projectID, savedViewsResource), config)
}

// DeleteSavedView deletes a saved view by ID using the top-level route.
func (c *Client) DeleteSavedView(ctx context.Context, viewID string, options ...RequestOption) (*SavedViewDeleteResult, error) {
	return requestJSON[SavedViewDeleteResult](c, ctx, http.MethodDelete, "/saved-views/"+url.PathEscape(viewID), newRequestConfig(options...))
}

// DeleteProjectSavedView deletes a saved view by project and view ID.
func (c *Client) DeleteProjectSavedView(ctx context.Context, projectID, viewID string, options ...RequestOption) (*SavedViewDeleteResult, error) {
	return requestJSON[SavedViewDeleteResult](c, ctx, http.MethodDelete, projectMemberResourcePath(projectID, savedViewsResource, viewID), newRequestConfig(options...))
}

// ListSavedKeywords lists saved research keywords for a project.
func (c *Client) ListSavedKeywords(ctx context.Context, projectID string, pagination *PaginationOptions, options ...RequestOption) (*ListResponse[SavedKeyword], error) {
	config := newRequestConfig(options...)
	addPagination(config.query, pagination)
	return requestJSON[ListResponse[SavedKeyword]](c, ctx, http.MethodGet, projectResourcePath(projectID, savedKeywordsResource), config)
}

// CreateSavedKeywords saves research keywords for a project. Keywords already
// saved are reported as duplicates instead of failing the request.
func (c *Client) CreateSavedKeywords(ctx context.Context, projectID string, input CreateSavedKeywordsInput, options ...RequestOption) (*CreateSavedKeywordsResult, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[CreateSavedKeywordsResult](c, ctx, http.MethodPost, projectResourcePath(projectID, savedKeywordsResource), config)
}

// DeleteProjectSavedKeyword deletes a saved keyword by project and saved keyword ID.
func (c *Client) DeleteProjectSavedKeyword(ctx context.Context, projectID, savedKeywordID string, options ...RequestOption) (*SavedKeywordDeleteResult, error) {
	return requestJSON[SavedKeywordDeleteResult](c, ctx, http.MethodDelete, projectMemberResourcePath(projectID, savedKeywordsResource, savedKeywordID), newRequestConfig(options...))
}

// ListCompetitors lists managed competitors and market metadata for a project.
func (c *Client) ListCompetitors(ctx context.Context, projectID string, pagination *PaginationOptions, options ...RequestOption) (*ListCompetitorsResponse, error) {
	config := newRequestConfig(options...)
	addPagination(config.query, pagination)
	return requestJSON[ListCompetitorsResponse](c, ctx, http.MethodGet, projectResourcePath(projectID, "competitors"), config)
}

// AddCompetitor adds a managed competitor to a project.
func (c *Client) AddCompetitor(ctx context.Context, projectID string, input AddCompetitorInput, options ...RequestOption) (*Competitor, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[Competitor](c, ctx, http.MethodPost, projectResourcePath(projectID, "competitors"), config)
}

// RemoveCompetitor removes a competitor by ID using the top-level route.
func (c *Client) RemoveCompetitor(ctx context.Context, competitorID string, options ...RequestOption) (*CompetitorRemoveResult, error) {
	return requestJSON[CompetitorRemoveResult](c, ctx, http.MethodDelete, "/competitors/"+url.PathEscape(competitorID), newRequestConfig(options...))
}

// RemoveProjectCompetitor removes a competitor by project and competitor ID.
func (c *Client) RemoveProjectCompetitor(ctx context.Context, projectID, competitorID string, options ...RequestOption) (*CompetitorRemoveResult, error) {
	return requestJSON[CompetitorRemoveResult](c, ctx, http.MethodDelete, projectMemberResourcePath(projectID, "competitors", competitorID), newRequestConfig(options...))
}

// GetNotificationPreferences gets project notification preferences for the current user.
func (c *Client) GetNotificationPreferences(ctx context.Context, projectID string, options ...RequestOption) (*NotificationPreferences, error) {
	return requestJSON[NotificationPreferences](c, ctx, http.MethodGet, projectResourcePath(projectID, "notification-preferences"), newRequestConfig(options...))
}

// UpdateNotificationPreferences patches project notification preferences for the current user.
func (c *Client) UpdateNotificationPreferences(ctx context.Context, projectID string, input UpdateNotificationPreferencesInput, options ...RequestOption) (*UpdatedNotificationPreferences, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[UpdatedNotificationPreferences](c, ctx, http.MethodPatch, projectResourcePath(projectID, "notification-preferences"), config)
}

// ListMigrationTokens lists active migration tokens and import job metadata for a project.
func (c *Client) ListMigrationTokens(ctx context.Context, projectID string, options ...RequestOption) (*ListMigrationTokensResponse, error) {
	return requestJSON[ListMigrationTokensResponse](c, ctx, http.MethodGet, projectResourcePath(projectID, migrationTokensResource), newRequestConfig(options...))
}

// MintMigrationToken mints a migration token for a project.
func (c *Client) MintMigrationToken(ctx context.Context, projectID string, input MintMigrationTokenInput, options ...RequestOption) (*IssuedMigrationToken, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[IssuedMigrationToken](c, ctx, http.MethodPost, projectResourcePath(projectID, migrationTokensResource), config)
}

// RevokeMigrationToken revokes a migration token by ID using the top-level route.
func (c *Client) RevokeMigrationToken(ctx context.Context, tokenID string, options ...RequestOption) (*RevokedMigrationToken, error) {
	return requestJSON[RevokedMigrationToken](c, ctx, http.MethodDelete, "/migration-tokens/"+url.PathEscape(tokenID), newRequestConfig(options...))
}

// RevokeProjectMigrationToken revokes a migration token by project and token ID.
func (c *Client) RevokeProjectMigrationToken(ctx context.Context, projectID, tokenID string, options ...RequestOption) (*RevokedMigrationToken, error) {
	return requestJSON[RevokedMigrationToken](c, ctx, http.MethodDelete, projectMemberResourcePath(projectID, migrationTokensResource, tokenID), newRequestConfig(options...))
}

// CreateSignal ingests a signal for the API key's project. The API responds
// 201 with the created signal resource.
func (c *Client) CreateSignal(ctx context.Context, input CreateSignalInput, options ...RequestOption) (*Signal, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[Signal](c, ctx, http.MethodPost, "/signals", config)
}

// ListSignals lists signals for a project, newest first.
func (c *Client) ListSignals(ctx context.Context, projectID string, filters *ListSignalsOptions, options ...RequestOption) (*ListResponse[Signal], error) {
	config := newRequestConfig(options...)
	if filters != nil {
		addPagination(config.query, &PaginationOptions{Cursor: filters.Cursor, Limit: filters.Limit})
		if !filters.From.IsZero() {
			config.query.Set("from", filters.From.Format(time.RFC3339Nano))
		}
		addQuery(config.query, "source", string(filters.Source))
		if !filters.To.IsZero() {
			config.query.Set("to", filters.To.Format(time.RFC3339Nano))
		}
		addQuery(config.query, "type", filters.Type)
	}

	return requestJSON[ListResponse[Signal]](c, ctx, http.MethodGet, projectResourcePath(projectID, "signals"), config)
}
