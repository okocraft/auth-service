package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/Siroshun09/logs"
	"github.com/okocraft/auth-service/internal/config"
	"github.com/okocraft/auth-service/internal/handler/http/server"
	"github.com/okocraft/auth-service/internal/repositories/database"
)

func main() {
	slogLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger := logs.NewLoggerWithSlog(slogLogger)

	ctx := context.Background()
	ctx = logs.WithContext(ctx, logger)

	cfg, err := config.NewHTTPServerConfigFromEnv()
	if err != nil {
		logger.Error(ctx, err)
		os.Exit(1)
	}

	db, err := database.New(cfg.DBConfig, 10*time.Minute)
	if err != nil {
		logger.Error(ctx, err)
		os.Exit(1)
	}
	defer func() {
		closeErr := db.Close()
		if closeErr != nil {
			logger.Error(ctx, closeErr)
		}
	}()

	httpServer := server.NewHTTPServerFactory(cfg, slogLogger, db).NewHTTPServer()

	srvCtx, stop := httpServer.Run(ctx)
	logger.Info(ctx, "http server started")
	defer stop()

	<-srvCtx.Done()
	if err := httpServer.Shutdown(1 * time.Minute); err != nil {
		logger.Error(ctx, err)
		os.Exit(1)
	}

	logger.Info(ctx, "http server has been stopped")
	os.Exit(0)
}
