package rest

import (
	"net/http"
)

type recorderResponseWriter struct {
	ResponseWriter
	statusCode  int
	wroteHeader bool
}

// Record the status code
func (w *recorderResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	w.statusCode = code
	w.wroteHeader = true
}

// Make sure the this WriteHeader is called
func (w *recorderResponseWriter) WriteJson(v interface{}) error {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.WriteJson(v)
}

// Provided in order to implement the http.Flusher interface.
func (w *recorderResponseWriter) Flush() {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	flusher := w.ResponseWriter.(http.Flusher)
	flusher.Flush()
}

// Provided in order to implement the http.CloseNotifier interface.
func (w *recorderResponseWriter) CloseNotify() <-chan bool {
	notifier := w.ResponseWriter.(http.CloseNotifier)
	return notifier.CloseNotify()
}

// Provided in order to implement the http.ResponseWriter interface.
func (w *recorderResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	writer := w.ResponseWriter.(http.ResponseWriter)
	return writer.Write(b)
}

// Middleware function.
func (rh *ResourceHandler) recorderWrapper(h HandlerFunc) HandlerFunc {
	return func(w ResponseWriter, r *Request) {

		writer := &recorderResponseWriter{w, 0, false}

		// call the handler
		h(writer, r)

		rh.env.setVar(r, "statusCode", writer.statusCode)
	}
}
