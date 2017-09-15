package test

import (
	"encoding/json"
	"net/http/httptest"
)

type ResponseRecorder struct {
	*httptest.ResponseRecorder
}

func (r *ResponseRecorder) WriteJson(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = r.Write(b)
	return err
}

func (r *ResponseRecorder) EncodeJson(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func NewRecorder() *ResponseRecorder {
	return &ResponseRecorder{
		ResponseRecorder: httptest.NewRecorder(),
	}
}
