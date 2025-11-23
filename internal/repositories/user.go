package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/okocraft/auth-service/api/user"
	"github.com/okocraft/auth-service/internal/domain"
	"github.com/okocraft/auth-service/internal/repositories/database"
	"github.com/okocraft/auth-service/internal/repositories/queries"
)

type UserRepository interface {
	GetUserIDBySub(ctx context.Context, conn database.Connection, sub string) (user.ID, error)
	GetUserIDByLoginKey(ctx context.Context, conn database.Connection, loginKey domain.LoginKey) (user.ID, error)
	SaveLoginKeyForUserID(ctx context.Context, conn database.Connection, id user.ID, loginKey domain.LoginKey, now time.Time) error
	DeleteLoginKeyByUserID(ctx context.Context, conn database.Connection, id user.ID) error
	SaveUserSub(ctx context.Context, conn database.Connection, userID user.ID, sub string, now time.Time) error
}

func NewUserRepository() UserRepository {
	return &userRepository{}
}

type userRepository struct{}

func (r userRepository) GetUserIDBySub(ctx context.Context, conn database.Connection, sub string) (user.ID, error) {
	id, err := conn.Queries().GetUserIDBySub(ctx, sub)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, domain.UserNotFoundBySubError
	} else if err != nil {
		return 0, database.NewDBErrorWithStackTrace(err)
	}

	return user.ID(id), nil
}

func (r userRepository) GetUserIDByLoginKey(ctx context.Context, conn database.Connection, loginKey domain.LoginKey) (user.ID, error) {
	id, err := conn.Queries().GetUserIDByLoginKey(ctx, int64(loginKey))
	if errors.Is(err, sql.ErrNoRows) {
		return 0, domain.UserNotFoundByLoginKeyError
	} else if err != nil {
		return 0, database.NewDBErrorWithStackTrace(err)
	}

	return user.ID(id), nil
}

func (r userRepository) SaveLoginKeyForUserID(ctx context.Context, conn database.Connection, id user.ID, loginKey domain.LoginKey, now time.Time) error {
	err := conn.Queries().InsertLoginKeyForUserID(ctx, queries.InsertLoginKeyForUserIDParams{
		UserID:    int32(id),
		LoginKey:  int64(loginKey),
		CreatedAt: now,
	})
	if err != nil {
		return database.NewDBErrorWithStackTrace(err)
	}
	return nil
}

func (r userRepository) DeleteLoginKeyByUserID(ctx context.Context, conn database.Connection, id user.ID) error {
	err := conn.Queries().DeleteLoginKey(ctx, int32(id))
	if err != nil {
		return database.NewDBErrorWithStackTrace(err)
	}
	return nil
}

func (r userRepository) SaveUserSub(ctx context.Context, conn database.Connection, userID user.ID, sub string, now time.Time) error {
	row, err := conn.Queries().InsertSubForUserID(ctx, queries.InsertSubForUserIDParams{
		UserID:    int32(userID),
		Sub:       sub,
		CreatedAt: now,
	})
	if err != nil {
		return database.NewDBErrorWithStackTrace(err)
	} else if row == 0 {
		return domain.SubAlreadyLinkedError
	}

	return nil
}
