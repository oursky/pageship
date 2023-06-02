package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.SetTagName("binding")

	validate.RegisterValidation("dnsLabel", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return config.ValidateDNSLabel(value)
	})
}

const maxJSONSize = 10 * 1024 * 1024 // 10MB

func bindJSON[T any](w http.ResponseWriter, r *http.Request, body *T) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxJSONSize))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(body); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: err})
		return false
	}

	if err := validate.Struct(body); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: err})
		return false
	}
	return true
}

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

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(value); err != nil {
		panic(err)
	}
}

func writeResponse(w http.ResponseWriter, result any, err error) {
	if err == nil {
		writeJSON(w, http.StatusOK, response{Result: result})
		return
	}

	switch {
	case errors.Is(err, models.ErrAppUsedID):
		writeJSON(w, http.StatusConflict, response{Error: err})
	case errors.Is(err, models.ErrAppNotFound):
		writeJSON(w, http.StatusNotFound, response{Error: err})
	case errors.Is(err, models.ErrUndefinedSite):
		writeJSON(w, http.StatusBadRequest, response{Error: err})
	case errors.Is(err, models.ErrSiteNotFound):
		writeJSON(w, http.StatusNotFound, response{Error: err})
	case errors.Is(err, models.ErrDeploymentNotFound):
		writeJSON(w, http.StatusNotFound, response{Error: err})
	case errors.Is(err, models.ErrDeploymentUsedName):
		writeJSON(w, http.StatusConflict, response{Error: err})
	case errors.Is(err, models.ErrDeploymentNotUploaded):
		writeJSON(w, http.StatusBadRequest, response{Error: err})
	case errors.Is(err, models.ErrDeploymentAlreadyUploaded):
		writeJSON(w, http.StatusBadRequest, response{Error: err})
	case errors.Is(err, models.ErrDeploymentExpired):
		writeJSON(w, http.StatusBadRequest, response{Error: err})
	case errors.Is(err, models.ErrUserNotFound):
		writeJSON(w, http.StatusNotFound, response{Error: err})
	case errors.Is(err, models.ErrDeleteCurrentUser):
		writeJSON(w, http.StatusBadRequest, response{Error: err})
	case errors.Is(err, models.ErrInvalidCredentials):
		writeJSON(w, http.StatusUnauthorized, response{Error: err})
	default:
		panic(err)
	}
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

func requestID(r *http.Request) string {
	return middleware.GetReqID(r.Context())
}
