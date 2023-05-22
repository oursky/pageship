package site

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"regexp"
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
	hostRegex   *regexp.Regexp
	defaultSite string
	cache       *siteCache
}

func NewHandler(logger Logger, resolver Resolver, conf HandlerConfig) (*Handler, error) {
	hostRegex, err := regexp.Compile(`^` + conf.HostPattern + `(?:\:\d+)?$`)
	if err != nil {
		return nil, fmt.Errorf("invalid host pattern: %w", err)
	}

	cache, err := newSiteCache()
	if err != nil {
		return nil, fmt.Errorf("setup cache: %w", err)
	}

	return &Handler{
		logger:      logger,
		resolver:    resolver,
		hostRegex:   hostRegex,
		defaultSite: conf.DefaultSite,
		cache:       cache,
	}, nil
}

func (h *Handler) resolveSite(r *http.Request) (*Descriptor, error) {
	matches := h.hostRegex.FindStringSubmatch(r.Host)
	if len(matches) != 2 {
		return nil, ErrSiteNotFound
	}

	matchedID := matches[1]
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

	return h.cache.Load(matchedID, h.resolver.Resolve)
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
	h.logger.Debug("resolved site: %s/%s", site.AppID, site.SiteName)

	h.serve(site, w, r)
}

func (h *Handler) serve(site *Descriptor, w http.ResponseWriter, r *http.Request) {
	publicFS, err := fs.Sub(site.FS, site.Config.Public)
	if err != nil {
		h.logger.Error("construct site fs", err)
	}

	http.FileServer(http.FS(publicFS)).ServeHTTP(w, r)
}
