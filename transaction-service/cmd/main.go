package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"transaction-service/internal/client"
	"transaction-service/internal/handler"
	"transaction-service/internal/repository"
	"transaction-service/internal/service"
	_ "github.com/lib/pq"
)

func main() {
	port := envOrDefault("PORT", "8081")
	fraudURL := envOrDefault("FRAUD_SERVICE_URL", "http://127.0.0.1:8082")
	ledgerURL := envOrDefault("LEDGER_SERVICE_URL", "http://127.0.0.1:8080")
	dbURL := envOrDefault("DATABASE_URL", "postgres://postgres:123123@localhost:5432/ddbank?sslmode=disable")
	timeout := durationFromEnv("HTTP_TIMEOUT_MS", 2000*time.Millisecond)
	retries := intFromEnv("HTTP_RETRY_COUNT", 1)

	// Initialize Database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to open database connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	fraudClient := client.NewFraudClient(fraudURL, timeout, retries)
	ledgerClient := client.NewLedgerClient(ledgerURL, timeout, retries)
	transferService := service.NewTransferService(fraudClient, ledgerClient)
	transactionQueryService := service.NewTransactionQueryService(ledgerClient)
	
	userRepo := repository.NewUserRepository(db)
	authHandler := handler.NewAuthHandler(userRepo)
	
	httpHandler := handler.NewHTTPHandler(transferService, transactionQueryService, authHandler)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           httpHandler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf(
		"transaction_service_started port=%s fraud_url=%s ledger_url=%s timeout_ms=%d retry_count=%d",
		port,
		fraudURL,
		ledgerURL,
		timeout.Milliseconds(),
		retries,
	)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

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
