package rest

import (
	"github.com/ant0ine/go-json-rest/rest/test"
	"testing"
)

func TestPoweredByMiddleware(t *testing.T) {

	// api with a simple app
	api := NewApi(AppSimple(func(w ResponseWriter, r *Request) {
		w.WriteJson(map[string]string{"Id": "123"})
	}))

	// the middleware to test
	api.Use(&PoweredByMiddleware{
		XPoweredBy: "test",
	})

	// wrap all
	handler := api.MakeHandler()

	req := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	recorded := test.RunRequest(t, handler, req)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()
	recorded.HeaderIs("X-Powered-By", "test")
}
