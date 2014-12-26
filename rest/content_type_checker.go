package rest

import (
	"mime"
	"net/http"
	"strings"
)

// contentTypeCheckerMiddleware verify the Request content type and returns a
// StatusUnsupportedMediaType (415) HTTP error response if it's incorrect.
type contentTypeCheckerMiddleware struct{}

// MiddlewareFunc returns a HandlerFunc that implements the middleware.
func (mw *contentTypeCheckerMiddleware) MiddlewareFunc(handler HandlerFunc) HandlerFunc {

	return func(w ResponseWriter, r *Request) {

		mediatype, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		charset, ok := params["charset"]
		if !ok {
			charset = "UTF-8"
		}

		if r.ContentLength > 0 && // per net/http doc, means that the length is known and non-null
			!(mediatype == "application/json" && strings.ToUpper(charset) == "UTF-8") {

			Error(w,
				"Bad Content-Type or charset, expected 'application/json'",
				http.StatusUnsupportedMediaType,
			)
			return
		}

		// call the wrapped handler
		handler(w, r)
	}
}
