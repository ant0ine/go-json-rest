
# Go-Json-Rest

*A quick and easy way to setup a RESTful JSON API*

[![Build Status](https://travis-ci.org/ant0ine/go-json-rest.png?branch=master)](https://travis-ci.org/ant0ine/go-json-rest) [![GoDoc](https://godoc.org/github.com/ant0ine/go-json-rest?status.png)](https://godoc.org/github.com/ant0ine/go-json-rest/rest)


**Go-Json-Rest** is a thin layer on top of `net/http` that helps building RESTful JSON APIs easily. It provides fast URL routing using a Trie based implementation, helpers to deal with JSON requests and responses, and middlewares for additional functionalities like CORS, Auth, Gzip ...


## What's new in version 2

* Middlewares, the notion of middleware is now formally defined. They can be setup as global pre-routing Middlewares wrapping all the endpoints, or on a per endpoint basis.
In fact the internal code of **go-json-rest** is itself implemented with Middlewares, they are just hidden behind configuration boolean flags to make these very common options even easier to use.

* A new ResponseWriter. This is now an interface, and allows Middlewares to wrap the writer. The provided writer implements, in addition of *rest.ResponseWriter*, *http.Flusher*, *http.CloseNotifier*, and *http.ResponseWriter*. A lot more Go-ish, and very similar to `net/http`.

* The AuthBasic and CORS Middlewares have been added. More to come in the future.

* Faster, more tasks are performed at init time, and less for each request.

* New documentation, with more examples.

* A lot of other small improvements, See the [Migration guide to v2](#migration-guide-from-v1-to-v2)


## Table of content

- [Features](#features)
- [Install](#install)
- [Vendoring](#vendoring)
- [Examples](#examples)
  - [Basics](#basics)
	  - [Hello World!](#hello-world)
	  - [Countries](#countries)
	  - [Users](#users)
	  - [Lookup](#lookup)
  - [Applications](#applications)
	  - [API and static files](#api-and-static-files)
	  - [GORM](#gorm)
	  - [CORS](#cors)
	  - [Basic Auth](#basic-auth)
	  - [Status](#status)
	  - [Status Auth](#status-auth)
  - [Advanced](#advanced)
	  - [Streaming](#streaming)
	  - [Non JSON payload](#non-json-payload)
	  - [API Versioning](#api-versioning)
	  - [Statsd](#statsd)
	  - [SPDY](#spdy)
	  - [Google App Engine](#gae)
	  - [Basic Auth Custom](#basic-auth-custom)
	  - [CORS Custom](#cors-custom)
- [External Documentation](#external-documentation)
- [Options](#options)
- [Migration guide from v1 to v2](#migration-guide-from-v1-to-v2)
- [Thanks](#thanks)


## Features

- Many examples.
- Fast and scalable URL routing. It implements the classic route description syntax using a scalable trie data structure.
- Use Middlewares in order to implement and extend the functionalities. (Logging, Gzip, CORS, Auth, ...)
- Implemented as a `net/http` Handler. This standard interface allows combinations with other Handlers.
- Test package to help writing tests for your API.
- Monitoring statistics inspired by Memcached.


## Install

This package is "go-gettable", just do:

    go get github.com/ant0ine/go-json-rest/rest


## Vendoring

The recommended way of using this library in your project is to use the **"vendoring"** method,
where this library code is copied in your repository at a specific revision.
[This page](http://nathany.com/go-packages/) is a good summary of package management in Go.


## Examples

All the following examples can be found in dedicated examples repository: https://github.com/ant0ine/go-json-rest-examples

### Basics

First examples to try, as an introduction to go-json-rest.

#### Hello World!

Tradition!

The curl demo:
``` sh
curl -i http://127.0.0.1:8080/message
```


Go code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
)

type Message struct {
	Body string
}

func main() {
	handler := rest.ResourceHandler{}
	err := handler.SetRoutes(
		&rest.Route{"GET", "/message", func(w rest.ResponseWriter, req *rest.Request) {
			w.WriteJson(&Message{
				Body: "Hello World!",
			})
		}},
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":8080", &handler))
}

```

#### Countries

Demonstrate simple POST GET and DELETE operations

The curl demo:
```
curl -i -d '{"Code":"FR","Name":"France"}' http://127.0.0.1:8080/countries
curl -i -d '{"Code":"US","Name":"United States"}' http://127.0.0.1:8080/countries
curl -i http://127.0.0.1:8080/countries/FR
curl -i http://127.0.0.1:8080/countries/US
curl -i http://127.0.0.1:8080/countries
curl -i -X DELETE http://127.0.0.1:8080/countries/FR
curl -i http://127.0.0.1:8080/countries
curl -i -X DELETE http://127.0.0.1:8080/countries/US
curl -i http://127.0.0.1:8080/countries
```

Go code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
	"sync"
)

func main() {

	handler := rest.ResourceHandler{
		EnableRelaxedContentType: true,
	}
	err := handler.SetRoutes(
		&rest.Route{"GET", "/countries", GetAllCountries},
		&rest.Route{"POST", "/countries", PostCountry},
		&rest.Route{"GET", "/countries/:code", GetCountry},
		&rest.Route{"DELETE", "/countries/:code", DeleteCountry},
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":8080", &handler))
}

type Country struct {
	Code string
	Name string
}

var store = map[string]*Country{}

var lock = sync.RWMutex{}

func GetCountry(w rest.ResponseWriter, r *rest.Request) {
	code := r.PathParam("code")

	lock.RLock()
	var country *Country
	if store[code] != nil {
		country = &Country{}
		*country = *store[code]
	}
	lock.RUnlock()

	if country == nil {
		rest.NotFound(w, r)
		return
	}
	w.WriteJson(country)
}

func GetAllCountries(w rest.ResponseWriter, r *rest.Request) {
	lock.RLock()
	countries := make([]Country, len(store))
	i := 0
	for _, country := range store {
		countries[i] = *country
		i++
	}
	lock.RUnlock()
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
	lock.Lock()
	store[country.Code] = &country
	lock.Unlock()
	w.WriteJson(&country)
}

func DeleteCountry(w rest.ResponseWriter, r *rest.Request) {
	code := r.PathParam("code")
	lock.Lock()
	delete(store, code)
	lock.Unlock()
	w.WriteHeader(http.StatusOK)
}

```

#### Users

Demonstrate how to use `rest.RouteObjectMethod`

`rest.RouteObjectMethod` helps create a Route that points to an object method instead of just a function.

The curl demo:
```
curl -i -d '{"Name":"Antoine"}' http://127.0.0.1:8080/users
curl -i http://127.0.0.1:8080/users/0
curl -i -X PUT -d '{"Name":"Antoine Imbert"}' http://127.0.0.1:8080/users/0
curl -i -X DELETE http://127.0.0.1:8080/users/0
curl -i http://127.0.0.1:8080/users
```

Go code:
``` go
package main

import (
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
	"sync"
)

func main() {

	users := Users{
		Store: map[string]*User{},
	}

	handler := rest.ResourceHandler{
		EnableRelaxedContentType: true,
	}
	err := handler.SetRoutes(
		rest.RouteObjectMethod("GET", "/users", &users, "GetAllUsers"),
		rest.RouteObjectMethod("POST", "/users", &users, "PostUser"),
		rest.RouteObjectMethod("GET", "/users/:id", &users, "GetUser"),
		rest.RouteObjectMethod("PUT", "/users/:id", &users, "PutUser"),
		rest.RouteObjectMethod("DELETE", "/users/:id", &users, "DeleteUser"),
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":8080", &handler))
}

type User struct {
	Id   string
	Name string
}

type Users struct {
	sync.RWMutex
	Store map[string]*User
}

func (u *Users) GetAllUsers(w rest.ResponseWriter, r *rest.Request) {
	u.RLock()
	users := make([]User, len(u.Store))
	i := 0
	for _, user := range u.Store {
		users[i] = *user
		i++
	}
	u.RUnlock()
	w.WriteJson(&users)
}

func (u *Users) GetUser(w rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")
	u.RLock()
	var user *User
	if u.Store[id] != nil {
		user = &User{}
		*user = *u.Store[id]
	}
	u.RUnlock()
	if user == nil {
		rest.NotFound(w, r)
		return
	}
	w.WriteJson(user)
}

func (u *Users) PostUser(w rest.ResponseWriter, r *rest.Request) {
	user := User{}
	err := r.DecodeJsonPayload(&user)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	u.Lock()
	id := fmt.Sprintf("%d", len(u.Store)) // stupid
	user.Id = id
	u.Store[id] = &user
	u.Unlock()
	w.WriteJson(&user)
}

func (u *Users) PutUser(w rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")
	u.Lock()
	if u.Store[id] == nil {
		rest.NotFound(w, r)
		u.Unlock()
		return
	}
	user := User{}
	err := r.DecodeJsonPayload(&user)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		u.Unlock()
		return
	}
	user.Id = id
	u.Store[id] = &user
	u.Unlock()
	w.WriteJson(&user)
}

func (u *Users) DeleteUser(w rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")
	u.Lock()
	delete(u.Store, id)
	u.Unlock()
	w.WriteHeader(http.StatusOK)
}

```

#### Lookup

Demonstrate how to use the relaxed placeholder (notation #paramName).
This placeholder matches everything until the first `/`, including `.`

The curl demo:
```
curl -i http://127.0.0.1:8080/lookup/google.com
curl -i http://127.0.0.1:8080/lookup/notadomain
```

Go code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net"
	"net/http"
)

type Message struct {
	Body string
}

func main() {
	handler := rest.ResourceHandler{}
	err := handler.SetRoutes(
		&rest.Route{"GET", "/lookup/#host", func(w rest.ResponseWriter, req *rest.Request) {
			ip, err := net.LookupIP(req.PathParam("host"))
			if err != nil {
				rest.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteJson(&ip)
		}},
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":8080", &handler))
}

```


### Applications

Common use cases, found in many applications.

#### API and static files

Combine Go-Json-Rest with other handlers.

`rest.ResourceHandler` is a valid `http.Handler`, and can be combined with other handlers.
In this example the ResourceHandler is used under the `/api/` prefix, while a FileServer is instantiated under the `/static/` prefix.

The curl demo:
```
curl -i http://127.0.0.1:8080/api/message
curl -i http://127.0.0.1:8080/static/main.go
```

Go code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
)

type Message struct {
	Body string
}

func main() {
	handler := rest.ResourceHandler{}
	err := handler.SetRoutes(
		&rest.Route{"GET", "/message", func(w rest.ResponseWriter, req *rest.Request) {
			w.WriteJson(&Message{
				Body: "Hello World!",
			})
		}},
	)
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/api/", http.StripPrefix("/api", &handler))

	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("."))))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

```

#### GORM

Demonstrate basic CRUD operation using a store based on MySQL and GORM

[GORM](https://github.com/jinzhu/gorm) is simple ORM library for Go.
In this example the same struct is used both as the GORM model and as the JSON model.

The curl demo:
```
curl -i -d '{"Message":"this is a test"}' http://127.0.0.1:8080/reminders
curl -i http://127.0.0.1:8080/reminders/1
curl -i http://127.0.0.1:8080/reminders
curl -i -X PUT -d '{"Message":"is updated"}' http://127.0.0.1:8080/reminders/1
curl -i -X DELETE http://127.0.0.1:8080/reminders/1
```

Go code:
``` go
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
	err := handler.SetRoutes(
		rest.RouteObjectMethod("GET", "/reminders", &api, "GetAllReminders"),
		rest.RouteObjectMethod("POST", "/reminders", &api, "PostReminder"),
		rest.RouteObjectMethod("GET", "/reminders/:id", &api, "GetReminder"),
		rest.RouteObjectMethod("PUT", "/reminders/:id", &api, "PutReminder"),
		rest.RouteObjectMethod("DELETE", "/reminders/:id", &api, "DeleteReminder"),
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":8080", &handler))
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
	w.WriteHeader(http.StatusOK)
}

```

#### CORS

Demonstrate how to setup CorsMiddleware around all the API endpoints.

The curl demo:
```
curl -i http://127.0.0.1:8080/countries
```


Go code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
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
				AllowedMethods: []string{"GET", "POST", "PUT"},
				AllowedHeaders: []string{
					"Accept", "Content-Type", "X-Custom-Header", "Origin"},
				AccessControlAllowCredentials: true,
				AccessControlMaxAge:           3600,
			},
		},
	}
	err := handler.SetRoutes(
		&rest.Route{"GET", "/countries", GetAllCountries},
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":8080", &handler))
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

```

#### Basic Auth

Demonstrate how to setup AuthBasicMiddleware as a pre-routing middleware.

The curl demo:
```
curl -i http://127.0.0.1:8080/countries
curl -i -u admin:admin http://127.0.0.1:8080/countries
```

Go code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
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
	err := handler.SetRoutes(
		&rest.Route{"GET", "/countries", GetAllCountries},
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":8080", &handler))
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

```

#### Status

Demonstrate how to setup a `/.status` endpoint

Inspired by memcached "stats", this optional feature can be enabled to help monitoring the service.
This example shows how to enable the stats, and how to setup the `/.status` route.


The curl demo:
```
curl -i http://127.0.0.1:8080/.status
curl -i http://127.0.0.1:8080/.status
...
```

Output example:
```
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
```

Go code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
)

func main() {
	handler := rest.ResourceHandler{
		EnableStatusService: true,
	}
	err := handler.SetRoutes(
		&rest.Route{"GET", "/.status",
			func(w rest.ResponseWriter, r *rest.Request) {
				w.WriteJson(handler.GetStatus())
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":8080", &handler))
}

```

#### Status Auth

Demonstrate how to setup a /.status endpoint protected with basic authentication.

This is a good use case of middleware applied to only one API endpoint.

The Curl Demo:
```
curl -i http://127.0.0.1:8080/countries
curl -i http://127.0.0.1:8080/.status
curl -i -u admin:admin http://127.0.0.1:8080/.status
...
```

Go code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
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
	err := handler.SetRoutes(
		&rest.Route{"GET", "/countries", GetAllCountries},
		&rest.Route{"GET", "/.status",
			auth.MiddlewareFunc(
				func(w rest.ResponseWriter, r *rest.Request) {
					w.WriteJson(handler.GetStatus())
				},
			),
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":8080", &handler))
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

```


### Advanced

Less common use cases.

#### Streaming

Demonstrate a streaming REST API, where the data is "flushed" to the client ASAP.

The stream format is a Line Delimited JSON.

The curl demo:
```
curl -i http://127.0.0.1:8080/stream
```

Output:
```
HTTP/1.1 200 OK
Content-Type: application/json
Date: Sun, 16 Feb 2014 00:39:19 GMT
Transfer-Encoding: chunked

{"Name":"thing #1"}
{"Name":"thing #2"}
{"Name":"thing #3"}
```

Go code:
``` go
package main

import (
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
	"time"
)

func main() {

	handler := rest.ResourceHandler{
		EnableRelaxedContentType: true,
		DisableJsonIndent:        true,
	}
	err := handler.SetRoutes(
		&rest.Route{"GET", "/stream", StreamThings},
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":8080", &handler))
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

```

#### Non JSON payload

Exceptional use of non JSON payloads.

The ResponseWriter implementation provided by go-json-rest is designed
to build JSON responses. In order to serve different kind of content,
it is recommended to either:
a) use another server and configure CORS
   (see the cors/ example)
b) combine the rest.ResourceHandler with another http.Handler
   (see api-and-static/ example)

That been said, exceptionally, it can be convenient to return a
different content type on a JSON endpoint. In this case, setting the
Content-Type and using the type assertion to access the Write method
is enough. As shown in this example.

The curl demo:
```
curl -i http://127.0.0.1:8080/message.txt
```

Go code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
)

func main() {
	handler := rest.ResourceHandler{}
	err := handler.SetRoutes(
		&rest.Route{"GET", "/message.txt", func(w rest.ResponseWriter, req *rest.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.(http.ResponseWriter).Write([]byte("Hello World!"))
		}},
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":8080", &handler))
}

```

#### API Versioning

First, API versioning is not easy and you may want to favor a mechanism that uses only backward compatible changes and deprecation cycles.

That been said, here is an example of API versioning using [Semver](http://semver.org/)

It defines a middleware that parses the version, checks a min and a max, and makes it available in the `request.Env`.

(TODO, there is an obvious need for PostRoutingMiddlewares here.)

The curl demo:
``` sh
curl -i http://127.0.0.1:8080/api/1.0.0/message
curl -i http://127.0.0.1:8080/api/2.0.0/message
curl -i http://127.0.0.1:8080/api/2.0.1/message
curl -i http://127.0.0.1:8080/api/0.0.1/message
curl -i http://127.0.0.1:8080/api/4.0.1/message

```


Go code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/coreos/go-semver/semver"
	"log"
	"net/http"
)

type SemVerMiddleware struct {
	MinVersion string
	MaxVersion string
}

func (mw *SemVerMiddleware) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {

	minVersion, err := semver.NewVersion(mw.MinVersion)
	if err != nil {
		panic(err)
	}

	maxVersion, err := semver.NewVersion(mw.MaxVersion)
	if err != nil {
		panic(err)
	}

	return func(writer rest.ResponseWriter, request *rest.Request) {

		version, err := semver.NewVersion(request.PathParam("version"))
		if err != nil {
			rest.Error(writer, "Invalid version: "+err.Error(), http.StatusBadRequest)
			return
		}

		if version.LessThan(*minVersion) {
			rest.Error(writer, "Min supported version is "+minVersion.String(), http.StatusBadRequest)
			return
		}

		if maxVersion.LessThan(*version) {
			rest.Error(writer, "Max supported version is "+maxVersion.String(), http.StatusBadRequest)
			return
		}

		request.Env["VERSION"] = version
		handler(writer, request)
	}
}

type Message struct {
	Body string
}

func main() {
	handler := rest.ResourceHandler{}
	svmw := SemVerMiddleware{
		MinVersion: "1.0.0",
		MaxVersion: "3.0.0",
	}
	err := handler.SetRoutes(
		&rest.Route{"GET", "/#version/message", svmw.MiddlewareFunc(
			func(w rest.ResponseWriter, req *rest.Request) {
				version := req.Env["VERSION"].(*semver.Version)
				if version.Major == 2 {
                                        // http://en.wikipedia.org/wiki/Second-system_effect
					w.WriteJson(&Message{"Hello broken World!"})
				} else {
					w.WriteJson(&Message{"Hello World!"})
				}
			},
		)},
	)
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/api/", http.StripPrefix("/api", &handler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

```

#### Statsd

Demonstrate how to use OuterMiddlewares to do additional logging and reporting.

Here `request.Env["STATUS_CODE"]` and `request.Env["ELAPSED_TIME"]` that are available to outer middlewares are used with the [g2s](https://github.com/peterbourgon/g2s) statsd client to send these metrics to statsd.

The curl demo:
``` sh
# start statsd server
# monitor network
ngrep -d any port 8125

curl -i http://127.0.0.1:8080/message
curl -i http://127.0.0.1:8080/doesnotexist

```

Go code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/peterbourgon/g2s"
	"log"
	"net/http"
	"strconv"
	"time"
)

type StatsdMiddleware struct {
	IpPort string
	Prefix string
}

func (mw *StatsdMiddleware) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {

	statsd, err := g2s.Dial("udp", mw.IpPort)
	if err != nil {
		panic(err)
	}

	keyBase := ""
	if mw.Prefix != "" {
		keyBase += mw.Prefix + "."
	}
	keyBase += "response."

	return func(writer rest.ResponseWriter, request *rest.Request) {

		handler(writer, request)

		statusCode := request.Env["STATUS_CODE"].(int)
		statsd.Counter(1.0, keyBase+"status_code."+strconv.Itoa(statusCode), 1)

		elapsedTime := request.Env["ELAPSED_TIME"].(*time.Duration)
		statsd.Timing(1.0, keyBase+"elapsed_time", *elapsedTime)
	}
}

type Message struct {
	Body string
}

func main() {
	handler := rest.ResourceHandler{
		OuterMiddlewares: []rest.Middleware{
			&StatsdMiddleware{
				IpPort: "localhost:8125",
			},
		},
	}
	err := handler.SetRoutes(
		&rest.Route{"GET", "/message", func(w rest.ResponseWriter, req *rest.Request) {

			// take more than 1ms so statsd can report it
			time.Sleep(100 * time.Millisecond)

			w.WriteJson(&Message{
				Body: "Hello World!",
			})
		}},
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":8080", &handler))
}

```

#### SPDY

Demonstrate how to use SPDY with https://github.com/shykes/spdy-go

For a command line client, install spdycat from:
https://github.com/tatsuhiro-t/spdylay

The spdycat demo:
```
spdycat -v --no-tls -2 http://localhost:8080/users/0
```

Go code:
``` go
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
	err := handler.SetRoutes(
		&rest.Route{"GET", "/users/:id", GetUser},
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(spdy.ListenAndServeTCP(":8080", &handler))
}

```

#### GAE

Demonstrate a simple Google App Engine app

Here are my steps to make it work with the GAE SDK.
(Probably not the best ones)

Assuming that go-json-rest is installed using "go get"
and that the GAE SDK is also installed.

Setup:
 * copy this examples/gae/ dir outside of the go-json-rest/ tree
 * cd gae/
 * mkdir -p github.com/ant0ine
 * cp -r $GOPATH/src/github.com/ant0ine/go-json-rest github.com/ant0ine/go-json-rest
 * rm -rf github.com/ant0ine/go-json-rest/examples/
 * path/to/google_appengine/dev_appserver.py .

The curl demo:
```
curl -i http://127.0.0.1:8080/message
```

Go code:
``` go
package gaehelloworld

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
)

type Message struct {
	Body string
}

func init() {
	handler := rest.ResourceHandler{}
	err := handler.SetRoutes(
		&rest.Route{"GET", "/message", func(w rest.ResponseWriter, req *rest.Request) {
			w.WriteJson(&Message{
				Body: "Hello World!",
			})
		}},
	)
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", &handler)
}

```

#### Basic Auth Custom

Demonstrate how to implement a custom AuthBasic middleware, used to protect all endpoints.

This is a very simple version supporting only one user.

The curl demo:
```
curl -i http://127.0.0.1:8080/countries
```

Go code:
``` go
package main

import (
	"encoding/base64"
	"errors"
	"github.com/ant0ine/go-json-rest/rest"
	"log"
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
	err := handler.SetRoutes(
		&rest.Route{"GET", "/countries", GetAllCountries},
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":8080", &handler))
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

```

#### CORS Custom

Demonstrate how to implement a custom CORS middleware, used to on all endpoints.

The curl demo:
```
curl -i http://127.0.0.1:8080/countries
```

Go code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
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
	err := handler.SetRoutes(
		&rest.Route{"GET", "/countries", GetAllCountries},
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":8080", &handler))
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

```



## External Documentation

- [Online Documentation (godoc.org)](http://godoc.org/github.com/ant0ine/go-json-rest/rest)

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
``` go
import (
        "github.com/ant0ine/go-json-rest"
)
```
has to be changed to this:
``` go
import (
        "github.com/ant0ine/go-json-rest/rest"
)
```


#### rest.ResponseWriter is now an interface

This change allows the `ResponseWriter` to be wrapped, like the one of the `net/http` package.
This is much more powerful, and allows the creation of Middlewares that wrap the writer. The gzip option, for instance, uses this to encode the payload (see gzip.go).

This:
``` go
func (w *rest.ResponseWriter, req *rest.Request) {
        ...
}
```
has to be changed to this:
``` go
func (w rest.ResponseWriter, req *rest.Request) {
        ...
}
```


#### SetRoutes now takes pointers to Route

Instead of copying Route structures everywhere, pointers are now used. This is more elegant, more efficient, and will allow more sophisticated Route manipulations in the future (like reverse route resolution).

This:
``` go
handler.SetRoutes(
		rest.Route{
		      // ...
		},
)
```
has to be changed to this:
``` go
handler.SetRoutes(
		&rest.Route{
		      // ...
		},
)
```


####  The notion of Middleware is now formally defined

A middleware is an object satisfying this interface:
``` go
type Middleware interface {
	MiddlewareFunc(handler HandlerFunc) HandlerFunc
}
```

Code using PreRoutingMiddleware will have to be adapted to provide a list of Middleware objects.
See the [Basic Auth example](https://github.com/ant0ine/go-json-rest-examples/blob/master/auth-basic/main.go).


#### Flush(), CloseNotify() and Write() are not directly exposed anymore

They used to be public methods of the ResponseWriter. The implementation is still there but a type assertion of the corresponding interface is now necessary.
Regarding these features, a rest.ResponseWriter now behaves exactly as the http.ResponseWriter implementation provided by net/http.

This:
``` go
writer.Flush()
```
has to be changed to this:
``` go
writer.(http.Flusher).Flush()
```


#### The /.status endpoint is not created automatically anymore

The route has to be manually defined.
See the [Status example](https://github.com/ant0ine/go-json-rest-examples/blob/master/status/main.go).
This is more flexible (the route is customizable), and allows combination with Middlewarres.
See for instance how to [protect this status endpoint with the AuthBasic middleware](https://github.com/ant0ine/go-json-rest-examples/blob/master/status-auth/main.go).


#### Request utility methods have changed

Overall, they provide the same features, but with two methods instead of three, better names, and without the confusing `UriForWithParams`.

- `func (r *Request) UriBase() url.URL` is now `func (r *Request) BaseUrl() *url.URL`, Note the pointer as the returned value.

- `func (r *Request) UriForWithParams(path string, parameters map[string][]string) url.URL` is now `func (r *Request) UrlFor(path string, queryParams map[string][]string) *url.URL`.

- `func (r *Request) UriFor(path string) url.URL` has be removed.


## Thanks

- [Franck Cuny](https://github.com/franckcuny)
- [Yann Kerhervé](https://github.com/yannk)
- [Ask Bjørn Hansen](https://github.com/abh)
- [Paul Lam](https://github.com/Quantisan)


Copyright (c) 2013-2014 Antoine Imbert

[MIT License](https://github.com/ant0ine/go-json-rest/blob/master/LICENSE)

[![Analytics](https://ga-beacon.appspot.com/UA-309210-4/go-json-rest/master/readme)](https://github.com/igrigorik/ga-beacon)

