package rest

const xPoweredByDefault = "go-json-rest"

// poweredByMiddleware adds the "X-Powered-By" header to the HTTP response.
type poweredByMiddleware struct {

	// If specified, used as the value for the "X-Powered-By" response header.
	// Defaults to "go-json-rest".
	XPoweredBy string
}

// MiddlewareFunc makes poweredByMiddleware implement the Middleware interface.
func (mw *poweredByMiddleware) MiddlewareFunc(h HandlerFunc) HandlerFunc {

	poweredBy := xPoweredByDefault
	if mw.XPoweredBy != "" {
		poweredBy = mw.XPoweredBy
	}

	return func(w ResponseWriter, r *Request) {

		w.Header().Add("X-Powered-By", poweredBy)

		// call the handler
		h(w, r)

	}
}
