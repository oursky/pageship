package command

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type HTTPServerLogger interface {
	Info(format string, args ...any)
	Error(msg string, err error)
}

type HTTPServer struct {
	Logger  HTTPServerLogger
	Addr    string
	Handler http.Handler
}

func (s *HTTPServer) Run(ctx context.Context) error {
	server := http.Server{
		Addr:    s.Addr,
		Handler: s.Handler,
	}

	shutdown := make(chan struct{})
	go func() {
		<-ctx.Done()
		s.Logger.Info("server stopping...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		close(shutdown)
	}()

	s.Logger.Info("server starting at %s", server.Addr)

	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.Logger.Error("failed to start server", err)
		return fmt.Errorf("failed to start server: %w", err)
	}
	<-shutdown

	return nil
}
