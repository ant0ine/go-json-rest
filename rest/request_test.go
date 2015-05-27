package rest

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func defaultRequest(method string, urlStr string, body io.Reader, t *testing.T) *Request {
	origReq, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		t.Fatal(err)
	}
	return &Request{
		origReq,
		nil,
		map[string]interface{}{},
	}
}

func TestRequestEmptyJson(t *testing.T) {
	req := defaultRequest("POST", "http://localhost", strings.NewReader(""), t)
	err := req.DecodeJsonPayload(nil)
	if err != ErrJsonPayloadEmpty {
		t.Error("Expected ErrJsonPayloadEmpty")
	}
}

func TestRequestBaseUrl(t *testing.T) {
	req := defaultRequest("GET", "http://localhost", nil, t)
	urlBase := req.BaseUrl()
	urlString := urlBase.String()

	expected := "http://localhost"
	if urlString != expected {
		t.Error(expected + " was the expected URL base, but instead got " + urlString)
	}
}

func TestRequestUrlScheme(t *testing.T) {
	req := defaultRequest("GET", "https://localhost", nil, t)
	urlBase := req.BaseUrl()

	expected := "https"
	if urlBase.Scheme != expected {
		t.Error(expected + " was the expected scheme, but instead got " + urlBase.Scheme)
	}
}

func TestRequestUrlFor(t *testing.T) {
	req := defaultRequest("GET", "http://localhost", nil, t)

	path := "/foo/bar"

	urlObj := req.UrlFor(path, nil)
	if urlObj.Path != path {
		t.Error(path + " was expected to be the path, but got " + urlObj.Path)
	}

	expected := "http://localhost/foo/bar"
	if urlObj.String() != expected {
		t.Error(expected + " was expected, but the returned URL was " + urlObj.String())
	}
}

func TestRequestUrlForQueryString(t *testing.T) {
	req := defaultRequest("GET", "http://localhost", nil, t)

	params := map[string][]string{
		"id": []string{"foo", "bar"},
	}

	urlObj := req.UrlFor("/foo/bar", params)

	expected := "http://localhost/foo/bar?id=foo&id=bar"
	if urlObj.String() != expected {
		t.Error(expected + " was expected, but the returned URL was " + urlObj.String())
	}
}

func TestCorsInfoSimpleCors(t *testing.T) {
	req := defaultRequest("GET", "http://localhost", nil, t)
	req.Request.Header.Set("Origin", "http://another.host")

	corsInfo := req.GetCorsInfo()
	if corsInfo == nil {
		t.Error("Expected non nil CorsInfo")
	}
	if corsInfo.IsCors == false {
		t.Error("This is a CORS request")
	}
	if corsInfo.IsPreflight == true {
		t.Error("This is not a Preflight request")
	}
}

func TestCorsInfoNullOrigin(t *testing.T) {
	req := defaultRequest("GET", "http://localhost", nil, t)
	req.Request.Header.Set("Origin", "null")

	corsInfo := req.GetCorsInfo()
	if corsInfo == nil {
		t.Error("Expected non nil CorsInfo")
	}
	if corsInfo.IsCors == false {
		t.Error("This is a CORS request")
	}
	if corsInfo.IsPreflight == true {
		t.Error("This is not a Preflight request")
	}
	if corsInfo.OriginUrl != nil {
		t.Error("OriginUrl cannot be set")
	}
}

func TestCorsInfoPreflightCors(t *testing.T) {
	req := defaultRequest("OPTIONS", "http://localhost", nil, t)
	req.Request.Header.Set("Origin", "http://another.host")

	corsInfo := req.GetCorsInfo()
	if corsInfo == nil {
		t.Error("Expected non nil CorsInfo")
	}
	if corsInfo.IsCors == false {
		t.Error("This is a CORS request")
	}
	if corsInfo.IsPreflight == true {
		t.Error("This is NOT a Preflight request")
	}

	// Preflight must have the Access-Control-Request-Method header
	req.Request.Header.Set("Access-Control-Request-Method", "PUT")
	corsInfo = req.GetCorsInfo()
	if corsInfo == nil {
		t.Error("Expected non nil CorsInfo")
	}
	if corsInfo.IsCors == false {
		t.Error("This is a CORS request")
	}
	if corsInfo.IsPreflight == false {
		t.Error("This is a Preflight request")
	}
	if corsInfo.Origin != "http://another.host" {
		t.Error("Origin must be identical to the header value")
	}
	if corsInfo.OriginUrl == nil {
		t.Error("OriginUrl must be set")
	}
}
