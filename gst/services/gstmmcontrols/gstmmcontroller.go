package gstmmcontrols

import (
	"encoding/json"
	"errors"
	"fmt"
	"mm/components/constants"
	"mm/properties"
	"mm/queue"

	//"mm/components/structures"
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

var Database string = properties.Prop.DATABASE
var logFileMutex sync.Mutex

// GetGST ...
func GetGST(w http.ResponseWriter, r *http.Request) {

	var logg Logg

	//arr := [10]string{"N1A", "N1B", "N1C", "N2A", "N2B", "N1", "N2", "N3", "N4", "N5"}
	logg.RequestStart = utils.GetTimeStampCurrent()
	logg.RequestStartValue = utils.GetTimeInNanoSeconds()
	logg.ServiceName = "gst_match_making"
	logg.ServiceURL = r.RequestURI
	logg.AnyError = make(map[string]string)
	logg.ExecTime = make(map[string]float64)

	_, st := utils.GetExecTime()

	if parts := strings.Split(r.RemoteAddr, ":"); len(parts) == 2 {
		logg.RemoteAddress = parts[0]
	}
	fmt.Println(logg.RemoteAddress, "Dev Testing")

	// if logg.RemoteAddress != constants.DATA_LEAD_ID && logg.RemoteAddress != constants.LOCAL_IP {
	// sendResponse(w, 401, "Blocked", GSTMatch{}, &logg, errors.New("request from not auth IP"))
	// WriteLog2(logg)
	// return
	// }

	fmt.Println(" logg.RemoteAddress : ", logg.RemoteAddress)
	fmt.Println(" constants.DATA_LEAD_ID : ", constants.DATA_LEAD_ID)
	fmt.Println(" constants.LOCAL_IP : ", constants.LOCAL_IP)

	logg.Request = Req{
		Glid:     r.URL.Query().Get("glid"),
		UniqueID: xid.New().String(),
	}

	defer func() {
		if panicCheck := recover(); panicCheck != nil {
			stack := string(debug.Stack())
			logg.StackTrace = stack
			fmt.Println(stack)
			sendResponse(w, 500, "Panic ..Pls inform Dev Team", GSTMatch{}, &logg, errors.New("Panic ..Pls inform Dev Team"))
			return
		}
	}()

	glidInt, _ := strconv.Atoi(logg.Request.Glid)
	fmt.Println(logg.Request.Glid, "Dev Testing")
	fmt.Println(glidInt, "RCA Testing", logg.RequestStart)
	if glidInt <= 0 {
		sendResponse(w, 400, "Invalid Glid", GSTMatch{}, &logg, errors.New("Invalid Glid"))
		return
	}

	logg.ExecTime["validation check"], _ = utils.GetExecTime(st)

	//Glusr Disabled Check

	_, st = utils.GetExecTime()

	glidDisabledData, err := GetDisabledGlidRecords(Database, glidInt)

	logg.ExecTime["glidDisabled check"], _ = utils.GetExecTime(st)

	if err != nil {
		sendResponse(w, 400, "Error in Disabled Query", GSTMatch{}, &logg, err)
		return
	}

	visited_Err_Msg := map[string]bool{
		"e-mail bounce":    true,
		"invalid cin-gstn": true,
	}

	if len(glidDisabledData) > 0 {
		disGlid := glidDisabledData["fk_glusr_usr_id"]
		fmt.Println(disGlid)
		errmsg_value := strings.ToLower(glidDisabledData["glusr_disable_errmsg_value"])
		fmt.Println(errmsg_value)
		if disGlid != "" {
			if !visited_Err_Msg[errmsg_value] {
				sendResponse(w, 200, "Glid is Disabled", GSTMatch{}, &logg, errors.New("Glid Disabled"))
				return
			}
		}
	}

	//END

	//start
	//cust_type==empFCP check and keywords check

	glidCustTypeCompanyName, err2 := GetCusttypeCompanyGlidRecords(Database, glidInt)

	if err2 != nil {
		sendResponse(w, 400, "Error in CustType,companyName Query", GSTMatch{}, &logg, err2)
		return
	}

	if len(glidCustTypeCompanyName) > 0 {
		var glusr_usr_id, gl_custtype_name, gl_companyname string
		glusr_usr_id = glidCustTypeCompanyName["glusr_usr_id"]
		gl_custtype_name = glidCustTypeCompanyName["glusr_usr_custtype_name"]
		gl_companyname = glidCustTypeCompanyName["glusr_usr_companyname"]
		fmt.Println("Initial gl companyname : ", gl_companyname)

		if glusr_usr_id != "" {
			if strings.ToLower(gl_custtype_name) == "empfcp" {
				logg.CustTypeFlag = gl_custtype_name
				sendResponse(w, 200, "CustType is empFCP", GSTMatch{}, &logg, errors.New("CustType is empFCP"))
				return
			}

			//changes started

			// check company_name contains below key words.
			// keywords case (Enterprise, Traders,Engineering, Store, Associate,Brothers)
			// flag_keyword := 0
			// Keywords := []string{"enterprise", "traders", "engineering", "store", "brothers", "associate"}
			// for _, s := range Keywords {
			//         if strings.Contains(strings.ToLower(gl_companyname), s) {
			//                 flag_keyword = 1
			//                 logg.KeyWordFlag = s
			//                 break
			//         }
			// }

			// if flag_keyword == 1 {
			//         sendResponse(w, 200, "glusr_companyname is having keyword "+logg.KeyWordFlag, GSTMatch{}, &logg, errors.New("glusr_companyname is having keyword "+logg.KeyWordFlag))
			//         return
			// }

			// changes ended

		}

	}
	//end

	//addition of contactdetails logic
	details, err := Contactdetails(logg.Request.Glid)
	if err != nil {
		// sendResponse(w, 200, "No GST Found", GSTMatch{}, &logg, err)
		logg.AnyError["Contactdetailsblockerror"] = err.Error()
		details = nil
	}

	if details != nil && len(details) > 0 {
		logg.ContactSource = "Y"
		gstDetails, _, _, exectime, arr, err := FindGstContactdetails(glidInt, details)
		for k, v := range exectime {
			logg.ExecTime[k] = v
		}
		//fmt.Println(err,"Dev Error Testing")
		//fmt.Println(gstDetails,"Dev-Testing-gstDetails")
		if err != nil {
			sendResponse(w, 200, "No GST Found", GSTMatch{}, &logg, err)
			return
		}
		logg.ScoreDetails = gstDetails.scores
		for i := 0; i < len(arr); i++ {
			if gstDetails.BucketName == arr[i] {
				logg.ScoreDetailsStg1 = gstDetails.scoresStage1
			}
		}
		sendResponse(w, 200, "SUCCESS", gstDetails, &logg, nil)
		return
	} else {
		logg.ContactSource = "N"
		gstDetails, _, _, exectime, arr, err := FindGst(glidInt)
		for k, v := range exectime {
			logg.ExecTime[k] = v
		}
		//fmt.Println(err,"Dev Error Testing")
		//fmt.Println(gstDetails,"Dev-Testing-gstDetails")
		if err != nil {
			sendResponse(w, 200, "No GST Found", GSTMatch{}, &logg, err)
			return
		}
		logg.ScoreDetails = gstDetails.scores
		for i := 0; i < len(arr); i++ {
			if gstDetails.BucketName == arr[i] {
				logg.ScoreDetailsStg1 = gstDetails.scoresStage1
			}
		}
		sendResponse(w, 200, "SUCCESS", gstDetails, &logg, nil)
		return
	}

	// gstDetails, _, _, exectime, arr, err := FindGst(glidInt)
	// for k, v := range exectime {
	// 	logg.ExecTime[k] = v
	// }
	// //fmt.Println(err,"Dev Error Testing")
	// //fmt.Println(gstDetails,"Dev-Testing-gstDetails")
	// if err != nil {
	// 	sendResponse(w, 200, "No GST Found", GSTMatch{}, &logg, err)
	// 	return
	// }
	// logg.ScoreDetails = gstDetails.scores
	// for i := 0; i < len(arr); i++ {
	// 	if gstDetails.BucketName == arr[i] {
	// 		logg.ScoreDetailsStg1 = gstDetails.scoresStage1
	// 	}
	// }
	// sendResponse(w, 200, "SUCCESS", gstDetails, &logg, nil)

	// return
}

func sendResponse(w http.ResponseWriter, code int, errMsg string, body GSTMatch, logg *Logg, err error) {

	_, st := utils.GetExecTime()

	w.Header().Set("Content-Type", "application/json")

	logg.Response = Res{
		Code: code,
		Err:  errMsg,
		Body: body,
	}

	if err != nil {
		logg.AnyError[errMsg] = err.Error()
	}

	json.NewEncoder(w).Encode(logg.Response)

	logg.RequestEndValue = utils.GetTimeInNanoSeconds()
	logg.ResponseTime = (logg.RequestEndValue - logg.RequestStartValue) / 1000000
	logg.RequestStartValue = 0
	logg.RequestEndValue = 0

	jsonLog, _ := json.Marshal(logg)

	fmt.Println("\n"+string(jsonLog), "Dev-Testing-Log")
	//gst/bucket type not found so no need to push into queue
	if len(logg.AnyError) != 0 || logg.Response.Body.BucketType == "" {
		WriteLog2(*logg)
		LogToNewFile(*logg)
		logg.ExecTime["func sendResponse"], _ = utils.GetExecTime(st)
		return
	}

	go qPush(*logg)
	logg.ExecTime["func sendResponse"], _ = utils.GetExecTime(st)
	return
}

// pushing in queue
func qPush(logg Logg) {

	raw, err := json.Marshal(logg)
	if err != nil {
		logg.AnyError["qPush Marshal"] = err.Error()
		WriteLog2(logg)
		return
	}

	enqData := make(map[string]string)
	enqData["publisher"] = "dhl"
	enqData["msgBody"] = enqData["publisher"]
	enqData["jsonDataStr"] = string(raw)

	// enqData["msgGroupID"] = enqData["publisher"]
	// enqData["msgDuplicationID"] = logg.Request.Glid

	_, err = queue.Send(enqData)

	if err != nil {
		logg.AnyError["qPush_queue_Send"] = err.Error()
	}
	WriteLog2(logg)
	LogToNewFile(logg)

	return
}

// WriteLog2 ...
func WriteLog2(logg Logg) {

	fileName := "gst_match_making_" + time.Now().Local().Format("02-01-2006") + ".txt"
	logsDir := properties.Prop.LOG_MATCH_MAKING_KIBANA + "/" + fileName

	//err ignored out of brevity
	jsonLog, _ := json.Marshal(logg)
	jsonLogString := string(jsonLog)

	f, err := os.OpenFile(logsDir, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer f.Close()

	f.WriteString("\n" + jsonLogString)
	return
}

func LogToNewFile(logg Logg) {
	logEntry := CreateLog(logg)

	logsDir := properties.Prop.LOG_MASTERINDIA + utils.TodayDir()
	// logsDir := serviceLogPath + utils.TodayDir()
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		e := os.MkdirAll(logsDir, os.ModePerm)
		if e != nil {
			fmt.Println(e)
			return
		}
	}

	logsDir += "/masterindia_wrapper_worker_logs.json"

	fmt.Println(logsDir)

	// Convert log entry to JSON
	jsonLog, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Println("Error marshaling log entry:", err)
		return
	}
	// Lock the mutex before writing to the file
	logFileMutex.Lock()
	defer logFileMutex.Unlock()

	// Open the log file and append the new log entry
	f, err := os.OpenFile(logsDir, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer f.Close()

	_, err = f.WriteString("\n" + string(jsonLog))
	if err != nil {
		fmt.Println("Error writing to log file:", err)
	}
}

func LogToWorkerFile(logg Logg) {
	logEntry := CreateWorkerLog(logg)

	logsDir := properties.Prop.LOG_MASTERINDIA + utils.TodayDir()
	// logsDir := serviceLogPath + utils.TodayDir()
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		e := os.MkdirAll(logsDir, os.ModePerm)
		if e != nil {
			fmt.Println(e)
			return
		}
	}

	logsDir += "/masterindia_wrapper_worker_logs.json"

	fmt.Println(logsDir)

	// Convert log entry to JSON
	jsonLog, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Println("Error marshaling log entry:", err)
		return
	}
	// Lock the mutex before writing to the file
	logFileMutex.Lock()
	defer logFileMutex.Unlock()

	// Open the log file and append the new log entry
	f, err := os.OpenFile(logsDir, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer f.Close()

	_, err = f.WriteString("\n" + string(jsonLog))
	if err != nil {
		fmt.Println("Error writing to log file:", err)
	}
}
