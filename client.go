package bisibility

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL    = "https://bisibility.com/api/v1"
	projectsPath      = "/projects"
	projectPathRoot   = projectsPath + "/"
	keywordsPath      = "/keywords"
	keywordPathRoot   = keywordsPath + "/"
	contentTypeHeader = "Content-Type"
)

// Version is the SDK version reported in the User-Agent header.
const Version = "0.2.0"

const userAgent = "bisibility-sdk-go/" + Version

// defaultHTTPTimeout bounds requests made with the default HTTP client.
const defaultHTTPTimeout = 30 * time.Second

// Client calls the Bisibility REST API.
type Client struct {
	apiKey     string
	baseURL    *url.URL
	headers    http.Header
	httpClient *http.Client
	maxRetries int
}

// WithMaxRetries sets retries after the initial attempt. The default is 2.
func WithMaxRetries(maxRetries int) Option {
	return func(c *Client) error {
		if maxRetries < 0 {
			return &ConfigurationError{Message: "maxRetries cannot be negative."}
		}
		c.maxRetries = maxRetries
		return nil
	}
}

// Option configures a Client.
type Option func(*Client) error

// WithAPIKey configures the bearer API key used for protected API methods.
func WithAPIKey(apiKey string) Option {
	return func(c *Client) error {
		c.apiKey = apiKey
		return nil
	}
}

// WithBaseURL configures the API v1 root URL.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) error {
		parsed, err := parseBaseURL(baseURL)
		if err != nil {
			return err
		}
		c.baseURL = parsed
		return nil
	}
}

// WithHTTPClient configures the HTTP client. The default is an *http.Client
// with a 30 second timeout; supply your own client to change the timeout,
// transport, or proxy behavior.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) error {
		if httpClient == nil {
			return &ConfigurationError{Message: "httpClient cannot be nil."}
		}
		c.httpClient = httpClient
		return nil
	}
}

// WithDefaultHeader configures a header sent with every request.
func WithDefaultHeader(key, value string) Option {
	return func(c *Client) error {
		c.headers.Set(key, value)
		return nil
	}
}

// projectHeader targets a project on personal access token routes that carry
// no project in the path.
const projectHeader = "X-Bisibility-Project"

// WithProjectID configures the project targeted by personal access token
// requests on routes without a project in the path. The project id or public
// id is sent as the X-Bisibility-Project header with every request; override
// it per request with WithRequestHeader.
func WithProjectID(projectID string) Option {
	return func(c *Client) error {
		if strings.TrimSpace(projectID) == "" {
			return &ConfigurationError{Message: "projectID cannot be empty."}
		}
		c.headers.Set(projectHeader, projectID)
		return nil
	}
}

// NewClient creates a Bisibility API client.
func NewClient(options ...Option) (*Client, error) {
	baseURL, err := parseBaseURL(defaultBaseURL)
	if err != nil {
		return nil, err
	}

	client := &Client{
		baseURL:    baseURL,
		headers:    make(http.Header),
		httpClient: &http.Client{Timeout: defaultHTTPTimeout},
		maxRetries: 2,
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}

// BaseURL returns the configured API v1 root URL.
func (c *Client) BaseURL() string {
	return c.baseURL.String()
}

// RequestOption configures an individual API request.
type RequestOption func(*requestConfig)

// WithIdempotencyKey sets the Idempotency-Key header for a write request.
func WithIdempotencyKey(key string) RequestOption {
	return func(config *requestConfig) {
		config.idempotencyKey = key
	}
}

// WithRequestHeader sets one header for an individual API request.
func WithRequestHeader(key, value string) RequestOption {
	return func(config *requestConfig) {
		config.headers.Set(key, value)
	}
}

type requestConfig struct {
	auth           bool
	body           any
	rawBody        io.Reader
	headers        http.Header
	idempotencyKey string
	migrationToken string
	query          url.Values
}

func newRequestConfig(options ...RequestOption) requestConfig {
	config := requestConfig{
		auth:    true,
		headers: make(http.Header),
		query:   make(url.Values),
	}
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}

	return config
}

