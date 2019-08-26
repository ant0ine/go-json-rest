package test

import (
	"compress/gzip"
	"io"
	"net/http/httptest"
	"reflect"
	"testing"
)

func testDecodedBody(t *testing.T, zip bool) {
	type Data struct {
		N int
	}
	input := `{"N": 1}`
	expectedData := Data{N: 1}

	w := httptest.NewRecorder()

	if zip {
		w.Header().Set("Content-Encoding", "gzip")
		enc := gzip.NewWriter(w)
		io.WriteString(enc, input)
		enc.Close()
	} else {
		io.WriteString(w, input)
	}

	var gotData Data
	if err := DecodeJsonPayload(w, &gotData); err != nil {
		t.Errorf("DecodeJsonPayload error: %s", err)
	}
	if !reflect.DeepEqual(expectedData, gotData) {
		t.Errorf("DecodeJsonPayload expected: %#v, got %#v", expectedData, gotData)
	}
}

func TestDecodedBodyUnzipped(t *testing.T) {
	testDecodedBody(t, false)
}

func TestDecodedBodyZipped(t *testing.T) {
	testDecodedBody(t, true)
}
