package dao

import (
	"time"

	"github.com/0xMain/subscription-hub/internal/domain"
)

type User struct {
	ID           int64     `db:"id"`
	Email        string    `db:"email"`
	FirstName    string    `db:"first_name"`
	LastName     string    `db:"last_name"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

func (d *User) ToDomain() *domain.User {
	if d == nil {
		return nil
	}

	return &domain.User{
		ID:        d.ID,
		FirstName: d.FirstName,
		LastName:  d.LastName,
		Email:     d.Email,
		CreatedAt: d.CreatedAt,
	}
}
