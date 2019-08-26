package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
)

// RecoverMiddleware catches the panic errors that occur in the wrapped HandleFunc,
// and convert them to 500 responses.
type RecoverMiddleware struct {

	// Custom logger used for logging the panic errors,
	// optional, defaults to log.New(os.Stderr, "", 0)
	Logger *log.Logger

	// If true, the log records will be printed as JSON. Convenient for log parsing.
	EnableLogAsJson bool

	// If true, when a "panic" happens, the error string and the stack trace will be
	// printed in the 500 response body.
	EnableResponseStackTrace bool
}

// MiddlewareFunc makes RecoverMiddleware implement the Middleware interface.
func (mw *RecoverMiddleware) MiddlewareFunc(h HandlerFunc) HandlerFunc {

	// set the default Logger
	if mw.Logger == nil {
		mw.Logger = log.New(os.Stderr, "", 0)
	}

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

func (mw *RecoverMiddleware) logError(message string) {
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
