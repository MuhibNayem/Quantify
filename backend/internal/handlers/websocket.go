package handlers

import (
	"encoding/json"
	"fmt"
	"inventory/backend/internal/auth"
	"inventory/backend/internal/config"
	"inventory/backend/internal/repository"
	"net/http"
	"strings"
	"time"

	ws "inventory/backend/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func buildOriginChecker(allowed []string) func(r *http.Request) bool {
	set := make(map[string]struct{}, len(allowed))
	for _, origin := range allowed {
		set[strings.TrimSpace(origin)] = struct{}{}
	}
	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return false
		}
		_, ok := set[origin]
		return ok
	}
}

// authMessageTimeout is the maximum time the server waits for the client to
// send a valid first-message auth after the WebSocket upgrade. If no valid
// auth arrives within this window, the connection is closed.
const authMessageTimeout = 5 * time.Second

// authMessageType is the expected value of the `type` field in the client's
// first WebSocket message. Any other message type is rejected.
const authMessageType = "auth"

// parseAuthMessage extracts the JWT token from a client's first-message auth
// payload. The expected shape is {"type":"auth","token":"<JWT>"}.
func parseAuthMessage(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("empty auth message")
	}
	var msg struct {
		Type  string `json:"type"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		return "", fmt.Errorf("invalid auth message JSON: %w", err)
	}
	if msg.Type != authMessageType {
		return "", fmt.Errorf("expected message type %q, got %q", authMessageType, msg.Type)
	}
	if msg.Token == "" {
		return "", fmt.Errorf("missing token in auth message")
	}
	return msg.Token, nil
}

// ServeWs handles websocket requests from the peer.
//
// Authentication is performed via a first-message exchange, NOT via the
// `?token=` query parameter. The legacy `?token=` query parameter is ignored
// entirely: if present, it is logged as a deprecation warning and discarded.
// Clients must send `{"type":"auth","token":"<JWT>"}` as the first WebSocket
// message after the upgrade. If no valid auth message arrives within
// authMessageTimeout, the connection is closed.
func ServeWs(hub *ws.Hub, c *gin.Context, notificationRepo repository.NotificationRepository, roleRepo repository.RoleRepository, cfg *config.Config) {
	// The `?token=` query parameter is deprecated and is no longer accepted as
	// an authentication mechanism. Logged here so operators can spot any
	// clients that haven't been updated yet. The value is intentionally NOT
	// used for auth.
	if tokenQuery := c.Query("token"); tokenQuery != "" {
		logrus.Warn("WebSocket connection supplied a deprecated ?token query parameter; ignoring it. The client must send first-message auth within 5s.")
	}

	requestUpgrader := upgrader
	requestUpgrader.CheckOrigin = buildOriginChecker(cfg.WSAllowedOrigins)
	conn, err := requestUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// Log the error if the WebSocket upgrade fails
		logrus.Errorf("Failed to upgrade to websocket: %v", err)
		return
	}

	// First-message auth. Read with a deadline so a client that never sends
	// auth cannot hold the connection open indefinitely.
	if err := conn.SetReadDeadline(time.Now().Add(authMessageTimeout)); err != nil {
		logrus.Errorf("Failed to set WebSocket auth read deadline: %v", err)
		conn.Close()
		return
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		logrus.Errorf("WebSocket first-message auth not received within %s (or read error): %v", authMessageTimeout, err)
		conn.Close()
		return
	}

	tokenString, err := parseAuthMessage(msg)
	if err != nil {
		logrus.Errorf("WebSocket first-message auth rejected: %v", err)
		conn.Close()
		return
	}

	claims, err := auth.ValidateJWT(tokenString)
	if err != nil {
		logrus.Errorf("Failed to validate WebSocket auth token: %v", err)
		conn.Close()
		return
	}

	// Fetch permissions for the role
	role, err := roleRepo.GetRoleByName(claims.Role)
	if err != nil {
		logrus.Errorf("Failed to fetch role permissions for user %d (role %s): %v", claims.UserID, claims.Role, err)
		conn.Close()
		return
	}

	permissions := make(map[string]bool)
	for _, p := range role.Permissions {
		permissions[p.Name] = true
	}

	// Auth succeeded: clear the read deadline so the normal ReadPump can
	// apply its own (longer) pong-based deadline.
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		logrus.Errorf("Failed to clear WebSocket read deadline after auth: %v", err)
		conn.Close()
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

	// Acknowledge successful auth to the client. Sent through the buffered
	// Send channel so WritePump delivers it.
	authOK, _ := json.Marshal(map[string]string{"type": "auth_ok"})
	select {
	case client.Send <- authOK:
	default:
		// Buffer full; the client will still see auth succeed via subsequent
		// server-pushed messages.
	}

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
