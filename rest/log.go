package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
	"time"
)

// TODO Future improvements:
// * support %{strftime}t ?
// * support %{<header>}o to print headers
// * split this middleware in two, Apache and JSON

// AccessLogFormat defines the format of the access log record.
// This implementation is a subset of Apache mod_log_config.
// (See http://httpd.apache.org/docs/2.0/mod/mod_log_config.html)
//
// %b content length, not implemented yet, -
// %D response elapsed time in microseconds
// %h remote address
// %H server protocol
// %l identd logname, not supported, -
// %m http method
// %P process id
// %q query string
// %r first line of the request
// %s status code
// %S status code preceeded by a terminal color
// %t time of the request
// %T response elapsed time in seconds, 3 decimal
// %u remote user, - if missing
// %{User-Agent}i user agent, - if missing
// %{Referer}i referer, - is missing
//
// Some predefined format are provided, see the contant below.
type AccessLogFormat string

const (
	// Common Log Format (CLF).
	ApacheCommon = "%h %l %u %t \"%r\" %s %b"

	// NCSA extended/combined log format.
	ApacheCombined = "%h %l %u %t \"%r\" %s %b \"%{Referer}i\" \"%{User-agent}i\""

	// Default format, colored output and response time, convenient for development.
	Default = "%t %S\033[0m \033[36;1m%Dμs\033[0m \"%r\" \033[1;30m%u \"%{User-Agent}i\"\033[0m"
)

// logMiddleware manages the Logger.
// It depends on request.Env["STATUS_CODE"] and request.Env["ELAPSED_TIME"].
type logMiddleware struct {
	Logger          *log.Logger
	EnableLogAsJson bool
	textTemplate    *template.Template
	format          AccessLogFormat
}

func (mw *logMiddleware) MiddlewareFunc(h HandlerFunc) HandlerFunc {

	// set the default Logger
	if mw.Logger == nil {
		mw.Logger = log.New(os.Stderr, "", 0)
	}

	// set default format
	if mw.format == "" {
		mw.format = Default
	}

	mw.convertFormat()

	return func(w ResponseWriter, r *Request) {

		// call the handler
		h(w, r)

		util := &accessLogUtil{w, r}

		mw.logRecord(util)
	}
}

var apacheAdapter = strings.NewReplacer(
	"%b", "{{.BytesWritten}}",
	"%D", "{{.ResponseTime | microseconds}}",
	"%h", "{{.ApacheRemoteAddr}}",
	"%H", "{{.R.Proto}}",
	"%l", "-",
	"%m", "{{.R.Method}}",
	"%P", "{{.Pid}}",
	"%q", "{{.ApacheQueryString}}",
	"%r", "{{.R.Method}} {{.R.URL.RequestURI}} {{.R.Proto}}",
	"%s", "{{.StatusCode}}",
	"%S", "\033[{{.StatusCode | statusCodeColor}}m{{.StatusCode}}",
	"%t", "{{.StartTime.Format \"02/Jan/2006:15:04:05 -0700\"}}",
	"%T", "{{.ResponseTime.Seconds | printf \"%.3f\"}}",
	"%u", "{{.RemoteUser | dashIfEmptyStr}}",
	"%{User-Agent}i", "{{.R.UserAgent | dashIfEmptyStr}}",
	"%{Referer}i", "{{.R.Referer | dashIfEmptyStr}}",
)

// Convert the Apache access log format into a text/template
func (mw *logMiddleware) convertFormat() {

	tmplText := apacheAdapter.Replace(string(mw.format))

	funcMap := template.FuncMap{
		"dashIfEmptyStr": func(value string) string {
			if value == "" {
				return "-"
			}
			return value
		},
		"microseconds": func(dur *time.Duration) string {
			return fmt.Sprintf("%d", dur.Nanoseconds()/1000)
		},
		"statusCodeColor": func(statusCode int) string {
			if statusCode >= 400 && statusCode < 500 {
				return "1;33"
			} else if statusCode >= 500 {
				return "0;31"
			}
			return "0;32"
		},
	}

	var err error
	mw.textTemplate, err = template.New("accessLog").Funcs(funcMap).Parse(tmplText)
	if err != nil {
		panic(err)
	}
}

func (mw *logMiddleware) executeTextTemplate(util *accessLogUtil) string {
	buf := bytes.NewBufferString("")
	err := mw.textTemplate.Execute(buf, util)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func (mw *logMiddleware) logRecord(util *accessLogUtil) {
	if mw.EnableLogAsJson {
		mw.Logger.Print(makeAccessLogJsonRecord(util).asJson())
	} else {
		mw.Logger.Print(mw.executeTextTemplate(util))
	}
}

// accessLogUtil provides a collection of utility functions that devrive data from the Request object.
// This object id used to provide data to the Apache Style template and the the JSON log record.
type accessLogUtil struct {
	W ResponseWriter
	R *Request
}

// As stored by the auth middlewares.
func (u *accessLogUtil) RemoteUser() string {
	if u.R.Env["REMOTE_USER"] != nil {
		return u.R.Env["REMOTE_USER"].(string)
	}
	return ""
}

// If qs exists then return it with a leadin "?", apache log style.
func (u *accessLogUtil) ApacheQueryString() string {
	if u.R.URL.RawQuery != "" {
		return "?" + u.R.URL.RawQuery
	}
	return ""
}

// When the request entered the timer middleware.
func (u *accessLogUtil) StartTime() *time.Time {
	return u.R.Env["START_TIME"].(*time.Time)
}

// If remoteAddr is set then return is without the port number, apache log style.
func (u *accessLogUtil) ApacheRemoteAddr() string {
	remoteAddr := u.R.RemoteAddr
	if remoteAddr != "" {
		parts := strings.SplitN(remoteAddr, ":", 2)
		return parts[0]
	}
	return ""
}

// As recorded by the recorder middleware.
func (u *accessLogUtil) StatusCode() int {
	return u.R.Env["STATUS_CODE"].(int)
}

// As mesured by the timer middleware.
func (u *accessLogUtil) ResponseTime() *time.Duration {
	return u.R.Env["ELAPSED_TIME"].(*time.Duration)
}

// Process id.
func (u *accessLogUtil) Pid() int {
	return os.Getpid()
}

// As recorded by the recorder middleware.
func (u *accessLogUtil) BytesWritten() int64 {
	return u.R.Env["BYTES_WRITTEN"].(int64)
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

func makeAccessLogJsonRecord(u *accessLogUtil) *AccessLogJsonRecord {
	return &AccessLogJsonRecord{
		Timestamp:    u.StartTime(),
		StatusCode:   u.StatusCode(),
		ResponseTime: u.ResponseTime(),
		HttpMethod:   u.R.Method,
		RequestURI:   u.R.URL.RequestURI(),
		RemoteUser:   u.RemoteUser(),
		UserAgent:    u.R.UserAgent(),
	}
}

// The preferred format for machine readable logs.
func (r *AccessLogJsonRecord) asJson() []byte {
	b, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return b
}
