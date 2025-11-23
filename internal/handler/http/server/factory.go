package server

import (
	"context"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"time"

	"github.com/Siroshun09/go-httplib"
	"github.com/Siroshun09/go-httplib/httplog"
	"github.com/Siroshun09/go-httplib/runner"
	"github.com/Siroshun09/logs"
	"github.com/Siroshun09/serrors"
	"github.com/Siroshun09/serrors/errorlogs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/okocraft/auth-service/internal/config"
	"github.com/okocraft/auth-service/internal/handler/http/oapi"
	"github.com/okocraft/auth-service/internal/repositories/database"
	"github.com/okocraft/auth-service/internal/usecases"
)

type HTTPServerFactory struct {
	cfg      config.HTTPServerConfig
	logger   logs.Logger
	database database.DB
}

func NewHTTPServerFactory(cfg config.HTTPServerConfig, logger *slog.Logger, database database.DB) HTTPServerFactory {
	return HTTPServerFactory{
		cfg: cfg,
		logger: errorlogs.NewLoggerWithOption(
			logs.NewLoggerWithSlog(slog.New(httplog.NewHTTPAttrHandler(logger.Handler()))),
			errorlogs.LoggerOption{
				StackTraceLogLevel:                  errorlogs.StackTraceLogLevelWarn,
				PrintStackTraceOnWarn:               cfg.Debug,
				PrintCurrentStackTraceIfNotAttached: true,
			},
		),
		database: database,
	}
}

func (f HTTPServerFactory) NewHTTPServer() runner.HTTPServerRunner {
	r := chi.NewRouter()

	r.Use(f.newRecoverer)
	r.Use(f.newLoggerMiddleware)
	r.Use(f.newRecovererForAPIHandler)
	r.Use(cors.Handler(cors.Options{
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			if _, ok := f.cfg.AllowedOrigins[origin]; ok {
				return true
			}

			if f.cfg.Debug {
				logs.Warnf(r.Context(), "Unknown origin: "+origin)
			}
			return false
		},
		AllowedOrigins:   slices.Collect(maps.Keys(f.cfg.AllowedOrigins)),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	return runner.NewHTTPServerRunner(
		&http.Server{
			Addr:    ":" + f.cfg.Port,
			Handler: oapi.HandlerFromMux(f.newAPIHandler(), r),
		},
		func(ctx context.Context, err error) {
			logs.Error(ctx, err)
		},
		func(ctx context.Context, rvr any) {
			logs.Error(ctx, serrors.Errorf("%v", rvr))
		},
	)
}

func (f HTTPServerFactory) newRecoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				logs.Error(r.Context(), serrors.Errorf("%v", rvr))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (f HTTPServerFactory) newRecovererForAPIHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				httplib.RenderInternalServerError(r.Context(), w, serrors.Errorf("%v", rvr))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (f HTTPServerFactory) newLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = logs.WithContext(ctx, f.logger)

		requestLog := httplib.NewRequestLog(r, time.Now())
		ctx = httplib.WithRequestLog(ctx, requestLog)

		responseLog := httplib.ResponseLog{}
		ctx = httplib.WithResponseLogPtr(ctx, &responseLog)

		next.ServeHTTP(w, r.WithContext(ctx))

		latency := time.Now().Sub(requestLog.Timestamp)
		ctx = httplib.WithLatency(ctx, latency)

		switch {
		case responseLog.Error == nil:
			logs.Info(ctx, "http access handled")
		case responseLog.StatusCode < http.StatusInternalServerError:
			logs.Warn(ctx, responseLog.Error)
		default:
			logs.Error(ctx, responseLog.Error)
		}
	})
}

func (f HTTPServerFactory) newAPIHandler() oapi.ServerInterface {
	usecaseFactory := usecases.NewUsecaseFactory(f.cfg.AuthConfig, f.database)
	accessLogUsecase := usecaseFactory.NewAccessLogUsecase()
	authUsecase := usecaseFactory.NewAuthUsecase()
	userUsecase := usecaseFactory.NewUserUsecase()
	return &apiHandler{
		authHandler:       newAuthHandler(authUsecase, accessLogUsecase),
		googleAuthHandler: newGoogleAuthHandler(f.cfg.GoogleAuthConfig, accessLogUsecase, authUsecase, userUsecase),
	}
}

type apiHandler struct {
	authHandler
	googleAuthHandler
}
