# Frontend Feature Implementation Plan (Exhaustive & Complete)

This document outlines the implementation plan for all missing features in the frontend application, based on a thorough analysis of the backend API documentation. The design for these new features should be consistent with the existing modern, colorful, and vibrant aesthetic.

## 1. Detailed Resource Views

### 1.1. Feature Overview
Create dedicated modals to display detailed information for individual resources. This provides a more focused and comprehensive view than the existing table rows, without requiring navigation to a new page.

### 1.2. Design & UI/UX
*   **Layout:** Use a `Dialog` (modal) component for displaying the details. The modal should have a modern, clean, and soothing aesthetic.
*   **Interaction:** The modal for a specific resource will be opened by clicking on the corresponding row in the resource's table. This makes discovery easy and intuitive. To prevent the modal from opening when an action button (e.g., "Edit", "Delete") within the row is clicked, event propagation must be stopped on those buttons (e.g., using `event.stopPropagation()`).
*   **Display:** Inside the modal, use a combination of `Card`s, `DescriptionList`s (a new component might be needed), and `Table`s to present the information in a structured and visually appealing way.
*   **Color Palette:** Use the color palette of the parent section (e.g., `sky` and `blue` for products, `violet` and `purple` for suppliers).

### 1.3. Subtasks
*   [x] **Task 1: Create a Reusable Detail Modal Component:**
    *   [x] Develop a generic modal component that can be adapted for different resources.
    *   [x] This component will handle fetching data based on a provided resource ID and endpoint.
*   [x] **Task 2: Integrate Modal for Products:**
    *   [x] On the "Catalog" page, make the product table rows clickable.
    *   [x] On row click, open the detail modal and fetch product details using `GET /products/{id}`.
    *   [x] The modal should include details like stock history (`GET /products/{productId}/stock/history`).
*   [ ] **Task 3: Integrate Modal for Categories & Sub-Categories:**
    *   [x] Make the category and sub-category table rows clickable.
    *   [x] On row click, open the detail modal and fetch details using `GET /categories/{id}` and `GET /subcategories/{id}`.
*   [x] **Task 4: Integrate Modal for Suppliers:**
    *   [x] Make the supplier table rows clickable.
    *   [x] On row click, open the detail modal and fetch details using `GET /suppliers/{id}`.
    *   [x] The modal should include the Supplier Performance Report.
*   [x] **Task 5: Integrate Modal for Locations:**
    *   [x] Make the location table rows clickable.
    *   [x] On row click, open the detail modal and fetch details using `GET /locations/{id}`.
*   [x] **Task 6: Integrate Modal for Alerts:**
    *   [x] On the "Alerts" page, make the alert table rows clickable.
    *   [x] On row click, open the detail modal and fetch details using `GET /alerts/{alertId}`.
*   [x] **Task 7: Integrate Modal for Forecasts:**
    *   [x] On the "Intelligence" page, make the forecast table rows clickable.
    *   [x] On row click, open the detail modal and fetch details using `GET /replenishment/forecast/{forecastId}`.
*   [x] **Task 8: Integrate Modal for Purchase Orders:**
    *   [x] On the "Purchase Orders" page, make the PO table rows clickable.
    *   [x] On row click, open the detail modal and fetch details using `GET /purchase-orders/{poId}`.

## 2. Enhanced Search & Filtering

### 2.1. Feature Overview
Improve the search functionality to allow users to find resources by alternative keys, such as name or SKU.

### 2.2. Design & UI/UX
*   [ ] **UI Location:** Enhance the existing search bars on the "Catalog" and other pages.
*   [ ] **Functionality:** Add a dropdown or radio buttons to allow users to select the search key (e.g., "Search by ID", "Search by Name", "Search by SKU").

### 2.3. Subtasks
*   [x] **Task 1: Enhance Product Search:**
    *   [x] Add an option to search for products by SKU using `GET /products/sku/{sku}`.
*   [x] **Task 2: Enhance Category Search:**
    *   [x] Add an option to search for categories by name using `GET /categories/name/{name}`.
*   [x] **Task 3: Enhance Supplier Search:**
    *   [x] Add an option to search for suppliers by name using `GET /suppliers/name/{name}`.
*   [ ] **Task 4: Enhance Time Tracking Search:**
    *   [ ] In the manager's view of the Time Tracking page, add an option to search for a user's last entry by username using `GET /time-tracking/last-entry/username/{username}`.

## 3. CRM & Loyalty Management

