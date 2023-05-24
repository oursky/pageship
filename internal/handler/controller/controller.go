package controller

import (
	"net/http"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/storage"
	apptime "github.com/oursky/pageship/internal/time"
	"go.uber.org/zap"
)

func init() {
	validate := binding.Validator.Engine().(*validator.Validate)

	binding.EnableDecoderDisallowUnknownFields = true

	validate.RegisterValidation("dnsLabel", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return config.ValidateDNSLabel(value)
	})
}

type Controller struct {
	Logger  *zap.Logger
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
	g.Use(ginzap.GinzapWithConfig(zap.L(), &ginzap.Config{
		TimeFormat: time.RFC3339,
		UTC:        true,
		Context: func(c *gin.Context) []zap.Field {
			var fields []zap.Field
			if userID, ok := c.Get("userID"); ok {
				fields = append(fields, zap.String("user-id", userID.(string)))
			}
			return fields
		},
	}))
	g.Use(ginzap.RecoveryWithZap(logger, true))

	g.GET("/healthz", c.handleHealthz)

	v1 := g.Group("/api/v1")
	v1.POST("/apps", c.handleAppCreate)
	v1.GET("/apps", c.handleAppList)
	v1.GET("/apps/:app-id", c.handleAppGet)
	v1.GET("/apps/:app-id/config", c.handleAppConfigGet)
	v1.PUT("/apps/:app-id/config", c.handleAppConfigSet)
	v1.GET("/apps/:app-id/users", c.handleAppUserList)
	v1.POST("/apps/:app-id/users", c.handleAppUserAdd)
	v1.DELETE("/apps/:app-id/users/:user-id", c.handleAppUserDelete)

	v1.GET("/apps/:app-id/sites", c.handleSiteList)
	v1.POST("/apps/:app-id/sites", c.handleSiteCreate)
	v1.PATCH("/apps/:app-id/sites/:site-name", c.handleSiteUpdate)

	v1.POST("/apps/:app-id/deployments", c.handleDeploymentCreate)
	v1.GET("/apps/:app-id/deployments", c.handleDeploymentList)
	v1.GET("/apps/:app-id/deployments/:deployment-name", c.handleDeploymentGet)
	v1.PUT("/apps/:app-id/deployments/:deployment-name/tarball", c.handleDeploymentUpload)

	v1.GET("/auth/github-ssh", c.handleAuthGithubSSH)

	return g.Handler()
}
