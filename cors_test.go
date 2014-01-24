package rest

import (
	"errors"
	"github.com/ant0ine/go-json-rest/test"
	"net/http"
	"testing"
)

func MakeCorsRequest(origin string, method string) *http.Request {
	request := test.MakeSimpleRequest("OPTIONS", "http://1.2.3.4/ok", nil)
	request.Header.Set("Origin", origin)
	request.Header.Set("Access-Control-Request-Method", method)
	request.Header.Set("Access-Control-Request-Header", "Accept-Encoding, Content-Type")
	return request
}

func MakeCorsHandler() *ResourceHandler {
	handler := ResourceHandler{
		DisableJsonIndent: true,
	}
	handler.SetRoutes(
		Route{"POST", "/ok",
			func(w *ResponseWriter, r *Request) {
				w.WriteJson(map[string]string{"Success": "total"})
			},
		},
		Route{"DELETE", "/ok",
			func(w *ResponseWriter, r *Request) {
				w.WriteJson(map[string]string{"Success": "total"})
			},
		},
		Route{"PUT", "/ok/:Id",
			func(w *ResponseWriter, r *Request) {
				w.WriteJson(map[string]string{"Success": "total"})
			},
		},
	)
	handler.SetOrigins(
		Origin{Host: "http://example.org"},
	)
	return &handler
}

func TestCorsPreflight(t *testing.T) {
	handler := MakeCorsHandler()
	request := MakeCorsRequest("http://example.org", "POST")
	recorded := test.RunRequest(t, handler, request)
	recorded.CodeIs(200)
	recorded.HeaderIs(HeaderAccessControlAllowOrigin, "http://example.org")
	recorded.HeaderIs(HeaderAccessControlAllowMethods, "POST, DELETE")
}

func TestCorsSplatPreflight(t *testing.T) {
	handler := MakeCorsHandler()
	handler.SetOrigins(
		Origin{Host: "*"},
	)
	request := MakeCorsRequest("http://example.org", "POST")
	recorded := test.RunRequest(t, handler, request)
	recorded.CodeIs(200)
	recorded.HeaderIs(HeaderAccessControlAllowOrigin, "*")
	recorded.HeaderIs(HeaderAccessControlAllowMethods, "POST, DELETE")
}

func TestCorsPreflightAccessControlModify(t *testing.T) {
	handler := MakeCorsHandler()
	handler.SetOrigins(
		Origin{Host: "*", AccessControl: func(req *CorsRequest, headers *CorsResponseHeaders) error {
			headers.AccessControlAllowMethods = []string{"POST"}
			return nil
		}},
	)
	request := MakeCorsRequest("http://example.org", "POST")
	recorded := test.RunRequest(t, handler, request)
	recorded.CodeIs(200)
	recorded.HeaderIs(HeaderAccessControlAllowOrigin, "*")
	recorded.HeaderIs(HeaderAccessControlAllowMethods, "POST")
}

func TestCorsPreflightAccessControlForbid(t *testing.T) {
	handler := MakeCorsHandler()
	handler.SetOrigins(
		Origin{Host: "*", AccessControl: func(req *CorsRequest, headers *CorsResponseHeaders) error {
			return errors.New(`Demo failure`)
		}},
	)
	request := MakeCorsRequest("http://example.org", "POST")
	recorded := test.RunRequest(t, handler, request)
	recorded.CodeIs(400)
	recorded.BodyIs(`{"Error":"Demo failure"}`)
}

func TestCorsWrongPreflight(t *testing.T) {
	handler := MakeCorsHandler()
	request := MakeCorsRequest("http://localhost", "POST")
	recorded := test.RunRequest(t, handler, request)
	recorded.CodeIs(400)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"Error":"CORS origin forbidden"}`)
}

func TestCorsActual(t *testing.T) {
	handler := MakeCorsHandler()
	request := test.MakeSimpleRequest(`POST`, `http://1.2.3.4/ok`, nil)
	request.Header.Set(`Origin`, `http://example.org`)
	recorded := test.RunRequest(t, handler, request)
	recorded.CodeIs(200)
	recorded.HeaderIs(HeaderAccessControlAllowOrigin, `http://example.org`)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"Success":"total"}`)
}

func TestCorsSplatActual(t *testing.T) {
	handler := MakeCorsHandler()
	handler.SetOrigins(
		Origin{Host: "*"},
	)
	request := test.MakeSimpleRequest(`POST`, `http://1.2.3.4/ok`, nil)
	request.Header.Set(`Origin`, `http://example.org`)
	recorded := test.RunRequest(t, handler, request)
	recorded.CodeIs(200)
	recorded.HeaderIs(HeaderAccessControlAllowOrigin, `*`)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"Success":"total"}`)
}
