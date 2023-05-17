package controller

import (
	"net/http"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/oursky/pageship/internal/db"
	"go.uber.org/zap"
)

type Controller struct {
	DB db.DB
}

func (c *Controller) Handler() http.Handler {
	logger := zap.L().Named("controller")

	g := gin.New()
	g.Use(ginzap.Ginzap(zap.L(), time.RFC3339, true))
	g.Use(ginzap.RecoveryWithZap(logger, true))

	g.GET("/healthz", c.handleHealthz)

	v1 := g.Group("/api/v1")
	v1.POST("/apps", c.handleAppCreate)
	v1.GET("/apps", c.handleAppList)

	return g.Handler()
}
