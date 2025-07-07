package gstvalidator

import (
	"github.com/gorilla/mux"
)

//Router ...
func Router() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	routes := router.PathPrefix("/v1").Subrouter()

	routesFun(routes)

	return routes
}

func routesFun(routes *mux.Router) *mux.Router {

	routes.HandleFunc("/gst", GstData).Methods("POSt")

	return routes
}


