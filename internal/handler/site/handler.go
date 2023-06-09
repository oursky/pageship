package site

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/oursky/pageship/internal/cache"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/httputil"
	"github.com/oursky/pageship/internal/site"
	"go.uber.org/zap"
)

const (
	cacheSize int           = 100
	cacheTTL  time.Duration = time.Second * 1
)

type HandlerConfig struct {
	HostPattern string
	Middlewares []Middleware
}

type Handler struct {
	ctx         context.Context
	logger      *zap.Logger
	resolver    site.Resolver
	hostPattern *config.HostPattern
	cache       *cache.Cache[*SiteHandler]
	middlewares []Middleware
}

func NewHandler(ctx context.Context, logger *zap.Logger, resolver site.Resolver, conf HandlerConfig) (*Handler, error) {
	h := &Handler{
		ctx:         ctx,
		logger:      logger,
		resolver:    resolver,
		hostPattern: config.NewHostPattern(conf.HostPattern),
		middlewares: conf.Middlewares,
	}

	cache, err := cache.NewCache(cacheSize, cacheTTL, h.doResolve)
	if err != nil {
		return nil, fmt.Errorf("setup cache: %w", err)
	}
	h.cache = cache

	return h, nil
}

func (h *Handler) resolveSite(host string) (*SiteHandler, error) {
	matchedID, ok := h.hostPattern.MatchString(host)
	if !ok {
		return nil, site.ErrSiteNotFound
	}

	return h.cache.Load(matchedID)
}

func (h *Handler) doResolve(matchedID string) (*SiteHandler, error) {
	desc, err := h.resolver.Resolve(h.ctx, matchedID)
	if err != nil {
		return nil, err
	}

	return NewSiteHandler(desc, h.middlewares), nil
}

func (h *Handler) AllowAnyDomain() bool {
	return h.resolver.AllowAnyDomain()
}

func (h *Handler) CheckValidDomain(name string) error {
	if h.resolver.AllowAnyDomain() {
		return nil
	}
	_, err := h.resolveSite(name)
	return err
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, err := h.resolveSite(r.Host)
	if errors.Is(err, site.ErrSiteNotFound) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		h.logger.Error("failed to resolve site", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Debug("resolved site", zap.String("site", handler.ID()))
	entry := middleware.GetLogEntry(r).(*httputil.LogEntry)
	entry.Logger = entry.Logger.With(zap.String("site", handler.ID()))

	handler.ServeHTTP(w, r)
}
