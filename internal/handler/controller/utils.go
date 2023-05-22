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
