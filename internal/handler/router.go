package handler

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/axism/composarr/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// StaticFS holds the embedded frontend files. Set by main.go when available.
var StaticFS fs.FS

func NewRouter(stackSvc *service.StackService) *gin.Engine {
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

	// API routes
	stackHandler := NewStackHandler(stackSvc)

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

			stacks.GET("/:id/status", stackHandler.GetStatus)
			stacks.GET("/:id/logs", stackHandler.GetLogs)
		}
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