func parseBaseURL(raw string) (*url.URL, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(raw), "/")
	if baseURL == "" {
		return nil, &ConfigurationError{Message: "baseURL cannot be empty."}
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, &ConfigurationError{Message: fmt.Sprintf("baseURL is invalid: %v.", err)}
	}
	if !parsed.IsAbs() || parsed.Host == "" {
		return nil, &ConfigurationError{Message: "baseURL must be an absolute URL."}
	}

	return parsed, nil
}

// GetHealth returns public API health information.
func (c *Client) GetHealth(ctx context.Context, options ...RequestOption) (*HealthResponse, error) {
	config := newRequestConfig(options...)
	config.auth = false
	return requestJSON[HealthResponse](c, ctx, http.MethodGet, "/health", config)
}

// GetOpenAPI returns the public OpenAPI document.
func (c *Client) GetOpenAPI(ctx context.Context, options ...RequestOption) (*OpenAPIDocument, error) {
	config := newRequestConfig(options...)
	config.auth = false
	return requestJSON[OpenAPIDocument](c, ctx, http.MethodGet, "/openapi.json", config)
}

// GetCapabilities returns the public capabilities envelope.
func (c *Client) GetCapabilities(ctx context.Context, options ...RequestOption) (*DataResponse[[]Capability], error) {
	config := newRequestConfig(options...)
	config.auth = false
	return requestJSON[DataResponse[[]Capability]](c, ctx, http.MethodGet, "/capabilities", config)
}

// GetLLMSText returns the public llms.txt document.
func (c *Client) GetLLMSText(ctx context.Context, options ...RequestOption) (string, error) {
	config := newRequestConfig(options...)
	config.auth = false
	return requestText(c, ctx, http.MethodGet, "/llms.txt", config)
}

// GetProviderRates returns the public SERP provider rate cards. No API key
// is required. The request is sent anonymously: any Authorization header set
// via WithDefaultHeader or WithRequestHeader is stripped before sending.
func (c *Client) GetProviderRates(ctx context.Context, options ...RequestOption) (*DataResponse[[]ProviderRate], error) {
	config := newRequestConfig(options...)
	config.auth = false
	return requestJSON[DataResponse[[]ProviderRate]](c, ctx, http.MethodGet, "/provider-rates", config)
}

// GetCostEstimate estimates the monthly SERP provider cost for a keyword
// portfolio using public rate cards. No API key is required. The request is
// sent anonymously: any Authorization header set via WithDefaultHeader or
// WithRequestHeader is stripped before sending. Keywords is always sent;
// zero-valued optional fields fall back to server defaults.
func (c *Client) GetCostEstimate(ctx context.Context, input CostEstimateOptions, options ...RequestOption) (*DataResponse[CostEstimate], error) {
	config := newRequestConfig(options...)
	config.auth = false
	config.query.Set("keywords", strconv.Itoa(input.Keywords))
	addIntQuery(config.query, "devices", input.Devices)
	addQuery(config.query, "frequency", string(input.Frequency))
	addIntQuery(config.query, "locations", input.Locations)
	addQuery(config.query, "option", input.Option)
	addQuery(config.query, "plan", input.Plan)
	addQuery(config.query, "provider", string(input.Provider))
	return requestJSON[DataResponse[CostEstimate]](c, ctx, http.MethodGet, "/cost-estimate", config)
}

// ListProjects lists projects visible to the configured API key.
func (c *Client) ListProjects(ctx context.Context, options ...RequestOption) (*ListResponse[Project], error) {
	return requestJSON[ListResponse[Project]](c, ctx, http.MethodGet, projectsPath, newRequestConfig(options...))
}

// Projects is an operation-style alias for ListProjects.
func (c *Client) Projects(ctx context.Context, options ...RequestOption) (*ListResponse[Project], error) {
	return c.ListProjects(ctx, options...)
}

// GetProject fetches one project by public project ID.
func (c *Client) GetProject(ctx context.Context, projectID string, options ...RequestOption) (*Project, error) {
	return requestJSON[Project](c, ctx, http.MethodGet, projectPathRoot+url.PathEscape(projectID), newRequestConfig(options...))
}

// UpdateProject patches project name or domain. At least one input field is required.
func (c *Client) UpdateProject(ctx context.Context, projectID string, input UpdateProjectInput, options ...RequestOption) (*Project, error) {
	if input.Domain == nil && input.Name == nil {
		return nil, &ConfigurationError{Message: "UpdateProject requires at least one of Domain or Name."}
	}

	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[Project](c, ctx, http.MethodPatch, projectPathRoot+url.PathEscape(projectID), config)
}

