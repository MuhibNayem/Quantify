package handlers

import (
	"fmt"
	"inventory/backend/internal/auth"
	"net/http"

	ws "inventory/backend/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *ws.Hub, c *gin.Context) {
	tokenString := c.Query("token")
	if tokenString == "" {
		logrus.Error("No token provided for websocket connection")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token not provided"})
		return
	}

	claims, err := auth.ValidateJWT(tokenString)
	if err != nil {
		logrus.Errorf("Failed to validate token for websocket connection: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// Log the error if the WebSocket upgrade fails
		fmt.Printf("Failed to upgrade to websocket: %v\n", err)
		return
	}
	client := &ws.Client{Hub: hub, Conn: conn, Send: make(chan []byte, 256), UserID: claims.UserID}
	client.Hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()
}
