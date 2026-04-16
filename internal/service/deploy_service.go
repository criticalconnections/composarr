package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/axism/composarr/internal/config"
	"github.com/axism/composarr/internal/models"
	"github.com/axism/composarr/internal/repository"
	ws "github.com/axism/composarr/internal/websocket"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Deployment status enum.
const (
	DeploymentStatusPending        = "pending"
	DeploymentStatusRunning        = "running"
	DeploymentStatusHealthChecking = "health_checking"
	DeploymentStatusSucceeded      = "succeeded"
	DeploymentStatusFailed         = "failed"
	DeploymentStatusRolledBack     = "rolled_back"
)

// Deployment trigger enum.
const (
	TriggerManual     = "manual"
	TriggerScheduled  = "scheduled"
	TriggerAutoUpdate = "auto_update"
	TriggerRollback   = "rollback"
)

type DeployOptions struct {
	Trigger         string // see Trigger* constants
	SkipPull        bool   // skip docker compose pull
	SkipHealthCheck bool   // skip health check (for emergency deploys)
	CommitMessage   string // optional commit message override
}

type DeployService struct {
	cfg            *config.Config
	stackRepo      *repository.StackRepository
	deployRepo     *repository.DeploymentRepository
	deployLogRepo  *repository.DeploymentLogRepository
	dockerSvc      *DockerService
	gitSvc         *GitService
	healthSvc      *HealthService
	depSvc         *DependencyService
	hub            *ws.Hub

	// Track in-flight deployments so they can be cancelled
	mu             sync.Mutex
	cancelFns      map[string]context.CancelFunc
}

func NewDeployService(
	cfg *config.Config,
	stackRepo *repository.StackRepository,
	deployRepo *repository.DeploymentRepository,
	deployLogRepo *repository.DeploymentLogRepository,
	dockerSvc *DockerService,
	gitSvc *GitService,
	healthSvc *HealthService,
	hub *ws.Hub,
) *DeployService {
	return &DeployService{
		cfg:           cfg,
		stackRepo:     stackRepo,
		deployRepo:    deployRepo,
		deployLogRepo: deployLogRepo,
		dockerSvc:     dockerSvc,
		gitSvc:        gitSvc,
		healthSvc:     healthSvc,
		hub:           hub,
		cancelFns:     make(map[string]context.CancelFunc),
	}
}

// SetDependencyService wires the dependency service after construction.
// This breaks an otherwise circular constructor dependency.
func (d *DeployService) SetDependencyService(depSvc *DependencyService) {
	d.depSvc = depSvc
}

// Deploy starts a new deployment for the stack and returns the deployment ID.
// The deployment runs in a goroutine; status updates are pushed via WebSocket.
func (d *DeployService) Deploy(stackID string, opts DeployOptions) (string, error) {
	stack, err := d.stackRepo.GetByID(stackID)
	if err != nil {
		return "", fmt.Errorf("get stack: %w", err)
	}

	// Get current HEAD commit (the version we're deploying)
	commitHash, err := d.gitSvc.GetHeadCommit(stack.Slug)
	if err != nil {
		return "", fmt.Errorf("get head commit: %w", err)
	}
	if commitHash == "" {
		return "", errors.New("stack has no committed compose file")
	}

	// Find the previous successful deployment for rollback reference
	previousCommit := ""
	previousDeployments, _ := d.deployRepo.List(stackID, 10)
	for _, prev := range previousDeployments {
		if prev.Status == DeploymentStatusSucceeded {
			previousCommit = prev.CommitHash
			break
		}
	}

	if opts.Trigger == "" {
		opts.Trigger = TriggerManual
	}

	now := time.Now().UTC()
	deployment := &models.Deployment{
		ID:             uuid.New().String(),
		StackID:        stackID,
		CommitHash:     commitHash,
		PreviousCommit: previousCommit,
		Status:         DeploymentStatusPending,
		TriggerType:    opts.Trigger,
		StartedAt:      &now,
		CreatedAt:      now,
	}

	if err := d.deployRepo.Create(deployment); err != nil {
		return "", fmt.Errorf("create deployment record: %w", err)
	}

	// Run pipeline in a goroutine so the HTTP response returns immediately
	ctx, cancel := context.WithCancel(context.Background())
	d.mu.Lock()
	d.cancelFns[deployment.ID] = cancel
	d.mu.Unlock()

	go func() {
		defer func() {
			d.mu.Lock()
			delete(d.cancelFns, deployment.ID)
			d.mu.Unlock()
			cancel()
		}()
		d.runPipeline(ctx, stack, deployment, opts)
	}()

	return deployment.ID, nil
}

