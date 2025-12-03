package database

import "github.com/Siroshun09/serrors"

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

func NewDBErrorWithStackTrace(err error) error {
	return serrors.WithStackTrace(NewDBError(err))
}
