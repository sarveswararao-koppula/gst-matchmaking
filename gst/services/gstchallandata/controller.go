package gstchallandata

import (
        "encoding/json"
        "errors"
        "fmt"
        api "mm/api/thirdpartyapi"
        "mm/components/constants"
        model "mm/model/masterindiamodel"
        "mm/properties"
        "mm/queue"
        "mm/utils"
        "net/http"
        "os"
        "runtime/debug"
        "strconv"
        "strings"
        "sync"
        "time"
        "github.com/rs/xid"
)

//loc
var loc *time.Location = utils.GetLocalTime()

//mutex
var mutex = &sync.Mutex{}

//S3LOGS
type S3Log struct {
        APIName            string                 `json:"APIName,omitempty"`
        APIUserID          string                 `json:"APIUserID,omitempty"`
        APIHit             string                 `json:"APIHit,omitempty"`
        Gst                string                 `json:"Gst,omitempty"`
        Modid              string                 `json:"Modid,omitempty"`
        RqstTime           string                 `json:"RqstTime,omitempty"`
        Result             map[string]interface{} `json:"Result,omitempty"`
        StackTrace         string                 `json:"stack_trace,omitempty"`
        AnyError           map[string]string      `json:"AnyError,omitempty"`
}

//FYSTRUCT
type fy struct {
        fy     string
        start  int
        end    int
        apihit bool
        maxDof string
}

//Tokken ...
type Tokken struct {
        Tok string    `json:"Tok,omitempty"`
        Exp time.Time `json:"Exp,omitempty"`
}

//Tokkens ...
var Tokkens map[string]*Tokken = make(map[string]*Tokken)

var database string = properties.Prop.DATABASE

