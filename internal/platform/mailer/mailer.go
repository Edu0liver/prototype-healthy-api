// Package mailer sends transactional email via Resend. When no API key is
// configured it degrades to logging (dev), so flows like invites still work.
package mailer

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/config"
	"github.com/resend/resend-go/v2"
	"go.uber.org/zap"
)

// Mailer sends emails.
type Mailer struct {
	client *resend.Client
	from   string
	log    *zap.Logger
}

// New builds the mailer.
func New(cfg *config.Config, log *zap.Logger) *Mailer {
	m := &Mailer{from: cfg.Email.FromAddress, log: log}
	if cfg.Email.ResendAPIKey != "" {
		m.client = resend.NewClient(cfg.Email.ResendAPIKey)
	}
	return m
}

// Send delivers an HTML email. In dev (no client) it logs instead.
func (m *Mailer) Send(to, subject, html string) error {
	if m.client == nil {
		m.log.Info("email (dev, not sent)", zap.String("to", to), zap.String("subject", subject))
		return nil
	}
	_, err := m.client.Emails.Send(&resend.SendEmailRequest{
		From:    m.from,
		To:      []string{to},
		Subject: subject,
		Html:    html,
	})
	return err
}
