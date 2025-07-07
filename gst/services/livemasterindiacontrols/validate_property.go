package livemasterindiacontrols

import (
	"errors"
)

type property struct {
	validaionKey string
	flag         string
	allowedCols  []string
}

var myProps map[string]property = map[string]property{
	"merpcsd": property{
		validaionKey: "bWVycF9zY3JlZW4=",
		flag:         "",
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
	"merpnsd": property{
		validaionKey: "bWVycF9zY3JlZW4=",
		flag:         "",
		allowedCols: []string{
			"gstin_number",
			"gst_insertion_date",
			"bussiness_fields_add",
			"state_name",
			"pincode",
			"gstin_status",
			"business_name",
			"business_activity_nature",
			"business_constitution",
			"bussiness_address_add",
			"trade_name",
			"location",
			"registration_date",
		},
	},
	"weberp": property{
		validaionKey: "d2ViZXJwX3NjcmVlbg==",
		flag:         "",
		allowedCols: []string{
			"gstin_number",
			"gst_insertion_date",
			"bussiness_fields_add",
			"state_name",
			"pincode",
			"gstin_status",
			"business_name",
			"business_activity_nature",
			"business_constitution",
			"bussiness_address_add",
			"trade_name",
			"location",
			"registration_date",
			"longitude_addl",
			"lattitude_addl",
			"street",
		},
	},
}

// ValidateProp ...
func ValidateProp(modid string, validationkey string, Flag string) ([]string, error) {

	if myProps[modid].validaionKey != validationkey || validationkey == "" || modid == "" {
		return nil, errors.New("Your request is taking longer to complete. Please retry after sometime or please email to bi-support@indiamart.com.")
	}

	if myProps[modid].flag == Flag {
		return myProps[modid].allowedCols, nil
	}
	return nil, errors.New("Your request is taking longer to complete. Please retry after sometime or please email to bi-support@indiamart.com.")
}
