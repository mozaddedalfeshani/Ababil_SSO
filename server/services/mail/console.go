package mail

import (
	"context"
	"log/slog"
)

// ConsoleMailer logs the email instead of sending it — the dev/self-host
// default when SMTP isn't configured, so verify/reset links are
// discoverable without setting up a mail server first.
type ConsoleMailer struct{}

func (ConsoleMailer) Send(ctx context.Context, to, subject, body string) error {
	slog.Info("email (console mailer, not actually sent)", "to", to, "subject", subject, "body", body)
	return nil
}
