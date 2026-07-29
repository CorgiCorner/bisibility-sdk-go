package bisibility

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// FlexibleFloat decodes API numeric fields that may be encoded as JSON numbers or strings.
type FlexibleFloat float64

// UnmarshalJSON implements json.Unmarshaler.
func (f *FlexibleFloat) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var number float64
	if err := json.Unmarshal(data, &number); err == nil {
		*f = FlexibleFloat(number)
		return nil
	}

	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}
	if text == "" {
		return nil
	}
	parsed, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return fmt.Errorf("parse flexible float %q: %w", text, err)
	}

	*f = FlexibleFloat(parsed)
	return nil
}

// Float64 returns the decoded value as a float64.
func (f FlexibleFloat) Float64() float64 {
	return float64(f)
}

// AlertConditionType identifies the condition used by an alert rule.
type AlertConditionType string

const (
	AlertConditionTypeThreshold          AlertConditionType = "threshold"
	AlertConditionTypeChangePct          AlertConditionType = "change_pct"
	AlertConditionTypeEntersTopN         AlertConditionType = "enters_top_n"
	AlertConditionTypeExitsTopN          AlertConditionType = "exits_top_n"
	AlertConditionTypeCompetitorOvertake AlertConditionType = "competitor_overtake"
	AlertConditionTypeSERPFeature        AlertConditionType = "serp_feature"
)

// AlertChannel identifies a delivery channel for alert rules.
type AlertChannel string

const (
	AlertChannelEmail   AlertChannel = "email"
	AlertChannelSlack   AlertChannel = "slack"
	AlertChannelWebhook AlertChannel = "webhook"
)

// AlertTargetType identifies the target set for an alert rule.
type AlertTargetType string

const (
	AlertTargetTypeAll     AlertTargetType = "all"
	AlertTargetTypeKeyword AlertTargetType = "keyword"
	AlertTargetTypeTag     AlertTargetType = "tag"
)

// AlertSeverity is the severity shown for alert rules and triggered alerts.
type AlertSeverity string

const (
	AlertSeverityUrgent  AlertSeverity = "urgent"
	AlertSeverityWarning AlertSeverity = "warning"
	AlertSeverityInfo    AlertSeverity = "info"
)

// AlertRuleStatus is the display status of an alert rule.
type AlertRuleStatus string

const (
	AlertRuleStatusActive   AlertRuleStatus = "active"
	AlertRuleStatusPaused   AlertRuleStatus = "paused"
	AlertRuleStatusLearning AlertRuleStatus = "learning"
	AlertRuleStatusSetup    AlertRuleStatus = "setup"
)

// CreateAlertRuleInput creates an alert rule for a project.
type CreateAlertRuleInput struct {
	Channels          []AlertChannel     `json:"channels,omitempty"`
	ChangePct         *float64           `json:"change_pct,omitempty"`
	CompetitorDomain  *string            `json:"competitor_domain,omitempty"`
	ConditionType     AlertConditionType `json:"condition_type"`
	Enabled           *bool              `json:"enabled,omitempty"`
	Name              string             `json:"name"`
	RecipientIDs      []string           `json:"recipient_ids,omitempty"`
	SERPFeature       *string            `json:"serp_feature,omitempty"`
	TargetIDs         []string           `json:"target_ids,omitempty"`
	TargetType        AlertTargetType    `json:"target_type,omitempty"`
	ThresholdPosition *int               `json:"threshold_position,omitempty"`
	TopN              *int               `json:"top_n,omitempty"`
}

// UpdateAlertRuleInput updates an alert rule.
type UpdateAlertRuleInput = CreateAlertRuleInput

