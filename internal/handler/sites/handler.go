package sites

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/oursky/pageship/internal/config"
	"go.uber.org/zap"
)

type siteRouterHandler struct {
	logger        *zap.Logger
	hostRegex     *regexp.Regexp
	defaultSite   string
	fs            fs.FS
	useAdhocSites bool
	sites         map[string]config.SitesConfigEntry

	next http.Handler
}

func NewHandler(fs fs.FS, conf *config.SitesConfig, next http.Handler) (http.Handler, error) {
	logger := zap.L().Named("sites")

	hostRegex, err := regexp.Compile(`^` + conf.HostPattern + `(?:\:\d+)?$`)
	if err != nil {
		return nil, fmt.Errorf("invalid host pattern: %w", err)
	}

	useAdhocSites := len(conf.Sites) == 0
	if useAdhocSites {
		logger.Info("using ad-hoc site resolution")
	}

	return &siteRouterHandler{
		logger:        logger,
		hostRegex:     hostRegex,
		defaultSite:   conf.DefaultSite,
		fs:            fs,
		useAdhocSites: useAdhocSites,
		sites:         conf.Sites,
		next:          next,
	}, nil
}

func (h *siteRouterHandler) lookupAdhocSite(site string) (*Descriptor, bool) {
	dnsLabels := strings.Split(path.Clean(site), ".")
	pathSegments := make([]string, len(dnsLabels))
	for i, label := range dnsLabels {
		pathSegments[len(pathSegments)-1-i] = label
	}

	fsys, err := fs.Sub(h.fs, path.Join(pathSegments...))
	if err != nil {
		h.logger.Warn("make context fs", zap.Error(err))
		return nil, false
	}
	if err := checkSiteFS(fsys); err != nil {
		if !os.IsNotExist(err) {
			h.logger.Warn("invalid context fs", zap.Error(err))
		}
		return nil, false
	}
	return &Descriptor{
		Site: site,
		FS:   fsys,
	}, true
}

func (h *siteRouterHandler) lookupConfigSite(site string) (*Descriptor, bool) {
	entry, ok := h.sites[site]
	if !ok {
		return nil, false
	}

	fs, err := fs.Sub(h.fs, entry.Context)
	if err != nil {
		h.logger.Warn("make context fs", zap.Error(err))
		return nil, false
	}

	if err := checkSiteFS(fs); err != nil {
		h.logger.Warn("invalid context fs", zap.Error(err))
		return nil, false
	}

	return &Descriptor{Site: site, FS: fs}, true
}

func (h *siteRouterHandler) lookupSite(r *http.Request) (*Descriptor, bool) {
	matches := h.hostRegex.FindStringSubmatch(r.Host)
	if len(matches) != 2 {
		return nil, false
	}

	id := matches[1]
	if id == h.defaultSite {
		// Default site must be accessed through empty ID
		return nil, false
	}

	if id == "" {
		if h.defaultSite == "" {
			// Default site is disabled; treat as not found
			return nil, false
		}
		id = h.defaultSite
	}

	if h.useAdhocSites {
		return h.lookupAdhocSite(id)
	} else {
		return h.lookupConfigSite(id)
	}
}

func (h *siteRouterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	desc, ok := h.lookupSite(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	h.logger.Debug("resolved site", zap.String("site", desc.Site))

	h.next.ServeHTTP(w, r.WithContext(withSite(r.Context(), desc)))
}

func checkSiteFS(fsys fs.FS) error {
	f, err := fsys.Open(".")
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("not directory: %s", fi.Name())
	}
	return nil
}
