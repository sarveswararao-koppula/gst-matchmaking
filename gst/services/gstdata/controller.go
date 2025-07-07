package gstdata

import (
	//"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	servapi "mm/api/servapi"
	"mm/components/constants"
	model "mm/model/masterindiamodel"
	"mm/properties"
	"mm/queue"
	authadvance "mm/services/authbridgeadvanced"
	"mm/services/gstapis/masterindia"
	"mm/utils"
	"net/http"
	"os"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rs/xid"
)

var database string = properties.Prop.DATABASE

type GstGlusrTurnover struct {
	Turnover []struct {
		FY    string `json:"fy"`
		Value string `json:"value"`
	} `json:"turnover"`
	UpdationDate string `json:"updation_date"`
}

// GstData ...
func GstData(w http.ResponseWriter, r *http.Request) {
	var (
		a, b, c, d, e, f, g, h, i, j, k, l, m, n, o, p, q, R, s, t string
	)
	var hsncode string
	a = "Start"
	uniqID := xid.New().String()
	var logg Logg
	logg.RequestStart = utils.GetTimeStampCurrent()
	logg.RequestStartValue = utils.GetTimeInNanoSeconds()
	logg.ServiceName = serviceName
	logg.ServiceURL = r.RequestURI
	logg.AnyError = make(map[string]string)
	logg.ExecTime = make(map[string]float64)

	HittingTime := logg.RequestStart
	Updateflag := 0
	if parts := strings.Split(r.RemoteAddr, ":"); len(parts) == 2 {
		logg.RemoteAddress = parts[0]
	}

	defer func() {
		if panicCheck := recover(); panicCheck != nil {
			stack := string(debug.Stack())
			logg.StackTrace = stack
			sendResponse(uniqID, w, 500, Updateflag, failure, errPanic, nil, Data{}, logg, "", hsncode)
			return
		}
	}()

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&logg.Request)

	gst_no := logg.Request.Gst

	if err != nil {
		sendResponse(uniqID, w, 400, Updateflag, failure, errParam, err, Data{}, logg, logg.Request.Gst, hsncode)
		return
	}
	cols, err := ValidateProp(logg.Request.ModID, logg.Request.Validationkey, logg.Request.Flag)
	//fmt.Println("cols",cols)
	if err != nil {
		sendResponse(uniqID, w, 400, Updateflag, failure, errNotAuth, err, Data{}, logg, logg.Request.Gst, hsncode)
		return
	}

	if logg.Request.ModID == "gst_otp" {

		// 1) authorize and get allowed columns
		_, err := ValidateProp("gst_otp", logg.Request.Validationkey, "")
		if err != nil {
			sendResponseTogstOtp(uniqID, w, 400, failure, errNotAuth, err, []VendorResp{}, logg)
			return
		}

		// 2) gst is mandatory
		gst := logg.Request.Gst
		if gst == "" {
			sendResponseTogstOtp(uniqID, w, 400, failure, "GST is required", nil, []VendorResp{}, logg)
			// sendResponse(uniqID, w, 400, 0, failure, errParam, errors.New("gst is required"), Data{}, logg, "", "")
			return
		}

		// 3) fetch all columns from DB
		data, err := GetGSTRecords(database, gst)
		if err != nil {
			sendResponseTogstOtp(uniqID, w, 400, failure, errFetchDB, err, []VendorResp{}, logg)
			// sendResponse(uniqID, w, 500, 0, failure, errFetchDB, err, Data{}, logg, gst, "")
			return
		}

		// 4) handle zero‐row
		if len(data) == 0 {
			sendResponseTogstOtp(uniqID, w, 400, failure, "No GST record found", nil, []VendorResp{}, logg)
			return
		}

		// 5) normalize nil → ""
		for k, v := range data {
			if v == nil {
				data[k] = ""
			}
		}

		// 5) extract just the four contact fields
		authMobile := data["mobile_number"].(string)
		authEmail := data["email_id"].(string)
		befiscMobile := data["gst_challan_mobile_by_befisc"].(string)
		befiscEmail := data["gst_challan_email_by_befisc"].(string)

		// 6) shape into the requested array

		respData := []VendorResp{
			{Vendor: "Authbridge", Mobile: authMobile, Email: authEmail},
			{Vendor: "Befisc", Mobile: befiscMobile, Email: befiscEmail},
		}

		// 8) respond with only vendor_contacts
		sendResponseTogstOtp(
			uniqID, w,
			200,
			success,
			"",
			nil,
			respData,
			logg,
		)
		return

	}

	if logg.Request.Gst == "" {
		if logg.Request.Glid == "" {
			sendResponse(uniqID, w, 400, Updateflag, failure, errParam, err, Data{}, logg, logg.Request.Gst, hsncode)
			return
		}
		if _, er := strconv.Atoi(logg.Request.Glid); er != nil {
			sendResponse(uniqID, w, 400, Updateflag, failure, errParam, err, Data{}, logg, logg.Request.Gst, hsncode)
			return
		}
	}

	var (
		data map[string]interface{}
	)

	if err != nil {
		sendResponse(uniqID, w, 400, Updateflag, failure, errFetchDB, err, Data{}, logg, logg.Request.Gst, hsncode)
		return
	}

	const apiName = "masterindia"
	if logg.Request.Flag == "2" {
		b = "Second"
		_, st := utils.GetExecTime()
		if logg.Request.Gst != "" {
			c = "Third"
			data, err = GetGSTRecords36(database, logg.Request.Gst)
			// logg.DbBeforeFlag2 = data
		} else {

			//data, err = GetGlidRecords(database, logg.Request.Glid)
			var (
				data_gst map[string]interface{}
			)
			//First Get GST and then go for Get GST Records 36
			data_gst, err = GetGstFromGlid(database, logg.Request.Glid)
			//fmt.Println(data_gst,"Length of MAP")
			if err != nil {
				sendResponse(uniqID, w, 400, Updateflag, failure, errFetchDB, err, Data{}, logg, logg.Request.Gst, hsncode)
				return
			}

			for k, v := range data_gst {
				if v == nil {
					data_gst[k] = ""
				}
			}

			if len(data_gst) == 0 || data_gst["gst"] == "" {
				sendResponse(uniqID, w, 400, Updateflag, failure, "There is no GST mapped to this user", nil, Data{}, logg, logg.Request.Gst, hsncode)
				return
			}

			if data_gst["gst"] != "" {
				d = "Fourth"
				logg.Request.Gst = data_gst["gst"].(string)
			}

			hsncode = CheckAndWriteHSN(logg.Request.Gst)

			data, err = GetGSTRecords36(database, logg.Request.Gst)

			if err != nil {
				sendResponse(uniqID, w, 400, Updateflag, failure, errFetchDB, err, Data{}, logg, logg.Request.Gst, hsncode)
				return
			} else {
				// Get the original value of k1
				DbTurnover, ok_new := data["gst_glusr_turnover"]

				// Check if the key is present and the original value is already a string
				if ok_new {
					DbTurnoverString, ok := DbTurnover.(string)
					if ok {
						// If the original value is a string, Do Format it
						Final := processJSON(DbTurnoverString)

						// Update the value of k1 with the new interface value
						data["annual_turnover_slab"] = Final
					}
				}

				grossincome_db, okk := data["gross_income"]

				if okk {
					grossincome_db_String, ok := grossincome_db.(string)
					if ok {
						// If the original value is a string, Do Format it
						Final := ReplaceBreakWithComma(grossincome_db_String)

						// Update the value of k1 with the new interface value
						data["gross_income"] = Final
					}
				}

				// ReplaceBreakWithComma(

			}
			// logg.DbBeforeFlag2 = data
			data_gst = make(map[string]interface{})

		}
		logg.ExecTime["DB_QUERY"], st = utils.GetExecTime(st)

		APIUserID, err := validateProp(logg.Request.ModID, logg.Request.Validationkey, apiName)

		if err != nil {
			sendResponse(uniqID, w, 400, Updateflag, failure, errValidationKey, err, Data{}, logg, logg.Request.Gst, hsncode)
			return
		}

		gstInsertionDate := ""
		gstinNumber := logg.Request.Gst
		gstinStatus := ""

		for k, v := range data {
			e = "Fifth"
			if v == nil {
				data[k] = ""
			}
		}
		//fmt.Println(data)

		if len(data) > 0 {
			gstInsertionDate = data["gst_insertion_date"].(string)
			gstinStatus = data["gstin_status"].(string)
			gstinStatus = strings.Trim(strings.ToUpper(gstinStatus), " ")
			buisness_constitutionid1, ok11 := data["business_constitution_group_id"].(string)
			if ok11 {
				data["business_constitution"] = utils.LegalStatusRead(buisness_constitutionid1)
			}

		}

		if len(data) == 0 {
			logg.MasterIndia.Hit = true
			logg.MasterIndia.User = APIUserID
		} else if days, err := utils.DiffDaysddmmyyyy(gstInsertionDate); err != nil || days > 30 || (gstinStatus != "ACTIVE" && days >= 1) || checkDist(gstInsertionDate) {
			logg.MasterIndia.Hit = true
			logg.MasterIndia.User = APIUserID
		}

		if logg.MasterIndia.Hit {
			f = "Sixth"
			wr := WorkRequest{
				APIName:     apiName,
				APIUserName: APIUserID,
				GstPan:      gstinNumber,
				Modid:       logg.Request.ModID,
				RqstTime:    logg.RequestStart,
			}

			raw, err := json.Marshal(wr)
			//fmt.Println(wr,"raw")
			if err != nil {
				sendResponse(uniqID, w, 400, Updateflag, failure, "failed at wr", err, Data{}, logg, logg.Request.Gst, hsncode)
				return
			}

			_, st = utils.GetExecTime()
			Updateflag = 0
			//queue
			enqData := make(map[string]string)
			enqData["publisher"] = "centralizedAPI"
			enqData["jsonDataStr"] = string(raw)
			enqData["msgBody"] = enqData["publisher"]
			// enqData["msgDuplicationID"] = wr.APIName + wr.GstPan
			// enqData["msgGroupID"] = wr.Modid

			msgID, err := queue.Send(enqData)
			if err != nil {
				logg.StackTrace = err.Error()
				sendResponse(uniqID, w, 500, Updateflag, failure, errPanic, nil, Data{}, logg, logg.Request.Gst, hsncode)
				return
			}
			logg.QueueMsgID = msgID
			logg.ExecTime["GSTData"], st = utils.GetExecTime(st)

			_, st = utils.GetExecTime()
			g = "Seventh"
			data, err = GetGSTRecords36(database, gstinNumber)

			// logg.DbDataFlag2 = data
			logg.ExecTime["DB_QUERY_2"], st = utils.GetExecTime(st)

			if err != nil {
				sendResponse(uniqID, w, 200, Updateflag, failure, errFetchDB, err, Data{}, logg, logg.Request.Gst, hsncode)
				return
			} else {
				DbTurnover, ok_new := data["gst_glusr_turnover"]

				// Check if the key is present and the original value is already a string
				if ok_new {
					DbTurnoverString, ok := DbTurnover.(string)
					if ok {
						// If the original value is a string, Do Format it
						Final := processJSON(DbTurnoverString)

						// Update the value of k1 with the new interface value
						data["annual_turnover_slab"] = Final
					}
				}

				buisness_constitutionid, ok11 := data["business_constitution_group_id"].(string)
				if ok11 {
					fmt.Println("lineno28666=")
					data["business_constitution"] = utils.LegalStatusRead(buisness_constitutionid)
					fmt.Println("288=", data["business_constitution"])
				}

				grossincome_db, okk := data["gross_income"]

				if okk {
					grossincome_db_String, ok := grossincome_db.(string)
					if ok {
						// If the original value is a string, Do Format it
						Final := ReplaceBreakWithComma(grossincome_db_String)

						// Update the value of k1 with the new interface value
						data["gross_income"] = Final
					}
				}
			}
			h = "Eighth"
		}

	} else {

		if logg.Request.ModID == "PAY" || logg.Request.ModID == "loans" || logg.Request.ModID == "loans2" || logg.Request.ModID == "merpnsd" {
			var (
				data_gst2 map[string]interface{}
			)
			//First Get GST and then go for Get GST Records 36
			data_gst2, err = GetGstFromGlid(database, logg.Request.Glid)
			if err != nil {
				sendResponse(uniqID, w, 400, Updateflag, failure, errFetchDB, err, Data{}, logg, logg.Request.Gst, hsncode)
				return
			}

			for k, v := range data_gst2 {
				//e = "Fifth"
				if v == nil {
					data_gst2[k] = ""
				}
			}

			if len(data_gst2) == 0 || data_gst2["gst"] == "" {
				sendResponse(uniqID, w, 400, Updateflag, failure, "There is no GST mapped to this user", nil, Data{}, logg, logg.Request.Gst, hsncode)
				return
			}

			if data_gst2["gst"] != "" {
				logg.Request.Gst = data_gst2["gst"].(string)
				// gstnum=logg.Request.Gst
				// fmt.Println("GST_NUM",gstnum)
				if logg.Request.ModID == "loans2" || logg.Request.ModID == "merpnsd" {
					if logg.Request.Gst != "" {

						data, err := GetGSTRecords75(database, logg.Request.Gst)
						if err != nil {
							sendResponse(uniqID, w, 400, Updateflag, failure, errValidationKey, err, Data{}, logg, logg.Request.Gst, hsncode)
							return
						}
						for k, v := range data {
							if v == nil {
								data[k] = ""
							}
						}
						result := make(map[string]interface{})
						for _, v := range cols {
							result[v] = data[v]
						}

						// sixthChar, err1 := GetSixthCharacter(logg.Request.Gst)
						// if err1 != nil {
						// 	sendResponse(uniqID, w, 400, Updateflag, failure, "There is no GST mapped to this user", nil, Data{}, logg, logg.Request.Gst, hsncode)
						// 	return
						// }

						// ownershipType := GetOwnershipType(sixthChar)
						legalstatusid, ok11 := result["business_constitution_group_id"].(string)
						if ok11 {
							result["business_constitution"] = utils.LegalStatusRead(legalstatusid)
						}

						// if sixthChar == 'F' {

						// 	trade_name_replaced_interface, ok := result["trade_name_replaced"]

						// 	if ok {
						// 		trade_name_replaced, ok2 := trade_name_replaced_interface.(string)

						// 		if ok2 {
						// 			llpornot := CheckLastWordLLP(trade_name_replaced)

						// 			if llpornot {
						// 				result["business_constitution"] = "Limited Liability Partnership"
						// 			} else {
						// 				result["business_constitution"] = "Partnership Firm"
						// 			}
						// 		}
						// 	}

						// }

						loansTurnover, ok_new := result["gst_glusr_turnover"]

						// Check if the key is present and the original value is already a string
						if ok_new {
							DbTurnoverString, ok := loansTurnover.(string)
							if ok {

								// Update the value of k1 with the new interface value
								result["gst_glusr_turnover"] = DbTurnoverString
							}
						}

						sendResponse(uniqID, w, 200, Updateflag, success, "", nil, Data{Values: result}, logg, logg.Request.Gst, hsncode)
						data_gst2 = make(map[string]interface{})
						return

					}
				}
			}

			data_gst2 = make(map[string]interface{})

		}

		i = "ninth"
		_, st := utils.GetExecTime()
		if logg.Request.Gst != "" {
			j = "tenth"
			data, err = GetGSTRecords(database, logg.Request.Gst)
			// logg.DbBeforeData = data
		} else {
			k = "Eleventh"
			data, err = GetGlidRecords(database, logg.Request.Glid)
		}

		logg.ExecTime["DB_QUERY"], st = utils.GetExecTime(st)

		APIUserID, err := masterindia.ValidateProp(logg.Request.ModID, logg.Request.Validationkey, apiName)

		if err != nil {
			sendResponse(uniqID, w, 400, Updateflag, failure, errValidationKey, err, Data{}, logg, logg.Request.Gst, hsncode)
			return
		}

		gstInsertionDate := ""
		gstinNumber := logg.Request.Gst
		primDist := ""
		gstinStatus := ""

		for k, v := range data {
			l = "twelvth"
			if v == nil {
				data[k] = ""
			}
		}
		//fmt.Println(data)

		if len(data) > 0 {
			gstInsertionDate = data["gst_insertion_date"].(string)
			primDist = data["bussiness_fields_add_district"].(string)
			gstinStatus = data["gstin_status"].(string)
			gstinStatus = strings.Trim(strings.ToUpper(gstinStatus), " ")
		}

		iu := ""
		if len(data) == 0 {
			logg.MasterIndia.Hit = true
			logg.MasterIndia.User = APIUserID
			iu = "I"
		} else if days, err := utils.DiffDaysddmmyyyy(gstInsertionDate); err != nil || days > 30 || gstinStatus != "ACTIVE" || checkDistPrim(gstInsertionDate, primDist) {
			logg.MasterIndia.Hit = true
			logg.MasterIndia.User = APIUserID
			iu = "U"
		}

		if logg.MasterIndia.Hit {
			m = "Thirteenth"
			wr := masterindia.Work{
				APIName:   apiName,
				APIUserID: logg.MasterIndia.User,
				GST:       gstinNumber,
				Modid:     logg.Request.ModID,
				UniqID:    uniqID,
			}
			Updateflag = 1
			_, st = utils.GetExecTime()
			_, params, err := wr.FetchGSTData(masterindiaAPILogs, 3000)
			logg.ExecTime["FetchGSTData"], st = utils.GetExecTime(st)

			if err != nil {
				sendResponse(uniqID, w, 200, Updateflag, failure, errFetchAPI, err, Data{}, logg, logg.Request.Gst, hsncode)
				return
			}

			_, st = utils.GetExecTime()
			if strings.ToUpper(iu) == "U" {
				n = "Fourteenth"
				_, err = model.UpdateGSTMasterData(database, params)
			} else if strings.ToUpper(iu) == "I" {
				o = "fifteen"
				_, err = model.InsertGSTMasterData(database, params)
			}
			logg.ExecTime["I_U_GST_DATA"], st = utils.GetExecTime(st)

			if err != nil {
				sendResponse(uniqID, w, 200, Updateflag, failure, errUpdateDB, err, Data{}, logg, logg.Request.Gst, hsncode)
				return
			}

			// 	data_glid, err := authadvance.GetGlidFromGstM(database, gstinNumber)
			//     if err != nil {
			// 	    fmt.Println("getting error from getglidfromgst function",err.Error())
			//     } else {
			// 	   for _, glid := range data_glid {
			// 		if err := authadvance.ProcessSingleGLIDPubapilogging(glid, "/gstdata/v1/gst"); err != nil {
			// 			// logg.AnyError["meshupdationerror"] = err.Error()
			// 			// s3log.Result["meshupdationerror_glid"] = fmt.Sprintf("GLID %s: %v", glid, err)
			// 			fmt.Println("meshupdationerror_glid",err.Error())
			// 		}
			//         }
			//    }

			if err := authadvance.ProcessingGST(gstinNumber, "/gstdata/v1/gst"); err != nil {
				// logg.AnyError["meshupdationerror"] = err.Error()
				fmt.Println("meshupdationerror_gst", err.Error())
			}

			_, st = utils.GetExecTime()
			p = "sixteenth"
			data, err = GetGSTRecords(database, gstinNumber)
			// logg.DbData = data
			logg.ExecTime["DB_QUERY_2"], st = utils.GetExecTime(st)

			if err != nil {
				sendResponse(uniqID, w, 200, Updateflag, failure, errFetchDB, err, Data{}, logg, logg.Request.Gst, hsncode)
				return
			}
		}
		q = "seventeenth"
	}

	result := make(map[string]interface{})
	for _, v := range cols {
		R = "eighteenth"
		result[v] = data[v]
	}

	//new changes added for loans team
	if logg.Request.ModID == "loans" {
		for k, v := range result {
			if k == "registration_date" {
				days, err := DaysBetweenDates(v.(string))
				if err != nil {
					fmt.Println("Error:", err)
					return
				}

				days = Abs(days)
				months := int(math.Ceil(float64(days) / 30.0)) // Floor division by 30 to get the number of full months.
				fmt.Printf("There are approximately %d full months (of 30 days each) between today and %s.\n", months, v.(string))
				result["gstvintage"] = strconv.Itoa(months)
			}

		}
	}

	//new changes added for Navison Team Purpose Integratin of Pincode to City API

	if _, ok11 := result["pincode"]; ok11 {

		if _, ok12 := result["bussiness_fields_add_district"]; ok12 {

			db_pin, _ := result["pincode"].(string)
			db_pin = strings.TrimSpace(db_pin)

			db_district, _ := result["bussiness_fields_add_district"].(string)
			db_district = strings.TrimSpace(db_district)

			if len(db_district) == 0 && len(db_pin) > 0 && len(db_pin) == 6 {
				//call pincode to city api
				citymap, e := CallcityFetch(db_pin)
				if e != nil {
					// logg.AnyError["CityFetch"] = err.Error()
					fmt.Println(e)
				}

				if _, ok13 := citymap["city_name"]; ok13 {

					result["bussiness_fields_add_district"] = citymap["city_name"].(string)
					latestCityName := result["bussiness_fields_add_district"]
					var params1 []interface{}

					params1 = append(params1, result["gstin_number"], latestCityName)
					_, err3 := model.UpdatePincode(database, params1)
					if err3 != nil {
						fmt.Println("Error in Updating district : ", err3)
					}
				}

			}

		}
	}

	// changes ended
	s = "nineteenth"
	sendResponse(uniqID, w, 200, Updateflag, success, "", nil, Data{Values: result}, logg, logg.Request.Gst, hsncode)
	//data=make(map[string]interface{})
	t = "twentith"

	arr := []string{gst_no, HittingTime, a, b, c, d, e, f, g, h, i, j, k, l, m, n, o, p, q, R, s, t}

	urlsJson, _ := json.Marshal(arr)
	fmt.Println(string(urlsJson))
	ioutil.WriteFile("/var/log/application/GST/GST_DATA_FLOW_LOGS/a.json", urlsJson, os.ModePerm)
	data = make(map[string]interface{})
	return
}

