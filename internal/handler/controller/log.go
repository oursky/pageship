package controller

import (
	"net/http"

	"go.uber.org/zap"
)

type loggers struct {
	*zap.Logger
	authn *zap.Logger
}

func log(r *http.Request) *zap.Logger {
	return get[*loggers](r).Logger
}
