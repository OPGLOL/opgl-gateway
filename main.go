package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OPGLOL/opgl-gateway-service/internal/api"
	"github.com/OPGLOL/opgl-gateway-service/internal/middleware"
	"github.com/OPGLOL/opgl-gateway-service/internal/proxy"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize zerolog with colorized console output for development
	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}).With().Timestamp().Caller().Logger()

	// Set global log level (can be configured via environment variable)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log.Info().Msg("Starting OPGL Gateway")

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

	authServiceURL := os.Getenv("OPGL_AUTH_URL")
	if authServiceURL == "" {
		authServiceURL = "http://localhost:8083"
	}

	log.Info().
		Str("port", port).
		Str("data_service_url", dataServiceURL).
		Str("cortex_service_url", cortexServiceURL).
		Str("auth_service_url", authServiceURL).
		Msg("Configuration loaded")

	// Initialize service proxy
	serviceProxy := proxy.NewServiceProxy(dataServiceURL, cortexServiceURL)

	// Initialize HTTP handler
	handler := api.NewHandler(serviceProxy)

	// Initialize rate limit client for auth service
	rateLimitClient := middleware.NewRateLimitServiceClient(authServiceURL)
	log.Info().
		Str("auth_service_url", authServiceURL).
		Msg("Rate limiting enabled via auth service")

	// Set up router with all handlers
	routerConfig := &api.RouterConfig{
		Handler:         handler,
		RateLimitClient: rateLimitClient,
	}
	router := api.SetupRouter(routerConfig)

	// Wrap router with CORS middleware first to handle preflight requests
	corsRouter := middleware.CORSMiddleware(router)

	// Wrap with logging middleware
	loggedRouter := middleware.LoggingMiddleware(corsRouter)

	// Create HTTP server
	serverAddress := fmt.Sprintf(":%s", port)
	server := &http.Server{
		Addr:    serverAddress,
		Handler: loggedRouter,
	}

	// Channel to listen for shutdown signals
	shutdownChannel := make(chan os.Signal, 1)
	signal.Notify(shutdownChannel, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		log.Info().
			Str("address", serverAddress).
			Str("port", port).
			Msg("OPGL Gateway listening")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for shutdown signal
	<-shutdownChannel
	log.Info().Msg("Shutting down server...")

	// Create shutdown context with timeout
	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	// Gracefully shutdown HTTP server
	if err := server.Shutdown(shutdownContext); err != nil {
		log.Error().Err(err).Msg("Server shutdown error")
	}

	log.Info().Msg("Server stopped")
}
