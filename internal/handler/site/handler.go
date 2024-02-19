package site

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/oursky/pageship/internal/cache"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/domain"
	"github.com/oursky/pageship/internal/handler/site/middleware"
	"github.com/oursky/pageship/internal/httputil"
	"github.com/oursky/pageship/internal/models"
	"github.com/oursky/pageship/internal/site"
	"go.uber.org/zap"
)

const (
	cacheSize int           = 100
	cacheTTL  time.Duration = time.Second * 1
)

type HandlerConfig struct {
	HostPattern         string
	MiddlewaresFunc     func(middleware.ContentCacheType) []middleware.Middleware
	ContentCacheMaxSize int64
}

type Handler struct {
	ctx            context.Context
	logger         *zap.Logger
	domainResolver domain.Resolver
	siteResolver   site.Resolver
	hostPattern    *config.HostPattern
	cache          *cache.Cache[*SiteHandler]
	middlewares    []middleware.Middleware
	contentCache   middleware.ContentCacheType
}

func NewHandler(ctx context.Context, logger *zap.Logger, domainResolver domain.Resolver, siteResolver site.Resolver, conf HandlerConfig) (*Handler, error) {
	h := &Handler{
		ctx:            ctx,
		logger:         logger,
		domainResolver: domainResolver,
		siteResolver:   siteResolver,
		hostPattern:    config.NewHostPattern(conf.HostPattern),
	}

	c, err := cache.NewCache(cacheSize, cacheTTL, h.doResolveHandler)
	if err != nil {
		return nil, fmt.Errorf("setup cache: %w", err)
	}
	h.cache = c

	load := func(r io.ReadSeeker) (*bytes.Buffer, int64, error) {
		b, err := io.ReadAll(r)
		nb := bytes.NewBuffer(b)
		if err != nil {
			return nb, 0, err
		}
		return nb, int64(nb.Len()), nil
	}
	cc, err := cache.NewContentCache[middleware.ContentCacheKey](conf.ContentCacheMaxSize, false, load) //TODO: pass from config
	if err != nil {
		return nil, fmt.Errorf("setup content cache: %w", err)
	}
	h.contentCache = cc

	h.middlewares = conf.MiddlewaresFunc(h.contentCache)
	return h, nil
}

func (h *Handler) resolveHandler(host string) (*SiteHandler, error) {
	return h.cache.Load(host)
}

func (h *Handler) ResolveSite(host string) (*site.Descriptor, error) {
	matchedID, ok := h.hostPattern.MatchString(host)
	if !ok {
		hostname, _, err := net.SplitHostPort(host)
		if err != nil {
			hostname = host
		}
		id, err := h.domainResolver.Resolve(h.ctx, hostname)
		if errors.Is(err, domain.ErrDomainNotFound) {
			return nil, site.ErrSiteNotFound
		} else if err != nil {
			return nil, err
		}
		matchedID = id
	}

	desc, err := h.siteResolver.Resolve(h.ctx, matchedID)
	if err != nil {
		return nil, err
	}

	return desc, nil
}

func (h *Handler) doResolveHandler(host string) (*SiteHandler, error) {
	desc, err := h.ResolveSite(host)
	if err != nil {
		return nil, err
	}

	return NewSiteHandler(desc, h.middlewares), nil
}

func (h *Handler) AcceptsAllDomain() bool {
	return h.siteResolver.IsWildcard()
}

func (h *Handler) CheckValidDomain(hostname string) error {
	if h.siteResolver.IsWildcard() {
		return nil
	}
	_, err := h.ResolveSite(hostname)
	return err
}

func (h *Handler) checkAuthz(r *http.Request, handler *SiteHandler) error {
	// Allow all access unless explicitly configured.
	access := handler.desc.Config.Access
	if len(access) == 0 {
		return nil
	}

	var credentials []models.CredentialID

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		credentials = append(credentials, models.CredentialIP(ip))
	}

	_, err = models.CheckACLAuthz(access, credentials)
	return err
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, err := h.resolveHandler(r.Host)
	if errors.Is(err, site.ErrSiteNotFound) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		h.logger.Error("failed to resolve site", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Debug("resolved site", zap.String("site", handler.ID()))
	entry := chiMiddleware.GetLogEntry(r)
	e := entry.(*httputil.LogEntry)
	e.Logger = e.Logger.With(zap.String("site", handler.ID()))

	if err := h.checkAuthz(r, handler); err != nil {
		http.NotFound(w, r)
		return
	}

	handler.ServeHTTP(w, r)
}
