package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"showcase-backend-go/pkg"
	"showcase-backend-go/pkg/configs"
	"showcase-backend-go/pkg/databases/postgres"
	"showcase-backend-go/pkg/databases/redis"
)

const backendApi = "backend_api"

func main() {
	mux := http.NewServeMux()
	ctx := context.Background()
	cfg, err := pkg.ConfigServerLoad(config.BACKEND_API_CONFIG_JSON)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return
	}
	listAddr := fmt.Sprintf("%s:%s",
		cfg.Listener.BackendApi.Address,
		strconv.Itoa(int(cfg.Listener.BackendApi.Port)))
	log.Printf("INFO: %s run on %s\n", backendApi, listAddr)

	RegistrarDbPostgresMain()
	RegistrarDbRedisMain()

	RegistrarAssets(mux)
	RegistrarHandlers(mux)

	defer func() {
		if db_pg.MainDb != nil {
			db_pg.MainDb.Close(ctx)
			log.Println("closing SQL db_pg.DbMain")
		}

		if db_rd.MainDb != nil {
			db_rd.MainDb.Close()
			log.Println("closing NoSQL db_rd.DbMain")
		}
	}()

	log.Fatal(http.ListenAndServe(listAddr, mux))
}
