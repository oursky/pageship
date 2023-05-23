package site

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/oursky/pageship/internal/cache"
	"github.com/oursky/pageship/internal/config"
)

type Logger interface {
	Debug(format string, args ...any)
	Error(msg string, err error)
}

type HandlerConfig struct {
	HostPattern string
}

type Handler struct {
	logger      Logger
	resolver    Resolver
	hostPattern *config.HostPattern
	cache       *cache.Cache[Descriptor]
}

func NewHandler(logger Logger, resolver Resolver, conf HandlerConfig) (*Handler, error) {
	cache, err := newSiteCache()
	if err != nil {
		return nil, fmt.Errorf("setup cache: %w", err)
	}

	return &Handler{
		logger:      logger,
		resolver:    resolver,
		hostPattern: config.NewHostPattern(conf.HostPattern),
		cache:       cache,
	}, nil
}

func (h *Handler) resolveSite(r *http.Request) (*Descriptor, error) {
	matchedID, ok := h.hostPattern.MatchString(r.Host)
	if !ok {
		return nil, ErrSiteNotFound
	}

	resolve := func(matchedID string) (*Descriptor, error) {
		return h.resolver.Resolve(r.Context(), matchedID)
	}
	return h.cache.Load(matchedID, resolve)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	site, err := h.resolveSite(r)
	if errors.Is(err, ErrSiteNotFound) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		h.logger.Error("failed to resolve site", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	h.logger.Debug("resolved site: %s", site.ID)

	h.serve(site, w, r)
}

func (h *Handler) serve(site *Descriptor, w http.ResponseWriter, r *http.Request) {
	fsys := site.FSFunc(r.Context())
	publicFS, err := fs.Sub(fsys, site.Config.Public)
	if err != nil {
		h.logger.Error("construct site fs", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.FileServer(http.FS(publicFS)).ServeHTTP(w, r)
}
