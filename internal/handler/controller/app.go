package controller

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/oursky/pageship/internal/db"
	"github.com/oursky/pageship/internal/models"
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

	userID := authn(r).UserID
	app, err := tx(r.Context(), c.DB, func(conn db.Conn) (*apiApp, error) {
		app := models.NewApp(c.Clock.Now().UTC(), request.ID)

		err := conn.CreateApp(r.Context(), app)
		if err != nil {
			return nil, err
		}

		err = conn.AssignAppUser(r.Context(), app.ID, userID)
		if err != nil {
			return nil, err
		}

		return c.makeAPIApp(app), nil
	})

	writeResponse(w, app, err)
}

func (c *Controller) handleAppGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "app-id")

	app, err := tx(r.Context(), c.DB, func(conn db.Conn) (*apiApp, error) {
		app, err := conn.GetApp(r.Context(), id)
		if err != nil {
			return nil, err
		}

		return c.makeAPIApp(app), nil
	})

	writeResponse(w, app, err)
}

func (c *Controller) handleAppList(w http.ResponseWriter, r *http.Request) {
	userID := authn(r).UserID
	apps, err := tx(r.Context(), c.DB, func(conn db.Conn) ([]*apiApp, error) {
		apps, err := conn.ListApps(r.Context(), userID)
		if err != nil {
			return nil, err
		}

		return mapModels(apps, c.makeAPIApp), nil
	})

	writeResponse(w, apps, err)
}

func (c *Controller) handleAppUserList(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "app-id")

	users, err := tx(r.Context(), c.DB, func(conn db.Conn) ([]*apiUser, error) {
		users, err := conn.ListAppUsers(r.Context(), id)
		if err != nil {
			return nil, err
		}

		return mapModels(users, c.makeAPIUser), nil
	})

	writeResponse(w, users, err)
}

func (c *Controller) handleAppUserAdd(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "app-id")

	var request struct {
		UserID string `json:"userID" binding:"required"`
	}
	if !bindJSON(w, r, &request) {
		return
	}

	result, err := tx(r.Context(), c.DB, func(conn db.Conn) (*struct{}, error) {
		user, err := conn.GetUser(r.Context(), request.UserID)
		if err != nil {
			return nil, err
		}

		err = conn.AssignAppUser(r.Context(), appID, user.ID)
		if err != nil {
			return nil, err
		}

		return &struct{}{}, nil
	})

	writeResponse(w, result, err)
}

func (c *Controller) handleAppUserDelete(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "app-id")
	userID := chi.URLParam(r, "user-id")

	if userID == authn(r).UserID {
		writeResponse(w, nil, models.ErrDeleteCurrentUser)
		return
	}

	result, err := tx(r.Context(), c.DB, func(conn db.Conn) (*struct{}, error) {
		err := conn.UnassignAppUser(r.Context(), appID, userID)
		if err != nil {
			return nil, err
		}

		return &struct{}{}, nil
	})

	writeResponse(w, result, err)
}
