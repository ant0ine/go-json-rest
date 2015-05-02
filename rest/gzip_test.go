package rest

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/ant0ine/go-json-rest/rest/test"
)

func TestGzipEnabled(t *testing.T) {

	api := NewApi()

	// the middleware to test
	api.Use(&GzipMiddleware{})

	// router app with success and error paths
	router, err := MakeRouter(
		Get("/ok", func(w ResponseWriter, r *Request) {
			w.WriteJson(map[string]string{"Id": "123"})
		}),
		Get("/error", func(w ResponseWriter, r *Request) {
			Error(w, "gzipped error", 500)
		}),
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
		Get("/ok", func(w ResponseWriter, r *Request) {
			w.WriteJson(map[string]string{"Id": "123"})
		}),
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

func TestGzipMinLength(t *testing.T) {
	api := NewApi()

	// the middleware to test
	api.Use(&GzipMiddleware{
		MinLength: 1024,
	})

	router, _ := MakeRouter(
		Post("/gzip", func(w ResponseWriter, r *Request) {
			defer r.Body.Close()
			io.Copy(w.(http.ResponseWriter), r.Body)
		}),
	)

	api.SetApp(router)
	ts := httptest.NewServer(api.MakeHandler())
	defer ts.Close()

	c := &http.Client{}

	// test payload content length over min length
	payload := map[string]string{
		"data": strings.Repeat("A", 1024),
	}

	req := test.MakeSimpleRequest("POST", ts.URL+"/gzip", payload)
	resp, _ := c.Do(req)

	length, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	if length >= 1024 {
		t.Error("Expect 'Content-Length' should less than 1024 but got", length)
	}

	// test payload content length less than min length
	payload = map[string]string{
		"data": strings.Repeat("A", 1),
	}
	b, _ := json.Marshal(payload)
	contentLength := len(b)

	req = test.MakeSimpleRequest("POST", ts.URL+"/gzip", payload)
	resp, _ = c.Do(req)

	length, _ = strconv.Atoi(resp.Header.Get("Content-Length"))
	if length != contentLength {
		t.Errorf("Expect 'Content-Length' should equal %d but got %d", contentLength, length)
	}
}
