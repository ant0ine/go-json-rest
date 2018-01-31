package test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/http/httptest"
	"strings"
)

// A TestReporter is an interface that is used to report errors in tests.
// It can be satisfied with *testing.T.
type TestReporter interface {
	Errorf(format string, args ...interface{})
}

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
func CodeIs(t TestReporter, r *httptest.ResponseRecorder, expectedCode int) {
	if r.Code != expectedCode {
		t.Errorf("Code %d expected, got: %d", expectedCode, r.Code)
	}
}

// HeaderIs tests the first value for the given headerKey
func HeaderIs(t TestReporter, r *httptest.ResponseRecorder, headerKey, expectedValue string) {
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

func ContentTypeIsJson(t TestReporter, r *httptest.ResponseRecorder) {

	mediaType, params, _ := mime.ParseMediaType(r.HeaderMap.Get("Content-Type"))
	charset := params["charset"]

	if mediaType != "application/json" {
		t.Errorf(
			"Content-Type media type: application/json expected, got: %s",
			mediaType,
		)
	}

	if charset != "" && strings.ToUpper(charset) != "UTF-8" {
		t.Errorf(
			"Content-Type charset: must be empty or UTF-8, got: %s",
			charset,
		)
	}
}

func ContentEncodingIsGzip(t TestReporter, r *httptest.ResponseRecorder) {
	HeaderIs(t, r, "Content-Encoding", "gzip")
}

func BodyIs(t TestReporter, r *httptest.ResponseRecorder, expectedBody string) {
	body, err := DecodedBody(r)
	if err != nil {
		t.Errorf("Body '%s' expected, got error: '%s'", expectedBody, err)
	}
	if string(body) != expectedBody {
		t.Errorf("Body '%s' expected, got: '%s'", expectedBody, body)
	}
}

func DecodeJsonPayload(r *httptest.ResponseRecorder, v interface{}) error {
	content, err := DecodedBody(r)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, v)
	if err != nil {
		return err
	}
	return nil
}

// DecodedBody returns the entire body read from r.Body, with it
// gunzipped if Content-Encoding is set to gzip
func DecodedBody(r *httptest.ResponseRecorder) ([]byte, error) {
	if r.Header().Get("Content-Encoding") != "gzip" {
		return ioutil.ReadAll(r.Body)
	}
	dec, err := gzip.NewReader(r.Body)
	if err != nil {
		return nil, err
	}
	b := new(bytes.Buffer)
	if _, err = io.Copy(b, dec); err != nil {
		return nil, err
	}
	if err = dec.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

type Recorded struct {
	T        TestReporter
	Recorder *httptest.ResponseRecorder
}

// RunRequest runs a HTTP request through the given handler
func RunRequest(t TestReporter, handler http.Handler, request *http.Request) *Recorded {
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
	ContentTypeIsJson(rd.T, rd.Recorder)
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

func (rd *Recorded) DecodedBody() ([]byte, error) {
	return DecodedBody(rd.Recorder)
}
