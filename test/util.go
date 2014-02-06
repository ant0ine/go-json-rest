// Utility functions to help writing tests for a Go-Json-Rest app
//
// Go comes with net/http/httptest to help writing test for an http
// server. When this http server implements a JSON REST API, some basic
// checks end up to be always the same. This test package tries to save
// some typing by providing helpers for this particular use case.
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

func MakeSimpleRequest(method string, urlStr string, payload interface{}) *http.Request {

	s := ""

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

func CodeIs(t *testing.T, r *httptest.ResponseRecorder, expectedCode int) {
	if r.Code != expectedCode {
		t.Errorf("Code %d expected, got: %d", expectedCode, r.Code)
	}
}

// Test the first value for the given headerKey
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

func RunRequest(t *testing.T, handler http.Handler, request *http.Request) *Recorded {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return &Recorded{t, recorder}
}

func (self *Recorded) CodeIs(expectedCode int) {
	CodeIs(self.T, self.Recorder, expectedCode)
}

func (self *Recorded) HeaderIs(headerKey, expectedValue string) {
	HeaderIs(self.T, self.Recorder, headerKey, expectedValue)
}

func (self *Recorded) ContentTypeIsJson() {
	self.HeaderIs("Content-Type", "application/json")
}

func (self *Recorded) ContentEncodingIsGzip() {
	self.HeaderIs("Content-Encoding", "gzip")
}

func (self *Recorded) BodyIs(expectedBody string) {
	BodyIs(self.T, self.Recorder, expectedBody)
}

func (self *Recorded) DecodeJsonPayload(v interface{}) error {
	return DecodeJsonPayload(self.Recorder, v)
}
