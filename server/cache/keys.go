package cache

import "fmt"

// Typed key builders keep the Redis key namespace in one place instead
// of scattered fmt.Sprintf calls across handlers/services.

func AuthRequestKey(id string) string { return fmt.Sprintf("authreq:%s", id) }

// MFAPendingKey scopes the half-authenticated state between password
// success and TOTP verification. Only /api/auth/login/totp may consume
// it; nothing else should ever read this key.
func MFAPendingKey(id string) string { return fmt.Sprintf("mfa_pending:%s", id) }

func RateLimitKey(bucket, key string) string { return fmt.Sprintf("rl:%s:%s", bucket, key) }

func LoginFailKey(emailHash string) string { return fmt.Sprintf("login_fail:%s", emailHash) }

// OTPFailKey counts wrong email-OTP guesses per user+purpose. After
// max attempts the outstanding OTP is invalidated.
func OTPFailKey(purpose, userID string) string {
	return fmt.Sprintf("otp_fail:%s:%s", purpose, userID)
}

func IPSaltKey(yyyymmdd string) string { return fmt.Sprintf("ipsalt:%s", yyyymmdd) }
