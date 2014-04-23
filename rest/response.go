package rest

import (
	"encoding/json"
	"net/http"
)

const xPoweredByDefault = "go-json-rest"

// A ResponseWriter interface dedicated to JSON HTTP response.
// Note that the object instantiated by the ResourceHandler that implements this interface,
// also happens to implement http.ResponseWriter, http.Flusher and http.CloseNotifier.
type ResponseWriter interface {

	// Identical to the http.ResponseWriter interface
	Header() http.Header

	// Use EncodeJson to generate the payload, write the headers with http.StatusOK if they
	// are not already written, then write the payload.
	WriteJson(v interface{}) error

	// Encode the data structure to JSON, mainly used to wrap ResponseWriter in middlewares.
	EncodeJson(v interface{}) ([]byte, error)

	// Similar to the http.ResponseWriter interface, with additional JSON related headers set.
	WriteHeader(int)

	// Identical to the Http.ResponseWriter interface
	Write([]byte) (int, error)
}

// Error produces an error response in JSON with the following structure, '{"Error":"My error message"}'
// The standard plain text net/http Error helper can still be called like this:
// http.Error(w, "error message", code)
func Error(w ResponseWriter, error string, code int) {
	w.WriteHeader(code)
	err := w.WriteJson(map[string]string{"Error": error})
	if err != nil {
		panic(err)
	}
}

// NotFound produces a 404 response with the following JSON, '{"Error":"Resource not found"}'
// The standard plain text net/http NotFound helper can still be called like this:
// http.NotFound(w, r.Request)
func NotFound(w ResponseWriter, r *Request) {
	Error(w, "Resource not found", http.StatusNotFound)
}

// Private responseWriter intantiated by the resource handler.
// It implements the following interfaces:
// ResponseWriter
// http.ResponseWriter
// http.Flusher
// http.CloseNotifier
type responseWriter struct {
	http.ResponseWriter
	wroteHeader bool
	isIndented  bool
	xPoweredBy  string
}

func (w *responseWriter) WriteHeader(code int) {
	w.Header().Set("content-type", "application/json")
	if len(w.xPoweredBy) > 0 {
		w.Header().Add("X-Powered-By", w.xPoweredBy)
	}
	w.ResponseWriter.WriteHeader(code)
	w.wroteHeader = true
}

func (w *responseWriter) EncodeJson(v interface{}) ([]byte, error) {
	var b []byte
	var err error
	if w.isIndented {
		b, err = json.MarshalIndent(v, "", "  ")
	} else {
		b, err = json.Marshal(v)
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Encode the object in JSON and call Write.
func (w *responseWriter) WriteJson(v interface{}) error {
	b, err := w.EncodeJson(v)
	if err != nil {
		return err
	}
	w.Write(b)
	return nil
}

// Provided in order to implement the http.ResponseWriter interface.
func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// Provided in order to implement the http.Flusher interface.
func (w *responseWriter) Flush() {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	flusher := w.ResponseWriter.(http.Flusher)
	flusher.Flush()
}

// Provided in order to implement the http.CloseNotifier interface.
func (w *responseWriter) CloseNotify() <-chan bool {
	notifier := w.ResponseWriter.(http.CloseNotifier)
	return notifier.CloseNotify()
}
