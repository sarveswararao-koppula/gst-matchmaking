package truthscreen

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mm/properties"
	"mm/utils"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/xid"
)

var mutex = &sync.Mutex{}

const (
	docType = "457"
	token   = "India@2608"
)

type GSTRequest struct {
	Modid         string `json:"modid,omitempty"`
	Validationkey string `json:"validationkey,omitempty"`
	GSTIN         string `json:"gstin,omitempty"`
}

type GSTResponse struct {
	Code           int    `json:"code"`
	Status         string `json:"status"`
	GSTIN_NUMBER   string `json:"gstin_number"`
	API_HIT        bool   `json:"api_hit"`
	TRANSACTION_ID string `json:"transaction_id"`
	ErrMessage     string `json:"err_message,omitempty"`
}

type Work struct {
	APIName   string `json:"APIName,omitempty"`
	APIUserId string `json:"APIUserId,omitempty"`
	GST       string `json:"gst,omitempty"`
	Modid     string `json:"modid,omitempty"`
	UniqID    string `json:"uniq_id,omitempty"`
}

type S3Log struct {
	RequestStart       string             `json:"RequestStart,omitempty"`
	RequestStartValue  float64            `json:"request_start_value,omitempty"`
	ServiceName        string             `json:"ServiceName,omitempty"`
	ServiceURL         string             `json:"ServiceURL,omitempty"`
	RequestEndValue    float64            `json:"request_end_value,omitempty"`
	FetchGSTDataTime   float64            `json:"response_time,omitempty"`
	Cred               string             `json:"cred,omitempty"`
	ExecTime           map[string]float64 `json:"exectime,omitempty"`
	Request            Work               `json:"Request,omitempty"`
	Result             map[string]string  `json:"Result,omitempty"`
	APIHit             bool               `json:"api_hit,omitempty"`
	AnyError           map[string]string  `json:"any_error,omitempty"`
	STATUS             int                `json:"STATUS"`
	RemoteAddress      string             `json:"RemoteAddress,omitempty"`
	RequestEnd         string             `json:"RequestEnd,omitempty"`
	ResponseTime       string             `json:"ResponseTime,omitempty"`
	ResponseTime_Float float64            `json:"ResponseTime_Float,omitempty"` // float64 type
	ServicePath        string             `json:"ServicePath,omitempty"`
	Response           GSTResponse        `json:"Response,omitempty"`
	RequestData        GSTRequest         `json:"RequestData,omitempty"`
}

func GetGSTDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var logs S3Log
	logs.RequestStart = utils.GetTimeStampCurrent()
	logs.RequestStartValue = utils.GetTimeInNanoSeconds()
	logs.ServicePath = r.URL.Path
	logs.ServiceURL = "/truthscreen/v1/gst"
	logs.RemoteAddress = utils.GetIPAdress(r)
	logs.RequestData = GSTRequest{}

	requiredParams := []string{"gstin", "modid", "validationkey"}
	for _, param := range requiredParams {
		if _, ok := r.URL.Query()[param]; !ok {
			http.Error(w, fmt.Sprintf("Missing '%s' parameter", param), http.StatusBadRequest)
			return
		}
	}

	gst_no := r.URL.Query().Get("gstin")
	modid := r.URL.Query().Get("modid")
	validationKey := r.URL.Query().Get("validationkey")

	logs.RequestData.Modid = modid
	logs.RequestData.GSTIN = gst_no
	logs.RequestData.Validationkey = validationKey

	// fmt.Println("logs.RequestData.GSTIN :",logs.RequestData.GSTIN)

	if logs.RequestData.GSTIN == "" {
		http.Error(w, fmt.Sprintf("Empty '%s' parameter", logs.RequestData.GSTIN), http.StatusBadRequest)
		return
	}

	if modid != "GlAdmin" && validationKey != "Z2xhZG1pbl9zY3JlZW4=" {
		logs.Response.ErrMessage = "Invalid modid or validationkey"
		Write2Kibana(logs)
		http.Error(w, "Invalid modid or validationkey", http.StatusUnauthorized)
		return
	}

	// fmt.Println("Line no 111")
	// var gstReq GSTRequest
	// err := json.NewDecoder(r.Body).Decode(&gstReq)
	// if err != nil {
	// 	fmt.Println("Line no 116")
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }

	response := GSTResponse{
		GSTIN_NUMBER: logs.RequestData.GSTIN,
		API_HIT:      true,
	}

	// fmt.Println("Line no 126")

	wr := Work{
		APIName:   "Gst_Advanced_Search",
		APIUserId: "prod5@indiamart.com",
		GST:       logs.RequestData.GSTIN,
		Modid:     "bi",
		UniqID:    xid.New().String(),
	}

	// fmt.Println("Line no 135")

	result, transID, _, err := fetchGSTData(wr)
	if err != nil {
		// fmt.Println("Line no 140")
		response.API_HIT = false
		response.ErrMessage = err.Error()
	}
	fmt.Println(result)
	response.Code = 200
	response.Status = "Success"
	response.TRANSACTION_ID = transID

	json.NewEncoder(w).Encode(response)
}

