package httputil

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type LogEntry struct {
	Logger  *zap.Logger
	request *http.Request
}

func (l *LogEntry) Panic(v interface{}, stack []byte) {
	var f zap.Field
	if err, ok := v.(error); ok {
		f = zap.Error(err)
	} else {
		f = zap.Any("error", v)
	}
	l.Logger.Error("panic", f,
		zap.String("stack", string(stack)))
}

func (l *LogEntry) Write(status int, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	l.Logger.Info("access",
		zap.String("host", l.request.Host),
		zap.String("uri", l.request.RequestURI),
		zap.String("method", l.request.Method),
		zap.String("remote", l.request.RemoteAddr),
		zap.String("user_agent", l.request.UserAgent()),
		zap.Bool("tls", l.request.TLS != nil),
		zap.Int64("request_length", l.request.ContentLength),
		zap.Int("response_length", bytes),
		zap.Int("status", status),
		zap.Duration("elapsed", elapsed),
	)
}

type LogFormatter struct{ *zap.Logger }

func (f LogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	return &LogEntry{Logger: f.Logger.With(
		zap.String("request_id", middleware.GetReqID(r.Context())),
	), request: r}
}
