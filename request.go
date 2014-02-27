package rest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Inherit from http.Request, and provide additional methods.
type Request struct {
	*http.Request
	// map of parameters that have been matched in the URL Path.
	PathParams map[string]string
}

func CToGoString(c []byte) string {
    n := -1
    for i, b := range c {
        if b == 0 {
            break
        }
        n = i
    }
    return string(c[:n+1])
}

// Provide a convenient access to the PathParams map
func (self *Request) PathParam(name string) string {
	return self.PathParams[name]
}

// Read the request body and decode the JSON using json.Unmarshal
func (self *Request) DecodeJsonPayload(v interface{}, decodeBody bool) error {
	content, err := ioutil.ReadAll(self.Body)
	self.Body.Close()
	if err != nil {
		return err
	}
    contentstr := CToGoString(content)
    if decodeBody == true {
        contentstr, _ = url.QueryUnescape(contentstr)
    }
	err = json.Unmarshal([]byte(contentstr), v)
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

// CORS request info derived from a rest.Request.
type CorsInfo struct {
	IsCors                      bool
	IsPreflight                 bool
	Origin                      string
	OriginUrl                   *url.URL
	AccessControlRequestMethod  string
	AccessControlRequestHeaders []string
}

// Derive CorsInfo from Request
func (self *Request) GetCorsInfo() *CorsInfo {

	origin := self.Header.Get("Origin")
	originUrl, err := url.ParseRequestURI(origin)

	isCors := err == nil && origin != "" && self.Host != originUrl.Host

	reqMethod := self.Header.Get("Access-Control-Request-Method")

	reqHeaders := []string{}
	rawReqHeaders := self.Header[http.CanonicalHeaderKey("Access-Control-Request-Headers")]
	for _, rawReqHeader := range rawReqHeaders {
		// net/http does not handle comma delimited headers for us
		for _, reqHeader := range strings.Split(rawReqHeader, ",") {
			reqHeaders = append(reqHeaders, http.CanonicalHeaderKey(strings.TrimSpace(reqHeader)))
		}
	}

	isPreflight := isCors && self.Method == "OPTIONS" && reqMethod != ""

	return &CorsInfo{
		IsCors:                      isCors,
		IsPreflight:                 isPreflight,
		Origin:                      origin,
		OriginUrl:                   originUrl,
		AccessControlRequestMethod:  reqMethod,
		AccessControlRequestHeaders: reqHeaders,
	}
}
