// Package handlers contains the full set of handler functions and routes
// supported by web api.
package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/mitrovicsinisaa/shorturl/business/auth"
	"github.com/mitrovicsinisaa/shorturl/business/data/shorturl"
	"github.com/mitrovicsinisaa/shorturl/business/data/user"
	"github.com/mitrovicsinisaa/shorturl/business/mid"
	"github.com/mitrovicsinisaa/shorturl/foundation/web"
)

// API construct an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, log *log.Logger, a *auth.Auth, db *sqlx.DB) *web.App {
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	cg := checkGroup{
		build: build,
		db:    db,
	}
	app.Handle(http.MethodGet, "/readiness", cg.readiness)
	app.Handle(http.MethodGet, "/liveness", cg.liveness)

	// Register user management and authentication endpoints.
	ug := userGroup{
		user: user.New(log, db),
		auth: a,
	}
	app.Handle(http.MethodGet, "/api/token/:kid", ug.token)

	// Register user management and authentication endpoints.
	sg := shorturlGroup{
		shorturl: shorturl.New(log, db),
		auth:     a,
	}

	app.Handle(http.MethodPost, "/api/shorturl", sg.create)
	app.Handle(http.MethodGet, "/:url", sg.queryByID)
	app.Handle(http.MethodGet, "/api/shorturl/:page/:rows", sg.query, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodGet, "/api/shorturl/:url", sg.queryVisitation, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodDelete, "/api/shorturl/:url", sg.delete, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))

	return app
}
