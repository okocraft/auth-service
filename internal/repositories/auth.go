package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/okocraft/auth-service/internal/domain"
	"github.com/okocraft/auth-service/internal/repositories/database"
	"github.com/okocraft/auth-service/internal/repositories/queries"
	"github.com/okocraft/authlib/user"
)

type AuthRepository interface {
	SaveRefreshToken(ctx context.Context, conn database.Connection, userID user.ID, jti uuid.UUID, loginID uuid.UUID, createdAt time.Time) error
	SaveAccessToken(ctx context.Context, conn database.Connection, refreshTokenID int64, jti uuid.UUID, createdAt time.Time) error
	GetUserIDAndRefreshTokenIDFromJTI(ctx context.Context, conn database.Connection, jti uuid.UUID) (user.ID, int64, error)
	DeleteAccessTokensByLoginID(ctx context.Context, conn database.Connection, loginID uuid.UUID) error
	DeleteRefreshTokensByLoginID(ctx context.Context, conn database.Connection, loginID uuid.UUID) error
	DeleteExpiredAccessTokens(ctx context.Context, conn database.Connection, expiredAt time.Time) (int64, error)
	DeleteExpiredRefreshTokens(ctx context.Context, conn database.Connection, expiredAt time.Time) (int64, error)
}

func NewAuthRepository() AuthRepository {
	return &authRepository{}
}

type authRepository struct{}

func (r authRepository) SaveRefreshToken(ctx context.Context, conn database.Connection, userID user.ID, jti uuid.UUID, loginID uuid.UUID, createdAt time.Time) error {
	q := conn.Queries()
	err := q.InsertRefreshToken(ctx, queries.InsertRefreshTokenParams{
		UserID:    int32(userID),
		Jti:       jti.Bytes(),
		LoginID:   loginID.Bytes(),
		CreatedAt: createdAt,
	})
	if err != nil {
		return database.NewDBErrorWithStackTrace(err)
	}
	return nil
}

func (r authRepository) SaveAccessToken(ctx context.Context, conn database.Connection, refreshTokenID int64, jti uuid.UUID, createdAt time.Time) error {
	q := conn.Queries()
	err := q.InsertAccessToken(ctx, queries.InsertAccessTokenParams{RefreshTokenID: refreshTokenID, Jti: jti.Bytes(), CreatedAt: createdAt})
	if err != nil {
		return database.NewDBErrorWithStackTrace(err)
	}
	return nil
}

func (r authRepository) GetUserIDAndRefreshTokenIDFromJTI(ctx context.Context, conn database.Connection, jti uuid.UUID) (user.ID, int64, error) {
	q := conn.Queries()
	row, err := q.GetUserIDAndRefreshTokenIDByJTI(ctx, jti.Bytes())
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, domain.RefreshTokenIDByJTINotFoundError
	} else if err != nil {
		return 0, 0, database.NewDBErrorWithStackTrace(err)
	}

	return user.ID(row.UserID), row.ID, nil
}

func (r authRepository) DeleteAccessTokensByLoginID(ctx context.Context, conn database.Connection, loginID uuid.UUID) error {
	q := conn.Queries()
	err := q.DeleteAccessTokensByLoginID(ctx, loginID.Bytes())
	if err != nil {
		return database.NewDBErrorWithStackTrace(err)
	}
	return nil
}

func (r authRepository) DeleteRefreshTokensByLoginID(ctx context.Context, conn database.Connection, loginID uuid.UUID) error {
	q := conn.Queries()
	err := q.DeleteRefreshTokensByLoginID(ctx, loginID.Bytes())
	if err != nil {
		return database.NewDBErrorWithStackTrace(err)
	}
	return nil
}

func (r authRepository) DeleteExpiredAccessTokens(ctx context.Context, conn database.Connection, expiredAt time.Time) (int64, error) {
	q := conn.Queries()
	rows, err := q.DeleteExpiredAccessTokens(ctx, expiredAt)
	if err != nil {
		return 0, database.NewDBErrorWithStackTrace(err)
	}
	return rows, nil
}

func (r authRepository) DeleteExpiredRefreshTokens(ctx context.Context, conn database.Connection, expiredAt time.Time) (int64, error) {
	q := conn.Queries()
	rows, err := q.DeleteExpiredRefreshTokens(ctx, expiredAt)
	if err != nil {
		return 0, database.NewDBErrorWithStackTrace(err)
	}
	return rows, nil
}