func fetchGSTData(wr Work) (map[string]string, string, string, error) {
	transID := ""
	message := ""
	logg := S3Log{
		APIHit:            false,
		RequestStart:      time.Now().Format(time.RFC3339),
		Cred:              "prod5@indiamart.com",
		RequestStartValue: float64(time.Now().UnixNano()),
		Request:           wr,
		ExecTime:          make(map[string]float64),
		Result:            make(map[string]string),
		AnyError:          make(map[string]string),
		ServiceName:       "truthscreen",
		ServiceURL:        "truthscreen/v1/gst",
	}

	logg.APIHit = true
	logg.STATUS = 0
	txnId := xid.New().String()

	_, st := utils.GetExecTime()
	encryptionKey := utils.MainFunction(txnId, docType, logg.Request.GST, token)
	logg.ExecTime["AuthBridge_Encryption_Hit_Time"], st = utils.GetExecTime(st)

	data, err := fetchGSTDetails(encryptionKey)
	logg.ExecTime["AuthBridge_API_Hit_Time"], st = utils.GetExecTime(st)

	if err != nil {
		logg.AnyError["FetchGSTDetails"] = err.Error()
		Write2Kibana(logg)
		return nil, "", "", err
	}

	if _, ok := data["responseData"]; !ok {
		logg.AnyError["FetchGSTDetails"] = fmt.Sprint(data)
		Write2Kibana(logg)
		return nil, "", "", fmt.Errorf("no response data found")
	}

	decryptionResult := utils.Decryption(data["responseData"].(string), token)
	logg.ExecTime["AuthBridge_Result_Decryption_Time"], st = utils.GetExecTime(st)

	var stringMap map[string]interface{}
	err = json.Unmarshal([]byte(decryptionResult), &stringMap)
	if err != nil {
		logg.AnyError["FetchGSTDetails"] = err.Error()
		Write2Kibana(logg)
		return nil, "", "", err
	}

	stringMap["status"] = int(stringMap["status"].(float64))
	status, _ := stringMap["status"]

	if status == 1 {
		logg.STATUS = 1
		transID, _ = stringMap["ts_trans_id"].(string)
		message, _ = stringMap["msg"].(string)
		logg.Result["Trans_Id"] = transID
		logg.Result["Message"] = message
		logg.ExecTime["Time_for_Seperating_Trans_id"], st = utils.GetExecTime(st)
	} else {
		logg.APIHit = false
	}

	logg.RequestEndValue = float64(time.Now().UnixNano())
	logg.FetchGSTDataTime = (logg.RequestEndValue - logg.RequestStartValue) / 1000000
	logg.RequestStartValue = 0
	logg.RequestEndValue = 0

	Write2Kibana(logg)
	return logg.Result, transID, message, nil
}

func fetchGSTDetails(requestData string) (map[string]interface{}, error) {
	url := "https://www.truthscreen.com/api/v1.0/gst"
	timeout := 30000

	client := http.Client{
		Timeout: time.Duration(timeout) * time.Millisecond,
	}

	payload := strings.NewReader(requestData)
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return nil, err
	}

	req.Header.Add("username", "prod5@indiamart.com")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	return data, err
}

// Write2Kibana ...
func Write2Kibana(logs S3Log) {

	logsDir := properties.Prop.LOG_MASTERINDIA + utils.TodayDir()

	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		e := os.MkdirAll(logsDir, os.ModePerm)
		fmt.Println(e)
	}

	logsDir += "/masterindia_wrapper.json"

	fmt.Println(logsDir)

	jsonLog, _ := json.Marshal(logs)

	jsonLogString := string(jsonLog[:len(jsonLog)])

	f, _ := os.OpenFile(logsDir, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()

	mutex.Lock()
	defer mutex.Unlock()
	f.WriteString("\n" + jsonLogString)
	return
}
