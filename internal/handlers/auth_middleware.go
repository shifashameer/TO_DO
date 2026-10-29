package handlers

import (
	"net/http"
	"strings"
	"todo-app/internal/auth"
	"todo-app/internal/store"
)

func RequireAuth(users store.UserStore, sessions *auth.Sessions, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		cookie, err := r.Cookie("session")
		if err == nil {
			if id, ok := sessions.Get(cookie.Value); ok {
				if user, err := users.GetByID(id); err == nil && user.ID == id {
					next.ServeHTTP(w, r)
					return
				}
			}
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})
}
