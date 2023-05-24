package site

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/oursky/pageship/internal/cache"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/site"
)

const (
	cacheSize int           = 100
	cacheTTL  time.Duration = time.Second * 1
)

type Logger interface {
	Debug(format string, args ...any)
	Error(msg string, err error)
}

type HandlerConfig struct {
	HostPattern string
	Middlewares []Middleware
}

type Handler struct {
	logger      Logger
	resolver    site.Resolver
	hostPattern *config.HostPattern
	cache       *cache.Cache[*SiteHandler]
	middlewares []Middleware
}

func NewHandler(logger Logger, resolver site.Resolver, conf HandlerConfig) (*Handler, error) {
	cache, err := cache.NewCache[*SiteHandler](cacheSize, cacheTTL)
	if err != nil {
		return nil, fmt.Errorf("setup cache: %w", err)
	}

	return &Handler{
		logger:      logger,
		resolver:    resolver,
		hostPattern: config.NewHostPattern(conf.HostPattern),
		cache:       cache,
		middlewares: conf.Middlewares,
	}, nil
}

func (h *Handler) resolveSite(r *http.Request) (*SiteHandler, error) {
	matchedID, ok := h.hostPattern.MatchString(r.Host)
	if !ok {
		return nil, site.ErrSiteNotFound
	}

	resolve := func(matchedID string) (*SiteHandler, error) {
		desc, err := h.resolver.Resolve(r.Context(), matchedID)
		if err != nil {
			return nil, err
		}

		return NewSiteHandler(desc, h.middlewares), nil
	}
	return h.cache.Load(matchedID, resolve)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, err := h.resolveSite(r)
	if errors.Is(err, site.ErrSiteNotFound) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		h.logger.Error("failed to resolve site", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	h.logger.Debug("resolved site: %s", handler.ID())

	handler.ServeHTTP(w, r)
}
