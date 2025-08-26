package entity

import (
	"time"

	"github.com/google/uuid"
)

type AccessToken struct {
	UserID    uint
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type RefreshToken struct {
	UUID        uuid.UUID
	CreatedAt   time.Time
	ExpiresAt   time.Time
	RevokedAt   *time.Time
	Fingerprint string
	UserID      uint
}

func (rt RefreshToken) IsValid() bool {
	now := time.Now()
	if rt.RevokedAt != nil && rt.RevokedAt.Before(now) {
		return false
	}
	return rt.ExpiresAt.After(now)
}
