package handler

import (
	"net/http"

	"github.com/axism/composarr/internal/service"
	"github.com/gin-gonic/gin"
)

type DependencyHandler struct {
	depSvc *service.DependencyService
}

func NewDependencyHandler(depSvc *service.DependencyService) *DependencyHandler {
	return &DependencyHandler{depSvc: depSvc}
}

// ListDependencies godoc
// GET /api/v1/stacks/:id/dependencies
func (h *DependencyHandler) ListDependencies(c *gin.Context) {
	stackID := c.Param("id")
	deps, err := h.depSvc.ListDependencies(stackID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, deps)
}

// ListDependents godoc
// GET /api/v1/stacks/:id/dependents
func (h *DependencyHandler) ListDependents(c *gin.Context) {
	stackID := c.Param("id")
	deps, err := h.depSvc.ListDependents(stackID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, deps)
}

// AddDependency godoc
// POST /api/v1/stacks/:id/dependencies
func (h *DependencyHandler) AddDependency(c *gin.Context) {
	stackID := c.Param("id")
	var req struct {
		DependsOnID    string `json:"dependsOnId"`
		DependencyType string `json:"dependencyType"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dep, err := h.depSvc.AddDependency(stackID, req.DependsOnID, req.DependencyType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dep)
}

// RemoveDependency godoc
// DELETE /api/v1/stacks/:id/dependencies/:depId
func (h *DependencyHandler) RemoveDependency(c *gin.Context) {
	depID := c.Param("depId")
	if err := h.depSvc.RemoveDependency(depID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "dependency removed"})
}

// GetGraph godoc
// GET /api/v1/dependencies/graph
func (h *DependencyHandler) GetGraph(c *gin.Context) {
	graph, err := h.depSvc.GetGraph()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, graph)
}
