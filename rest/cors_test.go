package rest

import (
	"net/http"
	"testing"

	"github.com/ant0ine/go-json-rest/rest/test"
)

func TestCorsMiddlewareEmptyAccessControlRequestHeaders(t *testing.T) {
	api := NewApi()

	// the middleware to test
	api.Use(&CorsMiddleware{
		OriginValidator: func(_ string, _ *Request) bool {
			return true
		},
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
		},
		AllowedHeaders: []string{
			"Origin",
			"Referer",
		},
	})

	// wrap all
	handler := api.MakeHandler()

	req, _ := http.NewRequest("OPTIONS", "http://localhost", nil)
	req.Header.Set("Origin", "http://another.host")
	req.Header.Set("Access-Control-Request-Method", "PUT")
	req.Header.Set("Access-Control-Request-Headers", "")

	recorded := test.RunRequest(t, handler, req)
	t.Logf("recorded: %+v\n", recorded.Recorder)
	recorded.CodeIs(200)
	recorded.HeaderIs("Access-Control-Allow-Methods", "GET,POST,PUT")
	recorded.HeaderIs("Access-Control-Allow-Headers", "Origin,Referer")
	recorded.HeaderIs("Access-Control-Allow-Origin", "http://another.host")
}
