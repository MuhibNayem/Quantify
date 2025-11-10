# Analysis of Intuitive Query Parameters

## 1. Introduction

This document provides an analysis of the backend codebase to identify opportunities for replacing internal database IDs with more user-friendly and intuitive identifiers in API endpoints. The goal is to make the API more accessible and easier to use for developers and end-users.

## 2. Analysis of Current Implementation

The current backend implementation consistently uses database primary keys (IDs) in API routes for fetching, updating, and deleting single resources. This pattern is prevalent across various modules of the application.

**Key Observations:**

*   **API Route Structure:** The API routes are structured as `/resource/:id`, where `:id` is the database primary key. This is evident in `internal/router/router.go`.
*   **Handler Logic:** The handlers, located in `internal/handlers/`, extract the ID from the URL using `c.Param("id")` and then use it to query the database. For example, in `internal/handlers/products.go`, the `GetProduct` function fetches a product by its ID.
*   **Data Models:** The data models, defined in `internal/domain/models.go`, contain several fields that are both unique and human-readable, making them excellent candidates for use in API routes.

## 3. Opportunities for Improvement

Here are the specific areas where we can introduce more intuitive query parameters:

### 3.1. Products

*   **Current:** `/products/:id`
*   **Analysis:** The `Product` model has two unique and user-friendly fields: `SKU` and `BarcodeUPC`. These are much more intuitive for users to work with than an internal database ID.
*   **Recommendation:**
    *   Introduce new endpoints:
        *   `/products/sku/:sku`
        *   `/products/barcode/:barcode`
    *   These endpoints would allow users to fetch, update, and delete products using their SKU or barcode.

### 3.2. Categories

*   **Current:** `/categories/:id`
*   **Analysis:** The `Category` model has a unique `Name` field.
*   **Recommendation:**
    *   Introduce a new endpoint:
        *   `/categories/name/:name`
    *   This would allow for managing categories by their name.

### 3.3. Suppliers

*   **Current:** `/suppliers/:id`
*   **Analysis:** The `Supplier` model has a unique `Name` field.
*   **Recommendation:**
    *   Introduce a new endpoint:
        *   `/suppliers/name/:name`
    *   This would allow for managing suppliers by their name.

## 4. Proposed Implementation Strategy

To implement these changes without breaking the existing API, we can introduce new endpoints and repository functions.

1.  **Update the Repository Layer:**
    *   For each resource, create new functions in the `internal/repository/` directory to fetch data by the new, user-friendly identifiers.
    *   **Example (`internal/repository/product_repository.go`):**
        ```go
        func (r *ProductRepository) GetProductBySKU(sku string) (*domain.Product, error) {
            var product domain.Product
            if err := r.db.Where("sku = ?", sku).First(&product).Error; err != nil {
                return nil, err
            }
            return &product, nil
        }
        ```

2.  **Update the Router:**
    *   In `internal/router/router.go`, add the new routes.
    *   **Example:**
        ```go
        productRoutes := router.Group("/products")
        {
            // ... existing routes
            productRoutes.GET("/sku/:sku", productHandler.GetProductBySKU)
        }
        ```

3.  **Update the Handlers:**
    *   Create new handler functions to handle the new routes.
    *   These handlers will call the new repository functions.
    *   **Example (`internal/handlers/products.go`):**
        ```go
        func (h *ProductHandler) GetProductBySKU(c *gin.Context) {
            sku := c.Param("sku")
            product, err := h.productRepo.GetProductBySKU(sku)
            if err != nil {
                // ... error handling
                return
            }
            c.JSON(http.StatusOK, product)
        }
        ```

## 5. Conclusion

By introducing these new endpoints, we can significantly improve the usability and intuitiveness of the API. This change will make the API more developer-friendly and align it better with the real-world workflows of an inventory management system. The proposed strategy of adding new endpoints ensures backward compatibility while providing enhanced functionality.
