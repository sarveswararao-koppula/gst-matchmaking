package livebefisccontrols

import (
	"encoding/json"
	"errors"
	"fmt"
	servapi "mm/api/servapi"
	api "mm/api/thirdpartyapi"
	model "mm/model/masterindiamodel"

	"mm/properties"
	authadvanced "mm/services/authbridgeadvanced"

	// authadvance "mm/services/authbridgeadvanced"
	"mm/utils"
	"os"
	"strings"
	"time"
)

// Tokken ...
type Tokken struct {
	Tok string    `json:"Tok,omitempty"`
	Exp time.Time `json:"Exp,omitempty"`
}

// Tokkens ...
var Tokkens map[string]*Tokken = make(map[string]*Tokken)

// S3Log ...
type S3Log struct {
	APIName   string                 `json:"APIName,omitempty"`
	APIUserID string                 `json:"APIUserID,omitempty"`
	AuthKey   string                 `json:"AuthKey,omitempty"`
	APIHit    string                 `json:"APIHit,omitempty"`
	Gst       string                 `json:"Gst,omitempty"`
	Modid     string                 `json:"Modid,omitempty"`
	RqstTime  string                 `json:"RqstTime,omitempty"`
	Result    map[string]interface{} `json:"Result,omitempty"`
	//ExecTime
	ExecTime    map[string]float64 `json:"ExecTime,omitempty"`
	RqstQEnTime string             `json:"RqstQEnTime,omitempty"`
}
type fy struct {
	fy     string
	start  int
	end    int
	apihit bool
	maxDof string
}

// BefiscSubcriberHandler ...
func BefiscSubcriberHandler(data string) error {

	// time.Sleep(10 * time.Second)

	wr := WorkRequest{}
	err := json.Unmarshal([]byte(data), &wr)
	if err != nil {
		return err
	}

	if wr.APIName == "befisc" {
		BefiscDataEnrichment(wr)
	}
	return nil
}

