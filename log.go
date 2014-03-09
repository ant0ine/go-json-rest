package rest

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type responseLogRecord struct {
	StatusCode   int
	ResponseTime *time.Duration
	HttpMethod   string
	RequestURI   string
}

func (rh *ResourceHandler) logResponseRecord(record *responseLogRecord) {
	if rh.EnableLogAsJson {
		b, err := json.Marshal(record)
		if err != nil {
			panic(err)
		}
		rh.Logger.Printf("%s", b)
	} else {
		statusCodeColor := "0;32"
		if record.StatusCode >= 400 && record.StatusCode < 500 {
			statusCodeColor = "1;33"
		} else if record.StatusCode >= 500 {
			statusCodeColor = "0;31"
		}
		rh.Logger.Printf("\033[%sm%d\033[0m \033[36;1m%.2fms\033[0m %s %s",
			statusCodeColor,
			record.StatusCode,
			float64(record.ResponseTime.Nanoseconds()/1e4)/100.0,
			record.HttpMethod,
			record.RequestURI,
		)
	}
}

func (rh *ResourceHandler) logWrapper(h http.HandlerFunc) http.HandlerFunc {

	// set a default Logger
	if rh.Logger == nil {
		rh.Logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	return func(w http.ResponseWriter, r *http.Request) {

		// call the handler
		h(w, r)

		rh.logResponseRecord(&responseLogRecord{
			rh.env.getVar(r, "statusCode").(int),
			rh.env.getVar(r, "elapsedTime").(*time.Duration),
			r.Method,
			r.URL.RequestURI(),
		})
	}
}
