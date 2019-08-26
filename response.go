package rest

import (
	"bufio"
	"encoding/json"
	"net"
	"net/http"
)

// A ResponseWriter interface dedicated to JSON HTTP response.
// Note, the responseWriter object instantiated by the framework also implements many other interfaces
// accessible by type assertion: http.ResponseWriter, http.Flusher, http.CloseNotifier, http.Hijacker.
type ResponseWriter interface {

	// Identical to the http.ResponseWriter interface
	Header() http.Header

	// Use EncodeJson to generate the payload, write the headers with http.StatusOK if
	// they are not already written, then write the payload.
	// The Content-Type header is set to "application/json", unless already specified.
	WriteJson(v interface{}) error

	// Encode the data structure to JSON, mainly used to wrap ResponseWriter in
	// middlewares.
	EncodeJson(v interface{}) ([]byte, error)

	// Similar to the http.ResponseWriter interface, with additional JSON related
	// headers set.
	WriteHeader(int)
}

// This allows to customize the field name used in the error response payload.
// It defaults to "Error" for compatibility reason, but can be changed before starting the server.
// eg: rest.ErrorFieldName = "errorMessage"
var ErrorFieldName = "Error"

// Error produces an error response in JSON with the following structure, '{"Error":"My error message"}'
// The standard plain text net/http Error helper can still be called like this:
// http.Error(w, "error message", code)
func Error(w ResponseWriter, error string, code int) {
	w.WriteHeader(code)
	err := w.WriteJson(map[string]string{ErrorFieldName: error})
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
// http.Hijacker
type responseWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (w *responseWriter) WriteHeader(code int) {
	if w.Header().Get("Content-Type") == "" {
		// Per spec, UTF-8 is the default, and the charset parameter should not
		// be necessary. But some clients (eg: Chrome) think otherwise.
		// Since json.Marshal produces UTF-8, setting the charset parameter is a
		// safe option.
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}
	w.ResponseWriter.WriteHeader(code)
	w.wroteHeader = true
}

func (w *responseWriter) EncodeJson(v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
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
	_, err = w.Write(b)
	if err != nil {
		return err
	}
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

// Provided in order to implement the http.Hijacker interface.
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker := w.ResponseWriter.(http.Hijacker)
	return hijacker.Hijack()
}
