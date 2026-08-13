package bisibility

import (
	"context"
	"net/http"
)

// AnalyzeDomainOverview estimates or loads a domain overview report.
// Non-estimate calls can spend provider budget and require an explicit maximum cost.
func (c *Client) AnalyzeDomainOverview(ctx context.Context, projectID string, input AnalyzeDomainOverviewOptions, options ...RequestOption) (*DomainOverviewAnalyzeResponse, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[DomainOverviewAnalyzeResponse](c, ctx, http.MethodPost, projectResourcePath(projectID, "domain-overview/analyze"), config)
}

// LoadDomainOverviewHistory loads the historical index series for an unexpired overview snapshot.
// The request can spend provider budget and requires an explicit maximum cost, including zero for
// a cache-only attempt.
func (c *Client) LoadDomainOverviewHistory(ctx context.Context, projectID string, input LoadDomainOverviewHistoryOptions, options ...RequestOption) (*DomainOverviewHistoryResponse, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[DomainOverviewHistoryResponse](c, ctx, http.MethodPost, projectResourcePath(projectID, "domain-overview/history"), config)
}

// LoadDomainOverviewKeywords loads one ranked-keyword page for a domain overview.
// The request can spend provider budget and requires an explicit maximum cost, including zero for
// a cache-only attempt.
func (c *Client) LoadDomainOverviewKeywords(ctx context.Context, projectID string, input LoadDomainOverviewKeywordsOptions, options ...RequestOption) (*DomainOverviewKeywordsResponse, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[DomainOverviewKeywordsResponse](c, ctx, http.MethodPost, projectResourcePath(projectID, "domain-overview/keywords"), config)
}

// LoadDomainOverviewPages loads one relevant-page page for a domain overview.
// The request can spend provider budget and requires an explicit maximum cost, including zero for
// a cache-only attempt.
func (c *Client) LoadDomainOverviewPages(ctx context.Context, projectID string, input LoadDomainOverviewPagesOptions, options ...RequestOption) (*DomainOverviewPagesResponse, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[DomainOverviewPagesResponse](c, ctx, http.MethodPost, projectResourcePath(projectID, "domain-overview/pages"), config)
}
