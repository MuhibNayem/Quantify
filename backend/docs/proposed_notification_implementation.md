# Proposed In-App Notification Implementation

## Introduction

This document outlines a detailed plan to enhance the existing alert and WebSocket mechanism into a robust, persistent, and user-specific in-app notification service. It addresses the current gaps identified in the system, providing technical tasks and sub-tasks for implementation.

## Current State (Summary)

Currently, the system triggers `Alert` events based on predefined conditions (e.g., low stock). These alerts are persisted in the database (`domain.Alert`) and published to RabbitMQ. A WebSocket server exists, broadcasting messages to all connected clients. However, there is no dedicated `Notification` model for user-specific messages, no mechanism for offline delivery, no notification history for disconnected users, and no read/unread status tracking.

## Proposed Enhancements

### 1. Persistent User-Specific Notifications & History

To provide a comprehensive notification experience, we need a dedicated `Notification` entity that stores messages for individual users and allows for historical access.

#### Task 1.1: Create `Notification` Database Model

*   **Description:** Define a new GORM model for `Notification` to store individual notification messages.
*   **Sub-tasks:**
    *   **1.1.1 Define `Notification` struct in `internal/domain/models.go`:**
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
    *   **1.1.2 Run database migration:** Generate and execute a migration to add the `notifications` table.

#### Task 1.2: Update Alert Triggering Logic to Create Notifications

*   **Description:** Modify the `triggerAlert` function (or a new service layer function) to create `Notification` records for relevant users when an `Alert` is triggered.
*   **Sub-tasks:**
    *   **1.2.1 Identify target users:** Determine which users should receive a notification for a given alert (e.g., all admins, specific managers, users subscribed to product alerts). This might involve querying `UserNotificationSettings` or `User` roles.
    *   **1.2.2 Create `Notification` records:** For each target user, create a new `domain.Notification` entry in the database.
    *   **1.2.3 Modify `triggerAlert` in `internal/handlers/alerts.go`:**
        *   After creating `domain.Alert`, iterate through relevant users.
        *   For each user, create a `domain.Notification` record.
        *   The `Payload` field can store `AlertTriggeredPayload` marshaled to JSON.

#### Task 1.3: Implement Notification Repository

*   **Description:** Create a dedicated repository for `Notification` operations to abstract database interactions.
*   **Sub-tasks:**
    *   **1.3.1 Create `internal/repository/notification_repository.go`:**
        *   Define an interface `NotificationRepository` with methods like `CreateNotification`, `GetNotificationsByUserID`, `MarkNotificationAsRead`, etc.
        *   Implement a GORM-based `notificationRepository` struct.

#### Task 1.4: Create Notification API Endpoints

*   **Description:** Expose API endpoints for users to retrieve and manage their notifications.
*   **Sub-tasks:**
    *   **1.4.1 `GET /users/{userId}/notifications`:**
        *   **Purpose:** Retrieve a list of notifications for a specific user.
        *   **Parameters:** `userId` (path), `isRead` (query, optional, filter by read status), `limit`, `offset` (for pagination).
        *   **Response:** Array of `Notification` objects.
        *   **Authorization:** User can only retrieve their own notifications or an Admin can retrieve any user's.
    *   **1.4.2 `GET /users/{userId}/notifications/unread/count`:**
        *   **Purpose:** Get the count of unread notifications for a user.
        *   **Parameters:** `userId` (path).
        *   **Response:** `{ "count": int }`.
        *   **Authorization:** User can only retrieve their own count.
    *   **1.4.3 `PATCH /users/{userId}/notifications/{notificationId}/read`:**
        *   **Purpose:** Mark a specific notification as read.
        *   **Parameters:** `userId` (path), `notificationId` (path).
        *   **Request Body:** Optional, could be empty or `{ "is_read": true }`.
        *   **Response:** Updated `Notification` object.
        *   **Authorization:** User can only mark their own notifications as read.
    *   **1.4.4 `PATCH /users/{userId}/notifications/read-all`:**
        *   **Purpose:** Mark all unread notifications for a user as read.
        *   **Parameters:** `userId` (path).
        *   **Response:** Success message or count of updated notifications.
        *   **Authorization:** User can only mark their own notifications as read.

### 2. Real-time User-Specific Delivery via WebSockets

The existing WebSocket hub needs modification to support sending messages to specific users rather than broadcasting to all.

#### Task 2.1: Modify WebSocket Hub for User-Specific Channels

*   **Description:** Update the `websocket.Hub` to manage clients based on their authenticated user ID.
*   **Sub-tasks:**
    *   **2.1.1 Associate `Client` with `UserID`:**
        *   Modify `websocket.Client` struct to include `UserID uint`.
        *   When a WebSocket connection is established, authenticate the user (e.g., via JWT in query param or header) and set the `UserID` for the `Client`.
    *   **2.1.2 Update `Hub` client management:**
        *   Change `Hub.clients` from `map[*Client]bool` to `map[uint]map[*Client]bool` (mapping `UserID` to a map of their active connections). This allows a user to have multiple active WebSocket connections (e.g., from different devices).
        *   Modify `Hub.Register` and `Hub.unregister` to handle this new structure.
    *   **2.1.3 Implement `Hub.SendToUser(userID uint, message interface{})`:**
        *   This new method will iterate through all WebSocket clients associated with the given `userID` and send the message.

