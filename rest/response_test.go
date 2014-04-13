package rest

import (
	"testing"

	"github.com/ant0ine/go-json-rest/rest/test"
)

func TestResponseNotIndent(t *testing.T) {

	writer := responseWriter{
		nil,
		false,
		false,
		xPoweredByDefault,
	}

	got, err := writer.EncodeJson(map[string]bool{"test": true})
	if err != nil {
		t.Error(err.Error())
	}
	gotStr := string(got)
	expected := "{\"test\":true}"
	if gotStr != expected {
		t.Error(expected + " was the expected, but instead got " + gotStr)
	}
}

func TestResponseIndent(t *testing.T) {
	testXPoweredBy(t, &ResourceHandler{}, xPoweredByDefault)
}

func TestXPoweredByCustom(t *testing.T) {
	testXPoweredBy(t, &ResourceHandler{XPoweredBy: "foo"}, "foo")
}

func TestXPoweredByDisabled(t *testing.T) {
	testXPoweredBy(t, &ResourceHandler{DisableXPoweredBy: true}, "")
}

func testXPoweredBy(t *testing.T, rh *ResourceHandler, expected string) {
	rh.SetRoutes(
		&Route{"GET", "/r/:id",
			func(w ResponseWriter, r *Request) {
				id := r.PathParam("id")
				w.WriteJson(map[string]string{"Id": id})
			},
		},
	)
	recorded := test.RunRequest(t, rh, test.MakeSimpleRequest("GET", "http://1.2.3.4/r/123", nil))
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()
	recorded.HeaderIs("X-Powered-By", expected)
}
