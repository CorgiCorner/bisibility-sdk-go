package bisibility

import (
	"context"
	"net/http"
	"strconv"
)

// ListStoredResearchReports reads saved summaries without invoking providers.
func (c *Client) ListStoredResearchReports(ctx context.Context, projectID string, options ...RequestOption) (*StoredResearchReportsResponse, error) {
	return requestJSON[StoredResearchReportsResponse](c, ctx, http.MethodGet, projectResourcePath(projectID, "research/reports"), newRequestConfig(options...))
}

// GetStoredResearchReport reads a saved report. FreshUntil determines its freshness.
func (c *Client) GetStoredResearchReport(ctx context.Context, projectID string, kind StoredResearchReportKind, input StoredResearchReportOptions, options ...RequestOption) (*StoredResearchReportResponse, error) {
	config := newRequestConfig(options...)
	for key, value := range map[string]string{"target": input.Target, "target_scope": input.TargetScope, "mode": input.Mode, "seed": input.Seed, "connection_id": input.ConnectionID, "language_code": input.LanguageCode} {
		if value != "" {
			config.query.Set(key, value)
		}
	}
	if input.IncludeSubdomains != nil {
		config.query.Set("include_subdomains", strconv.FormatBool(*input.IncludeSubdomains))
	}
	if input.IncludeClickstream != nil {
		config.query.Set("include_clickstream", strconv.FormatBool(*input.IncludeClickstream))
	}
	if input.LocationCode != nil {
		config.query.Set("location_code", strconv.Itoa(*input.LocationCode))
	}
	if input.ResultLimit != 0 {
		config.query.Set("result_limit", strconv.Itoa(input.ResultLimit))
	}
	return requestJSON[StoredResearchReportResponse](c, ctx, http.MethodGet, projectMemberResourcePath(projectID, "research/reports", string(kind)), config)
}

// ListProviderBudgets reads own-key and credit budgets independently.
func (c *Client) ListProviderBudgets(ctx context.Context, projectID string, options ...RequestOption) (*ListResponse[ProviderBudgets], error) {
	return requestJSON[ListResponse[ProviderBudgets]](c, ctx, http.MethodGet, projectResourcePath(projectID, "provider-budgets"), newRequestConfig(options...))
}

// UpdateProviderBudgets preserves omitted fields and clears explicitly null fields.
func (c *Client) UpdateProviderBudgets(ctx context.Context, projectID string, providerID ProviderID, input ProviderBudgetsUpdate, options ...RequestOption) (*ProviderBudgets, error) {
	if err := input.validate(); err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[ProviderBudgets](c, ctx, http.MethodPatch, projectMemberResourcePath(projectID, "providers", string(providerID))+"/budgets", config)
}
