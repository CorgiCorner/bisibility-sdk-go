package bisibility

import (
	"context"
	"regexp"
)

func promptTrackerResourcePath(projectID, resource, id, prefix string) (string, error) {
	if prefix != "" && !regexp.MustCompile("^"+prefix+"_[a-z][a-z0-9]{23}$").MatchString(id) {
		return "", &ConfigurationError{Message: "Expected a " + prefix + "_ public ID."}
	}
	path := projectResourcePath(projectID, "prompt-tracker/"+resource)
	if prefix != "" {
		path += "/" + id
	}
	return path, nil
}

// ListPromptTrackerTopics calls the project-scoped Prompt Tracker contract.
func (c *Client) ListPromptTrackerTopics(ctx context.Context, projectID string, input *PaginationOptions, options ...RequestOption) (*ListResponse[PromptTrackerTopic], error) {
	path, err := promptTrackerResourcePath(projectID, "topics", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if input != nil {
		addQuery(config.query, "cursor", input.Cursor)
		addIntQuery(config.query, "limit", input.Limit)
	}
	return requestJSON[ListResponse[PromptTrackerTopic]](c, ctx, "GET", path, config)
}

// CreatePromptTrackerTopic calls the project-scoped Prompt Tracker contract.
func (c *Client) CreatePromptTrackerTopic(ctx context.Context, projectID string, input PromptTrackerTopicInput, options ...RequestOption) (*DataResponse[PromptTrackerTopic], error) {
	path, err := promptTrackerResourcePath(projectID, "topics", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if err := validatePromptTrackerInput(input); err != nil {
		return nil, err
	}
	config.body = input
	return requestJSON[DataResponse[PromptTrackerTopic]](c, ctx, "POST", path, config)
}

// UpdatePromptTrackerTopic calls the project-scoped Prompt Tracker contract.
func (c *Client) UpdatePromptTrackerTopic(ctx context.Context, projectID string, id string, input PromptTrackerTopicPatch, options ...RequestOption) (*DataResponse[PromptTrackerTopic], error) {
	path, err := promptTrackerResourcePath(projectID, "topics", id, "ait")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if err := validatePromptTrackerInput(input); err != nil {
		return nil, err
	}
	config.body = input
	return requestJSON[DataResponse[PromptTrackerTopic]](c, ctx, "PATCH", path, config)
}

// ArchivePromptTrackerTopic calls the project-scoped Prompt Tracker contract.
func (c *Client) ArchivePromptTrackerTopic(ctx context.Context, projectID string, id string, options ...RequestOption) (*DataResponse[PromptTrackerTopic], error) {
	path, err := promptTrackerResourcePath(projectID, "topics", id, "ait")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	return requestJSON[DataResponse[PromptTrackerTopic]](c, ctx, "DELETE", path, config)
}

// ListPromptTrackerPrompts calls the project-scoped Prompt Tracker contract.
func (c *Client) ListPromptTrackerPrompts(ctx context.Context, projectID string, input *PaginationOptions, options ...RequestOption) (*ListResponse[PromptTrackerPrompt], error) {
	path, err := promptTrackerResourcePath(projectID, "prompts", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if input != nil {
		addQuery(config.query, "cursor", input.Cursor)
		addIntQuery(config.query, "limit", input.Limit)
	}
	return requestJSON[ListResponse[PromptTrackerPrompt]](c, ctx, "GET", path, config)
}

// CreatePromptTrackerPrompt calls the project-scoped Prompt Tracker contract.
func (c *Client) CreatePromptTrackerPrompt(ctx context.Context, projectID string, input PromptTrackerPromptInput, options ...RequestOption) (*DataResponse[PromptTrackerPrompt], error) {
	path, err := promptTrackerResourcePath(projectID, "prompts", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if err := validatePromptTrackerInput(input); err != nil {
		return nil, err
	}
	config.body = input
	return requestJSON[DataResponse[PromptTrackerPrompt]](c, ctx, "POST", path, config)
}

// UpdatePromptTrackerPrompt calls the project-scoped Prompt Tracker contract.
func (c *Client) UpdatePromptTrackerPrompt(ctx context.Context, projectID string, id string, input PromptTrackerPromptPatch, options ...RequestOption) (*DataResponse[PromptTrackerPrompt], error) {
	path, err := promptTrackerResourcePath(projectID, "prompts", id, "aip")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if err := validatePromptTrackerInput(input); err != nil {
		return nil, err
	}
	config.body = input
	return requestJSON[DataResponse[PromptTrackerPrompt]](c, ctx, "PATCH", path, config)
}

// ArchivePromptTrackerPrompt calls the project-scoped Prompt Tracker contract.
func (c *Client) ArchivePromptTrackerPrompt(ctx context.Context, projectID string, id string, options ...RequestOption) (*DataResponse[PromptTrackerPrompt], error) {
	path, err := promptTrackerResourcePath(projectID, "prompts", id, "aip")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	return requestJSON[DataResponse[PromptTrackerPrompt]](c, ctx, "DELETE", path, config)
}

// ListPromptTrackerSchedules calls the project-scoped Prompt Tracker contract.
func (c *Client) ListPromptTrackerSchedules(ctx context.Context, projectID string, input *PaginationOptions, options ...RequestOption) (*ListResponse[PromptTrackerSchedule], error) {
	path, err := promptTrackerResourcePath(projectID, "schedules", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if input != nil {
		addQuery(config.query, "cursor", input.Cursor)
		addIntQuery(config.query, "limit", input.Limit)
	}
	return requestJSON[ListResponse[PromptTrackerSchedule]](c, ctx, "GET", path, config)
}

// CreatePromptTrackerSchedule calls the project-scoped Prompt Tracker contract.
func (c *Client) CreatePromptTrackerSchedule(ctx context.Context, projectID string, input PromptTrackerScheduleInput, options ...RequestOption) (*DataResponse[PromptTrackerSchedule], error) {
	path, err := promptTrackerResourcePath(projectID, "schedules", "", "")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if err := validatePromptTrackerInput(input); err != nil {
		return nil, err
	}
	config.body = input
	return requestJSON[DataResponse[PromptTrackerSchedule]](c, ctx, "POST", path, config)
}

// UpdatePromptTrackerSchedule calls the project-scoped Prompt Tracker contract. The
// PATCH is a partial update: only fields the caller sets reach the server, and
// consent is sent at the top level because the API reads it from there to
// approve a reviewed configuration.
func (c *Client) UpdatePromptTrackerSchedule(ctx context.Context, projectID string, id string, input PromptTrackerSchedulePatch, options ...RequestOption) (*DataResponse[PromptTrackerSchedule], error) {
	path, err := promptTrackerResourcePath(projectID, "schedules", id, "ais")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	if err := validatePromptTrackerInput(input); err != nil {
		return nil, err
	}
	config.body = input
	return requestJSON[DataResponse[PromptTrackerSchedule]](c, ctx, "PATCH", path, config)
}

// ArchivePromptTrackerSchedule calls the project-scoped Prompt Tracker contract.
func (c *Client) ArchivePromptTrackerSchedule(ctx context.Context, projectID string, id string, options ...RequestOption) (*DataResponse[PromptTrackerSchedule], error) {
	path, err := promptTrackerResourcePath(projectID, "schedules", id, "ais")
	if err != nil {
		return nil, err
	}
	config := newRequestConfig(options...)
	return requestJSON[DataResponse[PromptTrackerSchedule]](c, ctx, "DELETE", path, config)
}
