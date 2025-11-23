package usecases

import (
	"context"
	"time"

	"github.com/okocraft/auth-service/api/user"
	"github.com/okocraft/auth-service/internal/domain"
	"github.com/okocraft/auth-service/internal/repositories"
	"github.com/okocraft/auth-service/internal/repositories/database"
)

type UserUsecase interface {
	GetUserIDBySub(ctx context.Context, sub string) (user.ID, error)
	SaveSubByLoginKey(ctx context.Context, loginKey domain.LoginKey, sub string) (user.ID, error)
}

func NewUserUsecase(db database.DB, repo repositories.UserRepository) UserUsecase {
	return &userUsecase{
		db:   db,
		repo: repo,
	}
}

type userUsecase struct {
	db   database.DB
	repo repositories.UserRepository
}

func (u userUsecase) GetUserIDBySub(ctx context.Context, sub string) (user.ID, error) {
	id, err := u.repo.GetUserIDBySub(ctx, u.db.Conn(), sub)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (u userUsecase) SaveSubByLoginKey(ctx context.Context, loginKey domain.LoginKey, sub string) (user.ID, error) {
	var result user.ID
	err := u.db.WithTx(ctx, func(ctx context.Context, tx database.Connection) error {
		id, err := u.repo.GetUserIDByLoginKey(ctx, tx, loginKey)
		if err != nil {
			return err
		}

		err = u.repo.DeleteLoginKeyByUserID(ctx, tx, id)
		if err != nil {
			return err
		}

		err = u.repo.SaveUserSub(ctx, tx, id, sub, time.Now())
		if err != nil {
			return err
		}

		result = id
		return nil
	})
	if err != nil {
		return 0, err
	}

	return result, nil
}
