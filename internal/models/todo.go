package models

import "time"

type Todo struct {
	ID          string `gorm:"primaryKey"`
	Title       string
	Description string
	Completed   bool
	CreatedAt   time.Time
}

//like a blueprint to know what the todo does