//GstChallanData..
func GstChallanData(w http.ResponseWriter, r *http.Request) {
        uniqID := xid.New().String()
        var logg Logg
        logg.RequestStart = utils.GetTimeStampCurrent()
        logg.RequestStartValue = utils.GetTimeInNanoSeconds()
        logg.ServiceName = serviceName
        logg.ServiceURL = r.RequestURI
        //fmt.Println(logg.ServiceURL, "Dev-Testing")
        logg.AnyError = make(map[string]string)
        logg.ExecTime = make(map[string]float64)
        var apiName = "challan"
        max_dof := ""
        if parts := strings.Split(r.RemoteAddr, ":"); len(parts) == 2 {
                logg.RemoteAddress = parts[0]
        }
        defer func() {
                if panicCheck := recover(); panicCheck != nil {
                        stack := string(debug.Stack())
                        logg.StackTrace = stack
                        sendResponse(uniqID, w, 500, failure, errPanic, nil, Data{}, max_dof, logg)
                        return
                }
        }()

        decoder := json.NewDecoder(r.Body)
        err := decoder.Decode(&logg.Request)
        gstinNumber := logg.Request.Gst
        //fmt.Println(gstinNumber, "Dev-GST")
        modid := logg.Request.ModID
        //fmt.Println(modid, "Dev-Modid")
        //fmt.Println(logg.Request.Validationkey, "Dev-Validation-Key")

        if err != nil {
                sendResponse(uniqID, w, 400, failure, errParam, err, Data{}, max_dof, logg)
                return
        }

        if modid == "weberp"  {
                apiName = "gstchallanscreen"
        }

        if modid == "loans2"{
                var (
                        data_gst2 map[string]interface{}
                )
                var (
                        data []map[string]string
                )
                //First Get GST and then go for Get GST Records 36
                data_gst2, err = GetGstFromGlid(database, logg.Request.Glid)
                if err != nil {
                        sendResponse(uniqID, w, 400, failure, errFetchDB, err, Data{}, "", logg)
                        return
                }

                for k, v := range data_gst2 {
                        //e = "Fifth"
                        if v == nil {
                                data_gst2[k] = ""
                        }
                }

                if len(data_gst2) == 0 || data_gst2["gst"] == "" {
                        sendResponse(uniqID, w, 400, failure, "There is no GST mapped to this user", err, Data{}, "", logg)
                        // sendResponse(uniqID, w, 400, Updateflag, failure, "There is no GST mapped to this user", nil, Data{}, logg, logg.Request.Gst, hsncode)
                        return
                }

                if data_gst2["gst"] != "" {
                        logg.Request.Gst = data_gst2["gst"].(string)
                }
                data_gst2 = make(map[string]interface{})

                if logg.Request.Gst != "" {
                        //fmt.Println(database)
                        data, err = GetChallanAllDetails(database, logg.Request.Gst)
                        //fmt.Println(data)
                }
                if err != nil {
                        sendResponse(uniqID, w, 400, failure, errFetchDB, err, Data{}, "", logg)
                        return
                }
                sendResponse(uniqID, w, 200, success, "", nil, Data{Values: data}, "", logg)
                return
        }

        user, err2 := validateProp(logg.Request.ModID, logg.Request.Validationkey, apiName)
        if err2 != nil {
                sendResponse(uniqID, w, 400, failure, errNotAuth, err2, Data{}, max_dof, logg)
                return
        }
        //fmt.Println(user, "Dev-Testing")
        credential := utils.GetCred(user)
        //fmt.Println(credential, "Dev-Credential")
        logg.MasterIndia.User = user

        if logg.Request.Gst == "" {
                sendResponse(uniqID, w, 400, failure, errParam, err, Data{}, max_dof, logg)
                return
        }

        var (
                data []map[string]string
        )
        flag0 := ""
        flag1 := ""
        max_dof_temp := ""
        _, st := utils.GetExecTime()
       // fmt.Println(logg.Request.Gst, "Dev-GST")
        if logg.Request.Flag == "1" {
                if logg.Request.Gst != "" {
                        //fmt.Println(database, "Dev-Database")
                        fetched_data, err := GetLatestDOF(database, logg.Request.Gst)
                        logg.ExecTime["DBQUERY0"], st = utils.GetExecTime(st)
                        max_dof = fetched_data["date_of_filing"]
                        max_dof_temp = max_dof
                        if err != nil {
                                sendResponse(uniqID, w, 400, failure, errFetchDB, err, Data{}, max_dof, logg)
                                return
                        }

                        if len(fetched_data) == 0 {
                                //fmt.Println("Dev-Data-Len-0")
                                max_dof, err, flag0, flag1 = hitChallan(credential, apiName, user, gstinNumber, modid, &logg, w, uniqID)
                                if err != nil {
                                        sendResponse(uniqID, w, 400, failure, errFetchDB, err, Data{}, max_dof, logg)
                                        return
                                }
                                if max_dof == "" {
                                        max_dof = max_dof_temp
                                }
                                if flag0 == "0" && flag1 == "0" {
                                        sendResponse(uniqID, w, 200, success, "", nil, Data{}, max_dof, logg)
                                        return
                                } else {
                                        return
                                }
                        }
                }
                if max_dof != "" && diffDays(max_dof) <= float64(90) {
                        //fmt.Println("Dev-Inside-Panic")
                        sendResponse(uniqID, w, 200, success, "", nil, Data{}, max_dof, logg)
                        return
                } else {
                       // fmt.Println("Dev-Hitting>90")
                        max_dof, err, flag0, flag1 = hitChallan(credential, apiName, user, gstinNumber, modid, &logg, w, uniqID)
                        if err != nil {
                                sendResponse(uniqID, w, 400, failure, errFetchDB, err, Data{}, max_dof, logg)
                                return
                        }
                        if max_dof == "" {
                                //fmt.Println("Inside-else")
                                max_dof = max_dof_temp
                        }
                        if flag0 == "0" && flag1 == "0" {
                                sendResponse(uniqID, w, 200, success, "", nil, Data{}, max_dof, logg)
                                return
                        } else {
                                return
                        }
                }
        } else {
                _, st := utils.GetExecTime()
                //fmt.Println(logg.Request.Gst, "GST")
                if logg.Request.Gst != "" {
                        //fmt.Println(database)
                        data, err = GetChallanDetails(database, logg.Request.Gst)
                        //fmt.Println(data)
                }

                logg.ExecTime["DB_QUERY"], st = utils.GetExecTime(st)
                if err != nil {
                        sendResponse(uniqID, w, 400, failure, errFetchDB, err, Data{}, "", logg)
                        return
                }
                sendResponse(uniqID, w, 200, success, "", nil, Data{Values: data}, "", logg)
        }
        return
}

