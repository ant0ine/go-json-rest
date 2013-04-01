package main

import (
	"github.com/ant0ine/go-json-rest"
	"net/http"
)

type User struct {
	Id   string
	Name string
}

func GetRedirect(w *rest.ResponseWriter, req *rest.Request) {
	w.Redirect("/users/1", 302)
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
		rest.Route{"GET", "/utilisateurs/:id", GetRedirect},
		rest.Route{"GET", "/users/:id", GetUser},
	)
	http.ListenAndServe(":8080", &handler)
}
