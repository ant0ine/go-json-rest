package rest

import (
	"net/url"
	"strings"
	"testing"
)

func TestFindRouteAPI(t *testing.T) {

	r := router{
		routes: []*Route{
			&Route{
				HttpMethod: "GET",
				PathExp:    "/",
			},
		},
	}

	err := r.start()
	if err != nil {
		t.Fatal()
	}

	// full url string
	input := "http://example.org/"
	route, params, pathMatched, err := r.findRoute("GET", input)
	if err != nil {
		t.Fatal()
	}
	if route.PathExp != "/" {
		t.Error()
	}
	if len(params) != 0 {
		t.Error()
	}
	if pathMatched != true {
		t.Error()
	}

	// part of the url string
	input = "/"
	route, params, pathMatched, err = r.findRoute("GET", input)
	if err != nil {
		t.Fatal()
	}
	if route.PathExp != "/" {
		t.Error()
	}
	if len(params) != 0 {
		t.Error()
	}
	if pathMatched != true {
		t.Error()
	}

	// url object
	urlObj, err := url.Parse("http://example.org/")
	if err != nil {
		t.Fatal()
	}
	route, params, pathMatched = r.findRouteFromURL("GET", urlObj)
	if route.PathExp != "/" {
		t.Error()
	}
	if len(params) != 0 {
		t.Error()
	}
	if pathMatched != true {
		t.Error()
	}
}

func TestNoRoute(t *testing.T) {

	r := router{
		routes: []*Route{},
	}

	err := r.start()
	if err != nil {
		t.Fatal()
	}

	input := "http://example.org/notfound"
	route, params, pathMatched, err := r.findRoute("GET", input)
	if err != nil {
		t.Fatal()
	}

	if route != nil {
		t.Error("should not be able to find a route")
	}
	if params != nil {
		t.Error("params must be nil too")
	}
	if pathMatched != false {
		t.Error()
	}
}

func TestEmptyPathExp(t *testing.T) {

	r := router{
		routes: []*Route{
			&Route{
				HttpMethod: "GET",
				PathExp:    "",
			},
		},
	}

	err := r.start()
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Error("expected the empty PathExp error")
	}
}

func TestInvalidPathExp(t *testing.T) {

	r := router{
		routes: []*Route{
			&Route{
				HttpMethod: "GET",
				PathExp:    "invalid",
			},
		},
	}

	err := r.start()
	if err == nil || !strings.Contains(err.Error(), "/") {
		t.Error("expected the / PathExp error")
	}
}

func TestUrlEncodedFind(t *testing.T) {

	r := router{
		routes: []*Route{
			&Route{
				HttpMethod: "GET",
				PathExp:    "/with space", // not urlencoded
			},
		},
	}

	err := r.start()
	if err != nil {
		t.Fatal()
	}

	input := "http://example.org/with%20space" // urlencoded
	route, _, pathMatched, err := r.findRoute("GET", input)
	if err != nil {
		t.Fatal()
	}
	if route.PathExp != "/with space" {
		t.Error()
	}
	if pathMatched != true {
		t.Error()
	}
}

func TestWithQueryString(t *testing.T) {

	r := router{
		routes: []*Route{
			&Route{
				HttpMethod: "GET",
				PathExp:    "/r/:id",
			},
		},
	}

	err := r.start()
	if err != nil {
		t.Fatal()
	}

	input := "http://example.org/r/123?arg=value"
	route, params, pathMatched, err := r.findRoute("GET", input)
	if err != nil {
		t.Fatal()
	}
	if route == nil {
		t.Fatal("Expected a match")
	}
	if params["id"] != "123" {
		t.Errorf("expected 123, got %s", params["id"])
	}
	if pathMatched != true {
		t.Error()
	}
}

func TestNonUrlEncodedFind(t *testing.T) {

	r := router{
		routes: []*Route{
			&Route{
				HttpMethod: "GET",
				PathExp:    "/with%20space", // urlencoded
			},
		},
	}

	err := r.start()
	if err != nil {
		t.Fatal()
	}

	input := "http://example.org/with space" // not urlencoded
	route, _, pathMatched, err := r.findRoute("GET", input)
	if err != nil {
		t.Fatal()
	}
	if route.PathExp != "/with%20space" {
		t.Error()
	}
	if pathMatched != true {
		t.Error()
	}
}

func TestDuplicatedRoute(t *testing.T) {

	r := router{
		routes: []*Route{
			&Route{
				HttpMethod: "GET",
				PathExp:    "/",
			},
			&Route{
				HttpMethod: "GET",
				PathExp:    "/",
			},
		},
	}

	err := r.start()
	if err == nil {
		t.Error("expected the duplicated route error")
	}
}

func TestSplatUrlEncoded(t *testing.T) {

	r := router{
		routes: []*Route{
			&Route{
				HttpMethod: "GET",
				PathExp:    "/r/*rest",
			},
		},
	}

	err := r.start()
	if err != nil {
		t.Fatal()
	}

	input := "http://example.org/r/123"
	route, params, pathMatched, err := r.findRoute("GET", input)
	if err != nil {
		t.Fatal()
	}
	if route == nil {
		t.Fatal("Expected a match")
	}
	if params["rest"] != "123" {
		t.Error()
	}
	if pathMatched != true {
		t.Error()
	}
}

func TestRouteOrder(t *testing.T) {

	r := router{
		routes: []*Route{
			&Route{
				HttpMethod: "GET",
				PathExp:    "/r/:id",
			},
			&Route{
				HttpMethod: "GET",
				PathExp:    "/r/*rest",
			},
		},
	}

	err := r.start()
	if err != nil {
		t.Fatal(err)
	}

	input := "http://example.org/r/123"
	route, params, pathMatched, err := r.findRoute("GET", input)
	if err != nil {
		t.Fatal(err)
	}
	if route == nil {
		t.Fatal("Expected one route to be matched")
	}
	if route.PathExp != "/r/:id" {
		t.Errorf("both match, expected the first defined, got %s", route.PathExp)
	}
	if params["id"] != "123" {
		t.Error()
	}
	if pathMatched != true {
		t.Error()
	}
}

func TestRelaxedPlaceholder(t *testing.T) {

	r := router{
		routes: []*Route{
			&Route{
				HttpMethod: "GET",
				PathExp:    "/r/:id",
			},
			&Route{
				HttpMethod: "GET",
				PathExp:    "/r/#filename",
			},
		},
	}

	err := r.start()
	if err != nil {
		t.Fatal()
	}

	input := "http://example.org/r/a.txt"
	route, params, pathMatched, err := r.findRoute("GET", input)
	if err != nil {
		t.Fatal(err)
	}
	if route == nil {
		t.Fatal("Expected one route to be matched")
	}
	if route.PathExp != "/r/#filename" {
		t.Errorf("expected the second route, got %s", route.PathExp)
	}
	if params["filename"] != "a.txt" {
		t.Error()
	}
	if pathMatched != true {
		t.Error()
	}
}

func TestSimpleExample(t *testing.T) {

	r := router{
		routes: []*Route{
			&Route{
				HttpMethod: "GET",
				PathExp:    "/resources/:id",
			},
			&Route{
				HttpMethod: "GET",
				PathExp:    "/resources",
			},
		},
	}

	err := r.start()
	if err != nil {
		t.Fatal()
	}

	input := "http://example.org/resources/123"
	route, params, pathMatched, err := r.findRoute("GET", input)
	if err != nil {
		t.Fatal()
	}

	if route.PathExp != "/resources/:id" {
		t.Error()
	}
	if params["id"] != "123" {
		t.Error()
	}
	if pathMatched != true {
		t.Error()
	}
}
