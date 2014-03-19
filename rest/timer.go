package rest

import (
	"time"
)

// timerMiddleware computes the elapsed time spent during the execution of the wrapped handler.
// The result is available to the wrapping handlers in request.Env["ELAPSED_TIME"] as a time.Duration.
type timerMiddleware struct{}

func (mw *timerMiddleware) MiddlewareFunc(h HandlerFunc) HandlerFunc {
	return func(w ResponseWriter, r *Request) {

		start := time.Now()

		// call the handler
		h(w, r)

		end := time.Now()
		elapsed := end.Sub(start)
		r.Env["ELAPSED_TIME"] = &elapsed
	}
}
