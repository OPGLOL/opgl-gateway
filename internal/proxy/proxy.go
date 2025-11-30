package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	apierrors "github.com/OPGLOL/opgl-gateway-service/internal/errors"
	"github.com/OPGLOL/opgl-gateway-service/internal/models"
)

// ServiceProxy handles communication with microservices
type ServiceProxy struct {
	dataServiceURL   string
	cortexServiceURL string
	httpClient       *http.Client
}

// NewServiceProxy creates a new ServiceProxy instance
func NewServiceProxy(dataServiceURL string, cortexServiceURL string) *ServiceProxy {
	return &ServiceProxy{
		dataServiceURL:   dataServiceURL,
		cortexServiceURL: cortexServiceURL,
		httpClient:       &http.Client{},
	}
}

// GetSummonerByRiotID retrieves summoner data from opgl-data service using Riot ID
func (proxy *ServiceProxy) GetSummonerByRiotID(region string, gameName string, tagLine string) (*models.Summoner, error) {
	url := proxy.dataServiceURL + "/api/v1/summoner"

	requestBody := map[string]string{
		"region":   region,
		"gameName": gameName,
		"tagLine":  tagLine,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, apierrors.InternalError("Failed to prepare request")
	}

	response, err := proxy.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, apierrors.DataServiceError("Unable to connect to data service")
	}
	defer response.Body.Close()

	// Handle different status codes from data service
	if response.StatusCode != http.StatusOK {
		return nil, proxy.handleDataServiceError(response, gameName, tagLine)
	}

	var summoner models.Summoner
	if err := json.NewDecoder(response.Body).Decode(&summoner); err != nil {
		return nil, apierrors.InternalError("Failed to process summoner data")
	}

	return &summoner, nil
}

// GetMatchesByRiotID retrieves match history from opgl-data service using Riot ID
func (proxy *ServiceProxy) GetMatchesByRiotID(region string, gameName string, tagLine string, count int) ([]models.Match, error) {
	url := proxy.dataServiceURL + "/api/v1/matches"

	requestBody := map[string]interface{}{
		"region":   region,
		"gameName": gameName,
		"tagLine":  tagLine,
		"count":    count,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, apierrors.InternalError("Failed to prepare request")
	}

	response, err := proxy.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, apierrors.DataServiceError("Unable to connect to data service")
	}
	defer response.Body.Close()

	// Handle different status codes from data service
	if response.StatusCode != http.StatusOK {
		return nil, proxy.handleDataServiceError(response, gameName, tagLine)
	}

	var matches []models.Match
	if err := json.NewDecoder(response.Body).Decode(&matches); err != nil {
		return nil, apierrors.InternalError("Failed to process match data")
	}

	return matches, nil
}

// GetMatchesByPUUID retrieves match history from opgl-data service using PUUID (internal use)
func (proxy *ServiceProxy) GetMatchesByPUUID(region string, puuid string, count int) ([]models.Match, error) {
	url := proxy.dataServiceURL + "/api/v1/matches"

	requestBody := map[string]interface{}{
		"region": region,
		"puuid":  puuid,
		"count":  count,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, apierrors.InternalError("Failed to prepare request")
	}

	response, err := proxy.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, apierrors.DataServiceError("Unable to connect to data service")
	}
	defer response.Body.Close()

	// Handle different status codes from data service
	if response.StatusCode != http.StatusOK {
		return nil, proxy.handleDataServiceErrorByPUUID(response)
	}

	var matches []models.Match
	if err := json.NewDecoder(response.Body).Decode(&matches); err != nil {
		return nil, apierrors.InternalError("Failed to process match data")
	}

	return matches, nil
}

// AnalyzePlayer sends analysis request to opgl-cortex-engine
func (proxy *ServiceProxy) AnalyzePlayer(summoner *models.Summoner, matches []models.Match) (*models.AnalysisResult, error) {
	requestBody := map[string]interface{}{
		"summoner": summoner,
		"matches":  matches,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, apierrors.InternalError("Failed to prepare request")
	}

	url := proxy.cortexServiceURL + "/api/v1/analyze"
	response, err := proxy.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, apierrors.CortexServiceError("Unable to connect to analysis service")
	}
	defer response.Body.Close()

	// Handle different status codes from cortex service
	if response.StatusCode != http.StatusOK {
		return nil, proxy.handleCortexServiceError(response)
	}

	var analysisResult models.AnalysisResult
	if err := json.NewDecoder(response.Body).Decode(&analysisResult); err != nil {
		return nil, apierrors.InternalError("Failed to process analysis data")
	}

	return &analysisResult, nil
}

// handleDataServiceError converts data service HTTP errors to APIErrors
func (proxy *ServiceProxy) handleDataServiceError(response *http.Response, gameName string, tagLine string) *apierrors.APIError {
	body, _ := io.ReadAll(response.Body)

	switch response.StatusCode {
	case http.StatusNotFound:
		return apierrors.PlayerNotFound(gameName, tagLine)
	case http.StatusBadRequest:
		return apierrors.InvalidRequestBody(string(body))
	default:
		return apierrors.DataServiceError("Data service error: " + string(body))
	}
}

// handleDataServiceErrorByPUUID converts data service HTTP errors to APIErrors when using PUUID
func (proxy *ServiceProxy) handleDataServiceErrorByPUUID(response *http.Response) *apierrors.APIError {
	body, _ := io.ReadAll(response.Body)

	switch response.StatusCode {
	case http.StatusNotFound:
		return apierrors.MatchesNotFound("No matches found for this player")
	case http.StatusBadRequest:
		return apierrors.InvalidRequestBody(string(body))
	default:
		return apierrors.DataServiceError("Data service error: " + string(body))
	}
}

// handleCortexServiceError converts cortex service HTTP errors to APIErrors
func (proxy *ServiceProxy) handleCortexServiceError(response *http.Response) *apierrors.APIError {
	body, _ := io.ReadAll(response.Body)

	switch response.StatusCode {
	case http.StatusBadRequest:
		return apierrors.InvalidRequestBody(string(body))
	default:
		return apierrors.CortexServiceError("Analysis service error: " + string(body))
	}
}
