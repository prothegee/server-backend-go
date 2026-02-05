package db_rd

import (
	"fmt"
	"os"
	"showcase-backend-go/pkg"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// --------------------------------------------------------- //

// redis connection type
type RdConn_t struct {
	Host     string
	Port     int16
	User     string
	Password string
	Db       int32
}

// @brief redis connection type json
type RdConn_tj struct {
	Host     string `json:"host"`
	Port     int16  `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Db       int32  `json:"db"`
}

// --------------------------------------------------------- //

// make connection from config server file
//
// params:
// 	fp string - file path
// 	pgConn *PgConn_tj
//
// return: (string, error)
func MakeConnFromConfigServerFile(fp string, rdConn *RdConn_tj) (*redis.Options, error) {
	content, err := pkg.ConfigServerLoad(fp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: db_rd fail to read file '%v'", err)
		return nil, err
	}

	rdConn = &RdConn_tj{
		Host:     content.Database.Redis.Main.Host,
		Port:     int16(content.Database.Redis.Main.Port),
		User:     content.Database.Redis.Main.User,
		Password: content.Database.Redis.Main.Password,
		Db:       content.Database.Redis.Main.Db,
	}

	addr := fmt.Sprintf("%s:%s", rdConn.Host, strconv.Itoa(int(rdConn.Port)))

	base := &redis.Options{
		Addr:     addr,
		Username: rdConn.User,
		Password: rdConn.Password,
		DB:       int(rdConn.Db),
	}

	return base, nil
}

// get instance of redis db
//
// closing db connection should be in main function who responsible to make the connection
//
// params:
// 	fp string - filepath
// 	cfg *RdConn_tj - redis connection type json
//
// @return (*redis.Client, error)
func RdDb(fp string, cfg *RdConn_tj) (*redis.Client, error) {
	base, err := MakeConnFromConfigServerFile(fp, cfg)
	if err != nil {
		return nil, err
	}

	return redis.NewClient(base), nil
}

// --------------------------------------------------------- //

var (
	// runtime "db redis main" section connection poll
	// do defer close before exit
	// the connection should be reuseable and no need to close in runtime
	MainDb *redis.Client = nil
)
