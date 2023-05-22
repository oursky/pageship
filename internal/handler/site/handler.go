package site

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/oursky/pageship/internal/config"
)

type Logger interface {
	Debug(format string, args ...any)
	Error(msg string, err error)
}

type HandlerConfig struct {
	DefaultSite string
	HostPattern string
}

type Handler struct {
	logger      Logger
	resolver    Resolver
	hostPattern config.HostPattern
	defaultSite string
	cache       *siteCache
}

func NewHandler(logger Logger, resolver Resolver, conf HandlerConfig) (*Handler, error) {
	cache, err := newSiteCache()
	if err != nil {
		return nil, fmt.Errorf("setup cache: %w", err)
	}

	defaultSite := conf.DefaultSite
	if defaultSite == "-" { // Allow setting '-' to disable default site
		defaultSite = ""
	}

	return &Handler{
		logger:      logger,
		resolver:    resolver,
		hostPattern: config.NewHostPattern(conf.HostPattern),
		defaultSite: defaultSite,
		cache:       cache,
	}, nil
}

func (h *Handler) resolveSite(r *http.Request) (*Descriptor, error) {
	matchedID, ok := h.hostPattern.MatchString(r.Host)
	if !ok {
		return nil, ErrSiteNotFound
	}

	if matchedID == h.defaultSite {
		// Default site must be accessed through empty ID
		return nil, ErrSiteNotFound
	}

	if matchedID == "" {
		if h.defaultSite == "" {
			// Default site is disabled; treat as not found
			return nil, ErrSiteNotFound
		}
		matchedID = h.defaultSite
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
	publicFS, err := fs.Sub(site.FS, site.Config.Public)
	if err != nil {
		h.logger.Error("construct site fs", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.FileServer(http.FS(publicFS)).ServeHTTP(w, r)
}
