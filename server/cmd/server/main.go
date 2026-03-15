package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
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
	"github.com/gatie-io/gatie-server/internal/repository"
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
	cli := humacli.New(func(hooks humacli.Hooks, opts *Options) {
		dbpool, err := pgxpool.New(context.Background(), opts.DatabaseURL)
		if err != nil {
			log.Fatalf("Unable to connect to database: %v", err)
		}

		database.RunMigrations(dbpool)

		vkClient, err := valkey.NewClient(valkey.MustParseURL(opts.ValkeyURL))
		if err != nil {
			log.Fatalf("Unable to connect to Valkey: %v", err)
		}

		queries := repository.New(dbpool)

		jwtSecret := opts.JWTSecret
		if jwtSecret == "" {
			b := make([]byte, 32)
			rand.Read(b)
			jwtSecret = hex.EncodeToString(b)
			log.Println("WARNING: JWT secret auto-generated. Set SERVICE_JWT_SECRET for persistent sessions across restarts.")
		}

		jwtManager := auth.NewJWTManager(jwtSecret, 15*time.Minute, 7*24*time.Hour)
		authMW := middleware.NewAuthMiddleware(jwtManager)

		authService := service.NewAuthService(queries, jwtManager)
		authHandler := handler.NewAuthHandler(authService)

		memberService := service.NewMemberService(queries)
		memberHandler := handler.NewMemberHandler(memberService, authMW, middleware.RequireAdmin)

		router := chi.NewMux()
		api := humachi.New(router, huma.DefaultConfig("GATIE", "1.0.0"))

		handler.RegisterHealth(api, dbpool, vkClient)
		authHandler.Register(api)
		memberHandler.Register(api)

		server := &http.Server{
			Addr:    fmt.Sprintf("%s:%d", opts.Host, opts.Port),
			Handler: router,
		}

		hooks.OnStart(func() {
			log.Printf("GATIE server starting on %s:%d", opts.Host, opts.Port)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server error: %v", err)
			}
		})

		hooks.OnStop(func() {
			log.Println("Shutting down...")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			server.Shutdown(ctx)
			dbpool.Close()
			vkClient.Close()
		})
	})

	cli.Run()
}
