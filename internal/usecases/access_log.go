package usecases

import (
	"context"

	"github.com/Siroshun09/serrors"
	"github.com/okocraft/auth-service/api/user"
	"github.com/okocraft/auth-service/internal/domain"
	"github.com/okocraft/auth-service/internal/repositories"
	"github.com/okocraft/auth-service/internal/repositories/database"
)

type AccessLogUsecase interface {
	SaveAccessLogByUserID(ctx context.Context, userID user.ID, accessLog domain.AccessLogParams) error
}

func NewAccessLogUsecase(db database.DB, repo repositories.AccessLogRepository, userRepo repositories.UserRepository) AccessLogUsecase {
	return &accessLogUsecase{
		db:       db,
		repo:     repo,
		userRepo: userRepo,
	}
}

type accessLogUsecase struct {
	db       database.DB
	repo     repositories.AccessLogRepository
	userRepo repositories.UserRepository
}

func (u accessLogUsecase) SaveAccessLogByUserID(ctx context.Context, userID user.ID, accessLog domain.AccessLogParams) error {
	err := u.repo.SaveAccessLog(ctx, u.db.Conn(), userID, accessLog)
	if err != nil {
		return serrors.WithStackTrace(err)
	}
	return nil
}
