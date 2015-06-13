
# Go-Json-Rest

*A quick and easy way to setup a RESTful JSON API*

[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/ant0ine/go-json-rest/rest) [![license](https://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/ant0ine/go-json-rest/master/LICENSE) [![build](https://img.shields.io/travis/ant0ine/go-json-rest.svg?style=flat)](https://travis-ci.org/ant0ine/go-json-rest)


**Go-Json-Rest** is a thin layer on top of `net/http` that helps building RESTful JSON APIs easily. It provides fast and scalable request routing using a Trie based implementation, helpers to deal with JSON requests and responses, and middlewares for functionalities like CORS, Auth, Gzip, Status ...


## Table of content

- [Features](#features)
- [Install](#install)
- [Vendoring](#vendoring)
- [Middlewares](#middlewares)
- [Examples](#examples)
  - [Basics](#basics)
	  - [Hello World!](#hello-world)
	  - [Lookup](#lookup)
	  - [Countries](#countries)
	  - [Users](#users)
  - [Applications](#applications)
	  - [API and static files](#api-and-static-files)
	  - [GORM](#gorm)
	  - [CORS](#cors)
	  - [JSONP](#jsonp)
	  - [Basic Auth](#basic-auth)
	  - [Status](#status)
	  - [Status Auth](#status-auth)
  - [Advanced](#advanced)
	  - [JWT](#jwt)
	  - [Streaming](#streaming)
	  - [Non JSON payload](#non-json-payload)
	  - [API Versioning](#api-versioning)
	  - [Statsd](#statsd)
	  - [NewRelic](#newrelic)
	  - [Graceful Shutdown](#graceful-shutdown)
	  - [SPDY](#spdy)
	  - [Google App Engine](#gae)
	  - [Websocket](#websocket)
- [External Documentation](#external-documentation)
- [Version 3 release notes](#version-3-release-notes)
- [Migration guide from v2 to v3](#migration-guide-from-v2-to-v3)
- [Version 2 release notes](#version-2-release-notes)
- [Migration guide from v1 to v2](#migration-guide-from-v1-to-v2)
- [Thanks](#thanks)


## Features

- Many examples.
- Fast and scalable URL routing. It implements the classic route description syntax using a Trie data structure.
- Architecture based on a router(App) sitting on top of a stack of Middlewares.
- The Middlewares implement functionalities like Logging, Gzip, CORS, Auth, Status, ...
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


## Middlewares

Core Middlewares:

| Name | Description |
|------|-------------|
| **AccessLogApache** | Access log inspired by Apache mod_log_config |
| **AccessLogJson** | Access log with records as JSON |
| **AuthBasic** | Basic HTTP auth |
| **ContentTypeChecker** | Verify the request content type |
| **Cors** | CORS server side implementation |
| **Gzip** | Compress the responses |
| **If** | Conditionally execute a Middleware at runtime |
| **JsonIndent** | Easy to read JSON |
| **Jsonp** | Response as JSONP |
| **PoweredBy** | Manage the X-Powered-By response header |
| **Recorder** | Record the status code and content length in the Env |
| **Status** | Memecached inspired stats about the requests |
| **Timer** | Keep track of the elapsed time in the Env |

Third Party Middlewares:

| Name | Description |
|------|-------------|
| **[Statsd](https://github.com/ant0ine/go-json-rest-middleware-statsd)** | Send stats to a statsd server |
| **[JWT](https://github.com/StephanDollberg/go-json-rest-middleware-jwt)** | Provides authentication via Json Web Tokens |
| **[AuthToken](https://github.com/grayj/go-json-rest-middleware-tokenauth)** | Provides a Token Auth implementation |

*If you have a Go-Json-Rest compatible middleware, feel free to submit a PR to add it in this list, and in the examples.*


## Examples

All the following examples can be found in dedicated examples repository: https://github.com/ant0ine/go-json-rest-examples

### Basics

First examples to try, as an introduction to go-json-rest.

#### Hello World!

Tradition!

curl demo:
``` sh
curl -i http://127.0.0.1:8080/
```


code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
)

func main() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.SetApp(rest.AppSimple(func(w rest.ResponseWriter, r *rest.Request) {
		w.WriteJson(map[string]string{"Body": "Hello World!"})
	}))
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

```

#### Lookup

Demonstrate how to use the relaxed placeholder (notation `#paramName`).
This placeholder matches everything until the first `/`, including `.`

curl demo:
```
curl -i http://127.0.0.1:8080/lookup/google.com
curl -i http://127.0.0.1:8080/lookup/notadomain
```

code:
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
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/lookup/#host", func(w rest.ResponseWriter, req *rest.Request) {
			ip, err := net.LookupIP(req.PathParam("host"))
			if err != nil {
				rest.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteJson(&ip)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

```

#### Countries

Demonstrate simple POST GET and DELETE operations

curl demo:
```
curl -i -H 'Content-Type: application/json' \
    -d '{"Code":"FR","Name":"France"}' http://127.0.0.1:8080/countries
curl -i -H 'Content-Type: application/json' \
    -d '{"Code":"US","Name":"United States"}' http://127.0.0.1:8080/countries
curl -i http://127.0.0.1:8080/countries/FR
curl -i http://127.0.0.1:8080/countries/US
curl -i http://127.0.0.1:8080/countries
curl -i -X DELETE http://127.0.0.1:8080/countries/FR
curl -i http://127.0.0.1:8080/countries
curl -i -X DELETE http://127.0.0.1:8080/countries/US
curl -i http://127.0.0.1:8080/countries
```

code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
	"sync"
)

func main() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/countries", GetAllCountries),
		rest.Post("/countries", PostCountry),
		rest.Get("/countries/:code", GetCountry),
		rest.Delete("/countries/:code", DeleteCountry),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
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

Demonstrate how to use Method Values.

Method Values have been [introduced in Go 1.1](https://golang.org/doc/go1.1#method_values).

This shows how to map a Route to a method of an instantiated object (i.e: receiver of the method)

curl demo:
```
curl -i -H 'Content-Type: application/json' \
    -d '{"Name":"Antoine"}' http://127.0.0.1:8080/users
curl -i http://127.0.0.1:8080/users/0
curl -i -X PUT -H 'Content-Type: application/json' \
    -d '{"Name":"Antoine Imbert"}' http://127.0.0.1:8080/users/0
curl -i -X DELETE http://127.0.0.1:8080/users/0
curl -i http://127.0.0.1:8080/users
```

code:
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

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/users", users.GetAllUsers),
		rest.Post("/users", users.PostUser),
		rest.Get("/users/:id", users.GetUser),
		rest.Put("/users/:id", users.PutUser),
		rest.Delete("/users/:id", users.DeleteUser),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
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


### Applications

Common use cases, found in many applications.

#### API and static files

Combine Go-Json-Rest with other handlers.

`api.MakeHandler()` is a valid `http.Handler`, and can be combined with other handlers.
In this example the api handler is used under the `/api/` prefix, while a FileServer is instantiated under the `/static/` prefix.

curl demo:
```
curl -i http://127.0.0.1:8080/api/message
curl -i http://127.0.0.1:8080/static/main.go
```

code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
)

func main() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)

	router, err := rest.MakeRouter(
		rest.Get("/message", func(w rest.ResponseWriter, req *rest.Request) {
			w.WriteJson(map[string]string{"Body": "Hello World!"})
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)

	http.Handle("/api/", http.StripPrefix("/api", api.MakeHandler()))

	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("."))))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

```

#### GORM

Demonstrate basic CRUD operation using a store based on MySQL and GORM

[GORM](https://github.com/jinzhu/gorm) is simple ORM library for Go.
In this example the same struct is used both as the GORM model and as the JSON model.

curl demo:
```
curl -i -H 'Content-Type: application/json' \
    -d '{"Message":"this is a test"}' http://127.0.0.1:8080/reminders
curl -i http://127.0.0.1:8080/reminders/1
curl -i http://127.0.0.1:8080/reminders
curl -i -X PUT -H 'Content-Type: application/json' \
    -d '{"Message":"is updated"}' http://127.0.0.1:8080/reminders/1
curl -i -X DELETE http://127.0.0.1:8080/reminders/1
```

code:
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

	i := Impl{}
	i.InitDB()
	i.InitSchema()

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/reminders", i.GetAllReminders),
		rest.Post("/reminders", i.PostReminder),
		rest.Get("/reminders/:id", i.GetReminder),
		rest.Put("/reminders/:id", i.PutReminder),
		rest.Delete("/reminders/:id", i.DeleteReminder),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

type Reminder struct {
	Id        int64     `json:"id"`
	Message   string    `sql:"size:1024" json:"message"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt time.Time `json:"-"`
}

type Impl struct {
	DB gorm.DB
}

func (i *Impl) InitDB() {
	var err error
	i.DB, err = gorm.Open("mysql", "gorm:gorm@/gorm?charset=utf8&parseTime=True")
	if err != nil {
		log.Fatalf("Got error when connect database, the error is '%v'", err)
	}
	i.DB.LogMode(true)
}

func (i *Impl) InitSchema() {
	i.DB.AutoMigrate(&Reminder{})
}

func (i *Impl) GetAllReminders(w rest.ResponseWriter, r *rest.Request) {
	reminders := []Reminder{}
	i.DB.Find(&reminders)
	w.WriteJson(&reminders)
}

func (i *Impl) GetReminder(w rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")
	reminder := Reminder{}
	if i.DB.First(&reminder, id).Error != nil {
		rest.NotFound(w, r)
		return
	}
	w.WriteJson(&reminder)
}

func (i *Impl) PostReminder(w rest.ResponseWriter, r *rest.Request) {
	reminder := Reminder{}
	if err := r.DecodeJsonPayload(&reminder); err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := i.DB.Save(&reminder).Error; err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(&reminder)
}

func (i *Impl) PutReminder(w rest.ResponseWriter, r *rest.Request) {

	id := r.PathParam("id")
	reminder := Reminder{}
	if i.DB.First(&reminder, id).Error != nil {
		rest.NotFound(w, r)
		return
	}

	updated := Reminder{}
	if err := r.DecodeJsonPayload(&updated); err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reminder.Message = updated.Message

	if err := i.DB.Save(&reminder).Error; err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(&reminder)
}

func (i *Impl) DeleteReminder(w rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")
	reminder := Reminder{}
	if i.DB.First(&reminder, id).Error != nil {
		rest.NotFound(w, r)
		return
	}
	if err := i.DB.Delete(&reminder).Error; err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

```

#### CORS

Demonstrate how to setup CorsMiddleware around all the API endpoints.

curl demo:
```
curl -i http://127.0.0.1:8080/countries
```

code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
)

func main() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.Use(&rest.CorsMiddleware{
		RejectNonCorsRequests: false,
		OriginValidator: func(origin string, request *rest.Request) bool {
			return origin == "http://my.other.host"
		},
		AllowedMethods: []string{"GET", "POST", "PUT"},
		AllowedHeaders: []string{
			"Accept", "Content-Type", "X-Custom-Header", "Origin"},
		AccessControlAllowCredentials: true,
		AccessControlMaxAge:           3600,
	})
	router, err := rest.MakeRouter(
		rest.Get("/countries", GetAllCountries),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
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

#### JSONP

Demonstrate how to use the JSONP middleware.

curl demo:
``` sh
curl -i http://127.0.0.1:8080/
curl -i http://127.0.0.1:8080/?cb=parseResponse
```

code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
)

func main() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.Use(&rest.JsonpMiddleware{
		CallbackNameKey: "cb",
	})
	api.SetApp(rest.AppSimple(func(w rest.ResponseWriter, r *rest.Request) {
		w.WriteJson(map[string]string{"Body": "Hello World!"})
	}))
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

```

#### Basic Auth

Demonstrate how to setup AuthBasicMiddleware as a pre-routing middleware.

curl demo:
```
curl -i http://127.0.0.1:8080/
curl -i -u admin:admin http://127.0.0.1:8080/
```

code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
)

func main() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.Use(&rest.AuthBasicMiddleware{
		Realm: "test zone",
		Authenticator: func(userId string, password string) bool {
			if userId == "admin" && password == "admin" {
				return true
			}
			return false
		},
	})
	api.SetApp(rest.AppSimple(func(w rest.ResponseWriter, r *rest.Request) {
		w.WriteJson(map[string]string{"Body": "Hello World!"})
	}))
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

```

#### Status

Demonstrate how to setup a `/.status` endpoint

Inspired by memcached "stats", this optional feature can be enabled to help monitoring the service.
This example shows how to enable the stats, and how to setup the `/.status` route.

curl demo:
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

code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
)

func main() {
	api := rest.NewApi()
	statusMw := &rest.StatusMiddleware{}
	api.Use(statusMw)
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/.status", func(w rest.ResponseWriter, r *rest.Request) {
			w.WriteJson(statusMw.GetStatus())
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

```

#### Status Auth

Demonstrate how to setup a /.status endpoint protected with basic authentication.

This is a good use case of middleware applied to only one API endpoint.

curl demo:
```
curl -i http://127.0.0.1:8080/countries
curl -i http://127.0.0.1:8080/.status
curl -i -u admin:admin http://127.0.0.1:8080/.status
...
```

code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
)

func main() {
	api := rest.NewApi()
	statusMw := &rest.StatusMiddleware{}
	api.Use(statusMw)
	api.Use(rest.DefaultDevStack...)
	auth := &rest.AuthBasicMiddleware{
		Realm: "test zone",
		Authenticator: func(userId string, password string) bool {
			if userId == "admin" && password == "admin" {
				return true
			}
			return false
		},
	}
	router, err := rest.MakeRouter(
		rest.Get("/countries", GetAllCountries),
		rest.Get("/.status", auth.MiddlewareFunc(
			func(w rest.ResponseWriter, r *rest.Request) {
				w.WriteJson(statusMw.GetStatus())
			},
		)),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
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

More advanced use cases.

#### JWT

Demonstrates how to use the [Json Web Token Auth Middleware](https://github.com/StephanDollberg/go-json-rest-middleware-jwt) to authenticate via a JWT token.

curl demo:
``` sh
curl -d '{"username": "admin", "password": "admin"}' -H "Content-Type:application/json" http://localhost:8080/api/login
curl -H "Authorization:Bearer TOKEN_RETURNED_FROM_ABOVE" http://localhost:8080/api/auth_test
curl -H "Authorization:Bearer TOKEN_RETURNED_FROM_ABOVE" http://localhost:8080/api/refresh_token
```

code:
``` go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/StephanDollberg/go-json-rest-middleware-jwt"
	"github.com/ant0ine/go-json-rest/rest"
)

func handle_auth(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(map[string]string{"authed": r.Env["REMOTE_USER"].(string)})
}

func main() {
	jwt_middleware := &jwt.JWTMiddleware{
		Key:        []byte("secret key"),
		Realm:      "jwt auth",
		Timeout:    time.Hour,
		MaxRefresh: time.Hour * 24,
		Authenticator: func(userId string, password string) bool {
			return userId == "admin" && password == "admin"
		}}

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	// we use the IfMiddleware to remove certain paths from needing authentication
	api.Use(&rest.IfMiddleware{
		Condition: func(request *rest.Request) bool {
			return request.URL.Path != "/login"
		},
		IfTrue: jwt_middleware,
	})
	api_router, _ := rest.MakeRouter(
		rest.Post("/login", jwt_middleware.LoginHandler),
		rest.Get("/auth_test", handle_auth),
		rest.Get("/refresh_token", jwt_middleware.RefreshHandler),
	)
	api.SetApp(api_router)

	http.Handle("/api/", http.StripPrefix("/api", api.MakeHandler()))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

```

#### Streaming

Demonstrate a streaming REST API, where the data is "flushed" to the client ASAP.

The stream format is a Line Delimited JSON.

curl demo:
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

code:
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
	api := rest.NewApi()
	api.Use(&rest.AccessLogApacheMiddleware{})
	api.Use(rest.DefaultCommonStack...)
	router, err := rest.MakeRouter(
		rest.Get("/stream", StreamThings),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
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
b) combine the api.MakeHandler() with another http.Handler
   (see api-and-static/ example)

That been said, exceptionally, it can be convenient to return a
different content type on a JSON endpoint. In this case, setting the
Content-Type and using the type assertion to access the Write method
is enough. As shown in this example.

curl demo:
```
curl -i http://127.0.0.1:8080/message.txt
```

code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
)

func main() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/message.txt", func(w rest.ResponseWriter, req *rest.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.(http.ResponseWriter).Write([]byte("Hello World!"))
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

```

#### API Versioning

First, API versioning is not easy and you may want to favor a mechanism that uses only backward compatible changes and deprecation cycles.

That been said, here is an example of API versioning using [Semver](http://semver.org/)

It defines a middleware that parses the version, checks a min and a max, and makes it available in the `request.Env`.

curl demo:
``` sh
curl -i http://127.0.0.1:8080/api/1.0.0/message
curl -i http://127.0.0.1:8080/api/2.0.0/message
curl -i http://127.0.0.1:8080/api/2.0.1/message
curl -i http://127.0.0.1:8080/api/0.0.1/message
curl -i http://127.0.0.1:8080/api/4.0.1/message

```

code:
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
			rest.Error(
				writer,
				"Invalid version: "+err.Error(),
				http.StatusBadRequest,
			)
			return
		}

		if version.LessThan(*minVersion) {
			rest.Error(
				writer,
				"Min supported version is "+minVersion.String(),
				http.StatusBadRequest,
			)
			return
		}

		if maxVersion.LessThan(*version) {
			rest.Error(
				writer,
				"Max supported version is "+maxVersion.String(),
				http.StatusBadRequest,
			)
			return
		}

		request.Env["VERSION"] = version
		handler(writer, request)
	}
}

func main() {

	svmw := SemVerMiddleware{
		MinVersion: "1.0.0",
		MaxVersion: "3.0.0",
	}
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/#version/message", svmw.MiddlewareFunc(
			func(w rest.ResponseWriter, req *rest.Request) {
				version := req.Env["VERSION"].(*semver.Version)
				if version.Major == 2 {
					// http://en.wikipedia.org/wiki/Second-system_effect
					w.WriteJson(map[string]string{
						"Body": "Hello broken World!",
					})
				} else {
					w.WriteJson(map[string]string{
						"Body": "Hello World!",
					})
				}
			},
		)),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	http.Handle("/api/", http.StripPrefix("/api", api.MakeHandler()))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

```

#### Statsd

Demonstrate how to use the [Statsd Middleware](https://github.com/ant0ine/go-json-rest-middleware-statsd) to collect statistics about the requests/reponses.
This middleware is based on the [g2s](https://github.com/peterbourgon/g2s) statsd client.

curl demo:
``` sh
# start statsd server
# monitor network
ngrep -d any port 8125

curl -i http://127.0.0.1:8080/message
curl -i http://127.0.0.1:8080/doesnotexist

```

code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest-middleware-statsd"
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
	"time"
)

func main() {
	api := rest.NewApi()
	api.Use(&statsd.StatsdMiddleware{})
	api.Use(rest.DefaultDevStack...)
	api.SetApp(rest.AppSimple(func(w rest.ResponseWriter, req *rest.Request) {

		// take more than 1ms so statsd can report it
		time.Sleep(100 * time.Millisecond)

		w.WriteJson(map[string]string{"Body": "Hello World!"})
	}))
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

```

#### NewRelic

NewRelic integration based on the GoRelic plugin: [github.com/yvasiyarov/gorelic](http://github.com/yvasiyarov/gorelic)

curl demo:
``` sh
curl -i http://127.0.0.1:8080/
```

code:
``` go
package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/yvasiyarov/go-metrics"
	"github.com/yvasiyarov/gorelic"
	"log"
	"net/http"
	"time"
)

type NewRelicMiddleware struct {
	License string
	Name    string
	Verbose bool
	agent   *gorelic.Agent
}

func (mw *NewRelicMiddleware) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {

	mw.agent = gorelic.NewAgent()
	mw.agent.NewrelicLicense = mw.License
	mw.agent.HTTPTimer = metrics.NewTimer()
	mw.agent.Verbose = mw.Verbose
	mw.agent.NewrelicName = mw.Name
	mw.agent.CollectHTTPStat = true
	mw.agent.Run()

	return func(writer rest.ResponseWriter, request *rest.Request) {

		handler(writer, request)

		// the timer middleware keeps track of the time
		startTime := request.Env["START_TIME"].(*time.Time)
		mw.agent.HTTPTimer.UpdateSince(*startTime)
	}
}

func main() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.Use(&NewRelicMiddleware{
		License: "<REPLACE WITH THE LICENSE KEY>",
		Name:    "<REPLACE WITH THE APP NAME>",
		Verbose: true,
	})
	api.SetApp(rest.AppSimple(func(w rest.ResponseWriter, r *rest.Request) {
		w.WriteJson(map[string]string{"Body": "Hello World!"})
	}))
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

```

#### Graceful Shutdown

This example uses [github.com/stretchr/graceful](https://github.com/stretchr/graceful) to try to be nice with the clients waiting for responses during a server shutdown (or restart).
The HTTP response takes 10 seconds to be completed, printing a message on the wire every second.
10 seconds is also the timeout set for the graceful shutdown.
You can play with these numbers to show that the server waits for the responses to complete.

curl demo:
``` sh
curl -i http://127.0.0.1:8080/message
```

code:
``` go
package main

import (
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/stretchr/graceful"
	"log"
	"net/http"
	"time"
)

func main() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/message", func(w rest.ResponseWriter, req *rest.Request) {
			for cpt := 1; cpt <= 10; cpt++ {

				// wait 1 second
				time.Sleep(time.Duration(1) * time.Second)

				w.WriteJson(map[string]string{
					"Message": fmt.Sprintf("%d seconds", cpt),
				})
				w.(http.ResponseWriter).Write([]byte("\n"))

				// Flush the buffer to client
				w.(http.Flusher).Flush()
			}
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)

	server := &graceful.Server{
		Timeout: 10 * time.Second,
		Server: &http.Server{
			Addr:    ":8080",
			Handler: api.MakeHandler(),
		},
	}

	log.Fatal(server.ListenAndServe())
}

```

#### SPDY

Demonstrate how to use SPDY with https://github.com/shykes/spdy-go

For a command line client, install spdycat from:
https://github.com/tatsuhiro-t/spdylay

spdycat demo:
```
spdycat -v --no-tls -2 http://localhost:8080/users/0
```

code:
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
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/users/:id", GetUser),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(spdy.ListenAndServeTCP(":8080", api.MakeHandler()))
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

curl demo:
```
curl -i http://127.0.0.1:8080/message
```

code:
``` go
package gaehelloworld

import (
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
)

func init() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		&rest.Get("/message", func(w rest.ResponseWriter, req *rest.Request) {
			w.WriteJson(map[string]string{"Body": "Hello World!"})
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	http.Handle("/", api.MakeHandler())
}

```

#### Websocket

Demonstrate how to run websocket in go-json-rest

go client demo:
```go
origin := "http://localhost:8080/"
url := "ws://localhost:8080/ws"
ws, err := websocket.Dial(url, "", origin)
if err != nil {
	log.Fatal(err)
}
if _, err := ws.Write([]byte("hello, world\n")); err != nil {
	log.Fatal(err)
}
var msg = make([]byte, 512)
var n int
if n, err = ws.Read(msg); err != nil {
	log.Fatal(err)
}
log.Printf("Received: %s.", msg[:n])
```

code:
``` go
package main

import (
	"io"
	"log"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	"golang.org/x/net/websocket"
)

func main() {
	wsHandler := websocket.Handler(func(ws *websocket.Conn) {
		io.Copy(ws, ws)
	})

	router, err := rest.MakeRouter(
		rest.Get("/ws", func(w rest.ResponseWriter, r *rest.Request) {
			wsHandler.ServeHTTP(w.(http.ResponseWriter), r.Request)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

```



## External Documentation

- [Online Documentation (godoc.org)](https://godoc.org/github.com/ant0ine/go-json-rest/rest)

Old v1 blog posts:

- [(Blog Post) Introducing Go-Json-Rest] (http://blog.ant0ine.com/typepad/2013/04/introducing-go-json-rest.html)
- [(Blog Post) Better URL Routing ?] (http://blog.ant0ine.com/typepad/2013/02/better-url-routing-golang-1.html)


## Version 3 release notes

### What's New in v3

* Public Middlewares. (12 included in the package)
* A new App interface. (the router being the provided App)
* A new Api object that manages the Middlewares and the App.
* Optional and interchangeable App/router.

### Here is for instance the new minimal "Hello World!"

```go
api := rest.NewApi()
api.Use(rest.DefaultDevStack...)
api.SetApp(rest.AppSimple(func(w rest.ResponseWriter, r *rest.Request) {
        w.WriteJson(map[string]string{"Body": "Hello World!"})
}))
http.ListenAndServe(":8080", api.MakeHandler())
```

*All 19 examples have been updated to use the new API. [See here](https://github.com/ant0ine/go-json-rest#examples)*

### Deprecating the ResourceHandler

V3 is about deprecating the ResourceHandler in favor of a new API that exposes the Middlewares. As a consequence, all the Middlewares are now public, and the new Api object helps putting them together as a stack. Some default stack configurations are offered. The router is now an App that sits on top of the stack of Middlewares. Which means that the router is no longer required to use Go-Json-Rest.

*Design ideas and discussion [See here](https://github.com/ant0ine/go-json-rest/issues/110)*


## Migration guide from v2 to v3

V3 introduces an API change (see [Semver](http://semver.org/)). But it was possible to maintain backward compatibility, and so, ResourceHandler still works.
ResourceHandler does the same thing as in V2, **but it is now considered as deprecated, and will be removed in a few months**. In the meantime, it logs a
deprecation warning.

### How to map the ResourceHandler options to the new stack of middlewares ?

* `EnableGzip bool`: Just include GzipMiddleware in the stack of middlewares.
* `DisableJsonIndent bool`: Just don't include JsonIndentMiddleware in the stack of middlewares.
* `EnableStatusService bool`: Include StatusMiddleware in the stack and keep a reference to it to access GetStatus().
* `EnableResponseStackTrace bool`: Same exact option but moved to RecoverMiddleware.
* `EnableLogAsJson bool`: Include AccessLogJsonMiddleware, and possibly remove AccessLogApacheMiddleware.
* `EnableRelaxedContentType bool`: Just don't include ContentTypeCheckerMiddleware.
* `OuterMiddlewares []Middleware`: You are now building the full stack, OuterMiddlewares are the first in the list.
* `PreRoutingMiddlewares []Middleware`: You are now building the full stack, OuterMiddlewares are the last in the list.
* `Logger *log.Logger`: Same option but moved to AccessLogApacheMiddleware and AccessLogJsonMiddleware.
* `LoggerFormat AccessLogFormat`: Same exact option but moved to AccessLogApacheMiddleware.
* `DisableLogger bool`: Just don't include any access log middleware.
* `ErrorLogger *log.Logger`: Same exact option but moved to RecoverMiddleware.
* `XPoweredBy string`: Same exact option but moved to PoweredByMiddleware.
* `DisableXPoweredBy bool`: Just don't include PoweredByMiddleware.


## Version 2 release notes

* Middlewares, the notion of middleware is now formally defined. They can be setup as global pre-routing Middlewares wrapping all the endpoints, or on a per endpoint basis.
In fact the internal code of **go-json-rest** is itself implemented with Middlewares, they are just hidden behind configuration boolean flags to make these very common options even easier to use.

* A new ResponseWriter. This is now an interface, and allows Middlewares to wrap the writer. The provided writer implements, in addition of *rest.ResponseWriter*, *http.Flusher*, *http.CloseNotifier*, *http.Hijacker*, and *http.ResponseWriter*. A lot more Go-ish, and very similar to `net/http`.

* The AuthBasic and CORS Middlewares have been added. More to come in the future.

* Faster, more tasks are performed at init time, and less for each request.

* New documentation, with more examples.

* A lot of other small improvements, See the [Migration guide to v2](#migration-guide-from-v1-to-v2)


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
- [Thanabodee Charoenpiriyakij](https://github.com/wingyplus)
- [Sebastien Estienne](https://github.com/sebest)


Copyright (c) 2013-2015 Antoine Imbert

[MIT License](https://github.com/ant0ine/go-json-rest/blob/master/LICENSE)

[![Analytics](https://ga-beacon.appspot.com/UA-309210-4/go-json-rest/master/readme)](https://github.com/igrigorik/ga-beacon)

