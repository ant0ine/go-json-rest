package rest

import (
	"net/http"
	"strings"
)

// SecureRedirectMiddleware redirects the client to the identical URL served via HTTPS
type SecureRedirectMiddleware struct{}

func (SecureRedirectMiddleware) MiddlewareFunc(h HandlerFunc) HandlerFunc {
	return func(w ResponseWriter, r *Request) {
		if strings.ToLower(r.Header.Get("X-Forwarded-Proto")) == "http" {
			redirectURL := r.URL
			redirectURL.Host = r.Host
			redirectURL.Scheme = "https"
			http.Redirect(w.(http.ResponseWriter), r.Request, redirectURL.String(), http.StatusMovedPermanently)
			return
		}
		h(w, r)
	}
}
