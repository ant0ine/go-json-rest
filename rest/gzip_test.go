package rest

import (
	"github.com/ant0ine/go-json-rest/rest/test"
	"testing"
)

func TestGzipEnabled(t *testing.T) {

	api := NewApi()

	// the middleware to test
	api.Use(&GzipMiddleware{})

	// router app with success and error paths
	router, err := MakeRouter(
		&Route{"GET", "/ok",
			func(w ResponseWriter, r *Request) {
				w.WriteJson(map[string]string{"Id": "123"})
			},
		},
		&Route{"GET", "/error",
			func(w ResponseWriter, r *Request) {
				Error(w, "gzipped error", 500)
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	api.SetApp(router)

	// wrap all
	handler := api.MakeHandler()

	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://localhost/ok", nil))
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()
	recorded.ContentEncodingIsGzip()
	recorded.HeaderIs("Vary", "Accept-Encoding")

	recorded = test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://localhost/error", nil))
	recorded.CodeIs(500)
	recorded.ContentTypeIsJson()
	recorded.ContentEncodingIsGzip()
	recorded.HeaderIs("Vary", "Accept-Encoding")
}

func TestGzipDisabled(t *testing.T) {

	api := NewApi()

	// router app with success and error paths
	router, err := MakeRouter(
		&Route{"GET", "/ok",
			func(w ResponseWriter, r *Request) {
				w.WriteJson(map[string]string{"Id": "123"})
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	api.SetApp(router)
	handler := api.MakeHandler()

	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://localhost/ok", nil))
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()
	recorded.HeaderIs("Content-Encoding", "")
	recorded.HeaderIs("Vary", "")
}
