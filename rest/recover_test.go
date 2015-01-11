package rest

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoverMiddleware(t *testing.T) {

	recov := &recoverMiddleware{
		Logger:                   log.New(ioutil.Discard, "", 0),
		EnableLogAsJson:          false,
		EnableResponseStackTrace: true,
	}

	app := func(w ResponseWriter, r *Request) {
		panic("test")
	}

	handlerFunc := WrapMiddlewares([]Middleware{recov}, app)

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

	// status code
	if recorder.Code != 500 {
		t.Error("Expected a 500 response")
	}

	// payload
	content, err := ioutil.ReadAll(recorder.Body)
	if err != nil {
		t.Fatal(err)
	}
	payload := map[string]string{}
	err = json.Unmarshal(content, &payload)
	if err != nil {
		t.Fatal(err)
	}
	if payload["Error"] == "" {
		t.Errorf("Expected an error message, got: %v", payload)
	}
}
