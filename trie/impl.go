// Special Trie implementation for HTTP routing.
//
// This Trie implementation is designed to support strings that includes
// :param and *splat parameters. Strings that are commonly used to represent
// the Path in HTTP routing. This implementation also maintain for each Path
// a map of HTTP Methods associated with the Route.
//
// You probably don't need to use this package directly.
//
package trie

import (
	"errors"
	"fmt"
)

func splitParam(remaining string) (string, string) {
	i := 0
	for len(remaining) > i && remaining[i] != '/' && remaining[i] != '.' {
		i++
	}
	return remaining[:i], remaining[i:]
}

type node struct {
	HttpMethodToRoute map[string]interface{}
	Children          map[string]*node
	ChildrenKeyLen    int
	ParamChild        *node
	ParamName         string
	SplatChild        *node
	SplatName         string
}

func (self *node) addRoute(httpMethod, pathExp string, route interface{}, usedParams []string) error {

	if len(pathExp) == 0 {
		// end of the path, leaf node, update the map
		if self.HttpMethodToRoute == nil {
			self.HttpMethodToRoute = map[string]interface{}{
				httpMethod: route,
			}
			return nil
		} else {
			if self.HttpMethodToRoute[httpMethod] != nil {
				return errors.New("node.Route already set, duplicated path and method")
			}
			self.HttpMethodToRoute[httpMethod] = route
			return nil
		}
	}

	token := pathExp[0:1]
	remaining := pathExp[1:]
	var nextNode *node

	if token[0] == ':' {
		// :param case
		var name string
		name, remaining = splitParam(remaining)

		// Check param name is unique
		for _, e := range usedParams {
			if e == name {
				return errors.New(
					fmt.Sprintf("A route can't have two params with the same name: %s", name),
				)
			}
		}
		usedParams = append(usedParams, name)

		if self.ParamChild == nil {
			self.ParamChild = &node{}
			self.ParamName = name
		} else {
			if self.ParamName != name {
				return errors.New(
					fmt.Sprintf(
						"Routes sharing a common placeholder MUST name it consistently: %s != %s",
						self.ParamName,
						name,
					),
				)
			}
		}
		nextNode = self.ParamChild
	} else if token[0] == '*' {
		// *splat case
		name := remaining
		remaining = ""
		if self.SplatChild == nil {
			self.SplatChild = &node{}
			self.SplatName = name
		}
		nextNode = self.SplatChild
	} else {
		// general case
		if self.Children == nil {
			self.Children = map[string]*node{}
			self.ChildrenKeyLen = 1
		}
		if self.Children[token] == nil {
			self.Children[token] = &node{}
		}
		nextNode = self.Children[token]
	}

	return nextNode.addRoute(httpMethod, remaining, route, usedParams)
}

// utility for the node.findRoutes recursive method
type findContext struct {
	paramStack []map[string]string
	matchFunc  func(httpMethod, path string, node *node)
}

func newFindContext() *findContext {
	return &findContext{
		paramStack: []map[string]string{},
	}
}

func (self *findContext) pushParams(name, value string) {
	self.paramStack = append(
		self.paramStack,
		map[string]string{name: value},
	)
}

func (self *findContext) popParams() {
	self.paramStack = self.paramStack[:len(self.paramStack)-1]
}

func (self *findContext) paramsAsMap() map[string]string {
	r := map[string]string{}
	for _, param := range self.paramStack {
		for key, value := range param {
			if r[key] != "" {
				// this is checked at addRoute time, and should never happen.
				panic(fmt.Sprintf(
					"placeholder %s already found, placeholder names should be unique per route",
					key,
				))
			}
			r[key] = value
		}
	}
	return r
}

type Match struct {
	// Same Route as in AddRoute
	Route interface{}
	// map of params matched for this result
	Params map[string]string
}

