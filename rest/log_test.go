package rest

import (
	"net/http/httptest"
	"testing"
        "net/http"
)

func TestLogMiddleware(t *testing.T) {

	recorder := &recorderMiddleware{}
	timer := &timerMiddleware{}
        logger := &logMiddleware{}

        // TODO test the defaults

	app := func(w ResponseWriter, r *Request) {
		w.WriteJson(map[string]string{"Id": "123"})
	}

        // same order as in ResourceHandler
	handlerFunc := WrapMiddlewares([]Middleware{logger, timer, recorder}, app)

	// fake request
        origRequest, _ := http.NewRequest("GET", "http://localhost/", nil)
	r := &Request{
                origRequest,
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

        // TODO actually test the output
}
