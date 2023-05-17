package controller

import "github.com/gin-gonic/gin"

func (*Controller) handleHealthz(ctx *gin.Context) {
	ctx.String(200, "OK")
}
