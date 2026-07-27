package bisibility

import (
	"context"
	"net/http"
)

// AnalyzeBacklinks analyzes a backlink target or returns a free estimate-only dry run.
// The endpoint requires write scope because cache misses can spend provider budget.
func (c *Client) AnalyzeBacklinks(ctx context.Context, projectID string, input AnalyzeBacklinksOptions, options ...RequestOption) (*BacklinksSnapshotResponse, error) {
	config := newRequestConfig(options...)
	if input.EstimateOnly {
		config.query.Set("estimate_only", "true")
	}
	if input.Fresh {
		config.query.Set("fresh", "true")
	}
	if input.IncludeSubdomains {
		config.query.Set("include_subdomains", "true")
	}
	addIntQuery(config.query, "max_cost_cents", input.MaxCostCents)
	addQuery(config.query, "mode", string(input.Mode))
	addIntQuery(config.query, "result_limit", input.ResultLimit)
	config.query.Set("target", input.Target)
	addQuery(config.query, "target_scope", string(input.TargetScope))
	return requestJSON[BacklinksSnapshotResponse](c, ctx, http.MethodGet, projectResourcePath(projectID, "backlinks"), config)
}

// LoadMoreBacklinkRows loads paid rows into an unexpired backlinks snapshot.
// The endpoint requires write scope because the provider call spends project budget.
func (c *Client) LoadMoreBacklinkRows(ctx context.Context, projectID string, input LoadMoreBacklinkRowsOptions, options ...RequestOption) (*BacklinksSnapshotResponse, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[BacklinksSnapshotResponse](c, ctx, http.MethodPost, projectResourcePath(projectID, "backlinks/rows"), config)
}
