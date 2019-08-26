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

func splitRelaxed(remaining string) (string, string) {
	i := 0
	for len(remaining) > i && remaining[i] != '/' {
		i++
	}
	return remaining[:i], remaining[i:]
}

type node struct {
	HttpMethodToRoute map[string]interface{}

	Children       map[string]*node
	ChildrenKeyLen int

	ParamChild *node
	ParamName  string

	RelaxedChild *node
	RelaxedName  string

	SplatChild *node
	SplatName  string
}

func (n *node) addRoute(httpMethod, pathExp string, route interface{}, usedParams []string) error {

	if len(pathExp) == 0 {
		// end of the path, leaf node, update the map
		if n.HttpMethodToRoute == nil {
			n.HttpMethodToRoute = map[string]interface{}{
				httpMethod: route,
			}
			return nil
		} else {
			if n.HttpMethodToRoute[httpMethod] != nil {
				return errors.New("node.Route already set, duplicated path and method")
			}
			n.HttpMethodToRoute[httpMethod] = route
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
					fmt.Sprintf("A route can't have two placeholders with the same name: %s", name),
				)
			}
		}
		usedParams = append(usedParams, name)

		if n.ParamChild == nil {
			n.ParamChild = &node{}
			n.ParamName = name
		} else {
			if n.ParamName != name {
				return errors.New(
					fmt.Sprintf(
						"Routes sharing a common placeholder MUST name it consistently: %s != %s",
						n.ParamName,
						name,
					),
				)
			}
		}
		nextNode = n.ParamChild
	} else if token[0] == '#' {
		// #param case
		var name string
		name, remaining = splitRelaxed(remaining)

		// Check param name is unique
		for _, e := range usedParams {
			if e == name {
				return errors.New(
					fmt.Sprintf("A route can't have two placeholders with the same name: %s", name),
				)
			}
		}
		usedParams = append(usedParams, name)

		if n.RelaxedChild == nil {
			n.RelaxedChild = &node{}
			n.RelaxedName = name
		} else {
			if n.RelaxedName != name {
				return errors.New(
					fmt.Sprintf(
						"Routes sharing a common placeholder MUST name it consistently: %s != %s",
						n.RelaxedName,
						name,
					),
				)
			}
		}
		nextNode = n.RelaxedChild
	} else if token[0] == '*' {
		// *splat case
		name := remaining
		remaining = ""

		// Check param name is unique
		for _, e := range usedParams {
			if e == name {
				return errors.New(
					fmt.Sprintf("A route can't have two placeholders with the same name: %s", name),
				)
			}
		}

		if n.SplatChild == nil {
			n.SplatChild = &node{}
			n.SplatName = name
		}
		nextNode = n.SplatChild
	} else {
		// general case
		if n.Children == nil {
			n.Children = map[string]*node{}
			n.ChildrenKeyLen = 1
		}
		if n.Children[token] == nil {
			n.Children[token] = &node{}
		}
		nextNode = n.Children[token]
	}

	return nextNode.addRoute(httpMethod, remaining, route, usedParams)
}

func (n *node) compress() {
	// *splat branch
	if n.SplatChild != nil {
		n.SplatChild.compress()
	}
	// :param branch
	if n.ParamChild != nil {
		n.ParamChild.compress()
	}
	// #param branch
	if n.RelaxedChild != nil {
		n.RelaxedChild.compress()
	}
	// main branch
	if len(n.Children) == 0 {
		return
	}
	// compressable ?
	canCompress := true
	for _, node := range n.Children {
		if node.HttpMethodToRoute != nil || node.SplatChild != nil || node.ParamChild != nil || node.RelaxedChild != nil {
			canCompress = false
		}
	}
	// compress
	if canCompress {
		merged := map[string]*node{}
		for key, node := range n.Children {
			for gdKey, gdNode := range node.Children {
				mergedKey := key + gdKey
				merged[mergedKey] = gdNode
			}
		}
		n.Children = merged
		n.ChildrenKeyLen++
		n.compress()
		// continue
	} else {
		for _, node := range n.Children {
			node.compress()
		}
	}
}

