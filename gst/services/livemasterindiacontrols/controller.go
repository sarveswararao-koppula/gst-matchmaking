package livemasterindiacontrols

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	api "mm/api/thirdpartyapi"
	"mm/components/constants"
	db "mm/components/database"
	model "mm/model/masterindiamodel"
	"mm/properties"
	authadvance "mm/services/authbridgeadvanced"
	"mm/utils"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	xid "github.com/rs/xid"
)

var (
	conn     *sql.DB
	stmtGlid *sql.Stmt
	stmtGST  *sql.Stmt
)

var database string = properties.Prop.DATABASE
var loc *time.Location = utils.GetLocalTime()

//var database = "dev"

var mutex = &sync.Mutex{}

// ResponseMasterindiacontrols ...
type ResponseMasterindiacontrols struct {
	Code       int                    `json:"code,omitempty"`
	Status     string                 `json:"status,omitempty"`
	ErrMessage string                 `json:"err_message,omitempty"`
	Body       map[string]interface{} `json:"body,omitempty"`
	ErrorCodes map[string]interface{} `json:"errorcodes,omitempty"`
}

// Masterindiacontrols ...
type Masterindiacontrols struct {
	RemoteAddress      string                      `json:"RemoteAddress,omitempty"`
	RequestStart       string                      `json:"RequestStart,omitempty"`
	RequestStartValue  float64                     `json:"RequestStartValue,omitempty"`
	RequestEnd         string                      `json:"RequestEnd,omitempty"`
	RequestEndValue    float64                     `json:"RequestEndValue,omitempty"`
	ResponseTime       string                      `json:"ResponseTime,omitempty"`
	ResponseTime_Float float64                     `json:"ResponseTime_Float,omitempty"` // float64 type
	ServiceName        string                      `json:"ServiceName,omitempty"`
	ServicePath        string                      `json:"ServicePath,omitempty"`
	ServiceURL         string                      `json:"ServiceURL,omitempty"`
	UniqueID           string                      `json:"UniqueID,omitempty"`
	Response           ResponseMasterindiacontrols `json:"Response,omitempty"`
	RequestData        Rqst                        `json:"RequestData,omitempty"`
	StackTrace         string                      `json:"StackTrace,omitempty"`
	QueueMsgID         string                      `json:"QueueMsgID,omitempty"`
	ReadResponseTime   string                      `json:"ReadResponseTime,omitempty"`
	WriteResponseTime  string                      `json:"WriteResponseTime,omitempty"`
}

// Rqst ...
type Rqst struct {
	Modid         string `json:"modid,omitempty"`
	Validationkey string `json:"validationkey,omitempty"`
	API           string `json:"api,omitempty"`
	Gst           string `json:"gst,omitempty"`
	Glid          string `json:"glid,omitempty"`
}

// WorkRequest ...
type WorkRequest struct {
	APIName     string `json:"APIName,omitempty"`
	APIUserName string `json:"APIUserName,omitempty"`
	GstPan      string `json:"GstPan,omitempty"`
	Modid       string `json:"Modid,omitempty"`
	RqstTime    string `json:"RqstTime,omitempty"`
}

// GetGSTData ...
func GetGSTData(w http.ResponseWriter, r *http.Request) {

	var logs Masterindiacontrols
	logs.RequestStart = utils.GetTimeStampCurrent()
	logs.RequestStartValue = utils.GetTimeInNanoSeconds()
	logs.ServiceName = "masterdata"
	logs.ServicePath = r.URL.Path
	logs.ServiceURL = "/realtimemasterindia/v1/gst"
	logs.RemoteAddress = utils.GetIPAdress(r)
	logs.UniqueID = xid.New().String() + xid.New().String()
	data := make(map[string]interface{})
	logs.RequestData = Rqst{}

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&logs.RequestData)

	if err != nil {
		fmt.Println(err)
		sendResponse(w, 400, "FAILURE", "Invalid Params", data, nil, &logs)
		return
	}

	user, err := validateProp(logs.RequestData.Modid, logs.RequestData.Validationkey, logs.RequestData.API)

	if err != nil {
		sendResponse(w, 400, "FAILURE", err.Error(), data, nil, &logs)
		return
	}

	if logs.RequestData.API == "challan" {
		challanHand(user, w, &logs)
		return
	}

	if logs.RequestData.API == "masterindia" {
		masterindiaHand(user, w, &logs)
		return
	}

}

