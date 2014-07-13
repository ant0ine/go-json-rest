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
