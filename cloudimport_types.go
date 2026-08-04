package bisibility

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// CloudImportProtocolVersion is the only export-package protocol version
// supported by the cloud-import API.
const CloudImportProtocolVersion = 5

// CloudImportScope selects the data included in an export package.
type CloudImportScope string

const (
	CloudImportScopeCurrent CloudImportScope = "current"
	CloudImportScopeHistory CloudImportScope = "history"
)

// CloudImportAlertCondition identifies a migrated alert-rule condition.
type CloudImportAlertCondition string

const (
	CloudImportAlertConditionChangePct          CloudImportAlertCondition = "change_pct"
	CloudImportAlertConditionCompetitorOvertake CloudImportAlertCondition = "competitor_overtake"
	CloudImportAlertConditionCTRDrop            CloudImportAlertCondition = "ctr_drop"
	CloudImportAlertConditionDowntrend          CloudImportAlertCondition = "downtrend"
	CloudImportAlertConditionEntersTopN         CloudImportAlertCondition = "enters_top_n"
	CloudImportAlertConditionExitsTopN          CloudImportAlertCondition = "exits_top_n"
	CloudImportAlertConditionPositionDrop       CloudImportAlertCondition = "position_drop"
	CloudImportAlertConditionSERPFeature        CloudImportAlertCondition = "serp_feature"
	CloudImportAlertConditionThreshold          CloudImportAlertCondition = "threshold"
	CloudImportAlertConditionURLMismatch        CloudImportAlertCondition = "url_mismatch"
)

// CloudImportSavedViewSurface identifies the saved-view surface.
type CloudImportSavedViewSurface string

const (
	CloudImportSavedViewSurfaceKeywords    CloudImportSavedViewSurface = "keywords"
	CloudImportSavedViewSurfaceCompetitors CloudImportSavedViewSurface = "competitors"
)

// CloudImportCompatibility is the unauthenticated schema-compatibility
// preflight. A valid response lists only protocol version 5.
type CloudImportCompatibility struct {
	AppVersion              string  `json:"app_version"`
	LatestMigration         *string `json:"latest_migration"`
	SchemaVersionsSupported []int   `json:"schema_versions_supported"`
}

// CloudImportCounts maps an imported resource name to the number of records
// created for it.
type CloudImportCounts map[string]int

// CloudImportFinalizeResponse is returned after a completed import. JobID is
// always a strict imp_ public ID and State is always done.
type CloudImportFinalizeResponse struct {
	Counts CloudImportCounts `json:"counts"`
	JobID  string            `json:"job_id"`
	State  CloudImportState  `json:"state"`
}

// CloudImportAlertRuleTarget is a sealed discriminated union. Use either
// CloudImportKeywordAlertTarget or CloudImportTagAlertTarget.
type CloudImportAlertRuleTarget interface {
	cloudImportAlertRuleTarget()
}

// CloudImportKeywordAlertTarget is the keyword variant of an alert-rule
// target. Its JSON representation always includes type: "keyword".
type CloudImportKeywordAlertTarget struct {
	Device    Device `json:"device,omitempty"`
	Keyword   string `json:"keyword,omitempty"`
	KeywordID string `json:"keyword_id"`
	Location  string `json:"location,omitempty"`
}

func (CloudImportKeywordAlertTarget) cloudImportAlertRuleTarget() {
	// This marker seals CloudImportAlertRuleTarget to supported SDK variants.
}

// CloudImportTagAlertTarget is the tag variant of an alert-rule target. Its
// JSON representation always includes type: "tag".
type CloudImportTagAlertTarget struct {
	Tag string `json:"tag"`
}

func (CloudImportTagAlertTarget) cloudImportAlertRuleTarget() {
	// This marker seals CloudImportAlertRuleTarget to supported SDK variants.
}

// CloudImportAlertRule is a migrated alert rule. ID and Name are required.
type CloudImportAlertRule struct {
	ChangePct         *float64                     `json:"change_pct,omitempty"`
	Channels          []AlertChannel               `json:"channels,omitempty"`
	CompetitorDomain  *string                      `json:"competitor_domain,omitempty"`
	ConditionType     CloudImportAlertCondition    `json:"condition_type,omitempty"`
	DropPositions     *int                         `json:"drop_positions,omitempty"`
	Enabled           *bool                        `json:"enabled,omitempty"`
	ID                string                       `json:"id"`
	Name              string                       `json:"name"`
	SerpFeature       *string                      `json:"serp_feature,omitempty"`
	TargetType        AlertTargetType              `json:"target_type,omitempty"`
	Targets           []CloudImportAlertRuleTarget `json:"targets,omitempty"`
	ThresholdPosition *int                         `json:"threshold_position,omitempty"`
	TopN              *int                         `json:"top_n,omitempty"`
}

