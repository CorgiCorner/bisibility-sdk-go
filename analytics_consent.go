package bisibility

import (
	"context"
	"net/http"
)

// GetAnalyticsConsent reads account analytics permission, independently of browser replay.
func (c *Client) GetAnalyticsConsent(ctx context.Context, options ...RequestOption) (*AccountAnalyticsConsent, error) {
	return requestJSON[AccountAnalyticsConsent](c, ctx, http.MethodGet, "/me/analytics-consent", newRequestConfig(options...))
}

// UpdateAnalyticsConsent stores an explicit idempotent account analytics decision.
func (c *Client) UpdateAnalyticsConsent(ctx context.Context, input UpdateAnalyticsConsentInput, options ...RequestOption) (*AccountAnalyticsConsent, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[AccountAnalyticsConsent](c, ctx, http.MethodPut, "/me/analytics-consent", config)
}
