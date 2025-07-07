package matchmake

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// Router sets up the routes for the matchmake service
func Router() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/companyattributes", companyAttributesHandler).Methods("POST")
	return router
}

// Input represents the incoming request structure
type Input struct {
	GLID          string `json:"glid"`
	PAN           string `json:"pan"`
	Modid         string `json:"modid"`
	ValidationKey string `json:"validationkey"`
}

// Response represents the outgoing response structure
type Response struct {
	Code         int         `json:"code"`
	Status       string      `json:"status"`
	Data         interface{} `json:"data,omitempty"`
	ErrorMessage string      `json:"ErrorMessage,omitempty"`
}

// companyAttributesHandler handles the dummy companyattributes endpoint
func companyAttributesHandler(w http.ResponseWriter, r *http.Request) {
	var in Input
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Check hardcoded modid and validationkey
	if in.Modid != "soa-users" || in.ValidationKey != "bWVycF9zY3JlWER=" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Response{
			Code:         401,
			Status:       "FAILURE",
			ErrorMessage: "Invalid modid or validationkey",
		})
		return
	}

	// Hardcoded dummy responses
	if in.GLID == "24736" && in.PAN == "ABCPD1234E" {
		resp := Response{
			Code:   200,
			Status: "success",
			Data: map[string]interface{}{
				"flag":              "AUTO APPROVED",
				"GLID":              in.GLID,
				"PAN":               in.PAN,
				"Matched_Attribute": map[string]string{"112": "7981642772"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	if in.GLID == "24737" && in.PAN == "ABCPD1234E" {
		resp := Response{
			Code:   200,
			Status: "success",
			Data: map[string]interface{}{
				"flag":              "AUTO REJECTED",
				"GLID":              in.GLID,
				"PAN":               in.PAN,
				"Matched_Attribute": map[string]string{},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// All other input: 400 Failure
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(Response{
		Code:         400,
		Status:       "FAILURE",
		ErrorMessage: "No Data Found",
	})
}
