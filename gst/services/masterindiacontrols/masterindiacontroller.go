package masterindiacontrols

import (
	"encoding/json"
	"errors"
	"fmt"
	api "mm/api/thirdpartyapi"
	"mm/components/constants"
	model "mm/model/masterindiamodel"
	"mm/properties"
	"mm/queue"
	authadvance "mm/services/authbridgeadvanced"
	"mm/utils"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	xid "github.com/rs/xid"
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
	logs.ServiceURL = r.RequestURI
	logs.RemoteAddress = utils.GetIPAdress(r)
	logs.UniqueID = xid.New().String() + xid.New().String()
	data := make(map[string]interface{})
	logs.RequestData = Rqst{}

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&logs.RequestData)

	if err != nil {
		fmt.Println(err)
		sendResponse(w, 400, "FAILURE", "Invalid Params", data, &logs)
		return
	}

	user, err := validateProp(logs.RequestData.Modid, logs.RequestData.Validationkey, logs.RequestData.API)

	if err != nil {
		sendResponse(w, 400, "FAILURE", err.Error(), data, &logs)
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

func masterindiaHand(user string, w http.ResponseWriter, logs *Masterindiacontrols) {

	fmt.Println("Request started: ")

	response := make(map[string]interface{})

	defer func() {
		if panicCheck := recover(); panicCheck != nil {
			stack := string(debug.Stack())
			fmt.Println(stack)
			logs.StackTrace = stack
			sendResponse(w, 500, "FAILURE", "Panic...Pls inform Dev Team", response, logs)
			return
		}
	}()

	if len(logs.RequestData.Gst) == 0 {
		sendResponse(w, 400, "FAILURE", "Param missing", response, logs)
		return
	}

	if len(logs.RequestData.Gst) < 13 {
		sendResponse(w, 400, "FAILURE", "Not valid GST", response, logs)
		return
	}

	//buyermy merp
	if logs.RequestData.Modid == "merp" {
		merp(w, logs, user)
		return
	}

	//seller modid
	if logs.RequestData.Modid == "seller" {
		seller(w, logs, user)
		return
	}
	//buyermy modid
	if logs.RequestData.Modid == "buyermy" {
		seller(w, logs, user)
		return
	}

	//buyermy merp
	// if logs.RequestData.Modid == "merp" {
	//      merp(w, logs, user)
	//      return
	// }

	//Api_Name Api_user_name  Gst_pan Modid Rqst_time

	wr := WorkRequest{
		APIName:     logs.RequestData.API,
		APIUserName: user,
		GstPan:      logs.RequestData.Gst,
		Modid:       logs.RequestData.Modid,
		RqstTime:    logs.RequestStart,
	}

	raw, err := json.Marshal(wr)
	if err != nil {
		sendResponse(w, 400, "FAILURE", "failed at wr", response, logs)
		return
	}

	enqData := make(map[string]string)
	enqData["publisher"] = "centralizedAPI"
	enqData["jsonDataStr"] = string(raw)

	enqData["msgBody"] = enqData["publisher"]
	// enqData["msgDuplicationID"] = wr.APIName + wr.GstPan
	// enqData["msgGroupID"] = wr.Modid

	msgID, err := queue.Send(enqData)

	if err != nil {
		logs.StackTrace = err.Error()
		sendResponse(w, 200, "FAILURE", "Panic...Pls inform Dev Team", response, logs)
		return
	}

	logs.QueueMsgID = msgID

	response["result"] = "GST Details would be updated"
	fmt.Println("RESPONSE: ", response)

	sendResponse(w, 200, "SUCCESS", "", response, logs)
	return

}

// func challanHand(user string, w http.ResponseWriter, logs *Masterindiacontrols) {

// 	response := make(map[string]interface{})
// 	// gstinNumber := logs.RequestData.Gst

// 	defer func() {
// 		if panicCheck := recover(); panicCheck != nil {
// 			stack := string(debug.Stack())
// 			fmt.Println(stack)
// 			logs.StackTrace = stack
// 			sendResponse(w, 500, "FAILURE", "Panic...Pls inform Dev Team", response, logs)
// 			return
// 		}
// 	}()

// 	if len(logs.RequestData.Gst) < 13 {
// 		sendResponse(w, 400, "FAILURE", "Param missing", response, logs)
// 		return
// 	}

// 		if logs.RequestData.Modid == "merp" || logs.RequestData.Modid == "gladmin" {
// 			response["result"] = "GST Details would be updated"
// 			sendResponse(w, 200, "SUCCESS", "", response, logs)
// 		}

// 	wr := WorkRequest{
// 		APIName:     logs.RequestData.API,
// 		APIUserName: user,
// 		GstPan:      logs.RequestData.Gst,
// 		Modid:       logs.RequestData.Modid,
// 		RqstTime:    logs.RequestStart,
// 	}

// 	raw, err := json.Marshal(wr)
// 	if err != nil {
// 		// sendResponse(w, 400, "FAILURE", "failed at wr", response, logs)
// 		return
// 	}

// 	if logs.RequestData.Modid == "merp" || logs.RequestData.Modid == "gladmin" {
// 		err2 := SubcriberHandler2(string(raw))
// 		if err2 != nil {
// 			fmt.Println("Subscriber Handler2 error : ", err2)
// 		}
// 		return
// 	}

// 	enqData := make(map[string]string)
// 	enqData["publisher"] = "centralizedAPI"
// 	enqData["jsonDataStr"] = string(raw)

// 	enqData["msgBody"] = enqData["publisher"]
// 	// enqData["msgDuplicationID"] = wr.APIName + wr.GstPan
// 	// enqData["msgGroupID"] = wr.Modid

// 	msgID, err := queue.Send(enqData)

// 	if err != nil {

// 		logs.StackTrace = err.Error()
// 		sendResponse(w, 200, "FAILURE", "Panic...Pls inform Dev Team", response, logs)
// 		return
// 	}

// 	logs.QueueMsgID = msgID

// 	response["result"] = "GST Details would be updated"
// 	sendResponse(w, 200, "SUCCESS", "", response, logs)
// 	return
// }

func challanHand(user string, w http.ResponseWriter, logs *Masterindiacontrols) {
	response := make(map[string]interface{})
	gstinNumber := logs.RequestData.Gst

	defer func() {
		if panicCheck := recover(); panicCheck != nil {
			stack := string(debug.Stack())
			fmt.Println(stack)
			logs.StackTrace = stack
			sendResponse(w, 500, "FAILURE", "Panic...Pls inform Dev Team", response, logs)
			return
		}
	}()

	if len(gstinNumber) < 13 {
		sendResponse(w, 400, "FAILURE", "Param missing", response, logs)
		return
	}

	// Send acknowledgment immediately
	response["result"] = "GST Details would be updated"
	sendResponse(w, 200, "SUCCESS", "", response, logs)

	// Run the remaining code in the background
	go func() {
		if logs.RequestData.Modid == "merp" || logs.RequestData.Modid == "gladmin" {
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

			if logs.RequestData.Modid == "merp" || logs.RequestData.Modid == "gladmin" {
				err2 := SubcriberHandler2(string(raw))
				if err2 != nil {
					fmt.Println("Subscriber Handler2 error:", err2)
				}
				return
			}

			enqData := make(map[string]string)
			enqData["publisher"] = "centralizedAPI"
			enqData["jsonDataStr"] = string(raw)
			enqData["msgBody"] = enqData["publisher"]

			msgID, err := queue.Send(enqData)
			if err != nil {
				logs.StackTrace = err.Error()
				fmt.Println("Queue send error:", err)
				return
			}

			logs.QueueMsgID = msgID
		}
	}()
}

func seller(w http.ResponseWriter, logs *Masterindiacontrols, user string) {

	response := make(map[string]interface{})

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
	dbData, err := model.GetGSTRecords(database, s3log.Gst)
	en := utils.GetTimeInNanoSeconds()
	fmt.Println("GetGSTRecords", (en-st)/1000000)

	if err != nil {
		s3log.APIHit = err.Error()
		sendResponse(w, 200, "FAILURE", "error in fetching gst records", response, logs)
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	dbGst, _ := dbData["gstin_number"].(string)
	dbGstInsertionDate, _ := dbData["gst_insertion_date"].(string)
	dbGstinStatus, _ := dbData["gstin_status"].(string)

	s3log.Result["gst_insertion_date_in_db"] = dbGstInsertionDate

	gapOfDays := DaysDiff(dbGstInsertionDate)
	if (dbGst == s3log.Gst) && ((dbGstinStatus == "Active" && gapOfDays <= 30) || (dbGstinStatus != "Active" && gapOfDays <= 1)) {

		s3log.APIHit = "GST Details already fetched successfully within 30 days"

		dbGstGroupid, _ := dbData["business_constitution_group_id"].(string)
		legalStatusValue := utils.LegalStatusRead(dbGstGroupid)
		response["business_constitution"] = legalStatusValue

		response["gstin_status"] = dbGstinStatus
		sendResponse(w, 200, "SUCCESS", "", response, logs)
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
		sendResponse(w, 200, "FAILURE", "error in fetching masterindia tokken", response, logs)
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
		} else {
			s3log.Result["api_error"] = fmt.Sprint(data)
		}

		apiDataMsg, _ := data["message"].(string)
		apiDataStr, _ := data["data"].(string)

		if strings.ToLower(apiDataMsg) == "the gstin passed in the request is invalid." || strings.ToLower(apiDataStr) == "the gstin passed in the request is invalid." {

			sendResponse(w, 200, "FAILURE", "the gstin passed in the request is invalid.", response, logs)
			Write2S3(&s3log)
			Write2Kibana(&s3log)
			return
		}

		sendResponse(w, 200, "FAILURE", "error from masterindia api", response, logs)
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	var params []interface{}
	var myerror error
	var ApiDataMap map[string]string

	st = utils.GetTimeInNanoSeconds()
	// s3log.Result["api_data"], params = utils.BusLogicOnMasterData_V2(s3log.Gst, apiData)
	//changes started
	ApiDataMap, params = utils.BusLogicOnMasterData_V2(s3log.Gst, apiData)
	ApiDataMap_Copy := ApiDataMap
	en = utils.GetTimeInNanoSeconds()
	fmt.Println("BusLogicOnMasterData_V2", (en-st)/1000000)

	st = utils.GetTimeInNanoSeconds()
	if dbGst == s3log.Gst {
		s3log.Result["i_u_d"] = "U"
		_, myerror = model.UpdateGSTMasterData(database, params)
	} else {

		s3log.Result["i_u_d"] = "I"
		_, myerror = model.InsertGSTMasterData(database, params)
	}
	en = utils.GetTimeInNanoSeconds()
	fmt.Println("i_u", (en-st)/1000000)

	if myerror != nil {
		s3log.Result["i_u_d_error"] = myerror.Error()
		sendResponse(w, 200, "FAILURE", "error in i_u records in db", response, logs)
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	//integration of pubapi

	// data_glid, err := authadvance.GetGlidFromGstM(database, s3log.Gst)
	// if err != nil {
	// 	s3log.Result["fetch_glidfromgst_error"] = err.Error()
	// } else {
	// 	for _, glid := range data_glid {
	// 		if err := authadvance.ProcessSingleGLIDPubapilogging(glid, "/masterindia/v1/gst"); err != nil {
	// 			// logg.AnyError["meshupdationerror"] = err.Error()
	// 			s3log.Result["meshupdationerror_glid"] = glid + "-" + err.Error()
	// 		}
	// 	}
	// }

	if err := authadvance.ProcessingGST(s3log.Gst, "/masterindia/v1/gst"); err != nil {
		// logg.AnyError["meshupdationerror"] = err.Error()
		s3log.Result["meshupdationerror_gst"] = s3log.Gst + "-" + err.Error()
	}

	// s3logResultAPIData, _ := s3log.Result["api_data"].(map[string]string)
	s3logResultAPIData := ApiDataMap
	response["gstin_status"] = s3logResultAPIData["gstin_status"]

    Gstgroupid := s3logResultAPIData["business_constitution_group_id"]
	legalStatusValue2 := utils.LegalStatusRead(Gstgroupid)
	response["business_constitution"] = legalStatusValue2

	sendResponse(w, 200, "SUCCESS", "", response, logs)
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
	//
	Write2S3(&s3log)
	return
}

func merp(w http.ResponseWriter, logs *Masterindiacontrols, user string) {

	response := make(map[string]interface{})

	var s3log S3Log
	s3log.APIName = logs.RequestData.API

	credential := utils.GetCred(user)
	s3log.APIUserID = credential["username"]
	s3log.Gst = logs.RequestData.Gst
	s3log.APIHit = ""
	s3log.Modid = logs.RequestData.Modid
	s3log.RqstTime = logs.RequestStart
	s3log.Result = make(map[string]interface{})

	response["result"] = "GST Details would be updated"
	sendResponse(w, 200, "SUCCESS", "", response, logs)
	st := utils.GetTimeInNanoSeconds()
	//Logging Read Responose Time

	startread := time.Now()

	dbData, err := model.GetGSTRecords(database, s3log.Gst)

	elapsedread := time.Since(startread)

	logs.ReadResponseTime = fmt.Sprint(elapsedread.Milliseconds())

	//Logging Ends
	en := utils.GetTimeInNanoSeconds()
	fmt.Println("GetGSTRecords", (en-st)/1000000)
	//response["result"] = "GST Details would be updated"

	if err != nil {
		s3log.APIHit = err.Error()
		//sendResponse(w, 200, "FAILURE", "error in fetching gst records", response, logs)
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	dbGst, _ := dbData["gstin_number"].(string)
	dbGstInsertionDate, _ := dbData["gst_insertion_date"].(string)
	dbGstinStatus, _ := dbData["gstin_status"].(string)

	s3log.Result["gst_insertion_date_in_db"] = dbGstInsertionDate

	gapOfDays := DaysDiff(dbGstInsertionDate)
	if (dbGst == s3log.Gst) && ((dbGstinStatus == "Active" && gapOfDays <= 30) || (dbGstinStatus != "Active" && gapOfDays <= 1)) {

		s3log.APIHit = "GST Details already fetched successfully within 30 days"
		// response["gstin_status"] = dbGstinStatus

		//sendResponse(w, 200, "SUCCESS", "", response, logs)
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

		//sendResponse(w, 200, "FAILURE", "error in fetching masterindia tokken", response, logs)
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

			if ErrorParam2[1] == 102 || ErrorParam2[1] == 104 || ErrorParam2[1] == 107 || ErrorParam2[1] == 112 {
				fmt.Println("ERROR Code that won't be inserted: ", ErrorParam2[1])
			} else {
				_, Mastererr2 := model.InsertGSTMasterErrorData(database, ErrorParam2)
				if Mastererr2 != nil {
					s3log.Result["table_error"] = Mastererr2.Error()
				}
				//end
			}
		} else {
			s3log.Result["api_error"] = fmt.Sprint(data)
			//new changes -start
			error_data := fmt.Sprint(data)
			ErrorParam3 := utils.GetErrorParams(s3log.Gst, error_data)
			if ErrorParam3[1] == 102 || ErrorParam3[1] == 104 || ErrorParam3[1] == 107 || ErrorParam3[1] == 112 {
				fmt.Println("ERROR Code that won't be inserted: ", ErrorParam3[1])
			} else {
				_, Mastererr3 := model.InsertGSTMasterErrorData(database, ErrorParam3)
				if Mastererr3 != nil {
					s3log.Result["table_error"] = Mastererr3.Error()
				}
				//end
			}
		}

		apiDataMsg, _ := data["message"].(string)
		apiDataStr, _ := data["data"].(string)

		if strings.ToLower(apiDataMsg) == "the gstin passed in the request is invalid." || strings.ToLower(apiDataStr) == "the gstin passed in the request is invalid." {

			//sendResponse(w, 200, "FAILURE", "the gstin passed in the request is invalid.", response, logs)
			Write2S3(&s3log)
			Write2Kibana(&s3log)
			return
		}

		//sendResponse(w, 200, "FAILURE", "error from masterindia api", response, logs)
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
		//sendResponse(w, 200, "FAILURE", "error in i_u records in db", response, logs)
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	// data_glid, err := authadvance.GetGlidFromGstM(database, s3log.Gst)
	// if err != nil {
	// 	s3log.Result["fetch_glidfromgst_error"] = err.Error()
	// } else {
	// 	for _, glid := range data_glid {
	// 		if err := authadvance.ProcessSingleGLIDPubapilogging(glid, "/masterindia/v1/gst"); err != nil {
	// 			// logg.AnyError["meshupdationerror"] = err.Error()
	// 			s3log.Result["meshupdationerror_glid"] = glid + "-" + err.Error()
	// 		}
	// 	}
	// }

	if err := authadvance.ProcessingGST(s3log.Gst, "/masterindia/v1/gst"); err != nil {
		// logg.AnyError["meshupdationerror"] = err.Error()
		s3log.Result["meshupdationerror_gst"] = s3log.Gst + "-" + err.Error()
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
	//

	//sendResponse(w, 200, "SUCCESS", "", response, logs)
	Write2S3(&s3log)
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

func sendResponse(w http.ResponseWriter, httpcode int, status string, errorMsg string, response map[string]interface{}, logs *Masterindiacontrols) {

	var serviceResponse ResponseMasterindiacontrols
	serviceResponse.Code = httpcode
	serviceResponse.Status = status
	serviceResponse.ErrMessage = errorMsg
	serviceResponse.Body = response
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
