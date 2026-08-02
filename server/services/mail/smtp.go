package mail

import (
	"context"
	"fmt"

	gomail "github.com/wneessen/go-mail"
)

type SMTPMailer struct {
	from   string
	client *gomail.Client
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func NewSMTPMailer(cfg SMTPConfig) (*SMTPMailer, error) {
	client, err := gomail.NewClient(cfg.Host,
		gomail.WithPort(cfg.Port),
		gomail.WithSMTPAuth(gomail.SMTPAuthPlain),
		gomail.WithUsername(cfg.Username),
		gomail.WithPassword(cfg.Password),
		gomail.WithTLSPolicy(gomail.TLSMandatory),
	)
	if err != nil {
		return nil, fmt.Errorf("create smtp client: %w", err)
	}
	return &SMTPMailer{from: cfg.From, client: client}, nil
}

func (m *SMTPMailer) Send(ctx context.Context, to, subject, body string) error {
	msg := gomail.NewMsg()
	if err := msg.From(m.from); err != nil {
		return fmt.Errorf("set from: %w", err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("set to: %w", err)
	}
	msg.Subject(subject)
	msg.SetBodyString(gomail.TypeTextPlain, body)

	return m.client.DialAndSendWithContext(ctx, msg)
}
