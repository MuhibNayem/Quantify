# API Documentation

This document provides a detailed description of the API endpoints, including their requests and responses.

## Alerts

### PUT /products/{productId}/alert-settings

-   **Summary:** Configure alert thresholds for a product
-   **Description:** Configures low-stock, overstock, and expiry alert thresholds for a specific product.
-   **Request:**
    -   **URL Params:**
        -   `productId` (integer, required): The ID of the product.
    -   **Body:**
        ```json
        {
            "lowStockLevel": 10,
            "overStockLevel": 100,
            "expiryAlertDays": 30
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "ProductID": 1,
            "LowStockLevel": 10,
            "OverStockLevel": 100,
            "ExpiryAlertDays": 30,
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:00:00Z"
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **404 Not Found:** If the product is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /alerts

-   **Summary:** Get a list of all active alerts
-   **Description:** Retrieves a list of all active stock-related alerts.
-   **Request:**
    -   **Query Params:**
        -   `type` (string, optional): Filter by alert type (LOW_STOCK, OUT_OF_STOCK, OVERSTOCK, EXPIRY_ALERT).
        -   `status` (string, optional): Filter by alert status (ACTIVE, RESOLVED). Defaults to ACTIVE.
        -   `productId` (integer, optional): Filter by Product ID.
-   **Response:**
    -   **200 OK:**
        ```json
        [
            {
                "ID": 1,
                "ProductID": 1,
                "Product": { ... },
                "BatchID": null,
                "Batch": null,
                "Type": "LOW_STOCK",
                "Message": "Product 'Sample Product' is running low. Current quantity: 5",
                "TriggeredAt": "2025-11-08T21:00:00Z",
                "Status": "ACTIVE",
                "ResolvedAt": null,
                "ResolvedBy": null
            }
        ]
        ```
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /alerts/{alertId}

-   **Summary:** Get an alert by ID
-   **Description:** Retrieves details of a specific alert by its ID.
-   **Request:**
    -   **URL Params:**
        -   `alertId` (integer, required): The ID of the alert.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "ProductID": 1,
            "Product": { ... },
            "BatchID": null,
            "Batch": null,
            "Type": "LOW_STOCK",
            "Message": "Product 'Sample Product' is running low. Current quantity: 5",
            "TriggeredAt": "2025-11-08T21:00:00Z",
            "Status": "ACTIVE",
            "ResolvedAt": null,
            "ResolvedBy": null
        }
        ```
    -   **404 Not Found:** If the alert is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### PATCH /alerts/{alertId}/resolve

-   **Summary:** Resolve an alert
-   **Description:** Marks a specific alert as resolved.
-   **Request:**
    -   **URL Params:**
        -   `alertId` (integer, required): The ID of the alert.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "ProductID": 1,
            "Product": { ... },
            "BatchID": null,
            "Batch": null,
            "Type": "LOW_STOCK",
            "Message": "Product 'Sample Product' is running low. Current quantity: 5",
            "TriggeredAt": "2025-11-08T21:00:00Z",
            "Status": "RESOLVED",
            "ResolvedAt": "2025-11-08T21:05:00Z",
            "ResolvedBy": 1
        }
        ```
    -   **404 Not Found:** If the alert is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### PUT /users/{userId}/notification-settings

-   **Summary:** Configure user notification preferences
-   **Description:** Configures email and SMS notification preferences for a user.
-   **Request:**
    -   **URL Params:**
        -   `userId` (integer, required): The ID of the user.
    -   **Body:**
        ```json
        {
            "emailNotificationsEnabled": true,
            "smsNotificationsEnabled": false,
            "emailAddress": "user@example.com",
            "phoneNumber": "+1234567890"
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "UserID": 1,
            "EmailNotificationsEnabled": true,
            "SMSNotificationsEnabled": false,
            "EmailAddress": "user@example.com",
            "PhoneNumber": "+1234567890",
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:00:00Z"
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **404 Not Found:** If the user is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

## Barcode

### GET /barcode/generate

-   **Summary:** Generate a barcode image for a product
-   **Description:** Generates a barcode image (PNG) for a given product SKU or ID.
-   **Request:**
    -   **Query Params:**
        -   `sku` (string, optional): Product SKU.
        -   `productId` (integer, optional): Product ID.
-   **Response:**
    -   **200 OK:** Returns a PNG image of the barcode.
    -   **400 Bad Request:** If neither `sku` nor `productId` is provided.
    -   **404 Not Found:** If the product is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /barcode/lookup

-   **Summary:** Lookup a product by barcode/UPC
-   **Description:** Retrieves product details by scanning its barcode or UPC.
-   **Request:**
    -   **Query Params:**
        -   `barcode` (string, required): Barcode or UPC value.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "SKU": "SKU123",
            "Name": "Sample Product",
            "Description": "This is a sample product.",
            "CategoryID": 1,
            "SubCategoryID": 1,
            "SupplierID": 1,
            "Brand": "Sample Brand",
            "PurchasePrice": 10.50,
            "SellingPrice": 20.00,
            "BarcodeUPC": "123456789012",
            "ImageURLs": [],
            "Status": "Active",
            "LocationID": 1,
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:00:00Z",
            "Category": { ... },
            "SubCategory": { ... },
            "Supplier": { ... },
            "Location": { ... }
        }
        ```
    -   **400 Bad Request:** If the `barcode` query param is missing.
    -   **404 Not Found:** If the product is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

## Bulk Operations

### GET /bulk/products/template

-   **Summary:** Download product import template
-   **Description:** Downloads a CSV/Excel template file with required headers for product creation.
-   **Request:** None
-   **Response:**
    -   **200 OK:** Returns a CSV file with the following headers: `SKU,Name,Description,CategoryID,SubCategoryID,SupplierID,Brand,PurchasePrice,SellingPrice,BarcodeUPC,ImageURLs,Status`

### POST /bulk/products/import

-   **Summary:** Upload a file for bulk product import
-   **Description:** Uploads a CSV/Excel file for bulk product creation/update. Returns a job ID for status tracking.
-   **Request:**
    -   **Form Data:**
        -   `file` (file, required): The CSV/Excel file to upload.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "jobId": "...",
            "status": "QUEUED",
            "message": "Bulk import job queued for processing.",
            "filePath": "...",
            "totalRecords": 0,
            "validRecords": 0,
            "invalidRecords": 0,
            "errors": [],
            "preview": []
        }
        ```
    -   **400 Bad Request:** If the file is not provided.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /bulk/products/import/{jobId}/status