// DeleteProject deletes one project and returns the deleted resource.
func (c *Client) DeleteProject(ctx context.Context, projectID string, options ...RequestOption) (*Project, error) {
	return requestJSON[Project](c, ctx, http.MethodDelete, projectPathRoot+url.PathEscape(projectID), newRequestConfig(options...))
}

// UpdateProjectDefaults patches project default market and schedule settings.
func (c *Client) UpdateProjectDefaults(ctx context.Context, projectID string, input ProjectDefaultsPatch, options ...RequestOption) (*ProjectDefaults, error) {
	if input.Frequency == "" {
		return nil, &ConfigurationError{Message: "UpdateProjectDefaults requires Frequency."}
	}
	if input.LocationKey == "" && (input.Country == "") != (input.Device == "") {
		return nil, &ConfigurationError{Message: "UpdateProjectDefaults requires Country and Device together unless LocationKey is set."}
	}

	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[ProjectDefaults](c, ctx, http.MethodPatch, projectPathRoot+url.PathEscape(projectID)+"/defaults", config)
}

// ListAPIKeys lists API keys for the configured API key's project.
func (c *Client) ListAPIKeys(ctx context.Context, pagination *PaginationOptions, options ...RequestOption) (*ListResponse[APIKey], error) {
	config := newRequestConfig(options...)
	addPagination(config.query, pagination)
	return requestJSON[ListResponse[APIKey]](c, ctx, http.MethodGet, "/api-keys", config)
}

// CreateAPIKey creates an API key for the configured API key's project.
func (c *Client) CreateAPIKey(ctx context.Context, input CreateAPIKeyInput, options ...RequestOption) (*CreatedAPIKey, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[CreatedAPIKey](c, ctx, http.MethodPost, "/api-keys", config)
}

// RevokeAPIKey revokes one API key.
func (c *Client) RevokeAPIKey(ctx context.Context, keyID string, options ...RequestOption) (*APIKey, error) {
	return requestJSON[APIKey](c, ctx, http.MethodDelete, "/api-keys/"+url.PathEscape(keyID), newRequestConfig(options...))
}

// ListKeywords lists keywords for a project.
func (c *Client) ListKeywords(ctx context.Context, projectID string, filters *ListKeywordsOptions, options ...RequestOption) (*ListResponse[Keyword], error) {
	config := newRequestConfig(options...)
	if filters != nil {
		addPagination(config.query, &PaginationOptions{Cursor: filters.Cursor, Limit: filters.Limit})
		addQuery(config.query, "filter[country]", filters.Country)
		addQuery(config.query, "filter[device]", string(filters.Device))
		addQuery(config.query, "filter[intent]", filters.Intent)
		addIntQuery(config.query, "filter[position_gt]", filters.PositionGT)
		addIntQuery(config.query, "filter[position_lt]", filters.PositionLT)
		addQuery(config.query, "filter[tag]", filters.Tag)
		addQuery(config.query, "filter[topic]", filters.Topic)
		addQuery(config.query, "search", filters.Search)
		addQuery(config.query, "sort", filters.Sort)
	}

	return requestJSON[ListResponse[Keyword]](c, ctx, http.MethodGet, projectPathRoot+url.PathEscape(projectID)+keywordsPath, config)
}

// KeywordsList is an operation-style alias for ListKeywords.
func (c *Client) KeywordsList(ctx context.Context, projectID string, filters *ListKeywordsOptions, options ...RequestOption) (*ListResponse[Keyword], error) {
	return c.ListKeywords(ctx, projectID, filters, options...)
}

// AddKeywords posts an array of keyword creation items.
func (c *Client) AddKeywords(ctx context.Context, projectID string, keywords []CreateKeywordInput, options ...RequestOption) (*CreateKeywordsResponse, error) {
	config := newRequestConfig(options...)
	config.body = keywords
	return requestJSON[CreateKeywordsResponse](c, ctx, http.MethodPost, projectPathRoot+url.PathEscape(projectID)+keywordsPath, config)
}