// CloudImportCompetitor is a migrated managed competitor. ID and Domain are
// required.
type CloudImportCompetitor struct {
	Domain string  `json:"domain"`
	ID     string  `json:"id"`
	Label  *string `json:"label,omitempty"`
}

// CloudImportRankingHistory is one migrated ranking-history point. CheckedAt
// is required and uses the API's camelCase nested wire property.
type CloudImportRankingHistory struct {
	CheckedAt        time.Time `json:"checkedAt"`
	Position         *int      `json:"position,omitempty"`
	PreviousPosition *int      `json:"previousPosition,omitempty"`
	RankingURL       *string   `json:"rankingUrl,omitempty"`
}

// CloudImportKeyword is a migrated keyword. ID, Keyword, Device, and Location
// are required. rankingHistory deliberately remains camelCase because that is
// the v5 OpenAPI wire contract.
type CloudImportKeyword struct {
	Device         Device                      `json:"device"`
	ID             string                      `json:"id"`
	Keyword        string                      `json:"keyword"`
	Location       string                      `json:"location"`
	RankingHistory []CloudImportRankingHistory `json:"rankingHistory,omitempty"`
	Tags           []string                    `json:"tags,omitempty"`
	TargetURL      *string                     `json:"target_url,omitempty"`
}

// CloudImportNotificationPreference is one migrated per-user notification
// preference set.
type CloudImportNotificationPreference struct {
	AlertEmail  *bool `json:"alert_email,omitempty"`
	AlertInApp  *bool `json:"alert_in_app,omitempty"`
	CheckEmail  *bool `json:"check_email,omitempty"`
	CheckInApp  *bool `json:"check_in_app,omitempty"`
	ImportEmail *bool `json:"import_email,omitempty"`
	ImportInApp *bool `json:"import_in_app,omitempty"`
	InviteEmail *bool `json:"invite_email,omitempty"`
	InviteInApp *bool `json:"invite_in_app,omitempty"`
	ReportEmail *bool `json:"report_email,omitempty"`
}

// CloudImportSavedView is a migrated saved view. Config is intentionally an
// unconstrained JSON value, matching the OpenAPI schema.
type CloudImportSavedView struct {
	Config  any                         `json:"config,omitempty"`
	ID      string                      `json:"id"`
	Name    string                      `json:"name"`
	Surface CloudImportSavedViewSurface `json:"surface,omitempty"`
}

// CloudImportPackage is the complete v5 export accepted by ImportCloudExport.
// The SDK writes version 5 itself. Every collection below is required and must
// be represented by a non-nil slice, including when it is empty.
type CloudImportPackage struct {
	AlertRules              []CloudImportAlertRule              `json:"alert_rules"`
	Competitors             []CloudImportCompetitor             `json:"competitors"`
	ExportedAt              *time.Time                          `json:"exported_at,omitempty"`
	Keywords                []CloudImportKeyword                `json:"keywords"`
	NotificationPreferences []CloudImportNotificationPreference `json:"notification_preferences"`
	ProjectID               string                              `json:"project_id"`
	SavedViews              []CloudImportSavedView              `json:"saved_views"`
	Scope                   CloudImportScope                    `json:"scope,omitempty"`
}

// CloudImportSessionTotals declares optional expected record totals for a
// chunked import session.
type CloudImportSessionTotals struct {
	Keywords   int `json:"keywords"`
	RankChecks int `json:"rank_checks"`
}

// CloudImportSessionCreate creates a chunked v5 cloud-import session. The SDK
// writes version 5 itself; ChunkCount and SourceProjectID are required.
type CloudImportSessionCreate struct {
	ChunkCount      int                       `json:"chunk_count"`
	SourceProjectID string                    `json:"source_project_id"`
	Totals          *CloudImportSessionTotals `json:"totals,omitempty"`
}

// CloudImportChunkLimits reports the per-chunk limits enforced by the server.
type CloudImportChunkLimits struct {
	MaxBodyBytes   int `json:"max_body_bytes"`
	MaxHistoryRows int `json:"max_history_rows"`
	MaxKeywords    int `json:"max_keywords"`
}

// CloudImportSessionCreateResponse is returned by CreateCloudImportSession.
// SessionID is an imp_ public ID despite the route retaining "sessions".
type CloudImportSessionCreateResponse struct {
	ChunkLimits CloudImportChunkLimits `json:"chunk_limits"`
	SessionID   string                 `json:"session_id"`
	State       CloudImportState       `json:"state"`
}

