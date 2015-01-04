package rest

import (
	"net/http"
)

// HandlerFunc defines the handler function. It is the go-json-rest equivalent of http.HandlerFunc.
type HandlerFunc func(ResponseWriter, *Request)

// Middleware defines the interface that objects must implement in order to wrap a HandlerFunc and
// be used in the middleware stack.
type Middleware interface {
	MiddlewareFunc(handler HandlerFunc) HandlerFunc
}

// WrapMiddlewares calls the MiddlewareFunc methods in the reverse order and returns an HandlerFunc
// ready to be executed. This can be used to wrap a set of middlewares, post routing, on a per Route
// basis.
func WrapMiddlewares(middlewares []Middleware, handler HandlerFunc) HandlerFunc {
	wrapped := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i].MiddlewareFunc(wrapped)
	}
	return wrapped
}

// Handle the transition between net/http and go-json-rest objects.
// It intanciates the rest.Request and rest.ResponseWriter, ...
func adapterFunc(handler HandlerFunc) http.HandlerFunc {

	return func(origWriter http.ResponseWriter, origRequest *http.Request) {

		// instantiate the rest objects
		request := &Request{
			origRequest,
			nil,
			map[string]interface{}{},
		}

		writer := &responseWriter{
			origWriter,
			false,
		}

		// call the wrapped handler
		handler(writer, request)
	}
}
