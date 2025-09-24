package notifications

import (
	"fmt"
	"inventory/backend/internal/config"
	"net/smtp"
)

func SendEmail(cfg *config.Config, to, subject, body string) error {
	if cfg.SMTPHost == "" {
		return fmt.Errorf("SMTP host not configured")
	}

	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost)

	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)

	return smtp.SendMail(addr, auth, cfg.SMTPSender, []string{to}, msg)
}
