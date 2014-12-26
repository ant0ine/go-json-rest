package rest

import (
	"net/http"
)

const xPoweredByDefault = "go-json-rest"

// Handle the transition between net/http and go-json-rest objects.
// It intanciates the rest.Request and rest.ResponseWriter, ...
type jsonAdapter struct {
	DisableJsonIndent bool
	XPoweredBy        string
	DisableXPoweredBy bool
}

func (ja *jsonAdapter) AdapterFunc(handler HandlerFunc) http.HandlerFunc {

	poweredBy := ""
	if !ja.DisableXPoweredBy {
		if ja.XPoweredBy == "" {
			poweredBy = xPoweredByDefault
		} else {
			poweredBy = ja.XPoweredBy
		}
	}

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
			poweredBy,
		}

		// call the wrapped handler
		handler(writer, request)
	}
}
