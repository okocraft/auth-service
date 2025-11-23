package usecases

import (
	"github.com/okocraft/auth-service/internal/config"
	"github.com/okocraft/auth-service/internal/repositories"
	"github.com/okocraft/auth-service/internal/repositories/database"
)

type UsecaseFactory struct {
	AuthConfig    config.AuthConfig
	DB            database.DB
	AccessLogRepo repositories.AccessLogRepository
	AuthRepo      repositories.AuthRepository
	UserRepo      repositories.UserRepository
}

func NewUsecaseFactory(conf config.AuthConfig, db database.DB) UsecaseFactory {
	return UsecaseFactory{
		AuthConfig:    conf,
		DB:            db,
		AccessLogRepo: repositories.NewAccessLogRepository(),
		AuthRepo:      repositories.NewAuthRepository(),
		UserRepo:      repositories.NewUserRepository(),
	}
}

func (f UsecaseFactory) NewAccessLogUsecase() AccessLogUsecase {
	return NewAccessLogUsecase(f.DB, f.AccessLogRepo, f.UserRepo)
}

func (f UsecaseFactory) NewAuthUsecase() AuthUsecase {
	return NewAuthUsecase(f.AuthConfig, f.DB, f.AuthRepo, f.UserRepo)
}

func (f UsecaseFactory) NewUserUsecase() UserUsecase {
	return NewUserUsecase(f.DB, f.UserRepo)
}
