package rest

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func runTestRequest(t *testing.T, handler *ResourceHandler, method, urlStr, payload string) *httptest.ResponseRecorder {

	urlObj, err := url.Parse(urlStr)
	if err != nil {
		t.Fatal(err)
	}
	r := http.Request{
		Method: method,
		URL:    urlObj,
	}
	r.Header = http.Header{}
	r.Header.Set("Accept-Encoding", "gzip")

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, &r)

	return recorder
}

func codeIs(t *testing.T, r *httptest.ResponseRecorder, expectedCode int) {
	if r.Code != expectedCode {
		t.Errorf("Code %d expected, got: %d", expectedCode, r.Code)
	}
}

func contentTypeIsJson(t *testing.T, r *httptest.ResponseRecorder) {
	ct := r.HeaderMap.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content type 'application/json' expected, got: %s", ct)
	}
}

func contentEncodingIsGzip(t *testing.T, r *httptest.ResponseRecorder) {
	ce := r.HeaderMap.Get("Content-Encoding")
	if ce != "gzip" {
		t.Errorf("Content encoding 'gzip' expected, got: %s", ce)
	}
}

func bodyIs(t *testing.T, r *httptest.ResponseRecorder, expectedBody string) {
	body := r.Body.String()
	if body != expectedBody {
		t.Errorf("Body '%s' expected, got: '%s'", expectedBody, body)
	}
}

func TestHandler(t *testing.T) {

	handler := ResourceHandler{
		DisableJsonIndent: true,
	}
	handler.SetRoutes(
		Route{"GET", "/r/:id",
			func(w *ResponseWriter, r *Request) {
				id := r.PathParam("id")
				w.WriteJson(map[string]string{"Id": id})
			},
		},
		Route{"GET", "/auto-fails",
			func(w *ResponseWriter, r *Request) {
				a := []int{}
				_ = a[0]
			},
		},
		Route{"GET", "/user-error",
			func(w *ResponseWriter, r *Request) {
				Error(w, "My error", 500)
			},
		},
		Route{"GET", "/user-notfound",
			func(w *ResponseWriter, r *Request) {
				NotFound(w, r)
			},
		},
	)

	// valid get resource
	recorder := runTestRequest(t, &handler, "GET", "http://1.2.3.4/r/123", "")
	codeIs(t, recorder, 200)
	contentTypeIsJson(t, recorder)
	bodyIs(t, recorder, `{"Id":"123"}`)

	// auto 405 on undefined route (wrong method)
	recorder = runTestRequest(t, &handler, "DELETE", "http://1.2.3.4/r/123", "")
	codeIs(t, recorder, 405)
	contentTypeIsJson(t, recorder)
	bodyIs(t, recorder, `{"Error":"Method not allowed"}`)

	// auto 404 on undefined route (wrong path)
	recorder = runTestRequest(t, &handler, "GET", "http://1.2.3.4/s/123", "")
	codeIs(t, recorder, 404)
	contentTypeIsJson(t, recorder)
	bodyIs(t, recorder, `{"Error":"Resource not found"}`)

	// auto 500 on unhandled userecorder error
	recorder = runTestRequest(t, &handler, "GET", "http://1.2.3.4/auto-fails", "")
	codeIs(t, recorder, 500)

	// userecorder error
	recorder = runTestRequest(t, &handler, "GET", "http://1.2.3.4/user-error", "")
	codeIs(t, recorder, 500)
	contentTypeIsJson(t, recorder)
	bodyIs(t, recorder, `{"Error":"My error"}`)

	// userecorder notfound
	recorder = runTestRequest(t, &handler, "GET", "http://1.2.3.4/user-notfound", "")
	codeIs(t, recorder, 404)
	contentTypeIsJson(t, recorder)
	bodyIs(t, recorder, `{"Error":"Resource not found"}`)
}

func TestGzip(t *testing.T) {

	handler := ResourceHandler{
		DisableJsonIndent: true,
		EnableGzip:        true,
	}
	handler.SetRoutes(
		Route{"GET", "/r",
			func(w *ResponseWriter, r *Request) {
				w.WriteJson(map[string]string{"Id": "123"})
			},
		},
	)

	recorder := runTestRequest(t, &handler, "GET", "http://1.2.3.4/r", "")
	codeIs(t, recorder, 200)
	contentTypeIsJson(t, recorder)
	contentEncodingIsGzip(t, recorder)
}