func challanHand(user string, w http.ResponseWriter, logs *Masterindiacontrols) {
	gstinNumber := logs.RequestData.Gst
	data := make(map[string]interface{})

	defer func() {
		if panicCheck := recover(); panicCheck != nil {
			stack := string(debug.Stack())
			fmt.Println(stack)
			logs.StackTrace = stack
			errorMsg := fmt.Sprintf("Panic occurred: %v", panicCheck)
			sendResponse(w, 400, "FAILURE", errorMsg, data, nil, logs)
			return
		}
	}()

	if len(gstinNumber) < 13 {
		sendResponse(w, 400, "FAILURE", "Param Missing", data, nil, logs)
		return
	}

	if logs.RequestData.Modid == "merpcsd" || logs.RequestData.Modid == "merpnsd" {
		wr := WorkRequest{
			APIName:     logs.RequestData.API,
			APIUserName: user,
			GstPan:      logs.RequestData.Gst,
			Modid:       logs.RequestData.Modid,
			RqstTime:    logs.RequestStart,
		}

		raw, err := json.Marshal(wr)
		if err != nil {
			fmt.Println("Error marshalling WorkRequest:", err)
			return
		}

		fyLatestFiling, err2 := SubcriberHandler2(string(raw))
		if err2 != nil {
			errStr := err2.Error()
			fmt.Println("Subscriber Handler2 error:", err2)
			sendResponse(w, 400, "FAILURE", errStr, data, nil, logs)
			return
		}
		data["dof"] = fyLatestFiling
		// fmt.Println("dof=",data)
		sendResponse(w, 200, "SUCCESS", "", data, nil, logs)
		return
	}
}

func masterindiaHand(user string, w http.ResponseWriter, logs *Masterindiacontrols) {

	fmt.Println("Request started: ")

	response := make(map[string]interface{})

	defer func() {
		if panicCheck := recover(); panicCheck != nil {
			stack := string(debug.Stack())
			fmt.Println(stack)
			logs.StackTrace = stack
			sendResponse(w, 500, "FAILURE", "Panic...Pls inform Dev Team", response, nil, logs)
			return
		}
	}()

	if logs.RequestData.Modid == "weberp" {
		weberp(w, logs, user)
		return
	}

	if len(logs.RequestData.Gst) == 0 {
		sendResponse(w, 400, "FAILURE", "Param missing", response, nil, logs)
		return
	}

	if len(logs.RequestData.Gst) < 13 {
		sendResponse(w, 400, "FAILURE", "Not valid GST", response, nil, logs)
		return
	}

	//merpcsd
	if logs.RequestData.Modid == "merpcsd" || logs.RequestData.Modid == "merpnsd" {
		merp(w, logs, user)
		return
	}

}

