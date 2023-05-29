package command

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type HTTPServer struct {
	Logger *zap.Logger
	http.Server
}

func (s *HTTPServer) Run(ctx context.Context) error {
	s.Server.ErrorLog = zap.NewStdLog(s.Logger)
	if s.Server.ReadHeaderTimeout == 0 {
		s.Server.ReadHeaderTimeout = 5 * time.Second
	}
	if s.Server.ReadTimeout == 0 {
		s.Server.ReadTimeout = 10 * time.Second
	}
	if s.Server.WriteTimeout == 0 {
		s.Server.WriteTimeout = 10 * time.Second
	}
	if s.Server.IdleTimeout == 0 {
		s.Server.IdleTimeout = 120 * time.Second
	}
	if s.Server.MaxHeaderBytes == 0 {
		s.Server.MaxHeaderBytes = 10 * 1024
	}

	shutdown := make(chan struct{})
	go func() {
		<-ctx.Done()
		s.Logger.Info("server stopping...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.Server.Shutdown(ctx)
		close(shutdown)
	}()

	s.Logger.Info("server starting", zap.String("addr", s.Server.Addr))

	err := s.Server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.Logger.Error("failed to start server", zap.Error(err))
		return fmt.Errorf("failed to start server: %w", err)
	}
	<-shutdown

	return nil
}
