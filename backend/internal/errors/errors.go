package errors

import "fmt"
import "net/http"

// AppError represents a custom application error with a message and an HTTP status code.
type AppError struct {
	Message    string
	StatusCode int
	Err        error // Original error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// NewAppError creates a new AppError.
func NewAppError(message string, statusCode int, err error) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// Common error types
var (
	ErrNotFound      = NewAppError("Resource not found", http.StatusNotFound, nil)
	ErrBadRequest    = NewAppError("Bad request", http.StatusBadRequest, nil)
	ErrUnauthorized  = NewAppError("Unauthorized", http.StatusUnauthorized, nil)
	ErrForbidden     = NewAppError("Forbidden", http.StatusForbidden, nil)
	ErrConflict      = NewAppError("Conflict", http.StatusConflict, nil)
	ErrInternal      = NewAppError("Internal server error", http.StatusInternalServerError, nil)
	ErrInvalidInput  = NewAppError("Invalid input", http.StatusBadRequest, nil)
	ErrAlreadyExists = NewAppError("Resource already exists", http.StatusConflict, nil)
)
