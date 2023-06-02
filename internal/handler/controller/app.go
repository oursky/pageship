package controller

import (
	"net/http"

	"github.com/go-chi/chi/v5"
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
	respond(w, withTx(r.Context(), c.DB, func(tx db.Tx) (any, error) {
		app := models.NewApp(c.Clock.Now().UTC(), request.ID)

		err := tx.CreateApp(r.Context(), app)
		if err != nil {
			return nil, err
		}

		c.Logger.Info("creating app",
			zap.String("request_id", requestID(r)),
			zap.String("user", authn(r).UserID),
			zap.String("app", app.ID),
		)

		err = tx.AssignAppUser(r.Context(), app.ID, userID)
		if err != nil {
			return nil, err
		}

		return c.makeAPIApp(app), nil
	}))
}

func (c *Controller) handleAppGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "app-id")

	respond(w, func() (any, error) {
		app, err := c.DB.GetApp(r.Context(), id)
		if err != nil {
			return nil, err
		}

		return c.makeAPIApp(app), nil
	})
}

func (c *Controller) handleAppList(w http.ResponseWriter, r *http.Request) {
	userID := authn(r).UserID
	respond(w, func() (any, error) {
		apps, err := c.DB.ListApps(r.Context(), userID)
		if err != nil {
			return nil, err
		}

		return mapModels(apps, c.makeAPIApp), nil
	})
}

func (c *Controller) handleAppUserList(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "app-id")

	respond(w, func() (any, error) {
		users, err := c.DB.ListAppUsers(r.Context(), id)
		if err != nil {
			return nil, err
		}

		return mapModels(users, c.makeAPIUser), nil
	})
}

func (c *Controller) handleAppUserAdd(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "app-id")

	var request struct {
		UserID string `json:"userID" binding:"required"`
	}
	if !bindJSON(w, r, &request) {
		return
	}

	respond(w, withTx(r.Context(), c.DB, func(tx db.Tx) (any, error) {
		user, err := tx.GetUser(r.Context(), request.UserID)
		if err != nil {
			return nil, err
		}

		err = tx.AssignAppUser(r.Context(), appID, user.ID)
		if err != nil {
			return nil, err
		}

		c.Logger.Info("adding user",
			zap.String("request_id", requestID(r)),
			zap.String("user", authn(r).UserID),
			zap.String("target_user", user.ID),
			zap.String("app", appID),
		)

		return &struct{}{}, nil
	}))
}

func (c *Controller) handleAppUserDelete(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "app-id")
	userID := chi.URLParam(r, "user-id")

	if userID == authn(r).UserID {
		writeResponse(w, nil, models.ErrDeleteCurrentUser)
		return
	}

	respond(w, withTx(r.Context(), c.DB, func(tx db.Tx) (any, error) {
		err := tx.UnassignAppUser(r.Context(), appID, userID)
		if err != nil {
			return nil, err
		}

		c.Logger.Info("removing user",
			zap.String("request_id", requestID(r)),
			zap.String("user", authn(r).UserID),
			zap.String("target_user", userID),
			zap.String("app", appID),
		)

		return &struct{}{}, nil
	}))
}
