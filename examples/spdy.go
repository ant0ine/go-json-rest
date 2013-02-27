// Demonstrate how to use SPDY with github.com/shykes/spdy-go
//
// For a command line client, install spdycat from:
// https://github.com/tatsuhiro-t/spdylay
//
// Then:
//
// spdycat -v --no-tls -2 http://localhost:8080/users/0
//
package main

import (
	"github.com/ant0ine/go-json-rest"
	"github.com/shykes/spdy-go"
	"log"
)

type User struct {
	Id   string
	Name string
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
		rest.Route{"GET", "/users/:id", GetUser},
	)
	log.Fatal(spdy.ListenAndServeTCP(":8080", &handler))
}