// CloudImportSourceKeyword identifies a source keyword in a sections chunk.
type CloudImportSourceKeyword struct {
	Device   Device `json:"device"`
	Location string `json:"location"`
	Text     string `json:"text"`
}

// CloudImportSessionSections carries the non-keyword sections in a sections
// chunk. All section properties are optional, as specified by v5.
type CloudImportSessionSections struct {
	AlertRules              []CloudImportAlertRule              `json:"alert_rules,omitempty"`
	Competitors             []CloudImportCompetitor             `json:"competitors,omitempty"`
	NotificationPreferences []CloudImportNotificationPreference `json:"notification_preferences,omitempty"`
	SavedViews              []CloudImportSavedView              `json:"saved_views,omitempty"`
	SourceKeywordIDs        map[string]CloudImportSourceKeyword `json:"source_keyword_ids,omitempty"`
}

// CloudImportUploadChunk is a sealed discriminated union. Use either
// CloudImportKeywordsChunk or CloudImportSectionsChunk.
type CloudImportUploadChunk interface {
	cloudImportUploadChunk()
}

// CloudImportKeywordsChunk uploads a required keywords array. Its JSON
// representation always includes kind: "keywords".
type CloudImportKeywordsChunk struct {
	Checksum string               `json:"checksum"`
	Keywords []CloudImportKeyword `json:"keywords"`
}

func (CloudImportKeywordsChunk) cloudImportUploadChunk() {
	// This marker seals CloudImportUploadChunk to supported SDK variants.
}

// CloudImportSectionsChunk uploads a required sections object. Its JSON
// representation always includes kind: "sections".
type CloudImportSectionsChunk struct {
	Checksum string                     `json:"checksum"`
	Sections CloudImportSessionSections `json:"sections"`
}

func (CloudImportSectionsChunk) cloudImportUploadChunk() {
	// This marker seals CloudImportUploadChunk to supported SDK variants.
}

// CloudImportChunkResponse is returned when a chunk is accepted.
type CloudImportChunkResponse struct {
	ChunkCount     int              `json:"chunk_count"`
	ChunksReceived int              `json:"chunks_received"`
	State          CloudImportState `json:"state"`
}

func (input CloudImportCompatibility) MarshalJSON() ([]byte, error) {
	if err := validateCloudImportCompatibility(input); err != nil {
		return nil, err
	}
	type wire CloudImportCompatibility
	return json.Marshal(wire(input))
}

func (input *CloudImportCompatibility) UnmarshalJSON(data []byte) error {
	type wire CloudImportCompatibility
	var decoded wire
	if err := decodeStrictCloudImportObject(data, &decoded, "CloudImportCompatibility", []string{"app_version", "latest_migration", "schema_versions_supported"}, map[string]bool{"latest_migration": true}); err != nil {
		return err
	}
	value := CloudImportCompatibility(decoded)
	if err := validateCloudImportCompatibility(value); err != nil {
		return err
	}
	*input = value
	return nil
}

func (input CloudImportFinalizeResponse) MarshalJSON() ([]byte, error) {
	if err := validateCloudImportFinalizeResponse(input); err != nil {
		return nil, err
	}
	type wire CloudImportFinalizeResponse
	return json.Marshal(wire(input))
}

func (input *CloudImportFinalizeResponse) UnmarshalJSON(data []byte) error {
	type wire CloudImportFinalizeResponse
	var decoded wire
	if err := decodeStrictCloudImportObject(data, &decoded, "CloudImportFinalizeResponse", []string{"counts", "job_id", "state"}, nil); err != nil {
		return err
	}
	value := CloudImportFinalizeResponse(decoded)
	if err := validateCloudImportFinalizeResponse(value); err != nil {
		return err
	}
	*input = value
	return nil
}

func (input CloudImportKeywordAlertTarget) MarshalJSON() ([]byte, error) {
	if err := validateCloudImportKeywordAlertTarget(input, "CloudImportKeywordAlertTarget"); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Device    Device `json:"device,omitempty"`
		Keyword   string `json:"keyword,omitempty"`
		KeywordID string `json:"keyword_id"`
		Location  string `json:"location,omitempty"`
		Type      string `json:"type"`
	}{input.Device, input.Keyword, input.KeywordID, input.Location, "keyword"})
}

