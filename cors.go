package rest

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type cors struct {
	origins []Origin
	index   map[string]*Origin
}

type Origin struct {
	// Host which is allowed for cross origin request
	// '*' sign has a special meaning allowing any origin to access this resource.
	// HTTP authentication, client-side SSL certificates and cookies are not allowed to be sent in case of '*'
	Host string
	// Allows sending credntials using cors
	AllowCredentials bool
	// List of headers which are allowed to be exposed in cors response
	ExposeHeaders []string
	// Time in seconds for which preflights will be cached
	MaxAge uint64
	// List of headers allowed in cors request
	AllowHeaders []string
	// If this function is not nil it'll be called after populating CorsRequestInfo and filling default headers (this means it is possible to override them)
	// If this function returns nil this will mean that request is forbidden.
	// If this function is nil itself it will be silently ignored and response will contain default CORS headers
	AccessControl func(info *CorsRequest, headers *CorsResponseHeaders) error
}

func (self *Origin) newCorsPreflightHeaders(httpMethods []string) *CorsResponseHeaders {
	return &CorsResponseHeaders{
		AccessControlAllowOrigin:      self.Host,
		AccessControlAllowCredentials: self.AllowCredentials,
		AccessControlMaxAge:           self.MaxAge,
		AccessControlAllowHeaders:     self.AllowHeaders,
		AccessControlAllowMethods:     httpMethods,
	}
}

func (self *Origin) newCorsActualHeaders() *CorsResponseHeaders {
	return &CorsResponseHeaders{
		AccessControlAllowOrigin:      self.Host,
		AccessControlAllowCredentials: self.AllowCredentials,
		AccessControlExposeHeaders:    self.ExposeHeaders,
	}
}

type CorsRequest struct {
	*http.Request
	IsCors                      bool
	IsPreflight                 bool
	Origin                      string
	AccessControlRequestMethod  string
	AccessControlRequestHeaders []string
}

type CorsResponseHeaders struct {
	AccessControlAllowOrigin      string
	AccessControlAllowCredentials bool
	AccessControlExposeHeaders    []string
	AccessControlMaxAge           uint64
	AccessControlAllowMethods     []string
	AccessControlAllowHeaders     []string
}

func newCorsRequest(r *http.Request) *CorsRequest {
	origin := r.Header.Get("Origin")
	// Chrome and safari send Origin header even within the same domain
	originUrl, err := url.ParseRequestURI(origin)
	isCors := nil == err && origin != "" && r.Host != originUrl.Host
	reqMethod := r.Header.Get("Access-Control-Request-Method")
	reqHeaders := strings.Split(r.Header.Get("Access-Control-Request-Headers"), `, `)
	isPreflight := isCors && r.Method == "OPTIONS" && reqMethod != ""

	return &CorsRequest{
		Request:     r,
		IsCors:      isCors,
		IsPreflight: isPreflight,
		Origin:      origin,
		AccessControlRequestMethod:  reqMethod,
		AccessControlRequestHeaders: reqHeaders,
	}
}

const (
	HeaderAccessControlAllowOrigin      = `Access-Control-Allow-Origin`
	HeaderAccessControlAllowCredentials = `Access-Control-Allow-Credentials`
	HeaderAccessConrtolMaxAge           = `Access-Control-Max-Age`
	HeaderAccessControlAllowMethods     = `Access-Control-Allow-Methods`
	HeaderAccessControlAllowHeaders     = `Access-Control-Allow-Headers`
	HeaderAccessControlExposeHeaders    = `Access-Control-Expose-Headers`
)

func (self *CorsResponseHeaders) setPreflightHeaders(w *ResponseWriter) {
	w.Header().Set(HeaderAccessControlAllowOrigin, self.AccessControlAllowOrigin)
	w.Header().Set(HeaderAccessControlAllowCredentials, strconv.FormatBool(self.AccessControlAllowCredentials))
	w.Header().Set(HeaderAccessConrtolMaxAge, strconv.FormatUint(self.AccessControlMaxAge, 10))
	w.Header().Set(HeaderAccessControlAllowMethods, strings.Join(self.AccessControlAllowMethods, `, `))
	w.Header().Set(HeaderAccessControlAllowHeaders, strings.Join(self.AccessControlAllowHeaders, `, `))
}

func (self *CorsResponseHeaders) setActualHeaders(w *ResponseWriter) {
	w.Header().Set(HeaderAccessControlAllowOrigin, self.AccessControlAllowOrigin)
	w.Header().Set(HeaderAccessControlAllowCredentials, strconv.FormatBool(self.AccessControlAllowCredentials))
	w.Header().Set(HeaderAccessControlExposeHeaders, strings.Join(self.AccessControlExposeHeaders, `, `))
}
