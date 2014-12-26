package rest

import (
	"fmt"
	"net/url"
	"regexp"
	"testing"
)

func routes() []*Route {
	// simulate the routes of a real but reasonable app.
	// 6 + 10 * (5 + 2) + 1 = 77 routes
	routePaths := []string{
		"/",
		"/signin",
		"/signout",
		"/profile",
		"/settings",
		"/upload/*file",
	}
	for i := 0; i < 10; i++ {
		for j := 0; j < 5; j++ {
			routePaths = append(routePaths, fmt.Sprintf("/resource%d/:id/property%d", i, j))
		}
		routePaths = append(routePaths, fmt.Sprintf("/resource%d/:id", i))
		routePaths = append(routePaths, fmt.Sprintf("/resource%d", i))
	}
	routePaths = append(routePaths, "/*")

	routes := []*Route{}
	for _, path := range routePaths {
		routes = append(routes, &Route{
			HttpMethod: "GET",
			PathExp:    path,
		})
	}
	return routes
}

func requestUrls() []*url.URL {
	// simulate a few requests
	urlStrs := []string{
		"http://example.org/",
		"http://example.org/resource9/123",
		"http://example.org/resource9/123/property1",
		"http://example.org/doesnotexist",
	}
	urlObjs := []*url.URL{}
	for _, urlStr := range urlStrs {
		urlObj, _ := url.Parse(urlStr)
		urlObjs = append(urlObjs, urlObj)
	}
	return urlObjs
}

func BenchmarkNoCompression(b *testing.B) {

	b.StopTimer()

	r := router{
		Routes:                 routes(),
		disableTrieCompression: true,
	}
	r.start()
	urlObjs := requestUrls()

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		for _, urlObj := range urlObjs {
			r.findRouteFromURL("GET", urlObj)
		}
	}
}

func BenchmarkCompression(b *testing.B) {

	b.StopTimer()

	r := router{
		Routes: routes(),
	}
	r.start()
	urlObjs := requestUrls()

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		for _, urlObj := range urlObjs {
			r.findRouteFromURL("GET", urlObj)
		}
	}
}

func BenchmarkRegExpLoop(b *testing.B) {
	// reference benchmark using the usual RegExps + Loop strategy

	b.StopTimer()

	routes := routes()
	urlObjs := requestUrls()

	// build the route regexps
	r1, err := regexp.Compile(":[^/\\.]*")
	if err != nil {
		panic(err)
	}
	r2, err := regexp.Compile("\\*.*")
	if err != nil {
		panic(err)
	}
	routeRegexps := []*regexp.Regexp{}
	for _, route := range routes {

		// generate the regexp string
		regStr := r2.ReplaceAllString(route.PathExp, "([^/\\.]+)")
		regStr = r1.ReplaceAllString(regStr, "(.+)")
		regStr = "^" + regStr + "$"

		// compile it
		reg, err := regexp.Compile(regStr)
		if err != nil {
			panic(err)
		}

		routeRegexps = append(routeRegexps, reg)
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		// do it for a few urls
		for _, urlObj := range urlObjs {
			// stop at the first route that matches
			for index, reg := range routeRegexps {
				if reg.FindAllString(urlObj.Path, 1) != nil {
					_ = routes[index]
					break
				}
			}
		}
	}
}
