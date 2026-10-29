package main

import (
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"os"
	"time"
	"todo-app/internal/handlers"
	"todo-app/internal/store"
	"todo-app/static"
)

func newRouter(todos store.Store, users store.UserStore, secureCookies bool) http.Handler {
	r := chi.NewRouter()
	// Reject cross-origin unsafe browser requests, including login/logout.
	r.Use(http.NewCrossOriginProtection().Handler)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServerFS(static.Files)))
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("ok"))
	})
	todoHandler := handlers.NewTodoHandler(todos)
	authHandler := handlers.NewAuthHandler(users)
	authHandler.SecureCookies = secureCookies
	r.Get("/login", authHandler.LoginPage)
	r.Post("/login", authHandler.Login)
	r.Post("/logout", authHandler.Logout)
	r.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler { return handlers.RequireAuth(users, authHandler.Sessions, next) })
		r.Get("/", todoHandler.Index)
		r.Get("/api/todos", todoHandler.GetTodos)
		r.Post("/todos", todoHandler.CreateTodo)
		r.Patch("/api/todos/{id}/toggle", todoHandler.ToggleComplete)
		r.Get("/todos/{id}/edit", todoHandler.EditTodo)
		r.Post("/todos/{id}", todoHandler.UpdateTodo)
		r.Delete("/api/todos/{id}", todoHandler.DeleteTodo)
	})
	return r
}
func main() {
	db, err := store.NewPostgresStore(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	address := os.Getenv("ADDR")
	if address == "" {
		address = "127.0.0.1:8080"
	}
	server := &http.Server{Addr: address, Handler: newRouter(db, db, os.Getenv("COOKIE_SECURE") == "true"), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("Listening on %s", address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Print(err)
	}
}