// AlertRule is an alert rule returned by list, create, and update endpoints.
type AlertRule struct {
	ID                string             `json:"id"`
	Name              string             `json:"name"`
	Channel           string             `json:"channel,omitempty"`
	Channels          []AlertChannel     `json:"channels,omitempty"`
	ChangePct         *FlexibleFloat     `json:"change_pct,omitempty"`
	Condition         string             `json:"condition,omitempty"`
	ConditionType     AlertConditionType `json:"condition_type,omitempty"`
	CompetitorDomain  *string            `json:"competitor_domain,omitempty"`
	CreatedAt         *time.Time         `json:"created_at,omitempty"`
	Enabled           bool               `json:"enabled"`
	Fires             string             `json:"fires,omitempty"`
	Period            string             `json:"period,omitempty"`
	RecipientIDs      []string           `json:"recipient_ids"`
	Scope             string             `json:"scope,omitempty"`
	SERPFeature       *string            `json:"serp_feature,omitempty"`
	Severity          AlertSeverity      `json:"severity,omitempty"`
	Status            AlertRuleStatus    `json:"status,omitempty"`
	TargetIDs         []string           `json:"target_ids,omitempty"`
	TargetType        AlertTargetType    `json:"target_type,omitempty"`
	ThresholdPosition *int               `json:"threshold_position,omitempty"`
	TopN              *int               `json:"top_n,omitempty"`
	UpdatedAt         *time.Time         `json:"updated_at,omitempty"`
}

// TriggeredAlert is one alert event returned by ListTriggeredAlerts.
type TriggeredAlert struct {
	Action   string        `json:"action"`
	CTAs     []string      `json:"ctas"`
	Current  string        `json:"current"`
	Headline string        `json:"headline"`
	ID       string        `json:"id"`
	Keyword  string        `json:"keyword"`
	Previous string        `json:"previous"`
	Rule     string        `json:"rule"`
	Severity AlertSeverity `json:"severity"`
	Unread   bool          `json:"unread"`
	When     string        `json:"when"`
}

// TriggeredAlertMuteResult reports the alert snooze applied by MuteTriggeredAlert.
type TriggeredAlertMuteResult struct {
	Muted        bool       `json:"muted"`
	SnoozedUntil *time.Time `json:"snoozed_until"`
}

// TriggeredAlertsReadResult reports how many firing alerts were marked read.
type TriggeredAlertsReadResult struct {
	Updated int `json:"updated"`
}

// AlertRuleDeleteResult is returned after deleting an alert rule.
type AlertRuleDeleteResult struct {
	Deleted bool `json:"deleted"`
}

// TeamRoleValue is the API role value for project team access.
type TeamRoleValue string

const (
	TeamRoleAdmin   TeamRoleValue = "admin"
	TeamRoleAuditor TeamRoleValue = "auditor"
	TeamRoleMember  TeamRoleValue = "member"
	TeamRoleOwner   TeamRoleValue = "owner"
	TeamRoleViewer  TeamRoleValue = "viewer"
)

// TeamMember is one project team member.
type TeamMember struct {
	Color     string        `json:"color"`
	Email     string        `json:"email"`
	ID        string        `json:"id"`
	Initials  string        `json:"initials"`
	Name      string        `json:"name"`
	Role      string        `json:"role"`
	RoleValue TeamRoleValue `json:"role_value"`
}

// TeamInvite is one pending team invite returned by list endpoints.
type TeamInvite struct {
	Email        string        `json:"email"`
	ExpiresLabel string        `json:"expires_label"`
	ID           string        `json:"id"`
	InvitedLabel string        `json:"invited_label"`
	Role         string        `json:"role"`
	RoleValue    TeamRoleValue `json:"role_value"`
}

// SitemapMonitorStatus is the current sitemap monitoring state.
type SitemapMonitorStatus string

const (
	SitemapMonitorStatusActive   SitemapMonitorStatus = "active"
	SitemapMonitorStatusDisabled SitemapMonitorStatus = "disabled"
	SitemapMonitorStatusPending  SitemapMonitorStatus = "pending"
)

// SitemapMonitorLatestSnapshot summarizes the most recent sitemap fetch.
type SitemapMonitorLatestSnapshot struct {
	FetchedAt  time.Time `json:"fetched_at"`
	SitemapURL string    `json:"sitemap_url"`
	URLCount   int       `json:"url_count"`
}

// SitemapMonitor is the project sitemap monitor and latest snapshot state.
type SitemapMonitor struct {
	Enabled        bool                          `json:"enabled"`
	ID             string                        `json:"id"`
	LatestSnapshot *SitemapMonitorLatestSnapshot `json:"latest_snapshot"`
	ProjectID      string                        `json:"project_id"`
	SitemapURL     *string                       `json:"sitemap_url"`
	Status         SitemapMonitorStatus          `json:"status"`
}

