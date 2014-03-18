package rest

import (
	"net/http"
)

// recorderMiddleware keeps a record of the HTTP status code of the response.
// The result is available to the wrapping handlers in request.Env["STATUS_CODE"] as an int.
type recorderMiddleware struct{}

func (mw *recorderMiddleware) MiddlewareFunc(h HandlerFunc) HandlerFunc {
	return func(w ResponseWriter, r *Request) {

		writer := &recorderResponseWriter{w, 0, false}

		// call the handler
		h(writer, r)

		r.Env["STATUS_CODE"] = writer.statusCode
	}
}

// Private responseWriter intantiated by the recorder middleware.
// It keeps a record of the HTTP status code of the response.
// It implements the following interfaces:
// ResponseWriter
// http.ResponseWriter
// http.Flusher
// http.CloseNotifier
type recorderResponseWriter struct {
	ResponseWriter
	statusCode  int
	wroteHeader bool
}

// Record the status code.
func (w *recorderResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	w.statusCode = code
	w.wroteHeader = true
}

// Make sure the local WriteHeader is called, and call the parent WriteJson.
func (w *recorderResponseWriter) WriteJson(v interface{}) error {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.WriteJson(v)
}

// Make sure the local WriteHeader is called, and call the parent Flush.
// Provided in order to implement the http.Flusher interface.
func (w *recorderResponseWriter) Flush() {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	flusher := w.ResponseWriter.(http.Flusher)
	flusher.Flush()
}

// Call the parent CloseNotify.
// Provided in order to implement the http.CloseNotifier interface.
func (w *recorderResponseWriter) CloseNotify() <-chan bool {
	notifier := w.ResponseWriter.(http.CloseNotifier)
	return notifier.CloseNotify()
}

// Make sure the local WriteHeader is called, and call the parent Write.
// Provided in order to implement the http.ResponseWriter interface.
func (w *recorderResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	writer := w.ResponseWriter.(http.ResponseWriter)
	return writer.Write(b)
}
