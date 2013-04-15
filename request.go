package rest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
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

// Returns a URL structure for the base (scheme + host) of the application,
// without the trailing slash in the host
func (self *Request) UriBase() url.URL {
	scheme := self.URL.Scheme
	if scheme == "" {
		scheme = "http"
	}

	host := self.Host
	if len(host) > 0 && host[len(host)-1] == '/' {
		host = host[:len(host)-1]
	}

	url := url.URL{
		Scheme: scheme,
		Host:   host,
	}
	return url
}

// Returns an URL structure from the base and an additional path.
func (self *Request) UriFor(path string) url.URL {
	baseUrl := self.UriBase()
	baseUrl.Path = path
	return baseUrl
}

// Returns an URL structure from the base, the path and the parameters.
func (self *Request) UriForWithParams(path string, parameters map[string][]string) url.URL {
	query := url.Values{}
	for k, v := range parameters {
		for _, vv := range v {
			query.Add(k, vv)
		}
	}
	baseUrl := self.UriFor(path)
	baseUrl.RawQuery = query.Encode()
	return baseUrl
}
