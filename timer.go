package rest

import (
	"time"
)

func (rh *ResourceHandler) timerWrapper(h HandlerFunc) HandlerFunc {
	return func(w ResponseWriter, r *Request) {

		start := time.Now()

		// call the handler
		h(w, r)

		end := time.Now()
		elapsed := end.Sub(start)
		r.Env["elapsedTime"] = &elapsed
	}
}
