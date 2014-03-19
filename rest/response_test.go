package rest

import (
	"testing"
)

func TestResponseNotIndent(t *testing.T) {

	writer := responseWriter{
		nil,
		false,
		false,
	}

	got, err := writer.EncodeJson(map[string]bool{"test": true})
	if err != nil {
		t.Error(err.Error())
	}
	gotStr := string(got)
	expected := "{\"test\":true}"
	if gotStr != expected {
		t.Error(expected + " was the expected, but instead got " + gotStr)
	}
}

func TestResponseIndent(t *testing.T) {

	writer := responseWriter{
		nil,
		false,
		true,
	}

	got, err := writer.EncodeJson(map[string]bool{"test": true})
	if err != nil {
		t.Error(err.Error())
	}
	gotStr := string(got)
	expected := "{\n  \"test\": true\n}"
	if gotStr != expected {
		t.Error(expected + " was the expected, but instead got " + gotStr)
	}
}
