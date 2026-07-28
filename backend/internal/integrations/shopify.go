package integrations

import "github.com/gin-gonic/gin"

// Shopify receives and validates incoming webhooks from a Shopify store.
// At present this is a stub that acknowledges the request; real HMAC
// verification and order-event translation will be added in a follow-up.
type Shopify struct{}

func (s *Shopify) Name() string { return "shopify" }

func (s *Shopify) HandleWebhook(c *gin.Context) error {
	// TODO: verify HMAC, parse order events, translate to internal domain
	return nil
}