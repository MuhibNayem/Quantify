package requests

// BulkImportConfirmRequest represents the request body for confirming a bulk import.
type BulkImportConfirmRequest struct {
	JobID string `json:"jobId" binding:"required"`
}
