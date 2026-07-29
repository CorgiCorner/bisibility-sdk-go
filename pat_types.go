package bisibility

import "time"

// MeProject is one project membership returned by GetMe.
type MeProject struct {
	Domain string        `json:"domain"`
	ID     string        `json:"id"`
	Name   string        `json:"name"`
	Role   TeamRoleValue `json:"role"`
}

// Me is the authenticated personal access token user.
type Me struct {
	Email    string      `json:"email"`
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Projects []MeProject `json:"projects"`
}

// UpdateMeInput patches the authenticated user's profile.
type UpdateMeInput struct {
	Name string `json:"name"`
}

// TokenScope is the tier granted to a personal access token. The effective
// tier per project is the minimum of the token scope and the user's
// membership role.
type TokenScope string

const (
	TokenScopeRead  TokenScope = "read"
	TokenScopeWrite TokenScope = "write"
	TokenScopeAdmin TokenScope = "admin"
)

// PersonalAccessToken describes a personal access token without the raw secret.
type PersonalAccessToken struct {
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	ID         string     `json:"id"`
	LastUsedAt *time.Time `json:"last_used_at"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	RevokedAt  *time.Time `json:"revoked_at"`
	Scope      TokenScope `json:"scope"`
}

// CreatedPersonalAccessToken is returned once when minting a personal access
// token. Token carries the raw bsb_pat_live_ secret and is never shown again.
type CreatedPersonalAccessToken struct {
	PersonalAccessToken
	MaskedValue string `json:"masked_value"`
	Token       string `json:"token"`
}

// CreateMyTokenInput mints a personal access token. Scope defaults to
// TokenScopeRead when omitted. ExpiresInDays accepts 30, 90, or 365; leave it
// nil for a token that never expires.
type CreateMyTokenInput struct {
	ExpiresInDays *int       `json:"expires_in_days,omitempty"`
	Name          string     `json:"name"`
	Scope         TokenScope `json:"scope,omitempty"`
}

// TrackingScope selects whether a project tracks ranks at country or city level.
type TrackingScope string

const (
	TrackingScopeCountry TrackingScope = "country"
	TrackingScopeCity    TrackingScope = "city"
)

// CreateProjectInput creates a project. TrackingScope defaults to
// TrackingScopeCountry and Defaults falls back to server defaults when omitted.
type CreateProjectInput struct {
	Defaults      *ProjectDefaultsPatch `json:"defaults,omitempty"`
	Domain        string                `json:"domain"`
	Name          string                `json:"name"`
	TrackingScope TrackingScope         `json:"tracking_scope,omitempty"`
}

// Webhook is one project webhook. The HMAC secret is write-only and never
// returned by the API.
type Webhook struct {
	CreatedAt      time.Time  `json:"created_at"`
	Description    *string    `json:"description"`
	Enabled        bool       `json:"enabled"`
	ID             string     `json:"id"`
	LastDeliveryAt *time.Time `json:"last_delivery_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	URL            string     `json:"url"`
}

// CreateWebhookInput creates a project webhook. HMACSecret must be at least
// 16 characters; it is write-only and never returned by the API.
type CreateWebhookInput struct {
	Description string `json:"description,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
	HMACSecret  string `json:"hmac_secret"`
	URL         string `json:"url"`
}

// UpdateWebhookInput patches a project webhook. Omitted fields keep their
// current values.
type UpdateWebhookInput struct {
	Description *string `json:"description,omitempty"`
	Enabled     *bool   `json:"enabled,omitempty"`
	HMACSecret  string  `json:"hmac_secret,omitempty"`
	URL         string  `json:"url,omitempty"`
}
