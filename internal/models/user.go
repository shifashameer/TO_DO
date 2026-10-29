package models

import "time"

type User struct {
	ID           string `gorm:"primaryKey"`
	Username     string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null" json:"-"`
	CreatedAt    time.Time
}
