package integrations

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/middleware"
)

// Integration is the contract every third-party webhook provider must
// satisfy. The registry uses Name() to dispatch incoming HTTP webhooks
// to the correct handler.
type Integration interface {
	Name() string
	HandleWebhook(c *gin.Context) error
}

// Registry stores integrations keyed by Name() and dispatches incoming
// webhook requests to the matching implementation.
type Registry struct {
	integrations map[string]Integration
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{integrations: make(map[string]Integration)}
}

// Register adds (or replaces) an integration under its Name().
func (r *Registry) Register(i Integration) {
	r.integrations[i.Name()] = i
}

// HandleWebhook is a gin.HandlerFunc-compatible dispatcher. It looks up
// the integration by the :name path parameter, returns 404 if unknown,
// and writes 500 if the handler returns an error.
func (r *Registry) HandleWebhook(c *gin.Context) {
	name := c.Param("name")
	i, ok := r.integrations[name]
	if !ok {
		middleware.RespondError(c, appErrors.NewAppError("Unknown integration", http.StatusNotFound, nil))
		return
	}
	if err := i.HandleWebhook(c); err != nil {
		middleware.RespondError(c, appErrors.NewAppError("Integration handler failed", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}