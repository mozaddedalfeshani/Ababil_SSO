package mail

import "fmt"

// Templates are plain text on purpose — no HTML/JS in transactional
// auth email reduces phishing-look-alike risk and keeps the mailer
// dependency-free.

func VerifyEmailMessage(otp string) (subject, body string) {
	subject = "Verify your email"
	body = fmt.Sprintf(
		"Your Ababil SSO verification code is:\n\n%s\n\nEnter this code on the verify-email page. It expires in 10 minutes. If you didn't request this, ignore this email.",
		otp,
	)
	return
}

func ResetPasswordMessage(otp string) (subject, body string) {
	subject = "Reset your password"
	body = fmt.Sprintf(
		"Your Ababil SSO password-reset code is:\n\n%s\n\nEnter this code with your new password on the reset page. It expires in 10 minutes and can only be used once. If you didn't request this, you can ignore this email — your password will not change.",
		otp,
	)
	return
}
