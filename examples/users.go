// Demonstrate how to use rest.RouteObjectMethod
//
// rest.RouteObjectMethod helps create a Route that points to
// an object method instead of just a function.
//
// The Curl Demo:
// curl -i -d '{"Name":"Antoine"}' http://127.0.0.1:8080/users
// curl -i http://127.0.0.1:8080/users/0
//
package main

import (
	"fmt"
	"github.com/ant0ine/go-json-rest"
	"net/http"
)

type User struct {
	Id   string
	Name string
}

type Users struct {
	Store map[string]*User
}

func (self *Users) GetUser(w *rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")
	user := self.Store[id]
	if user == nil {
		http.NotFound(w, r.Request)
		return
	}
	w.WriteJson(&user)
}

func (self *Users) PostUser(w *rest.ResponseWriter, r *rest.Request) {
	user := User{}
	err := r.DecodeJsonPayload(&user)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	id := fmt.Sprintf("%d", len(self.Store)) // stupid
	user.Id = id
	self.Store[id] = &user
	w.WriteJson(&user)
}

func main() {

	users := Users{
		Store: map[string]*User{},
	}

	handler := rest.ResourceHandler{}
	handler.SetRoutes(
		rest.RouteObjectMethod("GET", "/users/:id", &users, "GetUser"),
		rest.RouteObjectMethod("POST", "/users", &users, "PostUser"),
	)
	http.ListenAndServe(":8080", &handler)
}
