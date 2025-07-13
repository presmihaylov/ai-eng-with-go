package models

import "time"

type Memory struct {
	ID            string    `json:"id" db:"id"`
	MemoryContent string    `json:"memory_content" db:"memory_content"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}