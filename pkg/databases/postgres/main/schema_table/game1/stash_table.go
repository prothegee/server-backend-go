package db_pg_main_game1_stash

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"

	"showcase-backend-go/pkg/databases/postgres/main"
	account "showcase-backend-go/pkg/databases/postgres/main/schema_table/account"
)

// --------------------------------------------------------- //

// stash holder type
type Stash struct{}

// game1.stash type
type Stash_t struct {
	Id        uuid.UUID
	Uid       uuid.UUID
	Name      string
	NameNorm  string
	Items     *json.RawMessage
	DtCreated *time.Time
	DtUpdated *time.Time
}

// game1.stash type json
type Stash_tj struct {
	Id        uuid.UUID        `json:"id"`
	Uid       uuid.UUID        `json:"uid"`
	Name      string           `json:"name"`
	NameNorm  string           `json:"name_norm"`
	Items     *json.RawMessage `json:"items"`
	DtCreated *time.Time       `json:"dt_created"`
	DtUpdated *time.Time       `json:"dt_updated"`
}

// game1.stash type json clean representation
//
// use thif for clean request without unwanted cols (which uid in this case)
type Stash_tjc struct {
	Id        uuid.UUID        `json:"id"`
	Name      string           `json:"name"`
	NameNorm  string           `json:"name_norm"`
	Items     *json.RawMessage `json:"items"`
	DtCreated *time.Time       `json:"dt_created"`
	DtUpdated *time.Time       `json:"dt_updated"`
}

// conversion Stash_t to Stash_tj
//
// receiver: d Stash_t
func (d Stash_t) ToJSON() Stash_tj {
	return Stash_tj{
		Id:        d.Id,
		Uid:       d.Uid,
		Name:      d.Name,
		NameNorm:  d.NameNorm,
		Items:     d.Items,
		DtCreated: d.DtCreated,
		DtUpdated: d.DtUpdated,
	}
}

// items data representation from game1.stash items
//
// example item
//
//	{
//	    "item": "branch",
//		   "quantity": 1
//	}
type StashItem_t struct {
	Item     string
	Quantity uint64
}

type StashItem_tj struct {
	Item     string `json:"item"`
	Quantity uint64 `json:"quantity"`
}

func (d StashItem_tj) ToDATA() StashItem_t {
	return StashItem_t{
		Item:     d.Item,
		Quantity: d.Quantity,
	}
}

// --------------------------------------------------------- //

const (
	TABLE_STASH              = "stash"
	SCHEMA_TABLE_GAME1_STASH = "game1.stash"

	GAME1_STASH_CONSTRAINT_TO_ACCOUNT_USER_ID = "fk_game1_stash_uid"
)

const (
	Game1StashCOL_id         = "id"
	Game1StashCOL_uid        = "uid"
	Game1StashCOL_name       = "name"
	Game1StashCOL_name_norm  = "name_norm"
	Game1StashCOL_items      = "items"
	Game1StashCOL_dt_created = "dt_created"
	Game1StashCOL_dt_updated = "dt_updated"
)

type Game1StashItemOperand_e int

const (
	GAME1_STASH_ITEM_OPERAND_UNDEFINED Game1StashItemOperand_e = iota
	GAME1_STASH_ITEM_OPERAND_ADDITION
	GAME1_STASH_ITEM_OPERAND_SUBSTRACTION
)

// --------------------------------------------------------- //

func SQL_TABLE_INIT() string {
	return fmt.Sprintf(`-- LAST UPDATED: YYYY-MM-DD
create table if not exists %[1]s(
    id       	uuid        unique not null primary key default uuidv7(),
	uid			uuid		not null,
    name        text        not null,
    name_norm   text        not null,
    items       jsonb       null,
    dt_created  timestamp   null default now(),
    dt_updated  timestamp   null
);

-- alter table
do $$
begin
    if not exists (
        select 1 from information_schema.table_constraints
        where table_schema = '%[5]s'
        and table_name = '%[4]s'
        and constraint_name = '%[2]s'
    ) then
        alter table %[1]s
            add constraint %[2]s
            foreign key (%[6]s)
            references %[3]s (id)
            on delete no action
            on update cascade;
    end if;
end $$;

-- indexes
create index if not exists idx_game1_stash_uid on game1.stash(uid);
create index if not exists idx_game1_stash_name_norm on game1.stash(name_norm);

-- functions
create or replace function game1.stash_dt_updated()
    returns trigger as $$
    begin
        new.dt_updated = now();

        return new;
    end;
    $$ language plpgsql;

-- triggers
create or replace trigger game1_stash_dt_updated_trigger
    before update on %[1]s
    for each row
    execute function game1.stash_dt_updated();`,
		SCHEMA_TABLE_GAME1_STASH,
		GAME1_STASH_CONSTRAINT_TO_ACCOUNT_USER_ID,
		account.SCHEMA_TABLE_ACCOUNT_USER,
		TABLE_STASH,
		db_pg_main.SchemaGame1,
		Game1StashCOL_uid)
}

