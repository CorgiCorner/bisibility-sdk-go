package bisibility

import "encoding/json"

// SavedViewReportView selects the keyword report presented by a version 2 view.
type SavedViewReportView string

const (
	SavedViewReportGeneral     SavedViewReportView = "general"
	SavedViewReportLocal       SavedViewReportView = "local"
	SavedViewReportAIOverviews SavedViewReportView = "ai_overviews"
)

// KeywordSavedViewFilters is the normalized filter shape in both definition versions.
// AIOverviewShown defaults to false when an older response omits it.
type KeywordSavedViewFilters struct {
	AIOverviewShown bool                      `json:"ai_overview_shown"`
	Change          string                    `json:"change"`
	Contains        string                    `json:"contains"`
	Intents         []string                  `json:"intents"`
	LastCheck       string                    `json:"last_check"`
	Position        []SavedViewPositionBucket `json:"position"`
	SERP            []string                  `json:"serp"`
	Tags            []string                  `json:"tags"`
	Topics          []string                  `json:"topics"`
	URLChanged      bool                      `json:"url_changed"`
	VolMax          float64                   `json:"vol_max"`
	VolMin          float64                   `json:"vol_min"`
	WrongURL        bool                      `json:"wrong_url"`
}

// SavedViewLens selects a keyword market and device.
type SavedViewLens struct {
	Device     string  `json:"device"`
	LocationID *string `json:"location_id"`
}

// SavedViewSort is a saved grid ordering.
type SavedViewSort struct {
	Direction string `json:"direction"`
	Field     string `json:"field"`
}

// SavedViewPresentation retains version 2 grid preferences.
type SavedViewPresentation struct {
	ColumnOrder     []string           `json:"column_order"`
	RegistryVersion int                `json:"registry_version"`
	Sizing          map[string]float64 `json:"sizing"`
	Version         int                `json:"version"`
	Visibility      map[string]bool    `json:"visibility"`
}

// SavedViewSubjectScope identifies a keyword or market report subject.
type SavedViewSubjectScope struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

// SavedViewReportSubject pins the report identity and assignment revision.
type SavedViewReportSubject struct {
	AssignmentRevision int                   `json:"assignment_revision"`
	IdentityRevisionID string                `json:"identity_revision_id"`
	LocationID         string                `json:"location_id"`
	Scope              SavedViewSubjectScope `json:"scope"`
}

// SavedViewReportScope retains a version 2 keyword report selection.
type SavedViewReportScope struct {
	AsOf           *string                 `json:"as_of"`
	ComparisonDays int                     `json:"comparison_days"`
	Subject        *SavedViewReportSubject `json:"subject"`
	View           SavedViewReportView     `json:"view"`
}

// KeywordSavedViewDefinition represents versions 1 and 2. Grid preferences and
// ReportScope are present in version 2; Version preserves the wire discriminator.
type KeywordSavedViewDefinition struct {
	Filters      KeywordSavedViewFilters `json:"filters"`
	Lens         SavedViewLens           `json:"lens"`
	Search       string                  `json:"search"`
	Surface      string                  `json:"surface"`
	Version      int                     `json:"version"`
	Grouped      *bool                   `json:"grouped,omitempty"`
	PageSize     int                     `json:"page_size,omitempty"`
	Presentation *SavedViewPresentation  `json:"presentation,omitempty"`
	ReportScope  *SavedViewReportScope   `json:"report_scope,omitempty"`
	Sort         *SavedViewSort          `json:"sort,omitempty"`
}

// CompetitorSavedViewFilters retains competitor grid filters.
type CompetitorSavedViewFilters struct {
	ExcludedKeywordIDs []string `json:"excluded_keyword_ids"`
	Position           string   `json:"position"`
	Tag                *string  `json:"tag"`
}

// CompetitorSavedViewScope selects the competitor market.
type CompetitorSavedViewScope struct {
	Device     string `json:"device"`
	Engine     string `json:"engine"`
	LocationID string `json:"location_id"`
}

// CompetitorSavedViewDefinition represents competitor definition versions 1 and 2.
type CompetitorSavedViewDefinition struct {
	Filters      CompetitorSavedViewFilters `json:"filters"`
	Scope        CompetitorSavedViewScope   `json:"scope"`
	Surface      string                     `json:"surface"`
	Version      int                        `json:"version"`
	Grouped      *bool                      `json:"grouped,omitempty"`
	PageSize     int                        `json:"page_size,omitempty"`
	Presentation *SavedViewPresentation     `json:"presentation,omitempty"`
	Sort         *SavedViewSort             `json:"sort,omitempty"`
}

// SavedViewDefinition is a surface-discriminated definition. Exactly one field
// is populated on decoding. Unknown surfaces remain readable as raw JSON.
type SavedViewDefinition struct {
	Keywords    *KeywordSavedViewDefinition
	Competitors *CompetitorSavedViewDefinition
	Unknown     json.RawMessage
}

// UnmarshalJSON decodes the definition for its saved-view surface.
func (d *SavedViewDefinition) UnmarshalJSON(b []byte) error {
	var tag struct {
		Surface string `json:"surface"`
	}
	if err := json.Unmarshal(b, &tag); err != nil {
		return err
	}
	*d = SavedViewDefinition{}
	switch tag.Surface {
	case "keywords":
		var value KeywordSavedViewDefinition
		if err := json.Unmarshal(b, &value); err != nil {
			return err
		}
		d.Keywords = &value
	case "competitors":
		var value CompetitorSavedViewDefinition
		if err := json.Unmarshal(b, &value); err != nil {
			return err
		}
		d.Competitors = &value
	default:
		d.Unknown = append(json.RawMessage(nil), b...)
	}
	return nil
}

// MarshalJSON emits the underlying surface definition without a wrapper.
func (d SavedViewDefinition) MarshalJSON() ([]byte, error) {
	if d.Keywords != nil {
		return json.Marshal(d.Keywords)
	}
	if d.Competitors != nil {
		return json.Marshal(d.Competitors)
	}
	if d.Unknown != nil {
		return json.Marshal(d.Unknown)
	}
	return []byte("null"), nil
}
