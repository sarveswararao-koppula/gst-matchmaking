package masterindiacontrols

import (
	"github.com/gorilla/mux"
)

//Router ...
func Router() *mux.Router {

	masterIndiaRouter := mux.NewRouter().StrictSlash(true)
	masterIndiaRoutes := masterIndiaRouter.PathPrefix("/v1").Subrouter()

	masterindiaRoutesFunc(masterIndiaRoutes)

	return masterIndiaRoutes
}

func masterindiaRoutesFunc(masterIndiaRoutes *mux.Router) *mux.Router {

	masterIndiaRoutes.HandleFunc("/gst", GetGSTData).Methods("POST")

	return masterIndiaRoutes
}
