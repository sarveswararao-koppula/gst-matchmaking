package gstvalidator

import (
	"errors"
)

type property struct {
	validaionKey string
	allowedCols  []string
}

var myProps map[string]property = map[string]property{
	"bi": property{
		validaionKey: "Ymlfc2NyZWVu",
		allowedCols: []string{
			"gstin_number",
			"gst_insertion_date",
			"bussiness_fields_add",
			"state_name",
			"pincode",
			"bussiness_fields_add_district",
			"gstin_status",
		},
	},
	"GlAdmin": property{
		validaionKey: "Z2xhZG1pbl9zY3JlZW4=",
		allowedCols: []string{
			"gstin_number",
			"gstin_status",
		},
	},
}

//ValidateProp ...
func ValidateProp(modid string, validationkey string) ([]string, error) {

	if myProps[modid].validaionKey != validationkey || validationkey == "" || modid == "" {
		return nil, errors.New(errNotAuth)
	}

	return myProps[modid].allowedCols, nil
}

