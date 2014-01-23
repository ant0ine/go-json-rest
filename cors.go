package rest

import "strconv"

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
	ExposeHeaders string
	// Time in seconds for which preflights will be cached
	MaxAge uint64
	// List of headers allowed in cors request
	AllowHeaders string
	// If this function is not nil it'll be called after populating CorsRequestInfo and filling default headers (this means it is possible to override them)
	// If this function returns nil this will mean that request is forbidden.
	// If this function is nil itself it will be silently ignored and response will contain default CORS headers
	AccessControl func(info *CorsRequest, headers *CorsResponseHeaders) error
}

func (self *Origin) newCorsPreflightHeaders(httpMethods []string) *CorsResponseHeaders {
	var methods string
	for _, httpMethod := range httpMethods {
		if "" == methods {
			methods = httpMethod
		} else {
			methods = methods + `, ` + httpMethod
		}
	}

	return &CorsResponseHeaders{
		AccessControlAllowOrigin:      self.Host,
		AccessControlAllowCredentials: strconv.FormatBool(self.AllowCredentials),
		AccessControlMaxAge:           strconv.FormatUint(self.MaxAge, 10),
		AccessControlAllowHeaders:     self.AllowHeaders,
		AccessControlAllowMethods:     methods,
	}
}

func (self *Origin) newCorsActualHeaders() *CorsResponseHeaders {
	return &CorsResponseHeaders{
		AccessControlAllowOrigin:      self.Host,
		AccessControlAllowCredentials: strconv.FormatBool(self.AllowCredentials),
		AccessControlExposeHeaders:    self.ExposeHeaders,
	}
}

type CorsRequest struct {
	*Request
	IsCors                      bool
	IsPreflight                 bool
	Origin                      string
	AccessControlRequestMethod  string
	AccessControlRequestHeaders string
}

type CorsResponseHeaders struct {
	AccessControlAllowOrigin      string
	AccessControlAllowCredentials string
	AccessControlExposeHeaders    string
	AccessControlMaxAge           string
	AccessControlAllowMethods     string
	AccessControlAllowHeaders     string
}

func newCorsRequest(r *Request) *CorsRequest {

	origin := r.Header.Get("Origin")
	isCors := origin != ""
	reqMethod := r.Header.Get("Access-Control-Request-Method")
	reqHeaders := r.Header.Get("Access-Control-Request-Headers")
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

func (self *CorsResponseHeaders) setPreflightHeaders(w *ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", self.AccessControlAllowOrigin)
	w.Header().Set("Access-Control-Allow-Credentials", self.AccessControlAllowCredentials)
	w.Header().Set("Access-Control-Max-Age", self.AccessControlMaxAge)
	w.Header().Set("Access-Control-Allow-Methods", self.AccessControlAllowMethods)
	w.Header().Set("Access-Control-Allow-Headers", self.AccessControlAllowHeaders)
}

func (self *CorsResponseHeaders) setActualHeaders(w *ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", self.AccessControlAllowOrigin)
	w.Header().Set("Access-Control-Allow-Credentials", self.AccessControlAllowCredentials)
	w.Header().Set("Access-Control-Expose-Headers", self.AccessControlExposeHeaders)
}
