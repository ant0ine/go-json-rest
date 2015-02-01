package rest

import (
	"log"
	"net/http"
)

// ResourceHandler implements the http.Handler interface and acts a router for the defined Routes.
// The defaults are intended to be developemnt friendly, for production you may want
// to turn on gzip and disable the JSON indentation for instance.
// ResourceHandler is now DEPRECATED in favor of the new Api object. See the migration guide.
type ResourceHandler struct {
	internalRouter   *router
	statusMiddleware *StatusMiddleware
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
	// See the AccessLogJsonRecord type for details of the access log JSON record.
	EnableLogAsJson bool

	// If true, the handler does NOT check the request Content-Type. Otherwise, it
	// must be set to 'application/json' if the content is non-null.
	// Note: If a charset parameter exists, it MUST be UTF-8
	EnableRelaxedContentType bool

	// Optional global middlewares that can be used to wrap the all REST endpoints.
	// They are used in the defined order, the first wrapping the second, ...
	// They are run first, wrapping all go-json-rest middlewares,
	// * request.PathParams is not set yet
	// * "panic" won't be caught and converted to 500
	// * request.Env["STATUS_CODE"] and request.Env["ELAPSED_TIME"] are set.
	// They can be used for extra logging, or reporting.
	// (see statsd example in in https://github.com/ant0ine/go-json-rest-examples)
	OuterMiddlewares []Middleware

	// Optional global middlewares that can be used to wrap the all REST endpoints.
	// They are used in the defined order, the first wrapping the second, ...
	// They are run pre REST routing, request.PathParams is not set yet.
	// They are run post auto error handling, "panic" will be converted to 500 errors.
	// They can be used for instance to manage CORS or authentication.
	// (see CORS and Auth examples in https://github.com/ant0ine/go-json-rest-examples)
	PreRoutingMiddlewares []Middleware

	// Custom logger for the access log,
	// optional, defaults to log.New(os.Stderr, "", 0)
	Logger *log.Logger

	// Define the format of the access log record.
	// When EnableLogAsJson is false, this format is used to generate the access log.
	// See AccessLogFormat for the options and the predefined formats.
	// Defaults to a developement friendly format specified by the Default constant.
	LoggerFormat AccessLogFormat

	// If true, the access log will be fully disabled.
	// (the log middleware is not even instantiated, avoiding any performance penalty)
	DisableLogger bool

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

	log.Print("ResourceHandler is now DEPRECATED in favor of the new Api object, see the migration guide")

	// intantiate all the middlewares based on the settings.
	middlewares := []Middleware{}

	middlewares = append(middlewares,
		rh.OuterMiddlewares...,
	)

	// log as the first, depends on timer and recorder.
	if !rh.DisableLogger {
		if rh.EnableLogAsJson {
			middlewares = append(middlewares,
				&AccessLogJsonMiddleware{
					Logger: rh.Logger,
				},
			)
		} else {
			middlewares = append(middlewares,
				&AccessLogApacheMiddleware{
					Logger: rh.Logger,
					Format: rh.LoggerFormat,
				},
			)
		}
	}

	// also depends on timer and recorder
	if rh.EnableStatusService {
		// keep track of this middleware for GetStatus()
		rh.statusMiddleware = &StatusMiddleware{}
		middlewares = append(middlewares, rh.statusMiddleware)
	}

	// after gzip in order to track to the content length and speed
	middlewares = append(middlewares,
		&TimerMiddleware{},
		&RecorderMiddleware{},
	)

	if rh.EnableGzip {
		middlewares = append(middlewares, &GzipMiddleware{})
	}

	if !rh.DisableXPoweredBy {
		middlewares = append(middlewares,
			&PoweredByMiddleware{
				XPoweredBy: rh.XPoweredBy,
			},
		)
	}

	if !rh.DisableJsonIndent {
		middlewares = append(middlewares, &JsonIndentMiddleware{})
	}

	// catch user errors
	middlewares = append(middlewares,
		&RecoverMiddleware{
			Logger:                   rh.ErrorLogger,
			EnableLogAsJson:          rh.EnableLogAsJson,
			EnableResponseStackTrace: rh.EnableResponseStackTrace,
		},
	)

	middlewares = append(middlewares,
		rh.PreRoutingMiddlewares...,
	)

	// verify the request content type
	if !rh.EnableRelaxedContentType {
		middlewares = append(middlewares,
			&ContentTypeCheckerMiddleware{},
		)
	}

	// instantiate the router
	rh.internalRouter = &router{
		Routes: routes,
	}
	err := rh.internalRouter.start()
	if err != nil {
		return err
	}

	// wrap everything
	rh.handlerFunc = adapterFunc(
		WrapMiddlewares(middlewares, rh.internalRouter.AppFunc()),
	)

	return nil
}

// This makes ResourceHandler implement the http.Handler interface.
// You probably don't want to use it directly.
func (rh *ResourceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rh.handlerFunc(w, r)
}

// GetStatus returns a Status object. EnableStatusService must be true.
func (rh *ResourceHandler) GetStatus() *Status {
	return rh.statusMiddleware.GetStatus()
}
