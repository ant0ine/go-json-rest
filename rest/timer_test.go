package rest

import (
	"github.com/ant0ine/go-json-rest/rest/test"
	"testing"
	"time"
)

func TestTimerMiddleware(t *testing.T) {

	api := NewApi()

	// a middleware carrying the Env tests
	api.Use(MiddlewareSimple(func(handler HandlerFunc) HandlerFunc {
		return func(w ResponseWriter, r *Request) {

			handler(w, r)

			if r.Env["ELAPSED_TIME"] == nil {
				t.Error("ELAPSED_TIME is nil")
			}
			elapsedTime := r.Env["ELAPSED_TIME"].(*time.Duration)
			if elapsedTime.Nanoseconds() <= 0 {
				t.Errorf(
					"ELAPSED_TIME is expected to be at least 1 nanosecond %d",
					elapsedTime.Nanoseconds(),
				)
			}

			if r.Env["START_TIME"] == nil {
				t.Error("START_TIME is nil")
			}
			start := r.Env["START_TIME"].(*time.Time)
			if start.After(time.Now()) {
				t.Errorf(
					"START_TIME is expected to be in the past %s",
					start.String(),
				)
			}
		}
	}))

	// the middleware to test
	api.Use(&TimerMiddleware{})

	// a simple app
	api.SetApp(AppSimple(func(w ResponseWriter, r *Request) {
		w.WriteJson(map[string]string{"Id": "123"})
	}))

	// wrap all
	handler := api.MakeHandler()

	req := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	recorded := test.RunRequest(t, handler, req)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()
}
