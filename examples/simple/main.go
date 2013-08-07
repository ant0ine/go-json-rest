/* The minimal example from the documentation

The Curl Demo:

        curl -i http://127.0.0.1:8080/users/123

*/
package main

import (
	"github.com/ant0ine/go-json-rest"
	"net/http"
)

type User struct {
	Id   string
	Name string
	Posts  string
}

func GetUser(w *rest.ResponseWriter, req *rest.Request) {
	params := map[string][]string {"userId": []string{ req.PathParam("id") }}
	url := req.UriForWithParams("/posts/exports", params)
	user := User{
		Id:   req.PathParam("id"),
		Name: "Antoine",
		Posts: url.String(),
	}
	w.WriteJson(&user)
}

func main() {
	handler := rest.ResourceHandler{}
	handler.SetRoutes(
		rest.Route{"GET", "/users/:id", GetUser},
	)
	http.ListenAndServe(":8080", &handler)
}
