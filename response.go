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
	isGzipped   bool
	isIndented  bool
	statusCode  int
	wroteHeader bool
}

// Overloading of the http.ResponseWriter method.
// Just record the status code for logging.
func (self *ResponseWriter) WriteHeader(code int) {
	self.ResponseWriter.WriteHeader(code)
	self.statusCode = code
	self.wroteHeader = true
}

// Overloading of the http.ResponseWriter method.
// Provide additional capabilities, like transparent gzip encoding.
func (self *ResponseWriter) Write(b []byte) (int, error) {

	if self.isGzipped {
		self.Header().Set("Content-Encoding", "gzip")
	}

	if !self.wroteHeader {
		self.WriteHeader(http.StatusOK)
	}

	if self.isGzipped {
		gzipWriter := gzip.NewWriter(self.ResponseWriter)
		defer gzipWriter.Close()
		return gzipWriter.Write(b)
	}

	return self.ResponseWriter.Write(b)
}

// Encode the object in JSON, set the content-type header,
// and call Write.
func (self *ResponseWriter) WriteJson(v interface{}) error {
	self.Header().Set("content-type", "application/json")
	var b []byte
	var err error
	if self.isIndented {
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

// Produce an error response in JSON with the following structure, '{"Error":"My error message"}'
// The standard plain text net/http Error helper can still be called like this:
// http.Error(w, "error message", code)
func Error(w *ResponseWriter, error string, code int) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	err := w.WriteJson(map[string]string{"Error": error})
	if err != nil {
		panic(err)
	}
}

// Produce a 404 response with the following JSON, '{"Error":"Resource not found"}'
// The standard plain text net/http NotFound helper can still be called like this:
// http.NotFound(w, r.Request)
func NotFound(w *ResponseWriter, r *Request) {
	Error(w, "Resource not found", http.StatusNotFound)
}
