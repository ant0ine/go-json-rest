package rest

import (
	"github.com/ant0ine/go-json-rest/test"
	"net/http"
	"net/url"
	"testing"
)

func makeRequest(method, urlStr, payload string) *http.Request {

	urlObj, err := url.Parse(urlStr)
	if err != nil {
		panic(err)
	}
	r := http.Request{
		Method: method,
		URL:    urlObj,
	}
	r.Header = http.Header{}
	r.Header.Set("Accept-Encoding", "gzip")
	return &r
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
	recorded := test.RunRequest(t, &handler, makeRequest("GET", "http://1.2.3.4/r/123", ""))
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"Id":"123"}`)

	// auto 405 on undefined route (wrong method)
	recorded = test.RunRequest(t, &handler, makeRequest("DELETE", "http://1.2.3.4/r/123", ""))
	recorded.CodeIs(405)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"Error":"Method not allowed"}`)

	// auto 404 on undefined route (wrong path)
	recorded = test.RunRequest(t, &handler, makeRequest("GET", "http://1.2.3.4/s/123", ""))
	recorded.CodeIs(404)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"Error":"Resource not found"}`)

	// auto 500 on unhandled userecorder error
	recorded = test.RunRequest(t, &handler, makeRequest("GET", "http://1.2.3.4/auto-fails", ""))
	recorded.CodeIs(500)

	// userecorder error
	recorded = test.RunRequest(t, &handler, makeRequest("GET", "http://1.2.3.4/user-error", ""))
	recorded.CodeIs(500)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"Error":"My error"}`)

	// userecorder notfound
	recorded = test.RunRequest(t, &handler, makeRequest("GET", "http://1.2.3.4/user-notfound", ""))
	recorded.CodeIs(404)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`{"Error":"Resource not found"}`)
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

	recorded := test.RunRequest(t, &handler, makeRequest("GET", "http://1.2.3.4/r", ""))
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()
	recorded.ContentEncodingIsGzip()
}