func validateProp(modid string, validationkey string, api string) (string, error) {
	if constants.Properties[modid].ValidaionKey != validationkey || validationkey == "" || modid == "" || api == "" {
		return "", errors.New("Not Authorized")
	}

	for k, v := range constants.Properties[modid].AllowedAPIS {
		if k == api {
			return v, nil
		}
	}
	return "", errors.New("Not Authorized")
}

func sendResponse(uniqID string, w http.ResponseWriter, httpcode int, updateflag int, status string, errorMsg string, err error, body Data, logg Logg, gst_in_num string, hsn_new string) {

	fmt.Println("hsn_new: ", hsn_new)
	w.Header().Set("Content-Type", "application/json")

	logg.Response = Res{
		Code:       httpcode,
		Error:      errorMsg,
		Status:     status,
		Body:       body,
		UniqID:     uniqID,
		UpdateFlag: updateflag,
		GstNum:     gst_in_num,
		HSNcode:    hsn_new,
	}

	if err != nil {
		logg.AnyError[errorMsg] = err.Error()
	}

	json.NewEncoder(w).Encode(logg.Response)

	logg.RequestEndValue = utils.GetTimeInNanoSeconds()
	logg.ResponseTime = (logg.RequestEndValue - logg.RequestStartValue) / 1000000
	logg.ResponseTime_Float = (logg.RequestEndValue - logg.RequestStartValue) / 1000000
	logg.RequestStartValue = 0
	logg.RequestEndValue = 0

	writeLog2(logg)

	if logg.Request.ModID != "weberp2" {
		writeLog2Kibana(logg)
	}

	return
}

