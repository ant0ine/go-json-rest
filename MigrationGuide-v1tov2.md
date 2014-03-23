
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

This change allows the `ResponseWriter` to be wrapped, like the one of the `net/http` package. Middlewares like Gzip used this to encode the payload (see gzip.go)
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

####  The notion of Middleware is now formally defined

A middleware is an object satisfying this interface:
~~~ go
type Middleware interface {
	MiddlewareFunc(handler HandlerFunc) HandlerFunc
}
~~~

Code using PreRoutingMiddleware will have to be adapted to provide a list of Middleware objects.
See the [Basic Auth example](https://github.com/ant0ine/go-json-rest-examples/blob/v2-alpha/auth-basic/main.go).


[![Analytics](https://ga-beacon.appspot.com/UA-309210-4/go-json-rest/v2-alpha/MigrationGuide-v1tov2.md)](https://github.com/igrigorik/ga-beacon)
