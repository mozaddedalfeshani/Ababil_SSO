package models

import "time"

type SigningKeyStatus string

const (
	SigningKeyActive  SigningKeyStatus = "active"
	SigningKeyNext    SigningKeyStatus = "next"
	SigningKeyRetired SigningKeyStatus = "retired"
)

type SigningKey struct {
	KID           string
	Alg           string
	PrivateKeyEnc []byte
	PublicJWK     []byte // raw JSON, assembled/parsed by services/keys
	Status        SigningKeyStatus
	ActivatesAt   time.Time
	RetiresAt     *time.Time
	CreatedAt     time.Time
}
