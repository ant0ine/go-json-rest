package rest

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// AccessLogJsonMiddleware produces the access log with records written as JSON. This middleware
// depends on TimerMiddleware and RecorderMiddleware that must be in the wrapped middlewares. It
// also uses request.Env["REMOTE_USER"].(string) set by the auth middlewares.
type AccessLogJsonMiddleware struct {

	// Logger points to the logger object used by this middleware, it defaults to
	// log.New(os.Stderr, "", 0).
	Logger *log.Logger
}

// MiddlewareFunc makes AccessLogJsonMiddleware implement the Middleware interface.
func (mw *AccessLogJsonMiddleware) MiddlewareFunc(h HandlerFunc) HandlerFunc {

	// set the default Logger
	if mw.Logger == nil {
		mw.Logger = log.New(os.Stderr, "", 0)
	}

	return func(w ResponseWriter, r *Request) {

		// call the handler
		h(w, r)

		mw.Logger.Printf("%s", makeAccessLogJsonRecord(r).asJson())
	}
}

// AccessLogJsonRecord is the data structure used by AccessLogJsonMiddleware to create the JSON
// records. (Public for documentation only, no public method uses it)
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

	var timestamp *time.Time
	if r.Env["START_TIME"] != nil {
		timestamp = r.Env["START_TIME"].(*time.Time)
	}

	var statusCode int
	if r.Env["STATUS_CODE"] != nil {
		statusCode = r.Env["STATUS_CODE"].(int)
	}

	var responseTime *time.Duration
	if r.Env["ELAPSED_TIME"] != nil {
		responseTime = r.Env["ELAPSED_TIME"].(*time.Duration)
	}

	var remoteUser string
	if r.Env["REMOTE_USER"] != nil {
		remoteUser = r.Env["REMOTE_USER"].(string)
	}

	return &AccessLogJsonRecord{
		Timestamp:    timestamp,
		StatusCode:   statusCode,
		ResponseTime: responseTime,
		HttpMethod:   r.Method,
		RequestURI:   r.URL.RequestURI(),
		RemoteUser:   remoteUser,
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
