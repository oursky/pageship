package controller

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
)

type authzAction string

const (
	authzReadApp  authzAction = "read-app"
	authzWriteApp authzAction = "write-app"
)

func (c *Controller) requireAuth(
	actions ...authzAction,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			info := authn(r)
			if info == nil {
				writeResponse(w, nil, models.ErrInvalidCredentials)
				return
			}

			err := db.WithTx(r.Context(), c.DB, func(c db.Conn) error {
				for _, a := range actions {
					switch a {
					case authzReadApp, authzWriteApp:
						appID := chi.URLParam(r, "app-id")
						if err := c.IsAppAccessible(r.Context(), appID, info.UserID); err != nil {
							return err
						}
					default:
						panic("unknown authz action: " + a)
					}
				}
				return nil
			})
			if err != nil {
				writeResponse(w, nil, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
