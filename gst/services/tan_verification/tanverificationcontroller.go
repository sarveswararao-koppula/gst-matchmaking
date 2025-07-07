package tan_verification

import (
        "encoding/json"
        "errors"
        "fmt"
        "mm/components/constants"
        "mm/properties"
        //"mm/queuetan"
        "mm/utils"
        "net/http"
        "os"
        "runtime/debug"
        "sync"
        "time"

        xid "github.com/rs/xid"
)

//var database string = properties.Prop.DATABASE
var loc *time.Location = utils.GetLocalTime()

//var database = "dev"

var mutex = &sync.Mutex{}

//ResponseTanverification ...
type ResponseTanverification struct {
        Code       int                    `json:"code,omitempty"`
        Status     string                 `json:"status,omitempty"`
        ErrMessage string                 `json:"err_message,omitempty"`
        Body       map[string]interface{} `json:"body,omitempty"`
}

//Tan_verification ...
type Tan_verification struct {
        RemoteAddress     string                  `json:"RemoteAddress,omitempty"`
        RequestStart      string                  `json:"RequestStart,omitempty"`
        RequestStartValue float64                 `json:"RequestStartValue,omitempty"`
        RequestEnd        string                  `json:"RequestEnd,omitempty"`
        RequestEndValue   float64                 `json:"RequestEndValue,omitempty"`
        ResponseTime      string                  `json:"ResponseTime,omitempty"`
        ResponseTime_Float float64 `json:"ResponseTime_Float,omitempty"` // float64 type
        ServiceName       string                  `json:"ServiceName,omitempty"`
        ServicePath       string                  `json:"ServicePath,omitempty"`
        ServiceURL        string                  `json:"ServiceURL,omitempty"`
        UniqueID          string                  `json:"UniqueID,omitempty"`
        Response          ResponseTanverification `json:"Response,omitempty"`
        RequestData       Rqst                    `json:"RequestData,omitempty"`
        StackTrace        string                  `json:"StackTrace,omitempty"`
        QueueMsgID        string                  `json:"QueueMsgID,omitempty"`
}

//Rqst ...
type Rqst struct {
        Modid         string `json:"modid,omitempty"`
        Glid          string `json:"glid,omitempty"`
        Validationkey string `json:"validationkey,omitempty"`
        API           string `json:"api,omitempty"`
        Tanid         string `json:"tanid,omitempty"`
}

//WorkRequest ...
type WorkRequest struct {
        Modid    string `json:"modid,omitempty"`
        APIName  string `json:"APIName,omitempty"`
        Tanid    string `json:"tanid,omitempty"`
        Glid     string `json:"glid,omitempty"`
        RqstTime string `json:"RqstTime,omitempty"`
}

//GetTANData
func GetTANData(w http.ResponseWriter, r *http.Request) {

        var logs Tan_verification
        logs.RequestStart = utils.GetTimeStampCurrent()
        logs.RequestStartValue = utils.GetTimeInNanoSeconds()
        logs.ServiceName = "tandata" //change
        logs.ServicePath = r.URL.Path
        logs.ServiceURL = r.RequestURI
        logs.RemoteAddress = utils.GetIPAdress(r)
        logs.UniqueID = xid.New().String() + xid.New().String()
        data := make(map[string]interface{})
        logs.RequestData = Rqst{}

        decoder := json.NewDecoder(r.Body)

        err := decoder.Decode(&logs.RequestData)

        if err != nil {
                //fmt.Println(err)
                sendResponse(w, 400, "FAILURE", "Invalid Params", data, &logs)
                return
        }

        _, err = validateProp(logs.RequestData.Modid, logs.RequestData.Validationkey)
        //fmt.Println("GLID ", logs.RequestData.Glid, "TAN_ID ", logs.RequestData.Tanid)

        if err != nil {
                sendResponse(w, 400, "AUTHENTICATION FAILURE", err.Error(), data, &logs)
                return
        }

        // if logs.RequestData.API == "masterindia" {
        //      masterindiaHand(user, w, &logs)
        //      return
        // }
        tanverify(w, &logs)
        return

}

