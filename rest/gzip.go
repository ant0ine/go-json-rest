package rest

import (
	"bufio"
	"compress/gzip"
	"net"
	"net/http"
	"strings"
)

// GzipMiddleware is responsible for compressing the payload with gzip and setting the proper
// headers when supported by the client. It must be wrapped by TimerMiddleware for the
// compression time to be captured. And It must be wrapped by RecorderMiddleware for the
// compressed BYTES_WRITTEN to be captured.
type GzipMiddleware struct{}

// MiddlewareFunc makes GzipMiddleware implement the Middleware interface.
func (mw *GzipMiddleware) MiddlewareFunc(h HandlerFunc) HandlerFunc {
	return func(w ResponseWriter, r *Request) {
		// gzip support enabled
		canGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
		// client accepts gzip ?
		writer := &gzipResponseWriter{w, false, canGzip, nil}
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
// http.Hijacker
type gzipResponseWriter struct {
	ResponseWriter
	wroteHeader bool
	canGzip     bool
	gzipWriter  *gzip.Writer
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
	_, err = w.Write(b)
	if err != nil {
		return err
	}
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

// Provided in order to implement the http.Hijacker interface.
func (w *gzipResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker := w.ResponseWriter.(http.Hijacker)
	return hijacker.Hijack()
}

// Make sure the local WriteHeader is called, and encode the payload if necessary.
// Provided in order to implement the http.ResponseWriter interface.
func (w *gzipResponseWriter) Write(b []byte) (int, error) {

	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	writer := w.ResponseWriter.(http.ResponseWriter)

	if w.canGzip {
		// Write can be called multiple times for a given response.
		// (see the streaming example:
		// https://github.com/ant0ine/go-json-rest-examples/tree/master/streaming)
		// The gzipWriter is instantiated only once, and flushed after
		// each write.
		if w.gzipWriter == nil {
			w.gzipWriter = gzip.NewWriter(writer)
		}
		count, errW := w.gzipWriter.Write(b)
		errF := w.gzipWriter.Flush()
		if errW != nil {
			return count, errW
		}
		if errF != nil {
			return count, errF
		}
		return count, nil
	}

	return writer.Write(b)
}
