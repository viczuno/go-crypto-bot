package main

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"log"
	"os"
)

func main() {
	url := os.Getenv("DATABASE_DIRECT_URL")
	if url == "" {
		log.Fatal("DATABASE_DIRECT_URL is required")
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		log.Fatal("open database failed")
	}
	defer db.Close()
	if err = goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}
	if err = goose.UpContext(context.Background(), db, "migrations"); err != nil {
		log.Fatal(err)
	}
}
