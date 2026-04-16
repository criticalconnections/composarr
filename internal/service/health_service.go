package service

import (
	"context"
	"strings"
	"time"

	"github.com/axism/composarr/internal/config"
	"github.com/axism/composarr/internal/models"
	"github.com/axism/composarr/internal/repository"
	ws "github.com/axism/composarr/internal/websocket"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// HealthOutcome describes the final outcome of a health-check loop.
type HealthOutcome int

const (
	HealthOutcomeHealthy HealthOutcome = iota
	HealthOutcomeUnhealthy
	HealthOutcomeTimeout
	HealthOutcomeCancelled
)

// Health classification for individual containers.
const (
	StatusHealthy   = "healthy"
	StatusUnhealthy = "unhealthy"
	StatusStarting  = "starting"
	StatusNone      = "none"
)

type HealthService struct {
	dockerSvc *DockerService
	repo      *repository.HealthCheckRepository
	hub       *ws.Hub
	cfg       *config.Config
}

func NewHealthService(dockerSvc *DockerService, repo *repository.HealthCheckRepository, hub *ws.Hub, cfg *config.Config) *HealthService {
	return &HealthService{
		dockerSvc: dockerSvc,
		repo:      repo,
		hub:       hub,
		cfg:       cfg,
	}
}

// MonitorDeployment polls container health for the deployment until either:
// - all containers are healthy (returns HealthOutcomeHealthy)
// - any container is fatally unhealthy (returns HealthOutcomeUnhealthy)
// - the timeout expires (returns HealthOutcomeTimeout)
// - the context is cancelled (returns HealthOutcomeCancelled)
//
// Each poll is broadcast as a websocket health.update event.
func (h *HealthService) MonitorDeployment(ctx context.Context, deploymentID, stackID, stackSlug string) HealthOutcome {
	timeout := time.Duration(h.cfg.HealthTimeout) * time.Second
	interval := time.Duration(h.cfg.HealthInterval) * time.Second

	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	startTime := time.Now()

	// First check immediately
	outcome := h.runOneCheck(ctx, deploymentID, stackID, stackSlug, startTime)
	if outcome == HealthOutcomeHealthy || outcome == HealthOutcomeUnhealthy {
		return outcome
	}

	for {
		select {
		case <-ctx.Done():
			return HealthOutcomeCancelled
		case <-deadline.C:
			log.Warn().
				Str("deployment", deploymentID).
				Dur("timeout", timeout).
				Msg("health check timed out")
			return HealthOutcomeTimeout
		case <-ticker.C:
			outcome := h.runOneCheck(ctx, deploymentID, stackID, stackSlug, startTime)
			if outcome == HealthOutcomeHealthy || outcome == HealthOutcomeUnhealthy {
				return outcome
			}
		}
	}
}

// runOneCheck performs a single round of inspecting containers, persisting and broadcasting results.
// Returns Healthy if all containers are healthy, Unhealthy if any is fatally unhealthy, otherwise -1 to keep polling.
func (h *HealthService) runOneCheck(ctx context.Context, deploymentID, stackID, stackSlug string, startTime time.Time) HealthOutcome {
	containers, err := h.dockerSvc.Ps(ctx, stackSlug)
	if err != nil {
		log.Warn().Err(err).Str("stack", stackSlug).Msg("health check: failed to list containers")
		return -1
	}

	if len(containers) == 0 {
		// No containers yet — keep waiting (they may still be starting)
		if time.Since(startTime) < 10*time.Second {
			return -1
		}
		// After grace period, no containers means deployment failed
		return HealthOutcomeUnhealthy
	}

	results := make([]models.HealthCheckResult, 0, len(containers))
	allHealthy := true
	anyFatal := false

	for _, c := range containers {
		status := classifyContainer(c, time.Since(startTime))
		result := models.HealthCheckResult{
			ID:            uuid.New().String(),
			DeploymentID:  deploymentID,
			ContainerName: c.Name,
			ServiceName:   c.Service,
			Status:        status,
			CheckOutput:   c.Status,
			CheckedAt:     time.Now().UTC(),
		}

		if err := h.repo.Create(&result); err != nil {
			log.Warn().Err(err).Msg("failed to persist health result")
		}
		results = append(results, result)

		if status != StatusHealthy {
			allHealthy = false
		}
		// "fatal" means the container is in a state from which we should not wait further
		if isFatal(c) {
			anyFatal = true
		}
	}

	// Broadcast aggregate result
	h.hub.Broadcast(ws.NewEvent(ws.EventHealthUpdate, stackID, deploymentID, gin_H{
		"results":    results,
		"allHealthy": allHealthy,
	}))

	if allHealthy {
		return HealthOutcomeHealthy
	}
	if anyFatal {
		return HealthOutcomeUnhealthy
	}
	return -1
}

// classifyContainer maps the Docker compose ps output to a health status string.
// gracePeriod (since deploy started) is used for containers without a Docker HEALTHCHECK.
func classifyContainer(c ContainerStatus, sinceStart time.Duration) string {
	state := strings.ToLower(c.State)
	health := strings.ToLower(c.Health)

	if health == "healthy" {
		return StatusHealthy
	}
	if health == "unhealthy" {
		return StatusUnhealthy
	}
	if health == "starting" {
		return StatusStarting
	}

	// No HEALTHCHECK defined
	if state == "running" {
		// Grace period of 10s before declaring healthy without a healthcheck
		if sinceStart >= 10*time.Second {
			return StatusHealthy
		}
		return StatusStarting
	}

	if state == "restarting" {
		return StatusStarting
	}

	return StatusUnhealthy
}

// isFatal returns true when a container is in a state from which we should immediately abort waiting.
func isFatal(c ContainerStatus) bool {
	state := strings.ToLower(c.State)
	if state == "exited" || state == "dead" || state == "removing" {
		return true
	}
	// Running but explicitly unhealthy is a strong signal too
	if strings.ToLower(c.Health) == "unhealthy" {
		return true
	}
	return false
}

// gin_H is a tiny helper to build map[string]interface{} payloads without importing gin from a service file.
type gin_H map[string]interface{}
