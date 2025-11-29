package handlers

import (
	"cent_mes_server/models"
	"crypto/sha256"
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

func AddRoutesRI(
	srvMx *http.ServeMux,
	ri *RuntimeInstance) {
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

func (ri *RuntimeInstance) handlerReturnFormAuth(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("user")
	if err != nil {
		http.ServeFile(w, r, "html/login.html")
		return
	}
	http.Redirect(w, r, "/mes", http.StatusSeeOther)
}

func (ri *RuntimeInstance) handlerReturnFormReg(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "html/register.html")
}

func (ri *RuntimeInstance) handlerLogin(w http.ResponseWriter, r *http.Request) {
	login := r.FormValue("login")
	paswd_form := r.FormValue("password")
	paswd_hash := sha256.Sum256([]byte(paswd_form))
	paswd := hex.EncodeToString(paswd_hash[:])
	log.Printf("Login attempt. name:\"%s\" passwd:\"%s\"", login, paswd)
	user_query, err := ri.GetUserByLogin(login)
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

func (ri *RuntimeInstance) handlerRegister(w http.ResponseWriter, r *http.Request) {
	login := r.FormValue("login")
	email := r.FormValue("email")
	paswd_form := r.FormValue("password")
	paswd_hash := sha256.Sum256([]byte(paswd_form))
	paswd := hex.EncodeToString(paswd_hash[:])
	user_query, err := ri.GetUserByLogin(login)
	if err == nil || user_query.Login == login {
		fmt.Fprintln(w, "Incorrect user or password")
		log.Println("Incorrect user or password")
		return
	}
	token_encoder := sha256.New()
	token_encoder.Write([]byte(login))
	token_encoder.Write(paswd_hash[:])
	token := hex.EncodeToString(token_encoder.Sum(nil))
	err = ri.RegisterUser(models.User{
		Login:    login,
		Password: paswd,
		Token:    token,
	})
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

func (ri *RuntimeInstance) handlerReturnFormMessages(w http.ResponseWriter, r *http.Request) {
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

func (ri *RuntimeInstance) APIhandlerGetMessages(w http.ResponseWriter, r *http.Request) {
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
	} else {
		http.Error(w, "No chat_id", http.StatusBadRequest)
		return
	}
}

func (ri *RuntimeInstance) APIhandlerGetChats(w http.ResponseWriter, r *http.Request) {
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

func (ri *RuntimeInstance) APIhandlerSendMessage(w http.ResponseWriter, r *http.Request) {
	login_cookie, err := ri.UsernameFromCookie(w, r, "user")
	if err != nil {
		log.Println(err)
		return
	}
	var req models.MessageCreateRequest
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