func (self *node) find(httpMethod, path string, context *findContext) {

	if self.HttpMethodToRoute != nil && path == "" {
		context.matchFunc(httpMethod, path, self)
	}

	if len(path) == 0 {
		return
	}

	// *splat branch
	if self.SplatChild != nil {
		context.pushParams(self.SplatName, path)
		self.SplatChild.find(httpMethod, "", context)
		context.popParams()
	}

	// :param branch
	if self.ParamChild != nil {
		value, remaining := splitParam(path)
		context.pushParams(self.ParamName, value)
		self.ParamChild.find(httpMethod, remaining, context)
		context.popParams()
	}

	// main branch
	length := self.ChildrenKeyLen
	if len(path) < length {
		return
	}
	token := path[0:length]
	remaining := path[length:]
	if self.Children[token] != nil {
		self.Children[token].find(httpMethod, remaining, context)
	}
}

func (self *node) compress() {
	// *splat branch
	if self.SplatChild != nil {
		self.SplatChild.compress()
	}
	// :param branch
	if self.ParamChild != nil {
		self.ParamChild.compress()
	}
	// main branch
	if len(self.Children) == 0 {
		return
	}
	// compressable ?
	canCompress := true
	for _, node := range self.Children {
		if node.HttpMethodToRoute != nil || node.SplatChild != nil || node.ParamChild != nil {
			canCompress = false
		}
	}
	// compress
	if canCompress {
		merged := map[string]*node{}
		for key, node := range self.Children {
			for gdKey, gdNode := range node.Children {
				mergedKey := key + gdKey
				merged[mergedKey] = gdNode
			}
		}
		self.Children = merged
		self.ChildrenKeyLen++
		self.compress()
		// continue
	} else {
		for _, node := range self.Children {
			node.compress()
		}
	}
}

type Trie struct {
	root *node
}

// Instanciate a Trie with an empty node as the root.
func New() *Trie {
	return &Trie{
		root: &node{},
	}
}

// Insert the route in the Trie following or creating the nodes corresponding to the path.
func (self *Trie) AddRoute(httpMethod, pathExp string, route interface{}) error {
	return self.root.addRoute(httpMethod, pathExp, route, []string{})
}

// Given a path and an http method, return all the matching routes.
func (self *Trie) FindRoutes(httpMethod, path string) []*Match {
	context := newFindContext()
	matches := []*Match{}
	context.matchFunc = func(httpMethod, path string, node *node) {
		if node.HttpMethodToRoute[httpMethod] != nil {
			// path and method match, found a route !
			matches = append(
				matches,
				&Match{
					Route:  node.HttpMethodToRoute[httpMethod],
					Params: context.paramsAsMap(),
				},
			)
		}
	}
	self.root.find(httpMethod, path, context)
	return matches
}

// Same as FindRoutes, but return in addition a boolean indicating if the path was matched.
// Useful to return 405
func (self *Trie) FindRoutesAndPathMatched(httpMethod, path string) ([]*Match, bool) {
	context := newFindContext()
	pathMatched := false
	matches := []*Match{}
	context.matchFunc = func(httpMethod, path string, node *node) {
		pathMatched = true
		if node.HttpMethodToRoute[httpMethod] != nil {
			// path and method match, found a route !
			matches = append(
				matches,
				&Match{
					Route:  node.HttpMethodToRoute[httpMethod],
					Params: context.paramsAsMap(),
				},
			)
		}
	}
	self.root.find(httpMethod, path, context)
	return matches, pathMatched
}

// Given a path, and whatever the http method, return all the matching routes.
func (self *Trie) FindRoutesForPath(path string) []*Match {
	context := newFindContext()
	matches := []*Match{}
	context.matchFunc = func(httpMethod, path string, node *node) {
		params := context.paramsAsMap()
		for _, route := range node.HttpMethodToRoute {
			matches = append(
				matches,
				&Match{
					Route:  route,
					Params: params,
				},
			)
		}
	}
	self.root.find("", path, context)
	return matches
}

// Reduce the size of the tree, must be done after the last AddRoute.
func (self *Trie) Compress() {
	self.root.compress()
}
