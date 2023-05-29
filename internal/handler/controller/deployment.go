package controller

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
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

func (c *Controller) handleDeploymentGet(ctx *gin.Context) {
	appID := ctx.Param("app-id")
	deploymentName := ctx.Param("deployment-name")

	if !c.requireAuthn(ctx) || !c.requireAuthz(ctx, authzReadApp(appID)) {
		return
	}

	deployment, err := tx(ctx, c.DB, func(conn db.Conn) (*apiDeployment, error) {
		app, err := conn.GetApp(ctx, appID)
		if err != nil {
			return nil, err
		}

		deployment, err := conn.GetDeploymentByName(ctx, appID, deploymentName)
		if err != nil {
			return nil, err
		}

		sites, err := conn.GetDeploymentSiteNames(ctx, deployment)
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

	writeResponse(ctx, deployment, err)
}

func (c *Controller) handleDeploymentCreate(ctx *gin.Context) {
	appID := ctx.Param("app-id")

	if !c.requireAuthn(ctx) || !c.requireAuthz(ctx, authzWriteApp(appID)) {
		return
	}

	var request struct {
		Name       string             `json:"name" binding:"required,dnsLabel"`
		Files      []models.FileEntry `json:"files" binding:"required"`
		SiteConfig *config.SiteConfig `json:"site_config" binding:"required"`
	}
	if err := checkBind(ctx, ctx.ShouldBindJSON(&request)); err != nil {
		return
	}
	name := request.Name
	files := request.Files
	siteConfig := request.SiteConfig

	if len(files) > models.MaxFiles {
		ctx.JSON(http.StatusBadRequest, response{Error: deploy.ErrTooManyFiles})
		return
	}

	if err := config.ValidateSiteConfig(siteConfig); err != nil {
		ctx.JSON(http.StatusBadRequest, response{Error: err})
		return
	}

	var totalSize int64 = 0
	for _, entry := range files {
		totalSize += entry.Size
	}
	if totalSize > c.Config.MaxDeploymentSize {
		ctx.JSON(http.StatusBadRequest, response{
			Error: fmt.Errorf(
				"deployment too large: %s > %s",
				humanize.Bytes(uint64(totalSize)),
				humanize.Bytes(uint64(c.Config.MaxDeploymentSize)),
			),
		})
		return
	}

	deployment, err := tx(ctx, c.DB, func(conn db.Conn) (*apiDeployment, error) {
		app, err := conn.GetApp(ctx, appID)
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

		err = conn.CreateDeployment(ctx, deployment)
		if err != nil {
			return nil, err
		}

		return c.makeAPIDeployment(app, db.DeploymentInfo{
			Deployment:    deployment,
			FirstSiteName: nil,
		}), nil
	})

	writeResponse(ctx, deployment, err)
}

func (c *Controller) handleDeploymentUpload(ctx *gin.Context) {
	appID := ctx.Param("app-id")
	deploymentName := ctx.Param("deployment-name")

	if !c.requireAuthn(ctx) || !c.requireAuthz(ctx, authzWriteApp(appID)) {
		return
	}

	deployment, err := tx(ctx, c.DB, func(conn db.Conn) (*models.Deployment, error) {
		deployment, err := conn.GetDeploymentByName(ctx, appID, deploymentName)
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
		writeResponse(ctx, nil, err)
		return
	}

	// Extract tarball to object stoarge

	handleFile := func(e models.FileEntry, r io.Reader) error {
		key := deployment.StorageKeyPrefix + e.Path
		return c.Storage.Upload(ctx, key, r)
	}

	reader := io.LimitReader(
		httputil.NewTimeoutReader(
			ctx.Request.Body,
			http.NewResponseController(ctx.Writer),
			10*time.Second,
		),
		c.Config.MaxDeploymentSize,
	)
	err = deploy.ExtractFiles(reader, deployment.Metadata.Files, handleFile)
	if errors.As(err, new(deploy.Error)) {
		ctx.JSON(http.StatusBadRequest, response{Error: err})
		return
	} else if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	now := c.Clock.Now().UTC()

	// Mark deployment as completed, but inactive
	result, err := tx(ctx, c.DB, func(conn db.Conn) (*apiDeployment, error) {
		app, err := conn.GetApp(ctx, appID)
		if err != nil {
			return nil, err
		}

		deployment, err = conn.GetDeployment(ctx, appID, deployment.ID)
		if err != nil {
			return nil, err
		}

		if deployment.UploadedAt != nil {
			return nil, models.ErrDeploymentAlreadyUploaded
		}

		err = conn.MarkDeploymentUploaded(ctx, now, deployment)
		if err != nil {
			return nil, err
		}

		return c.makeAPIDeployment(app, db.DeploymentInfo{
			Deployment:    deployment,
			FirstSiteName: nil,
		}), nil
	})

	writeResponse(ctx, result, err)
}

func (c *Controller) handleDeploymentList(ctx *gin.Context) {
	appID := ctx.Param("app-id")

	if !c.requireAuthn(ctx) || !c.requireAuthz(ctx, authzReadApp(appID)) {
		return
	}

	deployments, err := tx(ctx, c.DB, func(conn db.Conn) ([]*apiDeployment, error) {
		app, err := conn.GetApp(ctx, appID)
		if err != nil {
			return nil, err
		}

		deployments, err := conn.ListDeployments(ctx, appID)
		if err != nil {
			return nil, err
		}

		return mapModels(deployments, func(d db.DeploymentInfo) *apiDeployment {
			return c.makeAPIDeployment(app, d)
		}), nil
	})

	writeResponse(ctx, deployments, err)
}
