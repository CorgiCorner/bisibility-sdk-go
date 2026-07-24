package bisibility

// The cloud-import family lets a caller migrate an exported package into a
// Bisibility project. importCloudExport, createCloudImportSession,
// uploadCloudImportChunk, and finalizeCloudImportSession authenticate with a
// migration token (Authorization: Bearer mig_...) minted by
// MintMigrationToken; getCloudImportCompatibility is an unauthenticated
// preflight. Shapes mirror the CloudImport* components in the public OpenAPI
// document.

// CloudImportCompatibility is the schema-compatibility preflight for cloud
// imports. It reports which package schema versions the server accepts.
type CloudImportCompatibility struct {
	AppVersion              string  `json:"app_version"`
	LatestMigration         *string `json:"latest_migration"`
	SchemaVersionsSupported []int   `json:"schema_versions_supported"`
}

// CloudImportCounts maps an imported resource name to the number of records
// created for it.
type CloudImportCounts map[string]int

// CloudImportFinalizeResponse is returned by importCloudExport and
// finalizeCloudImportSession once an import completes.
type CloudImportFinalizeResponse struct {
	Counts CloudImportCounts `json:"counts"`
	JobID  string            `json:"job_id"`
	State  string            `json:"state"`
}

// CloudImportAlertRuleTarget scopes a migrated alert rule to a keyword or tag.
// Fields accept both snake_case and camelCase source keys; only the fields
// present in the source package are sent.
type CloudImportAlertRuleTarget struct {
	Device    string `json:"device,omitempty"`
	Keyword   string `json:"keyword,omitempty"`
	KeywordID string `json:"keyword_id,omitempty"`
	Location  string `json:"location,omitempty"`
	Tag       string `json:"tag,omitempty"`
	TagID     string `json:"tag_id,omitempty"`
	Type      string `json:"type,omitempty"`
}

// CloudImportAlertRule is a migrated alert rule. Name is required.
type CloudImportAlertRule struct {
	ChangePct         *float64                     `json:"change_pct,omitempty"`
	Channels          []string                     `json:"channels,omitempty"`
	CompetitorDomain  *string                      `json:"competitor_domain,omitempty"`
	ConditionType     string                       `json:"condition_type,omitempty"`
	Enabled           *bool                        `json:"enabled,omitempty"`
	Name              string                       `json:"name"`
	SerpFeature       *string                      `json:"serp_feature,omitempty"`
	TargetType        string                       `json:"target_type,omitempty"`
	Targets           []CloudImportAlertRuleTarget `json:"targets,omitempty"`
	ThresholdPosition *int                         `json:"threshold_position,omitempty"`
	TopN              *int                         `json:"top_n,omitempty"`
}

// CloudImportCompetitor is a migrated managed competitor. Domain is required.
type CloudImportCompetitor struct {
	Domain string  `json:"domain"`
	Label  *string `json:"label,omitempty"`
}

// CloudImportRankingHistory is one migrated ranking-history point for a
// keyword. CheckedAt is required.
type CloudImportRankingHistory struct {
	CheckedAt        string  `json:"checkedAt"`
	Position         *int    `json:"position,omitempty"`
	PreviousPosition *int    `json:"previousPosition,omitempty"`
	RankingURL       *string `json:"rankingUrl,omitempty"`
}

// CloudImportKeyword is a migrated keyword with optional ranking history.
// Supply either Keyword or Text.
type CloudImportKeyword struct {
	Country        string                      `json:"country,omitempty"`
	Device         string                      `json:"device,omitempty"`
	ID             string                      `json:"id,omitempty"`
	Keyword        string                      `json:"keyword,omitempty"`
	Location       string                      `json:"location,omitempty"`
	RankingHistory []CloudImportRankingHistory `json:"rankingHistory,omitempty"`
	Tags           []string                    `json:"tags,omitempty"`
	TargetURL      *string                     `json:"target_url,omitempty"`
	Text           string                      `json:"text,omitempty"`
}

// CloudImportNotificationPreference is a migrated per-user notification
// preference set. All fields are optional booleans.
type CloudImportNotificationPreference struct {
	AlertEmail  *bool `json:"alert_email,omitempty"`
	AlertInApp  *bool `json:"alert_in_app,omitempty"`
	CheckEmail  *bool `json:"check_email,omitempty"`
	CheckInApp  *bool `json:"check_in_app,omitempty"`
	ImportEmail *bool `json:"import_email,omitempty"`
	ImportInApp *bool `json:"import_in_app,omitempty"`
	InviteEmail *bool `json:"invite_email,omitempty"`
	InviteInApp *bool `json:"invite_in_app,omitempty"`
}

