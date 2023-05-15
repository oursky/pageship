package sites

import (
	"context"
	"io/fs"
)

type Descriptor struct {
	Site string
	FS   fs.FS
}

var contextKey struct{}

func withSite(ctx context.Context, desc *Descriptor) context.Context {
	return context.WithValue(ctx, contextKey, desc)
}

func SiteFromContext(ctx context.Context) (*Descriptor, bool) {
	desc, ok := ctx.Value(contextKey).(*Descriptor)
	return desc, ok
}
