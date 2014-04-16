package rest

import (
	"log"
	"mime"
	"net/http"
	"os"
	"strings"
)

// HandlerFunc defines the handler function. It is the go-json-rest equivalent of http.HandlerFunc.
type HandlerFunc func(ResponseWriter, *Request)

// Middleware defines the interface that objects must implement in order to wrap a HandlerFunc and
// be used in the middleware stack.
type Middleware interface {
	MiddlewareFunc(handler HandlerFunc) HandlerFunc
}

// ResourceHandler implements the http.Handler interface and acts a router for the defined Routes.
// The defaults are intended to be developemnt friendly, for production you may want
// to turn on gzip and disable the JSON indentation for instance.
type ResourceHandler struct {
	internalRouter   *router
	statusMiddleware *statusMiddleware
	handlerFunc      http.HandlerFunc

	// If true, and if the client accepts the Gzip encoding, the response payloads
	// will be compressed using gzip, and the corresponding response header will set.
	EnableGzip bool

	// If true, the JSON payload will be written in one line with no space.
	DisableJsonIndent bool

	// If true, the status service will be enabled. Various stats and status will
	// then be available at GET /.status in a JSON format.
	EnableStatusService bool

	// If true, when a "panic" happens, the error string and the stack trace will be
	// printed in the 500 response body.
	EnableResponseStackTrace bool

	// If true, the records logged to the access log and the error log will be
	// printed as JSON. Convenient for log parsing.
	EnableLogAsJson bool

	// If true, the handler does NOT check the request Content-Type. Otherwise, it
	// must be set to 'application/json' if the content is non-null.
	// Note: If a charset parameter exists, it MUST be UTF-8
	EnableRelaxedContentType bool

	// Optional global middlewares that can be used to wrap the all REST endpoints.
	// They are used in the defined order, the first wrapping the second, ...
	// They can be used for instance to manage CORS or authentication.
	// (see the CORS example in go-json-rest-example)
	// They are run pre REST routing, request.PathParams is not set yet.
	PreRoutingMiddlewares []Middleware

	// Custom logger for the access log,
	// optional, defaults to log.New(os.Stderr, "", 0)
	Logger *log.Logger

	// Custom logger used for logging the panic errors,
	// optional, defaults to log.New(os.Stderr, "", 0)
	ErrorLogger *log.Logger

	// Custom X-Powered-By value, defaults to "go-json-rest".
	XPoweredBy string

	// If true, the X-Powered-By header will NOT be set.
	DisableXPoweredBy bool
}

// SetRoutes defines the Routes. The order the Routes matters,
// if a request matches multiple Routes, the first one will be used.
func (rh *ResourceHandler) SetRoutes(routes ...*Route) error {

	// set the default Loggers
	if rh.Logger == nil {
		rh.Logger = log.New(os.Stderr, "", 0)
	}
	if rh.ErrorLogger == nil {
		rh.ErrorLogger = log.New(os.Stderr, "", 0)
	}

	// start the router
	rh.internalRouter = &router{
		routes: routes,
	}
	err := rh.internalRouter.start()
	if err != nil {
		return err
	}

	if rh.DisableXPoweredBy {
		rh.XPoweredBy = ""
	} else {
		if len(rh.XPoweredBy) == 0 {
			rh.XPoweredBy = xPoweredByDefault
		}
	}

	rh.instantiateMiddlewares()

	return nil
}

func (rh *ResourceHandler) instantiateMiddlewares() {

	// instantiate all the middlewares
	middlewares := []Middleware{
		// log as the first, depend on timer and recorder.
		&logMiddleware{
			rh.Logger,
			rh.EnableLogAsJson,
		},
	}

	if rh.EnableGzip {
		middlewares = append(middlewares, &gzipMiddleware{})
	}

	if rh.EnableStatusService {
		// keep track of this middleware for GetStatus()
		rh.statusMiddleware = newStatusMiddleware()
		middlewares = append(middlewares, rh.statusMiddleware)
	}

	middlewares = append(middlewares,
		&timerMiddleware{},
		&recorderMiddleware{},
		&errorMiddleware{
			rh.ErrorLogger,
			rh.EnableLogAsJson,
			rh.EnableResponseStackTrace,
		},
	)

	middlewares = append(middlewares,
		rh.PreRoutingMiddlewares...,
	)

	handler := rh.app()
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i].MiddlewareFunc(handler)
	}

	rh.handlerFunc = rh.adapter(handler)
}

// Middleware that handles the transition between http and rest objects.
func (rh *ResourceHandler) adapter(handler HandlerFunc) http.HandlerFunc {
	return func(origWriter http.ResponseWriter, origRequest *http.Request) {

		// instantiate the rest objects
		request := Request{
			origRequest,
			nil,
			map[string]interface{}{},
		}

		isIndented := !rh.DisableJsonIndent

		writer := responseWriter{
			origWriter,
			false,
			isIndented,
			rh.XPoweredBy,
		}

		// call the wrapped handler
		handler(&writer, &request)
	}
}

// Handle the REST routing and run the user code.
func (rh *ResourceHandler) app() HandlerFunc {
	return func(writer ResponseWriter, request *Request) {

		// check the Content-Type
		mediatype, params, _ := mime.ParseMediaType(request.Header.Get("Content-Type"))
		charset, ok := params["charset"]
		if !ok {
			charset = "UTF-8"
		}

		if rh.EnableRelaxedContentType == false &&
			request.ContentLength > 0 && // per net/http doc, means that the length is known and non-null
			!(mediatype == "application/json" && strings.ToUpper(charset) == "UTF-8") {

			Error(writer,
				"Bad Content-Type or charset, expected 'application/json'",
				http.StatusUnsupportedMediaType,
			)
			return
		}

		// find the route
		route, params, pathMatched := rh.internalRouter.findRouteFromURL(request.Method, request.URL)
		if route == nil {

			if pathMatched {
				// no route found, but path was matched: 405 Method Not Allowed
				Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			// no route found, the path was not matched: 404 Not Found
			NotFound(writer, request)
			return
		}

		// a route was found, set the PathParams
		request.PathParams = params

		// run the user code
		handler := route.Func
		handler(writer, request)
	}
}

// This makes ResourceHandler implement the http.Handler interface.
// You probably don't want to use it directly.
func (rh *ResourceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rh.handlerFunc(w, r)
}

// GetStatus returns a Status object. EnableStatusService must be true.
func (rh *ResourceHandler) GetStatus() *Status {
	return rh.statusMiddleware.getStatus()
}
