package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var Credentials map[string]map[string]string = map[string]map[string]string{

	"Credentials3": map[string]string{
		"username":      "Credentials3@indiamart.com",
		"password":      "Indiamart3@123",
		"client_id":     "fPDwunwKcasiWxWbDK",
		"client_secret": "HidLFvlAwk0SfemISenmMtoP",
		"grant_type":    "password",
	},

	"Credentials4": map[string]string{
		"username":      "Credentials4@indiamart.com",
		"password":      "Indiamart4@123",
		"client_id":     "PyAbrpmyImVPJXglcX",
		"client_secret": "93SsSh78ZpOVHhmYJyedAIJz",
		"grant_type":    "password",
	},

	"Credentials5": map[string]string{
		"username":      "Credentials5@indiamart.com",
		"password":      "Indiamart@123",
		"client_id":     "JRTFWlDwqbWipzLAOs",
		"client_secret": "hJki7yZ3QNkmXG6J2YX0mefw",
		"grant_type":    "password",
	},

	"credentials6": map[string]string{
		"username":      "credentials6@indiamart.com",
		"password":      "Cred6@123",
		"client_id":     "RZlPkGFZhlafuHDnbu",
		"client_secret": "xRCnv8k4Xyq2JnIDkWZ4fHDV",
		"grant_type":    "password",
	},

	"credentials8": map[string]string{
		"username":      "credentials8@indiamart.com",
		"password":      "Cred8@123",
		"client_id":     "dGzavdFhnNxAFYkIVY",
		"client_secret": "uJxMcSwA94tzbX3vERHhvm1W",
		"grant_type":    "password",
	},

	"amrita": map[string]string{
		"username":      "amrita@indiamart.com",
		"password":      "Amrita@123",
		"client_id":     "CembgvxJXHJzaXxPoP",
		"client_secret": "TIa5mAsga3AJtqVfsAC3F5Ay",
		"grant_type":    "password",
	},

	"sachin": map[string]string{
		"username":      "sachin63596@indiamart.com",
		"password":      "Sachin@123",
		"client_id":     "YXAMKIxwcWntMmImRF",
		"client_secret": "pnU5yh0enmKGB6OLgdV23070",
		"grant_type":    "password",
	},

	"kumar": map[string]string{
		"username":      "kumar.rahul2@indiamart.com",
		"password":      "Rahul@123",
		"client_id":     "oqxejOmsbPlMDKmkIV",
		"client_secret": "uebAHUPSZOvRJx9kcQsAfaZu",
		"grant_type":    "password",
	},

	"puneetkochale": map[string]string{
		"username":      "puneetkochale@indiamart.com",
		"password":      "Indiamart@123",
		"client_id":     "YRCFiWjBmhUkRvvdqt",
		"client_secret": "va7YMU7StANSZH8qMAcoPP18",
		"grant_type":    "password",
	},

	"vishnu": map[string]string{
		"username":      "vishnu.kumar@indiamart.com",
		"password":      "Vishnu@123",
		"client_id":     "SBnKDsjovFbULErSAI",
		"client_secret": "7cMtAXtOyomyrc7friRfr2l7",
		"grant_type":    "password",
	},

	"amisha": map[string]string{
		"username":      "amisha.tomar@indiamart.com",
		"password":      "Amisha@123",
		"client_id":     "AXuoVtlEXuSDKHhieY",
		"client_secret": "oFcd86belE1ZdvWAmeIGhExO",
		"grant_type":    "password",
	},

	"dharmesh": map[string]string{
		"username":      "dharmesh.meena@indiamart.com",
		"password":      "Dharmesh@123",
		"client_id":     "IpoCOmTZtOjfzIovWX",
		"client_secret": "PVgib5SlYk8wDskQRKmHLKxo",
		"grant_type":    "password",
	},

	"puneetsingh": map[string]string{
		"username":      "puneetsingh@indiamart.com",
		"password":      "Puneet@123",
		"client_id":     "QLlBzTUOzMpSGXwbMK",
		"client_secret": "sEQx7MXAzMbjfLdwExLOCEpF",
		"grant_type":    "password",
	},

	"credentials7": map[string]string{
		"username":      "credentials7@indiamart.com",
		"password":      "Cred7@123",
		"client_id":     "hKTJUGrWTVDTXioKVc",
		"client_secret": "HhT6bQJ6djiUoUj4Ic6CEskH",
		"grant_type":    "password",
	},
	"vivek": map[string]string{
		"username":      "vivek.arya@indiamart.com",
		"password":      "Vivek@123",
		"client_id":     "tMdsdkRkrgOlcaeUNl",
		"client_secret": "NiKHivn7P2pJqnabG4Swx4Fq",
		"grant_type":    "password",
	},
	"Credentials15": map[string]string{
		"username":      "Credentials15@indiamart.com",
		"password":      "India@123",
		"client_id":     "pwabdpXGWYYDsQlDZD",
		"client_secret": "JAV5Rgw4UdT5AOFdyBzkObG0",
		"grant_type":    "password",
	},
	"Gladminbefisc": map[string]string{
		"authKey": "TNPWR37HIMHVZTJ",
	},
}

