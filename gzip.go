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

func (w *gzipResponseWriter) Flush() {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	flusher := w.ResponseWriter.(http.Flusher)
	flusher.Flush()
}

func (w *gzipResponseWriter) CloseNotify() <-chan bool {
	notifier := w.ResponseWriter.(http.CloseNotifier)
	return notifier.CloseNotify()
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {

	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	if w.canGzip {
		gzipWriter := gzip.NewWriter(w.ResponseWriter)
		defer gzipWriter.Close()
		return gzipWriter.Write(b)
	}

	return w.ResponseWriter.Write(b)
}

func (rh *ResourceHandler) gzipWrapper(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

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
