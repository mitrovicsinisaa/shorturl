package handlers

import (
	"context"
	"net/http"

	"github.com/mitrovicsinisaa/shorturl/business/auth"
	"github.com/mitrovicsinisaa/shorturl/business/data/user"
	"github.com/mitrovicsinisaa/shorturl/foundation/web"
	"github.com/pkg/errors"
)

type userGroup struct {
	user user.User
	auth *auth.Auth
}

func (ug userGroup) token(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	email, pass, ok := r.BasicAuth()
	if !ok {
		err := errors.New("must provide email and password in Basic auth")
		return web.NewRequestError(err, http.StatusUnauthorized)
	}

	claims, err := ug.user.Authenticate(ctx, v.TraceID, v.Now, email, pass)
	if err != nil {
		switch errors.Cause(err) {
		case user.ErrAuthenticationFailure:
			return web.NewRequestError(err, http.StatusUnauthorized)
		default:
			return errors.Wrap(err, "authenticating")
		}
	}

	params := web.Params(r)
	var tkn struct {
		Token string `json:"token"`
	}
	tkn.Token, err = ug.auth.GenerateToken(params["kid"], claims)
	if err != nil {
		return errors.Wrap(err, "generating token")
	}

	return web.Respond(ctx, w, tkn, http.StatusOK)
}
