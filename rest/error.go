package rest

import (
	"encoding/json"
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
				message := fmt.Sprintf("%s\n%s", reco, trace)
				mw.logError(message)

				// write error response
				if mw.EnableResponseStackTrace {
					Error(w, message, http.StatusInternalServerError)
				} else {
					Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}
		}()

		// call the handler
		h(w, r)
	}
}

func (mw *errorMiddleware) logError(message string) {
	if mw.EnableLogAsJson {
		record := map[string]string{
			"error": message,
		}
		b, err := json.Marshal(&record)
		if err != nil {
			panic(err)
		}
		mw.Logger.Printf("%s", b)
	} else {
		mw.Logger.Print(message)
	}
}
