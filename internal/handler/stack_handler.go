package handler

import (
	"net/http"
	"strconv"

	"github.com/axism/composarr/internal/service"
	"github.com/gin-gonic/gin"
)

type StackHandler struct {
	stackSvc *service.StackService
}

func NewStackHandler(stackSvc *service.StackService) *StackHandler {
	return &StackHandler{stackSvc: stackSvc}
}

// ListStacks godoc
// GET /api/v1/stacks
func (h *StackHandler) ListStacks(c *gin.Context) {
	stacks, err := h.stackSvc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stacks)
}

// GetStack godoc
// GET /api/v1/stacks/:id
func (h *StackHandler) GetStack(c *gin.Context) {
	id := c.Param("id")
	stack, err := h.stackSvc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "stack not found"})
		return
	}
	c.JSON(http.StatusOK, stack)
}

// CreateStack godoc
// POST /api/v1/stacks
func (h *StackHandler) CreateStack(c *gin.Context) {
	var req service.CreateStackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if req.ComposeContent == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "composeContent is required"})
		return
	}

	stack, err := h.stackSvc.Create(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, stack)
}

// UpdateStack godoc
// PUT /api/v1/stacks/:id
func (h *StackHandler) UpdateStack(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdateStackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stack, err := h.stackSvc.Update(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stack)
}

// DeleteStack godoc
// DELETE /api/v1/stacks/:id
func (h *StackHandler) DeleteStack(c *gin.Context) {
	id := c.Param("id")
	if err := h.stackSvc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "stack deleted"})
}

// GetCompose godoc
// GET /api/v1/stacks/:id/compose
func (h *StackHandler) GetCompose(c *gin.Context) {
	id := c.Param("id")
	content, err := h.stackSvc.GetCompose(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"content": content})
}

// UpdateCompose godoc
// PUT /api/v1/stacks/:id/compose
func (h *StackHandler) UpdateCompose(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdateComposeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	commitHash, err := h.stackSvc.UpdateCompose(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "compose file updated",
		"commitHash": commitHash,
	})
}

// StartStack godoc
// POST /api/v1/stacks/:id/start
func (h *StackHandler) StartStack(c *gin.Context) {
	id := c.Param("id")
	output, err := h.stackSvc.Start(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "stack started", "output": output})
}

// StopStack godoc
// POST /api/v1/stacks/:id/stop
func (h *StackHandler) StopStack(c *gin.Context) {
	id := c.Param("id")
	output, err := h.stackSvc.Stop(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "stack stopped", "output": output})
}

// RestartStack godoc
// POST /api/v1/stacks/:id/restart
func (h *StackHandler) RestartStack(c *gin.Context) {
	id := c.Param("id")
	output, err := h.stackSvc.Restart(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "stack restarted", "output": output})
}

// GetStatus godoc
// GET /api/v1/stacks/:id/status
func (h *StackHandler) GetStatus(c *gin.Context) {
	id := c.Param("id")
	containers, err := h.stackSvc.Status(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, containers)
}

// GetLogs godoc
// GET /api/v1/stacks/:id/logs
func (h *StackHandler) GetLogs(c *gin.Context) {
	id := c.Param("id")
	svc := c.Query("service")
	tail := 100
	if t := c.Query("tail"); t != "" {
		if parsed, err := strconv.Atoi(t); err == nil && parsed > 0 {
			tail = parsed
		}
	}

	logs, err := h.stackSvc.Logs(c.Request.Context(), id, svc, tail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": logs})
}
