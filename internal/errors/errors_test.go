package errors

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewAPIError tests the NewAPIError constructor
func TestNewAPIError(t *testing.T) {
	apiError := NewAPIError(ErrCodePlayerNotFound, "Player not found", http.StatusNotFound)

	if apiError.Code != ErrCodePlayerNotFound {
		t.Errorf("Expected code '%s', got '%s'", ErrCodePlayerNotFound, apiError.Code)
	}

	if apiError.Message != "Player not found" {
		t.Errorf("Expected message 'Player not found', got '%s'", apiError.Message)
	}

	if apiError.Status != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, apiError.Status)
	}
}

// TestAPIError_Error tests the Error interface implementation
func TestAPIError_Error(t *testing.T) {
	apiError := NewAPIError(ErrCodeInternalError, "Something went wrong", http.StatusInternalServerError)

	errorMessage := apiError.Error()

	if errorMessage != "Something went wrong" {
		t.Errorf("Expected error message 'Something went wrong', got '%s'", errorMessage)
	}
}

// TestInvalidRequestBody tests the InvalidRequestBody constructor
func TestInvalidRequestBody(t *testing.T) {
	apiError := InvalidRequestBody("Invalid JSON format")

	if apiError.Code != ErrCodeInvalidRequestBody {
		t.Errorf("Expected code '%s', got '%s'", ErrCodeInvalidRequestBody, apiError.Code)
	}

	if apiError.Status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, apiError.Status)
	}
}

// TestMissingFields tests the MissingFields constructor
func TestMissingFields(t *testing.T) {
	apiError := MissingFields("region is required")

	if apiError.Code != ErrCodeMissingFields {
		t.Errorf("Expected code '%s', got '%s'", ErrCodeMissingFields, apiError.Code)
	}

	if apiError.Status != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, apiError.Status)
	}
}

// TestPlayerNotFound tests the PlayerNotFound constructor
func TestPlayerNotFound(t *testing.T) {
	apiError := PlayerNotFound("TestPlayer", "NA1")

	if apiError.Code != ErrCodePlayerNotFound {
		t.Errorf("Expected code '%s', got '%s'", ErrCodePlayerNotFound, apiError.Code)
	}

	if apiError.Status != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, apiError.Status)
	}

	expectedMessage := "Player not found: TestPlayer#NA1"
	if apiError.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, apiError.Message)
	}
}

// TestMatchesNotFound tests the MatchesNotFound constructor
func TestMatchesNotFound(t *testing.T) {
	apiError := MatchesNotFound("No matches found")

	if apiError.Code != ErrCodeMatchesNotFound {
		t.Errorf("Expected code '%s', got '%s'", ErrCodeMatchesNotFound, apiError.Code)
	}

	if apiError.Status != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, apiError.Status)
	}
}

// TestDataServiceError tests the DataServiceError constructor
func TestDataServiceError(t *testing.T) {
	apiError := DataServiceError("Data service unavailable")

	if apiError.Code != ErrCodeDataServiceError {
		t.Errorf("Expected code '%s', got '%s'", ErrCodeDataServiceError, apiError.Code)
	}

	if apiError.Status != http.StatusBadGateway {
		t.Errorf("Expected status %d, got %d", http.StatusBadGateway, apiError.Status)
	}
}

// TestCortexServiceError tests the CortexServiceError constructor
func TestCortexServiceError(t *testing.T) {
	apiError := CortexServiceError("Cortex service unavailable")

	if apiError.Code != ErrCodeCortexServiceError {
		t.Errorf("Expected code '%s', got '%s'", ErrCodeCortexServiceError, apiError.Code)
	}

	if apiError.Status != http.StatusBadGateway {
		t.Errorf("Expected status %d, got %d", http.StatusBadGateway, apiError.Status)
	}
}

// TestInternalError tests the InternalError constructor
func TestInternalError(t *testing.T) {
	apiError := InternalError("Unexpected error")

	if apiError.Code != ErrCodeInternalError {
		t.Errorf("Expected code '%s', got '%s'", ErrCodeInternalError, apiError.Code)
	}

	if apiError.Status != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, apiError.Status)
	}
}

// TestWriteError tests the WriteError function
func TestWriteError(t *testing.T) {
	apiError := PlayerNotFound("TestPlayer", "NA1")

	responseRecorder := httptest.NewRecorder()
	WriteError(responseRecorder, apiError)

	// Check status code
	if responseRecorder.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, responseRecorder.Code)
	}

	// Check content type
	contentType := responseRecorder.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	// Check response body structure
	var errorResponse ErrorResponse
	err := json.NewDecoder(responseRecorder.Body).Decode(&errorResponse)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if errorResponse.Error.Code != ErrCodePlayerNotFound {
		t.Errorf("Expected error code '%s', got '%s'", ErrCodePlayerNotFound, errorResponse.Error.Code)
	}

	expectedMessage := "Player not found: TestPlayer#NA1"
	if errorResponse.Error.Message != expectedMessage {
		t.Errorf("Expected error message '%s', got '%s'", expectedMessage, errorResponse.Error.Message)
	}
}

// TestWriteError_DifferentStatusCodes tests WriteError with various status codes
func TestWriteError_DifferentStatusCodes(t *testing.T) {
	testCases := []struct {
		name           string
		apiError       *APIError
		expectedStatus int
	}{
		{"bad request", InvalidRequestBody("Invalid JSON"), http.StatusBadRequest},
		{"not found", PlayerNotFound("Test", "NA1"), http.StatusNotFound},
		{"bad gateway", DataServiceError("Service down"), http.StatusBadGateway},
		{"internal error", InternalError("Unexpected"), http.StatusInternalServerError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			responseRecorder := httptest.NewRecorder()
			WriteError(responseRecorder, testCase.apiError)

			if responseRecorder.Code != testCase.expectedStatus {
				t.Errorf("Expected status %d, got %d", testCase.expectedStatus, responseRecorder.Code)
			}
		})
	}
}
