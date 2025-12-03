package testdb

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Siroshun09/serrors"
	"github.com/gofrs/uuid/v5"
	"github.com/okocraft/auth-service/internal/config"
	"github.com/okocraft/auth-service/internal/repositories/database"
	"github.com/stretchr/testify/require"
)

type TestDB interface {
	GetDB() database.DB
	Run(t *testing.T, f func(ctx context.Context, conn database.Connection))
	Cleanup() error
}

func NewTestDB(useTx bool) (TestDB, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, serrors.WithStackTrace(err)
	}

	dbConfig, err := config.NewDBConfigFromEnv()
	if err != nil {
		dbConfig = config.DBConfig{
			Host:     "localhost",
			Port:     "3306",
			User:     "auth_service_user",
			Password: "auth_service_pw",
		}
	}

	db, err := database.New(dbConfig, 15*time.Minute)
	if err != nil {
		return nil, serrors.WithStackTrace(err)
	}

	dbConfig.DBName = "testdb_" + strings.ReplaceAll(id.String(), "-", "")
	createDB := "CREATE " + "DATABASE " + dbConfig.DBName
	_, err = db.Base().Exec(createDB)
	if err != nil {
		return nil, serrors.WithStackTrace(err)
	}

	_, err = db.Base().Exec("USE " + dbConfig.DBName)
	if err != nil {
		return nil, serrors.WithStackTrace(err)
	}

	err = createTables(db.Base())
	if err != nil {
		return nil, serrors.WithStackTrace(err)
	}

	return &testDB{
		db:    db,
		dbCfg: dbConfig,
		useTx: useTx,
	}, nil
}

func createTables(db *sql.DB) error {
	rootDir, err := GetProjectRoot()
	if err != nil {
		return serrors.WithStackTrace(err)
	}

	schema, err := os.ReadFile(filepath.Join(rootDir, "schema/database/schema.sql"))
	if err != nil {
		return serrors.WithStackTrace(err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		return serrors.WithStackTrace(err)
	}

	return nil
}

type testDB struct {
	db    database.DB
	dbCfg config.DBConfig
	useTx bool
}

func (db *testDB) GetDB() database.DB {
	return db.db
}

func (db *testDB) Run(t *testing.T, f func(ctx context.Context, conn database.Connection)) {
	ctx := t.Context()

	if !db.useTx {
		f(ctx, db.db.Conn())
		return
	}

	err := db.db.WithTx(ctx, func(ctx context.Context, tx database.Connection) error {
		f(ctx, tx)
		return nil
	})
	require.NoError(t, err)
}

func (db *testDB) Cleanup() (err error) {
	err = db.db.Close()
	if err != nil {
		return serrors.WithStackTrace(err)
	}

	cfg := db.dbCfg
	cfg.DBName = ""
	dbForDrop, err := database.New(cfg, 15*time.Minute)
	if err != nil {
		return serrors.WithStackTrace(err)
	}
	defer func() {
		closeErr := dbForDrop.Close()
		if closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	_, err = dbForDrop.Base().Exec("DROP " + "DATABASE " + db.dbCfg.DBName)
	if err != nil {
		return serrors.WithStackTrace(err)
	}

	return nil
}