func sendResponseTogstOtp(uniqID string, w http.ResponseWriter, httpcode int, status string, errorMsg string, err error, respData []VendorResp, logg Logg) {

	w.Header().Set("Content-Type", "application/json")

	otpResp := OtpResponse{
		Code:   httpcode,
		Status: status,
		Error:  errorMsg,
		Data:   respData,
		UniqID: uniqID,
	}

	if err != nil {
		logg.AnyError[errorMsg] = err.Error()
	}

	json.NewEncoder(w).Encode(otpResp)

	logg.RequestEndValue = utils.GetTimeInNanoSeconds()
	logg.ResponseTime = (logg.RequestEndValue - logg.RequestStartValue) / 1000000
	logg.ResponseTime_Float = (logg.RequestEndValue - logg.RequestStartValue) / 1000000
	logg.RequestStartValue = 0
	logg.RequestEndValue = 0

	writeLog2(logg)

	writeLog2Kibana(logg)
	

	return
}

// writeLog2 ...
func writeLog2(logg Logg) {

	logsDir := serviceLogPath + utils.TodayDir()

	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		e := os.MkdirAll(logsDir, os.ModePerm)
		if e != nil {
			fmt.Println(e)
			return
		}
	}

	logsDir += "/" + logFileName

	jsonLog, _ := json.Marshal(logg)

	f, _ := os.OpenFile(logsDir, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()

	f.WriteString("\n" + string(jsonLog))

	fmt.Println("\n" + string(jsonLog))
	return
}

