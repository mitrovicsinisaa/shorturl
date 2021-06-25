package mid

import (
	"context"
	"expvar"
	"net/http"
	"runtime"

	"github.com/mitrovicsinisaa/shorturl/foundation/web"
)

// m contains the global program counters for application.
var m = struct {
	gr  *expvar.Int
	req *expvar.Int
	err *expvar.Int
}{
	gr:  expvar.NewInt("gorutines"),
	req: expvar.NewInt("requests"),
	err: expvar.NewInt("errors"),
}

// Metrics updates program counters.
func Metrics() web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			err := handler(ctx, w, r)

			// Increment request count
			m.req.Add(1)

			// Update the count for the number of active gorutines every 100 requests.
			if m.req.Value()%100 == 0 {
				m.gr.Set(int64(runtime.NumGoroutine()))
			}

			// Increment the errors counter if an error occured on this request.
			if err != nil {
				m.err.Add(1)
			}

			// Rerurn the error so it can be handled further up the chain.
			return err
		}

		return h
	}

	return m
}
