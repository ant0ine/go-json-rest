// A quick and easy way to setup a RESTful JSON API
//
// Go-JSON-REST is a thin layer on top of net/http that helps building RESTful JSON APIs easily.
// It provides fast URL routing using https://github.com/ant0ine/go-urlrouter, and helpers to deal
// with JSON requests and responses. It is not a high-level REST framework that transparently maps
// HTTP requests to language procedure calls, on the opposite, you constantly have access to the
// underlying net/http objects.
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
//              handler := ResourceHandler{}
//              handler.SetRoutes(
//                      rest.Route{"GET", "/users/:id", GetUser},
//              )
//              http.ListenAndServe(":8080", &handler)
//      }
//
package rest

import (
	"compress/gzip"
	"encoding/json"
	"github.com/ant0ine/go-urlrouter"
	"io/ioutil"
	"net/http"
	"strings"
)

// Implement the http.Handler interface and act as a router for the defined Routes.
// The defaults are intended to be developemnt friendly, for production you may want
// to turn on gzip and disable the JSON indentation.
type ResourceHandler struct {
	router urlrouter.Router
	// If true and if the client accepts the Gzip encoding, the response payloads
	// will be compressed using gzip, and the corresponding response header will set.
	EnableGzip bool
	// If true the JSON payload will be written in one line with no space.
	DisableJsonIndent bool
}

// Used with SetRoutes.
type Route struct {
	Method  string
	PathExp string
	Dest    func(*ResponseWriter, *Request)
}

// Define the Routes. The order the Routes matters,
// if a request matches multiple Routes, the first one will be used.
// Note that the underlying router is https://github.com/ant0ine/go-urlrouter.
func (self *ResourceHandler) SetRoutes(routes ...Route) error {
	self.router = urlrouter.Router{
		Routes: []urlrouter.Route{},
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
	return self.router.Start()
}

// This makes ResourceHandler implement the http.Handler interface
func (self *ResourceHandler) ServeHTTP(orig_writer http.ResponseWriter, orig_request *http.Request) {

	// find the route
	route, params, err := self.router.FindRoute(
		strings.ToUpper(orig_request.Method) + orig_request.URL.Path,
	)
	if err != nil {
		// should never happen has the URL has already been parsed
		panic(err)
	}
	if route == nil {
		// no route found
		http.NotFound(orig_writer, orig_request)
		return
	}

	// determine if gzip is needed
	is_gzipped := self.EnableGzip == true &&
		strings.Contains(orig_request.Header.Get("Accept-Encoding"), "gzip")

	is_indented := !self.DisableJsonIndent

	request := Request{
		orig_request,
		params,
	}

	writer := ResponseWriter{
		orig_writer,
		is_gzipped,
		is_indented,
	}

	// run the user code
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

// Inherit from an object implementing the http.ResponseWriter interface, and provide additional methods.
type ResponseWriter struct {
	http.ResponseWriter
	is_gzipped  bool
	is_indented bool
}

// Overloading of the http.ResponseWriter method.
// Provide additional capabilities, like transparent gzip encoding.
func (self *ResponseWriter) Write(b []byte) (int, error) {
	if self.is_gzipped {
		self.Header().Set("Content-Encoding", "gzip")
		gzip_writer := gzip.NewWriter(self.ResponseWriter)
		defer gzip_writer.Close()
		return gzip_writer.Write(b)
	}
	return self.ResponseWriter.Write(b)
}

// Encode the object in JSON, set the content-type header,
// and call Write
func (self *ResponseWriter) WriteJSON(v interface{}) error {
	self.Header().Set("content-type", "application/json")
	var b []byte
	var err error
	if self.is_indented {
		b, err = json.MarshalIndent(v, "", "  ")
	} else {
		b, err = json.Marshal(v)
	}
	if err != nil {
		return err
	}
	self.Write(b)
	return nil
}
