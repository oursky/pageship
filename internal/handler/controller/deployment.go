package controller

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi/v5"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/deploy"
	"github.com/oursky/pageship/internal/httputil"
	"github.com/oursky/pageship/internal/models"
)

type apiDeployment struct {
	*models.Deployment
	FirstSiteName *string `json:"siteName"`
	URL           string  `json:"url,omitempty"`
}

func (c *Controller) makeAPIDeployment(app *models.App, d db.DeploymentInfo) *apiDeployment {
	deployment := *d.Deployment
	deployment.Metadata.Files = nil // Avoid large file list

	siteName := ""
	if d.FirstSiteName != nil {
		siteName = *d.FirstSiteName
	} else if app.Config.Deployments.Accessible {
		siteName = d.Name
	}

	url := ""
	if d.CheckAlive(c.Clock.Now().UTC()) == nil && siteName != "" {
		sub := siteName
		if siteName == app.Config.DefaultSite {
			sub = ""
		}
		url = c.Config.HostPattern.MakeURL(
			c.Config.HostIDScheme.Make(app.ID, sub),
		)
	}

	return &apiDeployment{
		Deployment:    &deployment,
		FirstSiteName: d.FirstSiteName,
		URL:           url,
	}
}

func (c *Controller) handleDeploymentGet(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "app-id")
	deploymentName := chi.URLParam(r, "deployment-name")

	deployment, err := tx(r.Context(), c.DB, func(conn db.Conn) (*apiDeployment, error) {
		app, err := conn.GetApp(r.Context(), appID)
		if err != nil {
			return nil, err
		}

		deployment, err := conn.GetDeploymentByName(r.Context(), appID, deploymentName)
		if err != nil {
			return nil, err
		}

		sites, err := conn.GetDeploymentSiteNames(r.Context(), deployment)
		if err != nil {
			return nil, err
		}
		var siteName *string
		if len(sites) > 0 {
			siteName = &sites[0]
		}

		return c.makeAPIDeployment(app, db.DeploymentInfo{
			Deployment:    deployment,
			FirstSiteName: siteName,
		}), nil
	})

	writeResponse(w, deployment, err)
}

func (c *Controller) handleDeploymentCreate(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "app-id")

	var request struct {
		Name       string             `json:"name" binding:"required,dnsLabel"`
		Files      []models.FileEntry `json:"files" binding:"required"`
		SiteConfig *config.SiteConfig `json:"site_config" binding:"required"`
	}
	if !bindJSON(w, r, &request) {
		return
	}
	name := request.Name
	files := request.Files
	siteConfig := request.SiteConfig

	if len(files) > models.MaxFiles {
		writeJSON(w, http.StatusBadRequest, response{Error: deploy.ErrTooManyFiles})
		return
	}

	if err := config.ValidateSiteConfig(siteConfig); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: err})
		return
	}

	var totalSize int64 = 0
	for _, entry := range files {
		totalSize += entry.Size
	}
	if totalSize > c.Config.MaxDeploymentSize {
		writeJSON(w, http.StatusBadRequest, response{
			Error: fmt.Errorf(
				"deployment too large: %s > %s",
				humanize.Bytes(uint64(totalSize)),
				humanize.Bytes(uint64(c.Config.MaxDeploymentSize)),
			),
		})
		return
	}

	deployment, err := tx(r.Context(), c.DB, func(conn db.Conn) (*apiDeployment, error) {
		app, err := conn.GetApp(r.Context(), appID)
		if err != nil {
			return nil, err
		}

		now := c.Clock.Now().UTC()

		metadata := &models.DeploymentMetadata{
			Files:  files,
			Config: *siteConfig,
		}
		deployment := models.NewDeployment(now, name, appID, c.Config.StorageKeyPrefix, metadata)

		deploymentTTL, err := time.ParseDuration(app.Config.Deployments.TTL)
		if err != nil {
			return nil, err
		}
		expireAt := now.Add(deploymentTTL)
		deployment.ExpireAt = &expireAt

		err = conn.CreateDeployment(r.Context(), deployment)
		if err != nil {
			return nil, err
		}

		return c.makeAPIDeployment(app, db.DeploymentInfo{
			Deployment:    deployment,
			FirstSiteName: nil,
		}), nil
	})

	writeResponse(w, deployment, err)
}

func (c *Controller) handleDeploymentUpload(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "app-id")
	deploymentName := chi.URLParam(r, "deployment-name")

	deployment, err := tx(r.Context(), c.DB, func(conn db.Conn) (*models.Deployment, error) {
		deployment, err := conn.GetDeploymentByName(r.Context(), appID, deploymentName)
		if err != nil {
			return nil, err
		}

		if deployment.IsExpired(c.Clock.Now().UTC()) {
			return nil, models.ErrDeploymentExpired
		}
		if deployment.UploadedAt != nil {
			return nil, models.ErrDeploymentAlreadyUploaded
		}

		return deployment, nil
	})

	if err != nil {
		writeResponse(w, nil, err)
		return
	}

	// Extract tarball to object stoarge

	handleFile := func(e models.FileEntry, reader io.Reader) error {
		key := deployment.StorageKeyPrefix + e.Path
		return c.Storage.Upload(r.Context(), key, reader)
	}

	reader := io.LimitReader(
		httputil.NewTimeoutReader(
			r.Body,
			http.NewResponseController(w),
			10*time.Second,
		),
		c.Config.MaxDeploymentSize,
	)
	err = deploy.ExtractFiles(reader, deployment.Metadata.Files, handleFile)
	if errors.As(err, new(deploy.Error)) {
		writeJSON(w, http.StatusBadRequest, response{Error: err})
		return
	} else if err != nil {
		writeResponse(w, nil, err)
		return
	}

	now := c.Clock.Now().UTC()

	// Mark deployment as completed, but inactive
	result, err := tx(r.Context(), c.DB, func(conn db.Conn) (*apiDeployment, error) {
		app, err := conn.GetApp(r.Context(), appID)
		if err != nil {
			return nil, err
		}

		deployment, err = conn.GetDeployment(r.Context(), appID, deployment.ID)
		if err != nil {
			return nil, err
		}

		if deployment.UploadedAt != nil {
			return nil, models.ErrDeploymentAlreadyUploaded
		}

		err = conn.MarkDeploymentUploaded(r.Context(), now, deployment)
		if err != nil {
			return nil, err
		}

		return c.makeAPIDeployment(app, db.DeploymentInfo{
			Deployment:    deployment,
			FirstSiteName: nil,
		}), nil
	})

	writeResponse(w, result, err)
}

func (c *Controller) handleDeploymentList(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "app-id")

	deployments, err := tx(r.Context(), c.DB, func(conn db.Conn) ([]*apiDeployment, error) {
		app, err := conn.GetApp(r.Context(), appID)
		if err != nil {
			return nil, err
		}

		deployments, err := conn.ListDeployments(r.Context(), appID)
		if err != nil {
			return nil, err
		}

		return mapModels(deployments, func(d db.DeploymentInfo) *apiDeployment {
			return c.makeAPIDeployment(app, d)
		}), nil
	})

	writeResponse(w, deployments, err)
}
