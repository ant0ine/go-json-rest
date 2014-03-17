
Go-Json-Rest
============

*A quick and easy way to setup a RESTful JSON API*

[![Build Status](https://travis-ci.org/ant0ine/go-json-rest.png?branch=master)](https://travis-ci.org/ant0ine/go-json-rest)

**Go-Json-Rest** is a thin layer on top of `net/http` that helps building RESTful JSON APIs easily. It provides fast URL routing using a Trie based implementation, helpers to deal with JSON requests and responses, and middlewares for additional functionalities like CORS, Auth, Gzip ...


Features
--------
- Implemented as a `net/http` Handler. This standard interface allows combinations with other Handlers.
- Fast URL routing. It implements the classic route description syntax using a fast and scalable trie data structure.
- Test package to help writing tests for the API.
- Monitoring statistics.
- Many examples.


Install
-------

This package is "go-gettable", just do:

    go get github.com/ant0ine/go-json-rest


Example
-------

~~~ go
package main
import (
        "github.com/ant0ine/go-json-rest"
        "net/http"
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
        http.ListenAndServe(":8080", &handler)
}
~~~


More Examples
-------------

(See the dedicated examples repository: https://github.com/ant0ine/go-json-rest-examples)

- [Countries](https://github.com/ant0ine/go-json-rest-examples/blob/v2-alpha/countries/main.go) Demo very simple GET, POST, DELETE operations
- [Users](https://github.com/ant0ine/go-json-rest-examples/blob/v2-alpha/users/main.go) Demo the mapping to object methods
- [SPDY](https://github.com/ant0ine/go-json-rest-examples/blob/v2-alpha/spdy/main.go) Demo SPDY using github.com/shykes/spdy-go
- [GAE](https://github.com/ant0ine/go-json-rest-examples/tree/v2-alpha/gae) Demo go-json-rest on Google App Engine
- [GORM](https://github.com/ant0ine/go-json-rest-examples/blob/v2-alpha/gorm/main.go) Demo basic CRUD operations using MySQL and GORM
- [Streaming](https://github.com/ant0ine/go-json-rest-examples/blob/v2-alpha/streaming/main.go) Demo Line Delimited JSON stream
- [CORS](https://github.com/ant0ine/go-json-rest-examples/blob/v2-alpha/cors/main.go) Demo CORS support for all endpoints
- [Basic Auth](https://github.com/ant0ine/go-json-rest-examples/blob/v2-alpha/auth-basic/main.go) Demo an Authentication Basic impl for all endpoints
- [Status](https://github.com/ant0ine/go-json-rest-examples/blob/v2-alpha/status/main.go) Demo how to setup the /.status endpoint


Documentation
-------------

- [Online Documentation (godoc.org)](http://godoc.org/github.com/ant0ine/go-json-rest)
- [(Blog Post) Introducing Go-Json-Rest] (http://blog.ant0ine.com/typepad/2013/04/introducing-go-json-rest.html)
- [(Blog Post) Better URL Routing ?] (http://blog.ant0ine.com/typepad/2013/02/better-url-routing-golang-1.html)


Options
-------

Things to enable in production:
- Gzip compression (default: disabled)
- Custom Logger (default: Go default)

Things to enable in development:
- Json indentation (default: enabled)
- Relaxed ContentType (default: disabled)
- Error stack trace in the response body (default: disabled)


The Status Endpoint
-------------------

Inspired by memcached "stats", this optional feature can be enabled to help monitoring the service.
See the "status" example to install the following status route:

GET /.status returns something like:

    {
      "Pid": 21732,
      "UpTime": "1m15.926272s",
      "UpTimeSec": 75.926272,
      "Time": "2013-03-04 08:00:27.152986 +0000 UTC",
      "TimeUnix": 1362384027,
      "StatusCodeCount": {
        "200": 53,
        "404": 11
      },
      "TotalCount": 64,
      "TotalResponseTime": "16.777ms",
      "TotalResponseTimeSec": 0.016777,
      "AverageResponseTime": "262.14us",
      "AverageResponseTimeSec": 0.00026214
    }


Migration from v1 to v2
-----------------------

A few breaking changes have been introduced to the v2. (go-json-rest follows [Semver](http://semver.org/))

- rest.ResponseWriter is now an interface. This is the main change, and most of the program will be migrated with a simple s/\*\.rest\.ResponseWriter/rest\.ResponseWriter/g
- Flush(), CloseNotify() and Write() are not directly exposed anymore. A type assertion of the corresponding interface is necessary. eg: writer.(http.Flusher).Flush()
- The /.status endpoint is not created automatically anymore. The route has to be manually set as shown on the "status" example.
- The notion of Middleware is now formally defined, and code using PreRoutingMiddleware will have to be adapted to provide a list of Middleware objects. See the [Basic Auth example](https://github.com/ant0ine/go-json-rest-examples/blob/v2-alpha/auth-basic/main.go).


Thanks
------
- [Franck Cuny](https://github.com/franckcuny)
- [Yann Kerhervé](https://github.com/yannk)
- [Ask Bjørn Hansen](https://github.com/abh)


Copyright (c) 2013-2014 Antoine Imbert

[MIT License](https://github.com/ant0ine/go-json-rest/blob/master/LICENSE)

[![Analytics](https://ga-beacon.appspot.com/UA-309210-4/go-json-rest/v2-alpha/readme)](https://github.com/igrigorik/ga-beacon)


