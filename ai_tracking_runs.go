package bisibility

import "context"

// PreviewAITrackingRun calls the project-scoped AI tracking contract.
func (c *Client) PreviewAITrackingRun(ctx context.Context, projectID string, input AITrackingPreviewInput, options ...RequestOption) (*DataResponse[AITrackingPreview], error) {
	path, err := trackingResourcePath(projectID, "runs/preview", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if err := validateAITrackingInput(input); err != nil {
		return nil, err
	}
	config.body = input
	return requestJSON[DataResponse[AITrackingPreview]](c, ctx, "POST", path, config)
}

// CreateAITrackingRun calls the project-scoped AI tracking contract.
func (c *Client) CreateAITrackingRun(ctx context.Context, projectID string, input AITrackingRunInput, options ...RequestOption) (*DataResponse[AITrackingRun], error) {
	path, err := trackingResourcePath(projectID, "runs", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if err := validateAITrackingInput(input); err != nil {
		return nil, err
	}
	config.body = input
	if config.idempotencyKey == "" && config.headers.Get("Idempotency-Key") == "" {
		return nil, &ConfigurationError{Message: "Idempotency-Key is required for tracking launch."}
	}
	return requestJSON[DataResponse[AITrackingRun]](c, ctx, "POST", path, config)
}

// ListAITrackingRuns calls the project-scoped AI tracking contract.
func (c *Client) ListAITrackingRuns(ctx context.Context, projectID string, input *PaginationOptions, options ...RequestOption) (*ListResponse[AITrackingRun], error) {
	path, err := trackingResourcePath(projectID, "runs", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if input != nil {
		addQuery(config.query, "cursor", input.Cursor)
		addIntQuery(config.query, "limit", input.Limit)
	}
	return requestJSON[ListResponse[AITrackingRun]](c, ctx, "GET", path, config)
}

// GetAITrackingRun calls the project-scoped AI tracking contract.
func (c *Client) GetAITrackingRun(ctx context.Context, projectID string, id string, options ...RequestOption) (*DataResponse[AITrackingRun], error) {
	path, err := trackingResourcePath(projectID, "runs", id, "air")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	return requestJSON[DataResponse[AITrackingRun]](c, ctx, "GET", path, config)
}

// ListAITrackingSamples calls the project-scoped AI tracking contract.
func (c *Client) ListAITrackingSamples(ctx context.Context, projectID string, id string, input *PaginationOptions, options ...RequestOption) (*ListResponse[AITrackingSample], error) {
	path, err := trackingResourcePath(projectID, "runs", id, "air")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if input != nil {
		addQuery(config.query, "cursor", input.Cursor)
		addIntQuery(config.query, "limit", input.Limit)
	}
	path += "/samples"
	return requestJSON[ListResponse[AITrackingSample]](c, ctx, "GET", path, config)
}

// CancelAITrackingRun calls the project-scoped AI tracking contract.
func (c *Client) CancelAITrackingRun(ctx context.Context, projectID string, id string, options ...RequestOption) (*DataResponse[AITrackingRun], error) {
	path, err := trackingResourcePath(projectID, "runs", id, "air")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	path += "/cancel"
	return requestJSON[DataResponse[AITrackingRun]](c, ctx, "POST", path, config)
}

// RetryAITrackingRun calls the project-scoped AI tracking contract.
func (c *Client) RetryAITrackingRun(ctx context.Context, projectID string, id string, input AITrackingRunInput, options ...RequestOption) (*DataResponse[AITrackingRun], error) {
	path, err := trackingResourcePath(projectID, "runs", id, "air")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if config.idempotencyKey == "" && config.headers.Get("Idempotency-Key") == "" {
		return nil, &ConfigurationError{Message: "Idempotency-Key is required for tracking retry."}
	}
	if err := validateAITrackingInput(input); err != nil {
		return nil, err
	}
	config.body = input
	path += "/retry"
	return requestJSON[DataResponse[AITrackingRun]](c, ctx, "POST", path, config)
}

// GetAITrackingHistory calls the project-scoped AI tracking contract.
func (c *Client) GetAITrackingHistory(ctx context.Context, projectID string, input *PaginationOptions, options ...RequestOption) (*ListResponse[AITrackingRun], error) {
	path, err := trackingResourcePath(projectID, "history", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if input != nil {
		addQuery(config.query, "cursor", input.Cursor)
		addIntQuery(config.query, "limit", input.Limit)
	}
	return requestJSON[ListResponse[AITrackingRun]](c, ctx, "GET", path, config)
}

// GetAITrackingTrends calls the project-scoped AI tracking contract.
func (c *Client) GetAITrackingTrends(ctx context.Context, projectID string, input *AITrackingTrendOptions, options ...RequestOption) (*DataResponse[AITrackingTrends], error) {
	path, err := trackingResourcePath(projectID, "trends", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if input != nil {
		addQuery(config.query, "cursor", input.Cursor)
		addIntQuery(config.query, "limit", input.Limit)
		if _, err := trackingResourcePath(projectID, "runs", input.RunID, "air"); err != nil {
			return nil, err
		}
		config.query.Set("run_id", input.RunID)
		if input.PreviousRunID != "" {
			if _, err := trackingResourcePath(projectID, "runs", input.PreviousRunID, "air"); err != nil {
				return nil, err
			}
			config.query.Set("previous_run_id", input.PreviousRunID)
		}
	}
	return requestJSON[DataResponse[AITrackingTrends]](c, ctx, "GET", path, config)
}

// ExportAITrackingEvidence calls the project-scoped AI tracking contract.
func (c *Client) ExportAITrackingEvidence(ctx context.Context, projectID string, input *AITrackingEvidenceOptions, options ...RequestOption) (*AITrackingExportResponse, error) {
	path, err := trackingResourcePath(projectID, "export", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if input != nil {
		addQuery(config.query, "format", input.Format)
		addQuery(config.query, "cursor", input.Cursor)
		addIntQuery(config.query, "limit", input.Limit)
		if _, err := trackingResourcePath(projectID, "runs", input.RunID, "air"); err != nil {
			return nil, err
		}
		config.query.Set("run_id", input.RunID)
	}
	if input != nil && input.Format == "csv" {
		csv, err := requestText(c, ctx, "GET", path, config)
		if err != nil {
			return nil, err
		}
		return &AITrackingExportResponse{CSV: csv, Format: "csv"}, nil
	}
	response, err := requestJSON[AITrackingExportResponse](c, ctx, "GET", path, config)
	if response != nil {
		response.Format = "json"
	}
	return response, err
}

// SuggestAITrackingPrompts calls the project-scoped AI tracking contract.
func (c *Client) SuggestAITrackingPrompts(ctx context.Context, projectID string, options ...RequestOption) (*DataResponse[AITrackingSuggestions], error) {
	path, err := trackingResourcePath(projectID, "suggestions", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	config.body = map[string]any{}
	return requestJSON[DataResponse[AITrackingSuggestions]](c, ctx, "POST", path, config)
}

// AcceptAITrackingSuggestions calls the project-scoped AI tracking contract.
func (c *Client) AcceptAITrackingSuggestions(ctx context.Context, projectID string, input AITrackingAcceptanceInput, options ...RequestOption) (*DataResponse[AITrackingAcceptance], error) {
	path, err := trackingResourcePath(projectID, "suggestions/accept", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if err := validateAITrackingInput(input); err != nil {
		return nil, err
	}
	config.body = input
	return requestJSON[DataResponse[AITrackingAcceptance]](c, ctx, "POST", path, config)
}
