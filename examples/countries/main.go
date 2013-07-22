/* Demonstrate simple POST GET and DELETE operations

The Curl Demo:

        curl -i -d '{"Code":"FR","Name":"France"}' http://127.0.0.1:8080/countries
        curl -i -d '{"Code":"US","Name":"United States"}' http://127.0.0.1:8080/countries
        curl -i http://127.0.0.1:8080/countries/FR
        curl -i http://127.0.0.1:8080/countries/US
        curl -i http://127.0.0.1:8080/countries
        curl -i -X DELETE http://127.0.0.1:8080/countries/FR
        curl -i http://127.0.0.1:8080/countries
        curl -i -X DELETE http://127.0.0.1:8080/countries/US
        curl -i http://127.0.0.1:8080/countries

*/
package main

import (
	"github.com/ant0ine/go-json-rest"
	"net/http"
)

func main() {

	handler := rest.ResourceHandler{
                EnableRelaxedContentType: true,
        }
	handler.SetRoutes(
		rest.Route{"GET", "/countries", GetAllCountries},
		rest.Route{"POST", "/countries", PostCountry},
		rest.Route{"GET", "/countries/:code", GetCountry},
		rest.Route{"DELETE", "/countries/:code", DeleteCountry},
	)
	http.ListenAndServe(":8080", &handler)
}

type Country struct {
	Code string
	Name string
}

var store = map[string]*Country{}

func GetCountry(w *rest.ResponseWriter, r *rest.Request) {
	code := r.PathParam("code")
	country := store[code]
	if country == nil {
		rest.NotFound(w, r)
		return
	}
	w.WriteJson(&country)
}

func GetAllCountries(w *rest.ResponseWriter, r *rest.Request) {
	countries := make([]*Country, len(store))
	i := 0
	for _, country := range store {
		countries[i] = country
		i++
	}
	w.WriteJson(&countries)
}

func PostCountry(w *rest.ResponseWriter, r *rest.Request) {
	country := Country{}
	err := r.DecodeJsonPayload(&country)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if country.Code == "" {
		rest.Error(w, "country code required", 400)
		return
	}
	if country.Name == "" {
		rest.Error(w, "country name required", 400)
		return
	}
	store[country.Code] = &country
	w.WriteJson(&country)
}

func DeleteCountry(w *rest.ResponseWriter, r *rest.Request) {
	code := r.PathParam("code")
	delete(store, code)
}
