package rest

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// accessLogJsonMiddleware produces the access log with records written as JSON.
// It depends on the timer, recorder and auth middlewares.
type accessLogJsonMiddleware struct {
	Logger *log.Logger
}

func (mw *accessLogJsonMiddleware) MiddlewareFunc(h HandlerFunc) HandlerFunc {

	// set the default Logger
	if mw.Logger == nil {
		mw.Logger = log.New(os.Stderr, "", 0)
	}

	return func(w ResponseWriter, r *Request) {

		// call the handler
		h(w, r)

		mw.Logger.Print(makeAccessLogJsonRecord(r).asJson())
	}
}

// When EnableLogAsJson is true, this object is dumped as JSON in the Logger.
// (Public for documentation only, no public method uses it).
type AccessLogJsonRecord struct {
	Timestamp    *time.Time
	StatusCode   int
	ResponseTime *time.Duration
	HttpMethod   string
	RequestURI   string
	RemoteUser   string
	UserAgent    string
}

func makeAccessLogJsonRecord(r *Request) *AccessLogJsonRecord {
	return &AccessLogJsonRecord{
		Timestamp:    r.Env["START_TIME"].(*time.Time),
		StatusCode:   r.Env["STATUS_CODE"].(int),
		ResponseTime: r.Env["ELAPSED_TIME"].(*time.Duration),
		HttpMethod:   r.Method,
		RequestURI:   r.URL.RequestURI(),
		RemoteUser:   r.Env["REMOTE_USER"].(string),
		UserAgent:    r.UserAgent(),
	}
}

func (r *AccessLogJsonRecord) asJson() []byte {
	b, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return b
}
