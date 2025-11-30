package api

import (
	"encoding/json"
	"net/http"

	apierrors "github.com/OPGLOL/opgl-gateway-service/internal/errors"
	"github.com/OPGLOL/opgl-gateway-service/internal/models"
	"github.com/OPGLOL/opgl-gateway-service/internal/proxy"
)

// Handler manages HTTP request handlers for the gateway
type Handler struct {
	serviceProxy proxy.ServiceProxyInterface
}

// NewHandler creates a new Handler instance
func NewHandler(serviceProxy proxy.ServiceProxyInterface) *Handler {
	return &Handler{
		serviceProxy: serviceProxy,
	}
}

// HealthCheck handles health check requests
func (handler *Handler) HealthCheck(writer http.ResponseWriter, request *http.Request) {
	response := map[string]string{
		"status":  "healthy",
		"service": "opgl-gateway",
	}
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(response)
}

// GetSummoner proxies summoner requests to opgl-data service using Riot ID
func (handler *Handler) GetSummoner(writer http.ResponseWriter, request *http.Request) {
	var summonerRequest struct {
		Region   string `json:"region"`
		GameName string `json:"gameName"`
		TagLine  string `json:"tagLine"`
	}

	if err := json.NewDecoder(request.Body).Decode(&summonerRequest); err != nil {
		apierrors.WriteError(writer, apierrors.InvalidRequestBody("Invalid JSON format"))
		return
	}

	if summonerRequest.Region == "" || summonerRequest.GameName == "" || summonerRequest.TagLine == "" {
		apierrors.WriteError(writer, apierrors.MissingFields("region, gameName, and tagLine are required"))
		return
	}

	summoner, err := handler.serviceProxy.GetSummonerByRiotID(summonerRequest.Region, summonerRequest.GameName, summonerRequest.TagLine)
	if err != nil {
		// Check if the error is already an APIError
		if apiErr, ok := err.(*apierrors.APIError); ok {
			apierrors.WriteError(writer, apiErr)
			return
		}
		// Wrap unknown errors as internal errors
		apierrors.WriteError(writer, apierrors.InternalError("An unexpected error occurred"))
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(summoner)
}

// GetMatches proxies match history requests to opgl-data service
// Accepts either Riot ID (region, gameName, tagLine) or PUUID (region, puuid)
func (handler *Handler) GetMatches(writer http.ResponseWriter, request *http.Request) {
	var matchRequest struct {
		Region   string `json:"region"`
		GameName string `json:"gameName"`
		TagLine  string `json:"tagLine"`
		PUUID    string `json:"puuid"`
		Count    int    `json:"count"`
	}

	if err := json.NewDecoder(request.Body).Decode(&matchRequest); err != nil {
		apierrors.WriteError(writer, apierrors.InvalidRequestBody("Invalid JSON format"))
		return
	}

	// Set default count if not provided
	count := matchRequest.Count
	if count <= 0 {
		count = 20
	}

	var matches []models.Match
	var err error

	// Check if PUUID is provided for direct lookup
	if matchRequest.PUUID != "" {
		if matchRequest.Region == "" {
			apierrors.WriteError(writer, apierrors.MissingFields("region is required when using puuid"))
			return
		}
		matches, err = handler.serviceProxy.GetMatchesByPUUID(matchRequest.Region, matchRequest.PUUID, count)
	} else {
		// Use Riot ID lookup
		if matchRequest.Region == "" || matchRequest.GameName == "" || matchRequest.TagLine == "" {
			apierrors.WriteError(writer, apierrors.MissingFields("region, gameName, and tagLine are required (or use region and puuid)"))
			return
		}
		matches, err = handler.serviceProxy.GetMatchesByRiotID(matchRequest.Region, matchRequest.GameName, matchRequest.TagLine, count)
	}

	if err != nil {
		// Check if the error is already an APIError
		if apiErr, ok := err.(*apierrors.APIError); ok {
			apierrors.WriteError(writer, apiErr)
			return
		}
		// Wrap unknown errors as internal errors
		apierrors.WriteError(writer, apierrors.InternalError("An unexpected error occurred"))
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(matches)
}

// AnalyzePlayer orchestrates player analysis by calling both data and cortex services using Riot ID
func (handler *Handler) AnalyzePlayer(writer http.ResponseWriter, request *http.Request) {
	var analyzeRequest struct {
		Region   string `json:"region"`
		GameName string `json:"gameName"`
		TagLine  string `json:"tagLine"`
	}

	if err := json.NewDecoder(request.Body).Decode(&analyzeRequest); err != nil {
		apierrors.WriteError(writer, apierrors.InvalidRequestBody("Invalid JSON format"))
		return
	}

	if analyzeRequest.Region == "" || analyzeRequest.GameName == "" || analyzeRequest.TagLine == "" {
		apierrors.WriteError(writer, apierrors.MissingFields("region, gameName, and tagLine are required"))
		return
	}

	// Step 1: Get summoner data from opgl-data
	summoner, err := handler.serviceProxy.GetSummonerByRiotID(analyzeRequest.Region, analyzeRequest.GameName, analyzeRequest.TagLine)
	if err != nil {
		if apiErr, ok := err.(*apierrors.APIError); ok {
			apierrors.WriteError(writer, apiErr)
			return
		}
		apierrors.WriteError(writer, apierrors.InternalError("An unexpected error occurred"))
		return
	}

	// Step 2: Get match history from opgl-data (using internal method with PUUID)
	matches, err := handler.serviceProxy.GetMatchesByPUUID(analyzeRequest.Region, summoner.PUUID, 20)
	if err != nil {
		if apiErr, ok := err.(*apierrors.APIError); ok {
			apierrors.WriteError(writer, apiErr)
			return
		}
		apierrors.WriteError(writer, apierrors.InternalError("An unexpected error occurred"))
		return
	}

	// Step 3: Send data to opgl-cortex-engine for analysis
	analysisResult, err := handler.serviceProxy.AnalyzePlayer(summoner, matches)
	if err != nil {
		if apiErr, ok := err.(*apierrors.APIError); ok {
			apierrors.WriteError(writer, apiErr)
			return
		}
		apierrors.WriteError(writer, apierrors.InternalError("An unexpected error occurred"))
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(analysisResult)
}