-   **Summary:** Get bulk import job status
-   **Description:** Retrieves the status and validation results of a bulk import job.
-   **Request:**
    -   **URL Params:**
        -   `jobId` (string, required): The ID of the bulk import job.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "jobId": "...",
            "status": "PENDING_CONFIRMATION",
            "message": "Bulk import job validation complete.",
            "filePath": "...",
            "totalRecords": 100,
            "validRecords": 98,
            "invalidRecords": 2,
            "errors": [ ... ],
            "preview": [ ... ]
        }
        ```
    -   **404 Not Found:** If the job is not found.

### POST /bulk/products/import/{jobId}/confirm

-   **Summary:** Confirm and execute bulk import
-   **Description:** Confirms and executes the bulk import after preview.
-   **Request:**
    -   **URL Params:**
        -   `jobId` (string, required): The ID of the bulk import job.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "jobId": "...",
            "status": "PROCESSING",
            "message": "Bulk import initiated"
        }
        ```
    -   **400 Bad Request:** If the job is not in the `PENDING_CONFIRMATION` state.
    -   **404 Not Found:** If the job is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /bulk/products/export

-   **Summary:** Export product catalog
-   **Description:** Exports the entire product catalog or a filtered list of products to a CSV/Excel file.
-   **Request:**
    -   **Query Params:**
        -   `format` (string, optional): Export format (csv, excel). Defaults to `csv`.
        -   `category` (integer, optional): Filter by Category ID.
        -   `supplier` (integer, optional): Filter by Supplier ID.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "jobId": "...",
            "status": "QUEUED",
            "message": "Bulk export job queued for processing."
        }
        ```
    -   **500 Internal Server Error:** If there is a server-side error.

## Categories

### POST /categories

-   **Summary:** Create a new category
-   **Description:** Create a new product category.
-   **Request:**
    -   **Body:**
        ```json
        {
            "name": "Electronics"
        }
        ```
-   **Response:**
    -   **201 Created:**
        ```json
        {
            "ID": 1,
            "Name": "Electronics",
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:00:00Z"
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **409 Conflict:** If a category with the same name already exists.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /categories

-   **Summary:** Get a list of categories
-   **Description:** Get a list of all product categories.
-   **Request:** None
-   **Response:**
    -   **200 OK:**
        ```json
        [
            {
                "ID": 1,
                "Name": "Electronics",
                "SubCategories": [ ... ],
                "CreatedAt": "2025-11-08T21:00:00Z",
                "UpdatedAt": "2025-11-08T21:00:00Z"
            }
        ]
        ```
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /categories/{id}

-   **Summary:** Get a category by ID
-   **Description:** Get a single category by its ID.
-   **Request:**
    -   **URL Params:**
        -   `id` (integer, required): The ID of the category.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "Name": "Electronics",
            "SubCategories": [ ... ],
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:00:00Z"
        }
        ```
    -   **404 Not Found:** If the category is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /categories/name/{name}

-   **Summary:** Get a category by Name
-   **Description:** Get a single category by its Name.
-   **Request:**
    -   **URL Params:**
        -   `name` (string, required): The Name of the category.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "Name": "Electronics",
            "SubCategories": [ ... ],
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:00:00Z"
        }
        ```
    -   **404 Not Found:** If the category is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### PUT /categories/{id}

