package api

import (
	"github.com/OPGLOL/opgl-gateway-service/internal/middleware"
	"github.com/gorilla/mux"
)

// RouterConfig holds all dependencies for router setup
type RouterConfig struct {
	Handler         *Handler
	RateLimitClient *middleware.RateLimitServiceClient
}

// SetupRouter configures all routes for the gateway
func SetupRouter(config *RouterConfig) *mux.Router {
	router := mux.NewRouter()

	// Health check endpoint - no rate limiting
	router.HandleFunc("/health", config.Handler.HealthCheck).Methods("POST")

	// API routes subrouter
	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// Apply rate limiting middleware if configured
	if config.RateLimitClient != nil {
		apiRouter.Use(middleware.RateLimitMiddleware(config.RateLimitClient))
	}

	// Proxied data endpoints (rate limited)
	apiRouter.HandleFunc("/summoner", config.Handler.GetSummoner).Methods("POST")
	apiRouter.HandleFunc("/matches", config.Handler.GetMatches).Methods("POST")

	// Orchestrated analysis endpoint (rate limited)
	apiRouter.HandleFunc("/analyze", config.Handler.AnalyzePlayer).Methods("POST")

	return router
}

// SetupRouterSimple configures routes with minimal dependencies (for testing)
func SetupRouterSimple(handler *Handler, rateLimitClient *middleware.RateLimitServiceClient) *mux.Router {
	return SetupRouter(&RouterConfig{
		Handler:         handler,
		RateLimitClient: rateLimitClient,
	})
}
