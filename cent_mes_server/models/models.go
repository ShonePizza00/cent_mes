package models

import "time"

type Message struct {
	ID        int       `json:"id"`
	ChatID    int       `json:"chat_id"`
	Sender    string    `json:"sender_id"`
	CreatedAt time.Time `json:"created_at"`
	Body      string    `json:"body"`
}

type MessageCreateRequest struct {
	ID     int    `json:"id"`
	ChatID int    `json:"chat_id"`
	Sender string `json:"sender_id"`
	Getter string `json:"getter_id"`
	Body   string `json:"body"`
}

type User struct {
	Login    string
	Password string
	Token    string
}

type Chat struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type PageData struct {
	CurrentChatID int
	Chats         []Chat
	Messages      []Message
}
