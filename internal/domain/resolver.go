package domain

import (
	"context"
	"errors"
)

var ErrDomainNotFound = errors.New("domain not found")

type Resolver interface {
	Kind() string
	Resolve(ctx context.Context, hostname string) (string, error)
}
