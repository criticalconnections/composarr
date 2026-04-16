package handler

import (
	"net/http"

	ws "github.com/axism/composarr/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Permissive for local-host use; tighten if exposing externally
		return true
	},
}

type WSHandler struct {
	hub *ws.Hub
}

func NewWSHandler(hub *ws.Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

// Events godoc
// GET /api/v1/ws/events
// Upgrades the connection to a WebSocket and starts streaming events.
func (h *WSHandler) Events(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Warn().Err(err).Msg("ws upgrade failed")
		return
	}

	client := ws.NewClient(h.hub, conn)
	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}
