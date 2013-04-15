package rest

import (
	"net/http"
	"net/url"
	"testing"
)

func defaultRequest(uri string, method string, t *testing.T) Request {
	urlObj, err := url.Parse(uri)
	if err != nil {
		t.Fatal()
	}
	origReq := http.Request{
		Method: method,
		URL:    urlObj,
		Host:   "localhost",
	}
	req := Request{&origReq, nil}
	return req
}

func TestRequestUriBase(t *testing.T) {
	req := defaultRequest("http://localhost", "GET", t)
	uriBase := req.UriBase()
	uriString := uriBase.String()

	expected := "http://localhost"
	if uriString != expected {
		t.Error(expected + " was the expected URI base, but instead got " + uriString)
	}
}

func TestRequestUriScheme(t *testing.T) {
	req := defaultRequest("https://localhost", "GET", t)
	uriBase := req.UriBase()

	expected := "https"
	if uriBase.Scheme != expected {
		t.Error(expected + " was the expected scheme, but instead got " + uriBase.Scheme)
	}
}

func TestRequestUriFor(t *testing.T) {
	req := defaultRequest("http://locahost", "GET", t)

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
	req := defaultRequest("http://localhost", "GET", t)

	params := make(map[string][]string)
	params["id"] = []string{"foo", "bar"}

	uri := req.UriForWithParams("/foo/bar", params)

	expected := "http://localhost/foo/bar?id=foo&id=bar"
	if uri.String() != expected {
		t.Error(expected + " was expected, but the returned URI was " + uri.String())
	}
}