// UpdateSitemapMonitorInput enables or disables sitemap monitoring.
type UpdateSitemapMonitorInput struct {
	Enabled bool `json:"enabled"`
}

// CreateTeamInviteInput creates a project team invite.
type CreateTeamInviteInput struct {
	Email string `json:"email"`
	// Role must be TeamRoleAdmin, TeamRoleMember, or TeamRoleViewer. The API
	// rejects TeamRoleOwner for invites; owner only appears in member
	// responses.
	Role TeamRoleValue `json:"role"`
}

// CreatedTeamInvite is returned after creating a project team invite.
type CreatedTeamInvite struct {
	ExpiresAt  time.Time `json:"expires_at"`
	ID         string    `json:"id"`
	InviteLink string    `json:"invite_link"`
}

// RevokeTeamInviteResult is returned after revoking an invite.
type RevokeTeamInviteResult struct {
	ID string `json:"id"`
}

// ProviderID identifies a supported provider. The connectable provider ids
// accepted by connect, test, settings, and disconnect endpoints are
// dataforseo, serpapi, gsc, ga4, and plausible; ahrefs and semrush only
// appear in catalog listings.
type ProviderID string

const (
	ProviderIDDataForSEO ProviderID = "dataforseo"
	ProviderIDSerpAPI    ProviderID = "serpapi"
	ProviderIDGSC        ProviderID = "gsc"
	ProviderIDGA4        ProviderID = "ga4"
	ProviderIDPlausible  ProviderID = "plausible"
	ProviderIDAhrefs     ProviderID = "ahrefs"
	ProviderIDSemrush    ProviderID = "semrush"
)

// ProviderKind identifies the provider category.
type ProviderKind string

const (
	ProviderKindSERP       ProviderKind = "serp"
	ProviderKindAnalytics  ProviderKind = "analytics"
	ProviderKindEnrichment ProviderKind = "enrichment"
)

// ProviderStatus is a provider connection or catalog status.
type ProviderStatus string

const (
	ProviderStatusConnected   ProviderStatus = "connected"
	ProviderStatusNeedsReauth ProviderStatus = "needs_reauth"
	ProviderStatusOptional    ProviderStatus = "optional"
	ProviderStatusPlanned     ProviderStatus = "planned"
	ProviderStatusReady       ProviderStatus = "ready"
)

// ProviderMetaRow is one label/value row on provider catalog responses.
type ProviderMetaRow struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// ProviderCredentialField describes one credential field required by a provider.
type ProviderCredentialField struct {
	Label       string `json:"label"`
	Name        string `json:"name"`
	Placeholder string `json:"placeholder"`
	Type        string `json:"type,omitempty"`
}

// ProviderDrawerDefaults contains default UI values returned with provider catalog entries.
type ProviderDrawerDefaults struct {
	CostPerCheck float64 `json:"cost_per_check"`
	Depth        string  `json:"depth"`
	Device       string  `json:"device"`
	Enabled      *bool   `json:"enabled,omitempty"`
	Language     string  `json:"language"`
	Location     string  `json:"location"`
	Login        string  `json:"login"`
	Primary      bool    `json:"primary"`
	Priority     *int    `json:"priority,omitempty"`
	Secret       string  `json:"secret"`
}

// ProviderDrawer contains provider configuration metadata.
type ProviderDrawer struct {
	Activities         []ProviderMetaRow         `json:"activities"`
	CostHelp           string                    `json:"cost_help"`
	CredentialFields   []ProviderCredentialField `json:"credential_fields"`
	Defaults           ProviderDrawerDefaults    `json:"defaults"`
	EnvHint            string                    `json:"env_hint"`
	PrimaryToggleLabel string                    `json:"primary_toggle_label"`
}

