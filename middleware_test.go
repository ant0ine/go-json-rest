package rest

import (
	"testing"
)

type testMiddleware struct {
	name string
}

func (mw *testMiddleware) MiddlewareFunc(handler HandlerFunc) HandlerFunc {
	return func(w ResponseWriter, r *Request) {
		if r.Env["BEFORE"] == nil {
			r.Env["BEFORE"] = mw.name
		} else {
			r.Env["BEFORE"] = r.Env["BEFORE"].(string) + mw.name
		}
		handler(w, r)
		if r.Env["AFTER"] == nil {
			r.Env["AFTER"] = mw.name
		} else {
			r.Env["AFTER"] = r.Env["AFTER"].(string) + mw.name
		}
	}
}

func TestWrapMiddlewares(t *testing.T) {

	a := &testMiddleware{"A"}
	b := &testMiddleware{"B"}
	c := &testMiddleware{"C"}

	app := func(w ResponseWriter, r *Request) {
		// do nothing
	}

	handlerFunc := WrapMiddlewares([]Middleware{a, b, c}, app)

	// fake request
	r := &Request{
		nil,
		nil,
		map[string]interface{}{},
	}

	handlerFunc(nil, r)

	before := r.Env["BEFORE"].(string)
	if before != "ABC" {
		t.Error("middleware executed in the wrong order, expected ABC")
	}

	after := r.Env["AFTER"].(string)
	if after != "CBA" {
		t.Error("middleware executed in the wrong order, expected CBA")
	}
}
