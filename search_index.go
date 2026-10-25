package bisibility

import (
	"context"
	"net/http"
	"net/url"
	"time"
	"unicode/utf8"
)

// ListSearchIndexSitemaps lists sitemaps using the project's existing connection.
// Results are Google's indexed state, not a live test, and may be partial or delayed.
// This operation is free and accepts no property override.
func (c *Client) ListSearchIndexSitemaps(ctx context.Context, projectID string, options ...RequestOption) (*SearchIndexSitemaps, error) {
	return requestJSON[SearchIndexSitemaps](c, ctx, http.MethodGet, projectResourcePath(projectID, "search-index/sitemaps"), newRequestConfig(options...))
}

// GetSearchIndexSitemap gets a full sitemap URL within the project's connected property.
// Pass an unescaped absolute HTTP or HTTPS sitemap URL of 1-2048 characters;
// it is encoded as one path segment.
// Results are Google's indexed state, not a live test, and may be partial or delayed.
// This operation is free and accepts no property override.
func (c *Client) GetSearchIndexSitemap(ctx context.Context, projectID, path string, options ...RequestOption) (*SearchIndexSitemapResponse, error) {
	sitemapURL, err := url.Parse(path)
	if err != nil || (sitemapURL.Scheme != "http" && sitemapURL.Scheme != "https") ||
		sitemapURL.Host == "" || utf8.RuneCountInString(path) > 2048 {
		return nil, &ConfigurationError{Message: "GetSearchIndexSitemap requires an absolute HTTP or HTTPS sitemap URL of 1-2048 characters."}
	}

	return requestJSON[SearchIndexSitemapResponse](c, ctx, http.MethodGet, projectMemberResourcePath(projectID, "search-index/sitemaps", path), newRequestConfig(options...))
}

// InspectSearchIndexURLs inspects 1-10 URLs using the project's existing connection.
// Results are Google's indexed state, not a live test, and may be partial or delayed.
// This free operation requires write scope and accepts no property override.
// Manual and background calls share a project daily allowance, not Google's quota.
// Per-URL failures are returned in Results. Pass WithIdempotencyKey to replay a
// response without another inspection and enable automatic retries.
func (c *Client) InspectSearchIndexURLs(ctx context.Context, projectID string, input SearchIndexInspectionInput, options ...RequestOption) (*SearchIndexInspections, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[SearchIndexInspections](c, ctx, http.MethodPost, projectResourcePath(projectID, "search-index/inspections"), config)
}

// GetPageIndexingReport reads stored report completeness, freshness and grouped reasons.
// Results are Google's indexed state, not a live test, and may be partial or delayed.
// This free operation uses the project's existing connection and never calls Google.
func (c *Client) GetPageIndexingReport(ctx context.Context, projectID string, options ...RequestOption) (*PageIndexingReport, error) {
	return requestJSON[PageIndexingReport](c, ctx, http.MethodGet, projectResourcePath(projectID, "search-index/report"), newRequestConfig(options...))
}

// ListPageIndexingURLs lists one cursor page of stored sitemap and keyword target URLs.
// Manual-only URLs require Source=PageIndexingSourceManual. History includes up to
// 100 recent changes per URL. Limit is 1-100, default 50; pass cursors unchanged.
// ChangedSince includes URLs with a stored tracked change at or after that time,
// applied before pagination.
// Results are Google's indexed state, not a live test, and may be partial or delayed.
// This free operation uses the project's existing connection and never calls Google.
func (c *Client) ListPageIndexingURLs(ctx context.Context, projectID string, input *ListPageIndexingURLsOptions, options ...RequestOption) (*PageIndexingURLs, error) {
	config := newRequestConfig(options...)
	if input != nil {
		addQuery(config.query, "group", string(input.Group))
		addQuery(config.query, "coverage_state", input.CoverageState)
		addQuery(config.query, "sitemap", input.Sitemap)
		addQuery(config.query, "source", string(input.Source))
		addQuery(config.query, "search", input.Search)
		if input.ChangedSince != nil {
			config.query.Set("changed_since", input.ChangedSince.UTC().Format(time.RFC3339Nano))
		}
		addQuery(config.query, "cursor", input.Cursor)
		addIntQuery(config.query, "limit", input.Limit)
		if input.History {
			config.query.Set("history", "true")
		}
	}
	return requestJSON[PageIndexingURLs](c, ctx, http.MethodGet, projectResourcePath(projectID, "search-index/urls"), config)
}

// IteratePageIndexingURLs follows stored URL cursor pages while preserving filters.
// Results are Google's indexed state, not a live test, and may be partial or delayed.
func (c *Client) IteratePageIndexingURLs(ctx context.Context, projectID string, filters *ListPageIndexingURLsOptions, options ...RequestOption) *Pager[PageIndexingURL] {
	initial := ListPageIndexingURLsOptions{}
	if filters != nil {
		initial = *filters
	}
	return newPager(ctx, initial.Cursor, func(ctx context.Context, cursor string) ([]PageIndexingURL, *string, error) {
		pageFilters := initial
		pageFilters.Cursor = cursor
		page, err := c.ListPageIndexingURLs(ctx, projectID, &pageFilters, options...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.Meta.NextCursor, nil
	})
}

// GetPageIndexingSettings reads background inspection settings and estimated usage.
// This operation is free, uses the project's existing connection and never calls Google.
// Related results are Google's indexed state, not a live test, and may be partial or delayed.
func (c *Client) GetPageIndexingSettings(ctx context.Context, projectID string, options ...RequestOption) (*PageIndexingSettings, error) {
	return requestJSON[PageIndexingSettings](c, ctx, http.MethodGet, projectResourcePath(projectID, "search-index/settings"), newRequestConfig(options...))
}

// UpdatePageIndexingSettings partially updates background inspection settings.
// Supply at least one field. Nil fields are omitted; neither field accepts null.
// This free operation requires write scope, uses the project's existing connection
// and never calls Google. Pass WithIdempotencyKey to enable automatic retries.
// Related results are Google's indexed state, not a live test, and may be partial or delayed.
func (c *Client) UpdatePageIndexingSettings(ctx context.Context, projectID string, input PageIndexingSettingsPatch, options ...RequestOption) (*PageIndexingSettings, error) {
	if input.BudgetMode == nil && input.InspectionDailyLimit == nil {
		return nil, &ConfigurationError{Message: "UpdatePageIndexingSettings requires at least one of BudgetMode or InspectionDailyLimit."}
	}

	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[PageIndexingSettings](c, ctx, http.MethodPatch, projectResourcePath(projectID, "search-index/settings"), config)
}