func writeLog2Kibana(logg Logg) {

	logsDir := serviceLogPath + utils.TodayDir()

	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		e := os.MkdirAll(logsDir, os.ModePerm)
		if e != nil {
			fmt.Println(e)
			return
		}
	}

	logsDir += "/" + logKibanaFileName

	jsonLog, _ := json.Marshal(logg)

	f, _ := os.OpenFile(logsDir, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()

	f.WriteString("\n" + string(jsonLog))

	fmt.Println("\n" + string(jsonLog))
	return
}

func checkDist(gstInsertionDate string) bool {

	gstDate, err := time.Parse("02-01-2006", gstInsertionDate)

	if err != nil {
		return true
	}

	liveDate, _ := time.Parse("02-01-2006", "15-04-2021")

	if gstDate.Sub(liveDate).Hours() <= 0 {
		return true

	}

	return false
}

func checkDistPrim(gstInsertionDate, primDist string) bool {

	if primDist != "" {
		return false
	}

	gstDate, err := time.Parse("02-01-2006", gstInsertionDate)

	if err != nil {
		return true
	}

	liveDate, _ := time.Parse("02-01-2006", "15-04-2021")

	if gstDate.Sub(liveDate).Hours() <= 0 {
		return true

	}

	return false
}

func CheckAndWriteHSN(gst string) string {
	jsonStr, err := servapi.HsnReadDetails(gst)
	fin_hsn := ""
	fmt.Println("test-1:")
	if err != nil {
		fmt.Println("test-2:")
		fmt.Println("error: ", err)
		fin_hsn = ""
		return fin_hsn
	} else {
		fmt.Println("test-3:")
		data := make(map[string]interface{})
		err = json.Unmarshal([]byte(jsonStr), &data)
		if err != nil {
			fmt.Println("test-4:")
			fmt.Println("error in unmarshalling: ", err)
			fin_hsn = ""
			return fin_hsn
		}

		if val, ok := data["RESPONSE"]; ok {
			fmt.Println("test-5:")
			value, _ := val.(map[string]interface{})

			if value["CODE"].(string) == "200" && value["STATUS"].(string) == "SUCCESS" && value["MESSAGE"].(string) == "DATA NOT FOUND" {
				fmt.Println("test-6:")
				fin_hsn = ""
				return fin_hsn
			}
			// initial := 0
			if value["CODE"].(string) == "200" && value["STATUS"].(string) == "SUCCESS" && value["MESSAGE"].(string) == "DATA FOUND" {
				fmt.Println("test-7:")
				DATA, ok3 := value["DATA"].([]interface{})
				if ok3 && len(DATA) > 0 {
					fmt.Println("test-8: ", DATA)
					trace5 := ""
					// trace4 := ""
					k := 0
					for i := 0; i < len(DATA); i++ {
						fmt.Println("test-9:")
						record, ok4 := DATA[i].(map[string]interface{})

						if ok4 && len(record) > 0 {

							fmt.Println("test-9a: ", record)
							for k, v := range record {

								if v == nil {
									record[k] = ""
								}
								fmt.Println("test-9b: ", record[k])
							}

							case_sefntive_flag := 0
							//check whether hsn_code present
							if _, oksmall := record["hsn_code"]; oksmall {
								case_sefntive_flag = 1
							}

							//check whether HSN_CODE present
							if _, okCap := record["HSN_CODE"]; okCap {
								case_sefntive_flag = 2
							}

							if case_sefntive_flag == 1 {
								if record["hsn_code"].(string) != "" {
									k = k + 1
									hsn := record["hsn_code"].(string)
									fmt.Println("test-9c: ", hsn)
									if k == 1 {
										// trace4 = glid
										fmt.Println("test-9d: ", hsn, " ", k)
										trace5 = hsn
									} else {
										trace5 = trace5 + "," + hsn
										fmt.Println("test-9e: ", trace5)
									}

								}
							}

							if case_sefntive_flag == 2 {
								if record["HSN_CODE"].(string) != "" {
									k = k + 1
									hsn := record["HSN_CODE"].(string)
									fmt.Println("test-9c: ", hsn)
									if k == 1 {
										// trace4 = glid
										fmt.Println("test-9d: ", hsn, " ", k)
										trace5 = hsn
									} else {
										trace5 = trace5 + "," + hsn
										fmt.Println("test-9e: ", trace5)
									}

								}
							}

							// if record["hsn_code"].(string) != "" {
							//              k = k + 1
							//              hsn := record["hsn_code"].(string)
							//              fmt.Println("test-9c: ",hsn)
							//              if k == 1 {
							//               // trace4 = glid
							//               fmt.Println("test-9d: ",hsn," ",k)
							//               trace5 = hsn
							//              } else {
							//               trace5 = trace5 + "," + hsn
							//               fmt.Println("test-9e: ",trace5)
							//              }

							// }

						}

					}
					//trace5 = fmt.Printf("%s", trace5)
					fmt.Println("test-10:")
					fin_hsn = trace5
					return fin_hsn
				}

				fmt.Println("test-11:")
				fin_hsn = ""
				return fin_hsn

			}

		} else {
			fmt.Println("test-12:")
			fmt.Println("Response is not present:")
			fin_hsn = ""
			return fin_hsn
		}

	}

	fmt.Println("test-13:")

	return fin_hsn
}

