# Proposed In-App Notification Implementation

## Introduction

This document outlines a detailed plan to enhance the existing alert and WebSocket mechanism into a robust, persistent, and user-specific in-app notification service. It addresses the current gaps identified in the system, providing technical tasks and sub-tasks for implementation.

## Current State (Summary)

Currently, the system triggers `Alert` events based on predefined conditions (e.g., low stock). These alerts are persisted in the database (`domain.Alert`) and published to RabbitMQ. A WebSocket server exists, broadcasting messages to all connected clients. However, there is no dedicated `Notification` model for user-specific messages, no mechanism for offline delivery, no notification history for disconnected users, and no read/unread status tracking.

## Implemented Enhancements (as of current development)

### 1. Persistent User-Specific Notifications & History

To provide a comprehensive notification experience, we have implemented a dedicated `Notification` entity that stores messages for individual users and allows for historical access.

#### Task 1.1: Created Database Models

*   **Description:** Defined new GORM models for `Notification` and `AlertRoleSubscription`.
*   **Details:**
    *   **`Notification` struct in `internal/domain/models.go`:**
        ```go
        type Notification struct {
            gorm.Model
            UserID      uint   `gorm:"not null"`
            User        User
            Type        string `gorm:"not null"` // e.g., "ALERT", "SYSTEM", "PROMOTIONAL"
            Title       string `gorm:"not null"`
            Message     string `gorm:"not null"`
            Payload     string // JSON string for additional data (e.g., productID, orderID)
            IsRead      bool   `gorm:"default:false"`
            ReadAt      *time.Time
            TriggeredAt time.Time
        }
        ```
    *   **`AlertRoleSubscription` struct in `internal/domain/models.go`:**
        ```go
        // AlertRoleSubscription links an alert type to a user role.
        type AlertRoleSubscription struct {
            gorm.Model
            AlertType string `gorm:"not null;uniqueIndex:idx_alert_role"`
            Role      string `gorm:"not null;uniqueIndex:idx_alert_role"`
        }
        ```
    *   **Database migration:** A migration would be required to add the `notifications` and `alert_role_subscriptions` tables.

#### Task 1.2: Updated Alert Handling Logic

*   **Description:** Modified the `handleAlertDelivery` function in `cmd/server/main.go` to use the role-based subscription model for targeted notification delivery.
*   **Details:**
    *   When an `alert.triggered` message is received:
        1.  The `alert_type` is extracted from the payload.
        2.  The `alert_role_subscriptions` table is queried to find all roles subscribed to this `alert_type`.
        3.  Users with these subscribed roles are fetched from the database.
        4.  For each targeted user:
            *   Their `UserNotificationSettings` are checked.
            *   If email notifications are enabled, an email is sent.
            *   A `domain.Notification` record is created and persisted in the database.
            *   A user-specific WebSocket message is sent using `app.hub.SendToUser()`.

#### Task 1.3: Implemented Notification Repository

*   **Description:** Created a dedicated repository for `Notification` operations to abstract database interactions.
*   **Details:**
    *   `internal/repository/notification_repository.go` contains the `NotificationRepository` interface and its GORM implementation with methods like `CreateNotification`, `GetNotificationsByUserID`, `MarkNotificationAsRead`, etc.

### 2. Role-Based Subscription Management

To make the notification system flexible, administrators can now manage which roles receive which alerts via dedicated API endpoints.

#### Task 2.1: Created Subscription Management API Endpoints

*   **Description:** Exposed API endpoints for administrators to manage alert-role subscriptions.
*   **Details:**
    *   **`POST /api/v1/alerts/subscriptions`:** Creates a new subscription.
        *   **Body:** `{ "alertType": "LOW_STOCK", "role": "Manager" }`
        *   **Authorization:** Admin only.
    *   **`GET /api/v1/alerts/subscriptions`:** Lists all current alert-role subscriptions.
        *   **Authorization:** Admin only.
    *   **`DELETE /api/v1/alerts/subscriptions/:id`:** Deletes a subscription by its ID.
        *   **Authorization:** Admin only.
    *   **Implementation:** Handlers are in `internal/handlers/notification_subscriptions.go` and routes are defined in `internal/router/router.go`.

### 3. User-Facing Notification API (Pending)

Users need API endpoints to retrieve and interact with their notifications.

#### Task 3.1: Create Notification API Endpoints

*   **Description:** Expose API endpoints for users to retrieve and manage their notifications.
*   **Sub-tasks:**
    *   **3.1.1 `GET /api/v1/users/{userId}/notifications`:**
        *   **Purpose:** Retrieve a list of notifications for a specific user.
        *   **Parameters:** `userId` (path), `isRead` (query, optional, filter by read status), `limit`, `offset` (for pagination).
        *   **Authorization:** User can only retrieve their own notifications (or an Admin).
    *   **3.1.2 `GET /api/v1/users/{userId}/notifications/unread/count`:**
        *   **Purpose:** Get the count of unread notifications for a user.
        *   **Authorization:** User can only retrieve their own count.
    *   **3.1.3 `PATCH /api/v1/users/{userId}/notifications/{notificationId}/read`:**
        *   **Purpose:** Mark a specific notification as read.
        *   **Authorization:** User can only mark their own notifications as read.
    *   **3.1.4 `PATCH /api/v1/users/{userId}/notifications/read-all`:**
        *   **Purpose:** Mark all unread notifications for a user as read.
        *   **Authorization:** User can only mark their own notifications as read.

### 4. Real-time User-Specific Delivery via WebSockets

The existing WebSocket hub has been modified to support sending messages to specific users rather than broadcasting to all.

#### Task 4.1: Modified WebSocket Hub for User-Specific Channels

*   **Description:** Updated the `websocket.Hub` to manage clients based on their authenticated user ID.
*   **Details:**
    *   **`Client` with `UserID`:** `websocket.Client` struct now includes `UserID uint`.
    *   **`Hub` client management:** `Hub.clients` changed to `map[uint]map[*Client]bool`.
    *   **`Hub.SendToUser(userID uint, message interface{})`:** New method implemented to send messages to all active clients for a specific user.

#### Task 4.2: Integrated Notification Delivery with WebSocket

*   **Description:** When a new `Notification` is created, it is sent in real-time to the target user via WebSocket.
*   **Details:**
    *   `handleAlertDelivery` now calls `app.hub.SendToUser()` after creating a `domain.Notification` record.

#### Task 4.3: Handle WebSocket Connection/Disconnection (Pending)

*   **Description:** Ensure that when a user connects, any unread notifications are delivered.
*   **Sub-tasks:**
    *   **4.3.1 On `websocket.Client` connection:** After authentication, query the database for any unread `Notification`s for that `UserID` and send them to the newly connected client.

### 5. Frontend Integration & Testing (Conceptual)

*   **Frontend:** The frontend will need to connect to the WebSocket, display notifications, and use the new API endpoints to manage read status.
*   **Testing:** A comprehensive test suite should be developed, including unit tests for repositories and handlers, and integration tests for the API endpoints and WebSocket delivery.

## Overall Considerations

*   **Security:** All new endpoints must have proper authorization middleware.
*   **Performance:** Database queries for finding users based on roles and subscriptions should be optimized.
*   **User Preferences:** The `UserNotificationSettings` model can be extended to include preferences for in-app notifications (e.g., enable/disable specific alert types for a user, overriding the role-based subscription).
