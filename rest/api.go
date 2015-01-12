package rest

import (
	"net/http"
)

type Api struct {
        middlewares []Middleware
        app App
}

func NewApi(app App) *Api {

        if app == nil {
                panic("app is required")
        }

        return &Api{
                middlewares: []Middleware{},
                app: app,
        }
}

func (api *Api) Use(middlewares ...Middleware) {
        api.middlewares = append(api.middlewares, middlewares...)
}

func (api *Api) MakeHandler() http.Handler {
	return http.HandlerFunc(
                adapterFunc(
		        WrapMiddlewares(api.middlewares, api.app.AppFunc()),
	        ),
        )
}

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
