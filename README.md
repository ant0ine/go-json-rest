
Go-Json-Rest
============

*A quick and easy way to setup a RESTful JSON API*

**Version 2 of Go-Json-Rest is currently in development, you can try it by checking out the [“v2-alpha” branch](https://github.com/ant0ine/go-json-rest/tree/v2-alpha). Thanks in advance for your feedback.**

[![Build Status](https://travis-ci.org/ant0ine/go-json-rest.png?branch=master)](https://travis-ci.org/ant0ine/go-json-rest)

**Go-Json-Rest** is a thin layer on top of `net/http` that helps building RESTful JSON APIs easily. It provides fast URL routing using a Trie based implementation, and helpers to deal with JSON requests and responses. It is not a high-level REST framework that transparently maps HTTP requests to procedure calls, on the opposite, you constantly have access to the underlying
`net/http` objects.

Features
-----------
- Implemented as a `net/http` Handler. This standard interface allows combinations with other Handlers.
- Fast URL routing. It implements the classic route description syntax using a fast and scalable trie data structure.
- Test package to help writing tests for the API.
- Optional /.status endpoint for easy monitoring.
- Examples

Install
-------

This package is "go-gettable", just do:

    go get github.com/ant0ine/go-json-rest

Vendoring
---------

The recommended way of using this library in your project is to use the **"vendoring"** method,
where this library code is copied in your repository at a specific revision.
[This page](http://nathany.com/go-packages/) is a good summary of package management in Go.

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

- [Countries](https://github.com/ant0ine/go-json-rest-examples/blob/master/countries/main.go) Demo very simple GET, POST, DELETE operations
- [Users](https://github.com/ant0ine/go-json-rest-examples/blob/master/users/main.go) Demo the mapping to object methods
- [SPDY](https://github.com/ant0ine/go-json-rest-examples/blob/master/spdy/main.go) Demo SPDY using github.com/shykes/spdy-go
- [GAE](https://github.com/ant0ine/go-json-rest-examples/tree/master/gae) Demo go-json-rest on Google App Engine
- [GORM](https://github.com/ant0ine/go-json-rest-examples/blob/master/gorm/main.go) Demo basic CRUD operations using MySQL and GORM
- [Streaming](https://github.com/ant0ine/go-json-rest-examples/blob/master/streaming/main.go) Demo Line Delimited JSON stream
- [CORS](https://github.com/ant0ine/go-json-rest-examples/blob/master/cors/main.go) Demo CORS support for all endpoints
- [Basic Auth](https://github.com/ant0ine/go-json-rest-examples/blob/master/auth-basic/main.go) Demo an Authentication Basic impl for all endpoints


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

Thanks
------
- [Franck Cuny](https://github.com/franckcuny)
- [Yann Kerhervé](https://github.com/yannk)
- [Ask Bjørn Hansen](https://github.com/abh)


Copyright (c) 2013-2014 Antoine Imbert

[MIT License](https://github.com/ant0ine/go-json-rest/blob/master/LICENSE)

[![Analytics](https://ga-beacon.appspot.com/UA-309210-4/go-json-rest/readme)](https://github.com/igrigorik/ga-beacon)


