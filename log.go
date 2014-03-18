package rest

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// logMiddleware manages the Logger.
// It depends on request.Env["STATUS_CODE"] and request.Env["ELAPSED_TIME"].
type logMiddleware struct {
	Logger          *log.Logger
	EnableLogAsJson bool
}

func (mw *logMiddleware) MiddlewareFunc(h HandlerFunc) HandlerFunc {

	// set a default Logger
	if mw.Logger == nil {
		mw.Logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	return func(w ResponseWriter, r *Request) {

		// call the handler
		h(w, r)

		mw.logResponseRecord(&responseLogRecord{
			r.Env["STATUS_CODE"].(int),
			r.Env["ELAPSED_TIME"].(*time.Duration),
			r.Method,
			r.URL.RequestURI(),
		})
	}
}

type responseLogRecord struct {
	StatusCode   int
	ResponseTime *time.Duration
	HttpMethod   string
	RequestURI   string
}

func (mw *logMiddleware) logResponseRecord(record *responseLogRecord) {
	if mw.EnableLogAsJson {
		b, err := json.Marshal(record)
		if err != nil {
			panic(err)
		}
		mw.Logger.Printf("%s", b)
	} else {
		statusCodeColor := "0;32"
		if record.StatusCode >= 400 && record.StatusCode < 500 {
			statusCodeColor = "1;33"
		} else if record.StatusCode >= 500 {
			statusCodeColor = "0;31"
		}
		mw.Logger.Printf("\033[%sm%d\033[0m \033[36;1m%.2fms\033[0m %s %s",
			statusCodeColor,
			record.StatusCode,
			float64(record.ResponseTime.Nanoseconds()/1e4)/100.0,
			record.HttpMethod,
			record.RequestURI,
		)
	}
}
