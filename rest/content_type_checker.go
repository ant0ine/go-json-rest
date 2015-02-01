package rest

import (
	"mime"
	"net/http"
	"strings"
)

// ContentTypeCheckerMiddleware verifies the request Content-Type header and returns a
// StatusUnsupportedMediaType (415) HTTP error response if it's incorrect. The expected
// Content-Type is 'application/json' if the content is non-null. Note: If a charset parameter
// exists, it MUST be UTF-8.
type ContentTypeCheckerMiddleware struct{}

// MiddlewareFunc makes ContentTypeCheckerMiddleware implement the Middleware interface.
func (mw *ContentTypeCheckerMiddleware) MiddlewareFunc(handler HandlerFunc) HandlerFunc {

	return func(w ResponseWriter, r *Request) {

		mediatype, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		charset, ok := params["charset"]
		if !ok {
			charset = "UTF-8"
		}

		// per net/http doc, means that the length is known and non-null
		if r.ContentLength > 0 &&
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
