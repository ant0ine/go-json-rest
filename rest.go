// A quick and easy way to setup a RESTful JSON API
//
// Go-JSON-REST is a thin layer on top of net/http that helps building RESTful JSON APIs easily.
// It provides fast URL routing using go-urlrouter, and helpers to deal with JSON requests and responses.
// It is not a high-level REST framework that transparently maps HTTP requests to language procedure calls,
// on the opposite, you constantly have access to the underlying net/http objects.
//
// Example:
//
//      package main
//
//      import (
//              "github.com/ant0ine/go-json-rest"
//              "net/http"
//      )
//
//      type User struct {
//              Id   string
//              Name string
//      }
//
//      func GetUser(w *rest.ResponseWriter, req *rest.Request) {
//              user := User{
//                      Id:   req.PathParam("id"),
//                      Name: "Antoine",
//              }
//              w.WriteJSON(&user)
//      }
//
//      func main() {
//              handler := rest.NewResourceHandler(
//                      rest.Route{"GET", "/users/:id", GetUser},
//              )
//              http.ListenAndServe(":8080", handler)
//      }
//
package rest

import (
	"encoding/json"
	"github.com/ant0ine/go-urlrouter"
	"io/ioutil"
	"net/http"
	"strings"
)

// TODO
// * concatenating the method and the path in the router is kind of hacky,
//   maybe url-router should evolve to take the method

// NICETOHAVE
// * more friendly log or output
// * offer option for JSON indentation
// * offer option for gzipped output

// Implement the http.Handler interface and act as a router for the defined Routes.
type ResourceHandler struct {
	router urlrouter.Router
}

// Used during the instanciation of the ResourceHandler to define the Routes.
type Route struct {
	Method  string
	PathExp string
	Dest    func(*ResponseWriter, *Request)
}

// Instanciate a new ResourceHandler. The order the Routes matters,
// if a request matches multiple Routes, the first one will be used.
// Note that the underlying router is go-urlrouter.
func NewResourceHandler(routes ...Route) *ResourceHandler {
	self := ResourceHandler{
		router: urlrouter.Router{
			Routes: []urlrouter.Route{},
		},
	}
	for _, route := range routes {
		// make sure the method is uppercase
		method := strings.ToUpper(route.Method)

		self.router.Routes = append(
			self.router.Routes,
			urlrouter.Route{
				PathExp: method + route.PathExp,
				Dest:    route.Dest,
			},
		)
	}
	self.router.Start()
	return &self
}

// This makes ResourceHandler implement the http.Handler interface
func (self *ResourceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route, params, err := self.router.FindRoute(
		strings.ToUpper(r.Method) + r.URL.Path,
	)
	if err != nil {
		// should never happend has the URL has already been parsed
		panic(err)
	}
	if route == nil {
		// no route found
		http.NotFound(w, r)
		return
	}
	request := Request{r, params}
	writer := ResponseWriter{w}
	handler := route.Dest.(func(*ResponseWriter, *Request))
	handler(&writer, &request)
}

// Inherit from http.Request, and provide additional methods.
type Request struct {
	*http.Request
	// map of parameters that have been matched in the URL Path.
	PathParams map[string]string
}

// Provide a convenient access to the PathParams map
func (self *Request) PathParam(name string) string {
	return self.PathParams[name]
}

// Read the request body and decode the JSON using json.Unmarshal
func (self *Request) DecodeJSONPayload(v interface{}) error {
	content, err := ioutil.ReadAll(self.Body)
	self.Body.Close()
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, v)
	if err != nil {
		return err
	}
	return nil
}

// Inherit from a http.ResponseWriter interface, and provide additional methods.
type ResponseWriter struct {
	http.ResponseWriter
}

// Encode the object in JSON using json.Marshal, set the content-type header,
// and write the response.
func (self *ResponseWriter) WriteJSON(v interface{}) error {
	self.Header().Set("content-type", "application/json")
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	self.Write(b)
	return nil
}
