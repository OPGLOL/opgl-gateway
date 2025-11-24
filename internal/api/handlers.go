package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/OPGLOL/opgl-gateway/internal/proxy"
)

// Handler manages HTTP request handlers for the gateway
type Handler struct {
	serviceProxy *proxy.ServiceProxy
}

// NewHandler creates a new Handler instance
func NewHandler(serviceProxy *proxy.ServiceProxy) *Handler {
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

// GetSummoner proxies summoner requests to opgl-data service
func (handler *Handler) GetSummoner(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	region := vars["region"]
	summonerName := vars["summonerName"]

	summoner, err := handler.serviceProxy.GetSummoner(region, summonerName)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(summoner)
}

// GetMatches proxies match history requests to opgl-data service
func (handler *Handler) GetMatches(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	region := vars["region"]
	puuid := vars["puuid"]

	// Get count from query parameter (default: 20)
	countStr := request.URL.Query().Get("count")
	count := 20
	if countStr != "" {
		if parsedCount, err := strconv.Atoi(countStr); err == nil && parsedCount > 0 {
			count = parsedCount
		}
	}

	matches, err := handler.serviceProxy.GetMatches(region, puuid, count)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(matches)
}

// AnalyzePlayer orchestrates player analysis by calling both data and cortex services
func (handler *Handler) AnalyzePlayer(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	region := vars["region"]
	summonerName := vars["summonerName"]

	// Step 1: Get summoner data from opgl-data
	summoner, err := handler.serviceProxy.GetSummoner(region, summonerName)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	// Step 2: Get match history from opgl-data
	matches, err := handler.serviceProxy.GetMatches(region, summoner.PUUID, 20)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	// Step 3: Send data to opgl-cortex-engine for analysis
	analysisResult, err := handler.serviceProxy.AnalyzePlayer(summoner, matches)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(analysisResult)
}