func printFPadding(padding int, format string, a ...interface{}) {
	for i := 0; i < padding; i++ {
		fmt.Print(" ")
	}
	fmt.Printf(format, a...)
}

// Private function for now
func (n *node) printDebug(level int) {
	level++
	// *splat branch
	if n.SplatChild != nil {
		printFPadding(level, "*splat\n")
		n.SplatChild.printDebug(level)
	}
	// :param branch
	if n.ParamChild != nil {
		printFPadding(level, ":param\n")
		n.ParamChild.printDebug(level)
	}
	// #param branch
	if n.RelaxedChild != nil {
		printFPadding(level, "#relaxed\n")
		n.RelaxedChild.printDebug(level)
	}
	// main branch
	for key, node := range n.Children {
		printFPadding(level, "\"%s\"\n", key)
		node.printDebug(level)
	}
}

// utility for the node.findRoutes recursive method

type paramMatch struct {
	name  string
	value string
}

type findContext struct {
	paramStack []paramMatch
	matchFunc  func(httpMethod, path string, node *node)
}

func newFindContext() *findContext {
	return &findContext{
		paramStack: []paramMatch{},
	}
}

func (fc *findContext) pushParams(name, value string) {
	fc.paramStack = append(
		fc.paramStack,
		paramMatch{name, value},
	)
}

func (fc *findContext) popParams() {
	fc.paramStack = fc.paramStack[:len(fc.paramStack)-1]
}

func (fc *findContext) paramsAsMap() map[string]string {
	r := map[string]string{}
	for _, param := range fc.paramStack {
		if r[param.name] != "" {
			// this is checked at addRoute time, and should never happen.
			panic(fmt.Sprintf(
				"placeholder %s already found, placeholder names should be unique per route",
				param.name,
			))
		}
		r[param.name] = param.value
	}
	return r
}

type Match struct {
	// Same Route as in AddRoute
	Route interface{}
	// map of params matched for this result
	Params map[string]string
}

func (n *node) find(httpMethod, path string, context *findContext) {

	if n.HttpMethodToRoute != nil && path == "" {
		context.matchFunc(httpMethod, path, n)
	}

	if len(path) == 0 {
		return
	}

	// *splat branch
	if n.SplatChild != nil {
		context.pushParams(n.SplatName, path)
		n.SplatChild.find(httpMethod, "", context)
		context.popParams()
	}

	// :param branch
	if n.ParamChild != nil {
		value, remaining := splitParam(path)
		context.pushParams(n.ParamName, value)
		n.ParamChild.find(httpMethod, remaining, context)
		context.popParams()
	}

	// #param branch
	if n.RelaxedChild != nil {
		value, remaining := splitRelaxed(path)
		context.pushParams(n.RelaxedName, value)
		n.RelaxedChild.find(httpMethod, remaining, context)
		context.popParams()
	}

	// main branch
	length := n.ChildrenKeyLen
	if len(path) < length {
		return
	}
	token := path[0:length]
	remaining := path[length:]
	if n.Children[token] != nil {
		n.Children[token].find(httpMethod, remaining, context)
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
func (t *Trie) AddRoute(httpMethod, pathExp string, route interface{}) error {
	return t.root.addRoute(httpMethod, pathExp, route, []string{})
}

// Reduce the size of the tree, must be done after the last AddRoute.
func (t *Trie) Compress() {
	t.root.compress()
}

// Private function for now.
func (t *Trie) printDebug() {
	fmt.Print("<trie>\n")
	t.root.printDebug(0)
	fmt.Print("</trie>\n")
}

// Given a path and an http method, return all the matching routes.
func (t *Trie) FindRoutes(httpMethod, path string) []*Match {
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
	t.root.find(httpMethod, path, context)
	return matches
}

// Same as FindRoutes, but return in addition a boolean indicating if the path was matched.
// Useful to return 405
func (t *Trie) FindRoutesAndPathMatched(httpMethod, path string) ([]*Match, bool) {
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
	t.root.find(httpMethod, path, context)
	return matches, pathMatched
}

// Given a path, and whatever the http method, return all the matching routes.
func (t *Trie) FindRoutesForPath(path string) []*Match {
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
	t.root.find("", path, context)
	return matches
}
