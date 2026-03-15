package domain

import "time"

type User struct {
	ID        int64
	Email     string
	FirstName string
	LastName  string
	CreatedAt time.Time
}

type UserWithPassword struct {
	User         *User
	PasswordHash string
}