//hitchallan
func hitChallan(credential map[string]string, apiName string, user string, gstinNumber string, modid string, logg *Logg, w http.ResponseWriter, uniqID string) (string, error, string, string) {
        //var params []interface{}
        var s3log S3Log
        s3log.APIName = apiName
        s3log.APIUserID = credential["username"]
        s3log.Gst = gstinNumber
        //fmt.Println(gstinNumber, "Dev-GstinNumber")
        s3log.APIHit = ""
        s3log.Modid = modid
        s3log.RqstTime = utils.GetTimeStampCurrent()
        s3log.Result = make(map[string]interface{})
        var max_dof string
        flag0 := ""
        flag1 := ""
        //max_dof_s := ""
        fys := make([]fy, 4)
        getFY(fys)
        _, st := utils.GetExecTime()
        //fmt.Println(fys, "Dev-Fys")
        err := fyAPIStatus(fys, gstinNumber)
        logg.ExecTime["DBQUERY1"], st = utils.GetExecTime(st)
        //fmt.Println(fys, "Dev-After-Fys")
        if err != nil {
                s3log.APIHit = err.Error()
                Write2S3(&s3log)
                return max_dof, err, flag0, flag1
        }
        err = ValidateTokken(credential, user)
        if err != nil {
                s3log.Result["api_error"] = err.Error()
                Write2S3(&s3log)
                return max_dof, err, flag0, flag1
        }
        fmt.Println("tokken", Tokkens[user], "GST", s3log.Gst)
        _, st = utils.GetExecTime()
        regData, err := model.GetExistsMasterdata(database, s3log.Gst)
        logg.ExecTime["DBQUERY2"], st = utils.GetExecTime(st)
        //fmt.Println(regData, "Dev-RegData")
        if err != nil {
                s3log.APIHit = err.Error()
                Write2S3(&s3log)
                return max_dof, err, flag0, flag1
        }
        //latest filing date for all fys

        fyLatestFilingOrig := ""
        regisDateStr := ""
        if len(regData) > 0 {
                fyLatestFilingOrig = regData["date_of_filing"]
                regisDateStr = regData["registration_date"]
        }
        fyLatestFiling := fyLatestFilingOrig
        //fmt.Println(fyLatestFiling, "Dev-Latest-Filing")
        //fmt.Println("regisDateStr", regisDateStr)
        if regisDate, err := time.Parse("02-01-2006", regisDateStr); err == nil {

                regYear := -1
                //fmt.Println(regisDateStr, "Dev-RegisDateStr")
                var mmm time.Month
                regYear, mmm, _ = regisDate.Date()
                //fmt.Println(regYear, "Dev-RegYear")
                if int(mmm) < 4 {
                        regYear--
                }
                //fmt.Println(regYear, "Dev-RegYear-2")
                if regYear > 0 {
                        for i, fy := range fys {
                                if fy.start < regYear {
                                        fy.apihit = false
                                        fys[i] = fy
                                }
                        }
                }
                //fmt.Println(fys, "Dev-fys")
                //fmt.Println("regYear", regYear)
        }
        gstinNumber = logg.Request.Gst
        flag0 = "0"
        flag1 = "0"
        _, st = utils.GetExecTime()
        for _, fy := range fys {
                fmt.Println(fy)
                // if fy.fy == "fy0" {
                //      curFyLatestFiling = fy.maxDof
                // }
                fyLatestFiling = getMaxDate(fyLatestFiling, fy.maxDof)
                //fmt.Println(fyLatestFiling, "Dev-Fy")
                var s3log S3Log
                s3log.APIName = apiName
                s3log.APIUserID = credential["username"]

                fyYear := strconv.Itoa(fy.start) + "-" + strconv.Itoa(fy.end%100)

                s3log.Gst = gstinNumber + "#" + fyYear //gst#2020-21

                s3log.APIHit = "N"
                s3log.Modid = modid
                s3log.RqstTime = utils.GetTimeStampCurrent()
                s3log.Result = make(map[string]interface{})

                s3log.Result["max_dof"] = fy.maxDof

                if !fy.apihit {
                        Write2S3(&s3log)
                        continue
                }

                s3log.APIHit = "Y"
                //logging time
                //2021
                data, err := api.GetChallanData(gstinNumber, fyYear, map[string]string{
                        "client_id":    credential["client_id"],
                        "access_token": Tokkens[user].Tok,
                })
                //data fetched
                //fmt.Println(data, "Dev-Challan-Data")
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
                                //challan
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
                        continue
                }

                //DATA FOUND

                EFiledlist, _ := apiData["EFiledlist"].([]interface{})

                var chllandataArr []interface{}

                for _, v := range EFiledlist {
                        fmt.Println(fy)
                        apiDataClosure := make(map[string]interface{})
                        //each value
                        chllandata, ok := v.(map[string]interface{})
                        if ok {
                                // if fy.fy == "fy0" {
                                //      dofStr, _ := chllandata["dof"].(string)
                                //      curFyLatestFiling = getMaxDate(curFyLatestFiling, dofStr)
                                // }
                                dofStr, _ := chllandata["dof"].(string)
                                fyLatestFiling = getMaxDate(fyLatestFiling, dofStr)
                                //apiDataClosure["i_u_d"] = "I"
                                //fmt.Println(fyLatestFiling, "Dev-Fy2")
                        }
                        apiDataClosure["data"] = chllandata
                        chllandataArr = append(chllandataArr, apiDataClosure)
                }
                s3log.Result["apiData"] = chllandataArr
                //Write2S3(&s3log)
                //Putting the new variable here for storing latest filing
                s3log.Result["max_dof"] = fyLatestFiling
                //fmt.Println("Dev-Fy3", fyLatestFiling)
                s3log.Result["max_dof_orig"] = fyLatestFilingOrig
                //fmt.Println("Dev-Fy4", fyLatestFilingOrig)
                if fy.fy == "fy0" && fyLatestFiling != "" {
                        flag0 = "1"
                        //fmt.Println(flag0, "Dev-Flag")
                }
                if fy.fy == "fy1" && fyLatestFiling != "" {
                        flag1 = "1"
                }
                if fy.fy == "fy1" && flag0 == "0" && flag1 == "1" {
                        //fmt.Println("Dev-Financial-Year-1")
                        logg.ExecTime["fyloop"], st = utils.GetExecTime(st)
                        SendResponse(uniqID, w, 200, success, "", nil, Data{}, fyLatestFiling, logg)
                        go qPush(&s3log, s3log)
                        continue
                } else if fy.fy == "fy0" && flag0 == "1" && flag1 == "0" {
                        //fmt.Println("Dev-Financial-Year-0")
                        logg.ExecTime["fyloop"], st = utils.GetExecTime(st)
                        SendResponse(uniqID, w, 200, success, "", nil, Data{}, fyLatestFiling, logg)
                        //Write2S3(&s3log)
                        go qPush(&s3log, s3log)
                        continue
                }
                //Write2S3(&s3log)
                go qPush(&s3log, s3log)
                continue
        }
        fmt.Println("fyLatestFilingOrig= ", fyLatestFilingOrig)
        fmt.Println("fyLatestFiling= ", fyLatestFiling)
        //return to merp
        if flag0 == "0" && flag1 == "0" {
                return fyLatestFiling, nil, flag0, flag1
        }
        if (flag0 == "1" && flag1 == "0") || (flag1 == "1" && flag0 == "0") {
                return "", nil, flag0, flag1
        }
        return "", nil, flag0, flag1
}