// calling cityFetch API
// CallcityFetch calls the city fetch API and processes the response.
func CallcityFetch(pincode string) (map[string]interface{}, error) {
	jsonStr1, err := servapi.CityFetch(pincode)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	if err := json.Unmarshal([]byte(jsonStr1), &data); err != nil {
		return nil, err
	}

	// Convert all keys in the top-level data map to lowercase
	lowercasedData := make(map[string]interface{})
	for k, v := range data {
		lowercasedData[strings.ToLower(k)] = v
	}

	// Extract and check the CODE and MESSAGE fields
	code, codeOk := lowercasedData["code"].(string)
	message, messageOk := lowercasedData["message"].(string)

	// If CODE is not 200 or MESSAGE is not "Success", return an error
	if !codeOk || code != "200" || !messageOk || message != "Success" {
		return map[string]interface{}{
			"city_id":       "",
			"city_name":     "",
			"state_id":      "",
			"state_name":    "",
			"district_id":   "",
			"district_name": "",
		}, fmt.Errorf("API error: %v", message)
	}

	// Prepare the map to store the result with lowercase keys and default values
	result := map[string]interface{}{
		"city_id":       "",
		"city_name":     "",
		"state_id":      "",
		"state_name":    "",
		"district_id":   "",
		"district_name": "",
	}

	// Handle nested fields with potential case variations in keys
	if dataField, ok := lowercasedData["data"].(map[string]interface{}); ok {
		for _, key := range []string{"city", "state", "district"} {
			// Convert key to lowercase for consistency
			for nestedKey, nestedValue := range dataField {
				if strings.ToLower(nestedKey) == key {
					if subField, ok := nestedValue.(map[string]interface{}); ok {
						for k, v := range subField {
							result[strings.ToLower(k)] = v
						}
					}
				}
			}
		}
	}

	return result, nil
}

