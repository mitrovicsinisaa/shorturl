package mid

import (
	"context"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/mitrovicsinisaa/shorturl/foundation/web"
	"github.com/pkg/errors"
)

// Panics recovers form panics and converts panic to an error so it is
// reported in metrics and handled in errors.
func Panics(log *log.Logger) web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {

			// If the context is missing this value, request the service
			// to be shutdown gracefuly.
			v, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok {
				return web.NewShutdownError("web value missing from context")
			}

			// Defer a function to recover from a panic and set the err return
			// variable after the fact.
			defer func() {
				if r := recover(); r != nil {
					err = errors.Errorf("panic: %v", r)

					// Log to Go stack trace for this panic's gorutine.
					log.Printf("%s : PANIC     :\n%s", v.TraceID, debug.Stack())
				}
			}()

			// Call the next handler and set its return value in the err variable.
			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
