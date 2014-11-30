package rest

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAccessLogJsonMiddleware(t *testing.T) {

	recorder := &recorderMiddleware{}
	timer := &timerMiddleware{}

	buffer := bytes.NewBufferString("")
	logger := &accessLogJsonMiddleware{
		Logger: log.New(buffer, "", 0),
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

	decoded := &AccessLogJsonRecord{}
	err := json.Unmarshal(buffer.Bytes(), decoded)
	if err != nil {
		t.Fatal(err)
	}

	if decoded.StatusCode != 200 {
		t.Errorf("StatusCode 200 expected, got %d", decoded.StatusCode)
	}
	if decoded.RequestURI != "/" {
		t.Errorf("RequestURI / expected, got %s", decoded.RequestURI)
	}
	if decoded.HttpMethod != "GET" {
		t.Errorf("HttpMethod GET expected, got %s", decoded.HttpMethod)
	}
}
