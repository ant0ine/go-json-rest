// A quick and easy way to setup a RESTful JSON API
//
// http://ant0ine.github.io/go-json-rest/
//
// Go-Json-Rest is a thin layer on top of net/http that helps building RESTful JSON APIs easily.
// It provides fast and scalable request routing using a Trie based implementation, helpers to deal
// with JSON requests and responses, and middlewares for functionalities like CORS, Auth, Gzip,
// Status, ...
//
// Example:
//
//      package main
//
//      import (
//              "github.com/ant0ine/go-json-rest/rest"
//              "log"
//              "net/http"
//      )
//
//      type User struct {
//              Id   string
//              Name string
//      }
//
//      func GetUser(w rest.ResponseWriter, req *rest.Request) {
//              user := User{
//                      Id:   req.PathParam("id"),
//                      Name: "Antoine",
//              }
//              w.WriteJson(&user)
//      }
//
//      func main() {
//              api := rest.NewApi()
//              api.Use(rest.DefaultDevStack...)
//              router, err := rest.MakeRouter(
//                      rest.Get("/users/:id", GetUser),
//              )
//              if err != nil {
//                      log.Fatal(err)
//              }
//              api.SetApp(router)
//              log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
//      }
//
//
package rest
