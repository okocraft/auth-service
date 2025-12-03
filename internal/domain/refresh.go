package domain

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/okocraft/authlib/user"
)

type RefreshTokenParams struct {
	UserID         user.ID
	RefreshTokenID int64
	LoginID        uuid.UUID
	MaxExpiresAt   time.Time
}

type RefreshedToken struct {
	RefreshToken string
	AccessToken  string
	ExpiresAt    time.Time
}
