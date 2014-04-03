package rest

import (
	"compress/gzip"
	"net/http"
	"strings"
)

// gzipMiddleware is responsible for compressing the payload with gzip
// and setting the proper headers when supported by the client.
type gzipMiddleware struct{}

func (mw *gzipMiddleware) MiddlewareFunc(h HandlerFunc) HandlerFunc {
	return func(w ResponseWriter, r *Request) {
		// gzip support enabled
		canGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
		// client accepts gzip ?
		writer := &gzipResponseWriter{w, false, canGzip}
		// call the handler with the wrapped writer
		h(writer, r)
	}
}

// Private responseWriter intantiated by the gzip middleware.
// It encodes the payload with gzip and set the proper headers.
// It implements the following interfaces:
// ResponseWriter
// http.ResponseWriter
// http.Flusher
// http.CloseNotifier
type gzipResponseWriter struct {
	ResponseWriter
	wroteHeader bool
	canGzip     bool
}

// Set the right headers for gzip encoded responses.
func (w *gzipResponseWriter) WriteHeader(code int) {

	// Always set the Vary header, even if this particular request
	// is not gzipped.
	w.Header().Add("Vary", "Accept-Encoding")

	if w.canGzip {
		w.Header().Set("Content-Encoding", "gzip")
	}

	w.ResponseWriter.WriteHeader(code)
	w.wroteHeader = true
}

// Make sure the local Write is called.
func (w *gzipResponseWriter) WriteJson(v interface{}) error {
	b, err := w.EncodeJson(v)
	if err != nil {
		return err
	}
	w.Write(b)
	return nil
}

// Make sure the local WriteHeader is called, and call the parent Flush.
// Provided in order to implement the http.Flusher interface.
func (w *gzipResponseWriter) Flush() {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	flusher := w.ResponseWriter.(http.Flusher)
	flusher.Flush()
}

// Call the parent CloseNotify.
// Provided in order to implement the http.CloseNotifier interface.
func (w *gzipResponseWriter) CloseNotify() <-chan bool {
	notifier := w.ResponseWriter.(http.CloseNotifier)
	return notifier.CloseNotify()
}

// Make sure the local WriteHeader is called, and encode the payload if necessary.
// Provided in order to implement the http.ResponseWriter interface.
func (w *gzipResponseWriter) Write(b []byte) (int, error) {

	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	writer := w.ResponseWriter.(http.ResponseWriter)

	if w.canGzip {
		gzipWriter := gzip.NewWriter(writer)
		defer gzipWriter.Close()
		return gzipWriter.Write(b)
	}

	return writer.Write(b)
}
