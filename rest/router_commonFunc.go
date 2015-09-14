package rest

import (
	"net/http"
)

type commonFuncRouter struct {
	router

	beforeHandler HandlerFunc
	afterHandler  HandlerFunc
}

// MakeCommonFuncRouter returns the router app. Given a set of Routes, it dispatches the request to the
// HandlerFunc of the first route that matches. The order of the Routes matters.
func MakeCommonFuncRouter(routes ...*Route) (App, error) {
	r := &commonFuncRouter{}
	r.Routes = routes

	err := r.start()
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Handle the REST routing and run the user code.
func (rt *commonFuncRouter) AppFunc() HandlerFunc {

	return func(writer ResponseWriter, request *Request) {

		// find the route
		route, params, pathMatched := rt.findRouteFromURL(request.Method, request.URL)
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

		// common router code before user code
		if rt.beforeHandler != nil {
			rt.beforeHandler(writer, request)
		}

		handler(writer, request)

		// common router code after user code
		if rt.afterHandler != nil {
			rt.afterHandler(writer, request)
		}
	}
}

// Set the common function, run before/after user code.
func SetRouterCommonFunc(route App, beforeFunc HandlerFunc, afterFunc HandlerFunc) {
	rt, ok := route.(*commonFuncRouter)
	if ok == false {
		return
	}

	rt.beforeHandler = beforeFunc
	rt.afterHandler = afterFunc
}
