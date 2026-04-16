package handler

import (
	"net/http"

	"github.com/axism/composarr/internal/service"
	"github.com/gin-gonic/gin"
)

type ScheduleHandler struct {
	schedSvc *service.SchedulerService
}

func NewScheduleHandler(schedSvc *service.SchedulerService) *ScheduleHandler {
	return &ScheduleHandler{schedSvc: schedSvc}
}

// ListAll godoc
// GET /api/v1/schedules
func (h *ScheduleHandler) ListAll(c *gin.Context) {
	schedules, err := h.schedSvc.ListSchedules("")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, schedules)
}

// ListByStack godoc
// GET /api/v1/stacks/:id/schedules
func (h *ScheduleHandler) ListByStack(c *gin.Context) {
	stackID := c.Param("id")
	schedules, err := h.schedSvc.ListSchedules(stackID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, schedules)
}

// Get godoc
// GET /api/v1/schedules/:id
func (h *ScheduleHandler) Get(c *gin.Context) {
	id := c.Param("id")
	sched, err := h.schedSvc.GetSchedule(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
		return
	}
	c.JSON(http.StatusOK, sched)
}

// Create godoc
// POST /api/v1/stacks/:id/schedules
func (h *ScheduleHandler) Create(c *gin.Context) {
	stackID := c.Param("id")
	var req service.CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.StackID = stackID

	sched, err := h.schedSvc.CreateSchedule(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, sched)
}

// Update godoc
// PUT /api/v1/schedules/:id
func (h *ScheduleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sched, err := h.schedSvc.UpdateSchedule(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sched)
}

// Delete godoc
// DELETE /api/v1/schedules/:id
func (h *ScheduleHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.schedSvc.DeleteSchedule(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "schedule deleted"})
}

// NextWindow godoc
// GET /api/v1/schedules/:id/next
func (h *ScheduleHandler) NextWindow(c *gin.Context) {
	id := c.Param("id")
	next, err := h.schedSvc.NextWindow(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"nextWindow": next})
}

// QueueUpdate godoc
// POST /api/v1/stacks/:id/queue
func (h *ScheduleHandler) QueueUpdate(c *gin.Context) {
	stackID := c.Param("id")
	var req service.QueueUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.StackID = stackID

	update, err := h.schedSvc.QueueUpdate(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, update)
}

// ListQueuedUpdates godoc
// GET /api/v1/stacks/:id/queue
func (h *ScheduleHandler) ListQueuedUpdates(c *gin.Context) {
	stackID := c.Param("id")
	updates, err := h.schedSvc.ListQueuedUpdates(stackID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updates)
}

// ListAllQueuedUpdates godoc
// GET /api/v1/queued-updates
func (h *ScheduleHandler) ListAllQueuedUpdates(c *gin.Context) {
	updates, err := h.schedSvc.ListQueuedUpdates("")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updates)
}

// CancelQueuedUpdate godoc
// DELETE /api/v1/queued-updates/:id
func (h *ScheduleHandler) CancelQueuedUpdate(c *gin.Context) {
	id := c.Param("id")
	if err := h.schedSvc.CancelQueuedUpdate(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "queued update cancelled"})
}
