package rest

import (
	"testing"

	"github.com/ant0ine/go-json-rest/rest/test"
)

func TestJsonpMiddleware(t *testing.T) {

	api := NewApi()

	// the middleware to test
	api.Use(&JsonpMiddleware{})

	// router app with success and error paths
	router, err := MakeRouter(
		Get("/ok", func(w ResponseWriter, r *Request) {
			w.WriteJson(map[string]string{"Id": "123"})
		}),
		Get("/error", func(w ResponseWriter, r *Request) {
			Error(w, "jsonp error", 500)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	api.SetApp(router)

	// wrap all
	handler := api.MakeHandler()

	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://localhost/ok?callback=parseResponse", nil))
	recorded.CodeIs(200)
	recorded.HeaderIs("Content-Type", "text/javascript")
	recorded.HeaderIs("Content-Disposition", "filename=f.txt")
	recorded.HeaderIs("X-Content-Type-Options", "nosniff")
	recorded.BodyIs("/**/parseResponse({\"Id\":\"123\"})")

	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://localhost/error?callback=parseResponse", nil))
	recorded.CodeIs(500)
	recorded.HeaderIs("Content-Type", "text/javascript")
	recorded.HeaderIs("Content-Disposition", "filename=f.txt")
	recorded.HeaderIs("X-Content-Type-Options", "nosniff")
	recorded.BodyIs("/**/parseResponse({\"Error\":\"jsonp error\"})")
}
