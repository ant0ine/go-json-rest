package rest

import (
	"net/http"
)

type recorderResponseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func (w *recorderResponseWriter) WriteHeader(code int) {
	w.Header().Add("X-Powered-By", "go-json-rest")
	w.ResponseWriter.WriteHeader(code)
	w.statusCode = code
	w.wroteHeader = true
}

func (w *recorderResponseWriter) Flush() {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	flusher := w.ResponseWriter.(http.Flusher)
	flusher.Flush()
}

func (w *recorderResponseWriter) CloseNotify() <-chan bool {
	notifier := w.ResponseWriter.(http.CloseNotifier)
	return notifier.CloseNotify()
}

func (w *recorderResponseWriter) Write(b []byte) (int, error) {

	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	return w.ResponseWriter.Write(b)
}

func (rh *ResourceHandler) recorderWrapper(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		writer := &recorderResponseWriter{w, 0, false}

		// call the handler
		h(writer, r)

		rh.env.setVar(r, "statusCode", writer.statusCode)
	}
}