func (input *CloudImportKeywordAlertTarget) UnmarshalJSON(data []byte) error {
	var decoded struct {
		Device    Device `json:"device,omitempty"`
		Keyword   string `json:"keyword,omitempty"`
		KeywordID string `json:"keyword_id"`
		Location  string `json:"location,omitempty"`
		Type      string `json:"type"`
	}
	if err := decodeStrictCloudImportObject(data, &decoded, "CloudImportKeywordAlertTarget", []string{"keyword_id", "type"}, nil); err != nil {
		return err
	}
	if decoded.Type != "keyword" {
		return fmt.Errorf("CloudImportKeywordAlertTarget.type must be keyword")
	}
	value := CloudImportKeywordAlertTarget{Device: decoded.Device, Keyword: decoded.Keyword, KeywordID: decoded.KeywordID, Location: decoded.Location}
	if err := validateCloudImportKeywordAlertTarget(value, "CloudImportKeywordAlertTarget"); err != nil {
		return err
	}
	*input = value
	return nil
}

func (input CloudImportTagAlertTarget) MarshalJSON() ([]byte, error) {
	if err := validateCloudImportTagAlertTarget(input, "CloudImportTagAlertTarget"); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Tag  string `json:"tag"`
		Type string `json:"type"`
	}{input.Tag, "tag"})
}

func (input *CloudImportTagAlertTarget) UnmarshalJSON(data []byte) error {
	var decoded struct {
		Tag  string `json:"tag"`
		Type string `json:"type"`
	}
	if err := decodeStrictCloudImportObject(data, &decoded, "CloudImportTagAlertTarget", []string{"tag", "type"}, nil); err != nil {
		return err
	}
	if decoded.Type != "tag" {
		return fmt.Errorf("CloudImportTagAlertTarget.type must be tag")
	}
	value := CloudImportTagAlertTarget{Tag: decoded.Tag}
	if err := validateCloudImportTagAlertTarget(value, "CloudImportTagAlertTarget"); err != nil {
		return err
	}
	*input = value
	return nil
}

func (input CloudImportAlertRule) MarshalJSON() ([]byte, error) {
	if err := validateCloudImportAlertRule(input, "CloudImportAlertRule"); err != nil {
		return nil, err
	}
	type wire CloudImportAlertRule
	return json.Marshal(wire(input))
}

func (input *CloudImportAlertRule) UnmarshalJSON(data []byte) error {
	var decoded struct {
		ChangePct         *float64                  `json:"change_pct,omitempty"`
		Channels          []AlertChannel            `json:"channels,omitempty"`
		CompetitorDomain  *string                   `json:"competitor_domain,omitempty"`
		ConditionType     CloudImportAlertCondition `json:"condition_type,omitempty"`
		DropPositions     *int                      `json:"drop_positions,omitempty"`
		Enabled           *bool                     `json:"enabled,omitempty"`
		ID                string                    `json:"id"`
		Name              string                    `json:"name"`
		SerpFeature       *string                   `json:"serp_feature,omitempty"`
		TargetType        AlertTargetType           `json:"target_type,omitempty"`
		Targets           []json.RawMessage         `json:"targets,omitempty"`
		ThresholdPosition *int                      `json:"threshold_position,omitempty"`
		TopN              *int                      `json:"top_n,omitempty"`
	}
	if err := decodeStrictCloudImportObject(data, &decoded, "CloudImportAlertRule", []string{"id", "name"}, map[string]bool{"change_pct": true, "competitor_domain": true, "drop_positions": true, "serp_feature": true, "threshold_position": true, "top_n": true}); err != nil {
		return err
	}
	targets := make([]CloudImportAlertRuleTarget, len(decoded.Targets))
	for index, raw := range decoded.Targets {
		target, err := decodeCloudImportAlertRuleTarget(raw)
		if err != nil {
			return fmt.Errorf("CloudImportAlertRule.targets[%d]: %w", index, err)
		}
		targets[index] = target
	}
	value := CloudImportAlertRule{
		ChangePct: decoded.ChangePct, Channels: decoded.Channels, CompetitorDomain: decoded.CompetitorDomain,
		ConditionType: decoded.ConditionType, DropPositions: decoded.DropPositions, Enabled: decoded.Enabled,
		ID: decoded.ID, Name: decoded.Name, SerpFeature: decoded.SerpFeature, TargetType: decoded.TargetType,
		Targets: targets, ThresholdPosition: decoded.ThresholdPosition, TopN: decoded.TopN,
	}
	if err := validateCloudImportAlertRule(value, "CloudImportAlertRule"); err != nil {
		return err
	}
	*input = value
	return nil
}

func (input CloudImportCompetitor) MarshalJSON() ([]byte, error) {
	if err := validateCloudImportCompetitor(input, "CloudImportCompetitor"); err != nil {
		return nil, err
	}
	type wire CloudImportCompetitor
	return json.Marshal(wire(input))
}