// CreateKeywords creates one or more keywords using the wrapped API body shape.
func (c *Client) CreateKeywords(ctx context.Context, projectID string, input CreateKeywordsInput, options ...RequestOption) (*CreateKeywordsResponse, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[CreateKeywordsResponse](c, ctx, http.MethodPost, projectPathRoot+url.PathEscape(projectID)+keywordsPath, config)
}

// KeywordsCreate is an operation-style alias for CreateKeywords.
func (c *Client) KeywordsCreate(ctx context.Context, projectID string, input CreateKeywordsInput, options ...RequestOption) (*CreateKeywordsResponse, error) {
	return c.CreateKeywords(ctx, projectID, input, options...)
}

// GetKeyword fetches one keyword by public keyword ID.
func (c *Client) GetKeyword(ctx context.Context, keywordID string, options ...RequestOption) (*Keyword, error) {
	return requestJSON[Keyword](c, ctx, http.MethodGet, keywordPathRoot+url.PathEscape(keywordID), newRequestConfig(options...))
}

// UpdateKeyword patches keyword metadata.
func (c *Client) UpdateKeyword(ctx context.Context, keywordID string, input UpdateKeywordInput, options ...RequestOption) (*Keyword, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[Keyword](c, ctx, http.MethodPatch, keywordPathRoot+url.PathEscape(keywordID), config)
}

// SetKeywordTargetURL sets or clears a keyword target URL. Pass nil to clear it.
func (c *Client) SetKeywordTargetURL(ctx context.Context, keywordID string, targetURL *string, options ...RequestOption) (*Keyword, error) {
	input := UpdateKeywordInput{TargetURL: NullString()}
	if targetURL != nil {
		input.TargetURL = StringValue(*targetURL)
	}
	return c.UpdateKeyword(ctx, keywordID, input, options...)
}

// DeleteKeyword deletes one keyword and returns the deleted resource when the API sends one.
func (c *Client) DeleteKeyword(ctx context.Context, keywordID string, options ...RequestOption) (*Keyword, error) {
	return requestJSON[Keyword](c, ctx, http.MethodDelete, keywordPathRoot+url.PathEscape(keywordID), newRequestConfig(options...))
}

// BulkUpdateKeywords mutates many keywords.
func (c *Client) BulkUpdateKeywords(ctx context.Context, input KeywordBulkInput, options ...RequestOption) (*KeywordBulkResponse, error) {
	config := newRequestConfig(options...)
	config.body = input
	return requestJSON[KeywordBulkResponse](c, ctx, http.MethodPost, "/keywords/bulk", config)
}

// ListRankChecks lists rank checks for a keyword.
func (c *Client) ListRankChecks(ctx context.Context, keywordID string, filters *ListRankChecksOptions, options ...RequestOption) (*ListResponse[RankCheck], error) {
	config := newRequestConfig(options...)
	if filters != nil {
		addPagination(config.query, &PaginationOptions{Cursor: filters.Cursor, Limit: filters.Limit})
		if !filters.Since.IsZero() {
			config.query.Set("since", filters.Since.Format(time.RFC3339Nano))
		}
		addQuery(config.query, "status", string(filters.Status))
		if !filters.Until.IsZero() {
			config.query.Set("until", filters.Until.Format(time.RFC3339Nano))
		}
	}

	return requestJSON[ListResponse[RankCheck]](c, ctx, http.MethodGet, keywordPathRoot+url.PathEscape(keywordID)+"/rank-checks", config)
}

// RankHistory is an operation-style alias for ListRankChecks.
func (c *Client) RankHistory(ctx context.Context, keywordID string, filters *ListRankChecksOptions, options ...RequestOption) (*ListResponse[RankCheck], error) {
	return c.ListRankChecks(ctx, keywordID, filters, options...)
}

// RunRankCheck runs an immediate rank check for one keyword. When
// input.Async is true the check is enqueued with ?async=true and the API
// responds 202 with a RankCheck in status running.
func (c *Client) RunRankCheck(ctx context.Context, keywordID string, input *RunRankCheckInput, options ...RequestOption) (*RankCheck, error) {
	config := newRequestConfig(options...)
	if input != nil && input.ProviderID != "" {
		config.body = input
	}
	if input != nil && input.Async {
		config.query.Set("async", "true")
	}

	return requestJSON[RankCheck](c, ctx, http.MethodPost, keywordPathRoot+url.PathEscape(keywordID)+"/checks", config)
}

