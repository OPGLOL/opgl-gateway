package api

import (
	"github.com/gorilla/mux"
)

// SetupRouter configures all routes for the gateway
func SetupRouter(handler *Handler) *mux.Router {
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", handler.HealthCheck).Methods("GET")

	// Proxied data endpoints
	router.HandleFunc("/api/v1/summoner/{region}/{summonerName}", handler.GetSummoner).Methods("GET")
	router.HandleFunc("/api/v1/matches/{region}/{puuid}", handler.GetMatches).Methods("GET")

	// Orchestrated analysis endpoint
	router.HandleFunc("/api/v1/analyze/{region}/{summonerName}", handler.AnalyzePlayer).Methods("GET")

	return router
}
