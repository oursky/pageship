package domain

import (
	"context"
)

type ResolverNull struct{}

func (h *ResolverNull) Kind() string { return "null" }

func (h *ResolverNull) Resolve(ctx context.Context, hostname string) (string, error) {
	return "", ErrDomainNotFound
}
