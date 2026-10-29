package handlers

import (
	"net/http"
	"strings"
	"time"
	"todo-app/internal/auth"
	"todo-app/internal/store"
)

type AuthHandler struct {
	userStore     store.UserStore
	Sessions      *auth.Sessions
	SecureCookies bool
}

func NewAuthHandler(users store.UserStore) *AuthHandler {
	return &AuthHandler{userStore: users, Sessions: auth.NewSessions()}
}
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}
	username := strings.TrimSpace(r.PostForm.Get("username"))
	password := r.PostForm.Get("password")
	user, err := h.userStore.GetByUsername(username)
	if err != nil || user.ID == "" || !auth.CheckPassword(password, user.PasswordHash) {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}
	token, err := h.Sessions.Create(user.ID)
	if err != nil {
		http.Error(w, "Unable to create session", 500)
		return
	}
	if old, err := r.Cookie("session"); err == nil {
		h.Sessions.Delete(old.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "session", Value: token, Path: "/", HttpOnly: true, Secure: h.SecureCookies, SameSite: http.SameSiteStrictMode, MaxAge: int(auth.SessionTTL.Seconds()), Expires: time.Now().Add(auth.SessionTTL)})
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	render(w, "login.html", nil)
}
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session"); err == nil {
		h.Sessions.Delete(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "session", Value: "", Path: "/", MaxAge: -1, Expires: time.Unix(1, 0), HttpOnly: true, Secure: h.SecureCookies, SameSite: http.SameSiteStrictMode})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
