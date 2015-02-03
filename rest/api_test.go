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

func TestDevStack(t *testing.T) {

	api := NewApi()
	api.Use(DefaultDevStack...)
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
	recorded.BodyIs("{\n  \"Id\": \"123\"\n}")
}

func TestProdStack(t *testing.T) {

	api := NewApi()
	api.Use(DefaultProdStack...)
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
	recorded.ContentEncodingIsGzip()
}

func TestCommonStack(t *testing.T) {

	api := NewApi()
	api.Use(DefaultCommonStack...)
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
