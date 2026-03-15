package domain

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type InvoiceStatus string

const (
	StatusDraft     InvoiceStatus = "draft"
	StatusSent      InvoiceStatus = "sent"
	StatusPaid      InvoiceStatus = "paid"
	StatusCancelled InvoiceStatus = "cancelled"
)

type Invoice struct {
	ID         int64
	TenantID   int64
	CustomerID int64
	Number     string
	Status     InvoiceStatus
	Amount     decimal.Decimal
	IssuedDate *time.Time
	DueDate    time.Time
	PaidAt     *time.Time
	CreatedAt  *time.Time
}

func ParseInvoiceStatus(s string) (InvoiceStatus, error) {
	switch s {
	case "draft":
		return StatusDraft, nil
	case "sent":
		return StatusSent, nil
	case "paid":
		return StatusPaid, nil
	case "cancelled":
		return StatusCancelled, nil
	default:
		return "", fmt.Errorf("неизвестный статус счета: %s", s)
	}
}
