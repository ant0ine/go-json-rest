package rest

import (
	"github.com/ant0ine/go-json-rest/rest/test"
	"io/ioutil"
	"log"
	"testing"
)

func TestHandler(t *testing.T) {

	handler := ResourceHandler{
		DisableJsonIndent: true,
		// make the test output less verbose by discarding the error log
		ErrorLogger: log.New(ioutil.Discard, "", 0),
	}
	handler.SetRoutes(
		&Route{"GET", "/r/:id",
			func(w ResponseWriter, r *Request) {
				id := r.PathParam("id")
				w.WriteJson(map[string]string{"Id": id})
			},
		},
		&Route{"POST", "/r/:id",
			func(w ResponseWriter, r *Request) {
				// JSON echo
				data := map[string]string{}
				err := r.DecodeJsonPayload(&data)
				if err != nil {
					t.Fatal(err)
				}
				w.WriteJson(data)
			},
		},
		&Route{"GET", "/auto-fails",
			func(w ResponseWriter, r *Request) {
				a := []int{}
				_ = a[0]
			},
		},
		&Route{"GET", "/user-error",
			func(w ResponseWriter, r *Request) {
				Error(w, "My error", 500)
			},
		},
		&Route{"GET", "/user-notfound",
			func(w ResponseWriter, r *Request) {
				NotFound(w, r)
			},
		},
	)

	// valid get resource
	recorded := test.RunRequest(t, &handler, test.MakeSimpleRequest("GET", "http://1.2.3.4/r/123", nil))
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"Id":"123"}`)

	// valid post resource
	recorded = test.RunRequest(t, &handler, test.MakeSimpleRequest(
		"POST", "http://1.2.3.4/r/123", &map[string]string{"Test": "Test"}))
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"Test":"Test"}`)

	// broken Content-Type post resource
	request := test.MakeSimpleRequest("POST", "http://1.2.3.4/r/123", &map[string]string{"Test": "Test"})
	request.Header.Set("Content-Type", "text/html")
	recorded = test.RunRequest(t, &handler, request)
	recorded.CodeIs(415)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"Error":"Bad Content-Type or charset, expected 'application/json'"}`)

	// broken Content-Type post resource
	request = test.MakeSimpleRequest("POST", "http://1.2.3.4/r/123", &map[string]string{"Test": "Test"})
	request.Header.Set("Content-Type", "application/json; charset=ISO-8859-1")
	recorded = test.RunRequest(t, &handler, request)
	recorded.CodeIs(415)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"Error":"Bad Content-Type or charset, expected 'application/json'"}`)

	// Content-Type post resource with charset
	request = test.MakeSimpleRequest("POST", "http://1.2.3.4/r/123", &map[string]string{"Test": "Test"})
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	recorded = test.RunRequest(t, &handler, request)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"Test":"Test"}`)

	// auto 405 on undefined route (wrong method)
	recorded = test.RunRequest(t, &handler, test.MakeSimpleRequest("DELETE", "http://1.2.3.4/r/123", nil))
	recorded.CodeIs(405)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"Error":"Method not allowed"}`)

	// auto 404 on undefined route (wrong path)
	recorded = test.RunRequest(t, &handler, test.MakeSimpleRequest("GET", "http://1.2.3.4/s/123", nil))
	recorded.CodeIs(404)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"Error":"Resource not found"}`)

	// auto 500 on unhandled userecorder error
	recorded = test.RunRequest(t, &handler, test.MakeSimpleRequest("GET", "http://1.2.3.4/auto-fails", nil))
	recorded.CodeIs(500)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"Error":"Internal Server Error"}`)

	// userecorder error
	recorded = test.RunRequest(t, &handler, test.MakeSimpleRequest("GET", "http://1.2.3.4/user-error", nil))
	recorded.CodeIs(500)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"Error":"My error"}`)

	// userecorder notfound
	recorded = test.RunRequest(t, &handler, test.MakeSimpleRequest("GET", "http://1.2.3.4/user-notfound", nil))
	recorded.CodeIs(404)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"Error":"Resource not found"}`)
}
