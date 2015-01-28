package rest

import (
	"encoding/base64"
	"github.com/ant0ine/go-json-rest/rest/test"
	"testing"
)

func TestAuthBasic(t *testing.T) {

	// the middleware to test
	authMiddleware := &AuthBasicMiddleware{
		Realm: "test zone",
		Authenticator: func(userId string, password string) bool {
			if userId == "admin" && password == "admin" {
				return true
			}
			return false
		},
		Authorizator: func(userId string, request *Request) bool {
			if request.Method == "GET" {
				return true
			}
			return false
		},
	}

	// api for testing failure
	apiFailure := NewApi()
	apiFailure.Use(authMiddleware)
	apiFailure.SetApp(AppSimple(func(w ResponseWriter, r *Request) {
		t.Error("Should never be executed")
	}))
	handler := apiFailure.MakeHandler()

	// simple request fails
	recorded := test.RunRequest(t, handler, test.MakeSimpleRequest("GET", "http://localhost/", nil))
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// auth with wrong cred and right method fails
	wrongCredReq := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	encoded := base64.StdEncoding.EncodeToString([]byte("admin:AdmIn"))
	wrongCredReq.Header.Set("Authorization", "Basic "+encoded)
	recorded = test.RunRequest(t, handler, wrongCredReq)
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// auth with right cred and wrong method fails
	rightCredReq := test.MakeSimpleRequest("POST", "http://localhost/", nil)
	encoded = base64.StdEncoding.EncodeToString([]byte("admin:admin"))
	rightCredReq.Header.Set("Authorization", "Basic "+encoded)
	recorded = test.RunRequest(t, handler, rightCredReq)
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// api for testing success
	apiSuccess := NewApi()
	apiSuccess.Use(authMiddleware)
	apiSuccess.SetApp(AppSimple(func(w ResponseWriter, r *Request) {
		if r.Env["REMOTE_USER"] == nil {
			t.Error("REMOTE_USER is nil")
		}
		user := r.Env["REMOTE_USER"].(string)
		if user != "admin" {
			t.Error("REMOTE_USER is expected to be 'admin'")
		}
		w.WriteJson(map[string]string{"Id": "123"})
	}))

	// auth with right cred and right method succeeds
	rightCredReq = test.MakeSimpleRequest("GET", "http://localhost/", nil)
	encoded = base64.StdEncoding.EncodeToString([]byte("admin:admin"))
	rightCredReq.Header.Set("Authorization", "Basic "+encoded)
	recorded = test.RunRequest(t, apiSuccess.MakeHandler(), rightCredReq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()
}
