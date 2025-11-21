package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"
)

type runtimeInstance struct {
	DB *sql.DB
}

var (
	ErrNoCookie  = errors.New("no cookie")
	ErrCookieErr = errors.New("cookie error")
	ErrNoUser    = errors.New("no user with token")
)

func (ri *runtimeInstance) UsernameFromCookie(
	w http.ResponseWriter,
	r *http.Request,
	cookie_tag string) (string, error) {
	c, err_cookie := r.Cookie(cookie_tag)
	if err_cookie != nil {
		if err_cookie == http.ErrNoCookie {
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			return "", ErrNoCookie
		}
		http.Error(w, err_cookie.Error(), http.StatusBadRequest)
		return "", ErrCookieErr
	}
	row := ri.DB.QueryRow(`SELECT login FROM users_auth WHERE token=?`, c.Value)
	var username string
	err := row.Scan(&username)
	if err == sql.ErrNoRows {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return "", ErrNoUser
	}
	return username, nil
}

func (ri *runtimeInstance) UserChats(
	username string) ([]Chat, error) {
	rows, err := ri.DB.Query(
		`SELECT 
		cm.chat_id, 
		c.title,
		(SELECT m.created_at
		FROM messages m
		WHERE m.chat_id=cm.chat_id
		ORDER BY m.created_at DESC
		LIMIT 1) as last_mes_at
		FROM chat_members cm
		JOIN chats c ON cm.chat_id=c.id
		WHERE user_id=?
		ORDER BY last_mes_at DESC`, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var chats_slice []Chat = make([]Chat, 0, 10)
	for rows.Next() {
		var chat_t Chat
		var time_t time.Time
		rows.Scan(&chat_t.ID, &chat_t.Title, &time_t)
		temp := strings.Split(chat_t.Title, "|")
		if temp[0] == username {
			chat_t.Title = temp[1]
		} else {
			chat_t.Title = temp[0]
		}
		chats_slice = append(chats_slice, chat_t)
	}
	return chats_slice, nil
}

func (ri *runtimeInstance) MessagesInChat(
	chat_id int,
	after_id int) ([]Message, error) {
	rows, err := ri.DB.Query(`SELECT * FROM messages WHERE chat_id=? AND id>?`, chat_id, after_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	mess := make([]Message, 0, 32)
	for rows.Next() {
		var m Message
		rows.Scan(&m.ID, &m.ChatID, &m.Sender, &m.CreatedAt, &m.Body)
		mess = append(mess, m)
	}
	return mess, nil
}

func (ri *runtimeInstance) CreateNewChat(
	user1 string,
	user2 string) Chat {
	var chat_res Chat
	row := ri.DB.QueryRow(`SELECT MAX(chat_id) FROM chat_members`)
	row.Scan(&chat_res.ID)
	chat_res.ID++
	chat_res.Title = user1 + "|" + user2
	ri.DB.Exec(`INSERT INTO chat_members (chat_id, user_id) VALUES (?,?)`, chat_res.ID, user1)
	ri.DB.Exec(`INSERT INTO chat_members (chat_id, user_id) VALUES (?,?)`, chat_res.ID, user2)
	ri.DB.Exec(
		`INSERT INTO chats 
			(id, type, title, created_at) 
			VALUES (?,"talk", ?, CURRENT_TIMESTAMP)`,
		chat_res.ID, chat_res.Title)
	return chat_res
	// res := ri.DB.QueryRow(
	// 	`SELECT t1.chat_id
	// 		FROM chat_members t1
	// 		JOIN chat_members t2 ON t1.chat_id=t2.chat_id
	// 		WHERE t1.user_id=? AND t2.user_id=?`, user1, user2) //must select chat where only two users exist
	// var chat_q Chat
	// err := res.Scan(&chat_q.ID)
	// if err == sql.ErrNoRows {
	// 	res := ri.DB.QueryRow(`SELECT MAX(chat_id) FROM chat_members`)
	// 	res.Scan(&chat_q.ID)
	// 	chat_q.ID++
	// 	ri.DB.Exec(`INSERT INTO chat_members (chat_id, user_id) VALUES (?,?)`, chat_q.ID, user1)
	// 	ri.DB.Exec(`INSERT INTO chat_members (chat_id, user_id) VALUES (?,?)`, chat_q.ID, user2)
	// 	ri.DB.Exec(
	// 		`INSERT INTO chats
	// 			(id, type, title, created_at)
	// 			VALUES (?,"talk", ?, CURRENT_TIMESTAMP)`,
	// 		chat_q.ID, user2+user1)
	// }
	// ri.DB.Exec(
	// 	`INSERT INTO messages
	// 			(chat_id, sender_id, created_at, body)
	// 			VALUES (?,?,CURRENT_TIMESTAMP, ?)`,
	// 	chat_q.ID, user1, user2)
}

func (ri *runtimeInstance) SendMessage(
	msg *MessageCreateRequest) error {
	_, err := ri.DB.Exec(
		`INSERT INTO messages
				(chat_id, sender_id, created_at, body)
				VALUES (?,?,CURRENT_TIMESTAMP, ?)`,
		msg.ChatID, msg.Sender, msg.Body)
	res := ri.DB.QueryRow(`SELECT MAX(id) FROM messages WHERE chat_id=?`, msg.ChatID)
	res.Scan(&msg.ID)
	return err
}

func (ri *runtimeInstance) FindOrCreateNewChat(
	user1 string,
	user2 string) (*Chat, error) {
	{
		res := ri.DB.QueryRow(`SELECT login FROM users_auth WHERE login=?`, user2)
		var t string
		if err := res.Scan(&t); err == sql.ErrNoRows {
			return nil, ErrNoUser
		}
	}
	res := ri.DB.QueryRow(
		`SELECT t1.chat_id
	FROM chat_members t1
	JOIN chat_members t2 ON t1.chat_id=t2.chat_id
	WHERE t1.user_id=? AND t2.user_id=?`, user1, user2) //=>must select chat where only two users exist
	var chat_res Chat
	err := res.Scan(&chat_res.ID)
	if err == sql.ErrNoRows {
		chat_res = ri.CreateNewChat(user1, user2)
	}
	return &chat_res, nil
}

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
