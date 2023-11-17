package db

// schema.go provides data models in DB
import (
	"time"
)

// Task corresponds to a row in `tasks` table
type Task struct {
	ID        uint64    `db:"id"`
	Title     string    `db:"title"`
	CreatedAt time.Time `db:"created_at"`
	Deadline time.Time `db:"deadline"`
	IsDone    bool      `db:"is_done"`
	Comment   string    `db:"comment"`
	DaysLeft int
}

type User struct {
    ID        uint64    `db:"id"`
    Name      string    `db:"name"`
    Password  []byte    `db:"password"`
}