### 3.1. Feature Overview
A new section in the application to manage customers and their loyalty points. This will involve creating, viewing, updating, and deleting customer profiles, as well as managing their loyalty accounts.

### 3.2. Design & UI/UX
*   [ ] **New Route:** Create a new route at `/crm`.
*   [ ] **Layout:** Use a two-column layout. The left column will be a searchable list of customers, and the right column will display the details of the selected customer, including their loyalty account.
*   [ ] **Color Palette:** Use a new color palette to distinguish the CRM section. A warm palette based on `amber`, `yellow`, and `lime` would fit well with the existing design.
*   [ ] **Components:**
    *   [ ] Use `Card` components for customer details and loyalty information.
    *   [ ] Use a `Table` to display the list of customers.
    *   [ ] Use `Input` and `Button` components for forms.
    *   [ ] Use `Skeleton` components for loading states.
    *   [ ] Use `lucide-svelte` icons for actions (e.g., `UserPlus`, `Edit`, `Trash2`, `Star`).

### 3.3. Subtasks
*   [ ] **Task 1: Create CRM Route & Page:**
    *   [ ] Create a new directory `client/src/routes/crm`.
    *   [ ] Create a `+page.svelte` file inside the new directory.
    *   [ ] Add a link to the CRM page in the `Sidebar.svelte` component.
*   [ ] **Task 2: Implement Customer List:**
    *   [ ] Fetch and display a list of customers using the `GET /crm/customers/{identifier}` endpoint (or a new endpoint for listing all customers if available).
    *   [ ] Implement a search bar to filter customers by name, email, or phone.
*   [ ] **Task 3: Implement Customer Details View:**
    *   [ ] When a customer is selected from the list, display their details in a `Card`.
    *   [ ] Implement forms to create, update, and delete customers using the `POST /crm/customers`, `PUT /crm/customers/{userId}`, and `DELETE /crm/customers/{userId}` endpoints.
*   [ ] **Task 4: Implement Loyalty Management:**
    *   [ ] Display the selected customer's loyalty account information using the `GET /crm/loyalty/{userId}` endpoint.
    *   [ ] Implement forms to add and redeem loyalty points using the `POST /crm/loyalty/{userId}/points` and `POST /crm/loyalty/{userId}/redeem` endpoints.

## 4. Notification Center

### 4.1. Feature Overview
A UI to display user-specific notifications with read/unread status. This will allow users to stay informed about important events in the system.

### 4.2. Design & UI/UX
*   [ ] **UI Element:** Add a notification bell icon to the `Topbar.svelte` component.
*   [ ] **Dropdown Panel:** When the bell icon is clicked, a dropdown panel will appear, listing the user's notifications.
*   [ ] **Color Palette:** Use the existing `slate` and `blue` palette for consistency.
*   [ ] **Components:**
    *   [ ] Use a `DropdownMenu` or `Popover` component for the notification panel.
    *   [ ] Use a `ScrollArea` to make the notification list scrollable.
    *   [ ] Use `Badge` components to indicate unread notifications.
    *   [ ] Use `lucide-svelte` icons (`Bell`, `CheckCheck`).

### 4.3. Subtasks
*   [ ] **Task 1: Add Notification Icon to Topbar:**
    *   [ ] Add a bell icon with an unread count badge to the `Topbar.svelte` component.
    *   [ ] Fetch the unread count using the `GET /users/{userId}/notifications/unread/count` endpoint.
*   [ ] **Task 2: Implement Notification Panel:**
    *   [ ] Create a new component for the notification panel.
    *   [ ] Fetch and display a list of notifications using the `GET /users/{userId}/notifications` endpoint.
    *   [ ] Implement actions to mark notifications as read (`PATCH /users/{userId}/notifications/{notificationId}/read`) or mark all as read (`PATCH /users/{userId}/notifications/read-all`).
*   [ ] **Task 3: Real-time Updates:**
    *   [ ] Use the existing WebSocket connection to receive real-time notifications and update the unread count and notification list.

## 5. Full Purchase Order (PO) Management

### 5.1. Feature Overview
A new section to manage the entire lifecycle of a purchase order, from creation to cancellation.

