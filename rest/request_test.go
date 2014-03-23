package rest

import (
	"io"
	"net/http"
	"testing"
)

func defaultRequest(method string, urlStr string, body io.Reader, t *testing.T) *Request {
	origReq, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		t.Fatal()
	}
	return &Request{
		origReq,
		nil,
		map[string]interface{}{},
	}
}

func TestRequestUriBase(t *testing.T) {
	req := defaultRequest("GET", "http://localhost", nil, t)
	uriBase := req.UriBase()
	uriString := uriBase.String()

	expected := "http://localhost"
	if uriString != expected {
		t.Error(expected + " was the expected URI base, but instead got " + uriString)
	}
}

func TestRequestUriScheme(t *testing.T) {
	req := defaultRequest("GET", "https://localhost", nil, t)
	uriBase := req.UriBase()

	expected := "https"
	if uriBase.Scheme != expected {
		t.Error(expected + " was the expected scheme, but instead got " + uriBase.Scheme)
	}
}

func TestRequestUriFor(t *testing.T) {
	req := defaultRequest("GET", "http://localhost", nil, t)

	path := "/foo/bar"

	uri := req.UriFor(path)
	if uri.Path != path {
		t.Error(path + " was expected to be the path, but got " + uri.Path)
	}

	expected := "http://localhost/foo/bar"
	if uri.String() != expected {
		t.Error(expected + " was expected, but the returned URI was " + uri.String())
	}
}

func TestRequestUriForParams(t *testing.T) {
	req := defaultRequest("GET", "http://localhost", nil, t)

	params := make(map[string][]string)
	params["id"] = []string{"foo", "bar"}

	uri := req.UriForWithParams("/foo/bar", params)

	expected := "http://localhost/foo/bar?id=foo&id=bar"
	if uri.String() != expected {
		t.Error(expected + " was expected, but the returned URI was " + uri.String())
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
}
