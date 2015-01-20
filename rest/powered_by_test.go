package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPoweredByMiddleware(t *testing.T) {

	poweredBy := &PoweredByMiddleware{
		XPoweredBy: "test",
	}

	app := func(w ResponseWriter, r *Request) {
		w.WriteJson(map[string]string{"Id": "123"})
	}

	handlerFunc := WrapMiddlewares([]Middleware{poweredBy}, app)

	// fake request
	origRequest, _ := http.NewRequest("GET", "http://localhost/", nil)
	r := &Request{
		origRequest,
		nil,
		map[string]interface{}{},
	}

	// fake writer
	recorder := httptest.NewRecorder()
	w := &responseWriter{
		recorder,
		false,
	}

	// run
	handlerFunc(w, r)

	// header
	value := recorder.HeaderMap.Get("X-Powered-By")
	if value != "test" {
		t.Errorf("Expected X-Powered-By to be 'test', got %s", value)
	}
}
