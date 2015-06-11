package rest

import (
	"testing"
)

func TestReverseRouteResolution(t *testing.T) {

	noParam := &Route{"GET", "/", nil}
	got := noParam.MakePath(nil)
	expected := "/"
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}

	twoParams := &Route{"GET", "/:id.:format", nil}
	got = twoParams.MakePath(
		map[string]string{
			"id":     "123",
			"format": "json",
		},
	)
	expected = "/123.json"
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}

	splatParam := &Route{"GET", "/:id.*format", nil}
	got = splatParam.MakePath(
		map[string]string{
			"id":     "123",
			"format": "tar.gz",
		},
	)
	expected = "/123.tar.gz"
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}

	relaxedParam := &Route{"GET", "/#file", nil}
	got = relaxedParam.MakePath(
		map[string]string{
			"file": "a.txt",
		},
	)
	expected = "/a.txt"
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestShortcutMethods(t *testing.T) {

	r := Head("/", nil)
	if r.HttpMethod != "HEAD" {
		t.Errorf("expected HEAD, got %s", r.HttpMethod)
	}

	r = Get("/", nil)
	if r.HttpMethod != "GET" {
		t.Errorf("expected GET, got %s", r.HttpMethod)
	}

	r = Post("/", nil)
	if r.HttpMethod != "POST" {
		t.Errorf("expected POST, got %s", r.HttpMethod)
	}

	r = Put("/", nil)
	if r.HttpMethod != "PUT" {
		t.Errorf("expected PUT, got %s", r.HttpMethod)
	}

	r = Patch("/", nil)
	if r.HttpMethod != "PATCH" {
		t.Errorf("expected PATCH, got %s", r.HttpMethod)
	}

	r = Delete("/", nil)
	if r.HttpMethod != "DELETE" {
		t.Errorf("expected DELETE, got %s", r.HttpMethod)
	}

	r = Options("/", nil)
	if r.HttpMethod != "OPTIONS" {
		t.Errorf("expected OPTIONS, got %s", r.HttpMethod)
	}
}