func (input *CloudImportCompetitor) UnmarshalJSON(data []byte) error {
	type wire CloudImportCompetitor
	var decoded wire
	if err := decodeStrictCloudImportObject(data, &decoded, "CloudImportCompetitor", []string{"id", "domain"}, map[string]bool{"label": true}); err != nil {
		return err
	}
	value := CloudImportCompetitor(decoded)
	if err := validateCloudImportCompetitor(value, "CloudImportCompetitor"); err != nil {
		return err
	}
	*input = value
	return nil
}

func (input CloudImportRankingHistory) MarshalJSON() ([]byte, error) {
	if err := validateCloudImportRankingHistory(input, "CloudImportRankingHistory"); err != nil {
		return nil, err
	}
	type wire CloudImportRankingHistory
	return json.Marshal(wire(input))
}

func (input *CloudImportRankingHistory) UnmarshalJSON(data []byte) error {
	type wire CloudImportRankingHistory
	var decoded wire
	if err := decodeStrictCloudImportObject(data, &decoded, "CloudImportRankingHistory", []string{"checkedAt"}, map[string]bool{"position": true, "previousPosition": true, "rankingUrl": true}); err != nil {
		return err
	}
	value := CloudImportRankingHistory(decoded)
	if err := validateCloudImportRankingHistory(value, "CloudImportRankingHistory"); err != nil {
		return err
	}
	*input = value
	return nil
}

func (input CloudImportKeyword) MarshalJSON() ([]byte, error) {
	if err := validateCloudImportKeyword(input, "CloudImportKeyword"); err != nil {
		return nil, err
	}
	type wire CloudImportKeyword
	return json.Marshal(wire(input))
}

func (input *CloudImportKeyword) UnmarshalJSON(data []byte) error {
	type wire CloudImportKeyword
	var decoded wire
	if err := decodeStrictCloudImportObject(data, &decoded, "CloudImportKeyword", []string{"id", "keyword", "device", "location"}, map[string]bool{"target_url": true}); err != nil {
		return err
	}
	value := CloudImportKeyword(decoded)
	if err := validateCloudImportKeyword(value, "CloudImportKeyword"); err != nil {
		return err
	}
	*input = value
	return nil
}

func (input CloudImportNotificationPreference) MarshalJSON() ([]byte, error) {
	type wire CloudImportNotificationPreference
	return json.Marshal(wire(input))
}

func (input *CloudImportNotificationPreference) UnmarshalJSON(data []byte) error {
	type wire CloudImportNotificationPreference
	var decoded wire
	if err := decodeStrictCloudImportObject(data, &decoded, "CloudImportNotificationPreference", nil, nil); err != nil {
		return err
	}
	*input = CloudImportNotificationPreference(decoded)
	return nil
}

func (input CloudImportSavedView) MarshalJSON() ([]byte, error) {
	if err := validateCloudImportSavedView(input, "CloudImportSavedView"); err != nil {
		return nil, err
	}
	type wire CloudImportSavedView
	return json.Marshal(wire(input))
}

func (input *CloudImportSavedView) UnmarshalJSON(data []byte) error {
	type wire CloudImportSavedView
	var decoded wire
	if err := decodeStrictCloudImportObject(data, &decoded, "CloudImportSavedView", []string{"id", "name"}, map[string]bool{"config": true}); err != nil {
		return err
	}
	value := CloudImportSavedView(decoded)
	if err := validateCloudImportSavedView(value, "CloudImportSavedView"); err != nil {
		return err
	}
	*input = value
	return nil
}

func (input CloudImportPackage) MarshalJSON() ([]byte, error) {
	if err := validateCloudImportPackage(input, "CloudImportPackage"); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		AlertRules              []CloudImportAlertRule              `json:"alert_rules"`
		Competitors             []CloudImportCompetitor             `json:"competitors"`
		ExportedAt              *time.Time                          `json:"exported_at,omitempty"`
		Keywords                []CloudImportKeyword                `json:"keywords"`
		NotificationPreferences []CloudImportNotificationPreference `json:"notification_preferences"`
		ProjectID               string                              `json:"project_id"`
		SavedViews              []CloudImportSavedView              `json:"saved_views"`
		Scope                   CloudImportScope                    `json:"scope,omitempty"`
		Version                 int                                 `json:"version"`
	}{input.AlertRules, input.Competitors, input.ExportedAt, input.Keywords, input.NotificationPreferences, input.ProjectID, input.SavedViews, input.Scope, CloudImportProtocolVersion})
}

