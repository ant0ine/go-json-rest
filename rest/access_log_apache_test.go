package rest

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestAccessLogApacheMiddleware(t *testing.T) {

	recorder := &recorderMiddleware{}
	timer := &timerMiddleware{}

	buffer := bytes.NewBufferString("")
	logger := &accessLogApacheMiddleware{
		Logger:       log.New(buffer, "", 0),
		Format:       CommonLogFormat,
		textTemplate: nil,
	}

	app := func(w ResponseWriter, r *Request) {
		w.WriteJson(map[string]string{"Id": "123"})
	}

	// same order as in ResourceHandler
	handlerFunc := WrapMiddlewares([]Middleware{logger, timer, recorder}, app)

	// fake request
	origRequest, _ := http.NewRequest("GET", "http://localhost/", nil)
	origRequest.RemoteAddr = "127.0.0.1:1234"
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

	// eg: '127.0.0.1 - - 29/Nov/2014:22:28:34 +0000 "GET / HTTP/1.1" 200 12'
	apacheCommon := regexp.MustCompile(`127.0.0.1 - - \d{2}/\w{3}/\d{4}:\d{2}:\d{2}:\d{2} \+0000 "GET / HTTP/1.1" 200 12`)

	if !apacheCommon.Match(buffer.Bytes()) {
		t.Errorf("Got: %s", buffer.String())
	}
}