// Cancel attempts to cancel an in-flight deployment.
func (d *DeployService) Cancel(deploymentID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	cancel, ok := d.cancelFns[deploymentID]
	if !ok {
		return errors.New("deployment is not in flight")
	}
	cancel()
	return nil
}

// runPipeline executes the full deploy lifecycle for a single deployment.
func (d *DeployService) runPipeline(ctx context.Context, stack *models.Stack, dep *models.Deployment, opts DeployOptions) {
	logger := log.With().
		Str("deployment", dep.ID).
		Str("stack", stack.Slug).
		Str("commit", dep.CommitHash).
		Logger()

	d.broadcast(ws.EventDeployStarted, stack.ID, dep.ID, gin_H{"commitHash": dep.CommitHash})
	d.appendLog(dep.ID, "info", fmt.Sprintf("Deployment started (commit %s, trigger=%s)", short(dep.CommitHash), opts.Trigger))
	d.updateStatus(dep, DeploymentStatusRunning, "")

	// 1. Validate compose file
	d.broadcast(ws.EventDeployValidating, stack.ID, dep.ID, nil)
	d.appendLog(dep.ID, "info", "Validating compose file...")
	if err := d.dockerSvc.Validate(ctx, stack.Slug); err != nil {
		d.fail(stack, dep, fmt.Sprintf("validation failed: %v", err))
		return
	}

	// 1b. Check hard dependencies are running
	if d.depSvc != nil {
		d.appendLog(dep.ID, "info", "Checking dependencies...")
		if err := d.depSvc.PreDeployCheck(ctx, stack.ID); err != nil {
			d.fail(stack, dep, fmt.Sprintf("dependency check failed: %v", err))
			return
		}
	}

	// 2. Pull images
	if !opts.SkipPull {
		d.broadcast(ws.EventDeployPulling, stack.ID, dep.ID, nil)
		d.appendLog(dep.ID, "info", "Pulling images...")
		if output, err := d.dockerSvc.Pull(ctx, stack.Slug); err != nil {
			// Pulling failures shouldn't necessarily abort the deploy (image may be local)
			d.appendLog(dep.ID, "warn", fmt.Sprintf("docker compose pull warning: %v", err))
			logger.Warn().Err(err).Str("output", output).Msg("pull warning, continuing")
		}
	}

	// 3. Up
	d.broadcast(ws.EventDeployStarting, stack.ID, dep.ID, nil)
	d.appendLog(dep.ID, "info", "Starting containers (docker compose up -d)...")
	if output, err := d.dockerSvc.Up(ctx, stack.Slug); err != nil {
		d.appendLog(dep.ID, "error", fmt.Sprintf("compose up failed: %v\n%s", err, output))
		d.attemptRollback(stack, dep, fmt.Sprintf("compose up failed: %v", err))
		return
	}
	d.stackRepo.UpdateStatus(stack.ID, "running")

	// 4. Health check
	if opts.SkipHealthCheck {
		d.appendLog(dep.ID, "info", "Health check skipped per request")
		d.succeed(stack, dep)
		return
	}

	d.updateStatus(dep, DeploymentStatusHealthChecking, "")
	d.broadcast(ws.EventDeployHealthChecking, stack.ID, dep.ID, nil)
	d.appendLog(dep.ID, "info", fmt.Sprintf("Verifying container health (timeout %ds, interval %ds)...", d.cfg.HealthTimeout, d.cfg.HealthInterval))

	outcome := d.healthSvc.MonitorDeployment(ctx, dep.ID, stack.ID, stack.Slug)
	switch outcome {
	case HealthOutcomeHealthy:
		d.appendLog(dep.ID, "info", "All containers healthy")
		d.succeed(stack, dep)
	case HealthOutcomeUnhealthy:
		d.appendLog(dep.ID, "error", "Containers reported unhealthy")
		d.attemptRollback(stack, dep, "health check failed: containers unhealthy")
	case HealthOutcomeTimeout:
		d.appendLog(dep.ID, "error", fmt.Sprintf("Health check timed out after %ds", d.cfg.HealthTimeout))
		d.attemptRollback(stack, dep, "health check timed out")
	case HealthOutcomeCancelled:
		d.appendLog(dep.ID, "warn", "Deployment cancelled")
		d.fail(stack, dep, "deployment cancelled")
	}
}