func (input *CloudImportPackage) UnmarshalJSON(data []byte) error {
	var decoded struct {
		AlertRules              []CloudImportAlertRule              `json:"alert_rules"`
		Competitors             []CloudImportCompetitor             `json:"competitors"`
		ExportedAt              *time.Time                          `json:"exported_at,omitempty"`
		Keywords                []CloudImportKeyword                `json:"keywords"`
		NotificationPreferences []CloudImportNotificationPreference `json:"notification_preferences"`
		ProjectID               string                              `json:"project_id"`
		SavedViews              []CloudImportSavedView              `json:"saved_views"`
		Scope                   CloudImportScope                    `json:"scope,omitempty"`
		Version                 int                                 `json:"version"`
	}
	if err := decodeStrictCloudImportObject(data, &decoded, "CloudImportPackage", []string{"version", "project_id", "keywords", "alert_rules", "competitors", "notification_preferences", "saved_views"}, nil); err != nil {
		return err
	}
	if decoded.Version != CloudImportProtocolVersion {
		return fmt.Errorf("CloudImportPackage.version must be %d", CloudImportProtocolVersion)
	}
	value := CloudImportPackage{
		AlertRules: decoded.AlertRules, Competitors: decoded.Competitors, ExportedAt: decoded.ExportedAt,
		Keywords: decoded.Keywords, NotificationPreferences: decoded.NotificationPreferences,
		ProjectID: decoded.ProjectID, SavedViews: decoded.SavedViews, Scope: decoded.Scope,
	}
	if err := validateCloudImportPackage(value, "CloudImportPackage"); err != nil {
		return err
	}
	*input = value
	return nil
}

func (input CloudImportSessionTotals) MarshalJSON() ([]byte, error) {
	if err := validateCloudImportSessionTotals(input, "CloudImportSessionTotals"); err != nil {
		return nil, err
	}
	type wire CloudImportSessionTotals
	return json.Marshal(wire(input))
}

func (input *CloudImportSessionTotals) UnmarshalJSON(data []byte) error {
	type wire CloudImportSessionTotals
	var decoded wire
	if err := decodeStrictCloudImportObject(data, &decoded, "CloudImportSessionTotals", nil, nil); err != nil {
		return err
	}
	value := CloudImportSessionTotals(decoded)
	if err := validateCloudImportSessionTotals(value, "CloudImportSessionTotals"); err != nil {
		return err
	}
	*input = value
	return nil
}

func (input CloudImportSessionCreate) MarshalJSON() ([]byte, error) {
	if err := validateCloudImportSessionCreate(input, "CloudImportSessionCreate"); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		ChunkCount      int                       `json:"chunk_count"`
		SourceProjectID string                    `json:"source_project_id"`
		Totals          *CloudImportSessionTotals `json:"totals,omitempty"`
		Version         int                       `json:"version"`
	}{input.ChunkCount, input.SourceProjectID, input.Totals, CloudImportProtocolVersion})
}

func (input *CloudImportSessionCreate) UnmarshalJSON(data []byte) error {
	var decoded struct {
		ChunkCount      int                       `json:"chunk_count"`
		SourceProjectID string                    `json:"source_project_id"`
		Totals          *CloudImportSessionTotals `json:"totals,omitempty"`
		Version         int                       `json:"version"`
	}
	if err := decodeStrictCloudImportObject(data, &decoded, "CloudImportSessionCreate", []string{"version", "chunk_count", "source_project_id"}, nil); err != nil {
		return err
	}
	if decoded.Version != CloudImportProtocolVersion {
		return fmt.Errorf("CloudImportSessionCreate.version must be %d", CloudImportProtocolVersion)
	}
	value := CloudImportSessionCreate{ChunkCount: decoded.ChunkCount, SourceProjectID: decoded.SourceProjectID, Totals: decoded.Totals}
	if err := validateCloudImportSessionCreate(value, "CloudImportSessionCreate"); err != nil {
		return err
	}
	*input = value
	return nil
}

func (input CloudImportChunkLimits) MarshalJSON() ([]byte, error) {
	if err := validateCloudImportChunkLimits(input, "CloudImportChunkLimits"); err != nil {
		return nil, err
	}
	type wire CloudImportChunkLimits
	return json.Marshal(wire(input))
}

func (input *CloudImportChunkLimits) UnmarshalJSON(data []byte) error {
	type wire CloudImportChunkLimits
	var decoded wire
	if err := decodeStrictCloudImportObject(data, &decoded, "CloudImportChunkLimits", []string{"max_body_bytes", "max_history_rows", "max_keywords"}, nil); err != nil {
		return err
	}
	value := CloudImportChunkLimits(decoded)
	if err := validateCloudImportChunkLimits(value, "CloudImportChunkLimits"); err != nil {
		return err
	}
	*input = value
	return nil
}

func (input CloudImportSessionCreateResponse) MarshalJSON() ([]byte, error) {
	if err := validateCloudImportSessionCreateResponse(input); err != nil {
		return nil, err
	}
	type wire CloudImportSessionCreateResponse
	return json.Marshal(wire(input))
}

