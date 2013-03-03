// A quick and easy way to setup a RESTful JSON API
//
// Go-Json-Rest is a thin layer on top of net/http that helps building RESTful JSON APIs easily.
// It provides fast URL routing using https://github.com/ant0ine/go-urlrouter, and helpers to deal
// with JSON requests and responses. It is not a high-level REST framework that transparently maps
// HTTP requests to procedure calls, on the opposite, you constantly have access to the underlying
// net/http objects.
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
//              w.WriteJson(&user)
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
	"fmt"
	"github.com/ant0ine/go-urlrouter"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
	"time"
)

// Implement the http.Handler interface and act as a router for the defined Routes.
// The defaults are intended to be developemnt friendly, for production you may want
// to turn on gzip and disable the JSON indentation.
type ResourceHandler struct {
	router urlrouter.Router

	// If true, and if the client accepts the Gzip encoding, the response payloads
	// will be compressed using gzip, and the corresponding response header will set.
	EnableGzip bool

	// If true, the JSON payload will be written in one line with no space.
	DisableJsonIndent bool

	// If true, when a "panic" happens, the error string and the stack trace will be
	// printed in the 500 response body.
	EnableResponseStackTrace bool

	// If true, the record that is logged for each response will be printed as JSON
	// in the log. Convenient for log parsing.
	EnableLogAsJson bool

	// Custom logger, defaults to log.New(os.Stderr, "", log.LstdFlags)
	Logger *log.Logger
}

// Used with SetRoutes.
type Route struct {

	// Any http method. It will be used as uppercase to avoid common mistakes.
	HttpMethod string

	// A string like "/resource/:id.json".
	// Placeholders supported are:
	// :param that matches any char to the first '/' or '.'
	// *splat that matches everything to the end of the string
	PathExp string

	// Code that will be executed when this route is taken.
	Func func(*ResponseWriter, *Request)
}

// Create a Route that points to an object method. It can be convenient to point to an object method instead
// of a function, this helper makes it easy by passing the object instance and the method name as parameters.
func RouteObjectMethod(http_method string, path_exp string, object_instance interface{}, object_method string) Route {

	value := reflect.ValueOf(object_instance)
	func_value := value.MethodByName(object_method)
	if func_value.IsValid() == false {
		panic(fmt.Sprintf(
			"Cannot find the object method %s on %s",
			object_method,
			value,
		))
	}
	route_func := func(w *ResponseWriter, r *Request) {
		func_value.Call([]reflect.Value{
			reflect.ValueOf(w),
			reflect.ValueOf(r),
		})
	}

	return Route{
		HttpMethod: http_method,
		PathExp:    path_exp,
		Func:       route_func,
	}
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
		http_method := strings.ToUpper(route.HttpMethod)

		self.router.Routes = append(
			self.router.Routes,
			urlrouter.Route{
				PathExp: http_method + route.PathExp,
				Dest:    route.Func,
			},
		)
	}
	return self.router.Start()
}

type response_log_record struct {
	StatusCode   int
	ResponseTime *time.Duration
	HttpMethod   string
	RequestURI   string
}

func (self *ResourceHandler) log_response_record(record *response_log_record) {
	if self.EnableLogAsJson {
		b, err := json.Marshal(record)
		if err != nil {
			panic(err)
		}
		self.Logger.Printf("%s", b)
	} else {
		self.Logger.Printf("%d %v %s %s",
			record.StatusCode,
			record.ResponseTime,
			record.HttpMethod,
			record.RequestURI,
		)
	}
}

func (self *ResourceHandler) log_response(status_code int, start *time.Time, request *http.Request) {

	duration := time.Now().Sub(*start)

	self.log_response_record(&response_log_record{
		status_code,
		&duration,
		request.Method,
		request.URL.RequestURI(),
	})
}

// This makes ResourceHandler implement the http.Handler interface.
// You probably don't want to use it directly.
func (self *ResourceHandler) ServeHTTP(orig_writer http.ResponseWriter, orig_request *http.Request) {

	start := time.Now()

	// set a default Logger
	if self.Logger == nil {
		self.Logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	// catch user code's panic, and convert to http response
	defer func() {
		if r := recover(); r != nil {
			trace := debug.Stack()

			// log the trace
			self.Logger.Printf("%s\n%s", r, trace)

			// write error response
			message := "Internal Server Error"
			if self.EnableResponseStackTrace {
				message = fmt.Sprintf("%s\n\n%s", r, trace)
			}
			http.Error(orig_writer, message, http.StatusInternalServerError)

			// log response
			self.log_response(
				http.StatusNotFound,
				&start,
				orig_request,
			)
		}
	}()

	// find the route
	route, params, err := self.router.FindRoute(
		strings.ToUpper(orig_request.Method) + orig_request.URL.Path,
	)
	if err != nil {
		// should never happen as the URL has already been parsed
		panic(err)
	}
	if route == nil {
		// no route found
		http.NotFound(orig_writer, orig_request)

		// log response
		self.log_response(
			http.StatusNotFound,
			&start,
			orig_request,
		)
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
		0,
		false,
	}

	// run the user code
	handler := route.Dest.(func(*ResponseWriter, *Request))
	handler(&writer, &request)

	// log response
	self.log_response(
		writer.status_code,
		&start,
		orig_request,
	)
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
func (self *Request) DecodeJsonPayload(v interface{}) error {
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

// Inherit from an object implementing the http.ResponseWriter interface,
// and provide additional methods.
type ResponseWriter struct {
	http.ResponseWriter
	is_gzipped   bool
	is_indented  bool
	status_code  int
	wrote_header bool
}

// Overloading of the http.ResponseWriter method.
// Just record the status code for logging.
func (self *ResponseWriter) WriteHeader(code int) {
	self.ResponseWriter.WriteHeader(code)
	self.status_code = code
	self.wrote_header = true
}

// Overloading of the http.ResponseWriter method.
// Provide additional capabilities, like transparent gzip encoding.
func (self *ResponseWriter) Write(b []byte) (int, error) {

	if !self.wrote_header {
		self.WriteHeader(http.StatusOK)
	}

	if self.is_gzipped {
		self.Header().Set("Content-Encoding", "gzip")
		gzip_writer := gzip.NewWriter(self.ResponseWriter)
		defer gzip_writer.Close()
		return gzip_writer.Write(b)
	}

	return self.ResponseWriter.Write(b)
}

// Encode the object in JSON, set the content-type header,
// and call Write.
func (self *ResponseWriter) WriteJson(v interface{}) error {
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
