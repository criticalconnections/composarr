package handler

import (
	"net/http"
	"strconv"

	"github.com/axism/composarr/internal/repository"
	"github.com/axism/composarr/internal/service"
	"github.com/gin-gonic/gin"
)

type DeployHandler struct {
	deploySvc      *service.DeployService
	deployRepo     *repository.DeploymentRepository
	deployLogRepo  *repository.DeploymentLogRepository
	healthRepo     *repository.HealthCheckRepository
}

func NewDeployHandler(
	deploySvc *service.DeployService,
	deployRepo *repository.DeploymentRepository,
	deployLogRepo *repository.DeploymentLogRepository,
	healthRepo *repository.HealthCheckRepository,
) *DeployHandler {
	return &DeployHandler{
		deploySvc:     deploySvc,
		deployRepo:    deployRepo,
		deployLogRepo: deployLogRepo,
		healthRepo:    healthRepo,
	}
}

// Deploy godoc
// POST /api/v1/stacks/:id/deploy
func (h *DeployHandler) Deploy(c *gin.Context) {
	stackID := c.Param("id")

	var req struct {
		SkipPull        bool   `json:"skipPull"`
		SkipHealthCheck bool   `json:"skipHealthCheck"`
		Trigger         string `json:"trigger"`
	}
	_ = c.ShouldBindJSON(&req)

	deploymentID, err := h.deploySvc.Deploy(stackID, service.DeployOptions{
		SkipPull:        req.SkipPull,
		SkipHealthCheck: req.SkipHealthCheck,
		Trigger:         req.Trigger,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"deploymentId": deploymentID,
		"message":      "deployment started",
	})
}

// CancelDeploy godoc
// POST /api/v1/deployments/:id/cancel
func (h *DeployHandler) CancelDeploy(c *gin.Context) {
	deploymentID := c.Param("id")
	if err := h.deploySvc.Cancel(deploymentID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "cancellation requested"})
}

// ListDeployments godoc
// GET /api/v1/deployments?stackId=&limit=
func (h *DeployHandler) ListDeployments(c *gin.Context) {
	stackID := c.Query("stackId")
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	deployments, err := h.deployRepo.List(stackID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, deployments)
}

// GetDeployment godoc
// GET /api/v1/deployments/:id
func (h *DeployHandler) GetDeployment(c *gin.Context) {
	id := c.Param("id")
	deployment, err := h.deployRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}

	logs, _ := h.deployLogRepo.ListByDeployment(id)
	healthResults, _ := h.healthRepo.LatestByDeployment(id)

	c.JSON(http.StatusOK, gin.H{
		"deployment":    deployment,
		"logs":          logs,
		"healthResults": healthResults,
	})
}

// GetDeploymentLogs godoc
// GET /api/v1/deployments/:id/logs
func (h *DeployHandler) GetDeploymentLogs(c *gin.Context) {
	id := c.Param("id")
	logs, err := h.deployLogRepo.ListByDeployment(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}

// GetDeploymentHealth godoc
// GET /api/v1/deployments/:id/health
func (h *DeployHandler) GetDeploymentHealth(c *gin.Context) {
	id := c.Param("id")
	all, err := h.healthRepo.ListByDeployment(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	latest, _ := h.healthRepo.LatestByDeployment(id)
	c.JSON(http.StatusOK, gin.H{
		"all":    all,
		"latest": latest,
	})
}