### 5.2. Design & UI/UX
*   [ ] **New Route:** Create a new route at `/purchase-orders`.
*   [ ] **Layout:** Use a table-based layout to display a list of purchase orders. A modal or a separate page can be used to view the details of a PO.
*   [ ] **Color Palette:** Use a professional and clean palette based on `slate`, `gray`, and `blue`.
*   [ ] **Components:**
    *   [ ] Use a `Table` to list all purchase orders.
    *   [ ] Use `Dialog` or a new page to display PO details.
    *   [ ] Use `Button` components for actions like "Approve", "Send", "Receive", and "Cancel".
    *   [ ] Use `Badge` components to indicate the status of a PO (e.g., "Draft", "Approved", "Sent", "Partially Received", "Completed", "Cancelled").

### 5.3. Subtasks
*   [ ] **Task 1: Create PO Route & Page:**
    *   [ ] Create a new directory `client/src/routes/purchase-orders`.
    *   [ ] Create a `+page.svelte` file inside the new directory.
    *   [ ] Add a link to the PO page in the `Sidebar.svelte` component.
*   [ ] **Task 2: Implement PO List:**
    *   [ ] Fetch and display a list of purchase orders using a new endpoint (e.g., `GET /purchase-orders`).
    *   [ ] Implement filtering and sorting options for the PO list.
*   [ ] **Task 3: Implement PO Details View:**
    *   [ ] Display the details of a selected PO, including its items.
*   [ ] **Task 4: Implement PO Actions:**
    *   [ ] Implement buttons to approve, send, receive, and cancel POs using the respective API endpoints:
        *   `POST /purchase-orders/{poId}/approve`
        *   `POST /purchase-orders/{poId}/send`
        *   `POST /purchase-orders/{poId}/receive`
        *   `POST /purchase-orders/{poId}/cancel`
    *   [ ] Create a form for receiving goods against a PO.

## 6. Supplier Performance Report

### 6.1. Feature Overview
A UI to display a performance report for a selected supplier.

### 6.2. Design & UI/UX
*   [ ] **UI Location:** Add a "Performance" button or tab to the existing supplier management UI in the "Catalog" page.
*   [ ] **Display:** When the "Performance" button is clicked, display the report in a `Dialog` or a new section on the page.
*   [ ] **Color Palette:** Use the existing `violet` and `purple` palette from the suppliers section.
*   [ ] **Components:**
    *   [ ] Use `Card` components to display key performance metrics (e.g., on-time delivery rate, average lead time).
    *   [ ] Use charts (e.g., from `chart.js` or a Svelte-based charting library) to visualize performance trends.

### 6.3. Subtasks
*   [ ] **Task 1: Add Performance Button:**
    *   [ ] Add a "Performance" button to the suppliers table in `client/src/routes/catalog/+page.svelte`.
*   [ ] **Task 2: Implement Performance View:**
    *   [ ] Create a new component to display the supplier performance report.
    *   [ ] Fetch the performance data using the `GET /suppliers/{id}/performance` endpoint.
    *   [ ] Display the data using `Card`s and charts.

## 7. Time Tracking

### 7.1. Feature Overview
A UI for employees to clock in and out, and for managers to view time tracking data.

### 7.2. Design & UI/UX
*   [ ] **New Route:** Create a new route at `/time-tracking`.
*   [ ] **Layout:**
    *   [ ] For "Staff" role: A simple interface with "Clock In" and "Clock Out" buttons and a display of the last time clock entry.
    *   [ ] For "Manager" and "Admin" roles: A table-based view of all time clock entries, with filtering and exporting options.
*   [ ] **Color Palette:** Use a calm and focused palette based on `teal` and `cyan`.
*   [ ] **Components:**
    *   [ ] Use `Button` components for clocking in and out.
    *   [ ] Use a `Card` to display the user's last time clock entry.
    *   [ ] Use a `Table` to display time clock entries for managers.

### 7.3. Subtasks
*   [ ] **Task 1: Create Time Tracking Route & Page:**
    *   [ ] Create a new directory `client/src/routes/time-tracking`.
    *   [ ] Create a `+page.svelte` file inside the new directory.
    *   [ ] Add a link to the Time Tracking page in the `Sidebar.svelte` component.
*   [ ] **Task 2: Implement Clock In/Out for Staff:**
    *   [ ] Create a simple UI for staff to clock in and out using the `POST /time-tracking/clock-in/{userId}` and `POST /time-tracking/clock-out/{userId}` endpoints.
    *   [ ] Display the user's last time clock entry using the `GET /time-tracking/last-entry/{userId}` endpoint.
*   [ ] **Task 3: Implement Time Clock View for Managers:**
    *   [ ] Create a table to display all time clock entries (requires a new API endpoint, e.g., `GET /time-tracking/entries`).
    *   [ ] Implement filtering by user and date range.

