package main

import (
	"github.com/ant0ine/go-json-rest"
)

type User struct {
	Id   string
	Name string
}

func GetOldAPIUser(w *rest.ResponseWriter, req *rest.Request) {
	http.Redirect(w, req.Request, req.UriFor("/users/1"), 302)
}

func GetUser(w *rest.ResponseWriter, req *rest.Request) {
	user := User{
		Id:   req.PathParam("id"),
		Name: "Antoine",
	}
	w.WriteJson(&user)
}

func main() {
	handler := rest.ResourceHandler{}
	handler.SetRoutes(
		rest.Route{"GET", "/user/:id", GetOldAPIUser},
		rest.Route{"GET", "/users/:id", GetUser},
	)
	handler.Start(":8080")
}
