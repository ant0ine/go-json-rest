package rest

import (
	"net/http"
)

// Api defines a stack of middlewares and an app.
type Api struct {
	stack []Middleware
	app   App
}

// NewApi makes a new Api object, the App is required.
func NewApi(app App) *Api {

	if app == nil {
		panic("app is required")
	}

	return &Api{
		stack: []Middleware{},
		app:   app,
	}
}

// Use pushes one or multiple middlewares to the stack for middlewares
// maintained in the Api object.
func (api *Api) Use(middlewares ...Middleware) {
	api.stack = append(api.stack, middlewares...)
}

// MakeHandler wraps all the middlewares of the stack and the app together, and
// returns an http.Handler ready to be used.
func (api *Api) MakeHandler() http.Handler {
	return http.HandlerFunc(
		adapterFunc(
			WrapMiddlewares(api.stack, api.app.AppFunc()),
		),
	)
}

// Defines a stack of middlewares convenient for development. Among other things:
// console friendly logging, JSON indentation, error stack strace in the response.
var DefaultDevStack = []Middleware{
	&AccessLogApacheMiddleware{},
	&TimerMiddleware{},
	&RecorderMiddleware{},
	&JsonIndentMiddleware{},
	&PoweredByMiddleware{},
	&ContentTypeCheckerMiddleware{},
	&RecoverMiddleware{
		EnableResponseStackTrace: true,
	},
}

// Defines a stack of middlewares convenient for production. Among other things:
// Apache CombinedLogFormat logging, gzip compression.
var DefaultProdStack = []Middleware{
	&AccessLogApacheMiddleware{
		Format: CombinedLogFormat,
	},
	&TimerMiddleware{},
	&RecorderMiddleware{},
	&GzipMiddleware{},
	&PoweredByMiddleware{},
	&ContentTypeCheckerMiddleware{},
	&RecoverMiddleware{},
}
