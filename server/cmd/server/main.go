package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valkey-io/valkey-go"

	"github.com/gatie-io/gatie-server/internal/auth"
	"github.com/gatie-io/gatie-server/internal/database"
	"github.com/gatie-io/gatie-server/internal/handler"
	"github.com/gatie-io/gatie-server/internal/middleware"
	"github.com/gatie-io/gatie-server/internal/repository/postgres"
	"github.com/gatie-io/gatie-server/internal/service"
)

type Options struct {
	Host        string `doc:"Host to listen on" default:"0.0.0.0"`
	Port        int    `doc:"Port to listen on" short:"p" default:"8888"`
	DatabaseURL string `doc:"PostgreSQL connection URL" default:"postgres://gatie:gatie@localhost:5432/gatie?sslmode=disable"`
	ValkeyURL   string `doc:"Valkey connection URL" default:"valkey://localhost:6379"`
	JWTSecret   string `doc:"JWT signing secret (auto-generated if empty)" default:""`
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	cli := humacli.New(func(hooks humacli.Hooks, opts *Options) {
		dbpool, err := pgxpool.New(context.Background(), opts.DatabaseURL)
		if err != nil {
			slog.Error("unable to connect to database", "error", err)
			os.Exit(1)
		}

		database.RunMigrations(dbpool)

		vkClient, err := valkey.NewClient(valkey.MustParseURL(opts.ValkeyURL))
		if err != nil {
			slog.Error("unable to connect to Valkey", "error", err)
			os.Exit(1)
		}

		queries := postgres.New(dbpool)

		jwtSecret := opts.JWTSecret
		if jwtSecret == "" {
			b := make([]byte, 32)
			rand.Read(b)
			jwtSecret = hex.EncodeToString(b)
			slog.Warn("JWT secret auto-generated, set SERVICE_JWT_SECRET for persistent sessions across restarts")
		}

		jwtManager := auth.NewJWTManager(jwtSecret, 15*time.Minute, 7*24*time.Hour)

		router := chi.NewMux()
		config := huma.DefaultConfig("GATIE", "1.0.0")
		config.Transformers = append(config.Transformers, middleware.ErrorCaptureTransformer)
		api := humachi.New(router, config)
		api.UseMiddleware(middleware.NewRecover(api))
		api.UseMiddleware(middleware.NewRequestLogger())

		authMW := middleware.NewAuthMiddleware(api, jwtManager)
		adminMW := middleware.NewRequireAdmin(api)
		authRateLimitMW := middleware.NewRateLimit(api, 0.5, 5)

		authService := service.NewAuthService(queries, dbpool, jwtManager)
		authHandler := handler.NewAuthHandler(authService, authRateLimitMW)

		memberService := service.NewMemberService(queries, dbpool)
		memberHandler := handler.NewMemberHandler(memberService, authMW, adminMW)

		gateService := service.NewGateService(queries)
		gateHandler := handler.NewGateHandler(gateService, authMW, adminMW)

		handler.RegisterHealth(api, dbpool, vkClient)
		authHandler.Register(api)
		memberHandler.Register(api)
		gateHandler.Register(api)

		server := &http.Server{
			Addr:    fmt.Sprintf("%s:%d", opts.Host, opts.Port),
			Handler: router,
		}

		hooks.OnStart(func() {
			slog.Info("GATIE server starting", "host", opts.Host, "port", opts.Port)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("server error", "error", err)
				os.Exit(1)
			}
		})

		hooks.OnStop(func() {
			slog.Info("shutting down")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			server.Shutdown(ctx)
			dbpool.Close()
			vkClient.Close()
		})
	})

	cli.Run()
}
