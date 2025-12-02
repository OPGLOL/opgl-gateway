package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	apierrors "github.com/OPGLOL/opgl-gateway-service/internal/errors"
	"github.com/google/uuid"
)

// AuthServiceClient handles communication with the auth service
type AuthServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAuthServiceClient creates a new auth service client
func NewAuthServiceClient(baseURL string) *AuthServiceClient {
	return &AuthServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// validateTokenRequest represents the request to validate a token
type validateTokenRequest struct {
	Token string `json:"token"`
}

// validateTokenResponse represents the response from token validation
type validateTokenResponse struct {
	Valid  bool   `json:"valid"`
	UserID string `json:"userId,omitempty"`
	Email  string `json:"email,omitempty"`
}

// ValidateToken calls the auth service to validate a token
func (client *AuthServiceClient) ValidateToken(token string) (*validateTokenResponse, error) {
	requestBody := validateTokenRequest{Token: token}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	url := client.baseURL + "/api/v1/auth/validate"
	resp, err := client.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response validateTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// AuthMiddleware creates middleware that validates JWT access tokens via auth service
func AuthMiddleware(authClient *AuthServiceClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			// Extract Authorization header
			authHeader := request.Header.Get("Authorization")

			if authHeader == "" {
				apierrors.WriteError(responseWriter, apierrors.NewAPIError(
					apierrors.ErrCodeUnauthorized,
					"Authorization header is required",
					http.StatusUnauthorized,
				))
				return
			}

			// Check Bearer token format
			if !strings.HasPrefix(authHeader, "Bearer ") {
				apierrors.WriteError(responseWriter, apierrors.NewAPIError(
					apierrors.ErrCodeUnauthorized,
					"Invalid authorization format. Use: Bearer <token>",
					http.StatusUnauthorized,
				))
				return
			}

			// Extract token
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate token via auth service
			validationResult, err := authClient.ValidateToken(tokenString)
			if err != nil {
				apierrors.WriteError(responseWriter, apierrors.InternalError("Failed to validate token"))
				return
			}

			if !validationResult.Valid {
				apierrors.WriteError(responseWriter, apierrors.NewAPIError(
					apierrors.ErrCodeInvalidToken,
					"Invalid or expired access token",
					http.StatusUnauthorized,
				))
				return
			}

			// Parse user ID and add to context
			userID, err := uuid.Parse(validationResult.UserID)
			if err != nil {
				apierrors.WriteError(responseWriter, apierrors.InternalError("Invalid user ID in token"))
				return
			}

			// Add user ID to request context
			ctx := context.WithValue(request.Context(), "userID", userID)
			request = request.WithContext(ctx)

			// Proceed to next handler
			next.ServeHTTP(responseWriter, request)
		})
	}
}

// OptionalAuthMiddleware creates middleware that validates JWT tokens if present
// but allows requests without tokens to proceed
func OptionalAuthMiddleware(authClient *AuthServiceClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			// Extract Authorization header
			authHeader := request.Header.Get("Authorization")

			// If no auth header, proceed without user context
			if authHeader == "" {
				next.ServeHTTP(responseWriter, request)
				return
			}

			// Check Bearer token format
			if !strings.HasPrefix(authHeader, "Bearer ") {
				next.ServeHTTP(responseWriter, request)
				return
			}

			// Extract and validate token
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			validationResult, err := authClient.ValidateToken(tokenString)
			if err != nil || !validationResult.Valid {
				// Token invalid, proceed without user context
				next.ServeHTTP(responseWriter, request)
				return
			}

			// Parse user ID
			userID, err := uuid.Parse(validationResult.UserID)
			if err != nil {
				next.ServeHTTP(responseWriter, request)
				return
			}

			// Add user ID to request context
			ctx := context.WithValue(request.Context(), "userID", userID)
			request = request.WithContext(ctx)

			next.ServeHTTP(responseWriter, request)
		})
	}
}
