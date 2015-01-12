package rest

import (
	"github.com/ant0ine/go-json-rest/rest/test"
	"net/http"
	"testing"
)

func TestStatus(t *testing.T) {

	// the middlewares
	recorder := &recorderMiddleware{}
	timer := &timerMiddleware{}
	status := &statusMiddleware{}

	// the app
	app := func(w ResponseWriter, r *Request) {
		w.WriteJson(status.GetStatus())
	}

	// wrap all
	handler := http.HandlerFunc(adapterFunc(WrapMiddlewares([]Middleware{status, timer, recorder}, app)))

	// one request
	recorded := test.RunRequest(t, &handler, test.MakeSimpleRequest("GET", "http://1.2.3.4/1", nil))
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	// another request
	recorded = test.RunRequest(t, &handler, test.MakeSimpleRequest("GET", "http://1.2.3.4/2", nil))
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

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
