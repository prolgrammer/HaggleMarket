package model

import (
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/google/uuid"
)

type RefreshToken struct {
	UUID        uuid.UUID  `db:"uuid"`
	CreatedAt   time.Time  `db:"created_at"`
	ExpiresAt   time.Time  `db:"expires_at"`
	RevokedAt   *time.Time `db:"revoked_at"`
	Fingerprint string     `db:"fingerprint"`
	UserID      uint       `db:"user_id"`
}

func ConvertTokenFromEntity(t entity.RefreshToken) RefreshToken {
	return RefreshToken{
		UUID:        t.UUID,
		CreatedAt:   t.CreatedAt,
		ExpiresAt:   t.ExpiresAt,
		RevokedAt:   t.RevokedAt,
		Fingerprint: t.Fingerprint,
		UserID:      t.UserID,
	}
}

func ConvertTokenToEntity(t RefreshToken) entity.RefreshToken {
	return entity.RefreshToken{
		UUID:        t.UUID,
		CreatedAt:   t.CreatedAt,
		ExpiresAt:   t.ExpiresAt,
		RevokedAt:   t.RevokedAt,
		Fingerprint: t.Fingerprint,
		UserID:      t.UserID,
	}
}
