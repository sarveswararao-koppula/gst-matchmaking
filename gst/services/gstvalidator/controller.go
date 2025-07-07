package gstvalidator

import (
	"encoding/json"
	"fmt"
	model "mm/model/masterindiamodel"
	"mm/properties"
	"mm/services/gstapis/masterindia"
	authadvance "mm/services/authbridgeadvanced"
	"mm/utils"
	"net/http"
	"os"
	"runtime/debug"
	//"strconv"
	"strings"
	"time"
	"regexp"
	"unicode"

	"github.com/rs/xid"
)

var database string = properties.Prop.DATABASE

//GstData ...
func GstData(w http.ResponseWriter, r *http.Request) {

	uniqID := xid.New().String()
	var logg Logg
	logg.RequestStart = utils.GetTimeStampCurrent()
	logg.RequestStartValue = utils.GetTimeInNanoSeconds()
	logg.ServiceName = serviceName
	logg.ServiceURL = r.RequestURI
	logg.AnyError = make(map[string]string)
	logg.ExecTime = make(map[string]float64)

	if parts := strings.Split(r.RemoteAddr, ":"); len(parts) == 2 {
		logg.RemoteAddress = parts[0]
	}

	defer func() {
		if panicCheck := recover(); panicCheck != nil {
			stack := string(debug.Stack())
			logg.StackTrace = stack
			sendResponse(uniqID, w, 500, failure, errPanic, nil, Data{}, logg)
			return
		}
	}()

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&logg.Request)

	if err != nil {
		sendResponse(uniqID, w, 400, failure, errParam, err, Data{}, logg)
		return
	}

	cols, err := ValidateProp(logg.Request.ModID, logg.Request.Validationkey)

	if err != nil {
		sendResponse(uniqID, w, 400, failure, errNotAuth, err, Data{}, logg)
		return
	}

	if logg.Request.Gst == "" {
			sendResponse(uniqID, w, 400, failure, errParam, err, Data{}, logg)
			return
	}

	if !gstValidation(logg.Request.Gst) {
		sendResponse(uniqID, w, 401, failure, InvalidGst, err, Data{}, logg)
		return
	}


	// var (
	// 	data map[string]interface{}
	// )

	// var (
	// 	gstgliddata map[string]interface{}
	// )

	var (
		data       = make(map[string]interface{})
		gstgliddata = make(map[string]interface{})
	)
	
	if logg.Request.Gst != "" {
		gstgliddata, err = GetGstFromGlid(database, logg.Request.Gst)
	} 
	
	if len(gstgliddata) > 0 {
		sendResponse(uniqID, w, 402, failure, DuplicateGst, err, Data{}, logg)
		return
	}

	_, st := utils.GetExecTime()

	if logg.Request.Gst != "" {
		data, err = GetGSTRecords(database, logg.Request.Gst)
	} 

	logg.ExecTime["DB_QUERY"], st = utils.GetExecTime(st)

	if err != nil {
		sendResponse(uniqID, w, 300, failure, errFetchDB, err, Data{}, logg)
		return
	}

	const apiName = "masterindia"
	APIUserID, err := masterindia.ValidateProp(logg.Request.ModID, logg.Request.Validationkey, apiName)

	if err != nil {
		sendResponse(uniqID, w, 350, failure, errValidationKey, err, Data{}, logg)
		return
	}

	gstInsertionDate := ""
	gstinNumber := logg.Request.Gst
	primDist := ""
	gstinStatus := ""

	for k, v := range data {
		if v == nil {
			data[k] = ""
		}
	}
	fmt.Println(data)

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
	} else if days, err := utils.DiffDaysddmmyyyy(gstInsertionDate); err != nil || days > 30 || ( gstinStatus != "ACTIVE" && days >= 1 ) || checkDist(gstInsertionDate, primDist) {
		logg.MasterIndia.Hit = true
		logg.MasterIndia.User = APIUserID
		iu = "U"
	}

	if logg.MasterIndia.Hit {

		wr := masterindia.Work{
			APIName:   apiName,
			APIUserID: logg.MasterIndia.User,
			GST:       gstinNumber,
			Modid:     logg.Request.ModID,
			UniqID:    uniqID,
		}

		_, st = utils.GetExecTime()
		_, params, err := wr.FetchGSTData(masterindiaAPILogs, 3000)
		logg.ExecTime["FetchGSTData"], st = utils.GetExecTime(st)

		if err != nil {
			errString := strings.ToLower(err.Error()) // make error string lowercase
			if strings.Contains(errString, "the gstin passed in the request is invalid") {
				sendResponse(uniqID, w, 401, failure, InvalidGst, err, Data{}, logg)
				return
			}
			sendResponse(uniqID, w, 351, failure, errFetchAPI, err, Data{}, logg)
			return
		}

		_, st = utils.GetExecTime()
		if strings.ToUpper(iu) == "U" {
			_, err = model.UpdateGSTMasterData(database, params)
		} else if strings.ToUpper(iu) == "I" {
			_, err = model.InsertGSTMasterData(database, params)
		}
		logg.ExecTime["I_U_GST_DATA"], st = utils.GetExecTime(st)

		if err != nil {
			sendResponse(uniqID, w, 301, failure, errUpdateDB, err, Data{}, logg)
			return
		}

		// data_glid, err := authadvance.GetGlidFromGstM(database, gstinNumber)
		//     if err != nil {
		// 	    fmt.Println("getting error from getglidfromgst function",err.Error())
		//     } else {
		// 	   for _, glid := range data_glid {
		// 		if err := authadvance.ProcessSingleGLIDPubapilogging(glid, "/gstvalidator/v1/gst"); err != nil {
		// 			// logg.AnyError["meshupdationerror"] = err.Error()
		// 			// s3log.Result["meshupdationerror_glid"] = fmt.Sprintf("GLID %s: %v", glid, err)
		// 			fmt.Println("meshupdationerror_glid",err.Error())
		// 		}
		//         }
	    //    }

		if err := authadvance.ProcessingGST(gstinNumber, "/gstvalidator/v1/gst"); err != nil {
			// logg.AnyError["meshupdationerror"] = err.Error()
			fmt.Println("meshupdationerror_gst",err.Error())
		}

		_, st = utils.GetExecTime()
		data, err = GetGSTRecords(database, gstinNumber)
		logg.ExecTime["DB_QUERY_2"], st = utils.GetExecTime(st)

		if err != nil {
			sendResponse(uniqID, w, 300, failure, errFetchDB, err, Data{}, logg)
			return
		}
	}

	result := make(map[string]interface{})
	for _, v := range cols {
		result[v] = data[v]
	}
	sendResponse(uniqID, w, 200, success, "", nil, Data{Values: result}, logg)
	return
}

