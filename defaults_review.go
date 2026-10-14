package bisibility

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// ProjectDefaultsReview carries issuer-bound consent for one reviewed command.
type ProjectDefaultsReview struct {
	ExpectedSourceHash string `json:"expected_source_hash"`
	ExpectedRevision   string `json:"expected_revision"`
	PreviewToken       string `json:"preview_token"`
	IdempotencyKey     string `json:"idempotency_key"`
	Confirmed          bool   `json:"confirmed"`
}
type ProjectDefaultsPreviewRequest struct {
	Kind   string                     `json:"kind"`
	Values map[string]json.RawMessage `json:"values"`
}
type ProjectDefaultsPreviewDraft struct {
	ProjectID       string  `json:"project_id"`
	City            *string `json:"city"`
	Country         string  `json:"country"`
	Device          string  `json:"device"`
	LocationKey     string  `json:"location_key"`
	Frequency       string  `json:"frequency"`
	CronExpression  *string `json:"cron_expression"`
	JitterMinutes   int     `json:"jitter_minutes"`
	SerpDepth       int     `json:"serp_depth"`
	SerpStopOnMatch bool    `json:"serp_stop_on_match"`
	Timezone        string  `json:"timezone"`
}
type ProjectDefaultsPreviewResolved struct {
	LocationKey  string `json:"location_key"`
	Label        string `json:"label"`
	LanguageCode string `json:"language_code"`
	Device       string `json:"device"`
}
type ProjectDefaultsPreviewImpact struct {
	ExistingTargetRetargets int    `json:"existing_target_retargets"`
	NewTargets              int    `json:"new_targets"`
	InheritedDepthTargets   int    `json:"inherited_depth_targets"`
	RecurringEstimate       string `json:"recurring_estimate"`
	ScopeChanged            bool   `json:"scope_changed"`
	CadenceChanged          bool   `json:"cadence_changed"`
}
type ProjectDefaultsPreview struct {
	Request            ProjectDefaultsPreviewRequest  `json:"request"`
	Draft              ProjectDefaultsPreviewDraft    `json:"draft"`
	ExpectedSourceHash string                         `json:"expected_source_hash"`
	ExpectedRevision   string                         `json:"expected_revision"`
	PreviewToken       string                         `json:"preview_token"`
	ExpiresAt          string                         `json:"expires_at"`
	Resolved           ProjectDefaultsPreviewResolved `json:"resolved"`
	Impact             ProjectDefaultsPreviewImpact   `json:"impact"`
	PolicyReady        bool                           `json:"policy_ready"`
}

func (c *Client) assertReviewedDefaultsSupport(ctx context.Context, options ...RequestOption) error {
	capabilities, err := c.GetCapabilities(ctx, options...)
	if err != nil {
		return err
	}
	if capabilities.Features["reviewed_project_defaults"] != "v1" {
		return &ConfigurationError{Message: "This server does not advertise reviewed project defaults v1; no PATCH was sent."}
	}
	return nil
}
func reviewedDefaultsRequestConfig(review *ProjectDefaultsReview, config requestConfig) (requestConfig, error) {
	if review == nil {
		return config, nil
	}
	key := strings.TrimSpace(review.IdempotencyKey)
	if !regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(review.ExpectedSourceHash) || !regexp.MustCompile(`^(0|[1-9][0-9]*)$`).MatchString(review.ExpectedRevision) || len(review.PreviewToken) < 1 || len(review.PreviewToken) > 8192 || key != review.IdempotencyKey || len(key) < 8 || len(key) > 128 || !review.Confirmed {
		return config, &ConfigurationError{Message: "A complete explicit defaults review is required."}
	}
	for _, char := range key {
		if char < 32 || char == 127 {
			return config, &ConfigurationError{Message: "Invalid defaults review command key."}
		}
	}
	if config.idempotencyKey != "" && strings.TrimSpace(config.idempotencyKey) != key {
		return config, &ConfigurationError{Message: "Defaults review and request command keys must agree."}
	}
	config.idempotencyKey = key
	return config, nil
}

// PreviewProjectDefaults reads a signed draft after explicit server capability verification.
func (c *Client) PreviewProjectDefaults(ctx context.Context, projectID string, input ProjectDefaultsPatch, options ...RequestOption) (*ProjectDefaultsPreview, error) {
	if err := requirePublicID(projectID, "projectID", PublicIDPrefixProject); err != nil {
		return nil, err
	}
	if input.Review != nil {
		return nil, &ConfigurationError{Message: "Preview accepts draft settings only."}
	}
	if err := c.assertReviewedDefaultsSupport(ctx, options...); err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	config.query.Set("preview_only", "true")
	config.idempotencyKey = ""
	config.omitIdempotencyKey = true
	config.body = input
	return requestJSONWithValidation[ProjectDefaultsPreview](c, ctx, http.MethodPatch, projectPathRoot+url.PathEscape(projectID)+"/defaults", config, func(preview *ProjectDefaultsPreview) error {
		var project string
		if json.Unmarshal(preview.Request.Values["project_id"], &project) != nil || project != projectID || preview.Draft.ProjectID != projectID || preview.Request.Kind != "patch" || !regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(preview.ExpectedSourceHash) || !regexp.MustCompile(`^(0|[1-9][0-9]*)$`).MatchString(preview.ExpectedRevision) || len(preview.PreviewToken) < 1 || len(preview.PreviewToken) > 8192 {
			return fmt.Errorf("invalid defaults preview binding")
		}
		if _, err := time.Parse(time.RFC3339Nano, preview.ExpiresAt); err != nil {
			return fmt.Errorf("invalid defaults preview expiry")
		}
		return nil
	})
}

// Missing preview fields cannot silently become zero impacts or false consent readiness.
func (p *ProjectDefaultsPreview) UnmarshalJSON(raw []byte) error {
	type wire ProjectDefaultsPreview
	var objects map[string]json.RawMessage
	if err := json.Unmarshal(raw, &objects); err != nil {
		return err
	}
	groups := map[string][]string{
		"":         {"request", "draft", "expected_source_hash", "expected_revision", "preview_token", "expires_at", "resolved", "impact", "policy_ready"},
		"request":  {"kind", "values"},
		"draft":    {"project_id", "city", "country", "device", "location_key", "frequency", "cron_expression", "jitter_minutes", "serp_depth", "serp_stop_on_match", "timezone"},
		"resolved": {"location_key", "label", "language_code", "device"},
		"impact":   {"existing_target_retargets", "new_targets", "inherited_depth_targets", "recurring_estimate", "scope_changed", "cadence_changed"},
	}
	for group, names := range groups {
		fields := objects
		if group != "" {
			var nested map[string]json.RawMessage
			if err := json.Unmarshal(objects[group], &nested); err != nil {
				return err
			}
			fields = nested
		}
		for _, name := range names {
			value, ok := fields[name]
			nullable := group == "draft" && (name == "city" || name == "cron_expression")
			if !ok || (!nullable && string(value) == "null") {
				return fmt.Errorf("missing required defaults preview field: %s.%s", group, name)
			}
		}
	}
	var decoded wire
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return err
	}
	*p = ProjectDefaultsPreview(decoded)
	return nil
}
