package httpserver

import (
	"cent_mes_server/internal/auth"
	"cent_mes_server/internal/chats"
	"cent_mes_server/internal/messages"
	"context"
	"net/http"
)

type Services struct {
	AuthService *auth.Service
	ChatService *chats.Service
	MsgService  *messages.Service
}

func AddRoutes(ctx context.Context, mux *http.ServeMux, services *Services) {
	authHandler := auth.NewHandler(ctx, services.AuthService)
	chatsHandler := chats.NewHandler(ctx, services.ChatService)
	messagesHandler := messages.NewHandler(ctx, services.MsgService)

	auth.RegisterRoutes(mux, authHandler)
	chats.RegisterRoutes(mux, chatsHandler)
	messages.RegisterRoutes(mux, messagesHandler)
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js"))))
}
