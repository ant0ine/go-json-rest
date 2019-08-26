package rest

import (
	"strings"
)

// Route defines a route as consumed by the router. It can be instantiated directly, or using one
// of the shortcut methods: rest.Get, rest.Post, rest.Put, rest.Patch and rest.Delete.
type Route struct {

	// Any HTTP method. It will be used as uppercase to avoid common mistakes.
	HttpMethod string

	// A string like "/resource/:id.json".
	// Placeholders supported are:
	// :paramName that matches any char to the first '/' or '.'
	// #paramName that matches any char to the first '/'
	// *paramName that matches everything to the end of the string
	// (placeholder names must be unique per PathExp)
	PathExp string

	// Code that will be executed when this route is taken.
	Func HandlerFunc
}

// MakePath generates the path corresponding to this Route and the provided path parameters.
// This is used for reverse route resolution.
func (route *Route) MakePath(pathParams map[string]string) string {
	path := route.PathExp
	for paramName, paramValue := range pathParams {
		paramPlaceholder := ":" + paramName
		relaxedPlaceholder := "#" + paramName
		splatPlaceholder := "*" + paramName
		r := strings.NewReplacer(paramPlaceholder, paramValue, splatPlaceholder, paramValue, relaxedPlaceholder, paramValue)
		path = r.Replace(path)
	}
	return path
}

// Head is a shortcut method that instantiates a HEAD route. See the Route object the parameters definitions.
// Equivalent to &Route{"HEAD", pathExp, handlerFunc}
func Head(pathExp string, handlerFunc HandlerFunc) *Route {
	return &Route{
		HttpMethod: "HEAD",
		PathExp:    pathExp,
		Func:       handlerFunc,
	}
}

// Get is a shortcut method that instantiates a GET route. See the Route object the parameters definitions.
// Equivalent to &Route{"GET", pathExp, handlerFunc}
func Get(pathExp string, handlerFunc HandlerFunc) *Route {
	return &Route{
		HttpMethod: "GET",
		PathExp:    pathExp,
		Func:       handlerFunc,
	}
}

// Post is a shortcut method that instantiates a POST route. See the Route object the parameters definitions.
// Equivalent to &Route{"POST", pathExp, handlerFunc}
func Post(pathExp string, handlerFunc HandlerFunc) *Route {
	return &Route{
		HttpMethod: "POST",
		PathExp:    pathExp,
		Func:       handlerFunc,
	}
}

// Put is a shortcut method that instantiates a PUT route.  See the Route object the parameters definitions.
// Equivalent to &Route{"PUT", pathExp, handlerFunc}
func Put(pathExp string, handlerFunc HandlerFunc) *Route {
	return &Route{
		HttpMethod: "PUT",
		PathExp:    pathExp,
		Func:       handlerFunc,
	}
}

// Patch is a shortcut method that instantiates a PATCH route.  See the Route object the parameters definitions.
// Equivalent to &Route{"PATCH", pathExp, handlerFunc}
func Patch(pathExp string, handlerFunc HandlerFunc) *Route {
	return &Route{
		HttpMethod: "PATCH",
		PathExp:    pathExp,
		Func:       handlerFunc,
	}
}

// Delete is a shortcut method that instantiates a DELETE route. Equivalent to &Route{"DELETE", pathExp, handlerFunc}
func Delete(pathExp string, handlerFunc HandlerFunc) *Route {
	return &Route{
		HttpMethod: "DELETE",
		PathExp:    pathExp,
		Func:       handlerFunc,
	}
}

// Options is a shortcut method that instantiates an OPTIONS route.  See the Route object the parameters definitions.
// Equivalent to &Route{"OPTIONS", pathExp, handlerFunc}
func Options(pathExp string, handlerFunc HandlerFunc) *Route {
	return &Route{
		HttpMethod: "OPTIONS",
		PathExp:    pathExp,
		Func:       handlerFunc,
	}
}
