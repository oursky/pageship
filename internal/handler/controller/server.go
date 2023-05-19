package controller

import (
	"net/http"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/oursky/pageship/internal/db"
	apptime "github.com/oursky/pageship/internal/time"
	"go.uber.org/zap"
)

func init() {
	// Use vanilla tag name for interop.
	binding.Validator.Engine().(*validator.Validate).SetTagName("validate")
}

type Controller struct {
	Clock  apptime.Clock
	Config Config
	DB     db.DB
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

	v1.POST("/apps/:app-id/sites/:site/deployments", c.handleDeploymentCreate)
	v1.PUT("/apps/:app-id/sites/:site/deployments/:deployment-id/tarball", c.handleDeploymentUpload)

	return g.Handler()
}
