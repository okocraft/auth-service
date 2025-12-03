package repositories

import (
	"context"

	"github.com/okocraft/auth-service/internal/domain"
	"github.com/okocraft/auth-service/internal/repositories/database"
	"github.com/okocraft/auth-service/internal/repositories/queries"
	"github.com/okocraft/authlib/user"
)

type AccessLogRepository interface {
	SaveAccessLog(ctx context.Context, conn database.Connection, userID user.ID, accessLog domain.AccessLogParams) error
}

func NewAccessLogRepository() AccessLogRepository {
	return &accessLogRepository{}
}

type accessLogRepository struct{}

func (r accessLogRepository) SaveAccessLog(ctx context.Context, conn database.Connection, userID user.ID, accessLog domain.AccessLogParams) error {
	err := conn.Queries().InsertAccessLog(ctx, queries.InsertAccessLogParams{
		UserID:     int32(userID),
		ActionType: int8(accessLog.Action),
		LoginID:    accessLog.LoginID.Bytes(),
		Ip:         accessLog.IP,
		UserAgent:  accessLog.UserAgent,
		CreatedAt:  accessLog.CreatedAt,
	})
	if err != nil {
		return database.NewDBErrorWithStackTrace(err)
	}
	return nil
}