//qpush Function
func qPush(logg *S3Log, log S3Log) {
        //fylatestfiling,fylatestfilingorig
        //fmt.Println(*logg, "Dev-qpush")
        raw, err := json.Marshal(*logg)
        //fmt.Println("Dev-Raw", string(raw))
        //fmt.Printf("%+v\n", logg)
        if err != nil {
                log.AnyError["qPush Marshal"] = err.Error()
                Write2S3(&log)
                return
        }

        enqData := make(map[string]string)
        enqData["publisher"] = "instantact"
        enqData["msgBody"] = enqData["publisher"]
        enqData["jsonDataStr"] = string(raw)
        // enqData["msgGroupID"] = enqData["publisher"]
        // enqData["msgDuplicationID"] = logg.APIName + logg.Gst

        _, err = queue.Send(enqData)

        if err != nil {
                log.AnyError["qPush_queue_Send"] = err.Error()
        }
        Write2S3(&log)
        return
}

//GetMaxDate
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

//Getfy
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

//WRITE2S3
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

//FetchingUser
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

//validating tokken
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

//function1
func fyAPIStatus(fys []fy, gstinNumber string) error {

        var params []interface{}

        params = append(params, fys[0].end, fys[1].end, fys[2].end, fys[3].end, fys[3].start, gstinNumber)
        data, err := model.CheckChallanDetails(database, params)
        fmt.Println("Inseide-fyAPIstatus", data)
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

//diffdays function
func diffDays(dateOfFilingStr string) float64 {
        nowStr := time.Now().In(loc).Format("02-01-2006")
        now, _ := time.Parse("02-01-2006", nowStr)
        dateOfFiling, _ := time.Parse("02-01-2006", dateOfFilingStr)
        return float64(now.Sub(dateOfFiling).Hours() / 24)
}

//Have to change some things in sendResponse and write2
func sendResponse(uniqID string, w http.ResponseWriter, httpcode int, status string, errorMsg string, err error, body Data, max_dof string, logg Logg) {

        w.Header().Set("Content-Type", "application/json")

        logg.Response = Res{
                Code:    httpcode,
                Error:   errorMsg,
                Status:  status,
                Body:    body,
                Max_dof: max_dof,
                UniqID:  uniqID,
        }
        if err != nil {
                logg.AnyError[errorMsg] = err.Error()
        }

        json.NewEncoder(w).Encode(logg.Response)

        logg.RequestEndValue = utils.GetTimeInNanoSeconds()
        logg.ResponseTime = (logg.RequestEndValue - logg.RequestStartValue) / 1000000
        logg.RequestStartValue = 0
        logg.RequestEndValue = 0

        writeLog2(logg)
        return
}

//Have to change some things in sendResponse and write2
func SendResponse(uniqID string, w http.ResponseWriter, httpcode int, status string, errorMsg string, err error, body Data, max_dof string, logg *Logg) {
        w.Header().Set("Content-Type", "application/json")
        logg.Response = Res{
                Code:    httpcode,
                Error:   errorMsg,
                Status:  status,
                Body:    body,
                Max_dof: max_dof,
                UniqID:  uniqID,
        }
        if err != nil {
                logg.AnyError[errorMsg] = err.Error()
        }

        json.NewEncoder(w).Encode(logg.Response)

        logg.RequestEndValue = utils.GetTimeInNanoSeconds()
        logg.ResponseTime = (logg.RequestEndValue - logg.RequestStartValue) / 1000000
        logg.RequestStartValue = 0
        logg.RequestEndValue = 0

        writeLog2(*logg)
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
