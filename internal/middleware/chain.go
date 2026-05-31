package middleware

import "net/http"

// Chain composes middlewares around a base handler. Middlewares are applied
// in the order given, so the first argument is the outermost wrapper and the
// last is closest to h. Requests flow through middlewares in argument order,
// then into h.
//
// For example:
//
//	handler := middleware.Chain(mux,
//	    middleware.Recover(logger),
//	    middleware.Logging(logger),
//	    middleware.Auth(...),
//	)
//
// is equivalent to:
//
//	handler := middleware.Recover(logger)(
//	    middleware.Logging(logger)(
//	        middleware.Auth(...)(mux),
//	    ),
//	)
//
// Recover sees the request first and the response last, making it the right
// place for panic handling that should observe everything below it.
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
