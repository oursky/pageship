package controller

import (
	"net/http"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/storage"
	apptime "github.com/oursky/pageship/internal/time"
	"go.uber.org/zap"
)

type Controller struct {
	Clock   apptime.Clock
	Config  Config
	Storage *storage.Storage
	DB      db.DB
}

func (c *Controller) Handler() http.Handler {
	if c.Clock == nil {
		c.Clock = apptime.SystemClock
	}

	logger := zap.L().Named("controller")

	g := gin.New()
	g.Use(ginzap.Ginzap(zap.L(), time.RFC3339, true))
	g.Use(ginzap.RecoveryWithZap(logger, true))

	g.GET("/healthz", c.handleHealthz)

	v1 := g.Group("/api/v1")
	v1.POST("/apps", c.handleAppCreate)
	v1.GET("/apps", c.handleAppList)
	v1.GET("/apps/:app-id", c.handleAppGet)
	v1.GET("/apps/:app-id/config", c.handleAppConfigGet)
	v1.PUT("/apps/:app-id/config", c.handleAppConfigSet)

	v1.GET("/apps/:app-id/sites", c.handleSiteList)
	v1.POST("/apps/:app-id/sites/:site-name", c.handleSiteCreate)
	v1.PATCH("/apps/:app-id/sites/:site-name", c.handleSiteUpdate)

	v1.POST("/apps/:app-id/deployments", c.handleDeploymentCreate)
	v1.GET("/apps/:app-id/deployments", c.handleDeploymentList)
	v1.GET("/apps/:app-id/deployments/:deployment-name", c.handleDeploymentGet)
	v1.PUT("/apps/:app-id/deployments/:deployment-name/tarball", c.handleDeploymentUpload)

	return g.Handler()
}
