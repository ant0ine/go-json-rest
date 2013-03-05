package rest

import (
	"compress/gzip"
	"encoding/json"
	"net/http"
)

// Inherit from an object implementing the http.ResponseWriter interface,
// and provide additional methods.
type ResponseWriter struct {
	http.ResponseWriter
	is_gzipped   bool
	is_indented  bool
	status_code  int
	wrote_header bool
}

// Overloading of the http.ResponseWriter method.
// Just record the status code for logging.
func (self *ResponseWriter) WriteHeader(code int) {
	self.ResponseWriter.WriteHeader(code)
	self.status_code = code
	self.wrote_header = true
}

// Overloading of the http.ResponseWriter method.
// Provide additional capabilities, like transparent gzip encoding.
func (self *ResponseWriter) Write(b []byte) (int, error) {

	if !self.wrote_header {
		self.WriteHeader(http.StatusOK)
	}

	if self.is_gzipped {
		self.Header().Set("Content-Encoding", "gzip")
		gzip_writer := gzip.NewWriter(self.ResponseWriter)
		defer gzip_writer.Close()
		return gzip_writer.Write(b)
	}

	return self.ResponseWriter.Write(b)
}

// Encode the object in JSON, set the content-type header,
// and call Write.
func (self *ResponseWriter) WriteJson(v interface{}) error {
	self.Header().Set("content-type", "application/json")
	var b []byte
	var err error
	if self.is_indented {
		b, err = json.MarshalIndent(v, "", "  ")
	} else {
		b, err = json.Marshal(v)
	}
	if err != nil {
		return err
	}
	self.Write(b)
	return nil
}
