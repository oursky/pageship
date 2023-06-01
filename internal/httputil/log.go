package httputil

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type LogEntry struct {
	logger  *zap.Logger
	request *http.Request
	User    string
	Site    string
}

func (l *LogEntry) Panic(v interface{}, stack []byte) {
	var f zap.Field
	if err, ok := v.(error); ok {
		f = zap.Error(err)
	} else {
		f = zap.Any("error", v)
	}
	l.logger.Error("panic", f,
		zap.String("request_id", middleware.GetReqID(l.request.Context())),
		zap.String("stack", string(stack)))
}

func (l *LogEntry) Write(status int, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	l.logger.Info("access",
		zap.String("request_id", middleware.GetReqID(l.request.Context())),
		zap.String("host", l.request.Host),
		zap.String("uri", l.request.RequestURI),
		zap.String("method", l.request.Method),
		zap.String("remote", l.request.RemoteAddr),
		zap.String("user_agent", l.request.UserAgent()),
		zap.String("user", l.User),
		zap.String("site", l.Site),
		zap.Bool("tls", l.request.TLS != nil),
		zap.Int64("request_length", l.request.ContentLength),
		zap.Int("response_length", bytes),
		zap.Int("status", status),
		zap.Duration("elapsed", elapsed),
	)
}

type LogFormatter struct{ *zap.Logger }

func (f LogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	return &LogEntry{logger: f.Logger, request: r}
}
