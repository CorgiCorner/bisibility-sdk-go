package bisibility

import (
	"context"
	"net/http"
	"net/url"
)

// GetMe returns the authenticated user and project memberships. Only
// personal access tokens (bsp_) may call this method.
func (c *Client) GetMe(ctx context.Context, options ...RequestOption) (*Me, error) {
	return requestJSON[Me](c, ctx, http.MethodGet, "/me", newRequestConfig(options...))
}

// UpdateMe patches the authenticated user's profile. Requires a personal
// access token with write tier.
func (c *Client) UpdateMe(ctx context.Context, input UpdateMeInput, options ...RequestOption) (*Me, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[Me](c, ctx, http.MethodPatch, "/me", config)
}

// ListMyTokens lists the authenticated user's personal access tokens.
// Requires a personal access token with admin tier.
func (c *Client) ListMyTokens(ctx context.Context, options ...RequestOption) (*ListResponse[PersonalAccessToken], error) {
	return requestJSON[ListResponse[PersonalAccessToken]](c, ctx, http.MethodGet, "/me/tokens", newRequestConfig(options...))
}

// CreateMyToken mints a personal access token. The raw bsp_ secret is only
// returned once in the Token field. Requires a personal access token with
// admin tier or an OAuth access token bearing the tokens:write scope.
func (c *Client) CreateMyToken(ctx context.Context, input CreateMyTokenInput, options ...RequestOption) (*CreatedPersonalAccessToken, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[CreatedPersonalAccessToken](c, ctx, http.MethodPost, "/me/tokens", config)
}

// RevokeMyToken revokes one personal access token and returns the revoked
// resource. Revoking by ID requires admin tier; pass "current" to revoke the
// token used for the request, which works at any tier.
func (c *Client) RevokeMyToken(ctx context.Context, tokenID string, options ...RequestOption) (*PersonalAccessToken, error) {
	return requestJSON[PersonalAccessToken](c, ctx, http.MethodDelete, "/me/tokens/"+url.PathEscape(tokenID), newRequestConfig(options...))
}

// CreateProject creates a project. Requires a personal access token with
// write tier; project-scoped API keys cannot create projects.
func (c *Client) CreateProject(ctx context.Context, input CreateProjectInput, options ...RequestOption) (*Project, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[Project](c, ctx, http.MethodPost, "/projects", config)
}

// ListProjectAPIKeys lists API keys for a project using the nested route.
func (c *Client) ListProjectAPIKeys(ctx context.Context, projectID string, pagination *PaginationOptions, options ...RequestOption) (*ListResponse[APIKey], error) {
	config := newRequestConfig(options...)
	addPagination(config.query, pagination)
	return requestJSON[ListResponse[APIKey]](c, ctx, http.MethodGet, projectResourcePath(projectID, "api-keys"), config)
}

// CreateProjectAPIKey mints a project API key using the nested route. The raw
// bsk_ secret is only returned once in the Token field.
func (c *Client) CreateProjectAPIKey(ctx context.Context, projectID string, input CreateAPIKeyInput, options ...RequestOption) (*CreatedAPIKey, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[CreatedAPIKey](c, ctx, http.MethodPost, projectResourcePath(projectID, "api-keys"), config)
}

// ListWebhooks lists webhooks for a project.
func (c *Client) ListWebhooks(ctx context.Context, projectID string, pagination *PaginationOptions, options ...RequestOption) (*ListResponse[Webhook], error) {
	config := newRequestConfig(options...)
	addPagination(config.query, pagination)
	return requestJSON[ListResponse[Webhook]](c, ctx, http.MethodGet, projectResourcePath(projectID, "webhooks"), config)
}

// CreateWebhook creates a webhook for a project.
func (c *Client) CreateWebhook(ctx context.Context, projectID string, input CreateWebhookInput, options ...RequestOption) (*Webhook, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[Webhook](c, ctx, http.MethodPost, projectResourcePath(projectID, "webhooks"), config)
}

// UpdateWebhook patches a webhook by project and webhook ID.
func (c *Client) UpdateWebhook(ctx context.Context, projectID, webhookID string, input UpdateWebhookInput, options ...RequestOption) (*Webhook, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[Webhook](c, ctx, http.MethodPatch, projectMemberResourcePath(projectID, "webhooks", webhookID), config)
}

// DeleteWebhook deletes a webhook by project and webhook ID and returns the
// deleted resource. Requires admin tier.
func (c *Client) DeleteWebhook(ctx context.Context, projectID, webhookID string, options ...RequestOption) (*Webhook, error) {
	return requestJSON[Webhook](c, ctx, http.MethodDelete, projectMemberResourcePath(projectID, "webhooks", webhookID), newRequestConfig(options...))
}
