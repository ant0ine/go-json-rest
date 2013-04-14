// Efficient URL routing using a Trie data structure.
//
// This Package implements a URL Router, but instead of using the usual
// "evaluate all the routes and return the first regexp that matches" strategy,
// it uses a Trie data structure to perform the routing. This is more efficient,
// and scales better for a large number of routes.
// It supports the :param and *splat placeholders in the route strings.
//
// Example:
//	router := urlrouter.Router{
//		Routes: []urlrouter.Route{
//			urlrouter.Route{
//				PathExp: "/resources/:id",
//				Dest:    "one_resource",
//			},
//			urlrouter.Route{
//				PathExp: "/resources",
//				Dest:    "all_resources",
//			},
//		},
//	}
//
//	err := router.start()
//	if err != nil {
//		panic(err)
//	}
//
//	input := "http://example.org/resources/123"
//	route, params, err := router.findRoute(input)
//	if err != nil {
//		panic(err)
//	}
//	fmt.Print(route.Dest)  // one_resource
//	fmt.Print(params["id"])  // 123
//
// (Blog Post: http://blog.ant0ine.com/typepad/2013/02/better-url-routing-golang-1.html)
package rest

import (
	"errors"
	"github.com/ant0ine/go-json-rest/trie"
	"net/url"
)

// TODO
// support for #param placeholder ?

// XXX Router should be a private object
type Router struct {
	// list of Routes, the order matters, if multiple Routes match, the first defined will be used.
	Routes                 []Route
	disableTrieCompression bool
	index                  map[*Route]int
	trie                   *trie.Trie
}

// This validates the Routes and prepares the Trie data structure.
// It must be called once the Routes are defined and before trying to find Routes.
func (self *Router) start() error {

	self.trie = trie.New()
	self.index = map[*Route]int{}
	unique := map[string]bool{}

	for i, _ := range self.Routes {
		// pointer to the Route
		route := &self.Routes[i]
		// unique
		if unique[route.PathExp] == true {
			return errors.New("duplicated PathExp")
		}
		unique[route.PathExp] = true
		// index
		self.index[route] = i
		// insert in the Trie
		err := self.trie.AddRoute(route.PathExp, route)
		if err != nil {
			return err
		}
	}

	if self.disableTrieCompression == false {
		self.trie.Compress()
	}

	// TODO validation of the PathExp ? start with a /
	// TODO url encoding

	return nil
}

// Return the first matching Route and the corresponding parameters for a given URL object.
func (self *Router) findRouteFromURL(urlObj *url.URL) (*Route, map[string]string) {

	// lookup the routes in the Trie
	// TODO verify url encoding
	matches := self.trie.FindRoutes(urlObj.Path)

	// only return the first Route that matches
	minIndex := -1
	matchesByIndex := map[int]*trie.Match{}

	for _, match := range matches {
		route := match.RouteValue.(*Route)
		routeIndex := self.index[route]
		matchesByIndex[routeIndex] = match
		if minIndex == -1 || routeIndex < minIndex {
			minIndex = routeIndex
		}
	}

	if minIndex == -1 {
		// no route found
		return nil, nil
	}

	// and the corresponding params
	match := matchesByIndex[minIndex]

	return match.RouteValue.(*Route), match.Params
}

// Parse the url string (complete or just the path) and return the first matching Route and the corresponding parameters.
func (self *Router) findRoute(urlStr string) (*Route, map[string]string, error) {

	// parse the url
	urlObj, err := url.Parse(urlStr)
	if err != nil {
		return nil, nil, err
	}

	route, params := self.findRouteFromURL(urlObj)
	return route, params, nil
}
