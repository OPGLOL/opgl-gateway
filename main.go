package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/OPGLOL/opgl-gateway/internal/api"
	"github.com/OPGLOL/opgl-gateway/internal/proxy"
)

func main() {
	// Get configuration from environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dataServiceURL := os.Getenv("OPGL_DATA_URL")
	if dataServiceURL == "" {
		dataServiceURL = "http://localhost:8081"
	}

	cortexServiceURL := os.Getenv("OPGL_CORTEX_URL")
	if cortexServiceURL == "" {
		cortexServiceURL = "http://localhost:8082"
	}

	// Initialize service proxy
	serviceProxy := proxy.NewServiceProxy(dataServiceURL, cortexServiceURL)

	// Initialize HTTP handler
	handler := api.NewHandler(serviceProxy)

	// Set up router
	router := api.SetupRouter(handler)

	// Start server
	serverAddress := fmt.Sprintf(":%s", port)
	log.Printf("OPGL Gateway starting on port %s", port)
	log.Printf("Routing to Data Service: %s", dataServiceURL)
	log.Printf("Routing to Cortex Engine: %s", cortexServiceURL)
	log.Fatal(http.ListenAndServe(serverAddress, router))
}
