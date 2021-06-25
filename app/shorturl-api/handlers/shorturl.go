package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mitrovicsinisaa/shorturl/business/auth"
	"github.com/mitrovicsinisaa/shorturl/business/base62"
	"github.com/mitrovicsinisaa/shorturl/business/data/shorturl"
	"github.com/mitrovicsinisaa/shorturl/foundation/web"
	"github.com/pkg/errors"
)

type shorturlGroup struct {
	shorturl shorturl.Shorturl
	auth     *auth.Auth
}

func (sg shorturlGroup) query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	params := web.Params(r)
	pageNumber, err := strconv.Atoi(params["page"])
	if err != nil {
		return web.NewRequestError(fmt.Errorf("invalid page format: %s", params["page"]), http.StatusBadRequest)
	}
	rowsPerPage, err := strconv.Atoi(params["rows"])
	if err != nil {
		return web.NewRequestError(fmt.Errorf("invalid rows format: %s", params["rows"]), http.StatusBadRequest)
	}

	shorturls, err := sg.shorturl.Query(ctx, v.TraceID, pageNumber, rowsPerPage)
	if err != nil {
		return errors.Wrap(err, "unable to query for shorturls")
	}

	return web.Respond(ctx, w, shorturls, http.StatusOK)
}

func (sg shorturlGroup) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	params := web.Params(r)

	shorturlID, err := base62.Decode(params["url"])
	if err != nil {
		return errors.New("invalid url")
	}

	surl, err := sg.shorturl.QueryByID(ctx, v.TraceID, shorturlID)
	if err != nil {
		switch err {
		case shorturl.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "URL: %s", params["url"])
		}
	}

	http.Redirect(w, r, surl.URL, http.StatusSeeOther)

	return nil
}

func (sg shorturlGroup) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	var nsu shorturl.NewShorturl
	if err := web.Decode(r, &nsu); err != nil {
		return errors.Wrap(err, "unable to decode payload")
	}

	surl, err := sg.shorturl.Create(ctx, v.TraceID, nsu, v.Now)
	if err != nil {
		return errors.Wrapf(err, "Shorturl: %+v", &surl)
	}

	data := struct {
		ShortUrl string `json:"shorturl"`
	}{
		ShortUrl: r.Host + "/" + base62.Encode(surl.ID),
	}

	return web.Respond(ctx, w, data, http.StatusCreated)
}

func (sg shorturlGroup) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	params := web.Params(r)
	shorturlID, err := base62.Decode(params["url"])
	if err != nil {
		return errors.New("invalid url")
	}

	err = sg.shorturl.Delete(ctx, v.TraceID, claims, shorturlID)
	if err != nil {
		switch errors.Cause(err) {
		case shorturl.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case shorturl.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "URL: %s", params["url"])
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (sg shorturlGroup) queryVisitation(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return errors.New("claims missing from context")
	}

	params := web.Params(r)
	shorturlID, err := base62.Decode(params["url"])
	if err != nil {
		return errors.New("invalid url")
	}

	surl, err := sg.shorturl.QueryVisitation(ctx, v.TraceID, claims, shorturlID)
	if err != nil {
		switch errors.Cause(err) {
		case shorturl.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case shorturl.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "URL: %s", params["url"])
		}
	}

	return web.Respond(ctx, w, surl, http.StatusOK)
}
