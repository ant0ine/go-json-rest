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

	// api with simple app
	api := NewApi(AppSimple(func(w ResponseWriter, r *Request) {
		w.WriteJson(map[string]string{"Id": "123"})
	}))

	// the middlewares stack
	buffer := bytes.NewBufferString("")
	api.Use(&AccessLogApacheMiddleware{
		Logger:       log.New(buffer, "", 0),
		Format:       CommonLogFormat,
		textTemplate: nil,
	})
	api.Use(&TimerMiddleware{})
	api.Use(&RecorderMiddleware{})

	// wrap all
	handler := api.MakeHandler()

	// fake request
	r, _ := http.NewRequest("GET", "http://localhost/", nil)
	r.RemoteAddr = "127.0.0.1:1234"

	// fake writer
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	// eg: '127.0.0.1 - - 29/Nov/2014:22:28:34 +0000 "GET / HTTP/1.1" 200 12'
	apacheCommon := regexp.MustCompile(`127.0.0.1 - - \d{2}/\w{3}/\d{4}:\d{2}:\d{2}:\d{2} [+\-]\d{4}\ "GET / HTTP/1.1" 200 12`)

	if !apacheCommon.Match(buffer.Bytes()) {
		t.Errorf("Got: %s", buffer.String())
	}
}
