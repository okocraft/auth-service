package database

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"time"

	"github.com/Siroshun09/serrors"
	"github.com/go-sql-driver/mysql"
	"github.com/okocraft/auth-service/internal/config"
)

var (
	ErrFailedToBegin    = errors.New("tx begin err")
	ErrFailedToRollback = errors.New("rollback err")
	ErrFunctionError    = errors.New("function err")
	ErrFailedToCommit   = errors.New("commit err")
)

type DB interface {
	Base() *sql.DB
	Conn() Connection
	WithTx(ctx context.Context, fn func(ctx context.Context, tx Connection) error) error
	Close() error
}

func GenerateConfig(c config.DBConfig) *mysql.Config {
	cfg := mysql.NewConfig()
	cfg.User = c.User
	cfg.Passwd = c.Password
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(c.Host, c.Port)
	cfg.DBName = c.DBName
	cfg.MultiStatements = true
	cfg.ParseTime = true
	return cfg
}

func New(c config.DBConfig, maxLifeTime time.Duration) (DB, error) {
	conn, err := sql.Open("mysql", GenerateConfig(c).FormatDSN())
	if err != nil {
		return nil, serrors.WithStackTrace(err)
	}
	conn.SetConnMaxLifetime(maxLifeTime)
	if err := conn.Ping(); err != nil {
		return nil, serrors.WithStackTrace(err)
	}
	return db{base: conn}, nil
}

type db struct {
	base   *sql.DB
	txOpts *sql.TxOptions
}

func (db db) Base() *sql.DB {
	return db.base
}

func (db db) Conn() Connection {
	return newConnection(db.base)
}

func (db db) WithTx(ctx context.Context, fn func(ctx context.Context, tx Connection) error) (returnErr error) {
	tx, beginErr := db.base.BeginTx(ctx, db.txOpts)
	if beginErr != nil {
		return serrors.WithStackTrace(errors.Join(ErrFailedToBegin, beginErr))
	}

	var fnErr error
	defer func() {
		if fnErr != nil {
			rbErr := tx.Rollback()
			if rbErr != nil {
				returnErr = serrors.WithStackTrace(errors.Join(ErrFailedToRollback, fnErr, rbErr))
				return
			}
		}
	}()

	if fnErr = fn(ctx, newConnection(tx)); fnErr != nil {
		return serrors.WithStackTrace(errors.Join(ErrFunctionError, fnErr))
	}

	if err := tx.Commit(); err != nil {
		return serrors.WithStackTrace(errors.Join(ErrFailedToCommit, err))
	}

	return nil
}

func (db db) Close() error {
	err := db.base.Close()
	if err != nil {
		return serrors.WithStackTrace(err)
	}
	return nil
}
