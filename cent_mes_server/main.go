package main

import (
	"cent_mes_server/handlers"
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	addrString string = "127.0.0.1:80"
)

func main() {

	ctx, _ := context.WithCancel(context.Background())
	db, err := sql.Open("sqlite3", "staff/users.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	ri := handlers.RuntimeInstance{DB: db}
	ri.CreateTables(ctx)
	srvMx := http.NewServeMux()
	// log.Println(res.LastInsertId())
	server := &http.Server{
		Addr:         addrString,
		Handler:      srvMx,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	handlers.AddRoutesRI(srvMx, &ri)
	log.Println("Server is listening on", addrString)
	err = server.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}
