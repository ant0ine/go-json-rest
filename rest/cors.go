package rest

import (
	"net/http"
)

// CorsMiddleware provides a configurable CORS implementation.
type CorsMiddleware struct {

	// Reject non CORS requests if true. See CorsInfo.IsCors.
	RejectNonCorsRequests bool

	// Function excecuted for every CORS requests to validate the Origin. (Required)
	// Must return true if valid, false if invalid.
	// For instance: simple equality, regexp, DB lookup, ...
	OriginValidator func(origin string, request *Request) bool

	// TODO convert that to a slice ??? an uppercase strings ?
	// Method strings must be uppercase as CorsInfo.AccessControlRequestMethod is always uppercase.
	AllowedMethods map[string]bool

	// TODO
	AllowedHeaders map[string]bool

	// User to se the Access-Control-Allow-Credentials response header.
	AccessControlAllowCredentials bool

	// Used to set the Access-Control-Max-Age response header, in seconds.
	AccessControlMaxAge int
}

func (mw *CorsMiddleware) MiddlewareFunc(handler HandlerFunc) HandlerFunc {
	return func(writer ResponseWriter, request *Request) {

		corsInfo := request.GetCorsInfo()

		// non CORS requests
		if !corsInfo.IsCors {
			if mw.RejectNonCorsRequests {
				Error(writer, "Non CORS request", http.StatusForbidden)
				return
			}
			// continue, execute the wrapped middleware
			handler(writer, request)
			return
		}

		// Validate the Origin
		if mw.OriginValidator(corsInfo.Origin, request) == false {
			Error(writer, "Invalid Origin", http.StatusForbidden)
			return
		}

		if corsInfo.IsPreflight {

			// check the request methods
			if mw.AllowedMethods[corsInfo.AccessControlRequestMethod] == false {
				Error(writer, "Invalid Preflight Request", http.StatusForbidden)
				return
			}

			// check the request headers
			for _, requestedHeader := range corsInfo.AccessControlRequestHeaders {
				if !mw.AllowedHeaders[requestedHeader] {
					Error(writer, "Invalid Preflight Request", http.StatusForbidden)
					return
				}
			}

			for allowedMethod, _ := range allowedMethods {
				writer.Header().Add("Access-Control-Allow-Methods", allowedMethod)
			}
			for allowedHeader, _ := range allowedHeaders {
				writer.Header().Add("Access-Control-Allow-Headers", allowedHeader)
			}
			writer.Header().Set("Access-Control-Allow-Origin", corsInfo.Origin)
			if mw.AccessControlAllowCredentials == true {
				writer.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			writer.Header().Set("Access-Control-Max-Age", string(mw.AccessControlMaxAge))
			writer.WriteHeader(http.StatusOK)
			return
		} else {
			writer.Header().Set("Access-Control-Expose-Headers", "X-Powered-By") // TODO
			writer.Header().Set("Access-Control-Allow-Origin", corsInfo.Origin)
			if mw.AccessControlAllowCredentials == true {
				writer.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			// continure, execute the wrapped middleware
			handler(writer, request)
			return
		}
	}
}