-   **Summary:** Update an existing category
-   **Description:** Update an existing product category.
-   **Request:**
    -   **URL Params:**
        -   `id` (integer, required): The ID of the category.
    -   **Body:**
        ```json
        {
            "name": "Consumer Electronics"
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "Name": "Consumer Electronics",
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:05:00Z"
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **404 Not Found:** If the category is not found.
    -   **409 Conflict:** If a category with the same name already exists.
    -   **500 Internal Server Error:** If there is a server-side error.

### DELETE /categories/{id}

-   **Summary:** Delete a category
-   **Description:** Delete a product category by its ID. Cannot delete if products or subcategories are associated.
-   **Request:**
    -   **URL Params:**
        -   `id` (integer, required): The ID of the category.
-   **Response:**
    -   **204 No Content:** If the category is deleted successfully.
    -   **404 Not Found:** If the category is not found.
    -   **409 Conflict:** If the category has associated products or subcategories.
    -   **500 Internal Server Error:** If there is a server-side error.

### POST /categories/{categoryId}/subcategories

-   **Summary:** Create a new sub-category
-   **Description:** Create a new sub-category for a specific category.
-   **Request:**
    -   **URL Params:**
        -   `categoryId` (integer, required): The ID of the parent category.
    -   **Body:**
        ```json
        {
            "name": "Smartphones"
        }
        ```
-   **Response:**
    -   **201 Created:**
        ```json
        {
            "ID": 1,
            "Name": "Smartphones",
            "CategoryID": 1,
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:00:00Z"
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **404 Not Found:** If the parent category is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /categories/{categoryId}/subcategories

-   **Summary:** Get sub-categories for a category
-   **Description:** Get a list of sub-categories for a specific category.
-   **Request:**
    -   **URL Params:**
        -   `categoryId` (integer, required): The ID of the parent category.
-   **Response:**
    -   **200 OK:**
        ```json
        [
            {
                "ID": 1,
                "Name": "Smartphones",
                "CategoryID": 1,
                "CreatedAt": "2025-11-08T21:00:00Z",
                "UpdatedAt": "2025-11-08T21:00:00Z"
            }
        ]
        ```
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /subcategories/{id}

-   **Summary:** Get a sub-category by ID
-   **Description:** Get a single sub-category by its ID.
-   **Request:**
    -   **URL Params:**
        -   `id` (integer, required): The ID of the sub-category.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "Name": "Smartphones",
            "CategoryID": 1,
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:00:00Z"
        }
        ```
    -   **404 Not Found:** If the sub-category is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### PUT /subcategories/{id}

-   **Summary:** Update an existing sub-category
-   **Description:** Update an existing sub-category.
-   **Request:**
    -   **URL Params:**
        -   `id` (integer, required): The ID of the sub-category.
    -   **Body:**
        ```json
        {
            "name": "Mobile Phones"
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "Name": "Mobile Phones",
            "CategoryID": 1,
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:05:00Z"
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **404 Not Found:** If the sub-category is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### DELETE /subcategories/{id}

-   **Summary:** Delete a sub-category
-   **Description:** Delete a sub-category by its ID. Cannot delete if products are associated.
-   **Request:**
    -   **URL Params:**
        -   `id` (integer, required): The ID of the sub-category.
-   **Response:**
    -   **204 No Content:** If the sub-category is deleted successfully.
    -   **404 Not Found:** If the sub-category is not found.
    -   **409 Conflict:** If the sub-category has associated products.
    -   **500 Internal Server Error:** If there is a server-side error.

## CRM

### POST /crm/customers

-   **Summary:** Create a new customer
-   **Description:** Create a new customer with the provided details.
-   **Request:**
    -   **Body:**
        ```json
        {
            "username": "newcustomer",
            "password": "password123",
            "firstName": "John",
            "lastName": "Doe",
            "email": "john.doe@example.com",
            "phoneNumber": "1234567890"
        }
        ```
-   **Response:**
    -   **201 Created:**
        ```json
        {
            "ID": 1,
            "Username": "newcustomer",
            "Role": "Customer",
            "IsActive": true,
            "FirstName": "John",
            "LastName": "Doe",
            "Email": "john.doe@example.com",
            "PhoneNumber": "1234567890",
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:00:00Z"
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /crm/customers/{identifier}

-   **Summary:** Get a customer by ID, username, email, or phone
-   **Description:** Get a single customer by their ID, username, email, or phone number.
-   **Request:**
    -   **URL Params:**
        -   `identifier` (string, required): The ID, username, email, or phone number of the customer.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "Username": "newcustomer",
            ...
        }
        ```
    -   **404 Not Found:** If the customer is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### PUT /crm/customers/{userId}

-   **Summary:** Update an existing customer
-   **Description:** Update an existing customer with the provided details.
-   **Request:**
    -   **URL Params:**
        -   `userId` (integer, required): The ID of the customer.
    -   **Body:**
        ```json
        {
            "firstName": "Jane",
            "lastName": "Doe"
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "FirstName": "Jane",
            "LastName": "Doe",
            ...
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **404 Not Found:** If the customer is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### DELETE /crm/customers/{userId}

-   **Summary:** Delete a customer
-   **Description:** Delete a customer by their ID.
-   **Request:**
    -   **URL Params:**
        -   `userId` (integer, required): The ID of the customer.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "message": "Customer deleted successfully"
        }
        ```
    -   **404 Not Found:** If the customer is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /crm/loyalty/{userId}

-   **Summary:** Get a customer's loyalty account
-   **Description:** Get a customer's loyalty account by their user ID.
-   **Request:**
    -   **URL Params:**
        -   `userId` (integer, required): The ID of the user.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "UserID": 1,
            "Points": 100,
            "Tier": "Silver",
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:00:00Z"
        }
        ```
    -   **404 Not Found:** If the loyalty account is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### POST /crm/loyalty/{userId}/points

-   **Summary:** Add loyalty points to a customer's account
-   **Description:** Add loyalty points to a customer's account.
-   **Request:**
    -   **URL Params:**
        -   `userId` (integer, required): The ID of the user.
    -   **Body:**
        ```json
        {
            "points": 50
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "UserID": 1,
            "Points": 150,
            "Tier": "Silver",
            ...
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **404 Not Found:** If the loyalty account is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### POST /crm/loyalty/{userId}/redeem

-   **Summary:** Redeem loyalty points from a customer's account
-   **Description:** Redeem loyalty points from a customer's account.
-   **Request:**
    -   **URL Params:**
        -   `userId` (integer, required): The ID of the user.
    -   **Body:**
        ```json
        {
            "points": 50
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "UserID": 1,
            "Points": 100,
            "Tier": "Silver",
            ...
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid or if the user has insufficient points.
    -   **404 Not Found:** If the loyalty account is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

## Health

### GET /health

-   **Summary:** Show the status of the server
-   **Description:** Get the status of the server.
-   **Request:** None
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "status": "UP"
        }
        ```

## Inventory

### POST /inventory/transfers

-   **Summary:** Create a stock transfer
-   **Description:** Create a new stock transfer between two locations.
-   **Request:**
    -   **Body:**
        ```json
        {
            "productID": 1,
            "sourceLocationID": 1,
            "destLocationID": 2,
            "quantity": 10
        }
        ```
-   **Response:**
    -   **201 Created:**
        ```json
        {
            "ID": 1,
            "ProductID": 1,
            "SourceLocationID": 1,
            "DestLocationID": 2,
            "Quantity": 10,
            "TransferredBy": 1,
            "TransferredAt": "2025-11-08T21:00:00Z"
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **500 Internal Server Error:** If there is a server-side error.

## Locations

### POST /locations

-   **Summary:** Create a new location
-   **Description:** Create a new inventory location.
-   **Request:**
    -   **Body:**
        ```json
        {
            "name": "Warehouse A",
            "address": "123 Main St"
        }
        ```
-   **Response:**
    -   **201 Created:**
        ```json
        {
            "ID": 1,
            "Name": "Warehouse A",
            "Address": "123 Main St",
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:00:00Z"
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **409 Conflict:** If a location with the same name already exists.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /locations

-   **Summary:** Get a list of locations
-   **Description:** Get a list of all inventory locations.
-   **Request:** None
-   **Response:**
    -   **200 OK:**
        ```json
        [
            {
                "ID": 1,
                "Name": "Warehouse A",
                "Address": "123 Main St",
                "CreatedAt": "2025-11-08T21:00:00Z",
                "UpdatedAt": "2025-11-08T21:00:00Z"
            }
        ]
        ```
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /locations/{id}

-   **Summary:** Get a location by ID
-   **Description:** Get a single location by its ID.
-   **Request:**
    -   **URL Params:**
        -   `id` (integer, required): The ID of the location.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "Name": "Warehouse A",
            "Address": "123 Main St",
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:00:00Z"
        }
        ```
    -   **404 Not Found:** If the location is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### PUT /locations/{id}

-   **Summary:** Update an existing location
-   **Description:** Update an existing inventory location.
-   **Request:**
    -   **URL Params:**
        -   `id` (integer, required): The ID of the location.
    -   **Body:**
        ```json
        {
            "name": "Main Warehouse",
            "address": "456 Oak Ave"
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "Name": "Main Warehouse",
            "Address": "456 Oak Ave",
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:05:00Z"
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **404 Not Found:** If the location is not found.
    -   **409 Conflict:** If a location with the same name already exists.
    -   **500 Internal Server Error:** If there is a server-side error.

### DELETE /locations/{id}

-   **Summary:** Delete a location
-   **Description:** Delete an inventory location by its ID. Cannot delete if products, batches, or stock adjustments are associated.
-   **Request:**
    -   **URL Params:**
        -   `id` (integer, required): The ID of the location.
-   **Response:**
    -   **204 No Content:** If the location is deleted successfully.
    -   **404 Not Found:** If the location is not found.
    -   **409 Conflict:** If the location has associated data.
    -   **500 Internal Server Error:** If there is a server-side error.

## Products

### POST /products

-   **Summary:** Create a new product
-   **Description:** Create a new product with the provided details.
-   **Request:**
    -   **Body:**
        ```json
        {
            "sku": "SKU123",
            "name": "Sample Product",
            "description": "This is a sample product.",
            "categoryID": 1,
            "subCategoryID": 1,
            "supplierID": 1,
            "brand": "Sample Brand",
            "purchasePrice": 10.50,
            "sellingPrice": 20.00,
            "barcodeUPC": "123456789012",
            "imageURLs": [],
            "locationID": 1
        }
        ```
-   **Response:**
    -   **201 Created:**
        ```json
        {
            "ID": 1,
            "SKU": "SKU123",
            "Name": "Sample Product",
            "Description": "This is a sample product.",
            "CategoryID": 1,
            "SubCategoryID": 1,
            "SupplierID": 1,
            "Brand": "Sample Brand",
            "PurchasePrice": 10.50,
            "SellingPrice": 20.00,
            "BarcodeUPC": "123456789012",
            "ImageURLs": [],
            "Status": "Active",
            "LocationID": 1,
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:00:00Z"
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **409 Conflict:** If a product with the same SKU or BarcodeUPC already exists.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /products

-   **Summary:** Get a list of products
-   **Description:** Get a paginated, searchable, and filterable list of products.
-   **Request:**
    -   **Query Params:**
        -   `page` (integer, optional): Page number. Defaults to 1.
        -   `limit` (integer, optional): Number of items per page. Defaults to 10.
        -   `search` (string, optional): Search term for Product Name, SKU, or Barcode.
        -   `category` (integer, optional): Filter by Category ID.
        -   `supplier` (integer, optional): Filter by Supplier ID.
        -   `status` (string, optional): Filter by Status (Active, Archived, Discontinued).
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "products": [ ... ],
            "totalItems": 100,
            "currentPage": 1,
            "totalPages": 10,
            "itemsPerPage": 10
        }
        ```
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /products/{id}

-   **Summary:** Get a product by ID
-   **Description:** Get a single product by its ID.
-   **Request:**
    -   **URL Params:**
        -   `id` (integer, required): The ID of the product.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "SKU": "SKU123",
            "Name": "Sample Product",
            ...
        }
        ```
    -   **404 Not Found:** If the product is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /products/sku/{sku}

-   **Summary:** Get a product by SKU
-   **Description:** Get a single product by its SKU.
-   **Request:**
    -   **URL Params:**
        -   `sku` (string, required): The SKU of the product.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "SKU": "SKU123",
            "Name": "Sample Product",
            ...
        }
        ```
    -   **404 Not Found:** If the product is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /products/barcode/{barcode}

-   **Summary:** Get a product by Barcode
-   **Description:** Get a single product by its Barcode.
-   **Request:**
    -   **URL Params:**
        -   `barcode` (string, required): The Barcode of the product.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "SKU": "SKU123",
            "Name": "Sample Product",
            ...
        }
        ```
    -   **404 Not Found:** If the product is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### PUT /products/{id}

-   **Summary:** Update an existing product
-   **Description:** Update an existing product with the provided details.
-   **Request:**
    -   **URL Params:**
        -   `id` (integer, required): The ID of the product.
    -   **Body:**
        ```json
        {
            "name": "Updated Product Name",
            ...
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "Name": "Updated Product Name",
            ...
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **404 Not Found:** If the product is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### DELETE /products/{id}

-   **Summary:** Delete a product
-   **Description:** Delete a product by its ID. Restricted if product has associated sales or stock history.
-   **Request:**
    -   **URL Params:**
        -   `id` (integer, required): The ID of the product.
-   **Response:**
    -   **204 No Content:** If the product is deleted successfully.
    -   **404 Not Found:** If the product is not found.
    -   **409 Conflict:** If the product has associated data.
    -   **500 Internal Server Error:** If there is a server-side error.

### PATCH /products/{id}/archive

-   **Summary:** Archive a product
-   **Description:** Archive a product by setting its status to 'Archived'.
-   **Request:**
    -   **URL Params:**
        -   `id` (integer, required): The ID of the product.
    -   **Body:**
        ```json
        {
            "status": "Archived"
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "Status": "Archived",
            ...
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **404 Not Found:** If the product is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

## Replenishment

### POST /replenishment/forecast/generate

-   **Summary:** Trigger demand forecast generation
-   **Description:** Triggers a demand forecasting process for a product or all products.
-   **Request:**
    -   **Body:**
        ```json
        {
            "productID": 1,
            "periodInDays": 30
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "message": "Demand forecast generation initiated."
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /replenishment/forecast/{forecastId}

-   **Summary:** Get a specific demand forecast
-   **Description:** Retrieves details of a specific demand forecast by its ID.
-   **Request:**
    -   **URL Params:**
        -   `forecastId` (integer, required): The ID of the forecast.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "ProductID": 1,
            "ForecastPeriod": "30_DAYS",
            "PredictedDemand": 100,
            "GeneratedAt": "2025-11-08T21:00:00Z"
        }
        ```
    -   **404 Not Found:** If the forecast is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /replenishment/suggestions

-   **Summary:** Get a list of reorder suggestions
-   **Description:** Retrieves a list of suggested reorders based on forecast and stock levels.
-   **Request:**
    -   **Query Params:**
        -   `status` (string, optional): Filter by suggestion status (PENDING, APPROVED).
        -   `supplierId` (integer, optional): Filter by Supplier ID.
-   **Response:**
    -   **200 OK:**
        ```json
        [
            {
                "ID": 1,
                "ProductID": 1,
                "SupplierID": 1,
                "CurrentStock": 10,
                "PredictedDemand": 100,
                "SuggestedOrderQuantity": 90,
                "LeadTimeDays": 7,
                "Status": "PENDING",
                "SuggestedAt": "2025-11-08T21:00:00Z"
            }
        ]
        ```
    -   **500 Internal Server Error:** If there is a server-side error.

### POST /replenishment/suggestions/{suggestionId}/create-po

-   **Summary:** Create a draft Purchase Order from a reorder suggestion
-   **Description:** Creates a draft Purchase Order based on a selected reorder suggestion.
-   **Request:**
    -   **URL Params:**
        -   `suggestionId` (integer, required): The ID of the reorder suggestion.
-   **Response:**
    -   **201 Created:**
        ```json
        {
            "ID": 1,
            "SupplierID": 1,
            "Status": "DRAFT",
            "OrderDate": "2025-11-08T21:00:00Z",
            "CreatedBy": 1,
            "PurchaseOrderItems": [ ... ]
        }
        ```
    -   **400 Bad Request:** If the suggestion is not in the `PENDING` state.
    -   **404 Not Found:** If the suggestion is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### POST /purchase-orders/{poId}/send

-   **Summary:** Send a Purchase Order to the supplier
-   **Description:** Marks an approved Purchase Order as SENT.
-   **Request:**
    -   **URL Params:**
        -   `poId` (integer, required): The ID of the Purchase Order.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "Status": "SENT",
            ...
        }
        ```
    -   **400 Bad Request:** If the Purchase Order is not in the `APPROVED` state.
    -   **404 Not Found:** If the Purchase Order is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### POST /purchase-orders/{poId}/approve

-   **Summary:** Approve a draft Purchase Order
-   **Description:** Approves a draft Purchase Order, changing its status to APPROVED.
-   **Request:**
    -   **URL Params:**
        -   `poId` (integer, required): The ID of the Purchase Order.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "Status": "APPROVED",
            ...
        }
        ```
    -   **400 Bad Request:** If the Purchase Order is not in the `DRAFT` state.
    -   **404 Not Found:** If the Purchase Order is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /purchase-orders/{poId}

-   **Summary:** Get a Purchase Order by ID
-   **Description:** Retrieves details of a specific Purchase Order by its ID.
-   **Request:**
    -   **URL Params:**
        -   `poId` (integer, required): The ID of the Purchase Order.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            ...
        }
        ```
    -   **404 Not Found:** If the Purchase Order is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### PUT /purchase-orders/{poId}

