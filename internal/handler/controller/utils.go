package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

type response struct {
	Error  error
	Result any
}

func (r response) MarshalJSON() ([]byte, error) {
	if r.Error != nil {
		return json.Marshal(map[string]any{"error": r.Error.Error()})
	} else {
		return json.Marshal(map[string]any{"result": r.Result})
	}
}

func writeResponse(ctx *gin.Context, result any, err error) {
	if err == nil {
		ctx.JSON(http.StatusOK, response{Result: result})
		return
	}

	switch {
	case errors.Is(err, models.ErrAppUsedID):
		ctx.JSON(http.StatusConflict, response{Error: err})
	case errors.Is(err, models.ErrAppNotFound):
		ctx.JSON(http.StatusNotFound, response{Error: err})
	case errors.Is(err, models.ErrUndefinedSite):
		ctx.JSON(http.StatusBadRequest, response{Error: err})
	case errors.Is(err, models.ErrSiteNotFound):
		ctx.JSON(http.StatusNotFound, response{Error: err})
	case errors.Is(err, models.ErrDeploymentNotFound):
		ctx.JSON(http.StatusNotFound, response{Error: err})
	case errors.Is(err, models.ErrDeploymentUsedName):
		ctx.JSON(http.StatusConflict, response{Error: err})
	case errors.Is(err, models.ErrDeploymentNotUploaded):
		ctx.JSON(http.StatusBadRequest, response{Error: err})
	case errors.Is(err, models.ErrDeploymentAlreadyUploaded):
		ctx.JSON(http.StatusBadRequest, response{Error: err})
	case errors.Is(err, models.ErrUserNotFound):
		ctx.JSON(http.StatusNotFound, response{Error: err})
	case errors.Is(err, models.ErrDeleteCurrentUser):
		ctx.JSON(http.StatusBadRequest, response{Error: err})
	case errors.Is(err, models.ErrInvalidCredentials):
		ctx.JSON(http.StatusUnauthorized, response{Error: err})
	default:
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}
}

func checkBind(ctx *gin.Context, err error) error {
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response{Error: err})
	}
	return err
}

func mapModels[T, U any](models []T, mapper func(m T) U) []U {
	result := make([]U, len(models))
	for i, m := range models {
		result[i] = mapper(m)
	}
	return result
}

func tx[T any](ctx context.Context, d db.DB, fn func(c db.Conn) (T, error)) (T, error) {
	var result T
	err := db.WithTx(ctx, d, func(c db.Conn) (err error) {
		result, err = fn(c)
		return
	})
	return result, err
}
