package chats

import (
	"cent_mes_server/models"
	"context"
	"encoding/json"
	"net/http"
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
	mux.HandleFunc("GET /api/chats", h.GetChats)
}

func (h *Handler) GetChats(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("user")
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}
	chats, err := h.service.GetChats(h.ctx, &models.User{Token: c.Value})
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(chats)
}
