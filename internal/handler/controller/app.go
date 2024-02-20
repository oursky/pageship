package controller

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/oursky/pageship/internal/config"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
	"go.uber.org/zap"
)

type apiApp struct {
	*models.App
	URL string `json:"url"`
}

func (c *Controller) makeAPIApp(app *models.App) *apiApp {
	return &apiApp{
		App: app,
		URL: c.Config.HostPattern.MakeURL(
			c.Config.HostIDScheme.Make(app.ID, ""),
		),
	}
}

func (c *Controller) middlewareLoadApp() func(http.Handler) http.Handler {
	return middlewareLoadValue(func(r *http.Request) (*models.App, error) {
		id := chi.URLParam(r, "app-id")

		app, err := c.DB.GetApp(r.Context(), id)
		if err != nil {
			return nil, err
		}

		return app, nil
	})
}

func (c *Controller) handleAppCreate(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ID string `json:"id" binding:"required,dnsLabel"`
	}
	if !bindJSON(w, r, &request) {
		return
	}

	if _, reserved := c.Config.ReservedApps[request.ID]; reserved {
		writeResponse(w, nil, models.ErrAppUsedID)
		return
	}

	userID := getSubject(r)
	respond(w, withTx(r.Context(), c.DB, func(tx db.Tx) (any, error) {
		app := models.NewApp(c.Clock.Now().UTC(), request.ID, userID)

		err := tx.CreateApp(r.Context(), app)
		if err != nil {
			return nil, err
		}

		log(r).Info("creating app", zap.String("app", app.ID))

		return c.makeAPIApp(app), nil
	}))
}

func (c *Controller) handleAppGet(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, c.makeAPIApp(get[*models.App](r)), nil)
}

func (c *Controller) handleAppList(w http.ResponseWriter, r *http.Request) {
	authn := get[*authnInfo](r)
	respond(w, func() (any, error) {
		apps, err := c.DB.ListApps(r.Context(), authn.CredentialIDs)
		if err != nil {
			return nil, err
		}

		n := 0
		for _, a := range apps {
			if _, err := a.CheckAuthz(config.AccessLevelReader, authn.UserID(), authn.CredentialIDs); err == nil {
				apps[n] = a
				n++
			}
		}
		apps = apps[:n]

		return mapModels(apps, c.makeAPIApp), nil
	})
}
