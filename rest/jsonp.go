package rest

import (
	"bufio"
	"net"
	"net/http"
)

// JsonpMiddleware provides JSONP responses on demand, based on the presence
// of a query string argument specifying the callback name.
type JsonpMiddleware struct {

	// Name of the query string parameter used to specify the
	// the name of the JS callback used for the padding.
	// Defaults to "callback".
	CallbackNameKey string
}

// MiddlewareFunc returns a HandlerFunc that implements the middleware.
func (mw *JsonpMiddleware) MiddlewareFunc(h HandlerFunc) HandlerFunc {

	if mw.CallbackNameKey == "" {
		mw.CallbackNameKey = "callback"
	}

	return func(w ResponseWriter, r *Request) {

		callbackName := r.URL.Query().Get(mw.CallbackNameKey)
		// TODO validate the callbackName ?

		if callbackName != "" {
			// the client request JSONP, instantiate JsonpMiddleware.
			writer := &jsonpResponseWriter{w, false, callbackName}
			// call the handler with the wrapped writer
			h(writer, r)
		} else {
			// do nothing special
			h(w, r)
		}

	}
}

// Private responseWriter intantiated by the JSONP middleware.
// It adds the padding to the payload and set the proper headers.
// It implements the following interfaces:
// ResponseWriter
// http.ResponseWriter
// http.Flusher
// http.CloseNotifier
// http.Hijacker
type jsonpResponseWriter struct {
	ResponseWriter
	wroteHeader  bool
	callbackName string
}

// Overwrite the Content-Type to be text/javascript
func (w *jsonpResponseWriter) WriteHeader(code int) {

	w.Header().Set("Content-Type", "text/javascript")

	w.ResponseWriter.WriteHeader(code)
	w.wroteHeader = true
}

// Make sure the local Write is called.
func (w *jsonpResponseWriter) WriteJson(v interface{}) error {
	b, err := w.EncodeJson(v)
	if err != nil {
		return err
	}
	// TODO add "/**/" ?
	w.Write([]byte(w.callbackName + "("))
	w.Write(b)
	w.Write([]byte(")"))
	return nil
}

// Make sure the local WriteHeader is called, and call the parent Flush.
// Provided in order to implement the http.Flusher interface.
func (w *jsonpResponseWriter) Flush() {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	flusher := w.ResponseWriter.(http.Flusher)
	flusher.Flush()
}

// Call the parent CloseNotify.
// Provided in order to implement the http.CloseNotifier interface.
func (w *jsonpResponseWriter) CloseNotify() <-chan bool {
	notifier := w.ResponseWriter.(http.CloseNotifier)
	return notifier.CloseNotify()
}

// Provided in order to implement the http.Hijacker interface.
func (w *jsonpResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker := w.ResponseWriter.(http.Hijacker)
	return hijacker.Hijack()
}

// Make sure the local WriteHeader is called.
// Provided in order to implement the http.ResponseWriter interface.
func (w *jsonpResponseWriter) Write(b []byte) (int, error) {

	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	writer := w.ResponseWriter.(http.ResponseWriter)

	return writer.Write(b)
}
