package authbridgeadvanced

import (
	//"bytes"
	"encoding/json"
	"errors"
	"fmt"

	//"io/ioutil"
	"database/sql"
	servapi "mm/api/servapi"
	db "mm/components/database"
	model "mm/model/masterindiamodel"
	"mm/properties"
	"mm/utils"
	"net/http"
	"os"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"
)

var database string = properties.Prop.DATABASE

// GstChallanData..
func GstAuthData(w http.ResponseWriter, r *http.Request) {
	uniqID := ""
	var logg Logg
	var params []interface{}
	hsnstring := ""
	logg.RequestStart = utils.GetTimeStampCurrent()
	logg.RequestStartValue = utils.GetTimeInNanoSeconds()
	logg.ServiceName = serviceName
	logg.ServiceURL = r.RequestURI
	logg.AnyError = make(map[string]string)
	logg.ExecTime = make(map[string]float64)
	stringMap := make(map[string]interface{})
	if parts := strings.Split(r.RemoteAddr, ":"); len(parts) == 2 {
		logg.RemoteAddress = parts[0]
	}

	defer func() {
		if panicCheck := recover(); panicCheck != nil {
			stack := string(debug.Stack())
			logg.StackTrace = stack
			sendResponse(uniqID, w, 500, failure, errPanic, nil, "", logg)
			return
		}
	}()

	dataMap := r.PostFormValue("response_data")

	logg.Request.Responsedata = dataMap

	//dataMap := logg.Request.Responsedata

	err := json.Unmarshal([]byte(dataMap), &stringMap)
	if err != nil {
		sendResponse(uniqID, w, 400, failure, errUnmarshal, err, "", logg)
		return
	}

	uniqID = stringMap["ts_trans_id"].(string)

	if uniqID == "" {
		sendResponse(uniqID, w, 400, failure, errParam, err, "", logg)
		return
	}
	//status :=int(logg.Request.Status.(float64))

	stringMap["status"] = int(stringMap["status"].(float64))
	status := stringMap["status"]

	logg.STATUS = 0
	if status == 1 {
		logg.STATUS = 1
	}

	if status == 0 {

		sendResponse(uniqID, w, 400, failure, errFetchAPI, err, "", logg)

		sendResponseKibana(uniqID, 400, failure, errFetchAPI, err, "", logg)
		return
	}

	apiDataHSN := make(map[string]interface{})
	if status == 1 {

		apiData, _ := stringMap["msg"].(map[string]interface{})
		apiDataHSN, _ = apiData["goods_n_service"].(map[string]interface{})
		fmt.Println(apiDataHSN, "Dev-Third")

		Gst, _ := apiData["GSTIN/ UIN"].(string)
		fmt.Println(Gst)
		dbData, err2 := getGSTData(Gst)
		if err2 != nil {
			sendResponse(uniqID, w, 400, failure, errFetchDB, err, "", logg)
			return
		}
		//Bus Logic for Gst Additional Columns
		_, st := utils.GetExecTime()
		logg.Result, params = utils.BusLogicOnMasterData_V3(Gst, apiData)
		logg.ExecTime["BusLogic-GST-Columns-Mapping-Time"], st = utils.GetExecTime(st)

		//Bus Logic for Hsn Information
		logg.ResultHSN, hsnstring = utils.BusLogicOnAuthbridgeHSN_V1(Gst, apiDataHSN)
		logg.ExecTime["BusLogic-HSN-Columns-Mapping-Time"], st = utils.GetExecTime(st)

		//Insert and Update
		if len(dbData) == 0 {
			_, err = model.InsertAuthBridgeGSTMasterData("approvalPG", params)
			// _, err = model.InsertAuthBridgeGSTMasterData("dev", params)
			logg.Result["i_u"] = "i"
		} else {
			_, err = model.UpdateAuthBridgeGSTMasterData("approvalPG", params)
			// _, err = model.UpdateAuthBridgeGSTMasterData("dev", params)
			logg.Result["i_u"] = "u"
		}
		logg.ExecTime["i/u-DbTime"], st = utils.GetExecTime(st)

		if err != nil {
			sendResponse(uniqID, w, 200, failure, errUpdateDB, err, "", logg)
			return
		} else {

			turnover, ok := params[22].(string)
			if !ok {
				logg.AnyError["TurnoverJsonUpdationError"] = "Turnover Missing in response"
			} else {
				err2 := TurnoverUpdation(turnover, Gst)
				if err2 != nil {
					logg.AnyError["TurnoverJsonUpdationError"] = err.Error()
				}
			}

		}

		if len(hsnstring) > 0 {
			err := InsertHSNInfo(Gst, hsnstring)
			if err != nil {
				logg.AnyError["HsnUpdationError"] = err.Error()
			}
		}

		// data_glid, err := GetGlidFromGstM(database, Gst)
		// if err != nil {
		// 	sendResponse(uniqID, w, 400, failure, errFetchDB, err, "", logg)
		// 	return
		// }

		// // Check if no GLIDs were found
		// if len(data_glid) == 0 {
		// 	sendResponse(uniqID, w, 404, failure, "No GLIDs found for the given GST", nil, "", logg)
		// 	return
		// }

		// // err1 := ProcessSingleGLID(data_glid["fk_glusr_usr_id"].(string))
		// //         if err1 != nil {
		// //                 logg.AnyError["meshupdationerror"] = err.Error()
		// //         }

		// // Process each GLID

		// for _, glid := range data_glid {
		// 	if err := ProcessSingleGLIDPubapilogging(glid, "/authadvanced/v1/gst"); err != nil {
		// 		// logg.AnyError["meshupdationerror"] = err.Error()
		// 		logg.AnyError["meshupdationerror_glid"] = glid + "-" + err.Error()
		// 	}
		// }

		if err := ProcessingGST(Gst, "/authadvanced/v1/gst"); err != nil {
			// logg.AnyError["meshupdationerror"] = err.Error()
			logg.AnyError["meshupdationerror_gst"] = Gst + "-" + err.Error()
		}

		message := Gst

		sendResponse(uniqID, w, 200, success, "", nil, message, logg)

		sendResponseKibana(uniqID, 200, success, "", nil, message, logg)

		return
	}

}

