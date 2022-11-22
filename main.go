package main

import (
	"database/sql"
	"log"

	"github.com/karlib/simple_bank/api"
	db "github.com/karlib/simple_bank/db/sqlc"
	"github.com/karlib/simple_bank/util"
	_ "github.com/lib/pq"
)

func main() {
	// "." znamená načíst z aktuální složky protože app.env config file je ve stejné složce jako main.go
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannor conect to db:", err)
	}

	store := db.NewStore(conn)

	server := api.NewServer(store)

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start")
	}

}
