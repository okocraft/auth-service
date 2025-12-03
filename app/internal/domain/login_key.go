package domain

import (
	"strconv"

	"github.com/Siroshun09/serrors"
)

type LoginKey int64

func ParseLoginKey(loginKey string) (LoginKey, error) {
	key, err := strconv.ParseInt(loginKey, 10, 64)
	if err != nil {
		return 0, serrors.WithStackTrace(err)
	}
	return LoginKey(key), nil
}

func (key LoginKey) String() string {
	return strconv.FormatInt(int64(key), 10)
}
