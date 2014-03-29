package rest

import (
	"fmt"
	"log"
	"mime"
	"net/http"
	"reflect"
	"runtime/debug"
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
	handlerPath	     string

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

	// If true, the record that is logged for each response will be printed as JSON
	// in the log. Convenient for log parsing.
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

	// Custom logger, defaults to log.New(os.Stderr, "", log.LstdFlags)
	Logger *log.Logger
}

// Route defines a route. It's used with SetRoutes.
type Route struct {

	// Any http method. It will be used as uppercase to avoid common mistakes.
	HttpMethod string

	// A string like "/resource/:id.json".
	// Placeholders supported are:
	// :param that matches any char to the first '/' or '.'
	// *splat that matches everything to the end of the string
	// (placeholder names should be unique per PathExp)
	PathExp string

	// Code that will be executed when this route is taken.
	Func HandlerFunc
}

// RouteObjectMethod creates a Route that points to an object method. It can be convenient to point to
// an object method instead of a function, this helper makes it easy by passing the object instance and
// the method name as parameters.
func RouteObjectMethod(httpMethod string, pathExp string, objectInstance interface{}, objectMethod string) Route {

	value := reflect.ValueOf(objectInstance)
	funcValue := value.MethodByName(objectMethod)
	if funcValue.IsValid() == false {
		panic(fmt.Sprintf(
			"Cannot find the object method %s on %s",
			objectMethod,
			value,
		))
	}
	routeFunc := func(w ResponseWriter, r *Request) {
		funcValue.Call([]reflect.Value{
			reflect.ValueOf(w),
			reflect.ValueOf(r),
		})
	}

	return Route{
		HttpMethod: httpMethod,
		PathExp:    pathExp,
		Func:       routeFunc,
	}
}

// SetRoutes defines the Routes. The order the Routes matters,
// if a request matches multiple Routes, the first one will be used.
func (rh *ResourceHandler) SetRoutes(routes ...Route) error {
	
	// start the router
	rh.internalRouter = &router{
		routes: routes,
	}
	err := rh.internalRouter.start()
	if err != nil {
		return err
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

		// catch user code's panic, and convert to http response
		// (this does not use the JSON error response on purpose)
		defer func() {
			if reco := recover(); reco != nil {
				trace := debug.Stack()

				// log the trace
				rh.Logger.Printf("%s\n%s", reco, trace)

				// write error response
				message := "Internal Server Error"
				if rh.EnableResponseStackTrace {
					message = fmt.Sprintf("%s\n\n%s", reco, trace)
				}
				http.Error(origWriter, message, http.StatusInternalServerError)
			}
		}()

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

		// remove handler path from URL
		var path = request.URL.Path
		if strings.HasPrefix(path, rh.handlerPath) {
			path = "/"+strings.TrimPrefix(path, rh.handlerPath)
		} else {
			Error(writer, "Handler path is not set up correctly", http.StatusInternalServerError)
		}
		request.URL.Path = path

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

// Set up http handler
func (rh *ResourceHandler) Handle(path string) {
	rh.handlerPath = path
	http.Handle(rh.handlerPath, rh)
}