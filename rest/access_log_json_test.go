package rest

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAccessLogJsonMiddleware(t *testing.T) {

	// api with simple app
	api := NewApi(AppSimple(func(w ResponseWriter, r *Request) {
		w.WriteJson(map[string]string{"Id": "123"})
	}))

	// the middlewares stack
	buffer := bytes.NewBufferString("")
	api.Use(&AccessLogJsonMiddleware{
		Logger: log.New(buffer, "", 0),
	})
	api.Use(&TimerMiddleware{})
	api.Use(&RecorderMiddleware{})

	// wrap all
	handler := api.MakeHandler()

	// fake request
	r, _ := http.NewRequest("GET", "http://localhost/", nil)
	r.RemoteAddr = "127.0.0.1:1234"

	// fake writer
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	decoded := &AccessLogJsonRecord{}
	err := json.Unmarshal(buffer.Bytes(), decoded)
	if err != nil {
		t.Fatal(err)
	}

	if decoded.StatusCode != 200 {
		t.Errorf("StatusCode 200 expected, got %d", decoded.StatusCode)
	}
	if decoded.RequestURI != "/" {
		t.Errorf("RequestURI / expected, got %s", decoded.RequestURI)
	}
	if decoded.HttpMethod != "GET" {
		t.Errorf("HttpMethod GET expected, got %s", decoded.HttpMethod)
	}
}
