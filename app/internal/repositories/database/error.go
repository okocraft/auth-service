package database

import "github.com/Siroshun09/serrors/v2"

type DBError struct {
	err error
}

func (e DBError) Error() string {
	return e.err.Error()
}

func (e DBError) Unwrap() error {
	return e.err
}

func NewDBError(err error) DBError {
	return DBError{err: err}
}

func NewDBErrorWrap(err error) error {
	return serrors.Wrap(NewDBError(err))
}
