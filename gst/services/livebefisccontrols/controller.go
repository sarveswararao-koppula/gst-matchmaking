package livebefisccontrols

import (
	"encoding/json"
	"errors"
	"fmt"
	"mm/components/constants"
	"mm/properties"
	"mm/utils"
	"net/http"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/rs/xid"
)

var database string = properties.Prop.DATABASE
var loc *time.Location = utils.GetLocalTime()

var mutex = &sync.Mutex{}

type Rqst struct {
	Modid         string `json:"modid,omitempty"`
	Validationkey string `json:"validationkey,omitempty"`
	API           string `json:"api,omitempty"`
	Gst           string `json:"gst,omitempty"`
}

// ResponseBefisccontrols ...
type ResponseBefisccontrols struct {
	Code       int                    `json:"code,omitempty"`
	Status     string                 `json:"status,omitempty"`
	ErrMessage string                 `json:"err_message,omitempty"`
	Body       map[string]interface{} `json:"body,omitempty"`
}

// WorkRequest ...
type WorkRequest struct {
	APIName     string `json:"APIName,omitempty"`
	APIUserName string `json:"APIUserName,omitempty"`
	GstPan      string `json:"GstPan,omitempty"`
	Modid       string `json:"Modid,omitempty"`
	RqstTime    string `json:"RqstTime,omitempty"`
}

// Befisccontrols ...
type Befisccontrols struct {
	RemoteAddress      string                 `json:"RemoteAddress,omitempty"`
	RequestStart       string                 `json:"RequestStart,omitempty"`
	RequestStartValue  float64                `json:"RequestStartValue,omitempty"`
	RequestEnd         string                 `json:"RequestEnd,omitempty"`
	RequestEndValue    float64                `json:"RequestEndValue,omitempty"`
	ResponseTime       string                 `json:"ResponseTime,omitempty"`
	ResponseTime_Float float64                `json:"ResponseTime_Float,omitempty"` // float64 type
	ServiceName        string                 `json:"ServiceName,omitempty"`
	ServicePath        string                 `json:"ServicePath,omitempty"`
	ServiceURL         string                 `json:"ServiceURL,omitempty"`
	UniqueID           string                 `json:"UniqueID,omitempty"`
	Response           ResponseBefisccontrols `json:"Response,omitempty"`
	RequestData        Rqst                   `json:"RequestData,omitempty"`
	StackTrace         string                 `json:"StackTrace,omitempty"`
	QueueMsgID         string                 `json:"QueueMsgID,omitempty"`
	ReadResponseTime   string                 `json:"ReadResponseTime,omitempty"`
	WriteResponseTime  string                 `json:"WriteResponseTime,omitempty"`
}

// GetBefiscData ...
func GetBefiscData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var logs Befisccontrols

	logs.RequestStart = utils.GetTimeStampCurrent()
	logs.RequestStartValue = utils.GetTimeInNanoSeconds()
	logs.ServiceName = "Befisc"
	logs.ServicePath = r.URL.Path
	logs.ServiceURL = "/befisc/v1/gst"
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

	if logs.RequestData.API == "befisc" {
		befiscHand(user, w, &logs)
		return
	}
}

func befiscHand(user string, w http.ResponseWriter, logs *Befisccontrols) {

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

	//GlAdmin
	if logs.RequestData.Modid == "gladmin" {
		gladmin(w, logs, user)
		return
	}
	return
}

func gladmin(w http.ResponseWriter, logs *Befisccontrols, user string) {
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

	// Send acknowledgment immediately
	response["result"] = "GST Details would be updated"
	sendResponse(w, 200, "SUCCESS", "", response, logs)

	// Run the remaining code in the background
	go func(logs *Befisccontrols, user string) {
		if logs.RequestData.Modid == "gladmin" {
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
				err2 := BefiscSubcriberHandler(string(raw))
				if err2 != nil {
					fmt.Println("Subscriber Handler2 error:", err2)
				}
				return
		}
	}(logs,user)
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

func sendResponse(w http.ResponseWriter, httpcode int, status string, errorMsg string, response map[string]interface{}, logs *Befisccontrols) {

	var serviceResponse ResponseBefisccontrols
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

func write2Log(logs Befisccontrols) {

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