func (d *DeployService) succeed(stack *models.Stack, dep *models.Deployment) {
	d.updateStatus(dep, DeploymentStatusSucceeded, "")
	d.broadcast(ws.EventDeploySucceeded, stack.ID, dep.ID, gin_H{"commitHash": dep.CommitHash})
	d.appendLog(dep.ID, "info", "Deployment succeeded ✓")
}

func (d *DeployService) fail(stack *models.Stack, dep *models.Deployment, message string) {
	d.updateStatus(dep, DeploymentStatusFailed, message)
	d.broadcast(ws.EventDeployFailed, stack.ID, dep.ID, gin_H{"error": message})
	d.appendLog(dep.ID, "error", message)
}

// attemptRollback rolls the stack back to the previous successful commit, if available.
func (d *DeployService) attemptRollback(stack *models.Stack, dep *models.Deployment, reason string) {
	if dep.PreviousCommit == "" {
		d.appendLog(dep.ID, "error", "Cannot rollback: no previous successful deployment")
		d.fail(stack, dep, reason+" (no previous commit to roll back to)")
		return
	}

	d.broadcast(ws.EventDeployRollingBack, stack.ID, dep.ID, gin_H{
		"reason":         reason,
		"targetCommit":   dep.PreviousCommit,
	})
	d.appendLog(dep.ID, "warn", fmt.Sprintf("Rolling back to commit %s (%s)", short(dep.PreviousCommit), reason))

	// Restore the previous compose file (creates a new git commit)
	rollbackHash, err := d.gitSvc.RollbackToCommit(stack.Slug, dep.PreviousCommit, stack.ComposePath)
	if err != nil {
		d.appendLog(dep.ID, "error", fmt.Sprintf("git rollback failed: %v", err))
		d.fail(stack, dep, fmt.Sprintf("rollback failed: %v", err))
		return
	}

	// Re-deploy with the restored compose file (no health check to avoid recursion)
	rollbackCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if output, err := d.dockerSvc.Up(rollbackCtx, stack.Slug); err != nil {
		d.appendLog(dep.ID, "error", fmt.Sprintf("rollback compose up failed: %v\n%s", err, output))
		d.fail(stack, dep, fmt.Sprintf("rollback failed: %v", err))
		return
	}

	d.updateStatus(dep, DeploymentStatusRolledBack, reason)
	d.broadcast(ws.EventDeployRolledBack, stack.ID, dep.ID, gin_H{
		"reason":         reason,
		"rollbackCommit": rollbackHash,
	})
	d.appendLog(dep.ID, "warn", fmt.Sprintf("Rolled back successfully (new commit %s)", short(rollbackHash)))

	// Create a separate deployment record for the rollback itself
	now := time.Now().UTC()
	rollbackDep := &models.Deployment{
		ID:             uuid.New().String(),
		StackID:        stack.ID,
		CommitHash:     rollbackHash,
		PreviousCommit: dep.CommitHash,
		Status:         DeploymentStatusSucceeded,
		TriggerType:    TriggerRollback,
		StartedAt:      &now,
		CompletedAt:    &now,
		CreatedAt:      now,
	}
	d.deployRepo.Create(rollbackDep)
}

func (d *DeployService) broadcast(eventType, stackID, deploymentID string, payload interface{}) {
	d.hub.Broadcast(ws.NewEvent(eventType, stackID, deploymentID, payload))
}

func (d *DeployService) updateStatus(dep *models.Deployment, status, errorMessage string) {
	dep.Status = status
	dep.ErrorMessage = errorMessage
	if status == DeploymentStatusSucceeded || status == DeploymentStatusFailed || status == DeploymentStatusRolledBack {
		now := time.Now().UTC()
		dep.CompletedAt = &now
	}
	if err := d.deployRepo.UpdateStatus(dep.ID, status, errorMessage); err != nil {
		log.Warn().Err(err).Msg("failed to update deployment status")
	}
}

func (d *DeployService) appendLog(deploymentID, level, message string) {
	entry := &models.DeploymentLog{
		ID:           uuid.New().String(),
		DeploymentID: deploymentID,
		Level:        level,
		Message:      message,
		Timestamp:    time.Now().UTC(),
	}
	if err := d.deployLogRepo.Create(entry); err != nil {
		log.Warn().Err(err).Msg("failed to write deployment log")
	}
	d.broadcast(ws.EventDeployLog, "", deploymentID, gin_H{
		"level":   level,
		"message": message,
		"time":    entry.Timestamp,
	})
}

func short(hash string) string {
	if len(hash) > 8 {
		return hash[:8]
	}
	return hash
}
