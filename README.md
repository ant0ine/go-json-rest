
# Go-Json-Rest

*A quick and easy way to setup a RESTful JSON API*

[![Build Status](https://travis-ci.org/ant0ine/go-json-rest.png?branch=master)](https://travis-ci.org/ant0ine/go-json-rest) [![GoDoc](https://godoc.org/github.com/ant0ine/go-json-rest?status.png)](https://godoc.org/github.com/ant0ine/go-json-rest)


**Warning: This is v2-alpha, a work in progress for the version 2 of Go-Json-Rest**

**Go-Json-Rest** is a thin layer on top of `net/http` that helps building RESTful JSON APIs easily. It provides fast URL routing using a Trie based implementation, helpers to deal with JSON requests and responses, and middlewares for additional functionalities like CORS, Auth, Gzip ...


## Table of content

- [Features](#features)
- [Install](#install)
- [Vendoring](#vendoring)
- [Examples](#examples)
  - [Hello World!](#hello-world)
  - [Countries](#countries)
  - [Users](#users)
  - [GORM](#gorm)
  - [CORS](#cors)
  - [Basic Auth](#basic-auth)
  - [Status](#status)
  - [Status Auth](#status-auth)
  - [Streaming](#streaming)
  - [SPDY](#spdy)
  - [Basic Auth Custom](#basic-auth-custom)
  - [CORS Custom](#cors-custom)
- [External Documentation](#external-documentation)
- [Options](#options)
- [Migration guide from v1 to v2](#migration-guide-from-v1-to-v2)
- [Thanks](#thanks)


## Features

- Many examples.
- Fast URL routing. It implements the classic route description syntax using a fast and scalable trie data structure.
- Use Middlewares in order to extend the functionalities.
- Implemented as a `net/http` Handler. This standard interface allows combinations with other Handlers.
- Test package to help writing tests for the API.
- Monitoring statistics inspired by Memcached.


## Install

This package is "go-gettable", just do:

    go get github.com/ant0ine/go-json-rest/rest


## Vendoring

The recommended way of using this library in your project is to use the **"vendoring"** method,
where this library code is copied in your repository at a specific revision.
[This page](http://nathany.com/go-packages/) is a good summary of package management in Go.


## Examples

(See the dedicated examples repository: https://github.com/ant0ine/go-json-rest-examples)

#### Hello World!

Tradition!

The minimal example: Hello World!

The Curl Demo:

        curl -i http://127.0.0.1:8080/message



~~~ go
/* */
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
)

type Message struct {
	Body   string
}

func main() {
	handler := rest.ResourceHandler{}
	handler.SetRoutes(
		rest.Route{"GET", "/message",
                        func(w rest.ResponseWriter, req *rest.Request) {
                                w.WriteJson(&Message{
                                        Body: "Hello World!",
                                })
                        },
                },
	)
	http.ListenAndServe(":8080", &handler)
}

~~~


#### Countries

Demo very simple GET, POST, DELETE operations

Demonstrate simple POST GET and DELETE operations

The Curl Demo:

        curl -i -d '{"Code":"FR","Name":"France"}' http://127.0.0.1:8080/countries
        curl -i -d '{"Code":"US","Name":"United States"}' http://127.0.0.1:8080/countries
        curl -i http://127.0.0.1:8080/countries/FR
        curl -i http://127.0.0.1:8080/countries/US
        curl -i http://127.0.0.1:8080/countries
        curl -i -X DELETE http://127.0.0.1:8080/countries/FR
        curl -i http://127.0.0.1:8080/countries
        curl -i -X DELETE http://127.0.0.1:8080/countries/US
        curl -i http://127.0.0.1:8080/countries



~~~ go
/* */
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
)

func main() {

	handler := rest.ResourceHandler{
                EnableRelaxedContentType: true,
        }
	handler.SetRoutes(
		rest.Route{"GET", "/countries", GetAllCountries},
		rest.Route{"POST", "/countries", PostCountry},
		rest.Route{"GET", "/countries/:code", GetCountry},
		rest.Route{"DELETE", "/countries/:code", DeleteCountry},
	)
	http.ListenAndServe(":8080", &handler)
}

type Country struct {
	Code string
	Name string
}

var store = map[string]*Country{}

func GetCountry(w rest.ResponseWriter, r *rest.Request) {
	code := r.PathParam("code")
	country := store[code]
	if country == nil {
		rest.NotFound(w, r)
		return
	}
	w.WriteJson(&country)
}

func GetAllCountries(w rest.ResponseWriter, r *rest.Request) {
	countries := make([]*Country, len(store))
	i := 0
	for _, country := range store {
		countries[i] = country
		i++
	}
	w.WriteJson(&countries)
}

func PostCountry(w rest.ResponseWriter, r *rest.Request) {
	country := Country{}
	err := r.DecodeJsonPayload(&country)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if country.Code == "" {
		rest.Error(w, "country code required", 400)
		return
	}
	if country.Name == "" {
		rest.Error(w, "country name required", 400)
		return
	}
	store[country.Code] = &country
	w.WriteJson(&country)
}

func DeleteCountry(w rest.ResponseWriter, r *rest.Request) {
	code := r.PathParam("code")
	delete(store, code)
}

~~~


#### Users

Demo the mapping to object methods

Demonstrate how to use rest.RouteObjectMethod

rest.RouteObjectMethod helps create a Route that points to
an object method instead of just a function.

The Curl Demo:

        curl -i -d '{"Name":"Antoine"}' http://127.0.0.1:8080/users
        curl -i http://127.0.0.1:8080/users/0
        curl -i -X PUT -d '{"Name":"Antoine Imbert"}' http://127.0.0.1:8080/users/0
        curl -i -X DELETE http://127.0.0.1:8080/users/0
        curl -i http://127.0.0.1:8080/users



~~~ go
/* */
package main

import (
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
)

func main() {

	users := Users{
		Store: map[string]*User{},
	}

	handler := rest.ResourceHandler{
                EnableRelaxedContentType: true,
        }
	handler.SetRoutes(
		rest.RouteObjectMethod("GET", "/users", &users, "GetAllUsers"),
		rest.RouteObjectMethod("POST", "/users", &users, "PostUser"),
		rest.RouteObjectMethod("GET", "/users/:id", &users, "GetUser"),
		rest.RouteObjectMethod("PUT", "/users/:id", &users, "PutUser"),
		rest.RouteObjectMethod("DELETE", "/users/:id", &users, "DeleteUser"),
	)
	http.ListenAndServe(":8080", &handler)
}

type User struct {
	Id   string
	Name string
}

type Users struct {
	Store map[string]*User
}

func (self *Users) GetAllUsers(w rest.ResponseWriter, r *rest.Request) {
	users := make([]*User, len(self.Store))
	i := 0
	for _, user := range self.Store {
		users[i] = user
		i++
	}
	w.WriteJson(&users)
}

func (self *Users) GetUser(w rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")
	user := self.Store[id]
	if user == nil {
		rest.NotFound(w, r)
		return
	}
	w.WriteJson(&user)
}

func (self *Users) PostUser(w rest.ResponseWriter, r *rest.Request) {
	user := User{}
	err := r.DecodeJsonPayload(&user)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := fmt.Sprintf("%d", len(self.Store)) // stupid
	user.Id = id
	self.Store[id] = &user
	w.WriteJson(&user)
}

func (self *Users) PutUser(w rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")
	if self.Store[id] == nil {
		rest.NotFound(w, r)
		return
	}
	user := User{}
	err := r.DecodeJsonPayload(&user)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user.Id = id
	self.Store[id] = &user
	w.WriteJson(&user)
}

func (self *Users) DeleteUser(w rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")
	delete(self.Store, id)
}

~~~


#### GORM

Demo basic CRUD operations using MySQL and GORM

Demonstrate basic CRUD operation using a store based on MySQL and GORM

The Curl Demo:

        curl -i -d '{"Message":"this is a test"}' http://127.0.0.1:8080/reminders
        curl -i http://127.0.0.1:8080/reminders/1
        curl -i http://127.0.0.1:8080/reminders
        curl -i -X PUT -d '{"Message":"is updated"}' http://127.0.0.1:8080/reminders/1
        curl -i -X DELETE http://127.0.0.1:8080/reminders/1



~~~ go
/* */
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
	"time"
)

func main() {

	api := Api{}
	api.InitDB()
	api.InitSchema()

	handler := rest.ResourceHandler{
		EnableRelaxedContentType: true,
	}
	handler.SetRoutes(
		rest.RouteObjectMethod("GET", "/reminders", &api, "GetAllReminders"),
		rest.RouteObjectMethod("POST", "/reminders", &api, "PostReminder"),
		rest.RouteObjectMethod("GET", "/reminders/:id", &api, "GetReminder"),
		rest.RouteObjectMethod("PUT", "/reminders/:id", &api, "PutReminder"),
		rest.RouteObjectMethod("DELETE", "/reminders/:id", &api, "DeleteReminder"),
	)
	http.ListenAndServe(":8080", &handler)
}

type Reminder struct {
	Id        int64     `json:"id"`
	Message   string    `sql:"size:1024" json:"message"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt time.Time `json:"-"`
}

type Api struct {
	DB gorm.DB
}

func (api *Api) InitDB() {
	var err error
	api.DB, err = gorm.Open("mysql", "gorm:gorm@/gorm?charset=utf8&parseTime=True")
	if err != nil {
		log.Fatalf("Got error when connect database, the error is '%v'", err)
	}
	api.DB.LogMode(true)
}

func (api *Api) InitSchema() {
	api.DB.AutoMigrate(Reminder{})
}

func (api *Api) GetAllReminders(w rest.ResponseWriter, r *rest.Request) {
	reminders := []Reminder{}
	api.DB.Find(&reminders)
	w.WriteJson(&reminders)
}

func (api *Api) GetReminder(w rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")
	reminder := Reminder{}
	if api.DB.First(&reminder, id).Error != nil {
		rest.NotFound(w, r)
		return
	}
	w.WriteJson(&reminder)
}

func (api *Api) PostReminder(w rest.ResponseWriter, r *rest.Request) {
	reminder := Reminder{}
	if err := r.DecodeJsonPayload(&reminder); err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := api.DB.Save(&reminder).Error; err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(&reminder)
}

func (api *Api) PutReminder(w rest.ResponseWriter, r *rest.Request) {

	id := r.PathParam("id")
	reminder := Reminder{}
	if api.DB.First(&reminder, id).Error != nil {
		rest.NotFound(w, r)
		return
	}

	updated := Reminder{}
	if err := r.DecodeJsonPayload(&updated); err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reminder.Message = updated.Message

	if err := api.DB.Save(&reminder).Error; err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(&reminder)
}

func (api *Api) DeleteReminder(w rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")
	reminder := Reminder{}
	if api.DB.First(&reminder, id).Error != nil {
		rest.NotFound(w, r)
		return
	}
	if err := api.DB.Delete(&reminder).Error; err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

~~~


#### CORS

Demo how to setup CorsMiddleware as a pre-routing middleware

Demonstrate how to setup CorsMiddleware around all the API endpoints.

The Curl Demo:

        curl -i http://127.0.0.1:8080/countries



~~~ go
/* */
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
)

func main() {

	handler := rest.ResourceHandler{
		PreRoutingMiddlewares: []rest.Middleware{
			&rest.CorsMiddleware{
				RejectNonCorsRequests: false,
				OriginValidator: func(origin string, request *rest.Request) bool {
					return origin == "http://my.other.host"
				},
				AllowedMethods:                []string{"GET", "POST", "PUT"},
				AllowedHeaders:                []string{"Accept", "Content-Type", "X-Custom-Header"},
				AccessControlAllowCredentials: true,
				AccessControlMaxAge:           3600,
			},
		},
	}
	handler.SetRoutes(
		rest.Route{"GET", "/countries", GetAllCountries},
	)
	http.ListenAndServe(":8080", &handler)
}

type Country struct {
	Code string
	Name string
}

func GetAllCountries(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(
		[]Country{
			Country{
				Code: "FR",
				Name: "France",
			},
			Country{
				Code: "US",
				Name: "United States",
			},
		},
	)
}

~~~


#### Basic Auth

Demo how to setup AuthBasicMiddleware as a pre-routing middleware

Demonstrate how to setup AuthBasicMiddleware as a pre-routing middleware.

The Curl Demo:

        curl -i http://127.0.0.1:8080/countries
        curl -i -u admin:admin http://127.0.0.1:8080/countries



~~~ go
/* */
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
)

func main() {

	handler := rest.ResourceHandler{
		PreRoutingMiddlewares: []rest.Middleware{
			&rest.AuthBasicMiddleware{
				Realm: "test zone",
				Authenticator: func(userId string, password string) bool {
					if userId == "admin" && password == "admin" {
						return true
					}
					return false
				},
			},
		},
	}
	handler.SetRoutes(
		rest.Route{"GET", "/countries", GetAllCountries},
	)
	http.ListenAndServe(":8080", &handler)
}

type Country struct {
	Code string
	Name string
}

func GetAllCountries(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(
		[]Country{
			Country{
				Code: "FR",
				Name: "France",
			},
			Country{
				Code: "US",
				Name: "United States",
			},
		},
	)
}

~~~


#### Status

Demo how to setup the /.status endpoint

Inspired by memcached "stats", this optional feature can be enabled to help monitoring the service.
See the "status" example to install the following status route:

GET /.status returns something like:

~~~ json
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
~~~

Demonstrate how to setup a /.status endpoint

The Curl Demo:

        curl -i http://127.0.0.1:8080/.status
        curl -i http://127.0.0.1:8080/.status
        ...



~~~ go
/* */
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
)

func main() {
	handler := rest.ResourceHandler{
                EnableStatusService: true,
        }
	handler.SetRoutes(
		rest.Route{"GET", "/.status",
			func(w rest.ResponseWriter, r *rest.Request) {
				w.WriteJson(handler.GetStatus())
			},
                },
	)
	http.ListenAndServe(":8080", &handler)
}

~~~


#### Status Auth

Demo how to setup the /.status endpoint protected with basic authentication

Demonstrate how to setup a /.status endpoint protected with basic authentication.

This is a good use case of middleware applied to only one API endpoint.

The Curl Demo:

        curl -i http://127.0.0.1:8080/countries
        curl -i http://127.0.0.1:8080/.status
        curl -i -u admin:admin http://127.0.0.1:8080/.status
        ...



~~~ go
/* */
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
)

func main() {
	handler := rest.ResourceHandler{
		EnableStatusService: true,
	}
	auth := &rest.AuthBasicMiddleware{
		Realm: "test zone",
		Authenticator: func(userId string, password string) bool {
			if userId == "admin" && password == "admin" {
				return true
			}
			return false
		},
	}
	handler.SetRoutes(
		rest.Route{"GET", "/countries", GetAllCountries},
		rest.Route{"GET", "/.status",
			auth.MiddlewareFunc(
				func(w rest.ResponseWriter, r *rest.Request) {
					w.WriteJson(handler.GetStatus())
				},
			),
		},
	)
	http.ListenAndServe(":8080", &handler)
}

type Country struct {
	Code string
	Name string
}

func GetAllCountries(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(
		[]Country{
			Country{
				Code: "FR",
				Name: "France",
			},
			Country{
				Code: "US",
				Name: "United States",
			},
		},
	)
}

~~~


#### Streaming

Demo Line Delimited JSON stream

Demonstrate a streaming REST API, where the data is "flushed" to the client ASAP.

The stream format is a Line Delimited JSON.

The Curl Demo:

        curl -i http://127.0.0.1:8080/stream

        HTTP/1.1 200 OK
        Content-Type: application/json
        Date: Sun, 16 Feb 2014 00:39:19 GMT
        Transfer-Encoding: chunked

        {"Name":"thing #1"}
        {"Name":"thing #2"}
        {"Name":"thing #3"}



~~~ go
/* */
package main

import (
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
	"time"
)

func main() {

	handler := rest.ResourceHandler{
		EnableRelaxedContentType: true,
		DisableJsonIndent:        true,
	}
	handler.SetRoutes(
		rest.Route{"GET", "/stream", StreamThings},
	)
	http.ListenAndServe(":8080", &handler)
}

type Thing struct {
	Name string
}

func StreamThings(w rest.ResponseWriter, r *rest.Request) {
	cpt := 0
	for {
		cpt++
		w.WriteJson(
			&Thing{
				Name: fmt.Sprintf("thing #%d", cpt),
			},
		)
		w.(http.ResponseWriter).Write([]byte("\n"))
		// Flush the buffer to client
		w.(http.Flusher).Flush()
		// wait 3 seconds
		time.Sleep(time.Duration(3) * time.Second)
	}
}

~~~


#### SPDY

Demo SPDY using raw.githubusercontent.com/shykes/spdy-go



~~~ go
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
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/shykes/spdy-go"
	"log"
)

type User struct {
	Id   string
	Name string
}

func GetUser(w rest.ResponseWriter, req *rest.Request) {
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

~~~


#### GAE

Demo go-json-rest on Google App Engine

Demonstrate a simple Google App Engine app

The Curl Demo:

        curl -i -d '{"Code":"FR","Name":"France"}' http://127.0.0.1:8080/countries
        curl -i -d '{"Code":"US","Name":"United States"}' http://127.0.0.1:8080/countries
        curl -i http://127.0.0.1:8080/countries/FR
        curl -i http://127.0.0.1:8080/countries/US
        curl -i http://127.0.0.1:8080/countries
        curl -i -X DELETE http://127.0.0.1:8080/countries/FR
        curl -i http://127.0.0.1:8080/countries
        curl -i -X DELETE http://127.0.0.1:8080/countries/US
        curl -i http://127.0.0.1:8080/countries



~~~ go
/* */
package gaecountries

import (
	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
)

func init() {

	handler := rest.ResourceHandler{}
	handler.SetRoutes(
		rest.Route{"GET", "/countries", GetAllCountries},
		rest.Route{"POST", "/countries", PostCountry},
		rest.Route{"GET", "/countries/:code", GetCountry},
		rest.Route{"DELETE", "/countries/:code", DeleteCountry},
	)
	http.Handle("/", &handler)
}

type Country struct {
	Code string
	Name string
}

var store = map[string]*Country{}

func GetCountry(w rest.ResponseWriter, r *rest.Request) {
	code := r.PathParam("code")
	country := store[code]
	if country == nil {
		rest.NotFound(w, r)
		return
	}
	w.WriteJson(&country)
}

func GetAllCountries(w rest.ResponseWriter, r *rest.Request) {
	countries := make([]*Country, len(store))
	i := 0
	for _, country := range store {
		countries[i] = country
		i++
	}
	w.WriteJson(&countries)
}

func PostCountry(w rest.ResponseWriter, r *rest.Request) {
	country := Country{}
	err := r.DecodeJsonPayload(&country)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if country.Code == "" {
		rest.Error(w, "country code required", 400)
		return
	}
	if country.Name == "" {
		rest.Error(w, "country name required", 400)
		return
	}
	store[country.Code] = &country
	w.WriteJson(&country)
}

func DeleteCountry(w rest.ResponseWriter, r *rest.Request) {
	code := r.PathParam("code")
	delete(store, code)
}

~~~


#### Basic Auth Custom

Demo a custom implementation of Authentication Basic

Demonstrate how to implement a custom AuthBasic middleware, used to protect all endpoints.

This is a very simple version supporting only one user.

The Curl Demo:

        curl -i http://127.0.0.1:8080/countries



~~~ go
/* */
package main

import (
	"encoding/base64"
	"errors"
	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
	"strings"
)

type MyAuthBasicMiddleware struct {
	Realm    string
	UserId   string
	Password string
}

func (mw *MyAuthBasicMiddleware) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	return func(writer rest.ResponseWriter, request *rest.Request) {

		authHeader := request.Header.Get("Authorization")
		if authHeader == "" {
			mw.unauthorized(writer)
			return
		}

		providedUserId, providedPassword, err := mw.decodeBasicAuthHeader(authHeader)

		if err != nil {
			rest.Error(writer, "Invalid authentication", http.StatusBadRequest)
			return
		}

		if !(providedUserId == mw.UserId && providedPassword == mw.Password) {
			mw.unauthorized(writer)
			return
		}

		handler(writer, request)
	}
}

func (mw *MyAuthBasicMiddleware) unauthorized(writer rest.ResponseWriter) {
	writer.Header().Set("WWW-Authenticate", "Basic realm="+mw.Realm)
	rest.Error(writer, "Not Authorized", http.StatusUnauthorized)
}

func (mw *MyAuthBasicMiddleware) decodeBasicAuthHeader(header string) (user string, password string, err error) {

	parts := strings.SplitN(header, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Basic") {
		return "", "", errors.New("Invalid authentication")
	}

	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", errors.New("Invalid base64")
	}

	creds := strings.SplitN(string(decoded), ":", 2)
	if len(creds) != 2 {
		return "", "", errors.New("Invalid authentication")
	}

	return creds[0], creds[1], nil
}

func main() {

	handler := rest.ResourceHandler{
		PreRoutingMiddlewares: []rest.Middleware{
			&MyAuthBasicMiddleware{
				Realm:    "Administration",
				UserId:   "admin",
				Password: "admin",
			},
		},
	}
	handler.SetRoutes(
		rest.Route{"GET", "/countries", GetAllCountries},
	)
	http.ListenAndServe(":8080", &handler)
}

type Country struct {
	Code string
	Name string
}

func GetAllCountries(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(
		[]Country{
			Country{
				Code: "FR",
				Name: "France",
			},
			Country{
				Code: "US",
				Name: "United States",
			},
		},
	)
}

~~~


#### CORS Custom

Demo a custom implementation of CORS

Demonstrate how to implement a custom CORS middleware, used to on all endpoints.

The Curl Demo:

        curl -i http://127.0.0.1:8080/countries



~~~ go
/* */
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
)

type MyCorsMiddleware struct{}

func (mw *MyCorsMiddleware) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	return func(writer rest.ResponseWriter, request *rest.Request) {

		corsInfo := request.GetCorsInfo()

		// Be nice with non CORS requests, continue
		// Alternatively, you may also chose to only allow CORS requests, and return an error.
		if !corsInfo.IsCors {
			// continure, execute the wrapped middleware
			handler(writer, request)
			return
		}

		// Validate the Origin
		// More sophisticated validations can be implemented, regexps, DB lookups, ...
		if corsInfo.Origin != "http://my.other.host" {
			rest.Error(writer, "Invalid Origin", http.StatusForbidden)
			return
		}

		if corsInfo.IsPreflight {
			// check the request methods
			allowedMethods := map[string]bool{
				"GET":  true,
				"POST": true,
				"PUT":  true,
				// don't allow DELETE, for instance
			}
			if !allowedMethods[corsInfo.AccessControlRequestMethod] {
				rest.Error(writer, "Invalid Preflight Request", http.StatusForbidden)
				return
			}
			// check the request headers
			allowedHeaders := map[string]bool{
				"Accept":          true,
				"Content-Type":    true,
				"X-Custom-Header": true,
			}
			for _, requestedHeader := range corsInfo.AccessControlRequestHeaders {
				if !allowedHeaders[requestedHeader] {
					rest.Error(writer, "Invalid Preflight Request", http.StatusForbidden)
					return
				}
			}

			for allowedMethod, _ := range allowedMethods {
				writer.Header().Add("Access-Control-Allow-Methods", allowedMethod)
			}
			for allowedHeader, _ := range allowedHeaders {
				writer.Header().Add("Access-Control-Allow-Headers", allowedHeader)
			}
			writer.Header().Set("Access-Control-Allow-Origin", corsInfo.Origin)
			writer.Header().Set("Access-Control-Allow-Credentials", "true")
			writer.Header().Set("Access-Control-Max-Age", "3600")
			writer.WriteHeader(http.StatusOK)
			return
		} else {
			writer.Header().Set("Access-Control-Expose-Headers", "X-Powered-By")
			writer.Header().Set("Access-Control-Allow-Origin", corsInfo.Origin)
			writer.Header().Set("Access-Control-Allow-Credentials", "true")
			// continure, execute the wrapped middleware
			handler(writer, request)
			return
		}
	}
}

func main() {

	handler := rest.ResourceHandler{
		PreRoutingMiddlewares: []rest.Middleware{
			&MyCorsMiddleware{},
		},
	}
	handler.SetRoutes(
		rest.Route{"GET", "/countries", GetAllCountries},
	)
	http.ListenAndServe(":8080", &handler)
}

type Country struct {
	Code string
	Name string
}

func GetAllCountries(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(
		[]Country{
			Country{
				Code: "FR",
				Name: "France",
			},
			Country{
				Code: "US",
				Name: "United States",
			},
		},
	)
}

~~~



## External Documentation

- [Online Documentation (godoc.org)](http://godoc.org/github.com/ant0ine/go-json-rest)

Old v1 blog posts:

- [(Blog Post) Introducing Go-Json-Rest] (http://blog.ant0ine.com/typepad/2013/04/introducing-go-json-rest.html)
- [(Blog Post) Better URL Routing ?] (http://blog.ant0ine.com/typepad/2013/02/better-url-routing-golang-1.html)


## Options

Things to enable in production:
- Gzip compression (default: disabled)
- Custom Logger (default: Go default)

Things to enable in development:
- Json indentation (default: enabled)
- Relaxed ContentType (default: disabled)
- Error stack trace in the response body (default: disabled)


## Migration guide from v1 to v2

**Go-Json-Rest** follows [Semver](http://semver.org/) and a few breaking changes have been introduced with the v2.


#### The import path has changed to `github.com/ant0ine/go-json-rest/rest`

This is more conform to Go style, and makes [goimports](https://godoc.org/code.google.com/p/go.tools/cmd/goimports) work.

This:
~~~ go
import (
        "github.com/ant0ine/go-json-rest"
)
~~~
has to be changed to this:
~~~ go
import (
        "github.com/ant0ine/go-json-rest/rest"
)
~~~


#### rest.ResponseWriter is now an interface

This change allows the `ResponseWriter` to be wrapped, like the one of the `net/http` package. Middlewares like Gzip used this to encode the payload (see gzip.go).

This:
~~~ go
func (w *rest.ResponseWriter, req *rest.Request) {
        ...
}
~~~
has to be changed to this:
~~~ go
func (w rest.ResponseWriter, req *rest.Request) {
        ...
}
~~~


####  The notion of Middleware is now formally defined

A middleware is an object satisfying this interface:
~~~ go
type Middleware interface {
	MiddlewareFunc(handler HandlerFunc) HandlerFunc
}
~~~

Code using PreRoutingMiddleware will have to be adapted to provide a list of Middleware objects.
See the [Basic Auth example](https://github.com/ant0ine/go-json-rest-examples/blob/v2-alpha/auth-basic/main.go).


#### Flush(), CloseNotify() and Write() are not directly exposed anymore

A type assertion of the corresponding interface is necessary.

This:
~~~ go
writer.Flush()
~~~
has to be changed to this:
~~~ go
writer.(http.Flusher).Flush()
~~~


#### The /.status endpoint is not created automatically anymore

The route has to be manually defined.
See the [Status example](https://github.com/ant0ine/go-json-rest-examples/blob/v2-alpha/status/main.go).


#### Request utility methods have changed

Overall, they provide the same features, but with two methods instead of three, better names, and without the confusing `UriForWithParams`.

`func (r *Request) UriBase() url.URL` is now `func (r *Request) BaseUrl() *url.URL`, Note the pointer as the returned value.

`func (r *Request) UriForWithParams(path string, parameters map[string][]string) url.URL` is now `func (r *Request) UrlFor(path string, queryParams map[string][]string) *url.URL` and `func (r *Request) UriFor(path string) url.URL` has be removed.

## Thanks

- [Franck Cuny](https://github.com/franckcuny)
- [Yann Kerhervé](https://github.com/yannk)
- [Ask Bjørn Hansen](https://github.com/abh)


Copyright (c) 2013-2014 Antoine Imbert

[MIT License](https://github.com/ant0ine/go-json-rest/blob/master/LICENSE)

[![Analytics](https://ga-beacon.appspot.com/UA-309210-4/go-json-rest/v2-alpha/readme)](https://github.com/igrigorik/ga-beacon)


