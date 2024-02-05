package context

import (
	"context"
	"net/http"
)

type contextKey struct{}

type contextValue struct {
	Error error
}

func WithSiteContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey{}, &contextValue{})
}

func siteContextValue(ctx context.Context) *contextValue {
	return ctx.Value(contextKey{}).(*contextValue)
}

func Error(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, "internal server error", http.StatusInternalServerError)
	siteContextValue(r.Context()).Error = err
}
