package services

import (
	"fmt"
	"inventory/backend/internal/config"
	"inventory/backend/internal/domain"
	"net/smtp"
	"strings"
)

type EmailService interface {
	SendPurchaseOrderEmail(po domain.PurchaseOrder) error
}

type emailService struct {
	cfg *config.Config
}

func NewEmailService(cfg *config.Config) EmailService {
	return &emailService{cfg: cfg}
}

func (s *emailService) SendPurchaseOrderEmail(po domain.PurchaseOrder) error {
	if s.cfg.SMTPHost == "" {
		return fmt.Errorf("SMTP host not configured")
	}

	to := []string{po.Supplier.Email}
	if len(to) == 0 || to[0] == "" {
		return fmt.Errorf("supplier email not found")
	}

	subject := fmt.Sprintf("Purchase Order #%d from Quantify", po.ID)

	// Build Email Body
	var body strings.Builder
	body.WriteString(fmt.Sprintf("Dear %s,\n\n", po.Supplier.ContactPerson))
	body.WriteString(fmt.Sprintf("Please find attached Purchase Order #%d dated %s.\n\n", po.ID, po.OrderDate.Format("2006-01-02")))
	body.WriteString("Items:\n")

	for _, item := range po.PurchaseOrderItems {
		productName := item.Product.Name
		if productName == "" {
			productName = fmt.Sprintf("Product ID %d", item.ProductID)
		}
		body.WriteString(fmt.Sprintf("- %s (SKU: %s): %d units @ $%.2f\n", productName, item.Product.SKU, item.OrderedQuantity, item.UnitPrice))
	}

	body.WriteString("\nExpected Delivery: ")
	if po.ExpectedDeliveryDate != nil {
		body.WriteString(po.ExpectedDeliveryDate.Format("2006-01-02"))
	} else {
		body.WriteString("ASAP")
	}

	body.WriteString("\n\nThank you,\nQuantify Inventory Team")

	// Construct Message
	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", to[0], subject, body.String()))

	// Send Email
	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, s.cfg.SMTPHost)
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)

	if err := smtp.SendMail(addr, auth, s.cfg.SMTPSender, to, msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
