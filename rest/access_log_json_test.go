package rest

import (
	"bytes"
	"encoding/json"
	"github.com/ant0ine/go-json-rest/rest/test"
	"log"
	"testing"
)

func TestAccessLogJsonMiddleware(t *testing.T) {

	api := NewApi()

	// the middlewares stack
	buffer := bytes.NewBufferString("")
	api.Use(&AccessLogJsonMiddleware{
		Logger: log.New(buffer, "", 0),
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

	// log tests
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