## 8. Alert Settings & Subscriptions

### 8.1. Feature Overview
A UI to configure product-specific alert thresholds and user notification preferences.

### 8.2. Design & UI/UX
*   [ ] **UI Location:** This can be integrated into the existing "Alerts" page (`/alerts`).
*   [ ] **Layout:** Add two new `Card` components to the "Alerts" page: one for "Product Alert Settings" and one for "User Notification Preferences".
*   [ ] **Color Palette:** Use the existing `amber` and `orange` palette from the alerts section.
*   [ ] **Components:**
    *   [ ] Use `Card` components to group the settings.
    *   [ ] Use `Input` and `Button` components for the forms.
    *   [ ] Use `Switch` or `Checkbox` components for enabling/disabling notifications.

### 8.3. Subtasks
*   [ ] **Task 1: Add Product Alert Settings Form:**
    *   [ ] Add a form to `client/src/routes/alerts/+page.svelte` to configure low-stock, overstock, and expiry alert thresholds for a product.
    *   [ ] Use the `PUT /products/{productId}/alert-settings` endpoint to save the settings.
*   [ ] **Task 2: Add User Notification Preferences Form:**
    *   [ ] Add a form to `client/src/routes/alerts/+page.svelte` to configure email and SMS notification preferences for a user.
    *   [ ] Use the `PUT /users/{userId}/notification-settings` endpoint to save the settings.
*   [ ] **Task 3: Add Alert Role Subscription Management:**
    *   [ ] For "Admin" users, add a section to manage alert role subscriptions.
    *   [ ] Use a `Table` to list existing subscriptions (`GET /alerts/subscriptions`).
    *   [ ] Implement forms to create (`POST /alerts/subscriptions`) and delete (`DELETE /alerts/subscriptions/{id}`) subscriptions.

## 9. Product Archiving

### 9.1. Feature Overview
Provide a dedicated and clear way for users to archive products.

### 9.2. Design & UI/UX
*   [ ] **UI Location:** Add an "Archive" button to the product actions in the "Catalog" page's table.
*   [ ] **Functionality:** When the "Archive" button is clicked, confirm the action with the user and then send a request to the `PATCH /products/{id}/archive` endpoint.

### 9.3. Subtasks
*   [ ] **Task 1: Add Archive Button:**
    *   [ ] Add an "Archive" button to the products table in `client/src/routes/catalog/+page.svelte`.
*   [ ] **Task 2: Implement Archive Logic:**
    *   [ ] Create a function to handle the archive action, including a confirmation dialog.

## 10. Asynchronous Report Generation

### 10.1. Feature Overview
Update the "Intelligence" page to use the asynchronous export endpoints for generating reports.

### 10.2. Design & UI/UX
*   [ ] **UI Location:** Modify the "Reports" section of the `client/src/routes/intelligence/+page.svelte` page.
*   [ ] **Functionality:**
    *   [ ] When a user requests a report, call the appropriate export endpoint (e.g., `POST /reports/sales-trends/export`).
    *   [ ] Display the job ID to the user and periodically check the job status using `GET /reports/jobs/{jobId}`.
    *   [ ] When the job is complete, provide a "Download" button that links to the report file using `GET /reports/download/{jobId}`.

### 10.3. Subtasks
*   [ ] **Task 1: Update Report Generation Logic:**
    *   [ ] Modify the `runReport` function in `client/src/routes/intelligence/+page.svelte` to use the export endpoints.
*   [ ] **Task 2: Implement Job Status Tracking:**
    *   [ ] Create a new UI element to display the status of report generation jobs.
*   [ ] **Task 3: Implement Report Downloading:**
    *   [ ] Add a "Download" button that appears when a report is ready.

## 11. Stock Adjustment History

### 11.1. Feature Overview
Provide a way for users to view the history of stock adjustments for a product.

### 11.2. Design & UI/UX
*   [ ] **UI Location:** Integrate this into the "Product Detail Page" (from section 1).
*   [ ] **Display:** Use a `Table` to display the stock adjustment history, including details like the type of adjustment, quantity, reason, and date.

### 11.3. Subtasks
*   [ ] **Task 1: Add History Table to Product Detail Page:**
    *   [ ] In the new `client/src/routes/products/[id]/+page.svelte` file, add a `Table` component.
*   [ ] **Task 2: Implement History Fetching:**
    *   [ ] Fetch the stock adjustment history using the `GET /products/{productId}/stock/history` endpoint and display it in the table.