func tanverify(w http.ResponseWriter, logs *Tan_verification) {

        response := make(map[string]interface{})

        defer func() {
                if panicCheck := recover(); panicCheck != nil {
                        stack := string(debug.Stack())
                        fmt.Println(stack)
                        logs.StackTrace = stack
			fmt.Println("issue: ",logs.StackTrace)
                       // response["result"] = "tan Details would be updated"
                       // sendResponse(w, 200, "SUCCESS", "", response, logs)
			//sendResponse(w, 500, "FAILURE", "Panic...Pls inform Dev Team", response, logs)
                        return
                }
        }()

        if len(logs.RequestData.Tanid) == 0 {
                sendResponse(w, 401, "FAILURE", "Param missing", response, logs)
                return
        }

        if len(logs.RequestData.Tanid) < 10 {
                sendResponse(w, 400, "FAILURE", "Not valid TAN", response, logs)
                return
        }

        if len(logs.RequestData.Tanid) > 10 {
                sendResponse(w, 400, "FAILURE", "Not valid TAN", response, logs)
                return
        }

        //Api_Name Api_user_name  tan glid Rqst_time

        wr := WorkRequest{
                Modid:    logs.RequestData.Modid,
                APIName:  logs.RequestData.API,
                Tanid:    logs.RequestData.Tanid,
                Glid:     logs.RequestData.Glid,
                RqstTime: logs.RequestStart,
        }

        raw, err := json.Marshal(wr)
        if err != nil {
                sendResponse(w, 400, "FAILURE", "failed at wr", response, logs)
                return
        }
        //fmt.Println("raw json ", string(raw))
        enqData := make(map[string]string)
        enqData["publisher"] = "tanAPI" //changepublisher
        enqData["jsonDataStr"] = string(raw)

        enqData["msgBody"] = enqData["publisher"]
        //enqData["msgDuplicationID"] = wr.APIName + wr.Tanid + wr.Glid
        //enqData["msgGroupID"] = wr.Modid

        // msgID, err := queuetan.Send(enqData)

        fmt.Println("enqData: ",enqData)
/*
        if err != nil {
                logs.StackTrace = err.Error()
                sendResponse(w, 200, "FAILURE", "Panic...Pls inform Dev Team", response, logs)
                return
        }

*/
        logs.QueueMsgID = ""    //msgID

        response["result"] = "tan Details would be updated"
	sendResponse(w, 200, "SUCCESS", "", response, logs)
        sherr:=SubcriberHandler(enqData["jsonDataStr"])
        fmt.Println("TAN Subscriber Handler error : ",sherr)

        //sendResponse(w, 200, "SUCCESS", "", response, logs)
        return

}

func validateProp(modid string, validationkey string) (string, error) {

        if constants.Propertiestan[modid].ValidaionKey != validationkey || validationkey == "" || modid == "" {
                return "", errors.New("Not Authorized")
        }

        return "", nil
}

func sendResponse(w http.ResponseWriter, httpcode int, status string, errorMsg string, response map[string]interface{}, logs *Tan_verification) {

        var serviceResponse ResponseTanverification
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

func write2Log(logs Tan_verification) {
        year, month, day := time.Now().Date()
        logsDir := properties.Prop.LOG_TAN + "/" + fmt.Sprint(year) + "/" + fmt.Sprint(int(month)) + "/" + fmt.Sprint(day)

        if _, err := os.Stat(logsDir); os.IsNotExist(err) {
                e := os.MkdirAll(logsDir, os.ModePerm)
                fmt.Println(e)
        }

        logsDir += "/tan_wrapper.json" //to be change

        jsonLog, err := json.Marshal(logs)
        fmt.Println(err)
        jsonLogString := string(jsonLog[:len(jsonLog)])
	//jsonLogString := string(jsonLog)
	//fmt.Println("wrapper-json-Log: ",jsonLogString)
        f, _ := os.OpenFile(logsDir, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	//fmt.Println("f file open:",f)
	//fmt.Println("par file open:",par)
        defer f.Close()

        mutex.Lock()
        defer mutex.Unlock()
        f.WriteString("\n" + jsonLogString)
	//fmt.Println("x-wrapper ",x)
	//fmt.Println("y-wrapper ",y)
        return
}

