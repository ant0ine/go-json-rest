// Utility functions to help writing tests for a Go-Json-Rest app
//
// Go comes with net/http/httptest to help writing test for an http
// server. When this http server implements a JSON REST API, some basic
// checks end up to be always the same. This test package tries to save
// some typing by providing helpers for this particular use case.
//
//	package main
//
//	import (
//		"github.com/ant0ine/go-json-rest/rest/test"
//		"testing"
//	)
//
//	func TestSimpleRequest(t *testing.T) {
//		handler := ResourceHandler{}
//		handler.SetRoutes(
//			&Route{"GET", "/r",
//				func(w ResponseWriter, r *Request) {
//					w.WriteJson(map[string]string{"Id": "123"})
//				},
//			},
//		)
//		recorded := test.RunRequest(t, &handler,
//			test.MakeSimpleRequest("GET", "http://1.2.3.4/r", nil))
//		recorded.CodeIs(200)
//		recorded.ContentTypeIsJson()
//	}
package test
