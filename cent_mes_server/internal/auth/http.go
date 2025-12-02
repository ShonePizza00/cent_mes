package auth

import (
	"cent_mes_server/models"
	"context"
	"net/http"
	"net/url"
	"time"
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
	mux.HandleFunc("/", h.ReturnLoginPage)
	mux.HandleFunc("GET /auth", h.ReturnLoginPage)
	mux.HandleFunc("GET /reg", h.ReturnRegisterPage)
	mux.HandleFunc("POST /auth", h.Login)
	mux.HandleFunc("POST /reg", h.Register)
}

func (h *Handler) ReturnLoginPage(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("user")
	if err != nil {
		http.ServeFile(w, r, "html/login.html")
		return
	}
	err2 := h.service.GetUserByToken(h.ctx, &models.User{Token: c.Value})
	if err2 != nil {
		http.ServeFile(w, r, "html/login.html")
		return
	}
	http.Redirect(w, r, "/messages", http.StatusSeeOther)
}

func (h *Handler) ReturnRegisterPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "html/register.html")
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	login := r.FormValue("login")
	password := r.FormValue("password")
	usr := models.User{
		Login:    login,
		Password: password,
	}
	if err := h.service.Login(h.ctx, &usr); err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "user",
		Value:    usr.Token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/messages", http.StatusSeeOther)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	login := r.FormValue("login")
	_ = r.FormValue("email")
	password := r.FormValue("password")
	usr := models.User{
		Login:    login,
		Password: password,
	}
	if err := h.service.RegisterUser(h.ctx, &usr); err != nil {
		http.Error(w, "User already exists", http.StatusBadRequest)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "user",
		Value:    url.QueryEscape(usr.Token),
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/messages", http.StatusSeeOther)
}
