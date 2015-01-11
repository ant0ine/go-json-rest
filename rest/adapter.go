package rest

import (
	"net/http"
)

// Handle the transition between net/http and go-json-rest objects.
// It intanciates the rest.Request and rest.ResponseWriter, ...
type jsonAdapter struct {
	DisableJsonIndent bool
}

func (ja *jsonAdapter) AdapterFunc(handler HandlerFunc) http.HandlerFunc {

	return func(origWriter http.ResponseWriter, origRequest *http.Request) {

		// instantiate the rest objects
		request := &Request{
			origRequest,
			nil,
			map[string]interface{}{},
		}

		writer := &responseWriter{
			origWriter,
			false,
			!ja.DisableJsonIndent,
		}

		// call the wrapped handler
		handler(writer, request)
	}
}