// Provider is one provider catalog item returned by ListProviders.
type Provider struct {
	CategoryID      string            `json:"category_id"`
	CategoryTitle   string            `json:"category_title"`
	Description     string            `json:"description"`
	Drawer          ProviderDrawer    `json:"drawer"`
	Enabled         *bool             `json:"enabled,omitempty"`
	Icon            string            `json:"icon"`
	ID              ProviderID        `json:"id"`
	LogoDomain      string            `json:"logo_domain,omitempty"`
	Meta            []ProviderMetaRow `json:"meta"`
	Name            string            `json:"name"`
	Primary         *bool             `json:"primary,omitempty"`
	Priority        *int              `json:"priority,omitempty"`
	SecondaryAction string            `json:"secondary_action,omitempty"`
	Status          ProviderStatus    `json:"status"`
	Tint            string            `json:"tint"`
}

// ProviderCredentialsInput contains provider credentials for connect and test requests.
type ProviderCredentialsInput struct {
	APIKey   string `json:"api_key,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
	Login    string `json:"login,omitempty"`
	Secret   string `json:"secret,omitempty"`
}

// ConnectProviderInput connects or updates provider credentials for a project.
type ConnectProviderInput struct {
	CostPerCheck *float64                  `json:"cost_per_check,omitempty"`
	Credentials  *ProviderCredentialsInput `json:"credentials,omitempty"`
	Enabled      *bool                     `json:"enabled,omitempty"`
	Login        string                    `json:"login,omitempty"`
	Primary      *bool                     `json:"primary,omitempty"`
	Priority     *int                      `json:"priority,omitempty"`
	Secret       string                    `json:"secret,omitempty"`
}

// TestProviderConnectionInput tests provider credentials.
type TestProviderConnectionInput struct {
	Credentials *ProviderCredentialsInput `json:"credentials,omitempty"`
	Login       string                    `json:"login,omitempty"`
	Secret      string                    `json:"secret,omitempty"`
}

// ProviderSettingsInput updates provider enabled, primary, and priority settings.
type ProviderSettingsInput struct {
	Enabled  *bool `json:"enabled,omitempty"`
	Primary  *bool `json:"primary,omitempty"`
	Priority *int  `json:"priority,omitempty"`
}

// ProviderConnection is returned by provider connect and settings endpoints.
type ProviderConnection struct {
	CostPerCheckCents *FlexibleFloat `json:"cost_per_check_cents,omitempty"`
	CreatedAt         *time.Time     `json:"created_at,omitempty"`
	CredentialsHash   *string        `json:"credentials_hash,omitempty"`
	Enabled           bool           `json:"enabled"`
	ID                string         `json:"id"`
	IsPrimary         bool           `json:"is_primary"`
	Kind              ProviderKind   `json:"kind"`
	LastUsedAt        *time.Time     `json:"last_used_at,omitempty"`
	Priority          int            `json:"priority"`
	ProjectID         string         `json:"project_id"`
	Provider          ProviderID     `json:"provider"`
	Status            ProviderStatus `json:"status"`
	UpdatedAt         *time.Time     `json:"updated_at,omitempty"`
}

// ProviderTestResult is returned by TestProviderConnection.
type ProviderTestResult struct {
	Balance *float64 `json:"balance,omitempty"`
	Message string   `json:"message"`
	OK      bool     `json:"ok"`
}

// ProviderDisconnectResult is returned by DisconnectProvider.
type ProviderDisconnectResult struct {
	OK bool `json:"ok"`
}

// SavedViewPositionBucket identifies a saved-view position bucket.
type SavedViewPositionBucket string

const (
	SavedViewPositionTop3    SavedViewPositionBucket = "top3"
	SavedViewPositionTop10   SavedViewPositionBucket = "top10"
	SavedViewPosition11To50  SavedViewPositionBucket = "11-50"
	SavedViewPosition51To100 SavedViewPositionBucket = "51-100"
)

// SavedViewFilters are keyword grid filters stored in a saved view.
type SavedViewFilters struct {
	Change   string                    `json:"change,omitempty"`
	Contains string                    `json:"contains,omitempty"`
	Country  string                    `json:"country,omitempty"`
	Device   string                    `json:"device,omitempty"`
	Position []SavedViewPositionBucket `json:"position,omitempty"`
	SERP     []string                  `json:"serp,omitempty"`
	Tags     []string                  `json:"tags,omitempty"`
	VolMax   int                       `json:"vol_max,omitempty"`
	VolMin   int                       `json:"vol_min,omitempty"`
	WrongURL bool                      `json:"wrong_url,omitempty"`
}

// SavedViewConfig is the keyword grid state stored in a saved view.
type SavedViewConfig struct {
	Filters SavedViewFilters `json:"filters"`
	Search  string           `json:"search,omitempty"`
}

// SavedView is a keyword saved view.
type SavedView struct {
	Config      SavedViewConfig `json:"config"`
	CreatedAt   time.Time       `json:"created_at"`
	CreatedByID *string         `json:"created_by_id"`
	ID          string          `json:"id"`
	Name        string          `json:"name"`
}

// CreateSavedViewInput creates a project saved view.
type CreateSavedViewInput struct {
	Config SavedViewConfig `json:"config"`
	Name   string          `json:"name"`
}

// SavedViewDeleteResult is returned after deleting a saved view.
type SavedViewDeleteResult struct {
	Deleted bool `json:"deleted"`
}

// ManagedCompetitor is a competitor item in the project competitor list.
type ManagedCompetitor struct {
	Domain   string `json:"domain"`
	ID       string `json:"id"`
	Initials string `json:"initials,omitempty"`
	Label    string `json:"label"`
}

// Competitor is returned by competitor write endpoints.
type Competitor struct {
	Domain string  `json:"domain"`
	ID     string  `json:"id"`
	Label  *string `json:"label"`
}

// AddCompetitorInput adds a managed competitor to a project.
type AddCompetitorInput struct {
	Domain string `json:"domain"`
	Label  string `json:"label,omitempty"`
}

// CompetitorColumn is one column in a competitor market response.
type CompetitorColumn struct {
	Domain string `json:"domain"`
	ID     string `json:"id,omitempty"`
	Kind   string `json:"kind"`
	Label  string `json:"label"`
}

// CompetitorShare is one share-of-voice entry in a competitor market response.
type CompetitorShare struct {
	Color          string `json:"color"`
	Domain         string `json:"domain"`
	ID             string `json:"id,omitempty"`
	Initials       string `json:"initials"`
	Kind           string `json:"kind"`
	Label          string `json:"label"`
	ShareOfVoice   int    `json:"share_of_voice"`
	SharedKeywords int    `json:"shared_keywords"`
}

// CompetitorHeadToHeadRow is one keyword row in a competitor market response.
type CompetitorHeadToHeadRow struct {
	Gap     *int            `json:"gap"`
	Keyword string          `json:"keyword"`
	Ranks   map[string]*int `json:"ranks"`
}

// CompetitorMarket summarizes competitors for one country, device, and engine.
type CompetitorMarket struct {
	CheckedKeywordCount int                       `json:"checked_keyword_count"`
	Columns             []CompetitorColumn        `json:"columns"`
	CompetitorCount     int                       `json:"competitor_count"`
	Country             string                    `json:"country"`
	Device              string                    `json:"device"`
	Engine              string                    `json:"engine"`
	HasRankData         bool                      `json:"has_rank_data"`
	Key                 string                    `json:"key"`
	Rows                []CompetitorHeadToHeadRow `json:"rows"`
	Shares              []CompetitorShare         `json:"shares"`
	SharedKeywordCount  int                       `json:"shared_keyword_count"`
	TrackedKeywordCount int                       `json:"tracked_keyword_count"`
}

// SuggestedCompetitor is a competitor suggestion returned in list metadata.
type SuggestedCompetitor struct {
	Domain   string `json:"domain"`
	Initials string `json:"initials"`
	Overlap  int    `json:"overlap"`
}

// CompetitorsMeta contains list pagination plus competitor market metadata.
type CompetitorsMeta struct {
	NextCursor  *string               `json:"next_cursor"`
	Markets     []CompetitorMarket    `json:"markets"`
	Suggestions []SuggestedCompetitor `json:"suggestions"`
}

// ListCompetitorsResponse is returned by ListCompetitors.
type ListCompetitorsResponse struct {
	Data []ManagedCompetitor `json:"data"`
	Meta CompetitorsMeta     `json:"meta"`
}

// CompetitorRemoveResult is returned after removing a competitor.
type CompetitorRemoveResult struct {
	Removed bool `json:"removed"`
}

// NotificationPreferences are the full notification preferences returned by GET.
type NotificationPreferences struct {
	AlertEmail        bool   `json:"alert_email"`
	AlertInApp        bool   `json:"alert_in_app"`
	AlertSlack        bool   `json:"alert_slack"`
	AlertWebhook      bool   `json:"alert_webhook"`
	CheckEmail        bool   `json:"check_email"`
	CheckInApp        bool   `json:"check_in_app"`
	Email             string `json:"email"`
	EmailVerification string `json:"email_verification"`
	ImportEmail       bool   `json:"import_email"`
	ImportInApp       bool   `json:"import_in_app"`
	InviteEmail       bool   `json:"invite_email"`
	InviteInApp       bool   `json:"invite_in_app"`
	ProjectID         string `json:"project_id"`
	SlackAvailable    bool   `json:"slack_available"`
	WebhookAvailable  bool   `json:"webhook_available"`
}

// UpdateNotificationPreferencesInput patches notification preferences.
type UpdateNotificationPreferencesInput struct {
	AlertEmail   *bool `json:"alert_email,omitempty"`
	AlertInApp   *bool `json:"alert_in_app,omitempty"`
	AlertSlack   *bool `json:"alert_slack,omitempty"`
	AlertWebhook *bool `json:"alert_webhook,omitempty"`
	CheckEmail   *bool `json:"check_email,omitempty"`
	CheckInApp   *bool `json:"check_in_app,omitempty"`
	ImportEmail  *bool `json:"import_email,omitempty"`
	ImportInApp  *bool `json:"import_in_app,omitempty"`
	InviteEmail  *bool `json:"invite_email,omitempty"`
	InviteInApp  *bool `json:"invite_in_app,omitempty"`
}

// UpdatedNotificationPreferences is returned after a preferences patch.
type UpdatedNotificationPreferences struct {
	AlertEmail   bool   `json:"alert_email"`
	AlertInApp   bool   `json:"alert_in_app"`
	AlertSlack   bool   `json:"alert_slack"`
	AlertWebhook bool   `json:"alert_webhook"`
	CheckEmail   bool   `json:"check_email"`
	CheckInApp   bool   `json:"check_in_app"`
	ImportEmail  bool   `json:"import_email"`
	ImportInApp  bool   `json:"import_in_app"`
	InviteEmail  bool   `json:"invite_email"`
	InviteInApp  bool   `json:"invite_in_app"`
	ProjectID    string `json:"project_id"`
}

// MigrationScope selects which data a migration token can import.
type MigrationScope string

const (
	MigrationScopeFull     MigrationScope = "full"
	MigrationScopeKeywords MigrationScope = "keywords"
)

// CloudImportState is the current state of a migration import job.
type CloudImportState string

const (
	CloudImportStateIdle      CloudImportState = "idle"
	CloudImportStateReceiving CloudImportState = "receiving"
	CloudImportStateImporting CloudImportState = "importing"
	CloudImportStateDone      CloudImportState = "done"
	CloudImportStateFailed    CloudImportState = "failed"
)

// MigrationTokenCreatedBy identifies who minted an active migration token.
type MigrationTokenCreatedBy struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// MigrationToken is an active migration token without the raw token value.
type MigrationToken struct {
	CreatedAt time.Time               `json:"created_at"`
	CreatedBy MigrationTokenCreatedBy `json:"created_by"`
	ExpiresAt time.Time               `json:"expires_at"`
	ID        string                  `json:"id"`
	Scope     MigrationScope          `json:"scope"`
	SingleUse bool                    `json:"single_use"`
}

// CloudImportJob describes migration import job status.
type CloudImportJob struct {
	Counts     json.RawMessage  `json:"counts"`
	CreatedAt  *time.Time       `json:"created_at"`
	Error      *string          `json:"error"`
	FinishedAt *time.Time       `json:"finished_at"`
	ID         *string          `json:"id"`
	Progress   int              `json:"progress"`
	StartedAt  *time.Time       `json:"started_at"`
	State      CloudImportState `json:"state"`
}

// MigrationTokensMeta contains list pagination plus import job metadata.
type MigrationTokensMeta struct {
	NextCursor *string        `json:"next_cursor"`
	ImportJob  CloudImportJob `json:"import_job"`
}

// ListMigrationTokensResponse is returned by ListMigrationTokens.
type ListMigrationTokensResponse struct {
	Data []MigrationToken    `json:"data"`
	Meta MigrationTokensMeta `json:"meta"`
}

// MintMigrationTokenInput creates a migration token.
type MintMigrationTokenInput struct {
	Scope MigrationScope `json:"scope,omitempty"`
}

// IssuedMigrationToken is returned once when minting a migration token.
type IssuedMigrationToken struct {
	CreatedAt time.Time      `json:"created_at"`
	ExpiresAt time.Time      `json:"expires_at"`
	ID        string         `json:"id"`
	ImportJob CloudImportJob `json:"import_job"`
	Scope     MigrationScope `json:"scope"`
	SingleUse bool           `json:"single_use"`
	Token     string         `json:"token"`
}

// RevokedMigrationToken is returned after revoking a migration token.
type RevokedMigrationToken struct {
	ID        string    `json:"id"`
	RevokedAt time.Time `json:"revoked_at"`
}

// SignalSource identifies where a signal originated. CreateSignal only
// accepts SignalSourceDeploy, SignalSourceCMS, and SignalSourceAPI; the
// remaining sources are emitted by Bisibility itself and appear in list
// responses and list filters.
type SignalSource string

const (
	SignalSourceRankTracker        SignalSource = "rank_tracker"
	SignalSourceSearchAnalytics    SignalSource = "search_analytics"
	SignalSourceURLInspection      SignalSource = "url_inspection"
	SignalSourceSitemap            SignalSource = "sitemap"
	SignalSourceDeploy             SignalSource = "deploy"
	SignalSourceCMS                SignalSource = "cms"
	SignalSourceSearchEngineStatus SignalSource = "search_engine_status"
	SignalSourceManual             SignalSource = "manual"
	SignalSourceAPI                SignalSource = "api"
)

// SignalSeverity is the severity attached to a signal.
type SignalSeverity string

const (
	SignalSeverityInfo     SignalSeverity = "info"
	SignalSeverityWarning  SignalSeverity = "warning"
	SignalSeverityCritical SignalSeverity = "critical"
)

// Signal is one ingested signal event.
type Signal struct {
	CreatedAt  time.Time      `json:"created_at"`
	HappenedAt time.Time      `json:"happened_at"`
	ID         string         `json:"id"`
	KeywordID  *string        `json:"keyword_id"`
	Payload    JSONValue      `json:"payload"`
	ProjectID  string         `json:"project_id"`
	PublicID   string         `json:"public_id"`
	Severity   SignalSeverity `json:"severity"`
	Source     SignalSource   `json:"source"`
	Type       string         `json:"type"`
	URL        *string        `json:"url"`
}

// CreateSignalInput ingests a signal for the API key's project. Source must
// be SignalSourceDeploy, SignalSourceCMS, or SignalSourceAPI. Type must match
// ^[a-z_]+\.[a-z_]+$ (for example deploy.completed). HappenedAt defaults to
// the ingestion time and Severity defaults to info when omitted. Payload must
// serialize to 8KB or less and URL must use http or https.
type CreateSignalInput struct {
	HappenedAt *time.Time     `json:"happened_at,omitempty"`
	KeywordID  string         `json:"keyword_id,omitempty"`
	Payload    JSONValue      `json:"payload,omitempty"`
	Severity   SignalSeverity `json:"severity,omitempty"`
	Source     SignalSource   `json:"source"`
	Type       string         `json:"type"`
	URL        string         `json:"url,omitempty"`
}

// ListSignalsOptions filters project signals. Signals are returned newest
// first by happened_at.
type ListSignalsOptions struct {
	Cursor string
	Limit  int
	From   time.Time
	Source SignalSource
	To     time.Time
	Type   string
}
