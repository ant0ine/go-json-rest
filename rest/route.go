package rest

import (
	"strings"
)

// Route defines a route. It's used with SetRoutes.
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
