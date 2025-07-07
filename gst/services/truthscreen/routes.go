package truthscreen

import (
	"github.com/gorilla/mux"
)

func Router() *mux.Router {
	gstmmRouter := mux.NewRouter().StrictSlash(true)
	gstmmRoutes := gstmmRouter.PathPrefix("/v1").Subrouter()
	gstmmRoutesFunc(gstmmRoutes)
	return gstmmRoutes
}

func gstmmRoutesFunc(gstmmRoutes *mux.Router) *mux.Router {
	gstmmRoutes.HandleFunc("/gst", GetGSTDetails).Methods("GET")
	return gstmmRoutes
}