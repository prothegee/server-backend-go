package db_pg_main

import (
	"context"
	"fmt"
	"log"

	"showcase-backend-go/pkg/configs"
	dbpg "showcase-backend-go/pkg/databases/postgres"

	"github.com/jackc/pgx/v5"
)

const (
	SchemaAccount = "account"
	SchemaGame1   = "game1"
)

func Schemas() []string {
	return []string{
		SchemaAccount,
		SchemaGame1,
	}
}

// initialize all schema for postgresql main
func InitSchemas(db *pgx.Conn) error {
	var pgConn dbpg.PgConn_tj

	db, err := dbpg.PgDb(config.BACKEND_API_CONFIG_JSON, &pgConn)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
		return err
	}

	for _, val := range Schemas() {
		sql := fmt.Sprintf("create schema if not exists %s;", val)

		if _, err := db.Exec(context.Background(), sql); err != nil {
			log.Fatalf("FATAL ERROR SQL: %v", err)
			return err
		}
	}

	return nil
}
