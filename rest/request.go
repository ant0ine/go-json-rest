package rest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Request inherits from http.Request, and provides additional methods.
type Request struct {
	*http.Request

	// Map of parameters that have been matched in the URL Path.
	PathParams map[string]string

	// Environement used by middlewares to communicate.
	Env map[string]interface{}
}

// PathParam provides a convenient access to the PathParams map.
func (r *Request) PathParam(name string) string {
	return r.PathParams[name]
}

// DecodeJsonPayload reads the request body and decodes the JSON using json.Unmarshal.
func (r *Request) DecodeJsonPayload(v interface{}) error {
	content, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, v)
	if err != nil {
		return err
	}
	return nil
}

// BaseUrl returns a new URL object with the Host and Scheme taken from the request.
// (without the trailing slash in the host)
func (r *Request) BaseUrl() *url.URL {
	scheme := r.URL.Scheme
	if scheme == "" {
		scheme = "http"
	}

	host := r.Host
	if len(host) > 0 && host[len(host)-1] == '/' {
		host = host[:len(host)-1]
	}

	return &url.URL{
		Scheme: scheme,
		Host:   host,
	}
}

// UrlFor returns the URL object from UriBase with the Path set to path, and the query
// string built with queryParams.
func (r *Request) UrlFor(path string, queryParams map[string][]string) *url.URL {
	baseUrl := r.BaseUrl()
	baseUrl.Path = path
	if queryParams != nil {
		query := url.Values{}
		for k, v := range queryParams {
			for _, vv := range v {
				query.Add(k, vv)
			}
		}
		baseUrl.RawQuery = query.Encode()
	}
	return baseUrl
}

// CorsInfo contains the CORS request info derived from a rest.Request.
type CorsInfo struct {
	IsCors      bool
	IsPreflight bool
	Origin      string
	OriginUrl   *url.URL

	// The header value is converted to uppercase to avoid common mistakes.
	AccessControlRequestMethod string

	// The header values are normalized with http.CanonicalHeaderKey.
	AccessControlRequestHeaders []string
}

// GetCorsInfo derives CorsInfo from Request.
func (r *Request) GetCorsInfo() *CorsInfo {

	origin := r.Header.Get("Origin")
	originUrl, err := url.ParseRequestURI(origin)

	isCors := (err == nil && origin != "" && r.Host != originUrl.Host) || origin == "null"

	reqMethod := r.Header.Get("Access-Control-Request-Method")

	reqHeaders := []string{}
	rawReqHeaders := r.Header[http.CanonicalHeaderKey("Access-Control-Request-Headers")]
	for _, rawReqHeader := range rawReqHeaders {
		// net/http does not handle comma delimited headers for us
		for _, reqHeader := range strings.Split(rawReqHeader, ",") {
			reqHeaders = append(reqHeaders, http.CanonicalHeaderKey(strings.TrimSpace(reqHeader)))
		}
	}

	isPreflight := isCors && r.Method == "OPTIONS" && reqMethod != ""

	return &CorsInfo{
		IsCors:                      isCors,
		IsPreflight:                 isPreflight,
		Origin:                      origin,
		OriginUrl:                   originUrl,
		AccessControlRequestMethod:  strings.ToUpper(reqMethod),
		AccessControlRequestHeaders: reqHeaders,
	}
}
