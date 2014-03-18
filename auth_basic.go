package rest

import (
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"strings"
)

// AuthBasicMiddleware provides a simple AuthBasic implementation.
// It can be used before routing to protect all the endpoints, see PreRoutingMiddlewares.
// Or it can be used to wrap a particular endpoint HandlerFunc.
type AuthBasicMiddleware struct {

	// Realm name to display to the user. (Required)
	Realm string

	// Callback function that should perform the authentication of the user based on userId and password.
	// Must return true on success, false on failure. (Required)
	Authenticator func(userId string, password string) bool
}

// MiddlewareFunc tries to authenticate the user. It sends a 401 on failure,
// and executes the wrapped handler on success.
// Note that, on success, the userId is made available in the environment at request.Env["REMOTE_USER"]
func (mw *AuthBasicMiddleware) MiddlewareFunc(handler HandlerFunc) HandlerFunc {

	if mw.Realm == "" {
		log.Fatal("Realm is required")
	}
	if mw.Authenticator == nil {
		log.Fatal("Authenticator is required")
	}

	return func(writer ResponseWriter, request *Request) {

		authHeader := request.Header.Get("Authorization")
		if authHeader == "" {
			mw.unauthorized(writer)
			return
		}

		providedUserId, providedPassword, err := mw.decodeBasicAuthHeader(authHeader)

		if err != nil {
			Error(writer, "Invalid authentication", http.StatusBadRequest)
			return
		}

		if !mw.Authenticator(providedUserId, providedPassword) {
			mw.unauthorized(writer)
			return
		}

		request.Env["REMOTE_USER"] = providedUserId

		handler(writer, request)
	}
}

func (mw *AuthBasicMiddleware) unauthorized(writer ResponseWriter) {
	writer.Header().Set("WWW-Authenticate", "Basic realm="+mw.Realm)
	Error(writer, "Not Authorized", http.StatusUnauthorized)
}

func (mw *AuthBasicMiddleware) decodeBasicAuthHeader(header string) (user string, password string, err error) {

	parts := strings.SplitN(header, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Basic") {
		return "", "", errors.New("Invalid authentication")
	}

	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", errors.New("Invalid base64")
	}

	creds := strings.SplitN(string(decoded), ":", 2)
	if len(creds) != 2 {
		return "", "", errors.New("Invalid authentication")
	}

	return creds[0], creds[1], nil
}