func (input *CloudImportSessionCreateResponse) UnmarshalJSON(data []byte) error {
	type wire CloudImportSessionCreateResponse
	var decoded wire
	if err := decodeStrictCloudImportObject(data, &decoded, "CloudImportSessionCreateResponse", []string{"session_id", "state", "chunk_limits"}, nil); err != nil {
		return err
	}
	value := CloudImportSessionCreateResponse(decoded)
	if err := validateCloudImportSessionCreateResponse(value); err != nil {
		return err
	}
	*input = value
	return nil
}

func (input CloudImportSourceKeyword) MarshalJSON() ([]byte, error) {
	if err := validateCloudImportSourceKeyword(input, "CloudImportSourceKeyword"); err != nil {
		return nil, err
	}
	type wire CloudImportSourceKeyword
	return json.Marshal(wire(input))
}

func (input *CloudImportSourceKeyword) UnmarshalJSON(data []byte) error {
	type wire CloudImportSourceKeyword
	var decoded wire
	if err := decodeStrictCloudImportObject(data, &decoded, "CloudImportSourceKeyword", []string{"device", "location", "text"}, nil); err != nil {
		return err
	}
	value := CloudImportSourceKeyword(decoded)
	if err := validateCloudImportSourceKeyword(value, "CloudImportSourceKeyword"); err != nil {
		return err
	}
	*input = value
	return nil
}

func (input CloudImportSessionSections) MarshalJSON() ([]byte, error) {
	if err := validateCloudImportSessionSections(input, "CloudImportSessionSections"); err != nil {
		return nil, err
	}
	type wire CloudImportSessionSections
	return json.Marshal(wire(input))
}

func (input *CloudImportSessionSections) UnmarshalJSON(data []byte) error {
	type wire CloudImportSessionSections
	var decoded wire
	if err := decodeStrictCloudImportObject(data, &decoded, "CloudImportSessionSections", nil, nil); err != nil {
		return err
	}
	value := CloudImportSessionSections(decoded)
	if err := validateCloudImportSessionSections(value, "CloudImportSessionSections"); err != nil {
		return err
	}
	*input = value
	return nil
}

func (input CloudImportKeywordsChunk) MarshalJSON() ([]byte, error) {
	if err := validateCloudImportKeywordsChunk(input, "CloudImportKeywordsChunk"); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Checksum string               `json:"checksum"`
		Kind     string               `json:"kind"`
		Keywords []CloudImportKeyword `json:"keywords"`
	}{input.Checksum, "keywords", input.Keywords})
}

func (input *CloudImportKeywordsChunk) UnmarshalJSON(data []byte) error {
	var decoded struct {
		Checksum string               `json:"checksum"`
		Kind     string               `json:"kind"`
		Keywords []CloudImportKeyword `json:"keywords"`
	}
	if err := decodeStrictCloudImportObject(data, &decoded, "CloudImportKeywordsChunk", []string{"checksum", "kind", "keywords"}, nil); err != nil {
		return err
	}
	if decoded.Kind != "keywords" {
		return fmt.Errorf("CloudImportKeywordsChunk.kind must be keywords")
	}
	value := CloudImportKeywordsChunk{Checksum: decoded.Checksum, Keywords: decoded.Keywords}
	if err := validateCloudImportKeywordsChunk(value, "CloudImportKeywordsChunk"); err != nil {
		return err
	}
	*input = value
	return nil
}

func (input CloudImportSectionsChunk) MarshalJSON() ([]byte, error) {
	if err := validateCloudImportSectionsChunk(input, "CloudImportSectionsChunk"); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Checksum string                     `json:"checksum"`
		Kind     string                     `json:"kind"`
		Sections CloudImportSessionSections `json:"sections"`
	}{input.Checksum, "sections", input.Sections})
}

func (input *CloudImportSectionsChunk) UnmarshalJSON(data []byte) error {
	var decoded struct {
		Checksum string                     `json:"checksum"`
		Kind     string                     `json:"kind"`
		Sections CloudImportSessionSections `json:"sections"`
	}
	if err := decodeStrictCloudImportObject(data, &decoded, "CloudImportSectionsChunk", []string{"checksum", "kind", "sections"}, nil); err != nil {
		return err
	}
	if decoded.Kind != "sections" {
		return fmt.Errorf("CloudImportSectionsChunk.kind must be sections")
	}
	value := CloudImportSectionsChunk{Checksum: decoded.Checksum, Sections: decoded.Sections}
	if err := validateCloudImportSectionsChunk(value, "CloudImportSectionsChunk"); err != nil {
		return err
	}
	*input = value
	return nil
}

