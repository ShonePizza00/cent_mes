package messages

import (
	"cent_mes_server/models"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"text/template"
)

type Handler struct {
	service *Service
	ctx     context.Context
}

func NewHandler(ctx context.Context, svc *Service) *Handler {
	return &Handler{
		service: svc,
		ctx:     ctx}
}

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("GET /messages", h.ReturnMessagesPage)
	mux.HandleFunc("GET /api/messages", h.APIGetMessages)
	mux.HandleFunc("POST /api/messages", h.APISendMessage)
}

func (h *Handler) ReturnMessagesPage(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("user")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	usr := models.User{
		Token: c.Value,
	}
	err = h.service.GetUserByToken(h.ctx, &usr)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}
	tmpl := template.Must(template.ParseFiles("html/messenger.html"))
	if err := tmpl.Execute(w, usr.Login); err != nil {
		return
	}
}

func (h *Handler) APIGetMessages(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("user")
	if err != nil {
		return
	}
	q_vals := r.URL.Query()
	if q_vals.Has("chat_id") {
		chat_id, _ := strconv.Atoi(q_vals.Get("chat_id"))
		after_id, _ := strconv.Atoi(q_vals.Get("after_id"))

		cf := models.ChatFetch{
			ChatID:  int64(chat_id),
			AfterID: int64(after_id),
		}
		mess, err := h.service.GetMessages(h.ctx, &cf, &models.User{Token: c.Value})
		if err != nil {
			if err == ErrForbidden {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
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

func (h *Handler) APISendMessage(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("user")
	if err != nil {
		return
	}
	var req models.MessageCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad json", http.StatusBadRequest)
		return
	}
	err = h.service.SendMessage(h.ctx, &req, c.Value)
	if err != nil {
		if err == ErrInvalidSender {
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; chatset=utf-8")
	json.NewEncoder(w).Encode(req)
}
