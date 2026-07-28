package integrations

import "github.com/gin-gonic/gin"

// QuickBooks receives and validates incoming webhooks from QuickBooks
// Online. At present this is a stub that acknowledges the request; real
// signature verification and accounting-event translation will be added
// in a follow-up.
type QuickBooks struct{}

func (q *QuickBooks) Name() string { return "quickbooks" }

func (q *QuickBooks) HandleWebhook(c *gin.Context) error {
	// TODO: verify signature, parse accounting events
	return nil
}