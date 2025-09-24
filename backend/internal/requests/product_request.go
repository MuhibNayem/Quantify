package requests

// ProductCreateRequest represents the request body for creating a new product.
type ProductCreateRequest struct {
	SKU           string  `json:"sku" binding:"required"`
	Name          string  `json:"name" binding:"required"`
	Description   string  `json:"description"`
	CategoryID    uint    `json:"categoryId" binding:"required"`
	SubCategoryID uint    `json:"subCategoryId"`
	SupplierID    uint    `json:"supplierId" binding:"required"`
	Brand         string  `json:"brand"`
	PurchasePrice float64 `json:"purchasePrice"`
	SellingPrice  float64 `json:"sellingPrice" binding:"required"`
	BarcodeUPC    string  `json:"barcodeUpc"`
	ImageURLs     string  `json:"imageUrls"`
	Status        string  `json:"status"`
	LocationID    uint    `json:"locationId" binding:"required"`
}

// ProductUpdateRequest represents the request body for updating an existing product.
type ProductUpdateRequest struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	CategoryID    uint    `json:"categoryId"`
	SubCategoryID uint    `json:"subCategoryId"`
	SupplierID    uint    `json:"supplierId"`
	Brand         string  `json:"brand"`
	PurchasePrice float64 `json:"purchasePrice"`
	SellingPrice  float64 `json:"sellingPrice"`
	BarcodeUPC    string  `json:"barcodeUpc"`
	ImageURLs     string  `json:json:"imageUrls"`
	Status        string  `json:"status"`
	LocationID    uint    `json:"locationId"`
}

// ProductArchiveRequest represents the request body for archiving a product.
type ProductArchiveRequest struct {
	Status string `json:"status" binding:"required,eq=Archived"`
}
