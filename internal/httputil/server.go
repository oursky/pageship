package httputil

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/caddyserver/certmagic"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type ServerTLSConfig struct {
	Storage       certmagic.Storage
	ACMEDirectory string
	ACMEEmail     string
	Addr          string
	DomainNames   []string
	CheckDomain   func(name string) error
}

type Server struct {
	Logger  *zap.Logger
	Addr    string
	Handler http.Handler

	TLS *ServerTLSConfig
}

func (s *Server) makeServer(handler http.Handler) *http.Server {
	return &http.Server{
		ErrorLog:          zap.NewStdLog(s.Logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    10 * 1024,
		Handler:           handler,
	}
}

func (s *Server) serveHTTP(ctx context.Context, server *http.Server) error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	s.Logger.Info("http server starting", zap.String("addr", ln.Addr().String()))
	err = server.Serve(ln)

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start http server: %w", err)
	}
	return nil
}

func (s *Server) serveTLS(ctx context.Context, server *http.Server) error {
	ln, err := tls.Listen("tcp", s.TLS.Addr, server.TLSConfig)
	if err != nil {
		return err
	}

	s.Logger.Info("https server starting", zap.String("addr", ln.Addr().String()))
	err = server.Serve(ln)

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start https server: %w", err)
	}
	return nil
}

func (s *Server) setupTLS(ctx context.Context, httpHandler *http.Handler) (*http.Server, error) {
	magic := &certmagic.Config{
		Storage: s.TLS.Storage,
		Logger:  s.Logger.Named("cert"),
		OnDemand: &certmagic.OnDemandConfig{
			DecisionFunc: s.TLS.CheckDomain,
		},
	}

	cache := certmagic.NewCache(certmagic.CacheOptions{
		Logger: zap.NewNop(),
		GetConfigForCert: func(cert certmagic.Certificate) (*certmagic.Config, error) {
			return magic, nil
		},
	})
	magic = certmagic.New(cache, *magic)

	ca := certmagic.LetsEncryptProductionCA
	if s.TLS.ACMEDirectory != "" {
		ca = s.TLS.ACMEDirectory
	}
	issuer := certmagic.NewACMEIssuer(magic, certmagic.ACMEIssuer{
		Logger: zap.NewNop(),
		CA:     ca,
		Email:  s.TLS.ACMEEmail,
		Agreed: true,
	})
	magic.Issuers = []certmagic.Issuer{issuer}

	if err := magic.ManageAsync(ctx, s.TLS.DomainNames); err != nil {
		return nil, err
	}

	tlsConf := magic.TLSConfig()
	tlsConf.NextProtos = append([]string{"h2", "http/1.1"}, tlsConf.NextProtos...)
	tlsConf.GetCertificate = func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
		// Don't timeout when handling certificate
		chi.Conn.SetReadDeadline(time.Time{})
		chi.Conn.SetWriteDeadline(time.Time{})

		cert, err := magic.GetCertificate(chi)

		dl := time.Now().Add(10 * time.Second)
		chi.Conn.SetReadDeadline(dl)
		chi.Conn.SetWriteDeadline(dl)

		return cert, err
	}

	server := s.makeServer(*httpHandler)
	server.TLSConfig = tlsConf
	*httpHandler = issuer.HTTPChallengeHandler(http.HandlerFunc(redirectToHTTPS))

	return server, nil
}

func (s *Server) Run(ctx context.Context) error {
	httpHandler := s.Handler
	var tlsServer *http.Server
	if s.TLS != nil {
		server, err := s.setupTLS(ctx, &httpHandler)
		if err != nil {
			s.Logger.Error("failed to setup TLS", zap.Error(err))
			return err
		}
		tlsServer = server
	}

	httpServer := s.makeServer(httpHandler)

	shutdown := make(chan struct{})
	go func() {
		<-ctx.Done()
		s.Logger.Info("server stopping...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		g, ctx := errgroup.WithContext(ctx)
		g.Go(func() error { return httpServer.Shutdown(ctx) })
		if tlsServer != nil {
			g.Go(func() error { return tlsServer.Shutdown(ctx) })
		}
		g.Wait()
		close(shutdown)
	}()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return s.serveHTTP(ctx, httpServer) })
	if tlsServer != nil {
		g.Go(func() error { return s.serveTLS(ctx, tlsServer) })
	}
	err := g.Wait()
	if err != nil {
		s.Logger.Error("failed to start server", zap.Error(err))
		return err
	}
	<-shutdown

	return nil
}

func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		http.Error(w, "Use HTTPS", http.StatusBadRequest)
		return
	}
	target := "https://" + stripPort(r.Host) + r.URL.RequestURI()
	http.Redirect(w, r, target, http.StatusFound)
}

func stripPort(hostport string) string {
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		return hostport
	}
	return net.JoinHostPort(host, "443")
}
