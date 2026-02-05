package test_unittest

import (
	"context"
	"testing"

	"showcase-backend-go/pkg/configs"

	"showcase-backend-go/pkg/databases/postgres"

	"github.com/jackc/pgx/v5"
)

// --------------------------------------------------------- //

var testDbContext context.Context = context.Background()

// --------------------------------------------------------- //

func Test_SqlCommand(t *testing.T) {
	var (
		pgConn db_pg.PgConn_tj
	)

	conn, err := db_pg.MakeConnFromConfigServerFile(
		config.BACKEND_API_CONFIG_JSON, &pgConn)
	if err != nil {
		t.Errorf("ERROR: %v\n", err)
	}

	db, err := pgx.Connect(testDbContext, conn)
	if err != nil {
		t.Errorf("ERROR: %v\n", err)
	}

	defer db.Close(context.Background())

	sql := `DO language plpgsql $$ BEGIN RAISE NOTICE 'test notice #1'; END $$`
	if _, err := db.Exec(testDbContext, sql); err != nil {
		t.Errorf("ERROR: %v", err)
	}
}

// test for db_pg direct initialization
func Test_SqlCommanDbPg(t *testing.T) {
	var (
		pgConn db_pg.PgConn_tj
	)

	db, err := db_pg.PgDb(config.BACKEND_API_CONFIG_JSON, &pgConn)
	if err != nil {
		t.Errorf("ERROR: %v", err)
	}
	defer db.Close(context.Background())

	sql := `DO language plpgsql $$ BEGIN RAISE NOTICE 'test notice #2'; END $$`

	if _, err := db.Exec(testDbContext, sql); err != nil {
		t.Errorf("ERROR: %v", err)
	}
}
