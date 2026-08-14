package bisibility

import (
	"encoding/json"
	"time"
)

// Device identifies the search device used for keyword rank tracking.
type Device string

const (
	DeviceDesktop Device = "desktop"
	DeviceMobile  Device = "mobile"
)

// RankCheckFrequency controls how often Bisibility checks a keyword.
type RankCheckFrequency string

const (
	RankCheckFrequencyPaused     RankCheckFrequency = "paused"
	RankCheckFrequencyManual     RankCheckFrequency = "manual"
	RankCheckFrequencyDaily      RankCheckFrequency = "daily"
	RankCheckFrequencyWeekly     RankCheckFrequency = "weekly"
	RankCheckFrequencyMonthly    RankCheckFrequency = "monthly"
	RankCheckFrequencyCustomCron RankCheckFrequency = "custom_cron"
)

// RankCheckStatus is the status filter for rank check history.
type RankCheckStatus string

const (
	RankCheckStatusCompleted RankCheckStatus = "completed"
	RankCheckStatusFailed    RankCheckStatus = "failed"
	RankCheckStatusRunning   RankCheckStatus = "running"
)

// KeywordBulkOperation identifies a bulk keyword mutation.
type KeywordBulkOperation string

const (
	KeywordBulkOperationAddTags      KeywordBulkOperation = "add_tags"
	KeywordBulkOperationDelete       KeywordBulkOperation = "delete"
	KeywordBulkOperationRemoveTags   KeywordBulkOperation = "remove_tags"
	KeywordBulkOperationSetFrequency KeywordBulkOperation = "set_frequency"
	KeywordBulkOperationSetTargetURL KeywordBulkOperation = "set_target_url"
)

// JSONValue is used for schema-like response fields.
type JSONValue map[string]any

// ProblemDetails is the Bisibility RFC problem details error body.
type ProblemDetails struct {
	Type     string          `json:"type"`
	Title    string          `json:"title"`
	Status   int             `json:"status"`
	Detail   string          `json:"detail"`
	Instance string          `json:"instance"`
	DocsURL  string          `json:"docs_url"`
	Errors   json.RawMessage `json:"errors,omitempty"`
	// Extensions preserves RFC 9457 extension members not known to this SDK.
	Extensions map[string]json.RawMessage `json:"-"`
}

// UnmarshalJSON tolerates mistyped known members while preserving extension members.
func (p *ProblemDetails) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	decodeString := func(name string, target *string) { _ = json.Unmarshal(fields[name], target) }
	decodeString("type", &p.Type)
	decodeString("title", &p.Title)
	decodeString("detail", &p.Detail)
	decodeString("instance", &p.Instance)
	decodeString("docs_url", &p.DocsURL)
	_ = json.Unmarshal(fields["status"], &p.Status)
	if raw, ok := fields["errors"]; ok {
		p.Errors = append(p.Errors[:0], raw...)
	}
	p.Extensions = make(map[string]json.RawMessage)
	for name, raw := range fields {
		switch name {
		case "type", "title", "status", "detail", "instance", "docs_url", "errors":
			continue
		default:
			p.Extensions[name] = append(json.RawMessage(nil), raw...)
		}
	}
	return nil
}

// ListMeta contains pagination metadata.
type ListMeta struct {
	NextCursor *string `json:"next_cursor"`
}

// ListResponse is returned by paginated list endpoints.
type ListResponse[T any] struct {
	Data []T      `json:"data"`
	Meta ListMeta `json:"meta"`
}

// DataResponse is returned by endpoints that wrap data in a response envelope.
type DataResponse[T any] struct {
	Data T         `json:"data"`
	Meta JSONValue `json:"meta,omitempty"`
}

// PaginationOptions configures cursor pagination.
type PaginationOptions struct {
	Cursor string
	Limit  int
}

// ProjectWriteMode reports whether a project accepts writes.
type ProjectWriteMode string

const (
	ProjectWriteModeActive        ProjectWriteMode = "active"
	ProjectWriteModeMigrationHold ProjectWriteMode = "migration_hold"
	ProjectWriteModeMigrated      ProjectWriteMode = "migrated"
)

