// services/pingcontrols/controller.go
package pingcontrols

import (
	"encoding/json"
	"fmt"
	"mm/properties"
	"mm/utils"
	"net/http"
	"os"
	"sync"
)

var mutex = &sync.Mutex{}

// HealthCheckResponse struct to hold the health check response
type HealthCheckResponse struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Value  string `json:"response_sent"`
}

// PingLogs struct for logging health check requests
type PingLogs struct {
	RemoteAddress      string  `json:"RemoteAddress,omitempty"`
	RequestStart       string  `json:"RequestStart,omitempty"`
	RequestEnd         string  `json:"RequestEnd,omitempty"`
	ResponseTime_Float float64 `json:"ResponseTime_Float,omitempty"`
	ServicePath        string  `json:"ServicePath,omitempty"`
	ServiceURL         string  `json:"ServiceURL,omitempty"`
	ResponseCode       int     `json:"ResponseCode,omitempty"`
	ResponseStatus     string  `json:"ResponseStatus,omitempty"`
	ResponseValue      string  `json:"ResponseValue,omitempty"`
}

// GetPingResponse handles the GET request to provide a health check response.
func GetPingResponse(w http.ResponseWriter, r *http.Request) {
	// Initialize logs for the request
	var logs PingLogs
	startTime := utils.GetTimeInNanoSeconds()
	logs.RequestStart = utils.GetTimeStampCurrent()
	logs.ServicePath = r.URL.Path
	logs.ServiceURL = "/ping/v1/gst"
	logs.RemoteAddress = utils.GetIPAdress(r)

	// Define the response body
	response := HealthCheckResponse{
		Code:   200,
		Status: "SUCCESS",
		Value:  "pong",
	}

	// Set response header and send the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	// Update log details
	logs.ResponseCode = response.Code
	logs.ResponseStatus = response.Status
	logs.ResponseValue = response.Value
	logs.RequestEnd = utils.GetTimeStampCurrent()
	endTime := utils.GetTimeInNanoSeconds()
	logs.ResponseTime_Float = (endTime - startTime) / 1000000

	// Write logs to Kibana
	Write2Kibana(logs)
}

// Write2Kibana writes logs to the specified file in an efficient manner
func Write2Kibana(logs PingLogs) {
	logsDir := properties.Prop.LOG_MASTERINDIA + utils.TodayDir()

	// Ensure the log directory exists
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logsDir, os.ModePerm); err != nil {
			fmt.Println("Error creating log directory:", err)
			return
		}
	}

	logsFile := logsDir + "/masterindia_wrapper.json"

	// Marshal the log to JSON
	jsonLog, err := json.Marshal(logs)
	if err != nil {
		fmt.Println("Error marshalling log to JSON:", err)
		return
	}

	// Open the log file in append mode with write lock for concurrency
	f, err := os.OpenFile(logsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer f.Close()

	mutex.Lock()
	defer mutex.Unlock()

	// Write the JSON log entry with a newline for separation
	if _, err := f.WriteString(string(jsonLog) + "\n"); err != nil {
		fmt.Println("Error writing log to file:", err)
	}
}
