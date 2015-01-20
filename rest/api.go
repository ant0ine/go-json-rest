package rest

import (
	"net/http"
)

// Api defines a stack of middlewares and an app.
type Api struct {
        middlewares []Middleware
        app App
}

// NewApi makes a new Api object, the App is required.
func NewApi(app App) *Api {

        if app == nil {
                panic("app is required")
        }

        return &Api{
                middlewares: []Middleware{},
                app: app,
        }
}

// Use pushes one or multiple middlewares to the stack for middlewares
// maintained in the Api object.
func (api *Api) Use(middlewares ...Middleware) {
        api.middlewares = append(api.middlewares, middlewares...)
}

// MakeHandler wraps all the middlewares of the stack and the app together, and
// returns an http.Handler ready to be used.
func (api *Api) MakeHandler() http.Handler {
	return http.HandlerFunc(
                adapterFunc(
		        WrapMiddlewares(api.middlewares, api.app.AppFunc()),
	        ),
        )
}

// Defines a stack of middlewares that is convenient for development.
var DefaultDevStack = []Middleware{
        &accessLogApacheMiddleware{},
        &timerMiddleware{},
        &recorderMiddleware{},
        &jsonIndentMiddleware{},
        &poweredByMiddleware{},
        &recoverMiddleware{
                EnableResponseStackTrace: true,
        },
}
