# Proposed Report and Graph Generation Enhancements

## Introduction

This document analyzes the current report generation capabilities of the Quantify backend, identifies existing gaps, and proposes a detailed plan with technical tasks and sub-tasks to enhance the system into a more comprehensive and flexible reporting and graph generation solution.

## Current State (Summary)

The Quantify backend currently features a dedicated reporting module (`internal/handlers/reports.go`, `internal/services/reporting_service.go`, `internal/repository/reporting_repository.go`). It supports several fundamental report types: Sales Trends, Inventory Turnover, Profit Margin, and a Daily Sales Summary. Reports can be filtered by date range, category, and location, with Sales Trends offering grouping options.

Key strengths include:
*   **Asynchronous Export:** Long-running reports are processed via RabbitMQ jobs, preventing API blocking.
*   **Multiple Export Formats:** Sales Trends reports can be exported to CSV, PDF, and Excel.
*   **File Storage:** Generated reports are stored in MinIO.
*   **Job Management:** Endpoints exist to track job status and download completed reports.
*   **Caching:** Report data is cached using Redis to improve performance.

However, the system is **not sufficient** for advanced analytics and flexible reporting due to several limitations.

## Identified Gaps

1.  **Limited Report Types:** The existing reports, while essential, do not cover the full spectrum of business intelligence needs for an inventory management system.
2.  **No Generic Report Generation:** Adding new report types currently requires significant code changes across handlers, services, and repositories. There's no flexible mechanism for defining or generating custom reports.
3.  **Basic Graph Generation Support:** The backend provides raw data, expecting the frontend to handle all visualization. It doesn't offer structured "chart-ready" data.
4.  **No Ad-hoc Querying/Custom Reports:** Users cannot define or save their own complex report criteria beyond the predefined filters.
5.  **Performance for Complex Queries:** While caching helps, the underlying data aggregation for some reports (e.g., `getInventoryValueAt`) can be inefficient for large datasets or long periods.
6.  **Lack of Report Scheduling UI/Management:** There's no exposed API or UI for users to schedule custom reports or manage existing scheduled tasks.
7.  **Limited Error Detail in Async Jobs:** Errors during asynchronous report generation might lack sufficient detail for effective debugging or user feedback.
8.  **Granular Authorization:** While general API authorization exists, specific authorization for *which user can access which report data* (e.g., location-specific data for a store manager) needs explicit implementation.

## Proposed Enhancements

### 1. Expand Report Types and Introduce a Report Definition Framework

To move beyond hardcoded reports, a framework for defining and generating various report types is needed.

#### Task 1.1: Define New Report Types (Initial Set)

*   **Description:** Implement a few additional, commonly requested report types to demonstrate the framework's flexibility.
*   **Sub-tasks:**
    *   **1.1.1 Stock Aging Report:** Identifies inventory that has been in stock for a long time.
    *   **1.1.2 Dead Stock Report:** Lists items with no sales over a defined period.
    *   **1.1.3 Supplier Performance Report:** Analyzes supplier delivery times, quality, etc.
    *   **1.1.4 Purchase Order Analysis Report:** Tracks PO fulfillment, costs, and discrepancies.

#### Task 1.2: Implement a Report Definition Model

*   **Description:** Create a database model to store metadata about report definitions, potentially allowing for user-defined custom reports in the future.
*   **Sub-tasks:**
    *   **1.2.1 Define `ReportDefinition` struct in `internal/domain/models.go`:**
        ```go
        type ReportDefinition struct {
            gorm.Model
            Name        string `gorm:"uniqueIndex;not null"`
            Description string
            Type        string `gorm:"not null"` // e.g., "SALES_TRENDS", "STOCK_AGING", "CUSTOM"
            Config      string // JSON string for report-specific configuration (e.g., default filters, columns)
            IsPublic    bool   `gorm:"default:false"` // Can all users access?
            CreatedBy   uint
            // Add fields for authorization if needed (e.g., AllowedRoles []string)
        }
        ```
    *   **1.2.2 Run database migration.**

#### Task 1.3: Generic Report Generation Service

*   **Description:** Refactor the `ReportingService` to use a more generic approach for report generation based on `ReportDefinition`.
*   **Sub-tasks:**
    *   **1.3.1 Create a `ReportGenerator` interface:** Define methods like `Generate(config string) (interface{}, error)` and `Export(data interface{}, format string) (io.Reader, error)`.
    *   **1.3.2 Implement specific `ReportGenerator` for each report type:** (e.g., `SalesTrendsGenerator`, `StockAgingGenerator`).
    *   **1.3.3 `ReportingService` orchestrates:** The service will load the `ReportDefinition`, select the appropriate `ReportGenerator`, and execute it.

### 2. Enhanced Data Aggregation and Querying

Improve the efficiency and flexibility of data retrieval for reports.

#### Task 2.1: Implement Flexible Query Builder

*   **Description:** Develop a more dynamic query builder to handle varied report requirements and filters.
*   **Sub-tasks:**
    *   **2.1.1 Create a utility for dynamic GORM query construction:** This could parse filter parameters (e.g., `field=value`, `field_gt=value`, `field_in=value1,value2`) into GORM `Where` clauses.
    *   **2.1.2 Apply to existing and new report repository methods:** Replace hardcoded `Where` clauses with dynamic ones.

#### Task 2.2: Optimize Database Indexes

*   **Description:** Review and add necessary database indexes to improve report query performance.
*   **Sub-tasks:**
    *   **2.2.1 Analyze common report query patterns:** Identify frequently filtered and joined columns.
    *   **2.2.2 Add indexes:** Specifically for `adjusted_at`, `product_id`, `category_id`, `location_id` in `stock_adjustments` and `products` tables.