func (input CloudImportChunkResponse) MarshalJSON() ([]byte, error) {
	if err := validateCloudImportChunkResponse(input); err != nil {
		return nil, err
	}
	type wire CloudImportChunkResponse
	return json.Marshal(wire(input))
}

func (input *CloudImportChunkResponse) UnmarshalJSON(data []byte) error {
	type wire CloudImportChunkResponse
	var decoded wire
	if err := decodeStrictCloudImportObject(data, &decoded, "CloudImportChunkResponse", []string{"state", "chunks_received", "chunk_count"}, nil); err != nil {
		return err
	}
	value := CloudImportChunkResponse(decoded)
	if err := validateCloudImportChunkResponse(value); err != nil {
		return err
	}
	*input = value
	return nil
}

func decodeCloudImportAlertRuleTarget(data []byte) (CloudImportAlertRuleTarget, error) {
	var discriminator struct {
		Type string `json:"type"`
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, fmt.Errorf("CloudImportAlertRuleTarget must be a JSON object: %w", err)
	}
	if fields == nil || fields["type"] == nil || bytes.Equal(bytes.TrimSpace(fields["type"]), []byte("null")) {
		return nil, fmt.Errorf("CloudImportAlertRuleTarget.type is required")
	}
	if err := json.Unmarshal(data, &discriminator); err != nil {
		return nil, err
	}
	switch discriminator.Type {
	case "keyword":
		var target CloudImportKeywordAlertTarget
		if err := json.Unmarshal(data, &target); err != nil {
			return nil, err
		}
		return target, nil
	case "tag":
		var target CloudImportTagAlertTarget
		if err := json.Unmarshal(data, &target); err != nil {
			return nil, err
		}
		return target, nil
	default:
		return nil, fmt.Errorf("CloudImportAlertRuleTarget.type must be keyword or tag")
	}
}

func decodeStrictCloudImportObject(data []byte, destination any, name string, required []string, nullable map[string]bool) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return fmt.Errorf("%s must be a JSON object: %w", name, err)
	}
	if fields == nil {
		return fmt.Errorf("%s must be a JSON object", name)
	}
	for _, field := range required {
		raw, ok := fields[field]
		if !ok || (bytes.Equal(bytes.TrimSpace(raw), []byte("null")) && !nullable[field]) {
			return fmt.Errorf("%s.%s is required", name, field)
		}
	}
	for field, raw := range fields {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) && !nullable[field] {
			return fmt.Errorf("%s.%s cannot be null", name, field)
		}
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("invalid %s: %w", name, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("invalid %s: multiple JSON values", name)
		}
		return fmt.Errorf("invalid %s: %w", name, err)
	}
	return nil
}

func isCloudImportScope(value CloudImportScope) bool {
	return value == CloudImportScopeCurrent || value == CloudImportScopeHistory
}

func isCloudImportAlertCondition(value CloudImportAlertCondition) bool {
	switch value {
	case CloudImportAlertConditionChangePct, CloudImportAlertConditionCompetitorOvertake,
		CloudImportAlertConditionCTRDrop, CloudImportAlertConditionDowntrend,
		CloudImportAlertConditionEntersTopN, CloudImportAlertConditionExitsTopN,
		CloudImportAlertConditionPositionDrop, CloudImportAlertConditionSERPFeature,
		CloudImportAlertConditionThreshold, CloudImportAlertConditionURLMismatch:
		return true
	default:
		return false
	}
}

func isCloudImportSavedViewSurface(value CloudImportSavedViewSurface) bool {
	return value == CloudImportSavedViewSurfaceKeywords || value == CloudImportSavedViewSurfaceCompetitors
}

func isCloudImportDevice(value Device) bool {
	return value == DeviceDesktop || value == DeviceMobile
}

func isCloudImportLocation(value string) bool {
	_, ok := cloudImportLocations[value]
	return ok
}

var cloudImportLocations = map[string]struct{}{
	"United States": {}, "United Kingdom": {}, "Canada": {}, "Australia": {}, "Germany": {},
	"France": {}, "Spain": {}, "Italy": {}, "Netherlands": {}, "Sweden": {}, "Poland": {},
	"Ireland": {}, "Portugal": {}, "Belgium": {}, "Switzerland": {}, "Austria": {},
	"Denmark": {}, "Norway": {}, "Finland": {}, "Brazil": {}, "Mexico": {}, "India": {},
	"Japan": {}, "Singapore": {}, "New Zealand": {}, "South Africa": {}, "United Arab Emirates": {},
}

func hasCloudImportText(value string, maximum int) bool {
	return strings.TrimSpace(value) != "" && len(value) <= maximum
}
