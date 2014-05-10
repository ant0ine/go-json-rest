package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// MakeSimpleRequest returns a http.Request. The returned request object can be
// further prepared by adding headers and query string parmaters, for instance.
func MakeSimpleRequest(method string, urlStr string, payload interface{}) *http.Request {
	var s string

	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			panic(err)
		}
		s = fmt.Sprintf("%s", b)
	}

	r, err := http.NewRequest(method, urlStr, strings.NewReader(s))
	if err != nil {
		panic(err)
	}
	r.Header.Set("Accept-Encoding", "gzip")
	if payload != nil {
		r.Header.Set("Content-Type", "application/json")
	}

	return r
}

// CodeIs compares the rescorded status code
func CodeIs(t *testing.T, r *httptest.ResponseRecorder, expectedCode int) {
	if r.Code != expectedCode {
		t.Errorf("Code %d expected, got: %d", expectedCode, r.Code)
	}
}

// HeaderIs tests the first value for the given headerKey
func HeaderIs(t *testing.T, r *httptest.ResponseRecorder, headerKey, expectedValue string) {
	value := r.HeaderMap.Get(headerKey)
	if value != expectedValue {
		t.Errorf(
			"%s: %s expected, got: %s",
			headerKey,
			expectedValue,
			value,
		)
	}
}

func ContentTypeIsJson(t *testing.T, r *httptest.ResponseRecorder) {
	HeaderIs(t, r, "Content-Type", "application/json")
}

func ContentEncodingIsGzip(t *testing.T, r *httptest.ResponseRecorder) {
	HeaderIs(t, r, "Content-Encoding", "gzip")
}

func BodyIs(t *testing.T, r *httptest.ResponseRecorder, expectedBody string) {
	body := r.Body.String()
	if body != expectedBody {
		t.Errorf("Body '%s' expected, got: '%s'", expectedBody, body)
	}
}

func DecodeJsonPayload(r *httptest.ResponseRecorder, v interface{}) error {
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, v)
	if err != nil {
		return err
	}
	return nil
}

type Recorded struct {
	T        *testing.T
	Recorder *httptest.ResponseRecorder
}

// RunRequest runs a HTTP request through the given handler
func RunRequest(t *testing.T, handler http.Handler, request *http.Request) *Recorded {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return &Recorded{t, recorder}
}

func (rd *Recorded) CodeIs(expectedCode int) {
	CodeIs(rd.T, rd.Recorder, expectedCode)
}

func (rd *Recorded) HeaderIs(headerKey, expectedValue string) {
	HeaderIs(rd.T, rd.Recorder, headerKey, expectedValue)
}

func (rd *Recorded) ContentTypeIsJson() {
	rd.HeaderIs("Content-Type", "application/json")
}

func (rd *Recorded) ContentEncodingIsGzip() {
	rd.HeaderIs("Content-Encoding", "gzip")
}

func (rd *Recorded) BodyIs(expectedBody string) {
	BodyIs(rd.T, rd.Recorder, expectedBody)
}

func (rd *Recorded) DecodeJsonPayload(v interface{}) error {
	return DecodeJsonPayload(rd.Recorder, v)
}