func TurnoverFormat(input string) string {
	if strings.Contains(input, "<br/>") && strings.Contains(input, "(For") {
		strs := strings.Split(input, "<br/>")
		return "[ [" + strings.TrimSpace(strs[1]) + " , " + strings.TrimSpace(strs[0]) + " ] ]"
	} else if strings.Contains(input, "(For") {
		forIndex := strings.Index(input, "(For")
		return "[ [" + input[forIndex:] + " , " + strings.TrimSpace(input[:forIndex-1]) + " ] ]"
	} else {
		return input
	}
}

func ReplaceBreakWithComma(input string) string {
	if strings.Contains(input, "<br/>") {
		input = strings.ReplaceAll(input, "<br/>", ",")
	}

	return input
}

// formatTurnoverData formats the turnover data from the JSON string
func formatTurnoverData(jsonStr string) (string, error) {
	if jsonStr == "" {
		// Return an empty string for empty input
		return "", nil
	}

	var data GstGlusrTurnover
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return "", err
	}

	// Sort the turnover slice by FY in descending order
	sort.SliceStable(data.Turnover, func(i, j int) bool {
		return data.Turnover[i].FY > data.Turnover[j].FY
	})

	var formattedList []string
	for _, t := range data.Turnover {
		formattedItem := fmt.Sprintf("[FY %s , %s]", t.FY, t.Value)
		formattedList = append(formattedList, formattedItem)
	}

	// Join the list items with a comma and space
	formattedString := strings.Join(formattedList, " , ")

	return formattedString, nil
}

// processJSON takes a JSON string and returns a formatted string
func processJSON(jsonStr string) string {
	formattedData, err := formatTurnoverData(jsonStr)
	if err != nil {
		// fmt.Println("Error formatting data:", err)
		return ""
	}
	return formattedData
}