func GetCred(user string) map[string]string {

	return Credentials[user]
}

// func BusLogicOnMasterData(gstin_number string, m map[string]interface{}) []interface{} {
// 	var (
// 		business_name, centre_juri, registration_date, cancel_date, business_constitution, business_activity_nature,
// 		gstin_status, last_update_date, state_jurisdiction_code, state_juri, centre_jurisdiction_code, trade_name, bussiness_fields_add string
// 		location, state_name, pin, taxpayer_type, building_name, street, door_number, floor_number    string
// 		longitude, lattitude, bussiness_place_add_nature, bussiness_fields_pp, building_name_addl     string
// 		street_addl, location_addl, door_number_addl, state_name_addl, floor_number_addl              string
// 		longitude_addl, lattitude_addl, pincode_addl_str, nature_of_business_addl, gst_insertion_date string
// 		bussiness_fields_add_district                                                                 string
// 		bussiness_fields_pp_district                                                                  string
// 	)

// 	business_name, _ = m["lgnm"].(string)
// 	centre_juri, _ = m["ctj"].(string)
// 	registration_date, _ = m["rgdt"].(string)
// 	cancel_date, _ = m["cxdt"].(string)
// 	business_constitution, _ = m["ctb"].(string)
// 	gstin_status, _ = m["sts"].(string)
// 	last_update_date, _ = m["lstupdt"].(string)
// 	state_jurisdiction_code, _ = m["stjCd"].(string)
// 	state_juri, _ = m["stj"].(string)
// 	centre_jurisdiction_code, _ = m["ctjCd"].(string)
// 	trade_name, _ = m["tradeNam"].(string)
// 	taxpayer_type, _ = m["dty"].(string)

// 	pradr, ok := m["pradr"].(map[string]interface{})
// 	if ok {
// 		business_activity_nature, _ = pradr["ntr"].(string)
// 		bussiness_place_add_nature, _ = pradr["ntr"].(string)
// 	}
// 	addr, ok := pradr["addr"].(map[string]interface{})
// 	if ok {
// 		building_name, _ = addr["bnm"].(string)
// 		street, _ = addr["st"].(string)
// 		door_number, _ = addr["bno"].(string)
// 		floor_number, _ = addr["flno"].(string)
// 		lattitude, _ = addr["lt"].(string)
// 		longitude, _ = addr["lg"].(string)
// 		location, _ = addr["loc"].(string)
// 		state_name, _ = addr["stcd"].(string)
// 		pin, _ = addr["pncd"].(string)
// 		city, _ := addr["city"].(string)
// 		dst, _ := addr["dst"].(string)
// 		bussiness_fields_add = floor_number + "," + door_number + "," + street + "," + building_name + "," + location + "," + city + "," + dst + "," + state_name + "," + pin
// 		bussiness_fields_add_district = dst
// 	}

// 	adadrArr, ok := m["adadr"].([]map[string]interface{})
// 	if ok {
// 		adadr := adadrArr[0]
// 		addr, ok := adadr["addr"].(map[string]interface{})
// 		if ok {
// 			building_name_addl, _ = addr["bnm"].(string)
// 			door_number_addl, _ = addr["bno"].(string)
// 			dst_add1, _ := addr["dst"].(string)
// 			floor_number_addl, _ = addr["flno"].(string)
// 			location_addl, _ = addr["loc"].(string)
// 			lattitude_addl, _ = addr["lt"].(string)
// 			longitude_addl, _ = addr["lg"].(string)
// 			pincode_addl_str, _ = addr["pncd"].(string)
// 			street_addl, _ = addr["st"].(string)
// 			state_name_addl, _ = addr["stcd"].(string)
// 			nature_of_business_addl, _ = addr["ntr"].(string)

