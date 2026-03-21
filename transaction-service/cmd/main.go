package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"transaction-service/internal/client"
	"transaction-service/internal/handler"
	"transaction-service/internal/repository"
	"transaction-service/internal/service"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	errorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total number of HTTP errors",
		},
		[]string{"method", "endpoint", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

func init() {
	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(errorCount)
	prometheus.MustRegister(requestDuration)
}

func main() {
	// Phase 7.1: Load all configuration from environment variables
	port := mustEnv("PORT", "8081")
	fraudURL := mustEnv("FRAUD_SERVICE_URL", "")
	ledgerURL := mustEnv("LEDGER_SERVICE_URL", "")
	dbURL := mustEnv("DATABASE_URL", "")
	timeout := durationFromEnv("HTTP_TIMEOUT_MS", 2000*time.Millisecond)
	maxAttempts := intFromEnv("HTTP_RETRY_COUNT", 3)

	// Validate required environment variables
	if fraudURL == "" {
		log.Fatal("Error: FRAUD_SERVICE_URL environment variable is required but not set")
	}
	if ledgerURL == "" {
		log.Fatal("Error: LEDGER_SERVICE_URL environment variable is required but not set")
	}
	if dbURL == "" {
		log.Fatal("Error: DATABASE_URL environment variable is required but not set. " +
			"Example: DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable")
	}

	// Initialize Database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to open database connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	fraudClient := client.NewFraudClient(fraudURL, timeout, maxAttempts)
	ledgerClient := client.NewLedgerClient(ledgerURL, timeout, maxAttempts)
	transferService := service.NewTransferService(fraudClient, ledgerClient)
	transactionQueryService := service.NewTransactionQueryService(ledgerClient)

	userRepo := repository.NewUserRepository(db)
	idempotencyRepo := repository.NewIdempotencyRepository(db)
	authHandler := handler.NewAuthHandler(userRepo)

	httpHandler := handler.NewHTTPHandlerWithURLs(
		transferService,
		transactionQueryService,
		idempotencyRepo,
		db,
		ledgerURL,
		fraudURL,
		authHandler,
	)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           httpHandler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf(
		"transaction_service_started port=%s fraud_url=%s ledger_url=%s timeout_ms=%d max_attempts=%d",
		port,
		fraudURL,
		ledgerURL,
		timeout.Milliseconds(),
		maxAttempts,
	)

	// Phase 7.2: Graceful shutdown with signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("received_signal signal=%v, initiating graceful shutdown", sig)

		// Create shutdown context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Shutdown server and finish active requests
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("server_shutdown_error error=%v", err)
		}

		// Close database connection
		if err := db.Close(); err != nil {
			log.Printf("database_close_error error=%v", err)
		}

		log.Println("graceful_shutdown_complete")
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

// Phase 7.1: Environment variable loading helper
// mustEnv returns the environment variable or a default value if present
func mustEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	if fallback != "" {
		return fallback
	}

	// If no fallback provided, the variable is required
	log.Fatalf("Error: Required environment variable %s not set", key)
	return ""
}

// envOrDefault returns the environment variable or a default value
func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func durationFromEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return time.Duration(parsed) * time.Millisecond
}

func intFromEnv(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return fallback
	}

	return parsed
}
