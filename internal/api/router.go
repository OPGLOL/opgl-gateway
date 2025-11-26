package api

import (
	"github.com/gorilla/mux"
)

// SetupRouter configures all routes for the gateway
func SetupRouter(handler *Handler) *mux.Router {
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", handler.HealthCheck).Methods("POST")

	// Proxied data endpoints
	router.HandleFunc("/api/v1/summoner", handler.GetSummoner).Methods("POST")
	router.HandleFunc("/api/v1/matches", handler.GetMatches).Methods("POST")

	// Orchestrated analysis endpoint
	router.HandleFunc("/api/v1/analyze", handler.AnalyzePlayer).Methods("POST")

	return router
}