// --------------------------------------------------------- //

// initialzie game1.stash table
//
// params:
// 	db *pgx.Conn - must db_pg.MainDb
// 	ctx context.Context
//
// receiver: _ Stash
//
// return: error
func (_ Stash) InitTable(db *pgx.Conn, ctx context.Context) error {
	query := SQL_TABLE_INIT()
	_, err := db.Exec(ctx, query)
	if err != nil {
		log.Fatalf("FATAL ERROR \"%s\": %v", SCHEMA_TABLE_GAME1_STASH, err)
		return errors.Wrapf(err, "can't init table %s", SCHEMA_TABLE_GAME1_STASH)
	}

	return nil
}

// create new data in game1.stash table
//
// you're has a responsible to check if userId exists from account.user id
//
// params:
// 	db *pgx.Conn - must db_pg.MainDb
// 	ctx context.Context
// 	userId uuid.UUID
// 	stashName string
//
// receiver: _ Stash
//
// return: error
func (_ Stash) InsertNewStash(db *pgx.Conn, ctx context.Context,
	userId uuid.UUID, stashName string) error {
	query := fmt.Sprintf(`insert into %[1]s (%[2]s, %[3]s, %[4]s) values ($1, $2, $3);`,
		SCHEMA_TABLE_GAME1_STASH,
		Game1StashCOL_uid,
		Game1StashCOL_name,
		Game1StashCOL_name_norm)

	_, err := db.Exec(ctx, query, userId, stashName, strings.ToLower(stashName))
	if err != nil {
		return errors.Wrap(err, "failed to create new stash")
	}

	return nil
}

// select stash id by existing user id
//
// params:
// 	db *pgx.Conn - must db_pg.MainDb
// 	ctx context.Context
// 	uid uuid.UUID - existing user id
// 	name string - stash name
//
// receiver: _ Stash
//
// return: (uuid.UUID, error) - (actual stash id, nil)
func (_ Stash) SelectStashIdByUidAndName(db *pgx.Conn, ctx context.Context,
	uid uuid.UUID, name string) (uuid.UUID, error) {
	id := uuid.Nil
	query := fmt.Sprintf(`select %[1]s from %[2]s where %[3]s=$1 and %[4]s=$2;`,
		Game1StashCOL_id,
		SCHEMA_TABLE_GAME1_STASH,
		Game1StashCOL_uid,
		Game1StashCOL_name)

	err := db.QueryRow(ctx, query, uid, name).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return id, errors.New("not found/doesn't exists")
		}
		return uuid.Nil, err
	}

	return id, nil
}

// select stash if exists where it required uid & name
//
// params:
// 	db *pgx.Conn - must pg_db.MainDb
// 	ctx context.Context
// 	uid uuid.UUID - existing user id
// 	name string - stash name to check
//
// receiver: _ Stash
//
// return: (bool, error) - (true mean exists, err)
func (_ Stash) SelectStashExistenceByUidAndName(db *pgx.Conn, ctx context.Context,
	uid uuid.UUID, name string) (bool, error) {
	query := fmt.Sprintf(`select %[1]s from %[2]s where %[3]s=$1 and %[4]s=$2;`,
		Game1StashCOL_id,
		SCHEMA_TABLE_GAME1_STASH,
		Game1StashCOL_uid,
		Game1StashCOL_name)

	res, err := db.Exec(ctx, query, uid, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, errors.New("stash not found/doesn't exists")
		}
		return false, errors.Wrap(err, "failed to select existence stash")
	}

	return res.RowsAffected() > 0, nil
}

