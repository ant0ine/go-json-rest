package rest

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func run_test_request(t *testing.T, handler *ResourceHandler, method, url_str, payload string) *httptest.ResponseRecorder {

	url_obj, err := url.Parse(url_str)
	if err != nil {
		t.Fatal(err)
	}
	r := http.Request{
		Method: method,
		URL:    url_obj,
	}
	r.Header = http.Header{}
	r.Header.Set("Accept-Encoding", "gzip")

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, &r)

	return recorder
}

func code_is(t *testing.T, r *httptest.ResponseRecorder, expected_code int) {
	if r.Code != expected_code {
		t.Errorf("Code %d expected, got: %d", expected_code, r.Code)
	}
}

func content_type_is_json(t *testing.T, r *httptest.ResponseRecorder) {
	ct := r.HeaderMap.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content type 'application/json' expected, got: %s", ct)
	}
}

func content_encoding_is_gzip(t *testing.T, r *httptest.ResponseRecorder) {
	ce := r.HeaderMap.Get("Content-Encoding")
	if ce != "gzip" {
		t.Errorf("Content encoding 'gzip' expected, got: %s", ce)
	}
}

func body_is(t *testing.T, r *httptest.ResponseRecorder, expected_body string) {
	body := r.Body.String()
	if body != expected_body {
		t.Errorf("Body '%s' expected, got: '%s'", expected_body, body)
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
	recorder := run_test_request(t, &handler, "GET", "http://1.2.3.4/r/123", "")
	code_is(t, recorder, 200)
	content_type_is_json(t, recorder)
	body_is(t, recorder, `{"Id":"123"}`)

	// auto 404 on undefined route (wrong method)
	recorder = run_test_request(t, &handler, "DELETE", "http://1.2.3.4/r/123", "")
	code_is(t, recorder, 404)
	content_type_is_json(t, recorder)
	body_is(t, recorder, `{"Error":"Resource not found"}`)

	// auto 404 on undefined route (wrong path)
	recorder = run_test_request(t, &handler, "GET", "http://1.2.3.4/s/123", "")
	code_is(t, recorder, 404)
	content_type_is_json(t, recorder)
	body_is(t, recorder, `{"Error":"Resource not found"}`)

	// auto 500 on unhandled userecorder error
	recorder = run_test_request(t, &handler, "GET", "http://1.2.3.4/auto-fails", "")
	code_is(t, recorder, 500)

	// userecorder error
	recorder = run_test_request(t, &handler, "GET", "http://1.2.3.4/user-error", "")
	code_is(t, recorder, 500)
	content_type_is_json(t, recorder)
	body_is(t, recorder, `{"Error":"My error"}`)

	// userecorder notfound
	recorder = run_test_request(t, &handler, "GET", "http://1.2.3.4/user-notfound", "")
	code_is(t, recorder, 404)
	content_type_is_json(t, recorder)
	body_is(t, recorder, `{"Error":"Resource not found"}`)
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

	recorder := run_test_request(t, &handler, "GET", "http://1.2.3.4/r", "")
	code_is(t, recorder, 200)
	content_type_is_json(t, recorder)
	content_encoding_is_gzip(t, recorder)
}
