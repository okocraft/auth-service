package domain

import (
	"net"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/okocraft/authlib/user"
)

type AccessLogActionType int8

const (
	AccessLogActionTypeLogin AccessLogActionType = iota
	AccessLogActionTypeLogout
	AccessLogActionTypeFirstLogin
	AccessLogActionTypeRefreshToken
)

type AccessLog struct {
	UserID    user.ID
	Action    AccessLogActionType
	LoginID   uuid.UUID
	IP        net.IP
	UserAgent string
	CreatedAt time.Time
}

type AccessLogParams struct {
	Action    AccessLogActionType
	LoginID   uuid.UUID
	IP        net.IP
	UserAgent string
	CreatedAt time.Time
}

const UserAgentMaxLength = 255

func TruncateUserAgent(userAgent string) string {
	if len(userAgent) <= UserAgentMaxLength {
		return userAgent
	}
	return userAgent[:UserAgentMaxLength]
}
