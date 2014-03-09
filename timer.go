package rest

import (
	"net/http"
	"time"
)

func (rh *ResourceHandler) timerWrapper(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		// call the handler
		h(w, r)

		end := time.Now()
		elapsed := end.Sub(start)
		rh.env.setVar(r, "elapsedTime", &elapsed)
	}
}
