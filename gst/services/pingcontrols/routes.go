// services/pingcontrols/routes.go
package pingcontrols

import (
	"github.com/gorilla/mux"
)

// Router sets up the health check routes
func Router() *mux.Router {
	pingRouter := mux.NewRouter().StrictSlash(true)
	pingRoutes := pingRouter.PathPrefix("/v1").Subrouter()

	pingRoutes.HandleFunc("/gst", GetPingResponse).Methods("GET")

	return pingRouter
}
