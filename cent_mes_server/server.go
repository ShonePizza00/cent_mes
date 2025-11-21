package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"text/template"
	"time"
)

func createTables(db *sql.DB) {
	db.Exec(`CREATE TABLE users_auth(
	login TEXT PRIMARY KEY,
	passwd TEXT NOT NULL,
	token TEXT NOT NULL);`)
	db.Exec(`CREATE TABLE chats(
	id INTEGER PRIMARY KEY,
	type TEXT NOT NULL,
	title TEXT,
	CREATED_AT DATETIME NOT NULL);`)
	db.Exec(`CREATE TABLE chat_members(
	chat_id INTEGER NOT NULL,
	user_id TEXT NOT NULL,
	PRIMARY KEY (chat_id, user_id));`)
	db.Exec(`CREATE TABLE messages(
	id INTEGER PRIMARY KEY,
	chat_id INTEGER NOT NULL,
	sender_id TEXT NOT NULL,
	created_at DATETIME NOT NULL,
	body TEXT NOT NULL);`)
}

func addRoutes(
	srvMx *http.ServeMux,
	ri *runtimeInstance) {
	srvMx.HandleFunc("/", ri.handlerReturnFormAuth)

	srvMx.HandleFunc("GET /auth", ri.handlerReturnFormAuth)
	srvMx.HandleFunc("GET /reg", ri.handlerReturnFormReg)
	srvMx.HandleFunc("POST /auth", ri.handlerLogin)
	srvMx.HandleFunc("POST /reg", ri.handlerRegister)

	srvMx.HandleFunc("GET /mes", ri.handlerReturnFormMessages)
	srvMx.HandleFunc("GET /api/chats", ri.APIhandlerGetChats)
	srvMx.HandleFunc("GET /api/mes", ri.APIhandlerGetMessages)
	srvMx.HandleFunc("POST /api/mes", ri.APIhandlerSendMessage)

	srvMx.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js"))))
}

func (ri *runtimeInstance) handlerReturnFormAuth(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("user")
	if err != nil {
		http.ServeFile(w, r, "html/login.html")
		return
	}
	http.Redirect(w, r, "/mes", http.StatusSeeOther)
}

func (ri *runtimeInstance) handlerReturnFormReg(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "html/register.html")
}

func (ri *runtimeInstance) handlerLogin(w http.ResponseWriter, r *http.Request) {
	login := r.FormValue("login")
	paswd_form := r.FormValue("password")
	paswd_hash := sha256.Sum256([]byte(paswd_form))
	paswd := string(paswd_hash[:])
	log.Printf("Login attempt. name:\"%s\" passwd:\"%s\"", login, paswd)
	res := ri.DB.QueryRow("SELECT * FROM users_auth WHERE login=?", login)
	var user_query User
	err := res.Scan(&user_query.Login, &user_query.Password, &user_query.Token)
	if err != nil || login != user_query.Login || paswd != user_query.Password {
		log.Println("Incorrect username or password")
		http.Redirect(w, r, "/auth", http.StatusSeeOther) //=> show message to user
		return
	}
	log.Println("Login successful")
	http.SetCookie(w, &http.Cookie{
		Name:     "user",
		Value:    url.QueryEscape(user_query.Token),
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/mes", http.StatusSeeOther)
}

func (ri *runtimeInstance) handlerRegister(w http.ResponseWriter, r *http.Request) {
	login := r.FormValue("login")
	email := r.FormValue("email")
	paswd_form := r.FormValue("password")
	paswd_hash := sha256.Sum256([]byte(paswd_form))
	paswd := hex.EncodeToString(paswd_hash[:])
	res := ri.DB.QueryRow("SELECT login FROM users_auth WHERE login=?", login)
	var user_query User
	err := res.Scan(&user_query.Login)
	if err == nil || user_query.Login == login {
		fmt.Fprintln(w, "Incorrect user or password")
		log.Println("Incorrect user or password")
		return
	}
	token_t := sha256.Sum256([]byte(login))
	token := hex.EncodeToString(token_t[:])
	_, err = ri.DB.Exec(
		`INSERT INTO users_auth 
		(login, passwd, token) 
		VALUES (?, ?, ?)`,
		login, paswd, token)
	if err != nil {
		log.Println(err)
		fmt.Fprintln(w, "Fatal error")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "user",
		Value:    url.QueryEscape(token),
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	log.Printf("New user registered. name:%s, email:%s, passwd:%s\n", login, email, paswd)
	http.Redirect(w, r, "/auth", http.StatusSeeOther)
}

func (ri *runtimeInstance) handlerReturnFormMessages(w http.ResponseWriter, r *http.Request) {
	username, err_c := ri.UsernameFromCookie(w, r, "user")
	if err_c != nil {
		log.Println(err_c)
		return
	}
	tmpl := template.Must(template.ParseFiles("html/messenger.html"))
	if err := tmpl.Execute(w, username); err != nil {
		return
	}
}

func (ri *runtimeInstance) APIhandlerGetMessages(w http.ResponseWriter, r *http.Request) {
	_, err := ri.UsernameFromCookie(w, r, "user")
	if err != nil {
		log.Println(err)
		return
	}
	q_vals := r.URL.Query()
	if q_vals.Has("chat_id") {
		chat_id, _ := strconv.Atoi(q_vals.Get("chat_id"))
		afterID, _ := strconv.Atoi(q_vals.Get("after_id"))
		mess, err := ri.MessagesInChat(chat_id, afterID)
		if err != nil {
			log.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(mess)
		return
	}
}

func (ri *runtimeInstance) APIhandlerGetChats(w http.ResponseWriter, r *http.Request) {
	username, err := ri.UsernameFromCookie(w, r, "user")
	if err != nil {
		log.Println(err)
		return
	}
	chats, err := ri.UserChats(username)
	if err != nil {
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(chats)
}

func (ri *runtimeInstance) APIhandlerSendMessage(w http.ResponseWriter, r *http.Request) {
	login_cookie, err := ri.UsernameFromCookie(w, r, "user")
	if err != nil {
		log.Println(err)
		return
	}
	var req MessageCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println(err)
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	req.Sender = login_cookie
	if req.Getter != "" {
		chat_, err := ri.FindOrCreateNewChat(login_cookie, req.Getter)
		if err != nil {
			return
		}
		req.ChatID = chat_.ID
		ri.SendMessage(&req)
	} else if req.ChatID != 0 {
		ri.SendMessage(&req)
	} else {
		http.Error(w, "No recipient", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(req)
}