func merp(w http.ResponseWriter, logs *Masterindiacontrols, user string) {

	response := make(map[string]interface{})
	err_response := make(map[string]interface{})

	var s3log S3Log
	s3log.APIName = logs.RequestData.API

	credential := utils.GetCred(user)
	s3log.APIUserID = credential["username"]
	s3log.Gst = logs.RequestData.Gst
	s3log.APIHit = ""
	s3log.Modid = logs.RequestData.Modid
	s3log.RqstTime = logs.RequestStart
	s3log.Result = make(map[string]interface{})

	st := utils.GetTimeInNanoSeconds()
	//Logging Read Responose Time

	startread := time.Now()

	cols, err := ValidateProp(logs.RequestData.Modid, logs.RequestData.Validationkey, "")
	//fmt.Println("cols",cols)
	if err != nil {
		sendResponse(w, 401, "FAILURE", "error in validation prop", response, nil, logs)
		return
	}

	dbData, err := model.GetGSTRecordsNew(database, s3log.Gst)

	elapsedread := time.Since(startread)

	logs.ReadResponseTime = fmt.Sprint(elapsedread.Milliseconds())

	//Logging Ends
	en := utils.GetTimeInNanoSeconds()
	fmt.Println("GetGSTRecords", (en-st)/1000000)
	//response["result"] = "GST Details would be updated"

	if err != nil {
		s3log.APIHit = err.Error()
		sendResponse(w, 500, "FAILURE", "error in fetching gst records", response, nil, logs)
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	dbGstGroupid, _ := dbData["business_constitution_group_id"].(string)
	legalStatusValue := utils.LegalStatusRead(dbGstGroupid)
	dbData["business_constitution"] = legalStatusValue

	dbGst, _ := dbData["gstin_number"].(string)
	dbGstInsertionDate, _ := dbData["gst_insertion_date"].(string)
	dbGstinStatus, _ := dbData["gstin_status"].(string)

	s3log.Result["gst_insertion_date_in_db"] = dbGstInsertionDate

	gapOfDays := DaysDiff(dbGstInsertionDate)
	if (dbGst == s3log.Gst) && ((dbGstinStatus == "Active" && gapOfDays <= 30) || (dbGstinStatus != "Active" && gapOfDays <= 1)) {

		s3log.APIHit = "GST Details already fetched successfully within 30 days"

		for k, v := range dbData {
			if v == nil {
				dbData[k] = ""
			}
		}

		for _, v := range cols {
			response[v] = dbData[v]
		}
		//response["gstin_status"] = dbGstinStatus

		sendResponse(w, 200, "SUCCESS", "", response, nil, logs)
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	st = utils.GetTimeInNanoSeconds()
	err = ValidateTokken(credential, user)
	en = utils.GetTimeInNanoSeconds()
	fmt.Println("GetTokken", (en-st)/1000000)

	if err != nil {
		s3log.Result["api_error"] = err.Error()
		//new Changes -> start
		ErrorParam1 := utils.GetErrorParams(s3log.Gst, err.Error())
		if ErrorParam1[1] == 102 || ErrorParam1[1] == 104 || ErrorParam1[1] == 107 || ErrorParam1[1] == 112 {
			fmt.Println("ERROR Code that won't be inserted: ", ErrorParam1[1])
		} else {
			_, Mastererr1 := model.InsertGSTMasterErrorData(database, ErrorParam1)
			if Mastererr1 != nil {
				s3log.Result["table_error"] = Mastererr1.Error()
				// Write2S3(&s3log)
				// return
			}
			//end
		}

		err_response["err_code"] = ErrorParam1[1]
		err_response["err_message"] = ErrorParam1[2]

		sendResponse(w, 503, "FAILURE", "error in fetching masterindia tokken", response, err_response, logs)
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	s3log.APIHit = "Y"
	st = utils.GetTimeInNanoSeconds()
	data, err := api.GetMasterData(s3log.Gst, map[string]string{
		"client_id":    credential["client_id"],
		"access_token": Tokkens[user].Tok,
	})
	en = utils.GetTimeInNanoSeconds()
	fmt.Println("GetMasterData", (en-st)/1000000)

	apiData, ok := data["data"].(map[string]interface{})
	dataErrorBool, _ := data["error"].(bool)
	dataErrorStr, _ := data["error"].(string)

	if err != nil || !ok || dataErrorBool || dataErrorStr != "" {

		if err != nil {
			s3log.Result["api_error"] = err.Error()
			//new changes -start
			ErrorParam2 := utils.GetErrorParams(s3log.Gst, err.Error())

			err_response["err_code"] = ErrorParam2[1]
			err_response["err_message"] = ErrorParam2[2]

			if ErrorParam2[1] == 102 || ErrorParam2[1] == 104 || ErrorParam2[1] == 107 || ErrorParam2[1] == 112 {
				fmt.Println("ERROR Code that won't be inserted: ", ErrorParam2[1])
			} else {
				_, Mastererr2 := model.InsertGSTMasterErrorData(database, ErrorParam2)
				if Mastererr2 != nil {
					s3log.Result["table_error"] = Mastererr2.Error()
				}
				//end
			}
			sendResponse(w, 503, "FAILURE", "error in fetching masterindia details", response, err_response, logs)
			return
		} else {
			s3log.Result["api_error"] = fmt.Sprint(data)
			//new changes -start
			error_data := fmt.Sprint(data)
			ErrorParam3 := utils.GetErrorParams(s3log.Gst, error_data)

			err_response["err_code"] = ErrorParam3[1]
			err_response["err_message"] = ErrorParam3[2]

			if ErrorParam3[1] == 102 || ErrorParam3[1] == 104 || ErrorParam3[1] == 107 || ErrorParam3[1] == 112 {
				fmt.Println("ERROR Code that won't be inserted: ", ErrorParam3[1])
			} else {
				_, Mastererr3 := model.InsertGSTMasterErrorData(database, ErrorParam3)
				if Mastererr3 != nil {
					s3log.Result["table_error"] = Mastererr3.Error()
				}
				//end
			}
			sendResponse(w, 503, "FAILURE", "error in fetching masterindia details", response, err_response, logs)
			return
		}

		apiDataMsg, _ := data["message"].(string)
		apiDataStr, _ := data["data"].(string)

		if strings.ToLower(apiDataMsg) == "the gstin passed in the request is invalid." || strings.ToLower(apiDataStr) == "the gstin passed in the request is invalid." {

			err_response["err_code"] = 102
			err_response["err_message"] = "the gstin passed in the request is invalid."
			sendResponse(w, 504, "FAILURE", "the gstin passed in the request is invalid.", response, err_response, logs)
			Write2S3(&s3log)
			Write2Kibana(&s3log)
			return
		}

		err_response["err_code"] = 117
		err_response["err_message"] = "error from masterindia api"

		sendResponse(w, 503, "FAILURE", "error from masterindia api", response, nil, logs)
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	var params []interface{}
	var myerror error
	var ApiDataMap map[string]string

	st = utils.GetTimeInNanoSeconds()
	// s3log.Result["api_data"], params = utils.BusLogicOnMasterData_V2(s3log.Gst, apiData)
	ApiDataMap, params = utils.BusLogicOnMasterData_V2(s3log.Gst, apiData)

	ApiDataMap_Copy := ApiDataMap

	en = utils.GetTimeInNanoSeconds()
	fmt.Println("BusLogicOnMasterData_V2", (en-st)/1000000)

	st = utils.GetTimeInNanoSeconds()

	//Loggin Write Response Time

	startwrite := time.Now()

	if dbGst == s3log.Gst {
		s3log.Result["i_u_d"] = "U"
		_, myerror = model.UpdateGSTMasterData(database, params)
	} else {

		s3log.Result["i_u_d"] = "I"
		_, myerror = model.InsertGSTMasterData(database, params)
	}

	elapsedwrite := time.Since(startwrite)

	logs.WriteResponseTime = fmt.Sprint(elapsedwrite.Milliseconds())

	//Logging Ends
	en = utils.GetTimeInNanoSeconds()
	fmt.Println("i_u", (en-st)/1000000)

	if myerror != nil {
		s3log.Result["i_u_d_error"] = myerror.Error()
		sendResponse(w, 500, "FAILURE", "error in i_u records in db", response, nil, logs)
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	s3logResultAPIData, _ := s3log.Result["api_data"].(map[string]string)
	fmt.Println(s3logResultAPIData["gstin_status"])

	//

	gstjsonString, marshalerr := json.Marshal(ApiDataMap)
	if marshalerr != nil {
		fmt.Println("Error converting map to JSON: %v", marshalerr)
	} else {
		// s3log.Result["LatestapiData"]= fmt.Sprintf("\"%s\"", gstjsonString)
		str := string(gstjsonString)
		str = strings.ReplaceAll(str, "\"", "")
		str = strings.ReplaceAll(str, "{", "")
		str = strings.ReplaceAll(str, "}", "")
		str = strings.ReplaceAll(str, ",", ", ")
		s3log.Result["LatestapiData"] = str
		Write2Kibana(&s3log)
	}

	s3log.Result["apiData"] = ApiDataMap_Copy

	s3logResultAPIDatanew := ApiDataMap

	timestampString := s3logResultAPIDatanew["date_of_verification"]
	timestamp, _ := time.Parse("02/01/2006", timestampString)
	response["date_of_verification"] = timestamp

	timestampString = s3logResultAPIDatanew["last_update_date"]
	timestamp, _ = time.Parse("02/01/2006", timestampString)
	response["last_update_date"] = timestamp

	response["gst_insertion_date"] = s3logResultAPIDatanew["gst_insertion_date"]

	timestampString = s3logResultAPIDatanew["registration_date"]
	timestamp, _ = time.Parse("02/01/2006", timestampString)
	response["registration_date"] = timestamp

	timestampString = s3logResultAPIDatanew["cancel_date"]
	timestamp, _ = time.Parse("02/01/2006", timestampString)
	response["cancel_date"] = timestamp

	timestampString = s3logResultAPIDatanew["date_of_filing"]
	timestamp, _ = time.Parse("02/01/2006", timestampString)
	response["date_of_filing"] = timestamp

	timestampString = s3logResultAPIDatanew["filing_last_updation_date"]
	timestamp, _ = time.Parse("02/01/2006", timestampString)
	response["filing_last_updation_date"] = timestamp

	response["fk_gl_locality_id"] = s3logResultAPIDatanew["fk_gl_locality_id"]
	response["pincode"] = s3logResultAPIDatanew["pincode"]
	response["updated_by"] = s3logResultAPIDatanew["updated_by"]
	response["fk_gstin_turnover_id"] = s3logResultAPIDatanew["fk_gstin_turnover_id"]

	response["gstin_number"] = s3logResultAPIDatanew["gstin_number"]
	response["gstin_status"] = s3logResultAPIDatanew["gstin_status"]
	response["business_name"] = s3logResultAPIDatanew["business_name"]
	response["centre_juri"] = s3logResultAPIDatanew["centre_juri"]
	response["state_juri"] = s3logResultAPIDatanew["state_juri"]
	response["business_activity_nature"] = s3logResultAPIDatanew["business_activity_nature"]
	response["taxpayer_type"] = s3logResultAPIDatanew["taxpayer_type"]

	// response["business_constitution"] = s3logResultAPIDatanew["business_constitution"]

	Gstgroupid := s3logResultAPIDatanew["business_constitution_group_id"]
	legalStatusValue2 := utils.LegalStatusRead(Gstgroupid)
	response["business_constitution"] = legalStatusValue2

	response["bussiness_address_add"] = s3logResultAPIDatanew["bussiness_address_add"]
	response["bussiness_fields_add"] = s3logResultAPIDatanew["bussiness_fields_add"]
	response["trade_name"] = s3logResultAPIDatanew["trade_name"]
	response["centre_jurisdiction_code"] = s3logResultAPIDatanew["centre_jurisdiction_code"]
	response["state_jurisdiction_code"] = s3logResultAPIDatanew["state_jurisdiction_code"]
	response["floor_number"] = s3logResultAPIDatanew["floor_number"]
	response["state_name"] = s3logResultAPIDatanew["state_name"]
	response["door_number"] = s3logResultAPIDatanew["door_number"]
	response["location"] = s3logResultAPIDatanew["location"]
	response["street"] = s3logResultAPIDatanew["street"]
	response["building_name"] = s3logResultAPIDatanew["building_name"]
	response["bussiness_place_add_nature"] = s3logResultAPIDatanew["bussiness_place_add_nature"]
	response["longitude"] = s3logResultAPIDatanew["longitude"]
	response["lattitude"] = s3logResultAPIDatanew["lattitude"]
	response["business_name_pp"] = s3logResultAPIDatanew["business_name_pp"]
	response["business_address_pp"] = s3logResultAPIDatanew["business_address_pp"]
	response["bussiness_fields_pp"] = s3logResultAPIDatanew["bussiness_fields_pp"]
	response["registration_date"] = s3logResultAPIDatanew["registration_date"]

	// response[""] = s3logResultAPIDatanew[""]

	sendResponse(w, 200, "SUCCESS", "", response, nil, logs)
	Write2S3(&s3log)

	// data_glid, err := authadvance.GetGlidFromGstM(database, logs.RequestData.Gst)
	// 	    if err != nil {
	// 		    fmt.Println("getting error from getglidfromgst function",err.Error())
	// 	    } else {
	// 		   for _, glid := range data_glid {
	// 			if err := authadvance.ProcessSingleGLIDPubapilogging(glid, "/realtimemasterindia/v1/gst"); err != nil {
	// 				// logg.AnyError["meshupdationerror"] = err.Error()
	// 				// s3log.Result["meshupdationerror_glid"] = fmt.Sprintf("GLID %s: %v", glid, err)
	// 				fmt.Println("meshupdationerror_glid",err.Error())
	// 			}
	// 	        }
	//        }
	if err := authadvance.ProcessingGST(logs.RequestData.Gst, "/realtimemasterindia/v1/gst"); err != nil {
		// logg.AnyError["meshupdationerror"] = err.Error()
		fmt.Println("meshupdationerror_gst", err.Error())
	}
	return
}

func weberp(w http.ResponseWriter, logs *Masterindiacontrols, user string) {

	response := make(map[string]interface{})
	err_response := make(map[string]interface{})

	var s3log S3Log
	s3log.APIName = logs.RequestData.API

	credential := utils.GetCred(user)
	s3log.APIUserID = credential["username"]
	// s3log.Gst = logs.RequestData.Gst
	s3log.Glid = logs.RequestData.Glid
	s3log.APIHit = ""
	s3log.Modid = logs.RequestData.Modid
	s3log.RqstTime = logs.RequestStart
	s3log.Result = make(map[string]interface{})

	st := utils.GetTimeInNanoSeconds()
	//Logging Read Responose Time

	startread := time.Now()

	// cols, err := ValidateProp(logs.RequestData.Modid, logs.RequestData.Validationkey, "")
	// //fmt.Println("cols",cols)
	// if err != nil {
	// 	sendResponse(w, 401, "FAILURE", "error in validation prop", response, nil, logs)
	// 	return
	// }

	dbData, err := GetGstFromGlid(database, s3log.Glid)

	elapsedread := time.Since(startread)

	logs.ReadResponseTime = fmt.Sprint(elapsedread.Milliseconds())

	//Logging Ends
	en := utils.GetTimeInNanoSeconds()
	fmt.Println("GetGSTRecords", (en-st)/1000000)
	//response["result"] = "GST Details would be updated"

	if err != nil {
		s3log.APIHit = err.Error()
		sendResponse(w, 500, "FAILURE", "error in fetching gst records", response, nil, logs)
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	dbGst, gstexists := dbData["gst"].(string)

	if !gstexists {
		sendResponse(w, 500, "FAILURE", "There is no GST mapped to this user", response, nil, logs)
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	s3log.Gst = dbGst
	st = utils.GetTimeInNanoSeconds()
	err = ValidateTokken(credential, user)
	en = utils.GetTimeInNanoSeconds()
	fmt.Println("GetTokken", (en-st)/1000000)

	if err != nil {
		s3log.Result["api_error"] = err.Error()
		//new Changes -> start
		ErrorParam1 := utils.GetErrorParams(s3log.Gst, err.Error())
		if ErrorParam1[1] == 102 || ErrorParam1[1] == 104 || ErrorParam1[1] == 107 || ErrorParam1[1] == 112 {
			fmt.Println("ERROR Code that won't be inserted: ", ErrorParam1[1])
		} else {
			_, Mastererr1 := model.InsertGSTMasterErrorData(database, ErrorParam1)
			if Mastererr1 != nil {
				s3log.Result["table_error"] = Mastererr1.Error()
				// Write2S3(&s3log)
				// return
			}
			//end
		}

		err_response["err_code"] = ErrorParam1[1]
		err_response["err_message"] = ErrorParam1[2]

		sendResponse(w, 503, "FAILURE", "error in fetching masterindia tokken", response, err_response, logs)
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	s3log.APIHit = "Y"
	st = utils.GetTimeInNanoSeconds()
	data, err := api.GetMasterData(dbGst, map[string]string{
		"client_id":    credential["client_id"],
		"access_token": Tokkens[user].Tok,
	})
	en = utils.GetTimeInNanoSeconds()
	fmt.Println("GetMasterData", (en-st)/1000000)

	apiData, ok := data["data"].(map[string]interface{})
	dataErrorBool, _ := data["error"].(bool)
	dataErrorStr, _ := data["error"].(string)

	if err != nil || !ok || dataErrorBool || dataErrorStr != "" {

		if err != nil {
			s3log.Result["api_error"] = err.Error()
			//new changes -start
			ErrorParam2 := utils.GetErrorParams(s3log.Gst, err.Error())

			err_response["err_code"] = ErrorParam2[1]
			err_response["err_message"] = ErrorParam2[2]

			if ErrorParam2[1] == 102 || ErrorParam2[1] == 104 || ErrorParam2[1] == 107 || ErrorParam2[1] == 112 {
				fmt.Println("ERROR Code that won't be inserted: ", ErrorParam2[1])
			} else {
				_, Mastererr2 := model.InsertGSTMasterErrorData(database, ErrorParam2)
				if Mastererr2 != nil {
					s3log.Result["table_error"] = Mastererr2.Error()
				}
				//end
			}
			sendResponse(w, 503, "FAILURE", "error in fetching masterindia details", response, err_response, logs)
			return
		} else {
			s3log.Result["api_error"] = fmt.Sprint(data)
			//new changes -start
			error_data := fmt.Sprint(data)
			ErrorParam3 := utils.GetErrorParams(s3log.Gst, error_data)

			err_response["err_code"] = ErrorParam3[1]
			err_response["err_message"] = ErrorParam3[2]

			if ErrorParam3[1] == 102 || ErrorParam3[1] == 104 || ErrorParam3[1] == 107 || ErrorParam3[1] == 112 {
				fmt.Println("ERROR Code that won't be inserted: ", ErrorParam3[1])
			} else {
				_, Mastererr3 := model.InsertGSTMasterErrorData(database, ErrorParam3)
				if Mastererr3 != nil {
					s3log.Result["table_error"] = Mastererr3.Error()
				}
				//end
			}
			sendResponse(w, 503, "FAILURE", "error in fetching masterindia details", response, err_response, logs)
			return
		}

		apiDataMsg, _ := data["message"].(string)
		apiDataStr, _ := data["data"].(string)

		if strings.ToLower(apiDataMsg) == "the gstin passed in the request is invalid." || strings.ToLower(apiDataStr) == "the gstin passed in the request is invalid." {

			err_response["err_code"] = 102
			err_response["err_message"] = "the gstin passed in the request is invalid."
			sendResponse(w, 504, "FAILURE", "the gstin passed in the request is invalid.", response, err_response, logs)
			Write2S3(&s3log)
			Write2Kibana(&s3log)
			return
		}

		err_response["err_code"] = 117
		err_response["err_message"] = "error from masterindia api"

		sendResponse(w, 503, "FAILURE", "error from masterindia api", response, nil, logs)
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	var params []interface{}
	var myerror error
	var ApiDataMap map[string]string

	st = utils.GetTimeInNanoSeconds()
	// s3log.Result["api_data"], params = utils.BusLogicOnMasterData_V2(s3log.Gst, apiData)
	ApiDataMap, params = utils.BusLogicOnMasterData_V2(s3log.Gst, apiData)

	ApiDataMap_Copy := ApiDataMap

	en = utils.GetTimeInNanoSeconds()
	fmt.Println("BusLogicOnMasterData_V2", (en-st)/1000000)

	st = utils.GetTimeInNanoSeconds()

	//Loggin Write Response Time

	startwrite := time.Now()

	if dbGst == s3log.Gst {
		s3log.Result["i_u_d"] = "U"
		_, myerror = model.UpdateGSTMasterData(database, params)
	} else {

		s3log.Result["i_u_d"] = "I"
		_, myerror = model.InsertGSTMasterData(database, params)
	}

	elapsedwrite := time.Since(startwrite)

	logs.WriteResponseTime = fmt.Sprint(elapsedwrite.Milliseconds())

	//Logging Ends
	en = utils.GetTimeInNanoSeconds()
	fmt.Println("i_u", (en-st)/1000000)

	if myerror != nil {
		s3log.Result["i_u_d_error"] = myerror.Error()
		sendResponse(w, 500, "FAILURE", "error in i_u records in db", response, nil, logs)
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	s3logResultAPIData, _ := s3log.Result["api_data"].(map[string]string)
	fmt.Println(s3logResultAPIData["gstin_status"])

	//

	gstjsonString, marshalerr := json.Marshal(ApiDataMap)
	if marshalerr != nil {
		fmt.Println("Error converting map to JSON: %v", marshalerr)
	} else {
		// s3log.Result["LatestapiData"]= fmt.Sprintf("\"%s\"", gstjsonString)
		str := string(gstjsonString)
		str = strings.ReplaceAll(str, "\"", "")
		str = strings.ReplaceAll(str, "{", "")
		str = strings.ReplaceAll(str, "}", "")
		str = strings.ReplaceAll(str, ",", ", ")
		s3log.Result["LatestapiData"] = str
		Write2Kibana(&s3log)
	}

	s3log.Result["apiData"] = ApiDataMap_Copy

	s3logResultAPIDatanew := ApiDataMap

	timestampString := s3logResultAPIDatanew["date_of_verification"]
	timestamp, _ := time.Parse("02/01/2006", timestampString)
	response["date_of_verification"] = timestamp

	timestampString = s3logResultAPIDatanew["last_update_date"]
	timestamp, _ = time.Parse("02/01/2006", timestampString)
	response["last_update_date"] = timestamp

	response["gst_insertion_date"] = s3logResultAPIDatanew["gst_insertion_date"]

	timestampString = s3logResultAPIDatanew["registration_date"]
	timestamp, _ = time.Parse("02/01/2006", timestampString)
	response["registration_date"] = timestamp

	timestampString = s3logResultAPIDatanew["cancel_date"]
	timestamp, _ = time.Parse("02/01/2006", timestampString)
	response["cancel_date"] = timestamp

	timestampString = s3logResultAPIDatanew["date_of_filing"]
	timestamp, _ = time.Parse("02/01/2006", timestampString)
	response["date_of_filing"] = timestamp

	timestampString = s3logResultAPIDatanew["filing_last_updation_date"]
	timestamp, _ = time.Parse("02/01/2006", timestampString)
	response["filing_last_updation_date"] = timestamp

	response["fk_gl_locality_id"] = s3logResultAPIDatanew["fk_gl_locality_id"]
	response["pincode"] = s3logResultAPIDatanew["pincode"]
	response["updated_by"] = s3logResultAPIDatanew["updated_by"]
	response["fk_gstin_turnover_id"] = s3logResultAPIDatanew["fk_gstin_turnover_id"]

	response["gstin_number"] = s3logResultAPIDatanew["gstin_number"]
	response["gstin_status"] = s3logResultAPIDatanew["gstin_status"]
	response["business_name"] = s3logResultAPIDatanew["business_name"]
	response["centre_juri"] = s3logResultAPIDatanew["centre_juri"]
	response["state_juri"] = s3logResultAPIDatanew["state_juri"]
	response["business_activity_nature"] = s3logResultAPIDatanew["business_activity_nature"]
	response["taxpayer_type"] = s3logResultAPIDatanew["taxpayer_type"]

	// response["business_constitution"] = s3logResultAPIDatanew["business_constitution"]

	Gstgroupid := s3logResultAPIDatanew["business_constitution_group_id"]
	legalStatusValue := utils.LegalStatusRead(Gstgroupid)
	response["business_constitution"] = legalStatusValue

	response["bussiness_address_add"] = s3logResultAPIDatanew["bussiness_address_add"]
	response["bussiness_fields_add"] = s3logResultAPIDatanew["bussiness_fields_add"]
	response["trade_name"] = s3logResultAPIDatanew["trade_name"]
	response["centre_jurisdiction_code"] = s3logResultAPIDatanew["centre_jurisdiction_code"]
	response["state_jurisdiction_code"] = s3logResultAPIDatanew["state_jurisdiction_code"]
	response["floor_number"] = s3logResultAPIDatanew["floor_number"]
	response["state_name"] = s3logResultAPIDatanew["state_name"]
	response["door_number"] = s3logResultAPIDatanew["door_number"]
	response["location"] = s3logResultAPIDatanew["location"]
	response["street"] = s3logResultAPIDatanew["street"]
	response["building_name"] = s3logResultAPIDatanew["building_name"]
	response["bussiness_place_add_nature"] = s3logResultAPIDatanew["bussiness_place_add_nature"]
	response["longitude"] = s3logResultAPIDatanew["longitude"]
	response["lattitude"] = s3logResultAPIDatanew["lattitude"]
	response["business_name_pp"] = s3logResultAPIDatanew["business_name_pp"]
	response["business_address_pp"] = s3logResultAPIDatanew["business_address_pp"]
	response["bussiness_fields_pp"] = s3logResultAPIDatanew["bussiness_fields_pp"]
	response["registration_date"] = s3logResultAPIDatanew["registration_date"]

	response["street"] = s3logResultAPIDatanew["street"]
	response["lattitude_addl"] = s3logResultAPIDatanew["lattitude_addl"]
	response["longitude_addl"] = s3logResultAPIDatanew["longitude_addl"]

	// response[""] = s3logResultAPIDatanew[""]

	sendResponse(w, 200, "SUCCESS", "", response, nil, logs)
	Write2S3(&s3log)
	// if err := authadvance.ProcessSingleGLIDPubapilogging(logs.RequestData.Glid, "/realtimemasterindia/v1/gst"); err != nil {
	// 	// logg.AnyError["meshupdationerror"] = err.Error()
	// 	// s3log.Result["meshupdationerror_glid"] = fmt.Sprintf("GLID %s: %v", glid, err)
	// 	fmt.Println("meshupdationerror_glid",err.Error())
	// }

	if err := authadvance.ProcessingGST(dbGst, "/realtimemasterindia/v1/gst"); err != nil {
		// logg.AnyError["meshupdationerror"] = err.Error()
		fmt.Println("meshupdationerror_gst", err.Error())
	}

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

func sendResponse(w http.ResponseWriter, httpcode int, status string, errorMsg string, response map[string]interface{}, err_response map[string]interface{}, logs *Masterindiacontrols) {

	var serviceResponse ResponseMasterindiacontrols
	serviceResponse.Code = httpcode
	serviceResponse.Status = status
	serviceResponse.ErrMessage = errorMsg
	serviceResponse.Body = response
	serviceResponse.ErrorCodes = err_response
	logs.Response = serviceResponse

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(serviceResponse)

	logs.RequestEndValue = utils.GetTimeInNanoSeconds()
	logs.RequestEnd = utils.GetTimeStampCurrent()
	logs.ResponseTime = fmt.Sprint((logs.RequestEndValue - logs.RequestStartValue) / 1000000)
	logs.ResponseTime_Float = (logs.RequestEndValue - logs.RequestStartValue) / 1000000
	logs.RequestStartValue = 0
	logs.RequestEndValue = 0

	logsCp := *logs

	go write2Log(logsCp)

}

func write2Log(logs Masterindiacontrols) {

	//2009 November 10
	year, month, day := time.Now().Date()
	logsDir := properties.Prop.LOG_MASTERINDIA + "/" + fmt.Sprint(year) + "/" + fmt.Sprint(int(month)) + "/" + fmt.Sprint(day)

	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		e := os.MkdirAll(logsDir, os.ModePerm)
		fmt.Println(e)
	}

	logsDir += "/masterindia_wrapper.json"

	jsonLog, err := json.Marshal(logs)
	fmt.Println(err)
	jsonLogString := string(jsonLog[:len(jsonLog)])

	f, _ := os.OpenFile(logsDir, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()

	mutex.Lock()
	defer mutex.Unlock()
	f.WriteString("\n" + jsonLogString)
	return
}

// GetGstFromGlid ...
func GetGstFromGlid(database string, glid string) (map[string]interface{}, error) {

	glidInt, _ := strconv.Atoi(glid)

	if conn == nil {
		stmtGlid = nil
		var err error
		if conn, err = db.GetDatabaseConnection(database); err != nil {
			return nil, err
		}
	}

	if stmtGlid == nil {
		var err error
		query := `
                SELECT
				GST::text
                FROM
                GLUSR_USR_COMP_REGISTRATIONS
                WHERE FK_GLUSR_USR_ID=$1
                ;
        `
		stmtGlid, err = conn.Prepare(query)
		if err != nil {
			return nil, err
		}
	}

	var params []interface{}
	params = append(params, glidInt)

	callrecords, err := selectWithStmt(stmtGlid, params)

	if err != nil {
		conn = nil
		return nil, err
	}

	res := make(map[string]interface{})
	for _, v := range callrecords["queryData"].([]interface{}) {

		for k, v1 := range v.(map[string]interface{}) {
			res[k] = v1
		}
	}

	return res, nil
}

func selectWithStmt(statement *sql.Stmt, params []interface{}, timeOutSeconds ...int) (map[string]interface{}, error) {

	if statement == nil {
		return nil, errors.New("stmt is nil")
	}

	timeOut := 3
	if len(timeOutSeconds) > 0 {
		timeOut = timeOutSeconds[0]
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(timeOut))
	defer cancel()
	result, err := statement.QueryContext(ctx, params...)
	if err != nil {
		return nil, err
	}
	defer result.Close()
	cols, err := result.Columns()
	if err != nil {
		return nil, err
	}
	finalResult := make([]interface{}, 0)

	for result.Next() {
		data := make(map[string]interface{})
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		err := result.Scan(columnPointers...)
		if err != nil {
			fmt.Println(err.Error())
		}
		for i, colName := range cols {
			data[colName] = columns[i]
		}
		finalResult = append(finalResult, data)
	}
	returnResult := make(map[string]interface{})
	returnResult["queryData"] = finalResult
	return returnResult, err
}
