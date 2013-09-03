/* Demonstrate a possible CORS validation logic

The Curl Demo:

        curl -i http://127.0.0.1:8080/countries

*/
package main

import (
	"github.com/ant0ine/go-json-rest"
	"net/http"
)

func main() {

	handler := rest.ResourceHandler{
		PreRoutingMiddleware: func(handler rest.HandleFunc) rest.HandleFunc {
			return func(writer *rest.ResponseWriter, request *rest.Request) {

				corsInfo := request.GetCorsInfo()

				// Be nice with non CORS requests, continue
				// Alternatively, you may also chose to only allow CORS requests, and return an error.
				if !corsInfo.IsCors {
					// continure, execute the wrapped middleware
					handler(writer, request)
					return
				}

				// Validate the Origin
				// More sophisticated validations can be implemented, regexps, DB lookups, ...
				if corsInfo.Origin != "http://my.other.host" {
					rest.Error(writer, "Invalid Origin", http.StatusForbidden)
					return
				}

				if corsInfo.IsPreflight {
					// check the request methods
					allowedMethods := map[string]bool{
						"GET":  true,
						"POST": true,
						"PUT":  true,
						// don't allow DELETE, for instance
					}
					if !allowedMethods[corsInfo.AccessControlRequestMethod] {
						rest.Error(writer, "Invalid Preflight Request", http.StatusForbidden)
						return
					}
					// check the request headers
					allowedHeaders := map[string]bool{
						"X-Custom-Header": true,
					}
					for _, requestedHeader := range corsInfo.AccessControlRequestHeaders {
						if !allowedHeaders[requestedHeader] {
							rest.Error(writer, "Invalid Preflight Request", http.StatusForbidden)
							return
						}
					}

					for allowedMethod, _ := range allowedMethods {
						writer.Header().Add("Access-Control-Allow-Methods", allowedMethod)
					}
					for allowedHeader, _ := range allowedHeaders {
						writer.Header().Add("Access-Control-Allow-Headers", allowedHeader)
					}
					writer.Header().Set("Access-Control-Allow-Origin", corsInfo.Origin)
					writer.Header().Set("Access-Control-Allow-Credentials", "true")
					writer.Header().Set("Access-Control-Max-Age", "3600")
					writer.WriteHeader(http.StatusOK)
					return
				} else {
					writer.Header().Set("Access-Control-Expose-Headers", "X-Powered-By")
					writer.Header().Set("Access-Control-Allow-Origin", corsInfo.Origin)
					writer.Header().Set("Access-Control-Allow-Credentials", "true")
					// continure, execute the wrapped middleware
					handler(writer, request)
					return
				}
			}
		},
	}
	handler.SetRoutes(
		rest.Route{"GET", "/countries", GetAllCountries},
	)
	http.ListenAndServe(":8080", &handler)
}

type Country struct {
	Code string
	Name string
}

func GetAllCountries(w *rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(
		[]Country{
			Country{
				Code: "FR",
				Name: "France",
			},
			Country{
				Code: "US",
				Name: "United States",
			},
		},
	)
}
