package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewResponseWriter tests the responseWriter constructor
func TestNewResponseWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	responseWriter := newResponseWriter(recorder)

	if responseWriter == nil {
		t.Fatal("Expected responseWriter to not be nil")
	}

	if responseWriter.statusCode != http.StatusOK {
		t.Errorf("Expected default status code %d, got %d", http.StatusOK, responseWriter.statusCode)
	}

	if responseWriter.ResponseWriter != recorder {
		t.Error("Expected ResponseWriter to be set correctly")
	}
}

// TestResponseWriterWriteHeader tests that WriteHeader captures status code
func TestResponseWriterWriteHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	responseWriter := newResponseWriter(recorder)

	responseWriter.WriteHeader(http.StatusNotFound)

	if responseWriter.statusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, responseWriter.statusCode)
	}
}

// TestLoggingMiddleware tests the logging middleware
func TestLoggingMiddleware(t *testing.T) {
	// Create a simple handler that returns 200 OK
	nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("OK"))
	})

	// Wrap with logging middleware
	middleware := LoggingMiddleware(nextHandler)

	// Create test request
	request, err := http.NewRequest("POST", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	request.RemoteAddr = "127.0.0.1:12345"
	request.Header.Set("User-Agent", "test-agent")

	// Record response
	responseRecorder := httptest.NewRecorder()

	// Execute middleware
	middleware.ServeHTTP(responseRecorder, request)

	// Verify response was passed through
	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	if responseRecorder.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", responseRecorder.Body.String())
	}
}

// TestLoggingMiddleware_4xxStatus tests logging for 4xx status codes
func TestLoggingMiddleware_4xxStatus(t *testing.T) {
	nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("Bad Request"))
	})

	middleware := LoggingMiddleware(nextHandler)

	request, _ := http.NewRequest("POST", "/api/v1/summoner", nil)
	responseRecorder := httptest.NewRecorder()

	middleware.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, responseRecorder.Code)
	}
}

// TestLoggingMiddleware_5xxStatus tests logging for 5xx status codes
func TestLoggingMiddleware_5xxStatus(t *testing.T) {
	nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Internal Server Error"))
	})

	middleware := LoggingMiddleware(nextHandler)

	request, _ := http.NewRequest("POST", "/api/v1/summoner", nil)
	responseRecorder := httptest.NewRecorder()

	middleware.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, responseRecorder.Code)
	}
}

// TestLoggingMiddleware_DefaultStatusCode tests that status code defaults to 200 if not explicitly set
func TestLoggingMiddleware_DefaultStatusCode(t *testing.T) {
	nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// Handler that writes body without explicit WriteHeader
		writer.Write([]byte("OK"))
	})

	middleware := LoggingMiddleware(nextHandler)

	request, _ := http.NewRequest("GET", "/health", nil)
	responseRecorder := httptest.NewRecorder()

	middleware.ServeHTTP(responseRecorder, request)

	// When Write is called without WriteHeader, Go defaults to 200
	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected default status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}
}

// TestLoggingMiddleware_PreservesResponseWriter tests that the original response writer is preserved
func TestLoggingMiddleware_PreservesResponseWriter(t *testing.T) {
	nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("X-Custom-Header", "test-value")
		writer.WriteHeader(http.StatusCreated)
		writer.Write([]byte("Created"))
	})

	middleware := LoggingMiddleware(nextHandler)

	request, _ := http.NewRequest("POST", "/api/v1/summoner", nil)
	responseRecorder := httptest.NewRecorder()

	middleware.ServeHTTP(responseRecorder, request)

	if responseRecorder.Header().Get("X-Custom-Header") != "test-value" {
		t.Error("Expected custom header to be preserved")
	}

	if responseRecorder.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, responseRecorder.Code)
	}

	if responseRecorder.Body.String() != "Created" {
		t.Errorf("Expected body 'Created', got '%s'", responseRecorder.Body.String())
	}
}

// TestLoggingMiddleware_DifferentMethods tests logging with different HTTP methods
func TestLoggingMiddleware_DifferentMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	middleware := LoggingMiddleware(nextHandler)

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			request, _ := http.NewRequest(method, "/test", nil)
			responseRecorder := httptest.NewRecorder()

			middleware.ServeHTTP(responseRecorder, request)

			if responseRecorder.Code != http.StatusOK {
				t.Errorf("Expected status code %d for method %s, got %d", http.StatusOK, method, responseRecorder.Code)
			}
		})
	}
}