// RunCheck is an operation-style alias for RunRankCheck.
func (c *Client) RunCheck(ctx context.Context, keywordID string, input *RunRankCheckInput, options ...RequestOption) (*RankCheck, error) {
	return c.RunRankCheck(ctx, keywordID, input, options...)
}

// GetRankCheckResult fetches one rank check by ID.
func (c *Client) GetRankCheckResult(ctx context.Context, checkID string, options ...RequestOption) (*RankCheck, error) {
	return requestJSON[RankCheck](c, ctx, http.MethodGet, "/rank-checks/"+url.PathEscape(checkID), newRequestConfig(options...))
}

// CreateAPIKeyInput creates an API key.
type CreateAPIKeyInput struct {
	Name string `json:"name"`
}

func addPagination(query url.Values, pagination *PaginationOptions) {
	if pagination == nil {
		return
	}
	addQuery(query, "cursor", pagination.Cursor)
	addIntQuery(query, "limit", pagination.Limit)
}

func addQuery(query url.Values, key, value string) {
	if value != "" {
		query.Set(key, value)
	}
}

func addIntQuery(query url.Values, key string, value int) {
	if value != 0 {
		query.Set(key, fmt.Sprintf("%d", value))
	}
}

func requestJSON[T any](c *Client, ctx context.Context, method, path string, config requestConfig) (*T, error) {
	body, statusCode, err := c.do(ctx, method, path, config)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, nil
	}

	var out T
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, &ResponseError{
			Body:       string(body),
			Cause:      err,
			Method:     method,
			StatusCode: statusCode,
			URL:        mustBuildURL(c, path, config.query),
		}
	}

	return &out, nil
}

func requestText(c *Client, ctx context.Context, method, path string, config requestConfig) (string, error) {
	body, _, err := c.do(ctx, method, path, config)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (c *Client) do(ctx context.Context, method, path string, config requestConfig) ([]byte, int, error) {
	if ctx == nil {
		return nil, 0, &ConfigurationError{Message: "context cannot be nil."}
	}
	if config.auth && c.apiKey == "" {
		return nil, 0, &ConfigurationError{Message: "apiKey is required for this Bisibility API method."}
	}

	for attempt := 0; ; attempt++ {
		req, requestURL, err := c.newRequest(ctx, method, path, config)
		if err != nil {
			return nil, 0, err
		}
		retryable := isIdempotentRequest(req)
		resp, err := c.httpClient.Do(req)
		if err != nil {
			networkErr := &NetworkError{Cause: err, Method: method, URL: requestURL}
			if !retryable || attempt >= c.maxRetries || ctx.Err() != nil {
				return nil, 0, networkErr
			}
			if err := waitForRetry(ctx, retryBackoff(attempt)); err != nil {
				return nil, 0, &NetworkError{Cause: err, Method: method, URL: requestURL}
			}
			continue
		}

		body, statusCode, responseErr := readResponse(resp, method, requestURL)
		resp.Body.Close()
		if responseErr == nil {
			return body, statusCode, nil
		}
		var networkErr *NetworkError
		if retryable && attempt < c.maxRetries && errors.As(responseErr, &networkErr) && ctx.Err() == nil {
			if err := waitForRetry(ctx, retryBackoff(attempt)); err != nil {
				return nil, 0, &NetworkError{Cause: err, Method: method, URL: requestURL}
			}
			continue
		}
		var apiErr *APIError
		if !retryable || attempt >= c.maxRetries || !errors.As(responseErr, &apiErr) || (apiErr.StatusCode != http.StatusTooManyRequests && apiErr.StatusCode != http.StatusServiceUnavailable) {
			return body, statusCode, responseErr
		}
		delay := retryBackoff(attempt)
		if retryAfter, ok := apiErr.RetryAfter(); ok {
			delay = retryAfter
		}
		if err := waitForRetry(ctx, delay); err != nil {
			return nil, 0, &NetworkError{Cause: err, Method: method, URL: requestURL}
		}
	}
}

func isIdempotentRequest(req *http.Request) bool {
	switch req.Method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete:
		return true
	default:
		return req.Header.Get("Idempotency-Key") != ""
	}
}

