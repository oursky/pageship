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

func (c *Controller) handleDeploymentGet(ctx *gin.Context) {
	appID := ctx.Param("app-id")
	siteName := ctx.Param("site")
	deploymentID := ctx.Param("deployment-id")

	err := db.WithTx(ctx, c.DB, func(conn db.Conn) error {
		deployment, err := conn.GetDeployment(ctx, appID, siteName, deploymentID)
		if err != nil {
			return err
		}

		ctx.JSON(http.StatusOK, response{Result: deployment})
		return nil
	})

	if err != nil {
		if errors.Is(err, models.ErrDeploymentNotFound) {
			ctx.JSON(http.StatusNotFound, response{Error: err})
		} else {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}
	}
}

func (c *Controller) handleDeploymentCreate(ctx *gin.Context) {
	appID := ctx.Param("app-id")
	siteName := ctx.Param("site")

	var request struct {
		Files      []models.FileEntry `json:"files" binding:"required"`
		SiteConfig *config.SiteConfig `json:"site_config" binding:"required"`
	}
	if err := checkBind(ctx, ctx.ShouldBindJSON(&request)); err != nil {
		return
	}
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

	err := db.WithTx(ctx, c.DB, func(conn db.Conn) error {
		app, err := conn.GetApp(ctx, appID)
		if errors.Is(err, models.ErrAppNotFound) {
			ctx.JSON(http.StatusNotFound, response{Error: err})
			return db.ErrRollback
		} else if err != nil {
			return err
		}

		_, ok := app.Config.ResolveSite(siteName)
		if !ok {
			ctx.JSON(http.StatusBadRequest, response{Error: models.ErrUndefinedSite})
			return db.ErrRollback
		}

		site := models.NewSite(c.Clock.Now().UTC(), appID, siteName)
		err = conn.CreateSiteIfNotExist(ctx, site)
		if err != nil {
			return err
		}

		metadata := &models.DeploymentMetadata{
			Files:  files,
			Config: *siteConfig,
		}
		deployment := models.NewDeployment(c.Clock.Now().UTC(), appID, site.ID, c.Config.StorageKeyPrefix, metadata)

		err = conn.CreateDeployment(ctx, deployment)
		if err != nil {
			return err
		}

		ctx.JSON(http.StatusOK, response{Result: deployment})
		return nil
	})

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}
}

func (c *Controller) handleDeploymentUpload(ctx *gin.Context) {
	appID := ctx.Param("app-id")
	siteName := ctx.Param("site")
	deploymentID := ctx.Param("deployment-id")

	var deployment *models.Deployment
	err := db.WithTx(ctx, c.DB, func(conn db.Conn) (err error) {
		deployment, err = conn.GetDeployment(ctx, appID, siteName, deploymentID)
		if err != nil {
			return
		}

		if deployment.Status != models.DeploymentStatusPending {
			err = models.ErrDeploymentInvalidStatus
			return
		}

		return
	})

	if err != nil {
		if errors.Is(err, models.ErrDeploymentNotFound) {
			ctx.JSON(http.StatusNotFound, response{Error: err})
		} else if errors.Is(err, models.ErrDeploymentInvalidStatus) {
			ctx.JSON(http.StatusConflict, response{Error: err})
		} else {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	if deployment.Status != models.DeploymentStatusPending {
		ctx.JSON(http.StatusConflict, response{Error: models.ErrDeploymentInvalidStatus})
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
	err = db.WithTx(ctx, c.DB, func(conn db.Conn) error {
		deployment, err = conn.GetDeployment(ctx, appID, siteName, deploymentID)
		if err != nil {
			return err
		}

		if deployment.Status != models.DeploymentStatusPending {
			return models.ErrDeploymentInvalidStatus
		}

		err = conn.MarkDeploymentUploaded(ctx, now, deployment)
		if err != nil {
			return err
		}

		ctx.JSON(http.StatusOK, response{Result: deployment})
		return nil
	})

	if err != nil {
		if errors.Is(err, models.ErrDeploymentNotFound) {
			ctx.JSON(http.StatusNotFound, response{Error: err})
		} else if errors.Is(err, models.ErrDeploymentInvalidStatus) {
			ctx.JSON(http.StatusConflict, response{Error: err})
		} else {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}
	}
}

func (c *Controller) handleDeploymentUpdate(ctx *gin.Context) {
	appID := ctx.Param("app-id")
	siteName := ctx.Param("site")
	deploymentID := ctx.Param("deployment-id")

	var request struct {
		Status *models.DeploymentStatus `json:"status,omitempty" binding:"omitempty,oneof=ACTIVE INACTIVE"`
	}
	if err := checkBind(ctx, ctx.ShouldBindJSON(&request)); err != nil {
		return
	}

	if request.Status != nil && !request.Status.IsValid() {
		ctx.JSON(http.StatusBadRequest, response{Result: errors.New("invalid status")})
		return
	}

	err := db.WithTx(ctx, c.DB, func(conn db.Conn) error {
		deployment, err := conn.GetDeployment(ctx, appID, siteName, deploymentID)
		if err != nil {
			return err
		}

		if request.Status != nil {
			switch deployment.Status {
			case models.DeploymentStatusActive, models.DeploymentStatusInactive:
				break
			default:
				return models.ErrDeploymentInvalidStatus
			}

			switch *request.Status {
			case models.DeploymentStatusActive:
				err = conn.ActivateSiteDeployment(ctx, deployment)
				if err != nil {
					return err
				}
			case models.DeploymentStatusInactive:
				err = conn.DeactivateSiteDeployment(ctx, deployment)
				if err != nil {
					return err
				}
			default:
				return models.ErrDeploymentInvalidStatus
			}
		}

		ctx.JSON(http.StatusOK, response{Result: deployment})
		return nil
	})

	if err != nil {
		if errors.Is(err, models.ErrDeploymentNotFound) {
			ctx.JSON(http.StatusNotFound, response{Error: err})
		} else if errors.Is(err, models.ErrDeploymentInvalidStatus) {
			ctx.JSON(http.StatusConflict, response{Error: err})
		} else {
			ctx.AbortWithError(http.StatusInternalServerError, err)
		}
	}
}
