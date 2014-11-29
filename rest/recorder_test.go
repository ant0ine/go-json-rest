package rest

import (
	"net/http/httptest"
	"testing"
)

func TestRecorderMiddleware(t *testing.T) {

	mw := &recorderMiddleware{}

	app := func(w ResponseWriter, r *Request) {
		w.WriteJson(map[string]string{"Id": "123"})
	}

	handlerFunc := WrapMiddlewares([]Middleware{mw}, app)

	// fake request
	r := &Request{
		nil,
		nil,
		map[string]interface{}{},
	}

	// fake writer
	w := &responseWriter{
		httptest.NewRecorder(),
		false,
		false,
		"",
	}

	handlerFunc(w, r)

	if r.Env["STATUS_CODE"] == nil {
		t.Error("STATUS_CODE is nil")
	}
	statusCode := r.Env["STATUS_CODE"].(int)
	if statusCode != 200 {
		t.Errorf("statusCode = 200 expected, got %d", statusCode)
	}
}