// 			bussiness_fields_pp = floor_number_addl + ", " + door_number_addl + "," + street_addl + "," + building_name_addl + "," + dst_add1 + "," + state_name_addl + "," + pincode_addl_str

// 			bussiness_fields_pp_district = dst_add1
// 		}
// 	}

// 	gstin_number = strings.ReplaceAll(gstin_number, "'", "")
// 	business_name = strings.ReplaceAll(business_name, "'", "")
// 	centre_juri = strings.ReplaceAll(centre_juri, "'", "")

// 	registration_date = strings.ReplaceAll(registration_date, "'", "")
// 	regis_date_Date, _ := time.Parse("02/01/2006", registration_date)
// 	registration_date = regis_date_Date.Format("2006-01-02")

// 	cancel_date = strings.ReplaceAll(cancel_date, "'", "")
// 	cancel_date_Date, _ := time.Parse("02/01/2006", cancel_date)
// 	cancel_date = cancel_date_Date.Format("2006-01-02")

// 	business_constitution = strings.ReplaceAll(business_constitution, "'", "")
// 	business_activity_nature = strings.ReplaceAll(business_activity_nature, "'", "")

// 	last_update_date = strings.ReplaceAll(last_update_date, "'", "")
// 	last_up_date_Date, _ := time.Parse("02/01/2006", last_update_date)
// 	last_update_date = last_up_date_Date.Format("2006-01-02")

// 	state_jurisdiction_code = strings.ReplaceAll(state_jurisdiction_code, "'", "")
// 	state_juri = strings.ReplaceAll(state_juri, "'", "")
// 	centre_jurisdiction_code = strings.ReplaceAll(centre_jurisdiction_code, "'", "")
// 	trade_name = strings.ReplaceAll(trade_name, "'", "")

// 	bussiness_fields_add = strings.ReplaceAll(bussiness_fields_add, "'", "")
// 	bussiness_fields_add = strings.ReplaceAll(bussiness_fields_add, ",", " ")
// 	bussiness_fields_add = strings.ReplaceAll(bussiness_fields_add, "-", "")
// 	bussiness_fields_add = strings.Trim(bussiness_fields_add, " ")

// 	location = strings.ReplaceAll(location, "'", "")
// 	state_name = strings.ReplaceAll(state_name, "'", "")
// 	pin = strings.ReplaceAll(pin, "'", "")
// 	taxpayer_type = strings.ReplaceAll(taxpayer_type, "'", "")
// 	building_name = strings.ReplaceAll(building_name, "'", "")
// 	street = strings.ReplaceAll(street, "'", "")
// 	door_number = strings.ReplaceAll(door_number, "'", "")
// 	floor_number = strings.ReplaceAll(floor_number, "'", "")
// 	lattitude = strings.ReplaceAll(lattitude, "'", "")
// 	longitude = strings.ReplaceAll(longitude, "'", "")

// 	bussiness_place_add_nature = strings.ReplaceAll(bussiness_place_add_nature, "'", "")
// 	bussiness_fields_pp = strings.ReplaceAll(bussiness_fields_pp, "'", "")
// 	bussiness_fields_pp = strings.ReplaceAll(bussiness_fields_pp, ",", " ")
// 	bussiness_fields_pp = strings.ReplaceAll(bussiness_fields_pp, "-", "")
// 	bussiness_fields_pp = strings.Trim(bussiness_fields_pp, " ")

// 	building_name_addl = strings.ReplaceAll(building_name_addl, "'", "")
// 	street_addl = strings.ReplaceAll(street_addl, "'", "")
// 	location_addl = strings.ReplaceAll(location_addl, "'", "")
// 	door_number_addl = strings.ReplaceAll(door_number_addl, "'", "")
// 	state_name_addl = strings.ReplaceAll(state_name_addl, "'", "")
// 	floor_number_addl = strings.ReplaceAll(floor_number_addl, "'", "")
// 	lattitude_addl = strings.ReplaceAll(lattitude_addl, "'", "")
// 	longitude_addl = strings.ReplaceAll(longitude_addl, "'", "")
// 	pincode_addl_str = strings.ReplaceAll(pincode_addl_str, "'", "")

