package controller

import (
	"net/http"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Controller struct {
}

func (c *Controller) Handler() http.Handler {
	logger := zap.L().Named("controller")

	g := gin.New()
	g.Use(ginzap.Ginzap(zap.L(), time.RFC3339, true))
	g.Use(ginzap.RecoveryWithZap(logger, true))

	g.GET("/healthz", c.handleHealthz)

	return g.Handler()
}
