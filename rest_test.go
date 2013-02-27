package rest

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestHandler(t *testing.T) {

	handler := ResourceHandler{
		DisableJsonIndent: true,
	}
	handler.SetRoutes(
		Route{"GET", "/r/:id", func(w *ResponseWriter, r *Request) {
			id := r.PathParam("id")
			w.WriteJson(map[string]string{"Id": id})
		},
		},
	)

	url, err := url.Parse("http://1.2.3.4/r/123")
	if err != nil {
		t.Fatal(err)
	}
	r := http.Request{
		Method: "GET",
		URL:    url,
	}

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, &r)

	if recorder.Code != 200 {
		t.Error("code 200 expected")
	}
	if recorder.Body.String() != `{"Id":"123"}` {
		t.Error("wrong body")
	}

}
