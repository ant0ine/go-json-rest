package rest

import (
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/dgrijalva/jwt-go"
	"testing"
	"time"
)

func makeTokenString(username string, key []byte) string {
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	token.Claims["id"] = "admin"
	token.Claims["exp"] = time.Now().Add(time.Hour).Unix()
	tokenString, _ := token.SignedString(key)
	return tokenString
}

func TestAuthJWT(t *testing.T) {
	key := []byte("secret key")

	// the middleware to test
	authMiddleware := &JWTMiddleware{
		Realm:   "test zone",
		Key:     key,
		Timeout: time.Hour,
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

	// auth with right cred and wrong method fails
	rightCredReq := test.MakeSimpleRequest("POST", "http://localhost/", nil)
	rightCredReq.Header.Set("Authorization", "Bearer "+makeTokenString("admin", key))
	recorded = test.RunRequest(t, handler, rightCredReq)
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// wrong Auth format - bearer lower case
	wrongAuthFormat := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	wrongAuthFormat.Header.Set("Authorization", "bearer "+makeTokenString("admin", key))
	recorded = test.RunRequest(t, handler, wrongAuthFormat)
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// wrong Auth format - no space after bearer
	wrongAuthFormat = test.MakeSimpleRequest("GET", "http://localhost/", nil)
	wrongAuthFormat.Header.Set("Authorization", "bearer"+makeTokenString("admin", key))
	recorded = test.RunRequest(t, handler, wrongAuthFormat)
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// wrong Auth format - empty auth header
	wrongAuthFormat = test.MakeSimpleRequest("GET", "http://localhost/", nil)
	wrongAuthFormat.Header.Set("Authorization", "")
	recorded = test.RunRequest(t, handler, wrongAuthFormat)
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// right credt, right method but wrong priv key
	wrongPrivKeyReq := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	wrongPrivKeyReq.Header.Set("Authorization", "Bearer "+makeTokenString("admin", []byte("sekret key")))
	recorded = test.RunRequest(t, handler, wrongPrivKeyReq)
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// right credt, right method, right priv key but timeout
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	token.Claims["id"] = "admin"
	token.Claims["exp"] = 0
	tokenString, _ := token.SignedString(key)

	expiredTimestampReq := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	expiredTimestampReq.Header.Set("Authorization", "Bearer "+tokenString)
	recorded = test.RunRequest(t, handler, expiredTimestampReq)
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
	rightCredReq.Header.Set("Authorization", "Bearer "+makeTokenString("admin", key))
	recorded = test.RunRequest(t, apiSuccess.MakeHandler(), rightCredReq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()

	// login tests
	loginApi := NewApi()
	loginApi.SetApp(AppSimple(authMiddleware.LoginHandler))

	// wrong login
	wrongLoginCreds := map[string]string{"username": "admin", "password": "admIn"}
	wrongLoginReq := test.MakeSimpleRequest("POST", "http://localhost/", wrongLoginCreds)
	recorded = test.RunRequest(t, loginApi.MakeHandler(), wrongLoginReq)
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// empty login
	emptyLoginCreds := map[string]string{}
	emptyLoginReq := test.MakeSimpleRequest("POST", "http://localhost/", emptyLoginCreds)
	recorded = test.RunRequest(t, loginApi.MakeHandler(), emptyLoginReq)
	recorded.CodeIs(401)
	recorded.ContentTypeIsJson()

	// correct login
	loginCreds := map[string]string{"username": "admin", "password": "admin"}
	rightCredReq = test.MakeSimpleRequest("POST", "http://localhost/", loginCreds)
	recorded = test.RunRequest(t, loginApi.MakeHandler(), rightCredReq)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()
}
