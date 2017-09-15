package rest

import (
	"testing"

	"github.com/ant0ine/go-json-rest/rest/test"
)

func CustomError(r *Request, error string, code int) interface{} {
	// r = nil when using test requests
	var header string
	switch code {
	case 400:
		header = "Bad Input"
		break
	case 404:
		header = "Not Found"
		break
	default:
		header = "API Error"
	}

	return map[string]interface{}{
		"error": map[string]interface{}{
			"header":  header,
			"code":    code,
			"message": error,
		},
	}
}

func TestResponseNotIndent(t *testing.T) {

	writer := responseWriter{
		nil,
		false,
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

// The following tests could instantiate only the responseWriter,
// but using the Api object allows to use the rest/test utilities,
// and make the tests easier to write.

func TestWriteJsonResponse(t *testing.T) {

	api := NewApi()
	api.SetApp(AppSimple(func(w ResponseWriter, r *Request) {
		w.WriteJson(map[string]string{"Id": "123"})
	}))

	recorded := test.RunRequest(t, api.MakeHandler(), test.MakeSimpleRequest("GET", "http://localhost/", nil))
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()
	recorded.BodyIs("{\"Id\":\"123\"}")
}

func TestErrorResponse(t *testing.T) {

	api := NewApi()
	api.SetApp(AppSimple(func(w ResponseWriter, r *Request) {
		Error(w, "test", 500)
	}))

	recorded := test.RunRequest(t, api.MakeHandler(), test.MakeSimpleRequest("GET", "http://localhost/", nil))
	recorded.CodeIs(500)
	recorded.ContentTypeIsJson()
	recorded.BodyIs("{\"Error\":\"test\"}")
}

func TestCustomErrorResponse(t *testing.T) {

	api := NewApi()
	ErrorFunc = CustomError

	api.SetApp(AppSimple(func(w ResponseWriter, r *Request) {
		Error(w, "test", 500)
	}))

	recorded := test.RunRequest(t, api.MakeHandler(), test.MakeSimpleRequest("GET", "http://localhost/", nil))
	recorded.CodeIs(500)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"error":{"code":500,"header":"API Error","message":"test"}}`)

	// reset the package variable to not effect other tests
	ErrorFunc = nil
}

func TestNotFoundResponse(t *testing.T) {

	api := NewApi()
	api.SetApp(AppSimple(func(w ResponseWriter, r *Request) {
		NotFound(w, r)
	}))

	recorded := test.RunRequest(t, api.MakeHandler(), test.MakeSimpleRequest("GET", "http://localhost/", nil))
	recorded.CodeIs(404)
	recorded.ContentTypeIsJson()
	recorded.BodyIs("{\"Error\":\"Resource not found\"}")
}

func TestCustomNotFoundResponse(t *testing.T) {

	api := NewApi()
	ErrorFunc = CustomError

	api.SetApp(AppSimple(func(w ResponseWriter, r *Request) {
		NotFound(w, r)
	}))

	recorded := test.RunRequest(t, api.MakeHandler(), test.MakeSimpleRequest("GET", "http://localhost/", nil))
	recorded.CodeIs(404)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"error":{"code":404,"header":"Not Found","message":"Resource not found"}}`)

	// reset the package variable to not effect other tests
	ErrorFunc = nil
}
