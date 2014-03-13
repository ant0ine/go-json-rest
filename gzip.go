package rest

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	ResponseWriter
	wroteHeader bool
	canGzip     bool
}

// Set the right headers for gzip encoded responses.
func (w *gzipResponseWriter) WriteHeader(code int) {

	// always set the Vary header, even if this particular request
	// is not gzipped.
	w.Header().Add("Vary", "Accept-Encoding")

	if w.canGzip {
		w.Header().Set("Content-Encoding", "gzip")
	}

	w.ResponseWriter.WriteHeader(code)
	w.wroteHeader = true
}

// Make sure to use the gzip Write method.
func (w *gzipResponseWriter) WriteJson(v interface{}) error {
	b, err := w.EncodeJson(v)
	if err != nil {
		return err
	}
	w.Write(b)
	return nil
}

// Provided in order to implement the http.Flusher interface.
func (w *gzipResponseWriter) Flush() {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	flusher := w.ResponseWriter.(http.Flusher)
	flusher.Flush()
}

// Provided in order to implement the http.CloseNotifier interface.
func (w *gzipResponseWriter) CloseNotify() <-chan bool {
	notifier := w.ResponseWriter.(http.CloseNotifier)
	return notifier.CloseNotify()
}

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

// The middleware function.
func (rh *ResourceHandler) gzipWrapper(h HandlerFunc) HandlerFunc {
	return func(w ResponseWriter, r *Request) {

		if rh.EnableGzip == true {
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
