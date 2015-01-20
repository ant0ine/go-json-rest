package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJsonIndentMiddleware(t *testing.T) {

	jsonIndent := &JsonIndentMiddleware{}

	app := func(w ResponseWriter, r *Request) {
		w.WriteJson(map[string]string{"Id": "123"})
	}

	handlerFunc := WrapMiddlewares([]Middleware{jsonIndent}, app)

	// fake request
	origRequest, _ := http.NewRequest("GET", "http://localhost/", nil)
	origRequest.RemoteAddr = "127.0.0.1:1234"
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

	handlerFunc(w, r)

	expected := "{\n  \"Id\": \"123\"\n}"
	if recorder.Body.String() != expected {
		t.Errorf("expected %s, got : %s", expected, recorder.Body)
	}
}