#### Task 2.2: Integrate Notification Delivery with WebSocket

*   **Description:** When a new `Notification` is created, send it in real-time to the target user(s) via WebSocket.
*   **Sub-tasks:**
    *   **2.2.1 Modify `triggerAlert` (or a new service function):** After creating `domain.Notification` records, call `websocket.Hub.SendToUser` for each target user, passing the newly created notification object.
    *   **2.2.2 Handle message format:** Ensure the message sent over WebSocket is a structured JSON object representing the `Notification`.

#### Task 2.3: Handle WebSocket Connection/Disconnection

*   **Description:** Ensure that when a user connects, any unread notifications are delivered, and when they disconnect, resources are cleaned up.
*   **Sub-tasks:**
    *   **2.3.1 On `websocket.Client` connection:**
        *   After authentication and associating `UserID`, query the database for any unread `Notification`s for that `UserID`.
        *   Send these unread notifications to the newly connected client.
    *   **2.3.2 On `websocket.Client` disconnection:**
        *   Ensure the client is properly unregistered from the `Hub` and its resources are cleaned up.

### 3. Notification Read/Unread Status

Tracking read status is essential for a good user experience and for filtering notifications.

#### Task 3.1: Add `IsRead` Field to `Notification` Model

*   **Description:** Already included in Task 1.1.1.
*   **Sub-tasks:**
    *   **3.1.1 Add `IsRead bool` and `ReadAt *time.Time` fields to `domain.Notification` struct.**
    *   **3.1.2 Update database migration.**

#### Task 3.2: Implement API Endpoint to Mark as Read

*   **Description:** Already included in Task 1.4.3 and 1.4.4.
*   **Sub-tasks:**
    *   **3.2.1 Implement `PATCH /users/{userId}/notifications/{notificationId}/read` handler.**
    *   **3.2.2 Implement `PATCH /users/{userId}/notifications/read-all` handler.**

#### Task 3.3: Update Notification Retrieval to Filter by Read Status

*   **Description:** Ensure the `GET /users/{userId}/notifications` endpoint can filter by `isRead` status.
*   **Sub-tasks:**
    *   **3.3.1 Modify `GetNotificationsByUserID` in `notification_repository.go` to accept an `isRead` filter parameter.**
    *   **3.3.2 Update `GET /users/{userId}/notifications` handler to use the `isRead` query parameter.**

### 4. Frontend Integration (Conceptual)

While this document focuses on the backend, a brief conceptual outline for frontend integration is useful.

#### Task 4.1: Implement WebSocket Client

*   **Description:** Connect to the backend WebSocket endpoint and handle incoming notification messages.
*   **Sub-tasks:**
    *   **4.1.1 Establish WebSocket connection:** Authenticate with JWT.
    *   **4.1.2 Listen for incoming messages:** Parse JSON notification objects.

#### Task 4.2: Display Notifications

*   **Description:** Render notifications in the UI, potentially as a notification bell icon with a count of unread messages, and a dropdown/modal for a list.
*   **Sub-tasks:**
    *   **4.2.1 Update unread count:** Increment/decrement based on incoming notifications and user actions.
    *   **4.2.2 Display notification list:** Show title, message, and timestamp.

#### Task 4.3: Mark Notifications as Read

*   **Description:** Provide UI elements to mark individual or all notifications as read.
*   **Sub-tasks:**
    *   **4.3.1 Call `PATCH /users/{userId}/notifications/{notificationId}/read` API.**
    *   **4.3.2 Call `PATCH /users/{userId}/notifications/read-all` API.**

### 5. Error Handling, Logging, and Testing

These are cross-cutting concerns that must be applied to all new and modified components.

#### Task 5.1: Implement Robust Error Handling

*   **Description:** Ensure all new and modified functions handle errors gracefully and return meaningful error messages.
*   **Sub-tasks:**
    *   **5.1.1 Use `appErrors` for API responses.**
    *   **5.1.2 Implement proper database transaction management for notification creation and updates.**

#### Task 5.2: Add Comprehensive Logging

*   **Description:** Log significant events, errors, and warnings related to notification processing and delivery.
*   **Sub-tasks:**
    *   **5.2.1 Use `logrus` for structured logging.**
    *   **5.2.2 Log notification creation, delivery attempts, and read status changes.**

#### Task 5.3: Develop Unit and Integration Tests

*   **Description:** Write tests to ensure the correctness and reliability of the new notification service.
*   **Sub-tasks:**
    *   **5.3.1 Unit tests for `NotificationRepository` methods.**
    *   **5.3.2 Integration tests for new API endpoints.**
    *   **5.3.3 Integration tests for WebSocket user-specific delivery.**
    *   **5.3.4 Tests for alert triggering leading to notification creation.**

## Overall Considerations

*   **Scalability:** Consider the volume of notifications and WebSocket connections. The proposed `Hub` modification might need further optimization for very high concurrency.
*   **Security:** Ensure proper authentication and authorization for all notification-related API endpoints and WebSocket connections. Prevent unauthorized access to other users' notifications.
*   **Performance:** Optimize database queries for fetching and updating notifications, especially for users with a large history.
*   **User Preferences:** Extend `UserNotificationSettings` to include preferences for in-app notifications (e.g., enable/disable, types of notifications).
*   **Notification Types:** Define a clear taxonomy for `Notification.Type` to allow for better filtering and UI presentation.
