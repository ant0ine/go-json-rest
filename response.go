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

// Produce an error response in JSON with the following structure, '{"error":"error message"}'
// The standard plain text net/http Error helper can still be called like this:
// http.Error(w, "error message", code)
func Error(w *ResponseWriter, error string, code int) {
	w.WriteHeader(code)
	err := w.WriteJson(map[string]string{"error": error})
	if err != nil {
		panic(err)
	}
}

// Produce a 404 response with the following JSON, '{"error":"Resource not found"}'
// The standard plain text net/http NotFound helper can still be called like this:
// http.NotFound(w, r.Request)
func NotFound(w *ResponseWriter, r *Request) {
	Error(w, "Resource not found", http.StatusNotFound)
}
