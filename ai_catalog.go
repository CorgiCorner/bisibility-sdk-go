package bisibility

import "context"

// GetAIResearchCatalog reads free capabilities without submitting paid work.
func (c *Client) GetAIResearchCatalog(ctx context.Context, projectID string, options ...RequestOption) (*DataResponse[AIResearchCatalog], error) {
	return requestJSON[DataResponse[AIResearchCatalog]](c, ctx, "GET", projectResourcePath(projectID, "ai-catalog"), newRequestConfig(options...))
}

type AINamedChoice struct {
	Code string `json:"code"`
	Name string `json:"name"`
}
type AIModelCapability struct {
	ID                string `json:"id"`
	Provider          string `json:"provider"`
	Label             string `json:"label"`
	Reasoning         bool   `json:"reasoning"`
	WebSearch         bool   `json:"web_search"`
	MinOutputTokens   int    `json:"min_output_tokens"`
	MaxOutputTokens   int    `json:"max_output_tokens"`
	PriceAvailable    bool   `json:"price_available"`
	AdmissionEnabled  bool   `json:"admission_enabled"`
	ActualCostEnabled bool   `json:"actual_cost_enabled"`
}
type AIVisibilityMarket struct {
	Platform     string          `json:"platform"`
	LocationCode int             `json:"location_code"`
	CountryName  string          `json:"country_name"`
	Languages    []AINamedChoice `json:"languages"`
}
type AICatalogLimits struct {
	MaxModels       int `json:"max_models"`
	MaxOutputTokens int `json:"max_output_tokens"`
}
type AIResearchCatalog struct {
	Models              []AIModelCapability  `json:"models"`
	VisibilityMarkets   []AIVisibilityMarket `json:"visibility_markets"`
	ResponseCountries   []AINamedChoice      `json:"response_countries"`
	ResponseLanguages   []AINamedChoice      `json:"response_languages"`
	Limits              AICatalogLimits      `json:"limits"`
	FetchedAt           string               `json:"fetched_at"`
	ActualCostAvailable *bool                `json:"actual_cost_available"`
}
