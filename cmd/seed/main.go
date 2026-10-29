package main

import (
	"github.com/google/uuid"
	"log"
	"os"
	"strings"
	"todo-app/internal/auth"
	"todo-app/internal/models"
	"todo-app/internal/store"
)

func main() {
	username := strings.TrimSpace(os.Getenv("SEED_USERNAME"))
	password := os.Getenv("SEED_PASSWORD")
	if username == "" || len(password) < 12 || len(password) > 72 {
		log.Fatal("SEED_USERNAME and SEED_PASSWORD (12–72 bytes) are required")
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatal("Unable to hash password")
	}
	db, err := store.NewPostgresStore(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.CreateUser(models.User{ID: uuid.NewString(), Username: username, PasswordHash: hash}); err != nil {
		log.Fatal("Unable to create user; username may already exist (existing users are not overwritten)")
	}
	log.Println("User created successfully")
}
