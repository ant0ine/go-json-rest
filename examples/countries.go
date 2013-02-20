package main

import (
	"github.com/ant0ine/go-json-rest"
	"net/http"
)

// The Curl Demo:
// curl -i -d '{"Code":"FR","Name":"France"}' http://127.0.0.1:8080/countries
// curl -i -d '{"Code":"US","Name":"United States"}' http://127.0.0.1:8080/countries
// curl -i http://127.0.0.1:8080/countries/FR
// curl -i http://127.0.0.1:8080/countries/US
// curl -i http://127.0.0.1:8080/countries
// curl -i -X DELETE http://127.0.0.1:8080/countries/FR
// curl -i http://127.0.0.1:8080/countries
// curl -i -X DELETE http://127.0.0.1:8080/countries/US
// curl -i http://127.0.0.1:8080/countries

type Country struct {
	Code string
	Name string
}

var store = map[string]*Country{}

func GetCountry(w *rest.ResponseWriter, r *rest.Request) {
	code := r.PathParam("code")
	country := store[code]
	if country == nil {
		http.NotFound(w, r.Request)
		return
	}
	w.WriteJson(&country)
}

func GetAllCountries(w *rest.ResponseWriter, r *rest.Request) {
	countries := []*Country{}
	for _, country := range store {
		countries = append(countries, country)
	}
	w.WriteJson(&countries)
}

func PostCountry(w *rest.ResponseWriter, r *rest.Request) {
	country := Country{}
	err := r.DecodeJsonPayload(&country)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if country.Code == "" {
		http.Error(w, "country code required", 400)
		return
	}
	if country.Name == "" {
		http.Error(w, "country name required", 400)
		return
	}
	store[country.Code] = &country
	w.WriteJson(&country)
}

func DeleteCountry(w *rest.ResponseWriter, r *rest.Request) {
	code := r.PathParam("code")
	delete(store, code)
}

func main() {

	handler := rest.ResourceHandler{}
	handler.SetRoutes(
		rest.Route{"GET", "/countries", GetAllCountries},
		rest.Route{"POST", "/countries", PostCountry},
		rest.Route{"GET", "/countries/:code", GetCountry},
		rest.Route{"DELETE", "/countries/:code", DeleteCountry},
	)
	http.ListenAndServe(":8080", &handler)
}
