package middleware

import "net/http"

// CORSMiddleware handles Cross-Origin Resource Sharing (CORS) preflight requests
// and adds appropriate headers to allow browser-based clients to access the API
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		// Set CORS headers to allow cross-origin requests
		responseWriter.Header().Set("Access-Control-Allow-Origin", "*")
		responseWriter.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		responseWriter.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight OPTIONS requests immediately
		if request.Method == http.MethodOptions {
			responseWriter.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(responseWriter, request)
	})
}
