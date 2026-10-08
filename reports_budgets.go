package bisibility

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
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

// GetProjectContext reads the project's guidance for agents.
func (c *Client) GetProjectContext(ctx context.Context, projectID string, options ...RequestOption) (*DataResponse[ProjectContext], error) {
	return requestJSON[DataResponse[ProjectContext]](c, ctx, http.MethodGet, projectResourcePath(projectID, "context"), newRequestConfig(options...))
}

// UpdateProjectContext replaces the project's guidance fields.
func (c *Client) UpdateProjectContext(ctx context.Context, projectID string, input ProjectContextInput, options ...RequestOption) (*DataResponse[ProjectContext], error) {
	for _, text := range []string{input.Business, input.Audience, input.Products, input.Goals, input.AgentRules} {
		if utf8.RuneCountInString(text) > 4000 {
			return nil, &ConfigurationError{Message: "Project context fields must contain at most 4000 characters."}
		}
	}
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[DataResponse[ProjectContext]](c, ctx, http.MethodPatch, projectResourcePath(projectID, "context"), config)
}

// ListAgentReports returns a cursor page of saved analysis summaries.
func (c *Client) ListAgentReports(ctx context.Context, projectID string, input *ListAgentReportsOptions, options ...RequestOption) (*ListResponse[AgentReportSummary], error) {
	config := newRequestConfig(options...)
	if input != nil {
		if input.Limit < 0 || input.Limit > 100 {
			return nil, &ConfigurationError{Message: "Report limit must be between 1 and 100."}
		}
		if input.Kind != "" && !regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]{0,63}$`).MatchString(input.Kind) {
			return nil, &ConfigurationError{Message: "Report kind is invalid."}
		}
		addQuery(config.query, "kind", input.Kind)
		addQuery(config.query, "cursor", input.Cursor)
		addIntQuery(config.query, "limit", input.Limit)
	}
	return requestJSON[ListResponse[AgentReportSummary]](c, ctx, http.MethodGet, projectResourcePath(projectID, "agent-reports"), config)
}

// CreateAgentReport saves an external report, preserving JSON body and provenance keys.
func (c *Client) CreateAgentReport(ctx context.Context, projectID string, input CreateAgentReportInput, options ...RequestOption) (*DataResponse[AgentReportResource], error) {
	if !regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]{0,63}$`).MatchString(input.Kind) || strings.EqualFold(input.Kind, "site_audit") || strings.EqualFold(input.Kind, "ai_visibility") || strings.EqualFold(input.Kind, "prompt_explorer") {
		return nil, &ConfigurationError{Message: "Report kind must identify external analysis."}
	}
	if len(strings.TrimSpace(input.Title)) == 0 || utf8.RuneCountInString(input.Title) > 160 || input.Body == nil {
		return nil, &ConfigurationError{Message: "Report title and JSON body are required."}
	}
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[DataResponse[AgentReportResource]](c, ctx, http.MethodPost, projectResourcePath(projectID, "agent-reports"), config)
}

// GetAgentReport reads a saved report scoped to the project.
func (c *Client) GetAgentReport(ctx context.Context, projectID, reportID string, options ...RequestOption) (*DataResponse[AgentReportResource], error) {
	return requestJSON[DataResponse[AgentReportResource]](c, ctx, http.MethodGet, projectMemberResourcePath(projectID, "agent-reports", reportID), newRequestConfig(options...))
}

// ListSiteAudits reads saved site-audit summaries.
func (c *Client) ListSiteAudits(ctx context.Context, projectID string, options ...RequestOption) (*DataResponse[[]AgentReportSummary], error) {
	return requestJSON[DataResponse[[]AgentReportSummary]](c, ctx, http.MethodGet, projectResourcePath(projectID, "site-audits"), newRequestConfig(options...))
}

// RunSiteAudit starts a bounded crawl or returns a cached report.
func (c *Client) RunSiteAudit(ctx context.Context, projectID string, input *RunSiteAuditOptions, options ...RequestOption) (*DataResponse[SiteAuditReport], error) {
	if input == nil {
		input = &RunSiteAuditOptions{}
	}
	if input.MaxPages < 0 || input.MaxPages > 15 {
		return nil, &ConfigurationError{Message: "max_pages must be between 1 and 15."}
	}
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[DataResponse[SiteAuditReport]](c, ctx, http.MethodPost, projectResourcePath(projectID, "site-audits"), config)
}

// GetSiteAudit reads a saved audit report scoped to the project.
func (c *Client) GetSiteAudit(ctx context.Context, projectID, reportID string, options ...RequestOption) (*DataResponse[SiteAuditReport], error) {
	return requestJSON[DataResponse[SiteAuditReport]](c, ctx, http.MethodGet, projectMemberResourcePath(projectID, "site-audits", reportID), newRequestConfig(options...))
}

func validateAIResearchInput(input AIResearchInput) error {
	if len(strings.TrimSpace(input.Brand)) == 0 || utf8.RuneCountInString(input.Brand) > 120 || len(strings.TrimSpace(input.Domain)) == 0 || utf8.RuneCountInString(input.Domain) > 63 || input.MaxCostCents < 0 || input.MaxCostCents > 1000 {
		return &ConfigurationError{Message: "AI analysis requires a brand, domain, and cost cap between 0 and 1000 cents."}
	}
	return nil
}

// AnalyzeAIVisibility estimates or reads observed AI-visibility datasets.
func (c *Client) AnalyzeAIVisibility(ctx context.Context, projectID string, input AnalyzeAIVisibilityOptions, options ...RequestOption) (*DataResponse[AIAnalysisOutcome], error) {
	if err := validateAIResearchInput(input.AIResearchInput); err != nil {
		return nil, err
	}
	if input.Limit < 0 || input.Limit > 20 || input.LocationCode < 0 || input.LocationCode > 9007199254740991 || (input.Platform != "" && input.Platform != "chat_gpt" && input.Platform != "google") || (input.TargetType != "" && input.TargetType != "brand" && input.TargetType != "domain") || (input.LanguageCode != "" && !regexp.MustCompile(`^[a-z]{2}(?:-[A-Z]{2})?$`).MatchString(input.LanguageCode)) {
		return nil, &ConfigurationError{Message: "AI visibility options are invalid."}
	}
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[DataResponse[AIAnalysisOutcome]](c, ctx, http.MethodPost, projectResourcePath(projectID, "ai-visibility"), config)
}

// CompareAIPrompts estimates or runs synthetic prompt comparisons.
func (c *Client) CompareAIPrompts(ctx context.Context, projectID string, input CompareAIPromptsOptions, options ...RequestOption) (*DataResponse[AIAnalysisOutcome], error) {
	if err := validateAIResearchInput(input.AIResearchInput); err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(input.Prompt)) == 0 || utf8.RuneCountInString(input.Prompt) > 500 || len(input.Models) > 2 {
		return nil, &ConfigurationError{Message: "Prompt and model selection are invalid."}
	}
	seen := map[string]bool{}
	for _, model := range input.Models {
		if (model != "gpt-4.1-mini" && model != "gpt-4.1-nano") || seen[model] {
			return nil, &ConfigurationError{Message: fmt.Sprintf("Unsupported or duplicate prompt model %q.", model)}
		}
		seen[model] = true
	}
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[DataResponse[AIAnalysisOutcome]](c, ctx, http.MethodPost, projectResourcePath(projectID, "prompt-explorer"), config)
}
