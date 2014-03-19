
# Migration guide from v1 to v2

**Go-Json-Rest** follows [Semver](http://semver.org/) and a few breaking changes have been introduced with the v2.

### The import path has changed to `github.com/ant0ine/go-json-rest/rest`

This is more conform to the Go style, and makes [goimports](https://godoc.org/code.google.com/p/go.tools/cmd/goimports) working.

### rest.ResponseWriter is now an interface.

This is the main change, and most of the program will be migrated with a simple s/\*\.rest\.ResponseWriter/rest\.ResponseWriter/g

### Flush(), CloseNotify() and Write() are not directly exposed anymore.

A type assertion of the corresponding interface is necessary. eg: writer.(http.Flusher).Flush()

### The /.status endpoint is not created automatically anymore.

The route has to be manually set as shown on the "status" example.

###  The notion of Middleware is now formally defined

and code using PreRoutingMiddleware will have to be adapted to provide a list of Middleware objects. See the [Basic Auth example](https://github.com/ant0ine/go-json-rest-examples/blob/v2-alpha/auth-basic/main.go).


[![Analytics](https://ga-beacon.appspot.com/UA-309210-4/go-json-rest/v2-alpha/MigrationGuide-v1tov2.md)](https://github.com/igrigorik/ga-beacon)
