# Frontend Guide: Implementing Bulk Product Import

This guide provides frontend developers with a comprehensive overview and step-by-step instructions for implementing the asynchronous bulk product import feature.

---

## 1. Overview of the Workflow

The bulk import process is **asynchronous** and designed to handle large files without blocking the UI. It is a two-phase process:

1.  **Phase 1: Upload & Validation:** The user uploads a file. The backend validates the file in the background. The frontend polls for the validation result.
2.  **Phase 2: Confirmation & Import:** The user reviews the validation results (e.g., number of valid products, list of errors) and confirms the import. The backend then starts the final import process in the background.

This ensures that no data is imported until the user has reviewed and confirmed the validated data.

---

## 2. API Endpoints

You will interact with the following four endpoints:

### A. Download CSV Template

-   **Endpoint:** `GET /api/v1/bulk/products/template`
-   **Description:** Downloads the official CSV template for the product import.
-   **Usage:** Provide a button for the user to download this template. This ensures they use the correct headers and format.
-   **Response:** A `text/csv` file attachment named `product_import_template.csv`.

### B. Upload File for Processing

-   **Endpoint:** `POST /api/v1/bulk/products/import`
-   **Description:** Uploads the user's completed CSV file to start the validation process.
-   **Request Body:** `multipart/form-data` with the file attached under the key `file`.
-   **Success Response (202 Accepted):** A job object containing the `id` for the import task.
    ```json
    {
        "ID": 123,
        "Type": "BULK_IMPORT",
        "Status": "QUEUED",
        "Payload": "{\"bucketName\":\"bulk-imports\",\"objectName\":\"...\",\"userId\":1}",
        "Result": "",
        "LastError": "",
        // ... other fields
    }
    ```
-   **Action:** Store the `ID` from the response. This is your `jobId` for the next steps.

### C. Get Job Status & Validation Results

-   **Endpoint:** `GET /api/v1/bulk/products/import/{jobId}/status`
-   **Description:** Poll this endpoint to get the current status of the import job.
-   **Usage:** After uploading the file, call this endpoint periodically (e.g., every 2-3 seconds) until the status is no longer `QUEUED` or `PROCESSING`.
-   **Job Statuses:**
    -   `QUEUED`: The job is waiting to be processed.
    -   `PROCESSING`: The backend is actively validating the CSV file.
    -   `PENDING_CONFIRMATION`: Validation is complete. The frontend should now display the results to the user.
    -   `COMPLETED`: The user has confirmed the import, and all valid products have been successfully saved to the database.
    -   `FAILED`: The job failed at some point. The `LastError` field will contain details.
-   **Response when `Status` is `PENDING_CONFIRMATION`:**
    The `Result` field will be a JSON string containing the validation outcome. You must parse this string to get the details.
    ```json
    // The 'Result' field of the job object, once parsed:
    {
        "totalRecords": 5,
        "validRecords": 4,
        "invalidRecords": 1,
        "errors": [
            "invalid number of columns"
        ],
        "validProducts": [
            { "ID": 0, "SKU": "SKU001", "Name": "Product 1", ... },
            { "ID": 0, "SKU": "SKU002", "Name": "Product 2", ... },
            { "ID": 0, "SKU": "SKU003", "Name": "Product 3", ... },
            { "ID": 0, "SKU": "SKU004", "Name": "Product 4", ... }
        ]
    }
    ```

### D. Confirm and Execute Import

-   **Endpoint:** `POST /api/v1/bulk/products/import/{jobId}/confirm`
-   **Description:** Confirms that the user wants to proceed with importing the `validProducts` identified in the validation phase.
-   **Usage:** Only call this after the job status is `PENDING_CONFIRMATION` and the user has given their approval.
-   **Success Response (202 Accepted):**
    ```json
    {
        "jobId": 123,
        "status": "PROCESSING",
        "message": "Bulk import confirmation received"
    }
    ```
-   **Action:** After this call, you can consider the import process finalized from the user's perspective. You can optionally continue polling the status endpoint until the status changes to `COMPLETED`.

---

## 3. Step-by-Step Frontend Workflow

Here is the recommended implementation flow:

1.  **Provide Template:** On your "Bulk Import" page, have a "Download Template" button that hits the `GET .../template` endpoint.

2.  **Handle File Upload:**
    -   The user selects a file and clicks "Upload".
    -   `POST` the file to `.../import`.
    -   On success, get the `jobId` from the response.
    -   Show a loading indicator to the user (e.g., "Validating your file, please wait...").

3.  **Poll for Validation Status:**
    -   Start a poller that calls `GET .../{jobId}/status` every 2-3 seconds.
    -   Continue polling as long as the `Status` is `QUEUED` or `PROCESSING`.

4.  **Display Validation Results:**
    -   When the poller receives a `Status` of `PENDING_CONFIRMATION`, stop polling.
    -   Hide the loading indicator.
    -   Parse the `Result` JSON string from the job object.
    -   Display a summary to the user:
        -   "Validation complete: **4 out of 5** products are ready for import."
        -   If `invalidRecords > 0`, display the list of errors from the `errors` array. This helps the user fix their CSV file for a future attempt.
    -   If `validRecords > 0`, show a "Confirm Import" button. You can also display the `validProducts` in a table for final review.

5.  **Confirm the Import:**
    -   When the user clicks "Confirm Import", `POST` to `.../{jobId}/confirm`.
    -   On success, you can show a final success message like "Import started! Your products will be available shortly." and navigate the user away or update the UI to reflect the completion.

6.  **Handle Failures:**
    -   If at any point an API call fails or the job `Status` becomes `FAILED`, show an appropriate error message to the user. The `LastError` field on the job object may contain useful information.

---

## 4. CSV Template Details

The template `product_import_template.csv` contains the following headers:

`SKU,Name,Description,CategoryName,SubCategoryName,SupplierName,Brand,PurchasePrice,SellingPrice,LocationName,Status`

**Important Notes for the User:**
-   `SKU` and `Name` are mandatory.
-   **Automatic Entity Creation**: For the `CategoryName`, `SubCategoryName`, and `SupplierName` columns, you can provide the name of the entity.
    -   If an entity with that name already exists, the product will be linked to it.
    -   If an entity with that name **does not exist**, a new one will be **automatically created** during the import confirmation step.
-   **Sub-Category Creation**: To create a new sub-category, you must provide both the `SubCategoryName` and the corresponding `CategoryName` in the same row.
-   `PurchasePrice` and `SellingPrice` should be numbers.
-   The `Status` column is optional and will default to "Active".
-   **Stock is not handled in this import.** Initial stock levels must be adjusted manually after the products are created.
