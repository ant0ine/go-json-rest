package rest

import (
	"errors"
	"github.com/ant0ine/go-json-rest/rest/trie"
	"net/http"
	"net/url"
	"strings"
)

type router struct {
	Routes []*Route

	disableTrieCompression bool
	index                  map[*Route]int
	trie                   *trie.Trie
}

// MakeRouter returns the router app. Given a set of Routes, it dispatches the request to the
// HandlerFunc of the first route that matches. The order of the Routes matters.
func MakeRouter(routes ...*Route) (App, error) {
	r := &router{
		Routes: routes,
	}
	err := r.start()
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Handle the REST routing and run the user code.
func (rt *router) AppFunc() HandlerFunc {
	return func(writer ResponseWriter, request *Request) {

		// find the route
		route, params, pathMatched := rt.findRouteFromURL(request.Method, request.URL)
		if route == nil {

			if pathMatched {
				// no route found, but path was matched: 405 Method Not Allowed
				Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			// no route found, the path was not matched: 404 Not Found
			NotFound(writer, request)
			return
		}

		// a route was found, set the PathParams
		request.PathParams = params

		// run the user code
		handler := route.Func
		handler(writer, request)
	}
}

// This is run for each new request, perf is important.
func escapedPath(urlObj *url.URL) string {
	// the escape method of url.URL should be public
	// that would avoid this split.
	parts := strings.SplitN(urlObj.RequestURI(), "?", 2)
	return parts[0]
}

var preEscape = strings.NewReplacer("*", "__SPLAT_PLACEHOLDER__", "#", "__RELAXED_PLACEHOLDER__")

var postEscape = strings.NewReplacer("__SPLAT_PLACEHOLDER__", "*", "__RELAXED_PLACEHOLDER__", "#")

// This is run at init time only.
func escapedPathExp(pathExp string) (string, error) {

	// PathExp validation
	if pathExp == "" {
		return "", errors.New("empty PathExp")
	}
	if pathExp[0] != '/' {
		return "", errors.New("PathExp must start with /")
	}
	if strings.Contains(pathExp, "?") {
		return "", errors.New("PathExp must not contain the query string")
	}

	// Get the right escaping
	// XXX a bit hacky

	pathExp = preEscape.Replace(pathExp)

	urlObj, err := url.Parse(pathExp)
	if err != nil {
		return "", err
	}

	// get the same escaping as find requests
	pathExp = urlObj.RequestURI()

	pathExp = postEscape.Replace(pathExp)

	return pathExp, nil
}

// This validates the Routes and prepares the Trie data structure.
// It must be called once the Routes are defined and before trying to find Routes.
// The order matters, if multiple Routes match, the first defined will be used.
func (rt *router) start() error {

	rt.trie = trie.New()
	rt.index = map[*Route]int{}

	for i, route := range rt.Routes {

		// work with the PathExp urlencoded.
		pathExp, err := escapedPathExp(route.PathExp)
		if err != nil {
			return err
		}

		// insert in the Trie
		err = rt.trie.AddRoute(
			strings.ToUpper(route.HttpMethod), // work with the HttpMethod in uppercase
			pathExp,
			route,
		)
		if err != nil {
			return err
		}

		// index
		rt.index[route] = i
	}

	if rt.disableTrieCompression == false {
		rt.trie.Compress()
	}

	return nil
}

// return the result that has the route defined the earliest
func (rt *router) ofFirstDefinedRoute(matches []*trie.Match) *trie.Match {
	minIndex := -1
	var bestMatch *trie.Match

	for _, result := range matches {
		route := result.Route.(*Route)
		routeIndex := rt.index[route]
		if minIndex == -1 || routeIndex < minIndex {
			minIndex = routeIndex
			bestMatch = result
		}
	}

	return bestMatch
}

// Return the first matching Route and the corresponding parameters for a given URL object.
func (rt *router) findRouteFromURL(httpMethod string, urlObj *url.URL) (*Route, map[string]string, bool) {

	// lookup the routes in the Trie
	matches, pathMatched := rt.trie.FindRoutesAndPathMatched(
		strings.ToUpper(httpMethod), // work with the httpMethod in uppercase
		escapedPath(urlObj),         // work with the path urlencoded
	)

	// short cuts
	if len(matches) == 0 {
		// no route found
		return nil, nil, pathMatched
	}

	if len(matches) == 1 {
		// one route found
		return matches[0].Route.(*Route), matches[0].Params, pathMatched
	}

	// multiple routes found, pick the first defined
	result := rt.ofFirstDefinedRoute(matches)
	return result.Route.(*Route), result.Params, pathMatched
}

// Parse the url string (complete or just the path) and return the first matching Route and the corresponding parameters.
func (rt *router) findRoute(httpMethod, urlStr string) (*Route, map[string]string, bool, error) {

	// parse the url
	urlObj, err := url.Parse(urlStr)
	if err != nil {
		return nil, nil, false, err
	}

	route, params, pathMatched := rt.findRouteFromURL(httpMethod, urlObj)
	return route, params, pathMatched, nil
}
