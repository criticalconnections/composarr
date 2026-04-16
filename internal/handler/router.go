package handler

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/axism/composarr/internal/repository"
	"github.com/axism/composarr/internal/service"
	ws "github.com/axism/composarr/internal/websocket"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// StaticFS holds the embedded frontend files. Set by main.go when available.
var StaticFS fs.FS

type RouterDeps struct {
	StackSvc      *service.StackService
	GitSvc        *service.GitService
	DiffSvc       *service.DiffService
	DeploySvc     *service.DeployService
	SchedSvc      *service.SchedulerService
	DepSvc        *service.DependencyService
	DeployRepo    *repository.DeploymentRepository
	DeployLogRepo *repository.DeploymentLogRepository
	HealthRepo    *repository.HealthCheckRepository
	Hub           *ws.Hub
}

func NewRouter(deps RouterDeps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(requestLogger())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Health check
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Handlers
	stackHandler := NewStackHandler(deps.StackSvc)
	versionHandler := NewVersionHandler(deps.StackSvc, deps.GitSvc, deps.DiffSvc)
	deployHandler := NewDeployHandler(deps.DeploySvc, deps.DeployRepo, deps.DeployLogRepo, deps.HealthRepo)
	schedHandler := NewScheduleHandler(deps.SchedSvc)
	depHandler := NewDependencyHandler(deps.DepSvc)
	wsHandler := NewWSHandler(deps.Hub)

	v1 := router.Group("/api/v1")
	{
		stacks := v1.Group("/stacks")
		{
			stacks.GET("", stackHandler.ListStacks)
			stacks.POST("", stackHandler.CreateStack)
			stacks.GET("/:id", stackHandler.GetStack)
			stacks.PUT("/:id", stackHandler.UpdateStack)
			stacks.DELETE("/:id", stackHandler.DeleteStack)

			stacks.GET("/:id/compose", stackHandler.GetCompose)
			stacks.PUT("/:id/compose", stackHandler.UpdateCompose)

			stacks.POST("/:id/start", stackHandler.StartStack)
			stacks.POST("/:id/stop", stackHandler.StopStack)
			stacks.POST("/:id/restart", stackHandler.RestartStack)
			stacks.POST("/:id/deploy", deployHandler.Deploy)

			stacks.GET("/:id/status", stackHandler.GetStatus)
			stacks.GET("/:id/logs", stackHandler.GetLogs)

			// Versions / Git
			stacks.GET("/:id/versions", versionHandler.ListVersions)
			stacks.GET("/:id/versions/:hash", versionHandler.GetVersion)
			stacks.GET("/:id/versions/:hash/diff", versionHandler.GetVersionDiff)
			stacks.POST("/:id/versions/:hash/rollback", versionHandler.Rollback)

			stacks.GET("/:id/diff", versionHandler.GetWorkingDiff)
			stacks.POST("/:id/diff", versionHandler.GetWorkingDiff)
			stacks.GET("/:id/diff/:from/:to", versionHandler.GetDiffBetween)

			// Schedules and queued updates
			stacks.GET("/:id/schedules", schedHandler.ListByStack)
			stacks.POST("/:id/schedules", schedHandler.Create)
			stacks.GET("/:id/queue", schedHandler.ListQueuedUpdates)
			stacks.POST("/:id/queue", schedHandler.QueueUpdate)

			// Dependencies
			stacks.GET("/:id/dependencies", depHandler.ListDependencies)
			stacks.POST("/:id/dependencies", depHandler.AddDependency)
			stacks.GET("/:id/dependents", depHandler.ListDependents)
			stacks.DELETE("/:id/dependencies/:depId", depHandler.RemoveDependency)
		}

		deployments := v1.Group("/deployments")
		{
			deployments.GET("", deployHandler.ListDeployments)
			deployments.GET("/:id", deployHandler.GetDeployment)
			deployments.GET("/:id/logs", deployHandler.GetDeploymentLogs)
			deployments.GET("/:id/health", deployHandler.GetDeploymentHealth)
			deployments.POST("/:id/cancel", deployHandler.CancelDeploy)
		}

		schedules := v1.Group("/schedules")
		{
			schedules.GET("", schedHandler.ListAll)
			schedules.GET("/:id", schedHandler.Get)
			schedules.PUT("/:id", schedHandler.Update)
			schedules.DELETE("/:id", schedHandler.Delete)
			schedules.GET("/:id/next", schedHandler.NextWindow)
		}

		v1.GET("/queued-updates", schedHandler.ListAllQueuedUpdates)
		v1.DELETE("/queued-updates/:id", schedHandler.CancelQueuedUpdate)

		dependencies := v1.Group("/dependencies")
		{
			dependencies.GET("/graph", depHandler.GetGraph)
		}

		// WebSocket
		v1.GET("/ws/events", wsHandler.Events)
	}

	// Serve frontend (SPA)
	if StaticFS != nil {
		router.NoRoute(spaHandler())
	}

	return router
}

func spaHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Don't serve SPA for API routes
		if strings.HasPrefix(path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		// Try to serve the file
		f, err := StaticFS.Open(strings.TrimPrefix(path, "/"))
		if err == nil {
			f.Close()
			c.FileFromFS(path, http.FS(StaticFS))
			return
		}

		// Fallback to index.html for SPA routing
		c.FileFromFS("/", http.FS(StaticFS))
	}
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		status := c.Writer.Status()
		log.Debug().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", status).
			Msg("request")
	}
}