func retryBackoff(attempt int) time.Duration {
	delay := 500 * time.Millisecond * time.Duration(1<<min(attempt, 4))
	if delay > 8*time.Second {
		return 8 * time.Second
	}
	return delay
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) newRequest(ctx context.Context, method, path string, config requestConfig) (*http.Request, string, error) {
	body, err := encodeBody(config.body, config.rawBody)
	if err != nil {
		return nil, "", err
	}
	requestURL, err := c.buildURL(path, config.query)
	if err != nil {
		return nil, "", err
	}
	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return nil, "", err
	}

	c.applyHeaders(req, config, body != nil)
	return req, requestURL, nil
}

func (c *Client) applyHeaders(req *http.Request, config requestConfig, hasBody bool) {
	for key, values := range c.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	for key, values := range config.headers {
		req.Header.Del(key)
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", userAgent)
	}
	req.Header.Set("X-Bisibility-Client", userAgent)
	switch {
	case config.migrationToken != "":
		// Cloud-import writes authenticate with a migration token rather than
		// the client API key. Any inherited Authorization header is replaced.
		req.Header.Set("Authorization", "Bearer "+config.migrationToken)
	case config.auth:
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	default:
		// Anonymous public endpoints must never send credentials, even when
		// the caller set an Authorization header via WithDefaultHeader or
		// WithRequestHeader.
		req.Header.Del("Authorization")
	}
	if config.idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", config.idempotencyKey)
	}
	if hasBody && req.Header.Get(contentTypeHeader) == "" {
		req.Header.Set(contentTypeHeader, "application/json")
	}
}

func readResponse(resp *http.Response, method, requestURL string) ([]byte, int, error) {
	responseBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, 0, &NetworkError{Cause: readErr, Method: method, URL: requestURL}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.StatusCode, apiErrorFromResponse(resp, method, requestURL, responseBody)
	}

	return responseBody, resp.StatusCode, nil
}

func encodeBody(body any, rawBody io.Reader) (io.Reader, error) {
	// A pre-serialized body (for example a gzip-compressed cloud-import chunk)
	// is sent verbatim and takes precedence over the JSON-encoded body.
	if rawBody != nil {
		return rawBody, nil
	}
	if body == nil {
		return nil, nil
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(payload), nil
}

func (c *Client) buildURL(routePath string, query url.Values) (string, error) {
	built := strings.TrimRight(c.baseURL.String(), "/") + "/" + strings.TrimLeft(routePath, "/")
	if encodedQuery := query.Encode(); encodedQuery != "" {
		built += "?" + encodedQuery
	}

	return built, nil
}

func apiErrorFromResponse(resp *http.Response, method, requestURL string, body []byte) error {
	apiErr := &APIError{
		Body:       string(body),
		Header:     resp.Header.Clone(),
		Method:     method,
		StatusCode: resp.StatusCode,
		URL:        requestURL,
	}
	redactResponseHeaders(apiErr.Header)
	if len(body) > 0 && strings.Contains(strings.ToLower(resp.Header.Get(contentTypeHeader)), "json") {
		var problem ProblemDetails
		if err := json.Unmarshal(body, &problem); err == nil && isProblemDetailsJSON(body) {
			apiErr.Problem = &problem
		}
	}

	return apiErr
}

// isProblemDetailsJSON accepts either a string type, or a string title plus a
// numeric status, as required by the cross-language SDK behavior spec.
func isProblemDetailsJSON(body []byte) bool {
	var fields map[string]json.RawMessage
	if json.Unmarshal(body, &fields) != nil {
		return false
	}
	var problemType string
	if json.Unmarshal(fields["type"], &problemType) == nil && problemType != "" {
		return true
	}
	var title string
	var status float64
	return json.Unmarshal(fields["title"], &title) == nil && title != "" && json.Unmarshal(fields["status"], &status) == nil
}

func redactResponseHeaders(header http.Header) {
	for _, name := range []string{"Authorization", "Cookie", "Proxy-Authorization", "Set-Cookie", "X-Api-Key", "X-Auth-Token"} {
		header.Del(name)
	}
}

func mustBuildURL(c *Client, path string, query url.Values) string {
	requestURL, err := c.buildURL(path, query)
	if err != nil {
		return ""
	}
	return requestURL
}

// IsAPIError reports whether err is an APIError.
func IsAPIError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr)
}
