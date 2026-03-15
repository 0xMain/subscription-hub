package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type InvoiceItem struct {
	ID          int64
	InvoiceID   int64
	TenantID    int64
	Description string
	Quantity    int64
	UnitPrice   decimal.Decimal
	Amount      decimal.Decimal
	CreatedAt   *time.Time
}
