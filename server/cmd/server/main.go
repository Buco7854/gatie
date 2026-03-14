package main

import (
	"context"
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

	"github.com/gatie-io/gatie-server/internal/database"
	"github.com/gatie-io/gatie-server/internal/handler"
	"github.com/gatie-io/gatie-server/internal/repository"
)

type Options struct {
	Host        string `doc:"Host to listen on" default:"0.0.0.0"`
	Port        int    `doc:"Port to listen on" short:"p" default:"8888"`
	DatabaseURL string `doc:"PostgreSQL connection URL" default:"postgres://gatie:gatie@localhost:5432/gatie?sslmode=disable"`
	ValkeyURL   string `doc:"Valkey connection URL" default:"valkey://localhost:6379"`
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

		router := chi.NewMux()
		api := humachi.New(router, huma.DefaultConfig("GATIE", "1.0.0"))

		handler.RegisterHealth(api, dbpool, vkClient)
		_ = queries

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
