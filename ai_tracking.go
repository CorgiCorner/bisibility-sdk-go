package bisibility

import (
	"context"
	"regexp"
)

func trackingResourcePath(projectID, resource, id, prefix string) (string, error) {
	if prefix != "" && !regexp.MustCompile("^"+prefix+"_[a-z][a-z0-9]{23}$").MatchString(id) {
		return "", &ConfigurationError{Message: "Expected a " + prefix + "_ public ID."}
	}
	path := projectResourcePath(projectID, "ai-tracking/"+resource)
	if prefix != "" {
		path += "/" + id
	}
	return path, nil
}

// ListAITrackingTopics calls the project-scoped AI tracking contract.
func (c *Client) ListAITrackingTopics(ctx context.Context, projectID string, input *PaginationOptions, options ...RequestOption) (*ListResponse[AITrackingTopic], error) {
	path, err := trackingResourcePath(projectID, "topics", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if input != nil {
		addQuery(config.query, "cursor", input.Cursor)
		addIntQuery(config.query, "limit", input.Limit)
	}
	return requestJSON[ListResponse[AITrackingTopic]](c, ctx, "GET", path, config)
}

// CreateAITrackingTopic calls the project-scoped AI tracking contract.
func (c *Client) CreateAITrackingTopic(ctx context.Context, projectID string, input AITrackingTopicInput, options ...RequestOption) (*DataResponse[AITrackingTopic], error) {
	path, err := trackingResourcePath(projectID, "topics", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if err := validateAITrackingInput(input); err != nil {
		return nil, err
	}
	config.body = input
	return requestJSON[DataResponse[AITrackingTopic]](c, ctx, "POST", path, config)
}

// UpdateAITrackingTopic calls the project-scoped AI tracking contract.
func (c *Client) UpdateAITrackingTopic(ctx context.Context, projectID string, id string, input AITrackingTopicPatch, options ...RequestOption) (*DataResponse[AITrackingTopic], error) {
	path, err := trackingResourcePath(projectID, "topics", id, "ait")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if err := validateAITrackingInput(input); err != nil {
		return nil, err
	}
	config.body = input
	return requestJSON[DataResponse[AITrackingTopic]](c, ctx, "PATCH", path, config)
}

// ArchiveAITrackingTopic calls the project-scoped AI tracking contract.
func (c *Client) ArchiveAITrackingTopic(ctx context.Context, projectID string, id string, options ...RequestOption) (*DataResponse[AITrackingTopic], error) {
	path, err := trackingResourcePath(projectID, "topics", id, "ait")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	return requestJSON[DataResponse[AITrackingTopic]](c, ctx, "DELETE", path, config)
}

// ListAITrackingPrompts calls the project-scoped AI tracking contract.
func (c *Client) ListAITrackingPrompts(ctx context.Context, projectID string, input *PaginationOptions, options ...RequestOption) (*ListResponse[AITrackingPrompt], error) {
	path, err := trackingResourcePath(projectID, "prompts", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if input != nil {
		addQuery(config.query, "cursor", input.Cursor)
		addIntQuery(config.query, "limit", input.Limit)
	}
	return requestJSON[ListResponse[AITrackingPrompt]](c, ctx, "GET", path, config)
}

// CreateAITrackingPrompt calls the project-scoped AI tracking contract.
func (c *Client) CreateAITrackingPrompt(ctx context.Context, projectID string, input AITrackingPromptInput, options ...RequestOption) (*DataResponse[AITrackingPrompt], error) {
	path, err := trackingResourcePath(projectID, "prompts", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if err := validateAITrackingInput(input); err != nil {
		return nil, err
	}
	config.body = input
	return requestJSON[DataResponse[AITrackingPrompt]](c, ctx, "POST", path, config)
}

// UpdateAITrackingPrompt calls the project-scoped AI tracking contract.
func (c *Client) UpdateAITrackingPrompt(ctx context.Context, projectID string, id string, input AITrackingPromptPatch, options ...RequestOption) (*DataResponse[AITrackingPrompt], error) {
	path, err := trackingResourcePath(projectID, "prompts", id, "aip")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if err := validateAITrackingInput(input); err != nil {
		return nil, err
	}
	config.body = input
	return requestJSON[DataResponse[AITrackingPrompt]](c, ctx, "PATCH", path, config)
}

// ArchiveAITrackingPrompt calls the project-scoped AI tracking contract.
func (c *Client) ArchiveAITrackingPrompt(ctx context.Context, projectID string, id string, options ...RequestOption) (*DataResponse[AITrackingPrompt], error) {
	path, err := trackingResourcePath(projectID, "prompts", id, "aip")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	return requestJSON[DataResponse[AITrackingPrompt]](c, ctx, "DELETE", path, config)
}

// ListAITrackingSchedules calls the project-scoped AI tracking contract.
func (c *Client) ListAITrackingSchedules(ctx context.Context, projectID string, input *PaginationOptions, options ...RequestOption) (*ListResponse[AITrackingSchedule], error) {
	path, err := trackingResourcePath(projectID, "schedules", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if input != nil {
		addQuery(config.query, "cursor", input.Cursor)
		addIntQuery(config.query, "limit", input.Limit)
	}
	return requestJSON[ListResponse[AITrackingSchedule]](c, ctx, "GET", path, config)
}

// CreateAITrackingSchedule calls the project-scoped AI tracking contract.
func (c *Client) CreateAITrackingSchedule(ctx context.Context, projectID string, input AITrackingScheduleInput, options ...RequestOption) (*DataResponse[AITrackingSchedule], error) {
	path, err := trackingResourcePath(projectID, "schedules", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if err := validateAITrackingInput(input); err != nil {
		return nil, err
	}
	config.body = input
	return requestJSON[DataResponse[AITrackingSchedule]](c, ctx, "POST", path, config)
}

// UpdateAITrackingSchedule calls the project-scoped AI tracking contract. The
// PATCH is a partial update: only fields the caller sets reach the server, and
// consent is sent at the top level because the API reads it from there to
// approve a reviewed configuration.
func (c *Client) UpdateAITrackingSchedule(ctx context.Context, projectID string, id string, input AITrackingSchedulePatch, options ...RequestOption) (*DataResponse[AITrackingSchedule], error) {
	path, err := trackingResourcePath(projectID, "schedules", id, "ais")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if err := validateAITrackingInput(input); err != nil {
		return nil, err
	}
	config.body = input
	return requestJSON[DataResponse[AITrackingSchedule]](c, ctx, "PATCH", path, config)
}

// ArchiveAITrackingSchedule calls the project-scoped AI tracking contract.
func (c *Client) ArchiveAITrackingSchedule(ctx context.Context, projectID string, id string, options ...RequestOption) (*DataResponse[AITrackingSchedule], error) {
	path, err := trackingResourcePath(projectID, "schedules", id, "ais")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	return requestJSON[DataResponse[AITrackingSchedule]](c, ctx, "DELETE", path, config)
}
