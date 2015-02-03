package rest

import (
	"github.com/ant0ine/go-json-rest/rest/test"
	"io/ioutil"
	"log"
	"testing"
)

func TestRecoverMiddleware(t *testing.T) {

	api := NewApi()

	// the middleware to test
	api.Use(&RecoverMiddleware{
		Logger:                   log.New(ioutil.Discard, "", 0),
		EnableLogAsJson:          false,
		EnableResponseStackTrace: true,
	})

	// a simple app that fails
	api.SetApp(AppSimple(func(w ResponseWriter, r *Request) {
		panic("test")
	}))

	// wrap all
	handler := api.MakeHandler()

	req := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	recorded := test.RunRequest(t, handler, req)
	recorded.CodeIs(500)
	recorded.ContentTypeIsJson()

	// payload
	payload := map[string]string{}
	err := recorded.DecodeJsonPayload(&payload)
	if err != nil {
		t.Fatal(err)
	}
	if payload["Error"] == "" {
		t.Errorf("Expected an error message, got: %v", payload)
	}
}
