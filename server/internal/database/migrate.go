package database

import (
	"embed"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

func RunMigrations(dbpool *pgxpool.Pool) {
	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Migration dialect error: %v", err)
	}

	db := dbpool.Config().ConnConfig.ConnString()
	conn, err := goose.OpenDBWithDriver("pgx", db)
	if err != nil {
		log.Fatalf("Migration connection error: %v", err)
	}
	defer conn.Close()

	if err := goose.Up(conn, "migrations"); err != nil {
		log.Fatalf("Migration error: %v", err)
	}
}