// Project is a Bisibility project visible to an API key.
type Project struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Domain    string           `json:"domain"`
	WriteMode ProjectWriteMode `json:"write_mode"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// UpdateProjectInput patches project metadata. At least one field is required.
type UpdateProjectInput struct {
	Domain *string `json:"domain,omitempty"`
	Name   *string `json:"name,omitempty"`
}

// ProjectDefaultsSource identifies how the effective default market was selected.
type ProjectDefaultsSource string

const (
	ProjectDefaultsSourceDerived  ProjectDefaultsSource = "derived"
	ProjectDefaultsSourceExplicit ProjectDefaultsSource = "explicit"
	ProjectDefaultsSourceFallback ProjectDefaultsSource = "fallback"
)

// ProjectDefaults are the project default market and schedule settings.
type ProjectDefaults struct {
	City            *string               `json:"city"`
	Country         string                `json:"country"`
	CronExpression  *string               `json:"cron_expression"`
	Device          Device                `json:"device"`
	Frequency       RankCheckFrequency    `json:"frequency"`
	JitterMinutes   int                   `json:"jitter_minutes"`
	LastCheckedAt   *time.Time            `json:"last_checked_at"`
	LocationKey     string                `json:"location_key"`
	NextCheckAt     *time.Time            `json:"next_check_at"`
	ProjectID       string                `json:"project_id"`
	SerpDepth       int                   `json:"serp_depth"`
	SerpStopOnMatch bool                  `json:"serp_stop_on_match"`
	Source          ProjectDefaultsSource `json:"source"`
	Timezone        string                `json:"timezone"`
	UpdatedAt       *time.Time            `json:"updated_at"`
}

// ProjectDefaultsPatch updates project default market and schedule settings.
// Frequency is required by the API. Country and Device must be provided
// together when LocationKey is omitted. Omitted schedule fields fall back to
// server defaults (jitter_minutes 60 and timezone UTC).
type ProjectDefaultsPatch struct {
	City            *string            `json:"city,omitempty"`
	Country         string             `json:"country,omitempty"`
	CronExpression  *string            `json:"cron_expression,omitempty"`
	Device          Device             `json:"device,omitempty"`
	Frequency       RankCheckFrequency `json:"frequency"`
	JitterMinutes   *int               `json:"jitter_minutes,omitempty"`
	LocationKey     string             `json:"location_key,omitempty"`
	SerpStopOnMatch *bool              `json:"serp_stop_on_match,omitempty"`
	Timezone        string             `json:"timezone,omitempty"`
}

// ProjectOverviewRange identifies the rank-history window used for overview comparisons.
type ProjectOverviewRange string

const (
	ProjectOverviewRange7Days  ProjectOverviewRange = "7d"
	ProjectOverviewRange28Days ProjectOverviewRange = "28d"
	ProjectOverviewRange90Days ProjectOverviewRange = "90d"
)

// ProjectOverviewDevice identifies the SERP device filter used for an overview.
type ProjectOverviewDevice string

const (
	ProjectOverviewDeviceAll     ProjectOverviewDevice = "all"
	ProjectOverviewDeviceDesktop ProjectOverviewDevice = "desktop"
	ProjectOverviewDeviceMobile  ProjectOverviewDevice = "mobile"
)

// ProjectOverviewOptions filters a project overview.
type ProjectOverviewOptions struct {
	Device ProjectOverviewDevice
	Range  ProjectOverviewRange
	Tag    string
}

// ProjectOverviewPositionBucket counts keywords within an inclusive position range.
type ProjectOverviewPositionBucket struct {
	Count *int `json:"count"`
	Max   int  `json:"max"`
	Min   int  `json:"min"`
}

// ProjectOverview summarizes tracked keyword rank performance for a project.
type ProjectOverview struct {
	AveragePosition        *float64                        `json:"average_position"`
	AveragePositionDelta   *float64                        `json:"average_position_delta"`
	KeywordsAddedThisMonth int                             `json:"keywords_added_this_month"`
	LastCheckAt            *time.Time                      `json:"last_check_at"`
	NextCheckAt            *time.Time                      `json:"next_check_at"`
	PositionDistribution   []ProjectOverviewPositionBucket `json:"position_distribution"`
	ProjectID              string                          `json:"project_id"`
	Top10Count             *int                            `json:"top_10_count"`
	Top10Delta             *int                            `json:"top_10_delta"`
	Top100Count            *int                            `json:"top_100_count"`
	Top3Count              *int                            `json:"top_3_count"`
	TrackedKeywordCount    int                             `json:"tracked_keyword_count"`
	Visibility             *float64                        `json:"visibility"`
	VisibilityDelta        *float64                        `json:"visibility_delta"`
}

// APIKey describes an API key without the raw token.
type APIKey struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
	RevokedAt  *time.Time `json:"revoked_at"`
}

// CreatedAPIKey is returned once when creating an API key.
type CreatedAPIKey struct {
	APIKey
	MaskedValue string `json:"masked_value"`
	Token       string `json:"token"`
}

// KeywordSchedule is a keyword schedule returned by the API.
type KeywordSchedule struct {
	CronExpression *string            `json:"cron_expression"`
	Frequency      RankCheckFrequency `json:"frequency"`
	JitterMinutes  int                `json:"jitter_minutes"`
	LastCheckedAt  *time.Time         `json:"last_checked_at"`
	NextCheckAt    *time.Time         `json:"next_check_at"`
	Timezone       string             `json:"timezone"`
}

// Keyword is a tracked keyword and its latest rank summary.
type Keyword struct {
	ID               string           `json:"id"`
	ProjectID        string           `json:"project_id"`
	Text             string           `json:"text"`
	Country          string           `json:"country"`
	LanguageCode     string           `json:"language_code"`
	LanguageLabel    string           `json:"language_label"`
	Location         string           `json:"location"`
	LocationKey      string           `json:"location_key"`
	Device           Device           `json:"device"`
	Intent           *string          `json:"intent"`
	Topic            *string          `json:"topic"`
	TargetURL        *string          `json:"target_url"`
	RankingURL       *string          `json:"ranking_url"`
	LatestPosition   *int             `json:"latest_position"`
	PreviousPosition *int             `json:"previous_position"`
	Schedule         *KeywordSchedule `json:"schedule"`
	Tags             []string         `json:"tags"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// KeywordScheduleInput is the camelCase schedule shape accepted by write endpoints.