// CloudImportSavedView is a migrated saved view. Name is required; Config is
// an opaque view configuration.
type CloudImportSavedView struct {
	Config  any    `json:"config,omitempty"`
	Name    string `json:"name"`
	Surface string `json:"surface,omitempty"`
}

// CloudImportTopLevelRankCheck is a migrated standalone rank check attached to
// a keyword by id or text.
type CloudImportTopLevelRankCheck struct {
	CheckedAt        string  `json:"checked_at,omitempty"`
	Keyword          string  `json:"keyword,omitempty"`
	KeywordID        string  `json:"keyword_id,omitempty"`
	Position         *int    `json:"position,omitempty"`
	PreviousPosition *int    `json:"previous_position,omitempty"`
	RankingURL       *string `json:"ranking_url,omitempty"`
	Text             string  `json:"text,omitempty"`
}

// CloudImportPackage is the full export package accepted by importCloudExport.
type CloudImportPackage struct {
	AlertRules              []CloudImportAlertRule              `json:"alert_rules,omitempty"`
	Competitors             []CloudImportCompetitor             `json:"competitors,omitempty"`
	ExportedAt              string                              `json:"exportedAt,omitempty"`
	Keywords                []CloudImportKeyword                `json:"keywords,omitempty"`
	NotificationPreferences []CloudImportNotificationPreference `json:"notification_preferences,omitempty"`
	ProjectID               string                              `json:"projectId,omitempty"`
	RankChecks              []CloudImportTopLevelRankCheck      `json:"rank_checks,omitempty"`
	SavedViews              []CloudImportSavedView              `json:"saved_views,omitempty"`
	Scope                   string                              `json:"scope,omitempty"`
	Version                 int                                 `json:"version,omitempty"`
}

// CloudImportSessionTotals declares the expected record totals for a chunked
// import session.
type CloudImportSessionTotals struct {
	Keywords   int `json:"keywords,omitempty"`
	RankChecks int `json:"rank_checks,omitempty"`
}

// CloudImportSessionCreate creates a chunked cloud import session. Version must
// be 3 and ChunkCount must be between 1 and 500.
type CloudImportSessionCreate struct {
	ChunkCount int                       `json:"chunk_count"`
	Totals     *CloudImportSessionTotals `json:"totals,omitempty"`
	Version    int                       `json:"version"`
}

// CloudImportChunkLimits reports the per-chunk size limits enforced by the
// server for a session.
type CloudImportChunkLimits struct {
	MaxBodyBytes   int `json:"max_body_bytes"`
	MaxHistoryRows int `json:"max_history_rows"`
	MaxKeywords    int `json:"max_keywords"`
}

// CloudImportSessionCreateResponse is returned by createCloudImportSession.
type CloudImportSessionCreateResponse struct {
	ChunkLimits CloudImportChunkLimits `json:"chunk_limits"`
	SessionID   string                 `json:"session_id"`
	State       string                 `json:"state"`
}

// CloudImportSourceKeyword identifies a source keyword by device, location, and
// text within an uploaded sections chunk.
type CloudImportSourceKeyword struct {
	Device   string `json:"device"`
	Location string `json:"location"`
	Text     string `json:"text"`
}

// CloudImportSessionSections carries the non-keyword sections uploaded in a
// "sections" chunk.
type CloudImportSessionSections struct {
	AlertRules              []CloudImportAlertRule              `json:"alert_rules,omitempty"`
	Competitors             []CloudImportCompetitor             `json:"competitors,omitempty"`
	NotificationPreferences []CloudImportNotificationPreference `json:"notification_preferences,omitempty"`
	SavedViews              []CloudImportSavedView              `json:"saved_views,omitempty"`
	SourceKeywordIDs        map[string]CloudImportSourceKeyword `json:"source_keyword_ids,omitempty"`
}

// CloudImportUploadChunk is one chunk uploaded to a session. Kind selects the
// payload: set Kind to "keywords" with Keywords populated, or "sections" with
// Sections populated. Checksum is a "sha256:<hex>" digest of the chunk body.
type CloudImportUploadChunk struct {
	Checksum string                      `json:"checksum"`
	Keywords []CloudImportKeyword        `json:"keywords,omitempty"`
	Kind     string                      `json:"kind"`
	Sections *CloudImportSessionSections `json:"sections,omitempty"`
}

// CloudImportChunkResponse is returned by uploadCloudImportChunk after a chunk
// is accepted.
type CloudImportChunkResponse struct {
	ChunkCount     int    `json:"chunk_count"`
	ChunksReceived int    `json:"chunks_received"`
	State          string `json:"state"`
}
