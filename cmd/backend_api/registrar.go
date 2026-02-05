package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"showcase-backend-go/cmd/backend_api/api"
	backend_api_account "showcase-backend-go/cmd/backend_api/api/account"
	backend_api_auth "showcase-backend-go/cmd/backend_api/api/auth"
	backend_api_game1 "showcase-backend-go/cmd/backend_api/api/game1"
	backend_path "showcase-backend-go/cmd/backend_api/path"
	backend_ws_stock "showcase-backend-go/cmd/backend_api/ws/stock"

	"showcase-backend-go/pkg"
	"showcase-backend-go/pkg/configs"
	"showcase-backend-go/pkg/middleware"

	"showcase-backend-go/pkg/databases/postgres"
	"showcase-backend-go/pkg/databases/postgres/main"
	account "showcase-backend-go/pkg/databases/postgres/main/schema_table/account"
	game1 "showcase-backend-go/pkg/databases/postgres/main/schema_table/game1"

	"showcase-backend-go/pkg/databases/redis"
)

// --------------------------------------------------------- //

func handlerMiddlewares(h http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}

// --------------------------------------------------------- //

// @brief registrar for postgresql main db
func RegistrarDbPostgresMain() {
	var (
		err  error
		cfg  string = config.BACKEND_API_CONFIG_JSON
		conn db_pg.PgConn_tj
	)

	thisDb := db_pg.DbPgMain{}
	thisDb.InitPgDbMain(cfg)

	ctx := context.Background()

	db_pg.MainDb, err = db_pg.PgDb(cfg, &conn)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	// schemas initializee
	{
		err = db_pg_main.InitSchemas(db_pg.MainDb)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	// account schema
	{
		account_user := account.User{}
		err = account_user.InitTable(db_pg.MainDb, ctx)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	// game1 schema
	{
		game1_stash := game1.Stash{}
		err = game1_stash.InitTable(db_pg.MainDb, ctx)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

// --------------------------------------------------------- //

func RegistrarDbRedisMain() {
	var (
		err  error
		cfg  string = config.BACKEND_API_CONFIG_JSON
		conn db_rd.RdConn_tj
	)

	db_rd.MainDb, err = db_rd.RdDb(cfg, &conn)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
}

// --------------------------------------------------------- //

// @brief registrar for assets dir
//
// @param mux *http.ServeMux
func RegistrarAssets(mux *http.ServeMux) {
	assetsDir := config.BACKEND_API_ASSETS_DIR
	publicDir := config.BACKEND_API_PUBLIC_DIR

	err := pkg.CopyDir(assetsDir, publicDir, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return
	}

	fs := http.FileServer(http.Dir(filepath.Join(config.BACKEND_API_PUBLIC_DIR)))

	mux.Handle("/", http.StripPrefix("/", fs))
}

// @brief registrar entry all endpoint handler
//
// @param mux *http.ServeMux
func RegistrarHandlers(mux *http.ServeMux) {
	// /api/status
	handlerBackendApiStatus := handlerMiddlewares(
		backend_api.BackendApiStatus,
		pkg_middleware.CheckHttpHost)
	mux.HandleFunc(backend_api.BackendApiStatusHint, handlerBackendApiStatus)

	// /api/account/user
	handlerBackendApiAccountUser := handlerMiddlewares(
		backend_api_account.BackendApiAccountUser,
		pkg_middleware.CheckHttpOrigin)
	mux.HandleFunc(backend_api_account.BackendApiAccountUserHint, handlerBackendApiAccountUser)

	// /api/auth/session
	handlerBackendApiAuthSession := handlerMiddlewares(
		backend_api_auth.BackendApiAuthSession,
		pkg_middleware.CheckHttpOrigin,
		pkg_middleware.CheckHeaderAuthorization)
	mux.HandleFunc(backend_api_auth.BackendApiAuthSessionHint, handlerBackendApiAuthSession)

	// /api/game1/stash
	handlerBackendApiGame1Stash := handlerMiddlewares(
		backend_api_game1.BackendApiGame1Stash,
		pkg_middleware.CheckHttpOrigin,
		pkg_middleware.CheckHeaderAuthorization)
	mux.HandleFunc(backend_api_game1.BackendApiGame1StashHint, handlerBackendApiGame1Stash)

	// --------------------------------------------------------- //

	// /ws/stock/trade
	handleBackendWsStockTrade := handlerMiddlewares(
		backend_ws_stock.BackendWsStockTrade,
		pkg_middleware.CheckHttpHost)
	mux.HandleFunc(backend_ws_stock.BackendWsStockTradeHint, handleBackendWsStockTrade)

	// --------------------------------------------------------- //

	// /path/{fist}/{second}
	handleBackendPathDynamic := handlerMiddlewares(
		backend_path.BackendPathDynamic,
		pkg_middleware.CheckHttpHost)
	mux.HandleFunc(backend_path.BackendPathDynamicFirstPathHint, handleBackendPathDynamic)
}
