package routes

import (
	"mm/services/authbridgeadvanced"
	"mm/services/gstchallandata"
	"mm/services/gstdata"
	"mm/services/gstmmcontrols"
	"mm/services/gstturnoverdetails"
	"mm/services/gstvalidator"
	"mm/services/livebefisccontrols"
	"mm/services/livemasterindiacontrols"
	"mm/services/livematchmakingapi"
	"mm/services/masterindiacontrols"
	"mm/services/pingcontrols"
	"mm/services/tan_verification"
	"mm/services/truthscreen"
	"mm/subscribers"
	"net/http"
	"strings"
	"mm/services/matchmake"

	"github.com/gorilla/mux"
)

// SetUpRoutes ...
func SetUpRoutes() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)

	mount(router, "/gstmm", gstmmcontrols.Router())
	mount(router, "/masterindia", masterindiacontrols.Router())
	mount(router, "/matchmaking", livematchmakingapi.Router())
	mount(router, "/gstdata", gstdata.Router())
	mount(router, "/gstchallandata", gstchallandata.Router())
	mount(router, "/authadvanced", authbridgeadvanced.Router())
	mount(router, "/tan_verification", tan_verification.Router())
	mount(router, "/gstvalidator", gstvalidator.Router())
	mount(router, "/gst-turnover-details", gstturnoverdetails.Router())
	mount(router, "/realtimemasterindia", livemasterindiacontrols.Router())
	mount(router, "/befisc", livebefisccontrols.Router())
	mount(router, "/truthscreen", truthscreen.Router())
	mount(router, "/ping", pingcontrols.Router())
	mount(router, "/matchmake", matchmake.Router())
	subscribers.StartDispatcher(5, 5)
	return router
}

func mount(r *mux.Router, path string, handler http.Handler) {
	r.PathPrefix(path).Handler(
		http.StripPrefix(
			strings.TrimSuffix(path, "/"),
			handler,
		),
	)
}
