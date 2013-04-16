package rest

import (
	"github.com/ant0ine/go-json-rest/trie"
	"net/url"
	"strings"
)

// TODO
// support for #param placeholder ?

type router struct {
	routes                 []Route
	disableTrieCompression bool
	index                  map[*Route]int
	trie                   *trie.Trie
}

// This validates the Routes and prepares the Trie data structure.
// It must be called once the Routes are defined and before trying to find Routes.
// The order matters, if multiple Routes match, the first defined will be used.
func (self *router) start() error {

	self.trie = trie.New()
	self.index = map[*Route]int{}

	for i, _ := range self.routes {

		// pointer to the Route
		route := &self.routes[i]

		// insert in the Trie
		err := self.trie.AddRoute(
			strings.ToUpper(route.HttpMethod),
			route.PathExp,
			route,
		)
		if err != nil {
			return err
		}

		// index
		self.index[route] = i
	}

	if self.disableTrieCompression == false {
		self.trie.Compress()
	}

	// TODO validation of the PathExp ? start with a /
	// TODO url encoding

	return nil
}

// Return the first matching Route and the corresponding parameters for a given URL object.
func (self *router) findRouteFromURL(httpMethod string, urlObj *url.URL) (*Route, map[string]string) {

	// lookup the routes in the Trie
	// TODO verify url encoding
	results := self.trie.FindRoutes(
		strings.ToUpper(httpMethod),
		urlObj.Path,
	)

	// short cut
	if len(results) == 1 {
		return results[0].Route.(*Route), results[0].Params
	}

	// only return the first Route that results
	minIndex := -1
	resultsByIndex := map[int]*trie.Result{}

	for _, result := range results {
		route := result.Route.(*Route)
		routeIndex := self.index[route]
		resultsByIndex[routeIndex] = result
		if minIndex == -1 || routeIndex < minIndex {
			minIndex = routeIndex
		}
	}

	if minIndex == -1 {
		// no route found
		return nil, nil
	}

	// and the corresponding params
	result := resultsByIndex[minIndex]

	return result.Route.(*Route), result.Params
}

// Parse the url string (complete or just the path) and return the first matching Route and the corresponding parameters.
func (self *router) findRoute(httpMethod, urlStr string) (*Route, map[string]string, error) {

	// parse the url
	urlObj, err := url.Parse(urlStr)
	if err != nil {
		return nil, nil, err
	}

	route, params := self.findRouteFromURL(httpMethod, urlObj)
	return route, params, nil
}
