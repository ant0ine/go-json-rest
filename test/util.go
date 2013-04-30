// Utility functions to help writing tests for a Go-Json-Rest app
//
package test

import (
	"net/http"
	"net/http/httptest"
	//	"net/url"
	"testing"
)

func CodeIs(t *testing.T, r *httptest.ResponseRecorder, expectedCode int) {
	if r.Code != expectedCode {
		t.Errorf("Code %d expected, got: %d", expectedCode, r.Code)
	}
}

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
