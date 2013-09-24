package rest

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (self *gzipResponseWriter) WriteHeader(code int) {
	self.Header().Set("Content-Encoding", "gzip")
	self.ResponseWriter.WriteHeader(code)
	self.wroteHeader = true
}

func (self *gzipResponseWriter) Write(b []byte) (int, error) {

	if !self.wroteHeader {
		self.WriteHeader(http.StatusOK)
	}

	gzipWriter := gzip.NewWriter(self.ResponseWriter)
	defer gzipWriter.Close()
	return gzipWriter.Write(b)
}

func (self *ResourceHandler) gzipWrapper(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// determine if gzip is needed
		if self.EnableGzip == true &&
			strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			writer := &gzipResponseWriter{w, false}
			// call the handler
			h(writer, r)
		} else {
			// call the handler
			h(w, r)
		}
	}
}
