package bisibility

import (
	"encoding/json"
	"fmt"
	"time"
)

// AnalyticsConsentSource describes where the account decision was recorded.
type AnalyticsConsentSource string

const (
	AnalyticsConsentSourceWeb AnalyticsConsentSource = "web"
	AnalyticsConsentSourceIOS AnalyticsConsentSource = "ios"
	AnalyticsConsentSourceAPI AnalyticsConsentSource = "api"
)

// AccountAnalyticsConsent is account analytics permission. A nil DecidedAt means pending.
// Browser replay permission is independent and is not represented here.
type AccountAnalyticsConsent struct {
	Analytics bool                    `json:"analytics"`
	DecidedAt *time.Time              `json:"decided_at"`
	Source    *AnalyticsConsentSource `json:"source"`
	Version   int                     `json:"version"`
}

// UnmarshalJSON preserves nullable pending fields without accepting missing permission evidence.
func (out *AccountAnalyticsConsent) UnmarshalJSON(data []byte) error {
	var raw struct {
		Analytics *bool           `json:"analytics"`
		DecidedAt json.RawMessage `json:"decided_at"`
		Source    json.RawMessage `json:"source"`
		Version   *int            `json:"version"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Analytics == nil || len(raw.DecidedAt) == 0 || len(raw.Source) == 0 || raw.Version == nil || *raw.Version < 1 {
		return fmt.Errorf("invalid required account analytics consent fields")
	}
	var value AccountAnalyticsConsent
	value.Analytics, value.Version = *raw.Analytics, *raw.Version
	if err := json.Unmarshal(raw.DecidedAt, &value.DecidedAt); err != nil {
		return err
	}
	if err := json.Unmarshal(raw.Source, &value.Source); err != nil {
		return err
	}
	if value.Source != nil && *value.Source != AnalyticsConsentSourceWeb && *value.Source != AnalyticsConsentSourceIOS && *value.Source != AnalyticsConsentSourceAPI {
		return fmt.Errorf("invalid account analytics consent source")
	}
	*out = value
	return nil
}

// UpdateAnalyticsConsentInput stores only account analytics; false is always serialized.
type UpdateAnalyticsConsentInput struct {
	Analytics bool `json:"analytics"`
}
