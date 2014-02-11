package rest

import (
	"net/http"
)

type recorderResponseWriter struct {
	http.ResponseWriter
	http.Flusher
	statusCode  int
	wroteHeader bool
}

type flusher struct {
}

func (f flusher) Flush() {
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

func responseFlusher(w http.ResponseWriter) http.Flusher {
	if f, ok := w.(http.Flusher); ok {
		return f
	}
	return flusher{}
}

func (self *ResourceHandler) recorderWrapper(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		writer := &recorderResponseWriter{w, responseFlusher(w), 0, false}

		// call the handler
		h(writer, r)

		self.env.setVar(r, "statusCode", writer.statusCode)
	}
}
