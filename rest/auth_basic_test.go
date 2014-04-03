package rest

import (
	"encoding/base64"
	"github.com/ant0ine/go-json-rest/rest/test"
	"testing"
)

func TestAuthBasic(t *testing.T) {

	handler := ResourceHandler{
		PreRoutingMiddlewares: []Middleware{
			&AuthBasicMiddleware{
				Realm: "test zone",
				Authenticator: func(userId string, password string) bool {
					if userId == "admin" && password == "admin" {
						return true
					}
					return false
				},
			},
		},
	}
	handler.SetRoutes(
		&Route{"GET", "/r",
			func(w ResponseWriter, r *Request) {
				w.WriteJson(map[string]string{"Id": "123"})
			},
		},
	)

	// simple request fails
	recorded := test.RunRequest(t, &handler, test.MakeSimpleRequest("GET", "http://1.2.3.4/r", nil))
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// auth with wrong cred fails
	wrongCredReq := test.MakeSimpleRequest("GET", "http://1.2.3.4/r", nil)
	encoded := base64.StdEncoding.EncodeToString([]byte("admin:AdmIn"))
	wrongCredReq.Header.Set("Authorization", "Basic "+encoded)
	recorded = test.RunRequest(t, &handler, wrongCredReq)
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// auth with right cred succeeds
	rightCredReq := test.MakeSimpleRequest("GET", "http://1.2.3.4/r", nil)
	encoded = base64.StdEncoding.EncodeToString([]byte("admin:admin"))
	rightCredReq.Header.Set("Authorization", "Basic "+encoded)
	recorded = test.RunRequest(t, &handler, rightCredReq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()
}
