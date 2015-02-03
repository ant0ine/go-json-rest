package rest

import (
	"bufio"
	"net"
	"net/http"
)

// RecorderMiddleware keeps a record of the HTTP status code of the response,
// and the number of bytes written.
// The result is available to the wrapping handlers as request.Env["STATUS_CODE"].(int),
// and as request.Env["BYTES_WRITTEN"].(int64)
type RecorderMiddleware struct{}

// MiddlewareFunc makes RecorderMiddleware implement the Middleware interface.
func (mw *RecorderMiddleware) MiddlewareFunc(h HandlerFunc) HandlerFunc {
	return func(w ResponseWriter, r *Request) {

		writer := &recorderResponseWriter{w, 0, false, 0}

		// call the handler
		h(writer, r)

		r.Env["STATUS_CODE"] = writer.statusCode
		r.Env["BYTES_WRITTEN"] = writer.bytesWritten
	}
}

// Private responseWriter intantiated by the recorder middleware.
// It keeps a record of the HTTP status code of the response.
// It implements the following interfaces:
// ResponseWriter
// http.ResponseWriter
// http.Flusher
// http.CloseNotifier
// http.Hijacker
type recorderResponseWriter struct {
	ResponseWriter
	statusCode   int
	wroteHeader  bool
	bytesWritten int64
}

// Record the status code.
func (w *recorderResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	w.statusCode = code
	w.wroteHeader = true
}

// Make sure the local Write is called.
func (w *recorderResponseWriter) WriteJson(v interface{}) error {
	b, err := w.EncodeJson(v)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	if err != nil {
		return err
	}
	return nil
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

// Provided in order to implement the http.Hijacker interface.
func (w *recorderResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker := w.ResponseWriter.(http.Hijacker)
	return hijacker.Hijack()
}

// Make sure the local WriteHeader is called, and call the parent Write.
// Provided in order to implement the http.ResponseWriter interface.
func (w *recorderResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	writer := w.ResponseWriter.(http.ResponseWriter)
	written, err := writer.Write(b)
	w.bytesWritten += int64(written)
	return written, err
}
