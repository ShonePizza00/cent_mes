package main

import (
	"cent_mes_server/internal/auth"
	"cent_mes_server/internal/chats"
	"cent_mes_server/internal/httpserver"
	"cent_mes_server/internal/messages"
	sqlite_repo "cent_mes_server/internal/storage"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, err := sql.Open("sqlite3", "staff/users.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	srvMx := http.NewServeMux()
	server := &http.Server{
		Addr:         addrString,
		Handler:      srvMx,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	serviceRepo := sqlite_repo.NewSQLiteRepo(db)
	authSvc := auth.NewService(serviceRepo)
	chatsSvc := chats.NewService(serviceRepo)
	messagesSvc := messages.NewService(serviceRepo)
	httpserver.AddRoutes(ctx, srvMx, &httpserver.Services{
		AuthService: authSvc,
		ChatService: chatsSvc,
		MsgService:  messagesSvc,
	})
	log.Println("Server is listening on", addrString)
	err = server.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}
