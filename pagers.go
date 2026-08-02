package bisibility

import (
	"context"
	"net/http"
)

func paginationCopy(options *PaginationOptions, cursor string) PaginationOptions {
	if options == nil {
		return PaginationOptions{Cursor: cursor}
	}
	copy := *options
	copy.Cursor = cursor
	return copy
}

func paginationInitial(options *PaginationOptions) PaginationOptions {
	if options == nil {
		return PaginationOptions{}
	}
	return *options
}

// IterateAPIKeys returns a pager over API keys.
func (c *Client) IterateAPIKeys(ctx context.Context, pagination *PaginationOptions, options ...RequestOption) *Pager[APIKey] {
	initial := paginationInitial(pagination)
	return newPager(ctx, initial.Cursor, func(ctx context.Context, cursor string) ([]APIKey, *string, error) {
		page, err := c.ListAPIKeys(ctx, ptr(paginationCopy(&initial, cursor)), options...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.Meta.NextCursor, nil
	})
}

// IterateProjectAPIKeys returns a pager over project API keys.
func (c *Client) IterateProjectAPIKeys(ctx context.Context, projectID string, pagination *PaginationOptions, options ...RequestOption) *Pager[APIKey] {
	initial := paginationInitial(pagination)
	return newPager(ctx, initial.Cursor, func(ctx context.Context, cursor string) ([]APIKey, *string, error) {
		page, err := c.ListProjectAPIKeys(ctx, projectID, ptr(paginationCopy(&initial, cursor)), options...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.Meta.NextCursor, nil
	})
}

// IterateWebhooks returns a pager over project webhooks.
func (c *Client) IterateWebhooks(ctx context.Context, projectID string, pagination *PaginationOptions, options ...RequestOption) *Pager[Webhook] {
	initial := paginationInitial(pagination)
	return newPager(ctx, initial.Cursor, func(ctx context.Context, cursor string) ([]Webhook, *string, error) {
		page, err := c.ListWebhooks(ctx, projectID, ptr(paginationCopy(&initial, cursor)), options...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.Meta.NextCursor, nil
	})
}

// IterateKeywords returns a pager over project keywords while preserving filters.
func (c *Client) IterateKeywords(ctx context.Context, projectID string, filters *ListKeywordsOptions, options ...RequestOption) *Pager[Keyword] {
	initial := ListKeywordsOptions{}
	if filters != nil {
		initial = *filters
	}
	return newPager(ctx, initial.Cursor, func(ctx context.Context, cursor string) ([]Keyword, *string, error) {
		pageFilters := initial
		pageFilters.Cursor = cursor
		page, err := c.ListKeywords(ctx, projectID, &pageFilters, options...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.Meta.NextCursor, nil
	})
}

// IterateRankChecks returns a pager over keyword rank checks while preserving filters.
func (c *Client) IterateRankChecks(ctx context.Context, keywordID string, filters *ListRankChecksOptions, options ...RequestOption) *Pager[RankCheck] {
	initial := ListRankChecksOptions{}
	if filters != nil {
		initial = *filters
	}
	return newPager(ctx, initial.Cursor, func(ctx context.Context, cursor string) ([]RankCheck, *string, error) {
		pageFilters := initial
		pageFilters.Cursor = cursor
		page, err := c.ListRankChecks(ctx, keywordID, &pageFilters, options...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.Meta.NextCursor, nil
	})
}

// IterateRankHistory returns a pager over the JSON rank-history export.
// CSV format is ignored because CSV exports are complete and not cursor-paginated.
func (c *Client) IterateRankHistory(ctx context.Context, projectID string, filters *ExportRankHistoryOptions, options ...RequestOption) *Pager[RankHistoryExportRow] {
	initial := ExportRankHistoryOptions{}
	if filters != nil {
		initial = *filters
	}
	initial.Format = RankHistoryExportFormatJSON
	return newPager(ctx, initial.Cursor, func(ctx context.Context, cursor string) ([]RankHistoryExportRow, *string, error) {
		pageFilters := initial
		pageFilters.Cursor = cursor
		page, err := c.ExportRankHistory(ctx, projectID, &pageFilters, options...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.Meta.NextCursor, nil
	})
}

// IterateSignals returns a pager over project signals while preserving filters.
func (c *Client) IterateSignals(ctx context.Context, projectID string, filters *ListSignalsOptions, options ...RequestOption) *Pager[Signal] {
	initial := ListSignalsOptions{}
	if filters != nil {
		initial = *filters
	}
	return newPager(ctx, initial.Cursor, func(ctx context.Context, cursor string) ([]Signal, *string, error) {
		pageFilters := initial
		pageFilters.Cursor = cursor
		page, err := c.ListSignals(ctx, projectID, &pageFilters, options...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.Meta.NextCursor, nil
	})
}

func simpleProjectPager[T any](ctx context.Context, pagination *PaginationOptions, fetch func(context.Context, *PaginationOptions) (*ListResponse[T], error)) *Pager[T] {
	initial := paginationInitial(pagination)
	return newPager(ctx, initial.Cursor, func(ctx context.Context, cursor string) ([]T, *string, error) {
		page, err := fetch(ctx, ptr(paginationCopy(&initial, cursor)))
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.Meta.NextCursor, nil
	})
}

func (c *Client) IterateAlertRules(ctx context.Context, projectID string, pagination *PaginationOptions, options ...RequestOption) *Pager[AlertRule] {
	return simpleProjectPager(ctx, pagination, func(ctx context.Context, page *PaginationOptions) (*ListResponse[AlertRule], error) {
		return c.ListAlertRules(ctx, projectID, page, options...)
	})
}
func (c *Client) IterateTriggeredAlerts(ctx context.Context, projectID string, pagination *PaginationOptions, options ...RequestOption) *Pager[TriggeredAlert] {
	return simpleProjectPager(ctx, pagination, func(ctx context.Context, page *PaginationOptions) (*ListResponse[TriggeredAlert], error) {
		return c.ListTriggeredAlerts(ctx, projectID, page, options...)
	})
}
func (c *Client) IterateTeamMembers(ctx context.Context, projectID string, pagination *PaginationOptions, options ...RequestOption) *Pager[TeamMember] {
	return simpleProjectPager(ctx, pagination, func(ctx context.Context, page *PaginationOptions) (*ListResponse[TeamMember], error) {
		return c.ListTeamMembers(ctx, projectID, page, options...)
	})
}
func (c *Client) IterateTeamInvites(ctx context.Context, projectID string, pagination *PaginationOptions, options ...RequestOption) *Pager[TeamInvite] {
	return simpleProjectPager(ctx, pagination, func(ctx context.Context, page *PaginationOptions) (*ListResponse[TeamInvite], error) {
		return c.ListTeamInvites(ctx, projectID, page, options...)
	})
}
func (c *Client) IterateProviders(ctx context.Context, projectID string, pagination *PaginationOptions, options ...RequestOption) *Pager[Provider] {
	return simpleProjectPager(ctx, pagination, func(ctx context.Context, page *PaginationOptions) (*ListResponse[Provider], error) {
		return c.ListProviders(ctx, projectID, page, options...)
	})
}
func (c *Client) IterateSavedKeywords(ctx context.Context, projectID string, pagination *PaginationOptions, options ...RequestOption) *Pager[SavedKeyword] {
	return simpleProjectPager(ctx, pagination, func(ctx context.Context, page *PaginationOptions) (*ListResponse[SavedKeyword], error) {
		return c.ListSavedKeywords(ctx, projectID, page, options...)
	})
}
func (c *Client) IterateSavedViews(ctx context.Context, projectID string, pagination *PaginationOptions, options ...RequestOption) *Pager[SavedView] {
	return simpleProjectPager(ctx, pagination, func(ctx context.Context, page *PaginationOptions) (*ListResponse[SavedView], error) {
		return c.ListSavedViews(ctx, projectID, page, options...)
	})
}

// IterateCompetitors returns a pager over managed project competitors.
func (c *Client) IterateCompetitors(ctx context.Context, projectID string, pagination *PaginationOptions, options ...RequestOption) *Pager[ManagedCompetitor] {
	initial := paginationInitial(pagination)
	return newPager(ctx, initial.Cursor, func(ctx context.Context, cursor string) ([]ManagedCompetitor, *string, error) {
		page, err := c.ListCompetitors(ctx, projectID, ptr(paginationCopy(&initial, cursor)), options...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.Meta.NextCursor, nil
	})
}

// IterateMigrationTokens returns a pager over active migration tokens.
func (c *Client) IterateMigrationTokens(ctx context.Context, projectID string, pagination *PaginationOptions, options ...RequestOption) *Pager[MigrationToken] {
	initial := paginationInitial(pagination)
	return newPager(ctx, initial.Cursor, func(ctx context.Context, cursor string) ([]MigrationToken, *string, error) {
		config := newRequestConfig(options...)
		pageOptions := paginationCopy(&initial, cursor)
		addPagination(config.query, &pageOptions)
		page, err := requestJSON[ListMigrationTokensResponse](c, ctx, http.MethodGet, projectResourcePath(projectID, migrationTokensResource), config)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.Meta.NextCursor, nil
	})
}

func ptr[T any](value T) *T { return &value }
