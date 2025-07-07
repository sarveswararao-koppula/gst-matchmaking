package server

import (
	"log"
	prop "mm/properties"
	"mm/routes"
	"net/http"
)

//StartServer ...
func StartServer() {

	//set up routes
	router := routes.SetUpRoutes()

	log.Fatal(http.ListenAndServe(prop.Prop.PORT, router))
}
