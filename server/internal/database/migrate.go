package database

import (
	"embed"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

func RunMigrations(dbpool *pgxpool.Pool) {
	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		slog.Error("migration dialect error", "error", err)
		os.Exit(1)
	}

	db := dbpool.Config().ConnConfig.ConnString()
	conn, err := goose.OpenDBWithDriver("pgx", db)
	if err != nil {
		slog.Error("migration connection error", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	if err := goose.Up(conn, "migrations"); err != nil {
		slog.Error("migration error", "error", err)
		os.Exit(1)
	}
}
