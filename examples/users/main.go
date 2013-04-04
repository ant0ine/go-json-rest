/* Demonstrate how to use rest.RouteObjectMethod

rest.RouteObjectMethod helps create a Route that points to
an object method instead of just a function.

The Curl Demo:

        curl -i -d '{"Name":"Antoine"}' http://127.0.0.1:8080/users
        curl -i http://127.0.0.1:8080/users/0
        curl -i -X PUT -d '{"Name":"Antoine Imbert"}' http://127.0.0.1:8080/users/0
        curl -i -X DELETE http://127.0.0.1:8080/users/0
        curl -i http://127.0.0.1:8080/users

*/
package main

import (
	"fmt"
	"github.com/ant0ine/go-json-rest"
	"net/http"
)

func main() {

	users := Users{
		Store: map[string]*User{},
	}

	handler := rest.ResourceHandler{}
	handler.SetRoutes(
		rest.RouteObjectMethod("GET", "/users", &users, "GetAllUsers"),
		rest.RouteObjectMethod("POST", "/users", &users, "PostUser"),
		rest.RouteObjectMethod("GET", "/users/:id", &users, "GetUser"),
		rest.RouteObjectMethod("PUT", "/users/:id", &users, "PutUser"),
		rest.RouteObjectMethod("DELETE", "/users/:id", &users, "DeleteUser"),
	)
	http.ListenAndServe(":8080", &handler)
}

type User struct {
	Id   string
	Name string
}

type Users struct {
	Store map[string]*User
}

func (self *Users) GetAllUsers(w *rest.ResponseWriter, r *rest.Request) {
	users := make([]*User, len(self.Store))
	i := 0
	for _, user := range self.Store {
		users[i] = user
		i++
	}
	w.WriteJson(&users)
}

func (self *Users) GetUser(w *rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")
	user := self.Store[id]
	if user == nil {
		rest.NotFound(w, r)
		return
	}
	w.WriteJson(&user)
}

func (self *Users) PostUser(w *rest.ResponseWriter, r *rest.Request) {
	user := User{}
	err := r.DecodeJsonPayload(&user)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := fmt.Sprintf("%d", len(self.Store)) // stupid
	user.Id = id
	self.Store[id] = &user
	w.WriteJson(&user)
}

func (self *Users) PutUser(w *rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")
	if self.Store[id] == nil {
		rest.NotFound(w, r)
		return
	}
	user := User{}
	err := r.DecodeJsonPayload(&user)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user.Id = id
	self.Store[id] = &user
	w.WriteJson(&user)
}

func (self *Users) DeleteUser(w *rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")
	delete(self.Store, id)
}
