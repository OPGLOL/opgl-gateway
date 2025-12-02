package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	apierrors "github.com/OPGLOL/opgl-gateway-service/internal/errors"
)

// RateLimitServiceClient handles communication with the auth service for rate limiting
type RateLimitServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewRateLimitServiceClient creates a new rate limit service client
func NewRateLimitServiceClient(baseURL string) *RateLimitServiceClient {
	return &RateLimitServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// checkRateLimitRequest represents the request to check rate limit
type checkRateLimitRequest struct {
	APIKey string `json:"apiKey"`
}

// checkRateLimitResponse represents the response from rate limit check
type checkRateLimitResponse struct {
	Allowed   bool  `json:"allowed"`
	Limit     int   `json:"limit"`
	Remaining int   `json:"remaining"`
	Reset     int64 `json:"reset"`
}

// CheckRateLimit calls the auth service to check rate limit
func (client *RateLimitServiceClient) CheckRateLimit(apiKey string) (*checkRateLimitResponse, error) {
	requestBody := checkRateLimitRequest{APIKey: apiKey}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	url := client.baseURL + "/api/v1/ratelimit/check"
	resp, err := client.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// If auth service returns non-200, API key is invalid
	if resp.StatusCode != http.StatusOK {
		return &checkRateLimitResponse{
			Allowed:   false,
			Limit:     0,
			Remaining: 0,
			Reset:     time.Now().Unix(),
		}, nil
	}

	var response checkRateLimitResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// RateLimitMiddleware creates middleware that enforces rate limiting via auth service
func RateLimitMiddleware(rateLimitClient *RateLimitServiceClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			// Extract API key from header
			apiKey := request.Header.Get("X-API-Key")

			// If no API key provided, reject the request
			if apiKey == "" {
				apierrors.WriteError(responseWriter, apierrors.NewAPIError(
					apierrors.ErrCodeMissingAPIKey,
					"API key is required. Include X-API-Key header in your request.",
					http.StatusUnauthorized,
				))
				return
			}

			// Check rate limit via auth service
			rateLimitResult, err := rateLimitClient.CheckRateLimit(apiKey)
			if err != nil {
				apierrors.WriteError(responseWriter, apierrors.InternalError("Rate limit check failed"))
				return
			}

			// Add rate limit headers to response
			responseWriter.Header().Set("X-RateLimit-Limit", strconv.Itoa(rateLimitResult.Limit))
			responseWriter.Header().Set("X-RateLimit-Remaining", strconv.Itoa(rateLimitResult.Remaining))
			responseWriter.Header().Set("X-RateLimit-Reset", strconv.FormatInt(rateLimitResult.Reset, 10))

			// If API key is invalid (Limit is 0), reject
			if rateLimitResult.Limit == 0 {
				apierrors.WriteError(responseWriter, apierrors.NewAPIError(
					apierrors.ErrCodeInvalidAPIKey,
					"Invalid or inactive API key.",
					http.StatusUnauthorized,
				))
				return
			}

			// If rate limit exceeded, reject with 429
			if !rateLimitResult.Allowed {
				retryAfter := rateLimitResult.Reset - time.Now().Unix()
				if retryAfter < 0 {
					retryAfter = 1
				}
				responseWriter.Header().Set("Retry-After", strconv.FormatInt(retryAfter, 10))

				apierrors.WriteError(responseWriter, apierrors.NewAPIError(
					apierrors.ErrCodeRateLimitExceeded,
					fmt.Sprintf("Rate limit exceeded. Try again in %d seconds.", retryAfter),
					http.StatusTooManyRequests,
				))
				return
			}

			// Request allowed, proceed to next handler
			next.ServeHTTP(responseWriter, request)
		})
	}
}

// OptionalRateLimitMiddleware creates middleware that enforces rate limiting only if API key is provided
func OptionalRateLimitMiddleware(rateLimitClient *RateLimitServiceClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			// Extract API key from header
			apiKey := request.Header.Get("X-API-Key")

			// If no API key provided, allow request without rate limiting
			if apiKey == "" {
				next.ServeHTTP(responseWriter, request)
				return
			}

			// Check rate limit via auth service
			rateLimitResult, err := rateLimitClient.CheckRateLimit(apiKey)
			if err != nil {
				apierrors.WriteError(responseWriter, apierrors.InternalError("Rate limit check failed"))
				return
			}

			// Add rate limit headers to response
			responseWriter.Header().Set("X-RateLimit-Limit", strconv.Itoa(rateLimitResult.Limit))
			responseWriter.Header().Set("X-RateLimit-Remaining", strconv.Itoa(rateLimitResult.Remaining))
			responseWriter.Header().Set("X-RateLimit-Reset", strconv.FormatInt(rateLimitResult.Reset, 10))

			// If API key is invalid, reject
			if rateLimitResult.Limit == 0 {
				apierrors.WriteError(responseWriter, apierrors.NewAPIError(
					apierrors.ErrCodeInvalidAPIKey,
					"Invalid or inactive API key.",
					http.StatusUnauthorized,
				))
				return
			}

			// If rate limit exceeded, reject with 429
			if !rateLimitResult.Allowed {
				responseWriter.Header().Set("Retry-After", strconv.FormatInt(rateLimitResult.Reset, 10))
				apierrors.WriteError(responseWriter, apierrors.NewAPIError(
					apierrors.ErrCodeRateLimitExceeded,
					"Rate limit exceeded.",
					http.StatusTooManyRequests,
				))
				return
			}

			next.ServeHTTP(responseWriter, request)
		})
	}
}
