package requests

type StockTransferRequest struct {
	ProductID      uint `json:"productId" binding:"required"`
	SourceLocationID uint `json:"sourceLocationId" binding:"required"`
	DestLocationID   uint `json:"destLocationId" binding:"required"`
	Quantity       int  `json:"quantity" binding:"required"`
}
