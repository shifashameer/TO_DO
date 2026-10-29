package store

import (
	"errors"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strings"

	"todo-app/internal/models"

	"gorm.io/gorm/logger"

	"github.com/google/uuid"
)

type PostgresStore struct {
	db *gorm.DB //A PostgresStore contains a GORM database connection.
}

// NewPostgresStore connects using caller-provided configuration and checks migrations.
func NewPostgresStore(dsn string) (*PostgresStore, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, errors.New("unable to connect to PostgreSQL; check DATABASE_URL and server availability")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.New("unable to initialize PostgreSQL pool")
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	if err := db.AutoMigrate(&models.Todo{}, &models.User{}); err != nil {
		sqlDB.Close()
		return nil, errors.New("database migration failed; check schema permissions")
	}
	return &PostgresStore{db: db}, nil
}
func (p *PostgresStore) Close() error {
	db, err := p.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
func databaseError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func (p *PostgresStore) Create(todo models.Todo) error {
	todo.ID = uuid.New().String()

	return p.db.Create(&todo).Error
}
func (p *PostgresStore) Get(id string) (models.Todo, error) {
	var todo models.Todo

	err := p.db.First(&todo, "id = ?", id).Error
	if err != nil {
		return models.Todo{}, databaseError(err)
	}

	return todo, nil
}
func (p *PostgresStore) List(options ListOptions) ([]models.Todo, error) {
	options, err := options.Normalize()
	if err != nil {
		return nil, err
	}
	todos := make([]models.Todo, 0)

	query := p.db

	// Filter by status
	switch options.Status {
	case "completed":
		query = query.Where("completed = ?", true)
	case "active":
		query = query.Where("completed = ?", false)
	}

	// Sort
	order := "ASC"

	if options.Order == "desc" {
		order = "DESC"
	}

	switch options.Sort {
	case "title":
		query = query.Order("title " + order)
	case "created_at":
		query = query.Order("created_at " + order)
	}

	err = query.Order("id ASC").Find(&todos).Error

	return todos, err
}
func (p *PostgresStore) Update(todo models.Todo) error {
	result := p.db.Model(&models.Todo{}).Where("id = ?", todo.ID).Select("title", "description", "completed").Updates(todo)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
func (p *PostgresStore) Delete(id string) error {
	result := p.db.Delete(&models.Todo{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
func (p *PostgresStore) GetByUsername(username string) (models.User, error) {
	var user models.User

	err := p.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return models.User{}, databaseError(err)
	}

	return user, nil
}
func (s *PostgresStore) GetByID(id string) (models.User, error) {
	var user models.User

	err := s.db.First(&user, "id = ?", id).Error
	if err != nil {
		return models.User{}, databaseError(err)
	}

	return user, nil
}

func (p *PostgresStore) CreateUser(user models.User) error {
	return p.db.Create(&user).Error
}

var _ Store = (*PostgresStore)(nil)
var _ UserStore = (*PostgresStore)(nil)
