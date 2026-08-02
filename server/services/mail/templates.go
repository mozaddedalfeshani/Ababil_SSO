package mail

import "fmt"

// Templates are plain text on purpose — no HTML/JS in transactional
// auth email reduces phishing-look-alike risk and keeps the mailer
// dependency-free.

func VerifyEmailMessage(verifyURL string) (subject, body string) {
	subject = "Verify your email"
	body = fmt.Sprintf(
		"Confirm your email address to finish setting up your account:\n\n%s\n\nThis link expires in 24 hours. If you didn't request this, ignore this email.",
		verifyURL,
	)
	return
}

func ResetPasswordMessage(resetURL string) (subject, body string) {
	subject = "Reset your password"
	body = fmt.Sprintf(
		"A password reset was requested for your account:\n\n%s\n\nThis link expires in 1 hour and can only be used once. If you didn't request this, you can ignore this email — your password will not change.",
		resetURL,
	)
	return
}
