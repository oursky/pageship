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
	Logger  *zap.Logger
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

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		close(shutdown)
	}()

	s.Logger.Info("server starting", zap.String("addr", server.Addr))

	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.Logger.Error("failed to start server", zap.Error(err))
		return fmt.Errorf("failed to start server: %w", err)
	}
	<-shutdown

	return nil
}
