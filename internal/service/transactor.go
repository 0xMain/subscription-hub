package service

import "context"

type transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