// InsertHSNInfo
func InsertHSNInfo(gst string, hsnstring string) (err error) {
	jsonStr, err := servapi.Hsnapi("PROD", gst, hsnstring)
	fmt.Println(jsonStr, "Prod-Insert-Prajjwal")
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

// Have to change some things in sendResponse and write2
func sendResponse(uniqID string, w http.ResponseWriter, httpcode int, status string, errorMsg string, err error, message string, logg Logg) {

	w.Header().Set("Content-Type", "application/json")

	logg.Response = Res{
		Code:    httpcode,
		Error:   errorMsg,
		Status:  status,
		Message: message,
		UniqID:  uniqID,
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

// Have to change some things in sendResponse and write2
func sendResponseKibana(uniqID string, httpcode int, status string, errorMsg string, err error, message string, logg Logg) {

	logg.Response = Res{
		Code:    httpcode,
		Error:   errorMsg,
		Status:  status,
		Message: message,
		UniqID:  uniqID,
	}
	if err != nil {
		logg.AnyError[errorMsg] = err.Error()
	}

	logg.RequestEndValue = utils.GetTimeInNanoSeconds()
	logg.ResponseTime = (logg.RequestEndValue - logg.RequestStartValue) / 1000000
	logg.ResponseTime_Float = (logg.RequestEndValue - logg.RequestStartValue) / 1000000
	logg.RequestStartValue = 0
	logg.RequestEndValue = 0

	kibanaLogg := logg
	kibanaLogg.Request.Responsedata = FlattenJson(kibanaLogg.Request.Responsedata)
	kibanaLogg.Result = nil
	kibanaLogg.ResultHSN = nil

	writeLogForKibana(kibanaLogg)
	return
}

// writeLog2 ...
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

// FlattenJson function
func FlattenJson(data string) string {
	str := strings.ReplaceAll(data, "{", "")
	str = strings.ReplaceAll(str, "}", "")
	str = strings.ReplaceAll(str, "\"", "")
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.ReplaceAll(str, " ", "")
	str = strings.ReplaceAll(str, ":", "=")
	str = strings.ReplaceAll(str, ",", ", ")
	return str
}

// writeLog2 ...
func writeLogForKibana(logg Logg) {

	fmt.Println("KIBANA")
	logsDir := serviceLogPath + utils.TodayDir()

	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		e := os.MkdirAll(logsDir, os.ModePerm)
		if e != nil {
			fmt.Println(e)
			return
		}
	}

	logsDir += "/" + logKibanaFileName

	jsonLog, _ := json.Marshal(logg)

	f, _ := os.OpenFile(logsDir, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()

	f.WriteString("\n" + string(jsonLog))

	fmt.Println("\n" + string(jsonLog))
	return
}

// Updated database connection function using provided details
// func getDBConnection() (*sql.DB, error) {
//         dbHost := "34.93.68.212"
//         dbPort := 5432
//         dbName := "approvalpg"
//         dbUser := "bi"
//         dbPassword := "bipass4impaypg"

//         // Construct the connection string
//         connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

//         // Open the database connection
//         return sql.Open("postgres", connStr)
// }

// Fetches the existing turnover data for a given GSTIN.
func fetchExistingTurnoverData(db *sql.DB, gstin string) ([]byte, error) {
	var jsonData []byte

	// Assuming gst_glusr_turnover is the JSONB column in the gst_gluser_masterdata table
	query := `SELECT gst_glusr_turnover FROM gst_gluser_masterdata WHERE gstin_number = $1`
	row := db.QueryRow(query, gstin)
	err := row.Scan(&jsonData)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

// Updates or appends turnover data based on the FY.
func updateOrAppendTurnoverData(existingData []byte, turnoverYear, turnoverValue string) ([]byte, error) {
	var data struct {
		Turnover []struct {
			FY    string `json:"fy"`
			Value string `json:"value"`
		} `json:"turnover"`
		UpdationDate string `json:"updation_date"`
	}

	// Unmarshal the existing data into our structure.
	// err := json.Unmarshal(existingData, &data)
	// if err != nil {
	//         return nil, err
	// }

	// Check if existingData is not nil and not an empty JSON object
	if len(existingData) > 0 && string(existingData) != "{}" {
		// Unmarshal the existing data into our structure.
		err := json.Unmarshal(existingData, &data)
		if err != nil {
			return nil, err
		}
	}

	// Check if the FY is already present.
	updated := false
	for i, turnover := range data.Turnover {
		if turnover.FY == turnoverYear {
			data.Turnover[i].Value = turnoverValue
			updated = true
			break
		}
	}

	// If the FY is not present, append a new entry.
	if !updated {
		data.Turnover = append(data.Turnover, struct {
			FY    string `json:"fy"`
			Value string `json:"value"`
		}{FY: turnoverYear, Value: turnoverValue})
	}

	// Sort the Turnover slice by financial year.
	sort.Slice(data.Turnover, func(i, j int) bool {
		return data.Turnover[i].FY < data.Turnover[j].FY
	})

	// Update the updation date to now.
	data.UpdationDate = time.Now().Format("2006-01-02 15:04:05")

	// Marshal the updated data back into JSON.
	return json.Marshal(data)
}

// Updates the master data business address in the database.
func updateMasterDataBusinessAddress(db *sql.DB, gstin string, result []byte) error {
	query := `UPDATE gst_gluser_masterdata SET gst_glusr_turnover = $1 WHERE gstin_number = $2`
	_, err := db.Exec(query, result, gstin)
	return err
}
func PreviousFinancialYear(currentDate time.Time) (string, string) {
	// Get the year and month of the current date
	currentYear, currentMonth, _ := currentDate.Date()

	// Determine the starting month of the financial year (April)
	financialYearStartMonth := time.April

	// Calculate the starting year of the financial year
	var financialYearStartYear int
	if currentMonth < financialYearStartMonth {
		financialYearStartYear = currentYear - 1
	} else {
		financialYearStartYear = currentYear
	}

	// Calculate the previous financial year
	previousFinancialYear := financialYearStartYear - 1

	// Convert years to strings
	previousFinancialYearStr := strconv.Itoa(previousFinancialYear)
	financialYearStartYearStr := strconv.Itoa(financialYearStartYear)

	return previousFinancialYearStr, financialYearStartYearStr
}

// func formatTurnover(input string) (turnoverValue, turnoverYear string) {
//         // Formatting logic
//         replacements := []struct {
//                 old string
//                 new string
//         }{
//                 {"<br/>", ""},
//                 {"Slab:", ""},
//         }

//         for _, r := range replacements {
//                 input = strings.ReplaceAll(input, r.old, r.new)
//         }
//         input = strings.TrimSpace(input)

//         // Extract turnover and year
//         if idx := strings.Index(input, "(For FY"); idx != -1 {
//                 turnoverValue = strings.TrimSpace(input[:idx])
//                 turnoverYear = strings.TrimSpace(input[idx+len("(For FY"):])

//                 if len(turnoverYear) > 0 && (turnoverYear[len(turnoverYear)-1] == ')' || isSpecialOrNonNumericASCII(rune(turnoverYear[len(turnoverYear)-1]))) {
//                         turnoverYear = turnoverYear[:len(turnoverYear)-1]
//                         turnoverYear = strings.TrimSpace(turnoverYear)
//                 }
//         } else {
//                 turnoverValue = input
//         }

//         if turnoverValue == "" {
//                 turnoverValue = "NA"
//         }

//         if turnoverYear == "" {
//                 currentDate := time.Now()
//                 previousFinancialYear, financialYearStartYear := PreviousFinancialYear(currentDate)
//                 turnoverYear = previousFinancialYear + "-" + financialYearStartYear
//         }

//         return turnoverValue, turnoverYear
// }

func isSpecialOrNonNumericASCII(char rune) bool {
	return char < ' ' || char > '~' || (char >= 0 && char <= 47) || (char >= 58 && char <= 64) || (char >= 91 && char <= 96) || (char >= 123 && char <= 127)
}

func TurnoverUpdation(turnover string, Gst string) error {
	pgConnection, err := db.GetDatabaseConnection("approvalPG")

	if err != nil {
		return err
	}

	existingData, err := fetchExistingTurnoverData(pgConnection, Gst)
	if err != nil {
		return err
	}

	turnoverValue, turnoverYear := formatTurnover(turnover)

	updatedData, err := updateOrAppendTurnoverData(existingData, turnoverYear, turnoverValue)
	if err != nil {
		return err
	}

	err1 := updateMasterDataBusinessAddress(pgConnection, Gst, updatedData)
	if err1 != nil {
		return err1
	}

	return nil
}
