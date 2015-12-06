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
		Get("/r/:id", func(w ResponseWriter, r *Request) {
			id := r.PathParam("id")
			w.WriteJson(map[string]string{"Id": id})
		}),
		Post("/r/:id", func(w ResponseWriter, r *Request) {
			// JSON echo
			data := map[string]string{}
			err := r.DecodeJsonPayload(&data)
			if err != nil {
				t.Fatal(err)
			}
			w.WriteJson(data)
		}),
	)

	// valid get resource
	recorded := test.RunRequest(t, &handler, test.MakeSimpleRequest("GET", "http://1.2.3.4/r/123", nil))
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"Id":"123"}`)

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
}
