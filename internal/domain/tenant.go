package domain

import "time"

type Tenant struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}
