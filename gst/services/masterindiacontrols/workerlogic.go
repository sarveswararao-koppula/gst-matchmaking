package masterindiacontrols

import (
	"encoding/json"
	"errors"
	"fmt"
	api "mm/api/thirdpartyapi"
	model "mm/model/masterindiamodel"
	"mm/properties"
	authadvance "mm/services/authbridgeadvanced"
	"mm/utils"
	"os"
	"strconv"
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

// SubcriberHandler ...
func SubcriberHandler(data string) error {

	wr := WorkRequest{}
	err := json.Unmarshal([]byte(data), &wr)
	if err != nil {
		return err
	}

	if wr.APIName == "challan" {
		Challan(wr)
	} else if wr.APIName == "masterindia" {
		masterindia(wr)
	}
	return nil
}

// SubcriberHandler ...
func SubcriberHandler2(data string) error {

	wr := WorkRequest{}
	err := json.Unmarshal([]byte(data), &wr)
	if err != nil {
		return err
	}

	if wr.APIName == "challan" {
		Challan2(wr)
	} else if wr.APIName == "masterindia" {
		masterindia(wr)
	}
	return nil
}

// Challan ...
func Challan(work WorkRequest) {

	var s3log S3Log
	s3log.APIName = work.APIName
	user := work.APIUserName
	credential := utils.GetCred(user)

	s3log.APIUserID = credential["username"]
	s3log.Gst = work.GstPan
	s3log.APIHit = ""
	s3log.Modid = work.Modid
	s3log.RqstTime = work.RqstTime
	s3log.Result = make(map[string]interface{})

	fys := make([]fy, 4)
	getFY(fys)
	err := fyAPIStatus(fys, s3log.Gst)

	if err != nil {
		s3log.APIHit = err.Error()
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	err = ValidateTokken(credential, user)
	if err != nil {
		s3log.Result["api_error"] = err.Error()
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	fmt.Println("tokken", Tokkens[user])

	regData, err := model.GetExistsMasterdata(database, s3log.Gst)

	if err != nil {
		s3log.APIHit = err.Error()
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}
	//latest filing date for all fys
	fyLatestFilingOrig := ""
	regisDateStr := ""
	if len(regData) > 0 {
		fyLatestFilingOrig = regData["date_of_filing"]
		regisDateStr = regData["registration_date"]
	}
	fyLatestFiling := fyLatestFilingOrig

	//fmt.Println("regisDateStr", regisDateStr)
	if regisDate, err := time.Parse("02-01-2006", regisDateStr); err == nil {

		regYear := -1

		var mmm time.Month
		regYear, mmm, _ = regisDate.Date()
		if int(mmm) < 4 {
			regYear--
		}

		if regYear > 0 {
			for i, fy := range fys {
				if fy.start < regYear {
					fy.apihit = false
					fys[i] = fy
				}
			}
		}
		//fmt.Println("regYear", regYear)
	}

	gstinNumber := work.GstPan
	for _, fy := range fys {

		fmt.Println(fy)
		// if fy.fy == "fy0" {
		//      curFyLatestFiling = fy.maxDof
		// }

		fyLatestFiling = getMaxDate(fyLatestFiling, fy.maxDof)

		var s3log S3Log
		s3log.APIName = work.APIName
		s3log.APIUserID = credential["username"]

		fyYear := strconv.Itoa(fy.start) + "-" + strconv.Itoa(fy.end%100)

		s3log.Gst = gstinNumber + "#" + fyYear //gst#2020-21

		s3log.APIHit = "N"
		s3log.Modid = work.Modid
		s3log.RqstTime = work.RqstTime
		s3log.Result = make(map[string]interface{})

		s3log.Result["max_dof"] = fy.maxDof

		if !fy.apihit {
			Write2S3(&s3log)
			Write2Kibana(&s3log)
			continue
		}

		s3log.APIHit = "Y"
		data, err := api.GetChallanData(gstinNumber, fyYear, map[string]string{
			"client_id":    credential["client_id"],
			"access_token": Tokkens[user].Tok,
		})

		apiData, ok := data["data"].(map[string]interface{})
		dataErrorBool, _ := data["error"].(bool)
		dataErrorStr, _ := data["error"].(string)

		if err != nil || !ok || dataErrorBool || dataErrorStr != "" {

			if err != nil {
				s3log.Result["api_error"] = err.Error()
			} else {
				s3log.Result["api_error"] = fmt.Sprint(data)
			}

			if dataErrorStr == "invalid_grant" {
				Tokkens[user].Exp = time.Now().In(loc).Add(-24 * time.Hour)
			}

			//No DATA FOUND
			dataError, _ := apiData["error"].(map[string]interface{})
			errMessage, _ := dataError["message"].(string)
			errMessage = strings.ToLower(errMessage)

			if (fy.fy == "fy2" || fy.fy == "fy3") && (errMessage == "please select a valid financial year") {

				apiDataClosure := make(map[string]interface{})
				var chllandataArr []interface{}

				fakeChllandata := make(map[string]interface{})
				fakeChllandata["arn"] = ""
				fakeChllandata["dof"] = ""
				fakeChllandata["mof"] = ""
				fakeChllandata["ret_prd"] = 40000 + fy.start //42019
				fakeChllandata["rtntype"] = ""
				fakeChllandata["status"] = "Data Not Found"
				fakeChllandata["valid"] = "N"

				apiDataClosure["data"] = fakeChllandata

				_, err := model.InsertChallanDetails(database, gstinNumber, fakeChllandata, "1")
				if err != nil {
					apiDataClosure["i_u_d_error"] = err.Error()
				}

				chllandataArr = append(chllandataArr, apiDataClosure)

				s3log.Result["apiData"] = chllandataArr
			}
			Write2S3(&s3log)
			Write2Kibana(&s3log)
			continue
		}

		//DATA FOUND

		EFiledlist, _ := apiData["EFiledlist"].([]interface{})

		var chllandataArr []interface{}

		for _, v := range EFiledlist {

			apiDataClosure := make(map[string]interface{})

			chllandata, ok := v.(map[string]interface{})
			if ok {

				// if fy.fy == "fy0" {
				//      dofStr, _ := chllandata["dof"].(string)
				//      curFyLatestFiling = getMaxDate(curFyLatestFiling, dofStr)
				// }

				dofStr, _ := chllandata["dof"].(string)
				fyLatestFiling = getMaxDate(fyLatestFiling, dofStr)

				apiDataClosure["i_u_d"] = "I"
				_, err := model.InsertChallanDetails(database, gstinNumber, chllandata)

				if err != nil {
					apiDataClosure["i_u_d_error"] = err.Error()
				}
			}

			apiDataClosure["data"] = chllandata

			chllandataArr = append(chllandataArr, apiDataClosure)
		}

		//changes started
		chllandataArrCopy := chllandataArr

		jsonBytes, err := json.Marshal(chllandataArrCopy)
		if err != nil {
			fmt.Println("Error marshalling the data:", err)
			// return
		} else {
			// jsonString := string(jsonBytes)
			// s3log.Result["LatestapiData"]=fmt.Sprintf("\"%s\"", jsonString)
			// Write2Kibana(&s3log)
			str := string(jsonBytes)
			str = strings.ReplaceAll(str, "\"", "")
			str = strings.ReplaceAll(str, "{", "")
			str = strings.ReplaceAll(str, "}", "")
			str = strings.ReplaceAll(str, ",", ", ")
			s3log.Result["LatestapiData"] = str
			Write2Kibana(&s3log)
		}

		//changes ended
		s3log.Result["apiData"] = chllandataArr
		Write2S3(&s3log)
		continue

	}

	fmt.Println("fyLatestFilingOrig= ", fyLatestFilingOrig)
	fmt.Println("fyLatestFiling= ", fyLatestFiling)

	if flDate, err1 := time.Parse("02-01-2006", fyLatestFiling); err1 == nil {

		flDateOrig, err2 := time.Parse("02-01-2006", fyLatestFilingOrig)

		if err2 != nil || (err2 == nil && flDate.Sub(flDateOrig) > 0) {

			fyLatestFiling = flDate.Format("2006-01-02")

			var params []interface{}
			now := time.Now().In(loc).Format("2006-01-02 15:04:05")

			params = append(params, gstinNumber, fyLatestFiling, now)

			errU := model.UpdateMasterDataFilingDate(database, params)
			fmt.Println("err_u", errU)

		}

	}

	return
}

// Challan2 ...
func Challan2(work WorkRequest) {

	var s3log S3Log
	s3log.APIName = work.APIName
	user := work.APIUserName
	credential := utils.GetCred(user)

	s3log.APIUserID = credential["username"]
	s3log.Gst = work.GstPan
	s3log.APIHit = ""
	s3log.Modid = work.Modid
	s3log.RqstTime = work.RqstTime
	s3log.Result = make(map[string]interface{})

	fys := make([]fy, 4)
	getFY(fys)
	err := fyAPIStatus(fys, s3log.Gst)

	if err != nil {
		s3log.APIHit = err.Error()
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	err = ValidateTokken(credential, user)
	if err != nil {
		s3log.Result["api_error"] = err.Error()
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	fmt.Println("tokken", Tokkens[user])

	regData, err := model.GetExistsMasterdata(database, s3log.Gst)

	if err != nil {
		s3log.APIHit = err.Error()
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}
	//latest filing date for all fys
	fyLatestFilingOrig := ""
	regisDateStr := ""
	if len(regData) > 0 {
		fyLatestFilingOrig = regData["date_of_filing"]
		regisDateStr = regData["registration_date"]
	}
	fyLatestFiling := fyLatestFilingOrig

	//fmt.Println("regisDateStr", regisDateStr)
	if regisDate, err := time.Parse("02-01-2006", regisDateStr); err == nil {

		regYear := -1

		var mmm time.Month
		regYear, mmm, _ = regisDate.Date()
		if int(mmm) < 4 {
			regYear--
		}

		if regYear > 0 {
			for i, fy := range fys {
				if fy.start < regYear {
					fy.apihit = false
					fys[i] = fy
				}
			}
		}
		//fmt.Println("regYear", regYear)
	}

	fy0errorflag := 0 //no recordfound error

	gstinNumber := work.GstPan
	for _, fy := range fys {

		fmt.Println(fy)
		// if fy.fy == "fy0" {
		//      curFyLatestFiling = fy.maxDof
		// }

		fyLatestFiling = getMaxDate(fyLatestFiling, fy.maxDof)

		var s3log S3Log
		s3log.APIName = work.APIName
		s3log.APIUserID = credential["username"]

		fyYear := strconv.Itoa(fy.start) + "-" + strconv.Itoa(fy.end%100)

		s3log.Gst = gstinNumber + "#" + fyYear //gst#2020-21

		s3log.APIHit = "N"
		s3log.Modid = work.Modid
		s3log.RqstTime = work.RqstTime
		s3log.Result = make(map[string]interface{})

		s3log.Result["max_dof"] = fy.maxDof

		if (fy0errorflag == 0 && fy.fy == "fy1" &&  work.Modid == "merp") || fy.fy == "fy2" || fy.fy == "fy3" {
			s3log.APIHit = "N"
			Write2S3(&s3log)
			Write2Kibana(&s3log)
			continue
		}

		if !fy.apihit {
			Write2S3(&s3log)
			Write2Kibana(&s3log)
			continue
		}

		s3log.APIHit = "Y"
		data, err := api.GetChallanData(gstinNumber, fyYear, map[string]string{
			"client_id":    credential["client_id"],
			"access_token": Tokkens[user].Tok,
		})

		apiData, ok := data["data"].(map[string]interface{})
		dataErrorBool, _ := data["error"].(bool)
		dataErrorStr, _ := data["error"].(string)

		if err != nil || !ok || dataErrorBool || dataErrorStr != "" {
            fy0errorflag = 1
			if err != nil {
				s3log.Result["api_error"] = err.Error()
			} else {
				s3log.Result["api_error"] = fmt.Sprint(data)
			}

			if dataErrorStr == "invalid_grant" {
				Tokkens[user].Exp = time.Now().In(loc).Add(-24 * time.Hour)
			}

			//No DATA FOUND
			dataError, _ := apiData["error"].(map[string]interface{})
			errMessage, _ := dataError["message"].(string)
			errMessage = strings.ToLower(errMessage)

			if (fy.fy == "fy2" || fy.fy == "fy3") && (errMessage == "please select a valid financial year") {

				apiDataClosure := make(map[string]interface{})
				var chllandataArr []interface{}

				fakeChllandata := make(map[string]interface{})
				fakeChllandata["arn"] = ""
				fakeChllandata["dof"] = ""
				fakeChllandata["mof"] = ""
				fakeChllandata["ret_prd"] = 40000 + fy.start //42019
				fakeChllandata["rtntype"] = ""
				fakeChllandata["status"] = "Data Not Found"
				fakeChllandata["valid"] = "N"

				apiDataClosure["data"] = fakeChllandata

				_, err := model.InsertChallanDetails(database, gstinNumber, fakeChllandata, "1")
				if err != nil {
					apiDataClosure["i_u_d_error"] = err.Error()
				}

				chllandataArr = append(chllandataArr, apiDataClosure)

				s3log.Result["apiData"] = chllandataArr
			}

			Write2Kibana(&s3log)
			Write2S3(&s3log)
			continue
		}

		//DATA FOUND

		EFiledlist, _ := apiData["EFiledlist"].([]interface{})

		var chllandataArr []interface{}

		for _, v := range EFiledlist {

			apiDataClosure := make(map[string]interface{})

			chllandata, ok := v.(map[string]interface{})
			if ok {

				// if fy.fy == "fy0" {
				//      dofStr, _ := chllandata["dof"].(string)
				//      curFyLatestFiling = getMaxDate(curFyLatestFiling, dofStr)
				// }

				dofStr, _ := chllandata["dof"].(string)
				fyLatestFiling = getMaxDate(fyLatestFiling, dofStr)

				apiDataClosure["i_u_d"] = "I"

				// Check for duplicate entries before inserting
				rowcount, err := model.CheckChallanDetails_dupl(database, gstinNumber, chllandata)

				if err != nil {
					apiDataClosure["i_u_d_error"] = err.Error()
				} else if rowcount == 0 {
					_, err := model.InsertChallanDetails(database, gstinNumber, chllandata)
					if err != nil {
						apiDataClosure["i_u_d_error"] = err.Error()
					}
				} else {
					apiDataClosure["i_u_d_error"] = "Duplicate entry found, insertion skipped"
				}

				// _, err := model.InsertChallanDetails(database, gstinNumber, chllandata)

				// if err != nil {
				// 	apiDataClosure["i_u_d_error"] = err.Error()
				// }
			}

			apiDataClosure["data"] = chllandata

			chllandataArr = append(chllandataArr, apiDataClosure)
		}

		//changes started
		chllandataArrCopy := chllandataArr

		jsonBytes, err := json.Marshal(chllandataArrCopy)
		if err != nil {
			fmt.Println("Error marshalling the data:", err)
			// return
		} else {
			// jsonString := string(jsonBytes)
			// s3log.Result["LatestapiData"]=fmt.Sprintf("\"%s\"", jsonString)
			// Write2Kibana(&s3log)
			str := string(jsonBytes)
			str = strings.ReplaceAll(str, "\"", "")
			str = strings.ReplaceAll(str, "{", "")
			str = strings.ReplaceAll(str, "}", "")
			str = strings.ReplaceAll(str, ",", ", ")
			s3log.Result["LatestapiData"] = str
			Write2Kibana(&s3log)
		}
		//changes ended

		s3log.Result["apiData"] = chllandataArr
		Write2S3(&s3log)
		continue

	}

	fmt.Println("fyLatestFilingOrig= ", fyLatestFilingOrig)
	fmt.Println("fyLatestFiling= ", fyLatestFiling)

	if flDate, err1 := time.Parse("02-01-2006", fyLatestFiling); err1 == nil {

		flDateOrig, err2 := time.Parse("02-01-2006", fyLatestFilingOrig)

		if err2 != nil || (err2 == nil && flDate.Sub(flDateOrig) > 0) {

			fyLatestFiling = flDate.Format("2006-01-02")

			var params []interface{}
			now := time.Now().In(loc).Format("2006-01-02 15:04:05")

			params = append(params, gstinNumber, fyLatestFiling, now)

			errU := model.UpdateMasterDataFilingDate(database, params)
			fmt.Println("err_u", errU)

		}

	}

	return
}

func masterindia(work WorkRequest) {

	var s3log S3Log
	s3log.APIName = work.APIName
	_, st1 := utils.GetExecTime()
	user := work.APIUserName
	credential := utils.GetCred(user)

	s3log.APIUserID = credential["username"]
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
	dbGstInsertionDate, _ := dbData["gst_insertion_date"].(string)
	dbGstinStatus, _ := dbData["gstin_status"].(string)

	s3log.Result["gst_insertion_date_in_db"] = dbGstInsertionDate

	gapOfDays := DaysDiff(dbGstInsertionDate)

	if work.Modid != "gladmin" && (dbGst == s3log.Gst) && ((dbGstinStatus == "Active" && gapOfDays <= 30) || (dbGstinStatus != "Active" && gapOfDays <= 1)) {
		s3log.APIHit = "GST Details already fetched successfully within 30 days"
		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	_, st = utils.GetExecTime()

	err = ValidateTokken(credential, user)

	s3log.ExecTime["validatetokken_time"], _ = utils.GetExecTime(st)

	fmt.Println("tokken", Tokkens[user])

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

		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return
	}

	s3log.APIHit = "Y"

	_, st = utils.GetExecTime() //

	data, err := api.GetMasterData(s3log.Gst, map[string]string{
		"client_id":    credential["client_id"],
		"access_token": Tokkens[user].Tok,
	})

	s3log.ExecTime["GetMasterData_time"], _ = utils.GetExecTime(st) //

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

		if dataErrorStr == "invalid_grant" {
			Tokkens[user].Exp = time.Now().In(loc).Add(-24 * time.Hour)
			/*
			   //new changes -start
			           ErrorParam4:=utils.GetErrorParams(s3log.Gst,dataErrorStr)
			           if ErrorParam4[1]==102 || ErrorParam4[1]==104 || ErrorParam4[1]==107 || ErrorParam4[1]==112 {
			                   fmt.Println("ERROR Code that won't be inserted: ",ErrorParam4[1])
			           }else{
			                   _,Mastererr4:=model.InsertGSTMasterErrorData(database,ErrorParam4)

			                   if Mastererr4!=nil{
			                           s3log.Result["table_error"] = Mastererr4.Error()
			                   }
			                   //end
			           }
			*/
		}

		Write2S3(&s3log)
		Write2Kibana(&s3log)
		return

	}

	var params []interface{}
	var myerror error

	var ApiDataMap map[string]string

	_, st = utils.GetExecTime() //

	//changes started
	// s3log.Result["apiData"], params = utils.BusLogicOnMasterData_V2(s3log.Gst, apiData)

	ApiDataMap, params = utils.BusLogicOnMasterData_V2(s3log.Gst, apiData)

	ApiDataMap_Copy := ApiDataMap

	// s3log.Result["NewapiData"]

	s3log.ExecTime["BusLogicMasterData_time"], _ = utils.GetExecTime(st) //

	if dbGst == s3log.Gst {

		s3log.Result["i_u_d"] = "U"

		_, st = utils.GetExecTime() //

		_, myerror = model.UpdateGSTMasterData(database, params)

		s3log.ExecTime["UpdateGSTMasterData_time"], _ = utils.GetExecTime(st) //

	} else {

		s3log.Result["i_u_d"] = "I"

		_, st = utils.GetExecTime() //

		_, myerror = model.InsertGSTMasterData(database, params)

		s3log.ExecTime["InsertGSTMasterData_time"], _ = utils.GetExecTime(st) //
	}

	if myerror != nil {
		s3log.Result["i_u_d_error"] = myerror.Error()
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

	s3log.ExecTime["Total_Execution_time"], _ = utils.GetExecTime(st1)

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

	//fmt.Println("sarvesh write2s3"
	return
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

// ValidateTokken ...
func ValidateTokken(credential map[string]string, user string) error {

	var generateTokken bool = false

	if Tokkens[user] == nil {
		generateTokken = true
	} else if Tokkens[user].Exp.IsZero() {
		generateTokken = true
	} else if int(time.Now().In(loc).Sub(Tokkens[user].Exp).Hours()) > 4 {
		generateTokken = true
	}

	if generateTokken {
		tok, e := api.GetTokken(credential)
		if e != nil {
			return e
		}

		accessToken, ok := tok["access_token"].(string)
		if ok {
			Tokkens[user] = &Tokken{accessToken, time.Now().In(loc)}
			return nil
		}
		return errors.New("err in generating access_token")
	}

	return nil

}

// DaysDiff ...
func DaysDiff(gstInsertionDateStr string) float64 {
	//days_check, _ := strconv.ParseFloat(days_check_str, 64)
	//days_check = math.Max(days_check, float64(30))
	if gstInsertionDateStr == "" {
		return 99999.0
	}

	gstInsertionDate, err := time.Parse("02-01-2006 15:04:05", gstInsertionDateStr)

	if err == nil {
		nowStr := time.Now().In(loc).Format("02-01-2006 15:04:05")
		now, _ := time.Parse("02-01-2006 15:04:05", nowStr)
		return float64(now.Sub(gstInsertionDate).Hours() / 24)
	}
	return 99999.0
}

func getFY(fys []fy) {

	today := time.Now().In(loc)
	year := today.Year()
	month := int(today.Month())

	if month >= 4 {
		year++
	}

	for i := 0; i < 4; i++ {
		fys[i] = fy{
			"fy" + strconv.Itoa(i), year - i - 1, year - i, true, "",
		}
	}
}

func fyAPIStatus(fys []fy, gstinNumber string) error {

	var params []interface{}

	params = append(params, fys[0].end, fys[1].end, fys[2].end, fys[3].end, fys[3].start, gstinNumber)
	data, err := model.CheckChallanDetails(database, params)
	if err != nil {
		return err
	}

	for _, v := range data {
		row, _ := v.(map[string]interface{})

		dateOfFiling, _ := row["dateOfFiling"].(string)
		fy, _ := row["fy"].(string)
		fy = strings.ToLower(fy)

		if fy == "fy0" {
			fys[0].maxDof = dateOfFiling

			if dateOfFiling != "" && diffDays(dateOfFiling) <= float64(90) {
				fys[0].apihit = false
			}

		} else if fy == "fy1" {

			fys[1].maxDof = dateOfFiling

			if dateOfFiling != "" && diffDays(dateOfFiling) <= float64(90) {
				fys[1].apihit = false
			}

		} else if fy == "fy2" {

			fys[2].maxDof = dateOfFiling
			fys[2].apihit = false

		} else if fy == "fy3" {

			fys[3].maxDof = dateOfFiling
			fys[3].apihit = false
		}

	}

	return nil
}

func diffDays(dateOfFilingStr string) float64 {
	nowStr := time.Now().In(loc).Format("02-01-2006")
	now, _ := time.Parse("02-01-2006", nowStr)
	dateOfFiling, _ := time.Parse("02-01-2006", dateOfFilingStr)
	return float64(now.Sub(dateOfFiling).Hours() / 24)
}

func getMaxDate(curFyLatestFilingStr string, filingDateStr string) string {

	filingDate, err1 := time.Parse("02-01-2006", filingDateStr)
	curFyLatestFiling, err2 := time.Parse("02-01-2006", curFyLatestFilingStr)

	if err1 == nil && err2 == nil {
		diff := filingDate.Sub(curFyLatestFiling).Hours()
		if diff > 0 {
			return filingDate.Format("02-01-2006")
		}
		return curFyLatestFiling.Format("02-01-2006")
	}

	if err1 != nil && err2 == nil {
		return curFyLatestFiling.Format("02-01-2006")
	}

	if err1 == nil && err2 != nil {
		return filingDate.Format("02-01-2006")
	}
	return ""
}
