package rest

import (
	"github.com/ant0ine/go-json-rest/rest/test"
	"testing"
)

func TestApiNoAppNoMiddleware(t *testing.T) {

	api := NewApi()
	if api == nil {
		t.Fatal("Api object must be instantiated")
	}

	handler := api.MakeHandler()
	if handler == nil {
		t.Fatal("the http.Handler must be have been create")
	}

	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://localhost/", nil))
	recorded.CodeIs(200)
}

func TestApiSimpleAppNoMiddleware(t *testing.T) {

	api := NewApi()
	if api == nil {
		t.Fatal("Api object must be instantiated")
	}

	api.SetApp(AppSimple(func(w ResponseWriter, r *Request) {
		w.WriteJson(map[string]string{"Id": "123"})
	}))

	handler := api.MakeHandler()
	if handler == nil {
		t.Fatal("the http.Handler must be have been create")
	}

	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://localhost/", nil))
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"Id":"123"}`)
}
