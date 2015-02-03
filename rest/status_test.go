package rest

import (
	"github.com/ant0ine/go-json-rest/rest/test"
	"testing"
)

func TestStatusMiddleware(t *testing.T) {

	api := NewApi()

	// the middlewares
	status := &StatusMiddleware{}
	api.Use(status)
	api.Use(&TimerMiddleware{})
	api.Use(&RecorderMiddleware{})

	// an app that return the Status
	api.SetApp(AppSimple(func(w ResponseWriter, r *Request) {
		w.WriteJson(status.GetStatus())
	}))

	// wrap all
	handler := api.MakeHandler()

	// one request
	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://localhost/1", nil))
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	// another request
	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://localhost/2", nil))
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	// payload
	payload := map[string]interface{}{}
	err := recorded.DecodeJsonPayload(&payload)
	if err != nil {
		t.Fatal(err)
	}

	if payload["Pid"] == nil {
		t.Error("Expected a non nil Pid")
	}

	if payload["TotalCount"].(float64) != 1 {
		t.Errorf("TotalCount 1 Expected, got: %f", payload["TotalCount"].(float64))
	}

	if payload["StatusCodeCount"].(map[string]interface{})["200"].(float64) != 1 {
		t.Errorf("StatusCodeCount 200 1 Expected, got: %f", payload["StatusCodeCount"].(map[string]interface{})["200"].(float64))
	}
}