// select all stash by uid
//
// db *pgx.Conn - must db_pg.MainDb
//
// ctx context.Context
//
// params:
// 	uid uuid.UUID
//
// receiver: _ Stash
//
// return: ([]Stash_tjc, error)
func (_ Stash) SelectAllStashByUid(db *pgx.Conn, ctx context.Context,
	uid uuid.UUID) ([]Stash_tjc, error) {
	stashs := []Stash_tjc{}

	query := fmt.Sprintf(`select %[1]s, %[2]s, %[3]s, %[4]s, %[5]s 
		from %[6]s 
		where %[7]s=$1;`,
		Game1StashCOL_id,
		Game1StashCOL_name,
		Game1StashCOL_items,
		Game1StashCOL_dt_created,
		Game1StashCOL_dt_updated,
		SCHEMA_TABLE_GAME1_STASH,
		Game1StashCOL_uid)

	rows, err := db.Query(ctx, query, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s Stash_tjc
		err := rows.Scan(&s.Id, &s.Name, &s.Items, &s.DtCreated, &s.DtUpdated)
		if err != nil {
			return nil, err
		}
		stashs = append(stashs, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return stashs, nil
}

func (_ Stash) SelectStashByIdAndUid(db *pgx.Conn, ctx context.Context,
	id uuid.UUID, uid uuid.UUID) (Stash_tjc, error) {
	stash := Stash_tjc{}

	query := fmt.Sprintf(`select %[1]s %[2]s %[3]s %[4]s %[5]s %[6]s
		from %[7]s
		where %[8]s=$1
		and %[9]s=$2;`,
		Game1StashCOL_id,
		Game1StashCOL_name,
		Game1StashCOL_name_norm,
		Game1StashCOL_items,
		Game1StashCOL_dt_created,
		Game1StashCOL_dt_updated,
		SCHEMA_TABLE_GAME1_STASH,
		Game1StashCOL_id,
		Game1StashCOL_uid)

	rows, err := db.Query(ctx, query, id, uid)
	if err != nil {
		return Stash_tjc{}, err
	}

	err = rows.Scan(&stash.Id, &stash.Name, &stash.Items,
		&stash.DtCreated, &stash.DtUpdated)
	if err != nil {
		return Stash_tjc{}, err
	}

	return stash, nil
}

// SelectStashByUidAndName

// update stash by uid and name
//
// it will check related stash first, if it's exists the operand will do the thing
//
// conditional second string return value is mostly 0 len, if it has something, currently meant that it's ok but the algo is not meant to be implement in database such as for operand substraction where item name doesn't exists
//
// params:
// 	db *pgx.Conn - must pg_db.MainDb
// 	ctx context.Context
// 	uid uuid.UUID
// 	name string
// 	item StashItem_t
// 	operand Game1StashItemOperand_e
//
// receiver: _ Stash
//
// return: (error, string) - (nil ok, message conditional)
func (_ Stash) UpdateStashByUidAndName(db *pgx.Conn, ctx context.Context,
	uid uuid.UUID, name string,
	item StashItem_t,
	operand Game1StashItemOperand_e) (error, string) {
	if operand == GAME1_STASH_ITEM_OPERAND_UNDEFINED {
		return errors.New("operand must be addition or subtraction"), ""
	}

	// current items
	var rawItems json.RawMessage
	queryGet := fmt.Sprintf(`select %[1]s from %[2]s where %[3]s=$1 and %[4]s=$2;`,
		Game1StashCOL_items,
		SCHEMA_TABLE_GAME1_STASH,
		Game1StashCOL_uid,
		Game1StashCOL_name)

	err := db.QueryRow(ctx, queryGet, uid, name).Scan(&rawItems)
	if err != nil {
		if err == pgx.ErrNoRows {
			return errors.New("stash not found for given uid and name"), ""
		}
		return errors.Wrap(err, "failed to fetch current stash items"), ""
	}

	var itemsList []StashItem_t
	if rawItems != nil {
		if err := json.Unmarshal(rawItems, &itemsList); err != nil {
			return errors.Wrap(err, "failed to unmarshal items JSON"), ""
		}
	}

	// find and update
	found := false
	for i := range itemsList {
		if itemsList[i].Item == item.Item {
			found = true
			switch operand {
			case GAME1_STASH_ITEM_OPERAND_ADDITION:
				itemsList[i].Quantity += item.Quantity
			case GAME1_STASH_ITEM_OPERAND_SUBSTRACTION:
				if item.Quantity > itemsList[i].Quantity {
					return errors.New("insufficient quantity for subtraction"), ""
				}
				itemsList[i].Quantity -= item.Quantity
				// delete item if quanity become 0
				if itemsList[i].Quantity == 0 {
					itemsList = append(itemsList[:i], itemsList[i+1:]...)
				}
			}
			break
		}
	}

	// add new item if not found
	if !found {
		if operand == GAME1_STASH_ITEM_OPERAND_SUBSTRACTION {
			return nil, "item not found in stash for subtraction"
		}
		itemsList = append(itemsList, item)
	}

	updatedJSON, err := json.Marshal(itemsList)
	if err != nil {
		return errors.Wrap(err, "failed to marshal updated items"), ""
	}

	queryUpdate := fmt.Sprintf(`update %[1]s set %[2]s=$1 where %[3]s=$2 and %[4]s=$3;`,
		SCHEMA_TABLE_GAME1_STASH,
		Game1StashCOL_items,
		Game1StashCOL_uid,
		Game1StashCOL_name)

	_, err = db.Exec(ctx, queryUpdate, updatedJSON, uid, name)
	if err != nil {
		return errors.Wrap(err, "failed to update stash items"), ""
	}

	return nil, ""
}

// delete stash by id
//
// params:
// 	db *pgx.Conn - must db_pg.MainDb
// 	ctx context.Context
// 	id uuid.UUID
//
// receiver: _ Stash
//
// return: error
func (_ Stash) DeleteStashById(db *pgx.Conn, ctx context.Context,
	id uuid.UUID) error {
	query := fmt.Sprintf(`delete from %[1]s where %[2]s=$1;`,
		SCHEMA_TABLE_GAME1_STASH,
		Game1StashCOL_id)

	_, err := db.Exec(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete stash")
	}

	return nil
}
