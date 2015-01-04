package rest

import (
	"github.com/ant0ine/go-json-rest/rest/test"
	"testing"
)

func TestJSONP(t *testing.T) {

	handler := ResourceHandler{
		DisableJsonIndent: true,
		PreRoutingMiddlewares: []Middleware{
			&JsonpMiddleware{},
		},
	}
	handler.SetRoutes(
		&Route{"GET", "/ok",
			func(w ResponseWriter, r *Request) {
				w.WriteJson(map[string]string{"Id": "123"})
			},
		},
		&Route{"GET", "/error",
			func(w ResponseWriter, r *Request) {
				Error(w, "jsonp error", 500)
			},
		},
	)

	recorded := test.RunRequest(t, &handler, test.MakeSimpleRequest("GET", "http://1.2.3.4/ok?callback=parseResponse", nil))
	recorded.CodeIs(200)
	recorded.HeaderIs("Content-Type", "text/javascript")
	recorded.BodyIs("parseResponse({\"Id\":\"123\"})")

	recorded = test.RunRequest(t, &handler, test.MakeSimpleRequest("GET", "http://1.2.3.4/error?callback=parseResponse", nil))
	recorded.CodeIs(500)
	recorded.HeaderIs("Content-Type", "text/javascript")
	recorded.BodyIs("parseResponse({\"Error\":\"jsonp error\"})")
}
