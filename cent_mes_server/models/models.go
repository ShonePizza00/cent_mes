package models

import "time"

type Message struct {
	ID        int64     `json:"id"`
	ChatID    int64     `json:"chat_id"`
	Sender    string    `json:"sender_id"`
	CreatedAt time.Time `json:"created_at"`
	Body      string    `json:"body"`
}

type MessageCreateRequest struct {
	ID     int64  `json:"id"`
	ChatID int64  `json:"chat_id"`
	Sender string `json:"sender_id"`
	Getter string `json:"getter_id"`
	Body   string `json:"body"`
}

type User struct {
	Login    string
	Password string
	Token    string
	Email    string
}

type Chat struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type ChatFetch struct {
	ChatID  int64 `json:"chatID"`
	AfterID int64 `json:"afterID"`
}

type PageData struct {
	CurrentChatID int64
	Chats         []Chat
	Messages      []Message
}
