package rest

import (
	"testing"
	"time"
)

func TestTimerMiddleware(t *testing.T) {

	mw := &timerMiddleware{}

	app := func(w ResponseWriter, r *Request) {
		// do nothing
	}

	handlerFunc := WrapMiddlewares([]Middleware{mw}, app)

	// fake request
	r := &Request{
		nil,
		nil,
		map[string]interface{}{},
	}

	handlerFunc(nil, r)

	if r.Env["ELAPSED_TIME"] == nil {
		t.Error("ELAPSED_TIME is nil")
	}
	elapsedTime := r.Env["ELAPSED_TIME"].(*time.Duration)
	if elapsedTime.Nanoseconds() <= 0 {
		t.Errorf("ELAPSED_TIME is expected to be at least 1 nanosecond %d", elapsedTime.Nanoseconds())
	}
}
