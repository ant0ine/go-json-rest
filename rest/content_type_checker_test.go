package rest

import (
	"github.com/ant0ine/go-json-rest/rest/test"
	"testing"
)

func TestContentTypeCheckerMiddleware(t *testing.T) {

	// api with a simple app
	api := NewApi(AppSimple(func(w ResponseWriter, r *Request) {
		w.WriteJson(map[string]string{"Id": "123"})
	}))

	// the middleware to test
	api.Use(&ContentTypeCheckerMiddleware{})

	// wrap all
	handler := api.MakeHandler()

	// no payload, no content length, no check
	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://localhost/", nil))
	recorded.CodeIs(200)

	// JSON payload with correct content type
	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("POST", "http://localhost/", map[string]string{"Id": "123"}))
	recorded.CodeIs(200)

	// JSON payload with correct content type specifying the utf-8 charset
	req := test.MakeSimpleRequest("POST", "http://localhost/", map[string]string{"Id": "123"})
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	recorded = test.RunRequest(t, handler, req)
	recorded.CodeIs(200)

	// JSON payload with incorrect content type
	req = test.MakeSimpleRequest("POST", "http://localhost/", map[string]string{"Id": "123"})
	req.Header.Set("Content-Type", "text/x-json")
	recorded = test.RunRequest(t, handler, req)
	recorded.CodeIs(415)
}
