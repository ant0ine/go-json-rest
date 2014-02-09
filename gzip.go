package rest

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	wroteHeader bool
	canGzip     bool
}

func (self *gzipResponseWriter) WriteHeader(code int) {

	// always set the Vary header, even if this particular request
	// is not gzipped.
	self.Header().Add("Vary", "Accept-Encoding")

	if self.canGzip {
		self.Header().Set("Content-Encoding", "gzip")
	}

	self.ResponseWriter.WriteHeader(code)
	self.wroteHeader = true
}

func (self *gzipResponseWriter) Write(b []byte) (int, error) {

	if !self.wroteHeader {
		self.WriteHeader(http.StatusOK)
	}

	if self.canGzip {
		gzipWriter := gzip.NewWriter(self.ResponseWriter)
		defer gzipWriter.Close()
		return gzipWriter.Write(b)
	}

	return self.ResponseWriter.Write(b)
}

func (self *ResourceHandler) gzipWrapper(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if self.EnableGzip == true {
			// gzip support enabled
			canGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
			// client accepts gzip ?
			writer := &gzipResponseWriter{w, false, canGzip}
			// call the handler with the wrapped writer
			h(writer, r)
		} else {
			// gzip support disabled, don't do anything
			h(w, r)
		}
	}
}
