package domain

import (
	"errors"
)

type unauthorizedError struct {
	err error
}

func (e *unauthorizedError) Error() string {
	return e.err.Error()
}

func (e *unauthorizedError) Unwrap() error {
	return e.err
}

func NewUnauthorizedError(err error) error {
	return &unauthorizedError{err: err}
}

func IsUnauthorizedError(err error) bool {
	var unauthorized *unauthorizedError
	return errors.As(err, &unauthorized)
}

var (
	RefreshTokenIDByJTINotFoundError = errors.New("refresh token id by jti not found")
	SubAlreadyLinkedError            = errors.New("sub already linked")
	UserNotFoundBySubError           = errors.New("user not found by sub")
	UserNotFoundByLoginKeyError      = errors.New("user not found by login key")
)