// BefiscDataEnrichment ...
func BefiscDataEnrichment(work WorkRequest) {

	var s3log S3Log
	s3log.APIName = work.APIName
	_, st1 := utils.GetExecTime()
	user := work.APIUserName
	credential := utils.GetCred(user)

	s3log.AuthKey = credential["authKey"]
	s3log.Gst = work.GstPan
	s3log.APIHit = ""
	s3log.Modid = work.Modid
	s3log.RqstTime = work.RqstTime
	s3log.RqstQEnTime = utils.GetTimeStampCurrent()
	s3log.Result = make(map[string]interface{})
	s3log.ExecTime = make(map[string]float64)

	_, st := utils.GetExecTime()

	dbData, err := model.GetGSTRecords(database, s3log.Gst)

	s3log.ExecTime["GetGSTRecords_time"], _ = utils.GetExecTime(st)

	if err != nil {
		s3log.APIHit = err.Error()
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	dbGst, _ := dbData["gstin_number"].(string)

	s3log.APIHit = "Y"
	_, st = utils.GetExecTime()
	data, BefiscRawJSON, err1 := api.GetBefiscGSTData(s3log.Gst, s3log.AuthKey)
	s3log.ExecTime["BefiscAPICall_time"], _ = utils.GetExecTime(st)
	if err1 != nil {
		s3log.Result["api_error"] = err1.Error()
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	s3log.Result["LatestapiData"] = BefiscRawJSON

	status, ok := data["status"].(float64)
	if ok {
		status1 := int(status)
		if status1 != 1 {
			msg := "Unknown error"
			if message, ok := data["message"].(string); ok {
				// s3log.Result["api_error"] = message
				msg = message
			}
			s3log.Result["Befis_api_error_message"] = msg

			Write2S3(&s3log)
			Write2Kibana(&s3log)
			return
		}
	}

	var params []interface{}
	var myerror error

	var ApiDataMap map[string]string

	_, st = utils.GetExecTime()

	ApiDataMap, params = utils.BusLogicOnMasterData_Befisc(s3log.Gst, data)

	s3log.ExecTime["BusLogicMasterData_time"], _ = utils.GetExecTime(st)

	ApiDataMap_Copy := ApiDataMap

	if dbGst == s3log.Gst {

		s3log.Result["i_u_d"] = "U"

		_, st = utils.GetExecTime() //

		_, myerror = model.UpdateGstBefiscData(database, params)

		s3log.ExecTime["UpdateGSTMasterData_time"], _ = utils.GetExecTime(st) //

	} else {

		s3log.Result["i_u_d"] = "I"

		_, st = utils.GetExecTime() //

		_, myerror = model.InsertGstBefiscData(database, params)

		s3log.ExecTime["InsertGSTMasterData_time"], _ = utils.GetExecTime(st) //
	}

	if myerror != nil {
		s3log.Result["i_u_d_error"] = myerror.Error()
	}

	turnover, ok := params[2].(string)
	if !ok {
		s3log.Result["TurnoverJsonUpdationError"] = "Turnover Missing in response"
	} else {
		err2 := authadvanced.TurnoverUpdation(turnover, s3log.Gst)
		if err2 != nil {
			s3log.Result["TurnoverJsonUpdationError"] = err.Error()
		}
	}

	result, ok := data["result"].(map[string]interface{})
	if ok {

		apiDataHSN, _ := result["business_details"].(map[string]interface{})
		//Bus Logic for Hsn Information
		_, hsnstring := utils.BusLogicOnBefiscHSN_V2(s3log.Gst, apiDataHSN)

		if len(hsnstring) > 0 {
			err := InsertHSNInfo(s3log.Gst, hsnstring)
			if err != nil {
				s3log.Result["HsnUpdationError"] = err.Error()
			}
		}

		//codeforchallan
		regData, err := model.GetExistsMasterdata(database, s3log.Gst)
	if err != nil {
		s3log.APIHit = err.Error()
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	
	//latest filing date for all fys
	fyLatestFilingOrig := ""
	if len(regData) > 0 {
		fyLatestFilingOrig = regData["date_of_filing"]
	}
	fyLatestFiling := fyLatestFilingOrig
// Access the "filing_status" field dynamically
	if filingStatus, ok := result["filing_status"].([]interface{}); ok {
		for _, outerArray := range filingStatus {
			if innerArray, ok := outerArray.([]interface{}); ok {
				for _, item := range innerArray {
					apiDataClosure := make(map[string]interface{})
					if record, ok := item.(map[string]interface{}); ok {
						dofStr, _ := record["dof"].(string)
				        fyLatestFiling = getMaxDate(fyLatestFiling, dofStr)
						apiDataClosure["i_u_d"] = "I"

				       // Check for duplicate entries before inserting
				        rowcount, err := model.CheckChallanDetails_dupl_befisc(database, s3log.Gst, record)
						if err != nil {
							apiDataClosure["i_u_d_error"] = err.Error()
						} else if rowcount == 0 {
							_, err := model.InsertChallanDetailsBefisc(database, s3log.Gst, record)
							if err != nil {
								apiDataClosure["i_u_d_error"] = err.Error()	
							}
						} else {
							apiDataClosure["i_u_d_error"] = "Duplicate entry found, insertion skipped"
						}
					}
				}
			}
		}
	}


    if flDate, err1 := time.Parse("02/01/2006", fyLatestFiling); err1 == nil {

		flDateOrig, err2 := time.Parse("02/01/2006", fyLatestFilingOrig)

		if err2 != nil || (err2 == nil && flDate.Sub(flDateOrig) > 0) {

			fyLatestFiling = flDate.Format("2006/01/02")

			var params []interface{}
			now := time.Now().In(loc).Format("2006/01/02 15:04:05")

			params = append(params, s3log.Gst, fyLatestFiling, now)

			errU := model.UpdateMasterDataFilingDate(database, params)
			fmt.Println("err_u", errU)

		}

	}

}

	if err := authadvanced.ProcessingGST(s3log.Gst, "/befisc/v1/gst"); err != nil {
		// logg.AnyError["meshupdationerror"] = err.Error()
		s3log.Result["meshupdationerror_gst"] = s3log.Gst + "-" + err.Error()
	}


	s3log.ExecTime["Total_Execution_time"], _ = utils.GetExecTime(st1)

	// Convert ApiDataMap_Copy (map[string]string) to JSON
	jsonBytes, err := json.Marshal(ApiDataMap_Copy)
	if err != nil {
		// handle the error gracefully
		s3log.Result["BusLogicData"] = fmt.Sprintf("Error converting to JSON: %v", err)
	} else {
		// Assign the JSON string to the same key (or a new key)
		s3log.Result["BusLogicData"] = string(jsonBytes)
	}

	Write2S3(&s3log)
	Write2Kibana(&s3log)

	return

}

func getMaxDate(curFyLatestFilingStr string, filingDateStr string) string {

	layout := "02/01/2006"

	filingDate, err1 := time.Parse(layout, filingDateStr)
	curFyLatestFiling, err2 := time.Parse(layout, curFyLatestFilingStr)

	if err1 == nil && err2 == nil {
		diff := filingDate.Sub(curFyLatestFiling).Hours()
		if diff > 0 {
			return filingDate.Format(layout)
		}
		return curFyLatestFiling.Format(layout)
	}

	if err1 != nil && err2 == nil {
		return curFyLatestFiling.Format(layout)
	}

	if err1 == nil && err2 != nil {
		return filingDate.Format(layout)
	}
	return ""
}


// InsertHSNInfo
func InsertHSNInfo(gst string, hsnstring string) (err error) {
	jsonStr, err := servapi.Hsnapi("PROD", gst, hsnstring)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return err
	}

	res := make(map[string]interface{})
	for k, v := range data {
		res[strings.ToLower(k)] = v
	}

	status, _ := res["status"].(string)

	if strings.ToLower(status) == "successful" {
		return nil
	}

	return errors.New(jsonStr)
}

// Write2S3 ...
func Write2S3(logs *S3Log) {

	logsDir := properties.Prop.LOG_MASTERINDIA + utils.TodayDir()

	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		e := os.MkdirAll(logsDir, os.ModePerm)
		fmt.Println(e)
	}

	logsDir += "/masterindia_wrapper_queue.json"

	fmt.Println(logsDir)

	jsonLog, _ := json.Marshal(*logs)

	jsonLogString := string(jsonLog[:len(jsonLog)])

	f, _ := os.OpenFile(logsDir, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()

	mutex.Lock()
	defer mutex.Unlock()
	f.WriteString("\n" + jsonLogString)
	return
}

// Write2Kibana ...
func Write2Kibana(logs *S3Log) {

	logsDir := properties.Prop.LOG_MASTERINDIA + utils.TodayDir()

	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		e := os.MkdirAll(logsDir, os.ModePerm)
		fmt.Println(e)
	}

	logsDir += "/masterindia_wrapper_worker_logs.json"

	fmt.Println(logsDir)

	jsonLog, _ := json.Marshal(*logs)

	jsonLogString := string(jsonLog[:len(jsonLog)])

	f, _ := os.OpenFile(logsDir, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()

	mutex.Lock()
	defer mutex.Unlock()
	f.WriteString("\n" + jsonLogString)
	return
}
