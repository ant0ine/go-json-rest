package rest

import (
	"encoding/json"
	"net/http"
)

// Inherit from an object implementing the http.ResponseWriter interface,
// and provide additional methods.
type ResponseWriter struct {
	http.ResponseWriter
	isIndented bool
}

// Make rest.ResponseWriter implement the http.Flusher interface.
// It propagates the Flush call to the wrapped ResponseWriter.
func (w *ResponseWriter) Flush() {
	flusher := w.ResponseWriter.(http.Flusher)
	flusher.Flush()
}

// Make rest.ResponseWriter implement the http.CloseNotifier interface.
func (w *ResponseWriter) CloseNotify() <-chan bool {
	notifier := w.ResponseWriter.(http.CloseNotifier)
	return notifier.CloseNotify()
}

// Encode the object in JSON, set the content-type header,
// and call Write.
func (w *ResponseWriter) WriteJson(v interface{}) error {
	w.Header().Set("content-type", "application/json")
	var b []byte
	var err error
	if w.isIndented {
		b, err = json.MarshalIndent(v, "", "  ")
	} else {
		b, err = json.Marshal(v)
	}
	if err != nil {
		return err
	}
	w.Write(b)
	return nil
}

// Produce an error response in JSON with the following structure, '{"Error":"My error message"}'
// The standard plain text net/http Error helper can still be called like this:
// http.Error(w, "error message", code)
func Error(w *ResponseWriter, error string, code int) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	err := w.WriteJson(map[string]string{"Error": error})
	if err != nil {
		panic(err)
	}
}

// Produce a 404 response with the following JSON, '{"Error":"Resource not found"}'
// The standard plain text net/http NotFound helper can still be called like this:
// http.NotFound(w, r.Request)
func NotFound(w *ResponseWriter, r *Request) {
	Error(w, "Resource not found", http.StatusNotFound)
}
