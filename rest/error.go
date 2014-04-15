package rest

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

// errorMiddleware catches the user panic errors and convert them to 500
type errorMiddleware struct {
	Logger                   *log.Logger
	EnableLogAsJson          bool
	EnableResponseStackTrace bool
}

func (mw *errorMiddleware) MiddlewareFunc(h HandlerFunc) HandlerFunc {

	return func(w ResponseWriter, r *Request) {

		// catch user code's panic, and convert to http response
		defer func() {
			if reco := recover(); reco != nil {
				trace := debug.Stack()

				// log the trace
				// TODO this should be logging JSON if EnableLogAsJson
				mw.Logger.Printf("%s\n%s", reco, trace)

				// write error response
				message := "Internal Server Error"
				if mw.EnableResponseStackTrace {
					message = fmt.Sprintf("%s\n\n%s", reco, trace)
				}
				Error(w, message, http.StatusInternalServerError)
			}
		}()

		// call the handler
		h(w, r)
	}
}
