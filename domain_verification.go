package bisibility

import (
	"context"
	"net/http"
	"time"
)

// DomainVerificationMethod identifies the proof mechanism. Response values remain open.
type DomainVerificationMethod string

const (
	DomainVerificationMethodDNSTXT        DomainVerificationMethod = "dns_txt"
	DomainVerificationMethodSearchConsole DomainVerificationMethod = "search_console"
)

// DomainVerification is the current proof and its verification history.
type DomainVerification struct {
	Domain          string                   `json:"domain"`
	Method          DomainVerificationMethod `json:"method"`
	DNSHost         string                   `json:"dns_host"`
	DNSValue        string                   `json:"dns_value"`
	VerifiedAt      *time.Time               `json:"verified_at"`
	LastCheckedAt   *time.Time               `json:"last_checked_at"`
	LastCheckResult *string                  `json:"last_check_result"`
	Failures        int                      `json:"failures"`
	RevokedAt       *time.Time               `json:"revoked_at"`
}

// DomainVerificationResponse has nil Data when no proof has been issued.
type DomainVerificationResponse struct {
	Data *DomainVerification `json:"data"`
}

// DomainVerificationInput chooses a supported proof mechanism.
type DomainVerificationInput struct {
	Method DomainVerificationMethod `json:"method"`
}

// GetDomainVerification reads the project's current domain proof.
func (c *Client) GetDomainVerification(ctx context.Context, projectID string, options ...RequestOption) (*DomainVerificationResponse, error) {
	return requestJSON[DomainVerificationResponse](c, ctx, http.MethodGet, projectResourcePath(projectID, "domain-verification"), newRequestConfig(options...))
}

// IssueDomainVerification issues a proof for a project domain.
func (c *Client) IssueDomainVerification(ctx context.Context, projectID string, input DomainVerificationInput, options ...RequestOption) (*DomainVerificationResponse, error) {
	if input.Method != DomainVerificationMethodDNSTXT && input.Method != DomainVerificationMethodSearchConsole {
		return nil, &ConfigurationError{Message: "method must be dns_txt or search_console."}
	}
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[DomainVerificationResponse](c, ctx, http.MethodPost, projectResourcePath(projectID, "domain-verification"), config)
}

// CheckDomainVerification checks the issued proof without replacing it.
func (c *Client) CheckDomainVerification(ctx context.Context, projectID string, options ...RequestOption) (*DomainVerificationResponse, error) {
	return requestJSON[DomainVerificationResponse](c, ctx, http.MethodPost, projectResourcePath(projectID, "domain-verification/check"), newRequestConfig(options...))
}
