package bisibility

import "context"

// PreviewPromptTrackerRun calls the project-scoped Prompt Tracker contract.
func (c *Client) PreviewPromptTrackerRun(ctx context.Context, projectID string, input PromptTrackerPreviewInput, options ...RequestOption) (*DataResponse[PromptTrackerPreview], error) {
	path, err := promptTrackerResourcePath(projectID, "runs/preview", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if err := validatePromptTrackerInput(input); err != nil {
		return nil, err
	}
	config.body = input
	return requestJSON[DataResponse[PromptTrackerPreview]](c, ctx, "POST", path, config)
}

// CreatePromptTrackerRun calls the project-scoped Prompt Tracker contract.
func (c *Client) CreatePromptTrackerRun(ctx context.Context, projectID string, input PromptTrackerRunInput, options ...RequestOption) (*DataResponse[PromptTrackerRun], error) {
	path, err := promptTrackerResourcePath(projectID, "runs", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if err := validatePromptTrackerInput(input); err != nil {
		return nil, err
	}
	config.body = input
	if config.idempotencyKey == "" && config.headers.Get("Idempotency-Key") == "" {
		return nil, &ConfigurationError{Message: "Idempotency-Key is required for tracking launch."}
	}
	return requestJSON[DataResponse[PromptTrackerRun]](c, ctx, "POST", path, config)
}

// ListPromptTrackerRuns calls the project-scoped Prompt Tracker contract.
func (c *Client) ListPromptTrackerRuns(ctx context.Context, projectID string, input *PaginationOptions, options ...RequestOption) (*ListResponse[PromptTrackerRun], error) {
	path, err := promptTrackerResourcePath(projectID, "runs", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if input != nil {
		addQuery(config.query, "cursor", input.Cursor)
		addIntQuery(config.query, "limit", input.Limit)
	}
	return requestJSON[ListResponse[PromptTrackerRun]](c, ctx, "GET", path, config)
}

// GetPromptTrackerRun calls the project-scoped Prompt Tracker contract.
func (c *Client) GetPromptTrackerRun(ctx context.Context, projectID string, id string, options ...RequestOption) (*DataResponse[PromptTrackerRun], error) {
	path, err := promptTrackerResourcePath(projectID, "runs", id, "air")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	return requestJSON[DataResponse[PromptTrackerRun]](c, ctx, "GET", path, config)
}

// ListPromptTrackerSamples calls the project-scoped Prompt Tracker contract.
func (c *Client) ListPromptTrackerSamples(ctx context.Context, projectID string, id string, input *PaginationOptions, options ...RequestOption) (*ListResponse[PromptTrackerSample], error) {
	path, err := promptTrackerResourcePath(projectID, "runs", id, "air")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if input != nil {
		addQuery(config.query, "cursor", input.Cursor)
		addIntQuery(config.query, "limit", input.Limit)
	}
	path += "/samples"
	return requestJSON[ListResponse[PromptTrackerSample]](c, ctx, "GET", path, config)
}

// CancelPromptTrackerRun calls the project-scoped Prompt Tracker contract.
func (c *Client) CancelPromptTrackerRun(ctx context.Context, projectID string, id string, options ...RequestOption) (*DataResponse[PromptTrackerRun], error) {
	path, err := promptTrackerResourcePath(projectID, "runs", id, "air")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	path += "/cancel"
	return requestJSON[DataResponse[PromptTrackerRun]](c, ctx, "POST", path, config)
}

// RetryPromptTrackerRun calls the project-scoped Prompt Tracker contract.
func (c *Client) RetryPromptTrackerRun(ctx context.Context, projectID string, id string, input PromptTrackerRunInput, options ...RequestOption) (*DataResponse[PromptTrackerRun], error) {
	path, err := promptTrackerResourcePath(projectID, "runs", id, "air")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if config.idempotencyKey == "" && config.headers.Get("Idempotency-Key") == "" {
		return nil, &ConfigurationError{Message: "Idempotency-Key is required for tracking retry."}
	}
	if err := validatePromptTrackerInput(input); err != nil {
		return nil, err
	}
	config.body = input
	path += "/retry"
	return requestJSON[DataResponse[PromptTrackerRun]](c, ctx, "POST", path, config)
}

// GetPromptTrackerHistory calls the project-scoped Prompt Tracker contract.
func (c *Client) GetPromptTrackerHistory(ctx context.Context, projectID string, input *PaginationOptions, options ...RequestOption) (*ListResponse[PromptTrackerRun], error) {
	path, err := promptTrackerResourcePath(projectID, "history", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if input != nil {
		addQuery(config.query, "cursor", input.Cursor)
		addIntQuery(config.query, "limit", input.Limit)
	}
	return requestJSON[ListResponse[PromptTrackerRun]](c, ctx, "GET", path, config)
}

// GetPromptTrackerTrends calls the project-scoped Prompt Tracker contract.
func (c *Client) GetPromptTrackerTrends(ctx context.Context, projectID string, input *PromptTrackerTrendOptions, options ...RequestOption) (*DataResponse[PromptTrackerTrends], error) {
	path, err := promptTrackerResourcePath(projectID, "trends", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if input != nil {
		addQuery(config.query, "cursor", input.Cursor)
		addIntQuery(config.query, "limit", input.Limit)
		if _, err := promptTrackerResourcePath(projectID, "runs", input.RunID, "air"); err != nil {
			return nil, err
		}
		config.query.Set("run_id", input.RunID)
		if input.PreviousRunID != "" {
			if _, err := promptTrackerResourcePath(projectID, "runs", input.PreviousRunID, "air"); err != nil {
				return nil, err
			}
			config.query.Set("previous_run_id", input.PreviousRunID)
		}
	}
	return requestJSON[DataResponse[PromptTrackerTrends]](c, ctx, "GET", path, config)
}

// ExportPromptTrackerEvidence calls the project-scoped Prompt Tracker contract.
func (c *Client) ExportPromptTrackerEvidence(ctx context.Context, projectID string, input *PromptTrackerEvidenceOptions, options ...RequestOption) (*PromptTrackerExportResponse, error) {
	path, err := promptTrackerResourcePath(projectID, "export", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if input != nil {
		addQuery(config.query, "format", input.Format)
		addQuery(config.query, "cursor", input.Cursor)
		addIntQuery(config.query, "limit", input.Limit)
		if _, err := promptTrackerResourcePath(projectID, "runs", input.RunID, "air"); err != nil {
			return nil, err
		}
		config.query.Set("run_id", input.RunID)
	}
	if input != nil && input.Format == "csv" {
		csv, err := requestText(c, ctx, "GET", path, config)
		if err != nil {
			return nil, err
		}
		return &PromptTrackerExportResponse{CSV: csv, Format: "csv"}, nil
	}
	response, err := requestJSON[PromptTrackerExportResponse](c, ctx, "GET", path, config)
	if response != nil {
		response.Format = "json"
	}
	return response, err
}
