package rest

import (
	"bytes"
	"github.com/ant0ine/go-json-rest/rest/test"
	"log"
	"regexp"
	"testing"
)

func TestAccessLogApacheMiddleware(t *testing.T) {

	api := NewApi()

	// the middlewares stack
	buffer := bytes.NewBufferString("")
	api.Use(&AccessLogApacheMiddleware{
		Logger:       log.New(buffer, "", 0),
		Format:       CommonLogFormat,
		textTemplate: nil,
	})
	api.Use(&TimerMiddleware{})
	api.Use(&RecorderMiddleware{})

	// a simple app
	api.SetApp(AppSimple(func(w ResponseWriter, r *Request) {
		w.WriteJson(map[string]string{"Id": "123"})
	}))

	// wrap all
	handler := api.MakeHandler()

	req := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	recorded := test.RunRequest(t, handler, req)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	// log tests, eg: '127.0.0.1 - - 29/Nov/2014:22:28:34 +0000 "GET / HTTP/1.1" 200 12'
	apacheCommon := regexp.MustCompile(`127.0.0.1 - - \d{2}/\w{3}/\d{4}:\d{2}:\d{2}:\d{2} [+\-]\d{4}\ "GET / HTTP/1.1" 200 12`)

	if !apacheCommon.Match(buffer.Bytes()) {
		t.Errorf("Got: %s", buffer.String())
	}
}

func TestAccessLogApacheMiddlewareMissingData(t *testing.T) {

	api := NewApi()

	// the uncomplete middlewares stack
	buffer := bytes.NewBufferString("")
	api.Use(&AccessLogApacheMiddleware{
		Logger:       log.New(buffer, "", 0),
		Format:       CommonLogFormat,
		textTemplate: nil,
	})

	// a simple app
	api.SetApp(AppSimple(func(w ResponseWriter, r *Request) {
		w.WriteJson(map[string]string{"Id": "123"})
	}))

	// wrap all
	handler := api.MakeHandler()

	req := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	recorded := test.RunRequest(t, handler, req)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	// not much to log when the Env data is missing, but this should still work
	apacheCommon := regexp.MustCompile(` - -  "GET / HTTP/1.1" 0 -`)

	if !apacheCommon.Match(buffer.Bytes()) {
		t.Errorf("Got: %s", buffer.String())
	}
}