-   **Summary:** Update a Purchase Order
-   **Description:** Updates details of a specific Purchase Order.
-   **Request:**
    -   **URL Params:**
        -   `poId` (integer, required): The ID of the Purchase Order.
    -   **Body:**
        ```json
        {
            "supplierID": 2,
            ...
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "SupplierID": 2,
            ...
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **404 Not Found:** If the Purchase Order is not found.
    -   **409 Conflict:** If the Purchase Order is not in the `DRAFT` state.
    -   **500 Internal Server Error:** If there is a server-side error.

### POST /purchase-orders/{poId}/receive

-   **Summary:** Record received goods for a Purchase Order
-   **Description:** Records received quantities for items in a Purchase Order and updates stock levels.
-   **Request:**
    -   **URL Params:**
        -   `poId` (integer, required): The ID of the Purchase Order.
    -   **Body:**
        ```json
        {
            "items": [
                {
                    "purchaseOrderItemID": 1,
                    "receivedQuantity": 10,
                    "batchNumber": "BATCH123",
                    "expiryDate": "2026-11-08T00:00:00Z"
                }
            ]
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "Status": "PARTIALLY_RECEIVED",
            ...
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **404 Not Found:** If the Purchase Order or an item is not found.
    -   **409 Conflict:** If the Purchase Order is not in the `SENT` state or if the received quantity exceeds the ordered quantity.
    -   **500 Internal Server Error:** If there is a server-side error.

### POST /purchase-orders/{poId}/cancel

-   **Summary:** Cancel a Purchase Order
-   **Description:** Cancels a Purchase Order if it's in DRAFT or APPROVED status.
-   **Request:**
    -   **URL Params:**
        -   `poId` (integer, required): The ID of the Purchase Order.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "Status": "CANCELLED",
            ...
        }
        ```
    -   **400 Bad Request:** If the Purchase Order is not in the `DRAFT` or `APPROVED` state.
    -   **404 Not Found:** If the Purchase Order is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

## Reports

### POST /reports/sales-trends

-   **Summary:** Get sales trends report
-   **Description:** Generates a report on sales trends over a specified period, with optional filters.
-   **Request:**
    -   **Body:**
        ```json
        {
            "startDate": "2025-10-01T00:00:00Z",
            "endDate": "2025-10-31T23:59:59Z",
            "categoryID": 1,
            "locationID": 1,
            "groupBy": "day"
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "period": "2025-10-01 to 2025-10-31",
            "totalSales": 12345.67,
            "averageDailySales": 411.52,
            "salesTrends": [ ... ],
            "topSellingProducts": [ ... ]
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **500 Internal Server Error:** If there is a server-side error.

### POST /reports/sales-trends/export

-   **Summary:** Export sales trends report
-   **Description:** Exports a sales trends report as a CSV, PDF, or Excel file.
-   **Request:**
    -   **Query Params:**
        -   `format` (string, optional): Export format (csv, pdf, excel). Defaults to `csv`.
    -   **Body:**
        ```json
        {
            "startDate": "2025-10-01T00:00:00Z",
            "endDate": "2025-10-31T23:59:59Z",
            "categoryID": 1,
            "locationID": 1,
            "groupBy": "day"
        }
        ```
-   **Response:**
    -   **202 Accepted:**
        ```json
        {
            "jobId": "..."
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /reports/jobs/{jobId}

-   **Summary:** Get report job status
-   **Description:** Get the status of a report generation job.
-   **Request:**
    -   **URL Params:**
        -   `jobId` (string, required): The ID of the report generation job.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "jobId": "...",
            "reportType": "sales_trends_csv",
            "params": { ... },
            "status": "COMPLETED",
            "fileUrl": "/api/v1/reports/download/..."
        }
        ```
    -   **404 Not Found:** If the job is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /reports/download/{jobId}

-   **Summary:** Download report file
-   **Description:** Download a generated report file.
-   **Request:**
    -   **URL Params:**
        -   `jobId` (string, required): The ID of the report generation job.
-   **Response:**
    -   **302 Found:** Redirects to the file URL in MinIO.
    -   **404 Not Found:** If the job is not found or the report is not ready.
    -   **500 Internal Server Error:** If there is a server-side error.

### POST /reports/inventory-turnover

-   **Summary:** Get inventory turnover report
-   **Description:** Generates a report on inventory turnover rate over a specified period.
-   **Request:**
    -   **Body:**
        ```json
        {
            "startDate": "2025-10-01T00:00:00Z",
            "endDate": "2025-10-31T23:59:59Z",
            "categoryID": 1,
            "locationID": 1
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "period": "2025-10-01 to 2025-10-31",
            "totalCostOfGoodsSold": 5432.10,
            "averageInventoryValue": 1234.56,
            "inventoryTurnoverRate": 4.40
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **500 Internal Server Error:** If there is a server-side error.

### POST /reports/profit-margin

-   **Summary:** Get profit margin report
-   **Description:** Generates a report on profit margins for products or categories over a specified period.
-   **Request:**
    -   **Body:**
        ```json
        {
            "startDate": "2025-10-01T00:00:00Z",
            "endDate": "2025-10-31T23:59:59Z",
            "categoryID": 1,
            "locationID": 1
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "period": "2025-10-01 to 2025-10-31",
            "totalRevenue": 12345.67,
            "totalCost": 5432.10,
            "grossProfit": 6913.57,
            "grossProfitMargin": 0.56
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **500 Internal Server Error:** If there is a server-side error.

## Stock

### POST /products/{productId}/stock/batches

-   **Summary:** Add new stock with batch information
-   **Description:** Adds a new batch of stock for a specific product.
-   **Request:**
    -   **URL Params:**
        -   `productId` (integer, required): The ID of the product.
    -   **Body:**
        ```json
        {
            "batchNumber": "BATCH123",
            "quantity": 100,
            "expiryDate": "2026-11-08T00:00:00Z"
        }
        ```
-   **Response:**
    -   **201 Created:**
        ```json
        {
            "ID": 1,
            "ProductID": 1,
            "BatchNumber": "BATCH123",
            "Quantity": 100,
            "ExpiryDate": "2026-11-08T00:00:00Z",
            "LocationID": 1,
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:00:00Z"
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **404 Not Found:** If the product is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /products/{productId}/stock

-   **Summary:** Get current stock levels for a product
-   **Description:** Retrieves current stock levels and batch breakdown for a specific product.
-   **Request:**
    -   **URL Params:**
        -   `productId` (integer, required): The ID of the product.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "productId": 1,
            "currentQuantity": 150,
            "batches": [ ... ]
        }
        ```
    -   **404 Not Found:** If the product is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### POST /products/{productId}/stock/adjustments

-   **Summary:** Perform a manual stock adjustment
-   **Description:** Performs a manual stock adjustment (stock-in or stock-out) for a product.
-   **Request:**
    -   **URL Params:**
        -   `productId` (integer, required): The ID of the product.
    -   **Body:**
        ```json
        {
            "type": "STOCK_OUT",
            "quantity": 5,
            "reasonCode": "DAMAGED",
            "notes": "Product was damaged during handling."
        }
        ```
-   **Response:**
    -   **201 Created:**
        ```json
        {
            "message": "Stock adjustment successful"
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **404 Not Found:** If the product is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /products/{productId}/stock/history

-   **Summary:** Get stock adjustment history for a product
-   **Description:** Retrieves the stock adjustment history for a specific product.
-   **Request:**
    -   **URL Params:**
        -   `productId` (integer, required): The ID of the product.
-   **Response:**
    -   **200 OK:**
        ```json
        [
            {
                "ID": 1,
                "ProductID": 1,
                "LocationID": 1,
                "Type": "STOCK_OUT",
                "Quantity": 5,
                "ReasonCode": "DAMAGED",
                "Notes": "Product was damaged during handling.",
                "AdjustedBy": 1,
                "AdjustedAt": "2025-11-08T21:00:00Z",
                "PreviousQuantity": 155,
                "NewQuantity": 150
            }
        ]
        ```
    -   **404 Not Found:** If the product is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

## Suppliers

### POST /suppliers

-   **Summary:** Create a new supplier
-   **Description:** Create a new product supplier.
-   **Request:**
    -   **Body:**
        ```json
        {
            "name": "Supplier A",
            "contactPerson": "John Doe",
            "email": "john.doe@supplier-a.com",
            "phone": "123-456-7890",
            "address": "123 Supplier St"
        }
        ```
-   **Response:**
    -   **201 Created:**
        ```json
        {
            "ID": 1,
            "Name": "Supplier A",
            "ContactPerson": "John Doe",
            "Email": "john.doe@supplier-a.com",
            "Phone": "123-456-7890",
            "Address": "123 Supplier St",
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:00:00Z"
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **409 Conflict:** If a supplier with the same name already exists.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /suppliers

-   **Summary:** Get a list of suppliers
-   **Description:** Get a list of all product suppliers.
-   **Request:** None
-   **Response:**
    -   **200 OK:**
        ```json
        [
            {
                "ID": 1,
                "Name": "Supplier A",
                ...
            }
        ]
        ```
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /suppliers/{id}

-   **Summary:** Get a supplier by ID
-   **Description:** Get a single supplier by its ID.
-   **Request:**
    -   **URL Params:**
        -   `id` (integer, required): The ID of the supplier.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "Name": "Supplier A",
            ...
        }
        ```
    -   **404 Not Found:** If the supplier is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /suppliers/name/{name}

-   **Summary:** Get a supplier by Name
-   **Description:** Get a single supplier by its Name.
-   **Request:**
    -   **URL Params:**
        -   `name` (string, required): The Name of the supplier.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "Name": "Supplier A",
            ...
        }
        ```
    -   **404 Not Found:** If the supplier is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### PUT /suppliers/{id}

-   **Summary:** Update an existing supplier
-   **Description:** Update an existing product supplier.
-   **Request:**
    -   **URL Params:**
        -   `id` (integer, required): The ID of the supplier.
    -   **Body:**
        ```json
        {
            "name": "Supplier A Updated",
            ...
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "Name": "Supplier A Updated",
            ...
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **404 Not Found:** If the supplier is not found.
    -   **409 Conflict:** If a supplier with the same name already exists.
    -   **500 Internal Server Error:** If there is a server-side error.

### DELETE /suppliers/{id}

-   **Summary:** Delete a supplier
-   **Description:** Delete a product supplier by its ID. Cannot delete if products are associated.
-   **Request:**
    -   **URL Params:**
        -   `id` (integer, required): The ID of the supplier.
-   **Response:**
    -   **204 No Content:** If the supplier is deleted successfully.
    -   **404 Not Found:** If the supplier is not found.
    -   **409 Conflict:** If the supplier has associated products.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /suppliers/{id}/performance

-   **Summary:** Get supplier performance report
-   **Description:** Generates a mock report on supplier performance (e.g., on-time delivery, quality).
-   **Request:**
    -   **URL Params:**
        -   `id` (integer, required): The ID of the supplier.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "supplierId": 1,
            "supplierName": "Supplier A",
            "averageLeadTimeDays": 7.5,
            "onTimeDeliveryRate": 0.95
        }
        ```
    -   **404 Not Found:** If the supplier is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

## Time Tracking

### POST /time-tracking/clock-in/{userId}

-   **Summary:** Clock in an employee
-   **Description:** Clock in an employee for a new time clock entry.
-   **Request:**
    -   **URL Params:**
        -   `userId` (integer, required): The ID of the user.
    -   **Body:**
        ```json
        {
            "notes": "Starting my shift."
        }
        ```
-   **Response:**
    -   **201 Created:**
        ```json
        {
            "ID": 1,
            "UserID": 1,
            "ClockIn": "2025-11-08T09:00:00Z",
            "ClockOut": null,
            "Notes": "Starting my shift."
        }
        ```
    -   **400 Bad Request:** If the user is already clocked in.
    -   **500 Internal Server Error:** If there is a server-side error.

### POST /time-tracking/clock-out/{userId}

-   **Summary:** Clock out an employee
-   **Description:** Clock out an employee, completing their time clock entry.
-   **Request:**
    -   **URL Params:**
        -   `userId` (integer, required): The ID of the user.
    -   **Body:**
        ```json
        {
            "notes": "Ending my shift."
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "UserID": 1,
            "ClockIn": "2025-11-08T09:00:00Z",
            "ClockOut": "2025-11-08T17:00:00Z",
            "Notes": "Ending my shift."
        }
        ```
    -   **400 Bad Request:** If the user is not clocked in.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /time-tracking/last-entry/{userId}

-   **Summary:** Get the last time clock entry for a user
-   **Description:** Get the last time clock entry for a user by their user ID.
-   **Request:**
    -   **URL Params:**
        -   `userId` (integer, required): The ID of the user.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "UserID": 1,
            "ClockIn": "2025-11-08T09:00:00Z",
            "ClockOut": "2025-11-08T17:00:00Z",
            "Notes": "Ending my shift."
        }
        ```
    -   **404 Not Found:** If no time clock entry is found for the user.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /time-tracking/last-entry/username/{username}

-   **Summary:** Get the last time clock entry for a user by username
-   **Description:** Get the last time clock entry for a user by their username.
-   **Request:**
    -   **URL Params:**
        -   `username` (string, required): The username of the user.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "UserID": 1,
            "ClockIn": "2025-11-08T09:00:00Z",
            "ClockOut": "2025-11-08T17:00:00Z",
            "Notes": "Ending my shift."
        }
        ```
    -   **404 Not Found:** If no time clock entry is found for the user.
    -   **500 Internal Server Error:** If there is a server-side error.

## Users

### POST /users/register

-   **Summary:** Register a new user
-   **Description:** Create a new user account. The first user registered will be an Admin and active by default. Subsequent users will be inactive by default and must be approved by an Admin.
-   **Request:**
    -   **Body:**
        ```json
        {
            "username": "newuser",
            "password": "password123",
            "role": "Staff"
        }
        ```
-   **Response:**
    -   **201 Created:**
        ```json
        {
            "ID": 2,
            "Username": "newuser",
            "Role": "Staff",
            "IsActive": false,
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:00:00Z"
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid or if the first user is not an Admin.
    -   **500 Internal Server Error:** If there is a server-side error.

### POST /users/login

-   **Summary:** Log in a user
-   **Description:** Authenticate user and return a JWT token.
-   **Request:**
    -   **Body:**
        ```json
        {
            "username": "admin",
            "password": "password123"
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "accessToken": "...",
            "refreshToken": "...",
            "user": { ... }
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **401 Unauthorized:** If the credentials are invalid or the user is not active.
    -   **500 Internal Server Error:** If there is a server-side error.

### POST /users/refresh-token

-   **Summary:** Refresh access token
-   **Description:** Get a new access token using a refresh token.
-   **Request:**
    -   **Body:**
        ```json
        {
            "refreshToken": "..."
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "accessToken": "...",
            "refreshToken": "..."
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **401 Unauthorized:** If the refresh token is invalid.
    -   **500 Internal Server Error:** If there is a server-side error.

### POST /users/logout

-   **Summary:** Log out a user
-   **Description:** Invalidate both the access and refresh tokens.
-   **Request:**
    -   **Body:**
        ```json
        {
            "refreshToken": "..."
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "message": "Logout successful"
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /users

-   **Summary:** List users with optional status and search filters
-   **Description:** Retrieves all users, optionally filtered by status (approved/pending) and search query (username or ID).
-   **Request:**
    -   **Query Params:**
        -   `status` (string, optional): Filter by user status (approved, pending).
        -   `q` (string, optional): Search by username or ID.
-   **Response:**
    -   **200 OK:**
        ```json
        [
            {
                "ID": 1,
                "Username": "admin",
                "Role": "Admin",
                "IsActive": true,
                "CreatedAt": "2025-11-08T21:00:00Z",
                "UpdatedAt": "2025-11-08T21:00:00Z"
            }
        ]
        ```
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /users/{id}

-   **Summary:** Get user details by ID
-   **Description:** Get details of a specific user by ID.
-   **Request:**
    -   **URL Params:**
        -   `id` (integer, required): The ID of the user.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "Username": "admin",
            "Role": "Admin",
            "IsActive": true,
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:00:00Z"
        }
        ```
    -   **404 Not Found:** If the user is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### PUT /users/{id}

-   **Summary:** Update user details
-   **Description:** Update details of a specific user by ID. Only Admins can change user roles.
-   **Request:**
    -   **URL Params:**
        -   `id` (integer, required): The ID of the user.
    -   **Body:**
        ```json
        {
            "username": "newusername",
            "password": "newpassword123",
            "role": "Staff"
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 1,
            "Username": "newusername",
            "Role": "Staff",
            "IsActive": true,
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:05:00Z"
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **403 Forbidden:** If a non-Admin user tries to change a user's role.
    -   **404 Not Found:** If the user is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### DELETE /users/{id}

-   **Summary:** Delete a user
-   **Description:** Delete a specific user by ID.
-   **Request:**
    -   **URL Params:**
        -   `id` (integer, required): The ID of the user.
-   **Response:**
    -   **204 No Content:** If the user is deleted successfully.
    -   **404 Not Found:** If the user is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

### PUT /users/{id}/approve

-   **Summary:** Approve a user
-   **Description:** Activate a user's account by setting IsActive to true.
-   **Request:**
    -   **URL Params:**
        -   `id` (integer, required): The ID of the user.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "ID": 2,
            "Username": "newuser",
            "Role": "Staff",
            "IsActive": true,
            "CreatedAt": "2025-11-08T21:00:00Z",
            "UpdatedAt": "2025-11-08T21:05:00Z"
        }
        ```
    -   **404 Not Found:** If the user is not found.
    -   **500 Internal Server Error:** If there is a server-side error.

## Webhooks

### POST /webhooks

-   **Summary:** Handle incoming webhooks
-   **Description:** A generic endpoint to handle incoming webhooks from various third-party integrations.
-   **Request:**
    -   **Body:**
        ```json
        {
            "source": "shopify",
            "event": "order_created",
            "data": { ... }
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "message": "Webhook received"
        }
        ```
    -   **400 Bad Request:** If the webhook payload is invalid.
    -   **500 Internal Server Error:** If there is a server-side error.

## WebSocket

### GET /ws

-   **Summary:** WebSocket endpoint
-   **Description:** Establishes a WebSocket connection for real-time communication.
-   **Request:** None
-   **Response:**
    -   **101 Switching Protocols:** If the WebSocket upgrade is successful.

## Payments

### POST /payment/create

-   **Summary:** Create a new payment
-   **Description:** Initiates a new payment with the specified payment method (bKash, card, or cash).
-   **Request:**
    -   **Body:**
        ```json
        {
            "order_id": "ORDER123",
            "amount": 100.50,
            "payment_method": "bkash"
        }
        ```
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "redirect_url": "https://gateway.bkash.com/..."
        }
        ```
    -   **400 Bad Request:** If the request payload is invalid.
    -   **500 Internal Server Error:** If there is a server-side error.

### GET /payment/bkash/callback

-   **Summary:** bKash payment callback
-   **Description:** The callback URL for bKash to redirect to after a payment attempt.
-   **Request:**
    -   **Query Params:**
        -   `paymentID` (string, required): The payment ID from bKash.
        -   `status` (string, required): The status of the payment attempt (e.g., "success", "failure", "cancelled").
-   **Response:**
    -   **302 Found:** Redirects to the success or failure URL.

### POST /payment/success

-   **Summary:** SSLCommerz success URL
-   **Description:** The URL for SSLCommerz to redirect to after a successful payment.
-   **Request:**
    -   **Form Data:** Contains transaction details from SSLCommerz.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "status": "Payment successful"
        }
        ```

### POST /payment/fail

-   **Summary:** SSLCommerz fail URL
-   **Description:** The URL for SSLCommerz to redirect to after a failed payment.
-   **Request:**
    -   **Form Data:** Contains transaction details from SSLCommerz.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "status": "Payment failed"
        }
        ```

### POST /payment/cancel

-   **Summary:** SSLCommerz cancel URL
-   **Description:** The URL for SSLCommerz to redirect to after a cancelled payment.
-   **Request:**
    -   **Form Data:** Contains transaction details from SSLCommerz.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "status": "Payment cancelled"
        }
        ```

### POST /payment/ipn

-   **Summary:** SSLCommerz Instant Payment Notification (IPN) listener
-   **Description:** The endpoint for SSLCommerz to send asynchronous payment status updates.
-   **Request:**
    -   **Form Data:** Contains transaction details from SSLCommerz.
-   **Response:**
    -   **200 OK:**
        ```json
        {
            "status": "SSLCommerz IPN handled successfully"
        }
        ```
    -   **400 Bad Request:** If the IPN data is invalid.
