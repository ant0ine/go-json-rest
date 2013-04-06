package rest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

// Inherit from http.Request, and provide additional methods.
type Request struct {
	*http.Request
	// map of parameters that have been matched in the URL Path.
	PathParams map[string]string
}

// Provide a convenient access to the PathParams map
func (self *Request) PathParam(name string) string {
	return self.PathParams[name]
}

// Read the request body and decode the JSON using json.Unmarshal
func (self *Request) DecodeJsonPayload(v interface{}) error {
	content, err := ioutil.ReadAll(self.Body)
	self.Body.Close()
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, v)
	if err != nil {
		return err
	}
	return nil
}

// Returns an absolute URI for the base (scheme + host) of the application,
// without the trailing slash.
func (self *Request) UriBase() string {
	scheme := self.URL.Scheme
	if scheme == "" {
		scheme = "http"
	}

	url := fmt.Sprintf("%s://%s", scheme, self.Host)

	trailingSlash, _ := regexp.Compile("/$")
	if trailingSlash.MatchString(url) == true {
		url = trailingSlash.ReplaceAllString(url, "")
	}
	return url
}

// Returns an URI from the base and path.
func (self *Request) UriFor(path string) string {
	url := fmt.Sprintf("%s%s", self.UriBase(), path)
	return url
}

// Returns an URI from the base, the path, and the parameters.
func (self *Request) UriForWithParams(path string, parameters map[string]string) string {
	query := url.Values{}
	for k, v := range parameters {
		query.Add(k, v)
	}
	url := fmt.Sprintf("%s?%s", self.UriFor(path), query.Encode())
	return url
}
