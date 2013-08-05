package rest

import (
	"net/http"
)

type recorderResponseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func (self *recorderResponseWriter) WriteHeader(code int) {
	self.ResponseWriter.WriteHeader(code)
	self.statusCode = code
	self.wroteHeader = true
}

func (self *recorderResponseWriter) Write(b []byte) (int, error) {

	if !self.wroteHeader {
		self.WriteHeader(http.StatusOK)
	}

	return self.ResponseWriter.Write(b)
}

func (self *ResourceHandler) recorderWrapper(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		writer := &recorderResponseWriter{w, 0, false}

		// call the handler
		h(writer, r)

		self.env.setVar(r, "statusCode", writer.statusCode)
	}
}
