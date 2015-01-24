package rest

import (
	"github.com/ant0ine/go-json-rest/rest/test"
	"testing"
)

func TestJsonIndentMiddleware(t *testing.T) {

	// api with a simple app
	api := NewApi(AppSimple(func(w ResponseWriter, r *Request) {
		w.WriteJson(map[string]string{"Id": "123"})
	}))

	// the middleware to test
	api.Use(&JsonIndentMiddleware{})

	// wrap all
	handler := api.MakeHandler()

	req := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	recorded := test.RunRequest(t, handler, req)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()
	recorded.BodyIs("{\n  \"Id\": \"123\"\n}")
}
