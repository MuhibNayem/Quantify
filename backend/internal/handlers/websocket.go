package handlers

import (
	"fmt"
	"inventory/backend/internal/auth"
	"inventory/backend/internal/repository"
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
func ServeWs(hub *ws.Hub, c *gin.Context, notificationRepo repository.NotificationRepository, roleRepo repository.RoleRepository) {
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

	// Fetch permissions for the role
	role, err := roleRepo.GetRoleByName(claims.Role)
	if err != nil {
		logrus.Errorf("Failed to fetch role permissions for user %d (role %s): %v", claims.UserID, claims.Role, err)
		c.JSON(http.StatusForbidden, gin.H{"error": "Failed to resolve permissions"})
		return
	}

	permissions := make(map[string]bool)
	for _, p := range role.Permissions {
		permissions[p.Name] = true
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// Log the error if the WebSocket upgrade fails
		fmt.Printf("Failed to upgrade to websocket: %v\n", err)
		return
	}
	client := &ws.Client{
		Hub:         hub,
		Conn:        conn,
		Send:        make(chan []byte, 256),
		UserID:      claims.UserID,
		Permissions: permissions,
	}
	client.Hub.Register <- client

	// Send unread notifications upon connection
	go func() {
		unreadNotifications, err := notificationRepo.GetNotificationsByUserID(client.UserID, new(bool), 50, 0) // Fetch latest 50 unread
		if err != nil {
			logrus.Errorf("Failed to fetch unread notifications for user %d: %v", client.UserID, err)
			return
		}
		for _, notification := range unreadNotifications {
			hub.SendToUser(client.UserID, notification)
		}
	}()

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()
}
