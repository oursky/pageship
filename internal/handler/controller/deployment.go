package controller

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/deploy"
	"github.com/oursky/pageship/internal/models"
)

type apiDeployment struct {
	*models.Deployment
	Sites []string `json:"sites"`
}

func (c *Controller) makeAPIDeployment(d *models.Deployment) *apiDeployment {
	deployment := *d
	deployment.Metadata.Files = nil // Avoid large file list

	return &apiDeployment{Deployment: d}
}

func (c *Controller) handleDeploymentGet(ctx *gin.Context) {
	appID := ctx.Param("app-id")
	deploymentName := ctx.Param("deployment-name")

	if !c.requireAuthn(ctx) || !c.requireAuthz(ctx, authzReadApp(appID)) {
		return
	}

	deployment, err := tx(ctx, c.DB, func(conn db.Conn) (*apiDeployment, error) {
		deployment, err := conn.GetDeployment(ctx, appID, deploymentName)
		if err != nil {
			return nil, err
		}

		return c.makeAPIDeployment(deployment), nil
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
		_, err := conn.GetApp(ctx, appID)
		if err != nil {
			return nil, err
		}

		metadata := &models.DeploymentMetadata{
			Files:  files,
			Config: *siteConfig,
		}
		deployment := models.NewDeployment(c.Clock.Now().UTC(), name, appID, c.Config.StorageKeyPrefix, metadata)

		err = conn.CreateDeployment(ctx, deployment)
		if err != nil {
			return nil, err
		}

		return c.makeAPIDeployment(deployment), nil
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

	err = deploy.ExtractFiles(ctx.Request.Body, deployment.Metadata.Files, handleFile)
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

		return c.makeAPIDeployment(deployment), nil
	})

	writeResponse(ctx, result, err)
}

func (c *Controller) handleDeploymentList(ctx *gin.Context) {
	appID := ctx.Param("app-id")

	if !c.requireAuthn(ctx) || !c.requireAuthz(ctx, authzReadApp(appID)) {
		return
	}

	deployments, err := tx(ctx, c.DB, func(conn db.Conn) ([]*apiDeployment, error) {
		deployments, err := conn.ListDeployments(ctx, appID)
		if err != nil {
			return nil, err
		}

		return mapModels(deployments, c.makeAPIDeployment), nil
	})

	writeResponse(ctx, deployments, err)
}
