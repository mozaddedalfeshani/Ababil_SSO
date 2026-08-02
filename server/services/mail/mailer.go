// Package mail defines the outbound-email boundary. The interface is
// pluggable so self-hosters bring their own SMTP and the project
// never depends on a third-party email API at runtime.
package mail

import "context"

type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}
