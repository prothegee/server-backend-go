package db_pg

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"showcase-backend-go/pkg"

	"github.com/jackc/pgx/v5"
)

// --------------------------------------------------------- //

// holder type for maindb postgres
type DbPgMain struct{}

// --------------------------------------------------------- //

// postgresql connection type
type PgConn_t struct {
	Host     string
	Port     int16
	User     string
	Password string
	Database string
	SslMode  string
}

// postgresql connection type json
type PgConn_tj struct {
	Host     string `json:"host"`
	Port     int16  `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SslMode  string `json:"sslmode"`
}

const (
	SslModeDisable    = "disable"
	SslModeRequire    = "require"
	SslModeVerifyCA   = "verify-ca"
	SslModeVerifyFULL = "verify-full"
)

// do not pkgify this on runtime
var sslModes = [4]string{
	SslModeDisable,
	SslModeRequire,
	SslModeVerifyCA,
	SslModeVerifyFULL,
}

func SslModes() [4]string {
	return sslModes
}

// --------------------------------------------------------- //

// this is only for postgres db main
//
// params:
//	fp string - file path
func (_ DbPgMain) InitPgDbMain(fp string) {
	var sb strings.Builder

	ctx := context.Background()
	content, err := pkg.ConfigServerLoad(fp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: db_pg fail to read file '%v'", err)
		return
	}

	pgConn := PgConn_tj{
		Host:     content.Database.PostgreSQL.Main.Host,
		Port:     int16(content.Database.PostgreSQL.Main.Port),
		User:     content.Database.PostgreSQL.Main.User,
		Password: content.Database.PostgreSQL.Main.Password,
		Database: content.Database.PostgreSQL.Main.Database,
		SslMode:  content.Database.PostgreSQL.Main.SslMode,
	}

	if len(pgConn.User) <= 0 {
		err = errors.New("user value is empty")
		log.Fatal(err)
		return
	}
	sb.WriteString("user=")
	sb.WriteString(pgConn.User)

	if len(pgConn.Password) > 0 {
		sb.WriteString(" password=")
		sb.WriteString(pgConn.Password)
	} // let it empty if password not supply

	if len(pgConn.Host) <= 0 {
		err = errors.New("host value is empty")
		log.Fatal(err)
		return
	}
	sb.WriteString(" host=")
	sb.WriteString(pgConn.Host)

	if pgConn.Port <= 3 {
		err = errors.New("port digit is less or equal than 3")
		// unless it's expected less than 3, my question is why?
		log.Fatal(err)
		return
	}
	sb.WriteString(" port=")
	sb.WriteString(fmt.Sprintf("%d", pgConn.Port))

	// skipped: database

	if len(pgConn.SslMode) <= 0 {
		err = errors.New("sslmode value can't be empty")
		log.Fatal(err)
		return
	}
	correctSslMode := false
	for _, val := range SslModes() {
		if val == pgConn.SslMode {
			correctSslMode = true
			sb.WriteString(" sslmode=")
			sb.WriteString(pgConn.SslMode)
			break
		}
	}
	if !correctSslMode {
		errMsg := fmt.Sprintf("sslmode is wrong, use: %s, %s, %s, or %s",
			SslModeDisable, SslModeRequire, SslModeVerifyCA, SslModeVerifyFULL)
		err = errors.New(errMsg)
	} // in-correct is not correct, wrong is antonym for correct

	connStr := sb.String()

	db, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("ERROR: fail establish connection to create database, connection string \"%s\"\n", connStr)
		return
	}
	defer db.Close(ctx)

	sqlCmd := fmt.Sprintf("create database %s;", pgConn.Database)
	_, err = db.Exec(ctx, sqlCmd)
	if err != nil {
		// allowing error treat as info
		if strings.Contains(err.Error(), "42P04") {
			log.Printf("INFO: \"%s\" already created; skipping creation\n", pgConn.Database)
		} else {
			log.Fatalf("ERROR: \"%s\" unexpected: %v\n", pgConn.Database, err.Error())
		}
	}
}

// --------------------------------------------------------- //

// make connection from config server file, first string result will be looks like
// "user=postgres password=mypassword host=127.0.0.1"
//
// any empty value from config may be ignored/required
//
// params:
// 	fp string - file path
// 	pgConn *PgConn_tj
//
// return: (string, error)
func MakeConnFromConfigServerFile(fp string, pgConn *PgConn_tj) (string, error) {
	var sb strings.Builder

	content, err := pkg.ConfigServerLoad(fp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: db_pg fail to read file '%v'", err)
		return "", err
	}

	pgConn = &PgConn_tj{
		Host:     content.Database.PostgreSQL.Main.Host,
		Port:     int16(content.Database.PostgreSQL.Main.Port),
		User:     content.Database.PostgreSQL.Main.User,
		Password: content.Database.PostgreSQL.Main.Password,
		Database: content.Database.PostgreSQL.Main.Database,
		SslMode:  content.Database.PostgreSQL.Main.SslMode,
	}

	if len(pgConn.User) <= 0 {
		return "", errors.New("user value is empty")
	}
	sb.WriteString("user=")
	sb.WriteString(pgConn.User)

	if len(pgConn.Password) > 0 {
		sb.WriteString(" password=")
		sb.WriteString(pgConn.Password)
	} // let it empty if password not supply

	if len(pgConn.Host) <= 0 {
		return "", errors.New("host value is empty")
	}
	sb.WriteString(" host=")
	sb.WriteString(pgConn.Host)

	if pgConn.Port <= 3 {
		return "", errors.New("port digit is less or equal than 3")
		// unless it's expected less than 3, my question is why?
	}
	sb.WriteString(" port=")
	sb.WriteString(fmt.Sprintf("%d", pgConn.Port))

	if len(pgConn.Database) <= 0 {
		return "", errors.New("database value can't be empty")
	}
	sb.WriteString(" dbname=")
	sb.WriteString(pgConn.Database)

	if len(pgConn.SslMode) <= 0 {
		return "", errors.New("sslmode value can't be empty")
	}
	correctSslMode := false
	for _, val := range SslModes() {
		if val == pgConn.SslMode {
			correctSslMode = true
			sb.WriteString(" sslmode=")
			sb.WriteString(pgConn.SslMode)
			break
		}
	}
	if !correctSslMode {
		errMsg := fmt.Sprintf("sslmode is wrong, use: %s, %s, %s, or %s",
			SslModeDisable, SslModeRequire, SslModeVerifyCA, SslModeVerifyFULL)
		return "", errors.New(errMsg)
	} // in-correct is not correct, wrong is antonym for correct

	return sb.String(), nil
}

// get instance of postgresql db
//
// closing db connection should be in main function who responsible to make the connection
//
// params:
// 	fp string - filepath
// 	cfg *PgConn_tj - postgresql connection type json
//
// return: (*sql.DB, error)
func PgDb(fp string, cfg *PgConn_tj) (*pgx.Conn, error) {
	ctx := context.Background()
	conn, err := MakeConnFromConfigServerFile(fp, cfg)
	if err != nil {
		return nil, err
	}

	base, err := pgx.Connect(ctx, conn)
	if err != nil {
		return nil, err
	}

	return base, nil
}

// --------------------------------------------------------- //

var (
	// runtime "db postgresql main" section connection pool
	// do defer close before exit server
	// this is not epoll/kqueu
	// the connection should be reuseable and no need to close in runtime
	MainDb *pgx.Conn = nil
)
