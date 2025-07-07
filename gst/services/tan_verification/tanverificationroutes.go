package tan_verification

import (
	"github.com/gorilla/mux"
)

//Router ...
func Router() *mux.Router {

	tanverificationRouter := mux.NewRouter().StrictSlash(true)
	tanverificationRoutes := tanverificationRouter.PathPrefix("/v1").Subrouter()

	tanverificationRoutesFunc(tanverificationRoutes)

	return tanverificationRoutes
}

func tanverificationRoutesFunc(tanverificationRoutes *mux.Router) *mux.Router {

	tanverificationRoutes.HandleFunc("/tan", GetTANData).Methods("POST")

	return tanverificationRoutes
}
