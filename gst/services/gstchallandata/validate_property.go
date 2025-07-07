package gstchallandata

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
			"return_period",
			"date_of_filing",
			"status",
			"entered_on ",
			"return_type",
		},
	},
	"weberp": property{
		validaionKey: "d2ViZXJwX3NjcmVlbg==",
		allowedCols: []string{
			"gstin_number",
			"return_period",
			"date_of_filing",
			"status",
			"entered_on ",
			"return_type",
		},
	},
	"loans2": property{
		validaionKey: "d2ViZXJwX3NjcmVlbg==",
		allowedCols: []string{
			"gstin_number",
			"return_period",
			"date_of_filing",
			"status",
			"entered_on ",
			"return_type",
			"gst_challan_detail_id",
			"is_valid",
			"mode_of_filing",
			"arn_number",
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

