package store

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"strings"
	"testing"
	"todo-app/internal/models"
)

// Integration tests require an explicit dedicated database ending in _test.
// Every test uses a fresh schema inside a transaction rolled back on cleanup.
func newTestStore(t *testing.T) *PostgresStore {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PostgreSQL integration requires TEST_DATABASE_URL (database name must end in _test)")
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal("invalid TEST_DATABASE_URL")
	}
	if !strings.HasSuffix(cfg.Database, "_test") {
		t.Fatal("refusing integration tests: database name must end in _test")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("cannot connect to dedicated test database")
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Close() })
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	t.Cleanup(func() { tx.Rollback() })
	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := tx.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Exec("SET LOCAL search_path TO " + schema).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.AutoMigrate(&models.Todo{}, &models.User{}); err != nil {
		t.Fatal(err)
	}
	return &PostgresStore{db: tx}
}
func TestPostgresStoreContract(t *testing.T) { storeContract(t, newTestStore(t)) }
func TestPostgresUserErrors(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.GetByID("missing"); err != ErrNotFound {
		t.Fatalf("missing user: %v", err)
	}
	if _, err := s.GetByUsername("missing"); err != ErrNotFound {
		t.Fatalf("missing username: %v", err)
	}
	u := models.User{ID: uuid.NewString(), Username: "unique", PasswordHash: "test-hash"}
	if err := s.CreateUser(u); err != nil {
		t.Fatal(err)
	}
	if got, err := s.GetByID(u.ID); err != nil || got.Username != u.Username {
		t.Fatal("user lookup failed")
	}
	u.ID = uuid.NewString()
	if err := s.CreateUser(u); err == nil {
		t.Fatal("duplicate username accepted")
	}
}
func TestMissingDatabaseConfig(t *testing.T) {
	if _, err := NewPostgresStore(""); err == nil {
		t.Fatal("missing configuration accepted")
	}
}
