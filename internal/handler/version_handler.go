package handler

import (
	"net/http"
	"strconv"

	"github.com/axism/composarr/internal/service"
	"github.com/gin-gonic/gin"
)

type VersionHandler struct {
	stackSvc *service.StackService
	gitSvc   *service.GitService
	diffSvc  *service.DiffService
}

func NewVersionHandler(stackSvc *service.StackService, gitSvc *service.GitService, diffSvc *service.DiffService) *VersionHandler {
	return &VersionHandler{
		stackSvc: stackSvc,
		gitSvc:   gitSvc,
		diffSvc:  diffSvc,
	}
}

// ListVersions godoc
// GET /api/v1/stacks/:id/versions
func (h *VersionHandler) ListVersions(c *gin.Context) {
	id := c.Param("id")
	slug, err := h.stackSvc.SlugForID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "stack not found"})
		return
	}

	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	commits, err := h.gitSvc.GetLog(slug, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, commits)
}

// GetVersion godoc
// GET /api/v1/stacks/:id/versions/:hash
func (h *VersionHandler) GetVersion(c *gin.Context) {
	id := c.Param("id")
	hash := c.Param("hash")

	slug, err := h.stackSvc.SlugForID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "stack not found"})
		return
	}
	composePath, _ := h.stackSvc.ComposePathForID(id)

	info, err := h.gitSvc.GetCommitInfo(slug, hash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	content, err := h.gitSvc.GetFileAtCommit(slug, hash, composePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"commit":  info,
		"content": string(content),
	})
}

// GetVersionDiff godoc
// GET /api/v1/stacks/:id/versions/:hash/diff
// Returns the diff between this commit and its parent.
func (h *VersionHandler) GetVersionDiff(c *gin.Context) {
	id := c.Param("id")
	hash := c.Param("hash")

	slug, err := h.stackSvc.SlugForID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "stack not found"})
		return
	}
	composePath, _ := h.stackSvc.ComposePathForID(id)

	diff, err := h.diffSvc.DiffVersions(slug, composePath, "", hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, diff)
}

// GetWorkingDiff godoc
// GET /api/v1/stacks/:id/diff
// Returns the diff between HEAD and the working copy (uncommitted changes).
// If the request includes ?content=..., diffs HEAD against the supplied content.
func (h *VersionHandler) GetWorkingDiff(c *gin.Context) {
	id := c.Param("id")

	slug, err := h.stackSvc.SlugForID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "stack not found"})
		return
	}
	composePath, _ := h.stackSvc.ComposePathForID(id)

	// Optional: client can POST a candidate content to preview a diff before commit
	var body struct {
		Content string `json:"content"`
	}
	_ = c.ShouldBindJSON(&body)

	if body.Content != "" {
		// Diff HEAD against supplied content
		headHash, _ := h.gitSvc.GetHeadCommit(slug)
		var oldContent []byte
		if headHash != "" {
			oldContent, _ = h.gitSvc.GetFileAtCommit(slug, headHash, composePath)
		}
		c.JSON(http.StatusOK, gin.H{
			"oldHash":    headHash,
			"newHash":    "",
			"oldContent": string(oldContent),
			"newContent": body.Content,
		})
		return
	}

	diff, err := h.diffSvc.DiffWorking(slug, composePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, diff)
}

// GetDiffBetween godoc
// GET /api/v1/stacks/:id/diff/:from/:to
func (h *VersionHandler) GetDiffBetween(c *gin.Context) {
	id := c.Param("id")
	from := c.Param("from")
	to := c.Param("to")

	slug, err := h.stackSvc.SlugForID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "stack not found"})
		return
	}
	composePath, _ := h.stackSvc.ComposePathForID(id)

	diff, err := h.diffSvc.DiffVersions(slug, composePath, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, diff)
}

// Rollback godoc
// POST /api/v1/stacks/:id/versions/:hash/rollback
func (h *VersionHandler) Rollback(c *gin.Context) {
	id := c.Param("id")
	hash := c.Param("hash")

	commitHash, err := h.stackSvc.Rollback(id, hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "rolled back",
		"commitHash": commitHash,
	})
}
