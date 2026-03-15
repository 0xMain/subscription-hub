package domain

import "time"

type Customer struct {
	ID        int64
	TenantID  int64
	FirstName string
	LastName  string
	Email     string
	CreatedAt *time.Time
}
