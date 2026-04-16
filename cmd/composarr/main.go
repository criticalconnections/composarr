package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/axism/composarr/internal/config"
	"github.com/axism/composarr/internal/database"
	"github.com/axism/composarr/internal/handler"
	"github.com/axism/composarr/internal/repository"
	"github.com/axism/composarr/internal/service"
	ws "github.com/axism/composarr/internal/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Load config
	cfg := config.Load()

	// Setup logging
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	log.Info().Str("version", "0.1.0").Msg("starting composarr")

	// Ensure data directories exist
	if err := os.MkdirAll(cfg.ReposDir, 0755); err != nil {
		log.Fatal().Err(err).Msg("failed to create repos directory")
	}

	// Initialize database
	db, err := database.New(cfg.DBPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize database")
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}

	// Initialize repositories
	stackRepo := repository.NewStackRepository(db.DB)
	deployRepo := repository.NewDeploymentRepository(db.DB)
	deployLogRepo := repository.NewDeploymentLogRepository(db.DB)
	healthRepo := repository.NewHealthCheckRepository(db.DB)

	// Initialize WebSocket hub
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Stop()

	// Initialize services
	dockerSvc := service.NewDockerService(cfg)
	gitSvc := service.NewGitService(cfg)
	diffSvc := service.NewDiffService(gitSvc)
	stackSvc := service.NewStackService(stackRepo, dockerSvc, gitSvc, cfg)
	healthSvc := service.NewHealthService(dockerSvc, healthRepo, hub, cfg)
	deploySvc := service.NewDeployService(cfg, stackRepo, deployRepo, deployLogRepo, dockerSvc, gitSvc, healthSvc, hub)

	// Setup router
	router := handler.NewRouter(handler.RouterDeps{
		StackSvc:      stackSvc,
		GitSvc:        gitSvc,
		DiffSvc:       diffSvc,
		DeploySvc:     deploySvc,
		DeployRepo:    deployRepo,
		DeployLogRepo: deployLogRepo,
		HealthRepo:    healthRepo,
		Hub:           hub,
	})

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	go func() {
		log.Info().Int("port", cfg.Port).Msg("server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server stopped")
}