type KeywordScheduleInput struct {
	CronExpression *string            `json:"cronExpression"`
	Frequency      RankCheckFrequency `json:"frequency"`
	JitterMinutes  *int               `json:"jitterMinutes,omitempty"`
	Timezone       string             `json:"timezone,omitempty"`
}

// CreateKeywordInput is one keyword item accepted by CreateKeywords and AddKeywords.
// LocationKey accepts canonical country, region, or city keys, optionally qualified with @language.
type CreateKeywordInput struct {
	Keyword     string                `json:"keyword"`
	City        string                `json:"city,omitempty"`
	Country     string                `json:"country,omitempty"`
	Location    string                `json:"location,omitempty"`
	LocationKey string                `json:"location_key,omitempty"`
	Device      Device                `json:"device,omitempty"`
	Intent      *string               `json:"intent,omitempty"`
	Topic       *string               `json:"topic,omitempty"`
	Schedule    *KeywordScheduleInput `json:"schedule,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
	TargetURL   *string               `json:"target_url,omitempty"`
}

// CreateKeywordsInput wraps one or more keyword creation items.
type CreateKeywordsInput struct {
	Keywords []CreateKeywordInput `json:"keywords"`
}

// CreateKeywordResult describes one create result.
type CreateKeywordResult struct {
	Keyword Keyword `json:"keyword"`
	Status  string  `json:"status"`
	Warning string  `json:"warning,omitempty"`
}

// CreateKeywordsResponse summarizes a keyword creation request.
type CreateKeywordsResponse struct {
	Created  int                   `json:"created"`
	Skipped  int                   `json:"skipped"`
	Results  []CreateKeywordResult `json:"results"`
	Warnings []string              `json:"warnings,omitempty"`
}

// KeywordMatchRequest identifies up to 50 keyword texts to match within a project.
type KeywordMatchRequest struct {
	Texts []string `json:"texts"`
}

// KeywordMatchMarket identifies one market where a keyword is tracked.
type KeywordMatchMarket struct {
	CountryCode   string `json:"country_code"`
	Device        Device `json:"device"`
	LanguageCode  string `json:"language_code"`
	LanguageLabel string `json:"language_label"`
	Location      string `json:"location"`
	LocationKey   string `json:"location_key"`
}

// KeywordMatch keeps the normalized request text separate from the stored keyword text.
type KeywordMatch struct {
	KeywordID      string `json:"keyword_id"`
	LatestPosition *int   `json:"latest_position"`
	// PreviousPosition is the previous observed rank position, if available.
	PreviousPosition *int `json:"previous_position"`
	// RankingURL is the URL that ranked at `latest_position` in the last completed check,
	// or null when the keyword has no completed check.
	RankingURL *string            `json:"ranking_url"`
	Market     KeywordMatchMarket `json:"market"`
	// MatchedText is the trimmed, lowercased request text used to match this keyword.
	MatchedText string `json:"matched_text"`
	// Text is the stored keyword text, which can differ from MatchedText in case and whitespace.
	Text string `json:"text"`
}

// KeywordMatchMeta reports normalized texts with more than 100 matching markets;
// their returned rows are partial.
type KeywordMatchMeta struct {
	TruncatedTexts []string `json:"truncated_texts"`
}

// KeywordMatchResponse contains matching keywords and truncation metadata.
type KeywordMatchResponse struct {
	Data []KeywordMatch   `json:"data"`
	Meta KeywordMatchMeta `json:"meta"`
}

// NullableString represents an optional string field that can be explicitly set to null.
type NullableString struct {
	set   bool
	value *string
}

// StringValue returns a NullableString set to a concrete string value.
func StringValue(value string) NullableString {
	return NullableString{set: true, value: &value}
}

// NullString returns a NullableString explicitly set to JSON null.
func NullString() NullableString {
	return NullableString{set: true}
}

// IsSet reports whether the field should be included in a request body.
func (n NullableString) IsSet() bool {
	return n.set
}

// Value returns the underlying string pointer. A nil pointer means JSON null when IsSet is true.
func (n NullableString) Value() *string {
	return n.value
}

// MarshalJSON implements json.Marshaler.
func (n NullableString) MarshalJSON() ([]byte, error) {
	if n.value == nil {
		return []byte("null"), nil
	}

	return json.Marshal(*n.value)
}

// UpdateKeywordInput updates keyword metadata. Use StringValue or NullString
// for TargetURL, Intent, and Topic; passing NullString clears the field.
type UpdateKeywordInput struct {
	City        *string
	Country     *string
	Device      *Device
	Frequency   *RankCheckFrequency
	Intent      NullableString
	Keyword     *string
	Location    *string
	LocationKey *string
	Schedule    *KeywordScheduleInput
	Tags        []string
	TargetURL   NullableString
	Topic       NullableString
}

// MarshalJSON includes only explicitly set update fields.
func (in UpdateKeywordInput) MarshalJSON() ([]byte, error) {
	body := make(map[string]any)
	if in.City != nil {
		body["city"] = *in.City
	}
	if in.Country != nil {
		body["country"] = *in.Country
	}
	if in.Device != nil {
		body["device"] = *in.Device
	}
	if in.Frequency != nil {
		body["frequency"] = *in.Frequency
	}
	if in.Intent.IsSet() {
		body["intent"] = in.Intent.Value()
	}
	if in.Keyword != nil {
		body["keyword"] = *in.Keyword
	}
	if in.Location != nil {
		body["location"] = *in.Location
	}
	if in.LocationKey != nil {
		body["location_key"] = *in.LocationKey
	}
	if in.Schedule != nil {
		body["schedule"] = in.Schedule
	}
	if in.Tags != nil {
		body["tags"] = in.Tags
	}
	if in.TargetURL.IsSet() {
		body["target_url"] = in.TargetURL.Value()
	}
	if in.Topic.IsSet() {
		body["topic"] = in.Topic.Value()
	}

	return json.Marshal(body)
}

// KeywordBulkInput mutates many keywords. Use StringValue or NullString for TargetURL.
type KeywordBulkInput struct {
	KeywordIDs []string
	Operation  KeywordBulkOperation
	Tags       []string
	Frequency  *RankCheckFrequency
	Schedule   *KeywordScheduleInput
	TargetURL  NullableString
}

// MarshalJSON includes only fields relevant to the selected bulk operation.
func (in KeywordBulkInput) MarshalJSON() ([]byte, error) {
	body := map[string]any{
		"keyword_ids": in.KeywordIDs,
		"operation":   in.Operation,
	}
	if in.Tags != nil {
		body["tags"] = in.Tags
	}
	if in.Frequency != nil {
		body["frequency"] = *in.Frequency
	}
	if in.Schedule != nil {
		body["schedule"] = in.Schedule
	}
	if in.TargetURL.IsSet() {
		body["target_url"] = in.TargetURL.Value()
	}

	return json.Marshal(body)
}

// KeywordBulkItemResult is one bulk mutation result.
type KeywordBulkItemResult struct {
	KeywordID string `json:"keyword_id"`
	Status    string `json:"status"`
}

// KeywordBulkResponse summarizes a bulk keyword mutation.
type KeywordBulkResponse struct {
	Operation KeywordBulkOperation    `json:"operation"`
	Results   []KeywordBulkItemResult `json:"results"`
}

// ListKeywordsOptions filters project keywords. Intent and Topic are
// case-insensitive exact filters up to 80 characters.
type ListKeywordsOptions struct {
	Cursor     string
	Limit      int
	Country    string
	Device     Device
	Intent     string
	PositionGT int
	PositionLT int
	Search     string
	Sort       string
	Tag        string
	Topic      string
}

// RankCheckAttempt is one provider fallback attempt recorded before the final
// rank-check status.
type RankCheckAttempt struct {
	Message  string `json:"message"`
	Provider string `json:"provider"`
}

// RankCheck is a keyword rank check. Status is running for async checks that
// have not completed yet.
type RankCheck struct {
	ID               string             `json:"id"`
	KeywordID        string             `json:"keyword_id"`
	Attempts         []RankCheckAttempt `json:"attempts"`
	CheckedAt        time.Time          `json:"checked_at"`
	CostCents        *float64           `json:"cost_cents"`
	Error            *string            `json:"error"`
	Position         *int               `json:"position"`
	PreviousPosition *int               `json:"previous_position"`
	Provider         string             `json:"provider"`
	RankingURL       *string            `json:"ranking_url"`
	Status           string             `json:"status"`
}

// ListRankChecksOptions filters rank check history.
type ListRankChecksOptions struct {
	Cursor string
	Limit  int
	Since  time.Time
	Status RankCheckStatus
	Until  time.Time
}

// RunRankCheckInput selects an optional provider for an immediate rank check.
// Set Async to enqueue the check instead of waiting for the result: the API
// responds 202 with a RankCheck in status running. Poll GetRankCheckResult
// until the status becomes completed or failed.
type RunRankCheckInput struct {
	ProviderID string `json:"provider_id,omitempty"`
	Async      bool   `json:"-"`
}

// HealthResponse is returned by GetHealth.
type HealthResponse struct {
	Status string `json:"status"`
}

// LivenessResponse is returned by GetLiveness.
type LivenessResponse struct {
	Status string `json:"status"`
}

// ReadinessResponse is returned by GetReadiness.
type ReadinessResponse struct {
	Status string `json:"status"`
}

// Capability describes an API capability advertised by Bisibility.
type Capability struct {
	Name        string    `json:"name"`
	OperationID string    `json:"operationId"`
	Description string    `json:"description"`
	InputSchema JSONValue `json:"input_schema"`
}

// OpenAPIDocument is the OpenAPI document returned by GetOpenAPI.
type OpenAPIDocument struct {
	OpenAPI    string         `json:"openapi"`
	Info       JSONValue      `json:"info"`
	Paths      JSONValue      `json:"paths"`
	Components JSONValue      `json:"components,omitempty"`
	Servers    []JSONValue    `json:"servers,omitempty"`
	Extra      map[string]any `json:"-"`
}

// EstimateFrequency is the rank-check frequency used by GetCostEstimate to
// compute monthly checks.
type EstimateFrequency string

const (
	EstimateFrequencyDaily   EstimateFrequency = "daily"
	EstimateFrequencyWeekly  EstimateFrequency = "weekly"
	EstimateFrequencyMonthly EstimateFrequency = "monthly"
)

// PricingModel identifies how a provider rate card is priced.
type PricingModel string

const (
	PricingModelFlat PricingModel = "flat"
	PricingModelPlan PricingModel = "plan"
)

// ProviderRateOption is one flat-rate pricing option on a provider rate card.
type ProviderRateOption struct {
	Key           string  `json:"key"`
	Label         string  `json:"label"`
	ShortLabel    string  `json:"short_label"`
	Turnaround    string  `json:"turnaround"`
	UnitCostCents float64 `json:"unit_cost_cents"`
	UnitCostUSD   float64 `json:"unit_cost_usd"`
}

// ProviderRatePlan is one subscription plan tier on a provider rate card.
type ProviderRatePlan struct {
	IncludedChecks    int     `json:"included_checks"`
	Label             string  `json:"label"`
	MonthlyPriceCents float64 `json:"monthly_price_cents"`
	MonthlyPriceUSD   float64 `json:"monthly_price_usd"`
	PlanKey           string  `json:"plan_key"`
}

// ProviderRate is one public SERP provider rate card. Flat rate cards carry
// Options and plan rate cards carry Plans.
type ProviderRate struct {
	CheckedAt    string               `json:"checked_at"`
	Label        string               `json:"label"`
	Notes        string               `json:"notes,omitempty"`
	Options      []ProviderRateOption `json:"options,omitempty"`
	Plans        []ProviderRatePlan   `json:"plans,omitempty"`
	PricingModel PricingModel         `json:"pricing_model"`
	ProviderID   ProviderID           `json:"provider_id"`
	SourceURL    string               `json:"source_url"`
}

// CostEstimateOptions are the query parameters accepted by GetCostEstimate.
// Keywords is required (0 is valid) and capped at 100000. Locations is capped
// at 100 and Devices at 2; both default to 1. Frequency defaults to daily and
// Provider defaults to dataforseo. Option pins a flat-rate option key and
// Plan pins a plan key; both are optional.
type CostEstimateOptions struct {
	Keywords  int
	Devices   int
	Frequency EstimateFrequency
	Locations int
	Option    string
	Plan      string
	Provider  ProviderID
}

// CostEstimate is a public monthly cost estimate for one provider rate card.
type CostEstimate struct {
	ChecksPerRun               int                 `json:"checks_per_run"`
	EffectiveCostPerCheckCents float64             `json:"effective_cost_per_check_cents"`
	ExceedsLargestPlan         bool                `json:"exceeds_largest_plan"`
	ExceedsSelectedPlan        bool                `json:"exceeds_selected_plan"`
	MonthlyChecks              int                 `json:"monthly_checks"`
	MonthlyCostCents           float64             `json:"monthly_cost_cents"`
	MonthlyCostUSD             float64             `json:"monthly_cost_usd"`
	PricingModel               PricingModel        `json:"pricing_model"`
	ProviderID                 ProviderID          `json:"provider_id"`
	RateCheckedAt              string              `json:"rate_checked_at"`
	RateSourceURL              string              `json:"rate_source_url"`
	SelectedOption             *ProviderRateOption `json:"selected_option,omitempty"`
	SelectedPlan               *ProviderRatePlan   `json:"selected_plan,omitempty"`
}
