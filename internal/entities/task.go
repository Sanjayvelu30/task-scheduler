package entities

import "time"

type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // e.g. "pending", "in-progress", "completed"
	CreatedAt   time.Time `json:"created_at"`
}