func sendResponse(uniqID string, w http.ResponseWriter, httpcode int, status string, errorMsg string, err error, body Data, logg Logg) {

	w.Header().Set("Content-Type", "application/json")

	logg.Response = Res{
		Code:   httpcode,
		Error:  errorMsg,
		Status: status,
		Body:   body,
		UniqID: uniqID,
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
	return
}

//writeLog2 ...
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

func checkDist(gstInsertionDate, primDist string) bool {

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

func gstStateCode() map[string]string {
	return map[string]string{
		"01": "Jammu & Kashmir",
		"02": "Himachal Pradesh",
		"03": "Punjab",
		"04": "Chandigarh",
		"05": "Uttarakhand",
		"06": "Haryana",
		"07": "Delhi",
		"08": "Rajasthan",
		"09": "Uttar Pradesh",
		"10": "Bihar",
		"11": "Sikkim",
		"12": "Arunachal Pradesh",
		"13": "Nagaland",
		"14": "Manipur",
		"15": "Mizoram",
		"16": "Tripura",
		"17": "Meghalaya",
		"18": "Assam",
		"19": "West Bengal",
		"20": "Jharkhand",
		"21": "Odisha",
		"22": "Chhattisgarh",
		"23": "Madhya Pradesh",
		"24": "Gujarat",
		"25": "Daman & Diu",
		"26": "Dadra and Nagar Haveli",
		"27": "Maharashtra",
		"28": "Andhra Pradesh (old)",
		"29": "Karnataka",
		"30": "Goa",
		"31": "Lakshadweep",
		"32": "Kerala",
		"33": "Tamil Nadu",
		"34": "Pondicherry",
		"35": "Andaman & Nicobar",
		"36": "Telangana",
		"37": "Andhra Pradesh (new)",
		"38": "Ladakh",
	}
}

func gstValidation(gst string) bool {
	
	gst = strings.ToUpper(strings.TrimSpace(gst))

	if len(gst) < 15 {
		return false
	}

	if len(gst) != 15 {
		return false
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(gst) {
		return false
	}

	stateCode := gst[0:2]
	if _, ok := gstStateCode()[stateCode]; !ok {
		return false
	}

	if !unicode.IsDigit(rune(gst[12])) {
		return false
	}

	//if gst[13] != 'Z' {
	//	return false
	//}

	return true
}

