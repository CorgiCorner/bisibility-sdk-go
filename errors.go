package bisibility

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// BisibilityError is implemented by every SDK-defined error.
type BisibilityError interface {
	error
	bisibilityError()
}

// ConfigurationError reports invalid client configuration or missing credentials.
type ConfigurationError struct {
	Message string
}

func (e *ConfigurationError) Error() string {
	return e.Message
}
func (e *ConfigurationError) bisibilityError() {
	// This marker seals BisibilityError to SDK-defined error types.
}

// APIError reports a non-2xx HTTP response from the Bisibility API.
type APIError struct {
	Body       string
	Header     http.Header
	Method     string
	Problem    *ProblemDetails
	StatusCode int
	URL        string
}

func (e *APIError) Error() string {
	if e.Problem != nil && e.Problem.Detail != "" {
		return e.Problem.Detail
	}
	if e.Body != "" {
		return e.Body
	}

	return fmt.Sprintf("Bisibility API request failed with status %d.", e.StatusCode)
}

func (e *APIError) bisibilityError() {
	// This marker seals BisibilityError to SDK-defined error types.
}

// IsRateLimit reports whether the API returned HTTP 429.
func (e *APIError) IsRateLimit() bool { return e.StatusCode == http.StatusTooManyRequests }

// IsNotFound reports whether the API returned HTTP 404.
func (e *APIError) IsNotFound() bool { return e.StatusCode == http.StatusNotFound }

// RetryAfter parses Retry-After as delta seconds or an HTTP date, capped at 60 seconds.
func (e *APIError) RetryAfter() (time.Duration, bool) {
	value := e.Header.Get("Retry-After")
	if value == "" {
		return 0, false
	}
	if seconds, err := strconv.ParseFloat(value, 64); err == nil {
		if seconds < 0 {
			return 0, false
		}
		return min(time.Duration(seconds*float64(time.Second)), 60*time.Second), true
	}
	date, err := http.ParseTime(value)
	if err != nil {
		return 0, false
	}
	return min(max(time.Until(date), 0), 60*time.Second), true
}

// NetworkError wraps transport-level failures while calling the Bisibility API.
type NetworkError struct {
	Cause  error
	Method string
	URL    string
}

func (e *NetworkError) Error() string {
	return "network error while calling the Bisibility API"
}

func (e *NetworkError) Unwrap() error {
	return e.Cause
}
func (e *NetworkError) bisibilityError() {
	// This marker seals BisibilityError to SDK-defined error types.
}

// TimeoutError reports that a queued rank-check run did not finish before the deadline.
type TimeoutError struct {
	Message string
}

func (e *TimeoutError) Error() string {
	return e.Message
}

func (e *TimeoutError) bisibilityError() {
	// This marker seals BisibilityError to SDK-defined error types.
}

// ResponseError reports invalid successful API responses, such as malformed JSON.
type ResponseError struct {
	Body       string
	Cause      error
	Method     string
	StatusCode int
	URL        string
}

func (e *ResponseError) Error() string {
	return "Bisibility API returned invalid JSON"
}

func (e *ResponseError) Unwrap() error {
	return e.Cause
}
func (e *ResponseError) bisibilityError() {
	// This marker seals BisibilityError to SDK-defined error types.
}

// ProviderPrioritySyncError reports a priority PATCH failure after a provider connect succeeds.
type ProviderPrioritySyncError struct {
	Cause error
}

func (e *ProviderPrioritySyncError) Error() string {
	return "provider connected but priority synchronization failed"
}

func (e *ProviderPrioritySyncError) Unwrap() error {
	return e.Cause
}

func (e *ProviderPrioritySyncError) bisibilityError() {
	// This marker seals BisibilityError to SDK-defined error types.
}
