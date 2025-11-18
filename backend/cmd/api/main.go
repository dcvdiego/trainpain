package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dcvdiego/trainpain-backend/internal/api"
	"github.com/dcvdiego/trainpain-backend/internal/api/handlers"
	"github.com/dcvdiego/trainpain-backend/internal/config"
	"github.com/dcvdiego/trainpain-backend/internal/external"
	"github.com/dcvdiego/trainpain-backend/internal/repository/postgres"
	"github.com/dcvdiego/trainpain-backend/internal/service"
	"github.com/dcvdiego/trainpain-backend/pkg/logger"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load() // Ignore error if .env doesn't exist

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	appLogger, err := logger.New(cfg.Server.Env)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	appLogger.Info("starting trainpain API server",
		zap.String("env", cfg.Server.Env),
		zap.String("port", cfg.Server.Port),
	)

	// Connect to database
	db, err := sqlx.Connect("pgx", cfg.Database.ConnectionString())
	if err != nil {
		appLogger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		appLogger.Fatal("failed to ping database", zap.Error(err))
	}
	appLogger.Info("database connection established")

	// Initialize repositories
	stationRepo := postgres.NewStationRepository(db)
	routeRepo := postgres.NewRouteRepository(db)

	// Initialize external API clients
	var hspClient *external.HSPClient
	if cfg.HSP.Username != "" && cfg.HSP.Password != "" {
		hspClient = external.NewHSPClient(cfg.HSP.BaseURL, cfg.HSP.Username, cfg.HSP.Password)
		appLogger.Info("HSP client initialized")
	} else {
		appLogger.Warn("HSP credentials not provided, some features will not work")
		// Create a dummy client for now
		hspClient = external.NewHSPClient(cfg.HSP.BaseURL, "dummy", "dummy")
	}

	// Initialize services
	routeService := service.NewRouteService(stationRepo, routeRepo, hspClient, appLogger)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	stationHandler := handlers.NewStationHandler(stationRepo, appLogger)
	routeHandler := handlers.NewRouteHandler(routeService, appLogger)

	// Setup router
	router := api.SetupRouter(healthHandler, stationHandler, routeHandler)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second, // Increased to handle parallel HSP API calls
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		appLogger.Info("server listening", zap.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Fatal("server forced to shutdown", zap.Error(err))
	}

	appLogger.Info("server exited")
}