#### Task 2.3: Consider Materialized Views for Complex Reports

*   **Description:** For very complex or frequently accessed reports, use materialized views to pre-aggregate data.
*   **Sub-tasks:**
    *   **2.3.1 Identify suitable reports:** (e.g., `GetInventoryTurnoverReport`'s `getInventoryValueAt` could benefit).
    *   **2.3.2 Create materialized views:** Define SQL for views that pre-calculate aggregates.
    *   **2.3.3 Implement refresh strategy:** Schedule periodic refreshes of materialized views.

### 3. Advanced Graph Generation Support

While frontend handles rendering, the backend can provide data in a more structured, chart-friendly format.

#### Task 3.1: Standardize Chart Data Output

*   **Description:** Define a common JSON structure for chart data that frontend libraries can easily consume.
*   **Sub-tasks:**
    *   **3.1.1 Define `ChartData` struct:**
        ```go
        type ChartData struct {
            Labels []string        `json:"labels"`
            Datasets []ChartDataset `json:"datasets"`
            // Add metadata like title, type (bar, line, pie)
        }
        type ChartDataset struct {
            Label string        `json:"label"`
            Data  []interface{} `json:"data"` // Can be float64, int, etc.
            // Add styling options like backgroundColor, borderColor
        }
        ```
    *   **3.1.2 Modify `GetSalesTrendsReport` to return `ChartData`:** Transform `SalesTrend` data into this format.
    *   **3.1.3 Extend to other reports:** Apply this structure to other reports that could benefit from visualization.

### 4. Report Scheduling and Management

Allow users to schedule reports for periodic generation and delivery.

#### Task 4.1: Implement Report Scheduling Mechanism

*   **Description:** Integrate a scheduling library or use existing job queue for recurring tasks.
*   **Sub-tasks:**
    *   **4.1.1 Define `ScheduledReport` struct in `internal/domain/models.go`:**
        ```go
        type ScheduledReport struct {
            gorm.Model
            ReportDefinitionID uint `gorm:"not null"`
            ReportDefinition   ReportDefinition
            Schedule           string `gorm:"not null"` // e.g., "daily", "weekly", "0 0 * * *" (cron expression)
            RecipientUserID    uint   // User to send the report to
            Format             string `gorm:"not null"` // e.g., "csv", "pdf", "xlsx"
            LastRunAt          *time.Time
            NextRunAt          *time.Time
            Status             string `gorm:"default:'ACTIVE'"` // ACTIVE, PAUSED, COMPLETED
        }
        ```
    *   **4.1.2 Create API endpoints for `ScheduledReport`:** CRUD operations (`POST /scheduled-reports`, `GET /scheduled-reports`, `PATCH /scheduled-reports/{id}`, `DELETE /scheduled-reports/{id}`).
    *   **4.1.3 Implement a scheduler worker:** A background worker that reads `ScheduledReport` entries and queues report generation jobs at the appropriate times.

#### Task 4.2: Report Delivery Options

*   **Description:** Extend report delivery beyond just download links.
*   **Sub-tasks:**
    *   **4.2.1 Email delivery:** Attach generated reports to emails sent to recipients.
    *   **In-app notification:** Notify users when their scheduled report is ready for download.

### 5. Enhanced Error Handling and Observability for Reporting Jobs

Improve the robustness and debuggability of asynchronous report generation.

#### Task 5.1: Detailed Job Status and Error Logging

*   **Description:** Provide more granular status updates and detailed error messages for `domain.Job` entries.
*   **Sub-tasks:**
    *   **5.1.1 Update `domain.Job` `Result` and `LastError` fields:** Ensure they can store more comprehensive information (e.g., stack traces, specific validation errors).
    *   **5.1.2 Implement structured logging within `ReportingService.GenerateReport`:** Log each step of the report generation process, including parameters, intermediate results, and any errors.

#### Task 5.2: Job Progress Tracking

*   **Description:** Allow users to see the progress of long-running report generation jobs.
*   **Sub-tasks:**
    *   **5.2.1 Add `Progress` field to `domain.Job`:** (e.g., `float64` from 0.0 to 1.0).
    *   **5.2.2 Update `Progress` during report generation:** Periodically update the job's progress in the database.
    *   **5.2.3 Expose `Progress` via `GetReportJobStatus` API.**

### 6. Granular Authorization for Reports

Ensure users only access reports and data they are permitted to see.

#### Task 6.1: Implement Report-Specific Authorization

*   **Description:** Add logic to check user permissions before generating or serving report data.
*   **Sub-tasks:**
    *   **6.1.1 Integrate with existing RBAC:** Use user roles to determine access to report types.
    *   **6.1.2 Implement data-level authorization:** For reports filtered by `locationID` or `categoryID`, ensure the requesting user has access to that specific location or category. This might involve passing user context (e.g., allowed locations) to the `ReportingService` and `ReportsRepository`.

## Overall Considerations

*   **Performance Testing:** Conduct thorough performance testing for all new report types and the generic generation framework, especially with large datasets.
*   **Security Audit:** Ensure all new API endpoints and data access patterns adhere to security best practices, particularly regarding data exposure and injection vulnerabilities.
*   **Scalability of Workers:** Design the report generation workers to be horizontally scalable to handle increased load.
*   **User Interface (Frontend):** While backend-focused, consider how these enhancements will translate into a user-friendly frontend interface for report selection, filtering, scheduling, and visualization.