// 	nature_of_business_addl = strings.ReplaceAll(nature_of_business_addl, "'", "")

// 	gst_insertion_date = time.Now().Format("2006-01-02 15:04:05")

// 	pincode, _ := strconv.Atoi(pin)
// 	pincode_addl, _ := strconv.Atoi(pincode_addl_str)
// 	gst_inserted_by := 111

// 	var params []interface{}

// 	params = append(params, gstin_number, business_name, centre_juri, registration_date,
// 		cancel_date, business_constitution, business_activity_nature, gstin_status, last_update_date,
// 		state_jurisdiction_code, state_juri, centre_jurisdiction_code, trade_name,
// 		bussiness_fields_add, location, state_name, pincode, taxpayer_type, building_name, street,
// 		door_number, floor_number, longitude, lattitude, bussiness_place_add_nature,
// 		bussiness_fields_pp, building_name_addl, street_addl, location_addl, door_number_addl,
// 		state_name_addl, floor_number_addl, longitude_addl, lattitude_addl, pincode_addl,
// 		nature_of_business_addl, gst_insertion_date, gst_inserted_by,
// 		bussiness_fields_add_district, bussiness_fields_pp_district)

// 	return params
// }

func BusLogicOnMasterData_V2(gstin_number string, m map[string]interface{}) (map[string]string, []interface{}) {

	var (
		business_name, centre_juri, registration_date, cancel_date, business_constitution, business_activity_nature,
		gstin_status, last_update_date, state_jurisdiction_code, state_juri, centre_jurisdiction_code, trade_name, bussiness_fields_add string
		location, state_name, pin, taxpayer_type, building_name, street, door_number, floor_number    string
		longitude, lattitude, bussiness_place_add_nature, bussiness_address_add, building_name_addl   string
		street_addl, location_addl, door_number_addl, state_name_addl, floor_number_addl              string
		longitude_addl, lattitude_addl, pincode_addl_str, nature_of_business_addl, gst_insertion_date string
		bussiness_fields_add_district                                                                 string
		bussiness_fields_pp_district                                                                  string
		business_constitution_group_id                                                                int
		landmark, locality,  geo_code_lvl , einvoice_status                                           string
	)

	business_name, _ = m["lgnm"].(string)
	centre_juri, _ = m["ctj"].(string)
	registration_date, _ = m["rgdt"].(string)
	cancel_date, _ = m["cxdt"].(string)
	business_constitution, _ = m["ctb"].(string)
	gstin_status, _ = m["sts"].(string)
	last_update_date, _ = m["lstupdt"].(string)
	state_jurisdiction_code, _ = m["stjCd"].(string)
	state_juri, _ = m["stj"].(string)
	centre_jurisdiction_code, _ = m["ctjCd"].(string)
	trade_name, _ = m["tradeNam"].(string)
	taxpayer_type, _ = m["dty"].(string)

	einvoice_status_v, einvoice_status_exists := m["einvoiceStatus"].(string)
		if einvoice_status_exists{
			einvoice_status=einvoice_status_v
		}else{
			einvoice_status=""
		}

	var nba string
	buisnessactivitynature_new, exists := m["nba"]

	// Check if "business_activity_nature" exists and handle different types
	if exists {
		switch v := buisnessactivitynature_new.(type) {
		case string:
			nba = v
		case []interface{}:
			var names []string
			for _, name := range v {
				if str, ok := name.(string); ok {
					names = append(names, str)
				}
			}
			nba = strings.Join(names, ",")
		default:
			// Handle unexpected types
			nba = ""
		}
	} else {
		nba = ""
	}

	business_activity_nature = nba

	pradr, ok := m["pradr"].(map[string]interface{})
	if ok {
		// business_activity_nature, _ = pradr["ntr"].(string)
		bussiness_place_add_nature, _ = pradr["ntr"].(string)
	}
	addr, ok := pradr["addr"].(map[string]interface{})
	if ok {
		building_name, _ = addr["bnm"].(string)
		street, _ = addr["st"].(string)
		door_number, _ = addr["bno"].(string)
		floor_number, _ = addr["flno"].(string)
		lattitude, _ = addr["lt"].(string)
		longitude, _ = addr["lg"].(string)
		location, _ = addr["loc"].(string)
		state_name, _ = addr["stcd"].(string)
		pin, _ = addr["pncd"].(string)
		city, _ := addr["city"].(string)
		dst, _ := addr["dst"].(string)
		// bussiness_fields_add = floor_number + " " + door_number + " " + street + " " + building_name + " " + location + " " + city + " " + dst + " " + state_name + " " + pin
		bussiness_fields_add_district = dst

		locality_v, loc_exists := addr["locality"].(string)
		if loc_exists{
			locality=locality_v
		}else{
			locality=""
		}

		landmark_v, landmark_exists := addr["landMark"].(string)
		if landmark_exists{
			landmark=landmark_v
		}else{
			landmark=""
		}

		geo_code_lvl_v, geo_code_lvl_exists := addr["geocodelvl"].(string)
		if geo_code_lvl_exists{
			geo_code_lvl=geo_code_lvl_v
		}else{
			geo_code_lvl=""
		}

        if landmark !="" && locality !="" {
			bussiness_fields_add = floor_number + " " + door_number + " " + building_name + " " + street + " " + landmark + " " + locality + " "  + location + " " + city + " " + dst + " " + state_name + " " + pin
		} else if landmark !="" && locality =="" {
			bussiness_fields_add = floor_number + " " + door_number + " " + building_name + " " + street + " " + landmark + " " + location + " " + city + " " + dst + " " + state_name + " " + pin
		} else if landmark =="" && locality !="" {
			bussiness_fields_add = floor_number + " " + door_number + " " + building_name + " " + street + " " + locality + " "  + location + " " + city + " " + dst + " " + state_name + " " + pin
		} else {
			bussiness_fields_add = floor_number + " " + door_number + " " + building_name + " " + street + " "  + location + " " + city + " " + dst + " " + state_name + " " + pin
		}

	}

	adadrArr, ok := m["adadr"].([]interface{})
	if ok && len(adadrArr) > 0 {

		adadr := adadrArr[0].(map[string]interface{})
		addr, ok := adadr["addr"].(map[string]interface{})
		if ok {
			building_name_addl, _ = addr["bnm"].(string)
			door_number_addl, _ = addr["bno"].(string)
			dst_add1, _ := addr["dst"].(string)
			floor_number_addl, _ = addr["flno"].(string)
			location_addl, _ = addr["loc"].(string)
			lattitude_addl, _ = addr["lt"].(string)
			longitude_addl, _ = addr["lg"].(string)
			pincode_addl_str, _ = addr["pncd"].(string)
			street_addl, _ = addr["st"].(string)
			state_name_addl, _ = addr["stcd"].(string)
			nature_of_business_addl, _ = addr["ntr"].(string)

			bussiness_address_add = floor_number_addl + " " + door_number_addl + " " + street_addl + " " + building_name_addl + " " + dst_add1 + " " + state_name_addl + " " + pincode_addl_str

			bussiness_fields_pp_district = dst_add1
		}
	}

	gst_inserted_by := 111

	res := make(map[string]string)

	res["gstin_number"] = gstin_number
	res["business_name"] = business_name
	res["centre_juri"] = centre_juri
	res["registration_date"] = registration_date
	res["cancel_date"] = cancel_date
	res["business_constitution"] = business_constitution
	res["business_activity_nature"] = business_activity_nature
	res["gstin_status"] = gstin_status
	res["last_update_date"] = last_update_date
	res["state_jurisdiction_code"] = state_jurisdiction_code
	res["state_juri"] = state_juri
	res["centre_jurisdiction_code"] = centre_jurisdiction_code
	res["trade_name"] = trade_name
	res["bussiness_fields_add"] = bussiness_fields_add
	res["location"] = location
	res["state_name"] = state_name
	res["pincode"] = pin
	res["taxpayer_type"] = taxpayer_type
	res["building_name"] = building_name
	res["street"] = street
	res["door_number"] = door_number
	res["floor_number"] = floor_number
	res["longitude"] = longitude
	res["lattitude"] = lattitude
	res["bussiness_place_add_nature"] = bussiness_place_add_nature
	res["bussiness_address_add"] = bussiness_address_add
	res["building_name_addl"] = building_name_addl
	res["street_addl"] = street_addl
	res["location_addl"] = location_addl
	res["door_number_addl"] = door_number_addl
	res["state_name_addl"] = state_name_addl
	res["floor_number_addl"] = floor_number_addl
	res["longitude_addl"] = longitude_addl
	res["lattitude_addl"] = lattitude_addl
	res["pincode_addl"] = pincode_addl_str
	res["nature_of_business_addl"] = nature_of_business_addl
	gst_insertion_date = time.Now().Format("2006-01-02 15:04:05")
	res["bussiness_fields_add_district"] = bussiness_fields_add_district
	res["bussiness_fields_pp_district"] = bussiness_fields_pp_district
	res["gst_insertion_date"] = gst_insertion_date
	res["gst_inserted_by"] = strconv.Itoa(gst_inserted_by)

	_, legalStatusID, err := GetLegalStatus(gstin_number)
	if err == nil {
		business_constitution_group_id = legalStatusID
	}

	res["business_constitution_group_id"] = strconv.Itoa(business_constitution_group_id)

	for i, v := range res {
		v = strings.ReplaceAll(v, "'", " ")
		v = strings.ReplaceAll(v, "-", " ")
		v = strings.ReplaceAll(v, ",", " ")
		v = strings.Trim(v, " ")
		res[i] = v
	}

	var registration_date_nil, cancel_date_nil, last_update_date_nil, pincode, pincode_addl interface{}

	regis_date_Date, err := time.Parse("02/01/2006", res["registration_date"])
	if err == nil {
		registration_date_nil = regis_date_Date.Format("2006-01-02")
	}

	cancel_date_Date, err := time.Parse("02/01/2006", res["cancel_date"])
	if err == nil {
		cancel_date_nil = cancel_date_Date.Format("2006-01-02")
	}

	last_up_date_Date, err := time.Parse("02/01/2006", res["last_update_date"])
	if err == nil {
		last_update_date_nil = last_up_date_Date.Format("2006-01-02")
	}

	if res["pincode"] != "" {
		val, err := strconv.Atoi(res["pincode"])
		if err == nil {
			pincode = val
		}
	}
	if res["pincode_addl"] != "" {
		val, err := strconv.Atoi(res["pincode_addl"])
		if err == nil {
			pincode_addl = val
		}
	}

	

	var params []interface{}

	params = append(params, gstin_number, business_name, centre_juri, registration_date_nil,
		cancel_date_nil, business_constitution, business_activity_nature, gstin_status,
		last_update_date_nil, state_jurisdiction_code, state_juri, centre_jurisdiction_code,
		trade_name, bussiness_fields_add, location, state_name, pincode, taxpayer_type,
		building_name,
		street, door_number, floor_number, longitude, lattitude, bussiness_place_add_nature,
		bussiness_address_add, building_name_addl, street_addl, location_addl, door_number_addl,
		state_name_addl, floor_number_addl, longitude_addl, lattitude_addl, pincode_addl,
		nature_of_business_addl,
		gst_insertion_date, gst_inserted_by,
		bussiness_fields_add_district, bussiness_fields_pp_district, business_constitution_group_id, landmark, locality, geo_code_lvl, einvoice_status)

	return res, params
}

