package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
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

func bindJSON(ctx *gin.Context, value any) error {
	if err := ctx.ShouldBindJSON(value); err != nil {
		ctx.JSON(http.StatusBadRequest, response{Error: err})
		return err
	}
	return nil
}
