package livematchmakingapi

import (
	"github.com/gorilla/mux"
)

//Router ...
func Router() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	routes := router.PathPrefix("/v1").Subrouter()

	routesFunc(routes)

	return routes
}

func routesFunc(routes *mux.Router) *mux.Router {

	routes.HandleFunc("/gst", Match).Methods("POST")

	return routes
}