func BusLogicOnMasterData_Befisc(gstin_number string, m map[string]interface{}) (map[string]string, []interface{}) {
	var (
		core_business_activity_nature, aggregate_turn_over, business_constitution, gstin_status, business_name, registration_date   string
		state_juri, taxpayer_type, centre_juri, trade_name, gst_challan_email_by_befisc, gst_challan_mobile_by_befisc, gross_income string
		cancel_date, einvoice_status, field_visit_conducted, proprieter_name, gst_refresh_date_advance_challan, gst_insertion_date  string
		business_constitution_group_id                                                                                              int
		business_activity_nature, bussiness_fields_add                                                                              string
	)

	_, legalStatusID, err := GetLegalStatus(gstin_number)
	if err == nil {
		business_constitution_group_id = legalStatusID
	}

	result, ok := m["result"].(map[string]interface{})
	if ok {

		// a,ok1 := result["primary_business_address"].(map[string]interface{})
		// if ok1 {
		// business_activity_nature, _ = a["business_nature"].(string)
		// }

		b, ok2 := result["primary_business_address"].(map[string]interface{})
		if ok2 {
			bussiness_fields_add, _ = b["registered_address"].(string)
		}

		// c,ok2 := result["other_business_address"].(map[string]interface{})
		// if ok2 {
		// 	bussiness_fields_add, _ = c[""].(string)
		// }

		var nba string
		core_business_activity_nature_new, exists := result["business_nature"]

		// Check if "business_activity_nature" exists and handle different types
		if exists {
			switch v := core_business_activity_nature_new.(type) {
			case string:
				nba = v
			case []interface{}:
				var names []string
				for _, name := range v {
					if str, ok := name.(string); ok {
						names = append(names, str)
					}
				}
				nba = strings.Join(names, ",")
			default:
				// Handle unexpected types
				nba = ""
			}
		} else {
			nba = ""
		}

		business_activity_nature = nba

		turnoverslab, _ := result["aggregate_turn_over"].(string)
		turnoverfinancialyear, _ := result["aggregate_turn_over_financial_year"].(string)

		aggregate_turn_over = formatAggregateTurnover(turnoverslab, turnoverfinancialyear)

		business_constitution, _ = result["business_constitution"].(string)
		gstin_status, _ = result["current_registration_status"].(string)
		business_name, _ = result["legal_name"].(string)
		registration_date, _ = result["register_date"].(string)
		registration_date = strings.Trim(registration_date, " ")
		if strings.ToLower(registration_date) == "na" || strings.ToLower(registration_date) == "" {
			registration_date = ""
		}
		state_juri, _ = result["state_jurisdiction"].(string)
		taxpayer_type, _ = result["tax_payer_type"].(string)
		centre_juri, _ = result["central_jurisdiction"].(string)
		trade_name, _ = result["trade_name"].(string)
		gst_challan_email_by_befisc, _ = result["business_email"].(string)
		gst_challan_mobile_by_befisc, _ = result["business_mobile"].(string)
		core_business_activity_nature, _ = result["nature_of_core_business_activity"].(string)
		gross_income, _ = result["gross_total_income"].(string)
		cancel_date, _ = result["register_cancellation_date"].(string)
		cancel_date = strings.Trim(cancel_date, " ")
		if strings.ToLower(cancel_date) == "na" || strings.ToLower(cancel_date) == "" {
			cancel_date = ""
		}
		einvoice_status, _ = result["mandate_e_invoice"].(string)
		field_visit_conducted, _ = result["is_field_visit_conducted"].(string)

		var pname string
		proprieter_name_new, exists := result["authorized_signatory"]

		// Check if "business_activity_nature" exists and handle different types
		if exists {
			switch v := proprieter_name_new.(type) {
			case string:
				pname = v
			case []interface{}:
				var names []string
				for _, name := range v {
					if str, ok := name.(string); ok {
						names = append(names, str)
					}
				}
				pname = strings.Join(names, ",")
			default:
				// Handle unexpected types
				pname = ""
			}
		} else {
			pname = ""
		}

		proprieter_name = pname
	}

	var Field_visit_conducted bool
	if strings.ToLower(field_visit_conducted) == "no" || strings.ToLower(field_visit_conducted) == "na" {
		Field_visit_conducted = false
	} else {
		Field_visit_conducted = true
	}

	gst_refresh_date_advance_challan = time.Now().Format("2006-01-02 15:04:05")

	gst_insertion_date = time.Now().Format("2006-01-02 15:04:05")

	var registration_date_nil, cancel_date_nil interface{}

	regis_date_Date, err := time.Parse("02/01/2006", registration_date)
	if err == nil {
		registration_date_nil = regis_date_Date.Format("2006-01-02")
	}

	cancel_date_Date, err := time.Parse("02/01/2006", cancel_date)
	if err == nil {
		cancel_date_nil = cancel_date_Date.Format("2006-01-02")
	}

	var params []interface{}

	params = append(params, gstin_number, core_business_activity_nature, aggregate_turn_over, business_constitution, gstin_status, business_name, registration_date_nil,
		state_juri, taxpayer_type, centre_juri, trade_name, gst_challan_email_by_befisc, gst_challan_mobile_by_befisc, gross_income,
		cancel_date_nil, einvoice_status, Field_visit_conducted, proprieter_name, gst_refresh_date_advance_challan,
		business_constitution_group_id, business_activity_nature, bussiness_fields_add, gst_insertion_date)

	res := make(map[string]string)

	res["state_jurisdiction"] = state_juri
	res["taxpayer_type"] = taxpayer_type

	return res, params
}

func formatAggregateTurnover(aggregateTurnOver, financialYear string) string {
	return fmt.Sprintf("%s (For FY %s)", aggregateTurnOver, financialYear)
}
