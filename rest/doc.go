// A quick and easy way to setup a RESTful JSON API
//
// http://ant0ine.github.io/go-json-rest/
//
// Go-Json-Rest is a thin layer on top of net/http that helps building RESTful JSON APIs easily.
// It provides fast URL routing using a Trie based implementation, helpers to deal with JSON
// requests and responses, and middlewares for additional functionalities like CORS, Auth, Gzip ...
//
// Example:
//
//      package main
//
//      import (
//              "github.com/ant0ine/go-json-rest/rest"
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
//              handler := rest.ResourceHandler{}
//              handler.SetRoutes(
//                      rest.Route{"GET", "/users/:id", GetUser},
//              )
//              http.ListenAndServe(":8080", &handler)
//      }
//
//
// Note about the URL routing: Instead of using the usual
// "evaluate all the routes and return the first regexp that matches" strategy,
// it uses a Trie data structure to perform the routing. This is more efficient,
// and scales better for a large number of routes.
// It supports the :param and *splat placeholders in the route strings.
//
package rest
