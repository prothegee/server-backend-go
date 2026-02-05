package db_pg_main_account_user

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"showcase-backend-go/pkg"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

// --------------------------------------------------------- //

// account schema of user holder type
type User struct{}

// account.user type
type User_t struct {
	Id           uuid.UUID
	Email        string
	PasswordHash string
	Dt_Created   *time.Time
	Dt_Updated   *time.Time
}

// account.user type json
type User_tj struct {
	Id           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"password_hash"`
	Dt_Created   *time.Time `json:"dt_created"`
	Dt_Updated   *time.Time `json:"dt_updated"`
}

// conversion User_t to User_tj
//
// receiver: d User_t
//
// return: User_tj
func (d User_t) ToJSON() User_tj {
	return User_tj{
		Id:           d.Id,
		Email:        d.Email,
		PasswordHash: d.PasswordHash,
		Dt_Created:   d.Dt_Created,
		Dt_Updated:   d.Dt_Updated,
	}
}

// --------------------------------------------------------- //

const (
	TABLE_USER                = "user"
	SCHEMA_TABLE_ACCOUNT_USER = "account.user"
)

const (
	AccountUserCOL_id            = "id"
	AccountUserCOL_email         = "email"
	AccountUserCOL_password_hash = "password_hash"
	AccountUserCOL_dt_created    = "dt_created"
	AccountUserCOL_dt_updated    = "dt_updated"
)

// --------------------------------------------------------- //

func SQL_TABLE_INIT() string {
	return fmt.Sprintf(`-- LAST UPDATED: YYYY-MM-DD
create table if not exists %[1]s(
    id          	uuid        unique not null primary key default uuidv7(),
    email       	text        unique not null,
    password_hash	text        not null,
    dt_created  	timestamp   null default now(),
    dt_updated  	timestamp   null
);

-- alter table

-- indexes
create index if not exists idx_account_user_email on account.user(email);

-- functions
create or replace function account.user_dt_updated()
    returns trigger as $$
    begin
        new.dt_updated = now();

        return new;
    end;
    $$ language plpgsql;

-- triggers
create or replace trigger account_user_dt_updated_trigger
    before update on account.user
    for each row
    execute function account.user_dt_updated();`,
		SCHEMA_TABLE_ACCOUNT_USER)
}

// --------------------------------------------------------- //

// @brief initialzie account.user table
//
// @param db *pgx.Conn - must db_pg.MainDb
//
// @receiver _ User
//
// @return error
func (_ User) InitTable(db *pgx.Conn, ctx context.Context) error {
	query := SQL_TABLE_INIT()
	_, err := db.Exec(ctx, query)
	if err != nil {
		log.Fatalf("FATAL ERROR \"%s\": %v", SCHEMA_TABLE_ACCOUNT_USER, err)
		return errors.Wrapf(err, "can't init table %s", SCHEMA_TABLE_ACCOUNT_USER)
	}

	return nil
}

// create new data in account.user table
//
// params:
// 	db *pgx.Conn - must db_pg.MainDb
// 	ctx context.Context
// 	email string - email to register
//
// receiver: _ User
//
// return: error
func (_ User) InsertNewUserByEmail(db *pgx.Conn, ctx context.Context,
	email string, password string) error {
	var hash string

	query := fmt.Sprintf(`insert into %[1]s (%[2]s, %[3]s) values ($1, $2);`,
		SCHEMA_TABLE_ACCOUNT_USER,
		AccountUserCOL_email,
		AccountUserCOL_password_hash)

	salt, err := pkg.GenerateSalt(pkg.ARGON2_MIN_SALT)
	if err != nil {
		return errors.Wrap(err, "failed to generate salt")
	}

	hash, err = pkg.Argon2id(password, salt, pkg.Argon2idParams_default)
	if err != nil {
		return errors.Wrap(err, "failed to hash pasword argon2id")
	}

	_, err = db.Exec(ctx, query, email, string(hash))
	if err != nil {
		return errors.Wrap(err, "fail to create new user")
	}

	return nil
}

// select id by email from account.user table
//
// params:
// 	db *pgx.Conn - must db_pg.MainDb
// 	ctx context.Context
// 	email string - email to check
//
// receiver: _ User
//
// return: (uuid.UUID, error)
func (_ User) SelectIdByEmail(db *pgx.Conn, ctx context.Context,
	email string) (uuid.UUID, error) {
	id := uuid.Nil

	query := fmt.Sprintf(`select %[1]s from %[2]s where %[3]s = $1;`,
		AccountUserCOL_id,
		SCHEMA_TABLE_ACCOUNT_USER,
		AccountUserCOL_email)

	err := db.QueryRow(ctx, query, email).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return id, errors.New("email not found/doesn't exists")
		}
		return id, errors.Wrap(err, "failed to select id by email")
	}

	return id, nil
}

// select to check if email exists
//
// params:
// 	db *pgx.Conn - must db_pg.MainDb
// 	ctx context.Context
// 	id uuid.UUID
//
// return: (bool, error) - true if exists
func (_ User) SelectIdIfExists(db *pgx.Conn, ctx context.Context,
	id uuid.UUID) (bool, error) {
	query := fmt.Sprintf(`select %[1]s from %[2]s where %[3]s=$1;`,
		AccountUserCOL_id,
		SCHEMA_TABLE_ACCOUNT_USER,
		AccountUserCOL_id)

	res, err := db.Exec(ctx, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, errors.New("id not found/doesn't exsts")
		}
		return false, errors.Wrap(err, "failed to select id")
	}

	return res.RowsAffected() > 0, nil
}

// select to check if email exists
//
// params:
// 	db *pgx.Conn - must db_pg.MainDb
// 	ctx context.Context
// 	email string - email to assign
//
// return: (bool, error) - true if exists
func (_ User) SelectEmailIfExists(db *pgx.Conn, ctx context.Context,
	email string) (bool, error) {
	query := fmt.Sprintf(`select %[1]s from %[2]s where %[3]s = $1;`,
		AccountUserCOL_id,
		SCHEMA_TABLE_ACCOUNT_USER,
		AccountUserCOL_email)

	res, err := db.Exec(ctx, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, errors.New("email not found/doesn't exists")
		}
		return false, errors.Wrap(err, "failed to select email")
	}

	return res.RowsAffected() > 0, nil
}

// select id by email from account.user table
//
// params:
// 	db *pgx.Conn - must db_pg.MainDb
// 	ctx context.Context
// 	id uuid.UUID - primary key
// 	email string - email to assign
//
// receiver: _ User
//
// return: error
func (_ User) UpdateEmailById(db *pgx.Conn, ctx context.Context,
	id uuid.UUID, email string) error {
	var (
		err error
	)

	query := fmt.Sprintf(`update %[1]s set %[2]s=$1 where id=$2;`,
		SCHEMA_TABLE_ACCOUNT_USER,
		AccountUserCOL_email)

	_, err = db.Exec(ctx, query, email, id)
	if err != nil {
		return errors.Wrap(err, "failed to update email by id")
	}

	return nil
}

// delete data from account.user table
//
// params:
// 	db *pgx.Conn - must db_pg.MainDb
// 	ctx context.Context
// 	id uuid.UUID - existing id
// 	email string - existing email
//
// receiver: _ User
//
// return: error
func (_ User) DeleteDataByIdAndEmail(db *pgx.Conn, ctx context.Context,
	id uuid.UUID, email string) error {
	var (
		err error
	)

	query := fmt.Sprintf(`delete from %[1]s where %[2]s=$1 and %[3]s=$2;`,
		SCHEMA_TABLE_ACCOUNT_USER,
		AccountUserCOL_id,
		AccountUserCOL_email)

	_, err = db.Exec(ctx, query, id, email)
	if err != nil {
		return errors.Wrap(err, "failed to delete user")
	}

	return nil
}
