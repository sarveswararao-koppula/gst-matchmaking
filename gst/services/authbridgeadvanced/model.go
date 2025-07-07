package authbridgeadvanced

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mm/components/constants"
	db "mm/components/database"
	"mm/utils"
	"os"

	// model "mm/services/livematchmakingapi" // Replace with the correct import path
	"database/sql"
	"net/http"
	"net/url"
	"strings" // Ensure the import path is correct
	"time"

	// "strconv"

	"github.com/rs/xid"
)

var (
	conn     *sql.DB
	stmtGlid *sql.Stmt
	stmtGST  *sql.Stmt
)

var (
	conn2     *sql.DB
	stmtGlid2 *sql.Stmt
)

// GetGSTRecords ...
func GetGSTRecords(database string, gstinNumber string) (map[string]string, error) {

	if conn == nil {
		stmtGST = nil
		var err error
		if conn, err = db.GetDatabaseConnection(database); err != nil {
			return nil, err
		}
	}

	if stmtGST == nil {
		var err error
		query := `
		SELECT 
		TRADE_NAME_REPLACED
		,trade_name::text
		,PINCODE::text
		,business_name::text
		,BUSINESS_NAME_REPLACED
		,BUSINESS_FIELDS_ADD_REPLACED
		,bussiness_fields_add::text
		,BUSINESS_ADDRESS_ADD_REPLACED
		,BUILDING_NAME_REPLACED
		,STREET_REPLACED
		,LOCATION_REPLACED
		,DOOR_NUMBER_REPLACED::text
		,FLOOR_NUMBER_REPLACED::text
		,GSTIN_NUMBER
		,GSTIN_STATUS
		,TO_CHAR(GST_INSERTION_DATE,'DD-MM-YYYY') GST_INSERTION_DATE
		,STATE_NAME
		,door_number::text
		,building_name::text
		,street::text
		,location::text
		,floor_number::text
		,business_constitution::text
		,core_business_activity_nature::text
		,proprieter_name::text
		,annual_turnover_slab::text
		,TO_CHAR(registration_date,'DD-MM-YYYY') registration_date
		,mobile_number::text
		,email_id::text
		,business_activity_nature::text
		,COALESCE(business_constitution_group_id,1927)::text as business_constitution_group_id
		FROM  GST_GLUSER_MASTERDATA 
		WHERE GSTIN_NUMBER = $1;
	`
		stmtGST, err = conn.Prepare(query)
		if err != nil {
			return nil, err
		}
	}

	var params []interface{}
	params = append(params, gstinNumber)

	callrecords, err := selectWithStmt(stmtGST, params)

	if err != nil {
		conn = nil
		return nil, err
	}

	res := make(map[string]string)
	for _, v := range callrecords["queryData"].([]interface{}) {

		for k, v1 := range v.(map[string]interface{}) {
			val := ""
			if v1 != nil {
				val, _ = v1.(string)
			}
			res[k] = val
		}
	}

	return res, nil
}

func selectWithStmt(statement *sql.Stmt, params []interface{}, timeOutSeconds ...int) (map[string]interface{}, error) {

	if statement == nil {
		return nil, errors.New("stmt is nil")
	}

	timeOut := 3
	if len(timeOutSeconds) > 0 {
		timeOut = timeOutSeconds[0]
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(timeOut))
	defer cancel()
	result, err := statement.QueryContext(ctx, params...)
	if err != nil {
		return nil, err
	}
	defer result.Close()
	cols, err := result.Columns()
	if err != nil {
		return nil, err
	}
	finalResult := make([]interface{}, 0)

	for result.Next() {
		data := make(map[string]interface{})
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		err := result.Scan(columnPointers...)
		if err != nil {
			fmt.Println(err.Error())
		}
		for i, colName := range cols {
			data[colName] = columns[i]
		}
		finalResult = append(finalResult, data)
	}
	returnResult := make(map[string]interface{})
	returnResult["queryData"] = finalResult
	return returnResult, err
}

func getGSTData(gst string) ([]map[string]string, error) {

	res := make([]map[string]string, 0)

	pgConnection, err := db.GetDatabaseConnection("approvalPG")

	if err != nil {
		return res, err
	}

	query := ""

	query = `
        select
                trade_name_replaced
                ,pincode::text
                ,business_name_replaced
                ,business_fields_add_replaced
                ,business_address_add_replaced
                ,building_name_replaced
                ,street_replaced
                ,location_replaced
                ,door_number_replaced::text
                ,floor_number_replaced::text
                ,gstin_number
                ,gstin_status
                ,to_char(gst_insertion_date,'dd-mm-yyyy') gst_insertion_date
                                ,mobile_number::text
                                ,email_id::text
                from
                gst_gluser_masterdata
                where gstin_number=$1
`

	var params []interface{}

	params = append(params, gst)

	challanrecords, err := db.SelectQuerySql(pgConnection, query, params)

	if err != nil {
		return res, err
	}

	for _, v := range challanrecords["queryData"].([]interface{}) {

		records := make(map[string]string)

		for k, v1 := range v.(map[string]interface{}) {
			val := ""
			if v1 != nil {
				val, _ = v1.(string)
			}
			records[k] = val
		}

		res = append(res, records)
	}

	return res, nil

}

// Generate a unique logging ID using xid
func generateUniqueLoggingID() string {
	id := xid.New()
	return id.String()
}

// Fetch GST from API and log responses
func fetchGST(glid string) (map[string]interface{}, error) {
	apiURL := fmt.Sprintf("http://users.imutils.com/wservce/users/otherdetail/?token=imobile@15061981&modid=WEBERP&glusrid=%s&type=CompRgst&AK=%s", glid, constants.ServerAK)

	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Print the raw response body for debugging
	log.Printf("API Response for GLID %s: %s\n", glid, string(body))

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	// Safely assert response["Response"]
	responseData, ok := response["Response"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type for 'Response' key")
	}

	// Safely assert responseData["Data"]
	data, ok := responseData["Data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type for 'Data' key")
	}

	return data, nil
}

// Process GST records to extract required fields
func processGSTRecords(gst string) (map[string]string, error) {
	gstRecords, err := GetGSTRecords("approvalPG", gst)
	if err != nil {
		return nil, err
	}

	// Map to store processed fields
	records := make(map[string]string)

	for k, v := range gstRecords {
		// Check if the value is empty or missing
		if v == "" {
			records[k] = "" // Assign an empty string
		} else {
			records[k] = v // Direct assignment since v is already a string
		}
	}

	// Extract required fields
	registrationDate := records["registration_date"]
	turnover, _ := formatTurnover(records["annual_turnover_slab"])
	annualTurnoverSlab := getShortName(turnover)
	proprieterName := records["proprieter_name"]
	coreBusinessActivity := records["core_business_activity_nature"]
	// businessConstitution := records["business_constitution"]
	business_constitution_group_id := records["business_constitution_group_id"]
    legalStatusValue := utils.LegalStatusRead(business_constitution_group_id)

	business_activity_nature := records["business_activity_nature"]

	// Return the required values in a map
	return map[string]string{
		"registration_date":             registrationDate,
		"annual_turnover_slab":          annualTurnoverSlab,
		"proprieter_name":               proprieterName,
		"core_business_activity_nature": coreBusinessActivity,
		"business_constitution":         legalStatusValue,
		"business_activity_nature":      business_activity_nature,
	}, nil
}

// Publish data to Pub API
func publishToPubAPI(details map[string]string, glid, gst, uniqueID string) error {
	apiURL := "http://prod-soa-rmq-api-mkp-messaging-imutils.imbi.prod/rmq/publish"
	timestamp := time.Now().Format("02-01-2006 03:04:05pm")

	msg := map[string]interface{}{
		"SERVICENAME":       "GLUSR_GST_DETAILS",
		"TIMESTAMP":         timestamp,
		"UNIQUE_LOGGING_ID": uniqueID,
		"GLUSR_USR_ID":      glid,
		"GST":               gst,
		"HISTORY_COLUMNS": map[string]string{
			"GLUSR_USR_UPDATEDBY":        "GST Auto Approval",
			"GLUSR_USR_IP":               "43.205.40.1",
			"GLUSR_USR_IP_COUNTRY":       "India",
			"GLUSR_USR_UPDATEDBY_ID":     "",
			"GLUSR_USR_UPDATEDBY_AGENCY": "",
			"GLUSR_USR_UPDATESCREEN":     "GST Auto Approval Process",
			"GLUSR_USR_UPDATEDBY_URL":    "",
		},
		"PROPRIETER_NAME":               details["proprieter_name"],
		"REGISTRATION_DATE":             details["registration_date"],
		"BUSINESS_CONSTITUTION":         details["business_constitution"],
		"CORE_BUSINESS_ACTIVITY_NATURE": details["core_business_activity_nature"],
		"ANNUAL_TURNOVER_SLAB":          details["annual_turnover_slab"],
		"BUSINESS_ACTIVITY_NATURE":      details["business_activity_nature"],
	}

	// Marshal the message into JSON
	msgBytes, _ := json.Marshal(msg)
	fmt.Println("msg json: ", string(msgBytes))
	data := url.Values{}
	data.Set("qname", "USER_GST_DETAILS")
	data.Set("rservice", "USER_GST_DETAILS")
	data.Set("rid", uniqueID)
	data.Set("msg", string(msgBytes))

	client := &http.Client{Timeout: 4 * time.Second}
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Make the API call and capture the response
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Log the Pub API response for debugging
	log.Printf("Pub API Response for GLID %s: %s\n", glid, string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to publish: %s", body)
	}

	fmt.Printf("Published GST details for GLID: %s\n", glid)
	return nil
}

func getShortName(turnover string) string {
	shortNameMap := map[string]string{
		"0 to 40 lakhs":       "0 - 40 L",
		"40 lakhs to 1.5 Cr.": "40 L - 1.5 Cr",
		"40 lakhs to 1.5 Cr":  "40 L - 1.5 Cr",
		"1.5 Cr. to 5 Cr.":    "1.5 - 5 Cr",
		"1.5 Cr to 5 Cr":      "1.5 - 5 Cr",
		"5 Cr. to 25 Cr.":     "5 - 25 Cr",
		"5 Cr to 25 Cr":       "5 - 25 Cr",
		"25 Cr. to 100 Cr.":   "25 - 100 Cr",
		"25 Cr to 100 Cr":     "25 - 100 Cr",
		"100 Cr. to 500 Cr.":  "100 - 500 Cr",
		"100 Cr to 500 Cr":    "100 - 500 Cr",
		"500 Cr. and above":   "> 500 Cr",
		"500 Cr and above":    "> 500 Cr",
		"NA":                  "NA",
		"":                    "NA",
	}

	shortName, ok := shortNameMap[turnover]
	if !ok {
		return "NA"
	}

	return shortName
}

func formatTurnover(input string) (turnoverValue, turnoverYear string) {
	// Formatting logic
	replacements := []struct {
		old string
		new string
	}{
		{"<br/>", ""},
		{"Slab:", ""},
		{"Rs.", ""},
	}

	for _, r := range replacements {
		input = strings.ReplaceAll(input, r.old, r.new)
	}
	input = strings.TrimSpace(input)

	// Extract turnover and year
	if idx := strings.Index(input, "(For FY"); idx != -1 {
		turnoverValue = strings.TrimSpace(input[:idx])
		turnoverYear = strings.TrimSpace(input[idx+len("(For FY"):])

		if len(turnoverYear) > 0 && (turnoverYear[len(turnoverYear)-1] == ')' || utils.IsSpecialOrNonNumericASCII(rune(turnoverYear[len(turnoverYear)-1]))) {
			turnoverYear = turnoverYear[:len(turnoverYear)-1]
			turnoverYear = strings.TrimSpace(turnoverYear)
		}

		if inidx := strings.Index(turnoverYear, "-"); inidx != -1 {
			turnoverYear = turnoverYear[:inidx+1] + turnoverYear[inidx+3:]
		}
	} else if  idx2 := strings.LastIndex(input, "("); idx2 != -1{
			turnoverValue = strings.TrimSpace(input[:idx2])
			turnoverYear  = strings.TrimSpace(input[idx2+1:])

			if len(turnoverYear) > 0 &&
           (turnoverYear[len(turnoverYear)-1] == ')' ||
            utils.IsSpecialOrNonNumericASCII(rune(turnoverYear[len(turnoverYear)-1]))) {
            turnoverYear = strings.TrimSpace(turnoverYear[:len(turnoverYear)-1])
        }
        // normalize “2020-2021” → “2020-21”
        if inidx := strings.Index(turnoverYear, "-"); inidx != -1 {
            turnoverYear = turnoverYear[:inidx+1] + turnoverYear[inidx+3:]
        }

	} else {
		turnoverValue = input
	}

	return turnoverValue, turnoverYear
}

// Function to process a single GLID
func ProcessSingleGLID(glid string) (err error) {
	uniqueID := generateUniqueLoggingID()

	// Fetch GST from the API
	data, err := fetchGST(glid)
	if err != nil {
		// log.Printf("Error fetching data for GLID %s: %v", glid, err)
		return err
	}

	gst, gstOk := data["GST"].(string)
	verificationSrcID, srcIDOk := data["FK_GST_VERIFICATION_SRC_ID"].(string)

	// Validate GST length and FK_GST_VERIFICATION_SRC_ID
	if !gstOk || len(gst) != 15 || !srcIDOk || (verificationSrcID != "1" && verificationSrcID != "2" && verificationSrcID != "3" && verificationSrcID != "4" && verificationSrcID != "5") {
		// log.Printf("Invalid GST or Verification Source ID for GLID %s", glid)
		return err
	}

	// Process GST records
	details, err := processGSTRecords(gst)
	if err != nil {
		// log.Printf("Error processing GST records for GLID %s: %v", glid, err)
		return err
	}

	// Publish to Pub API
	err = publishToPubAPI(details, glid, gst, uniqueID)
	if err != nil {
		// log.Printf("Error publishing for GLID %s: %v", glid, err)
		return err
	}

	return
	// fmt.Printf("Successfully processed GLID: %s\n", glid)
}

//adding

// writeLog2 ...
func WritePubApiLogForKibana(logg PubApiLogg) {

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

// Publish data to Pub API
func publishToPubAPILogging(details map[string]string, glid, gst, uniqueID string, gstapiurl string) error {
	apiURL := "http://prod-soa-rmq-api-mkp-messaging-imutils.imbi.prod/rmq/publish"
	timestamp := time.Now().Format("02-01-2006 03:04:05pm")

	msg := map[string]interface{}{
		"SERVICENAME":       "GLUSR_GST_DETAILS",
		"TIMESTAMP":         timestamp,
		"UNIQUE_LOGGING_ID": uniqueID,
		"GLUSR_USR_ID":      glid,
		"GST":               gst,
		"HISTORY_COLUMNS": map[string]string{
			"GLUSR_USR_UPDATEDBY":        "GST Auto Approval",
			"GLUSR_USR_IP":               "43.205.40.1",
			"GLUSR_USR_IP_COUNTRY":       "India",
			"GLUSR_USR_UPDATEDBY_ID":     "",
			"GLUSR_USR_UPDATEDBY_AGENCY": "",
			"GLUSR_USR_UPDATESCREEN":     "GST Auto Approval Process",
			"GLUSR_USR_UPDATEDBY_URL":    "",
		},
		"PROPRIETER_NAME":               details["proprieter_name"],
		"REGISTRATION_DATE":             details["registration_date"],
		"BUSINESS_CONSTITUTION":         details["business_constitution"],
		"CORE_BUSINESS_ACTIVITY_NATURE": details["core_business_activity_nature"],
		"ANNUAL_TURNOVER_SLAB":          details["annual_turnover_slab"],
		"BUSINESS_ACTIVITY_NATURE":      details["business_activity_nature"],
	}

	// Marshal the message into JSON
	msgBytes, _ := json.Marshal(msg)
	fmt.Println("msg json: ", string(msgBytes))
	data := url.Values{}
	data.Set("qname", "USER_GST_DETAILS")
	data.Set("rservice", "USER_GST_DETAILS")
	data.Set("rid", uniqueID)
	data.Set("msg", string(msgBytes))

	datamessage := string(msgBytes)

	client := &http.Client{Timeout: 4 * time.Second}
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		logFailureToKibana(glid, gst, apiURL, "Request creation failed", err, false, gstapiurl, datamessage)
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Make the API call and capture the response
	resp, err := client.Do(req)
	if err != nil {
		logFailureToKibana(glid, gst, apiURL, "API call failed", err, false, gstapiurl, datamessage)
		return err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logFailureToKibana(glid, gst, apiURL, "Response reading failed", err, false, gstapiurl, datamessage)
		return err
	}

	// Log the Pub API response for debugging
	log.Printf("Pub API Response for GLID %s: %s\n", glid, string(body))

	if resp.StatusCode != http.StatusOK {
		logFailureToKibana(glid, gst, apiURL, string(body), nil, false, gstapiurl, datamessage)
		return fmt.Errorf("failed to publish: %s", body)
	}

	// Log success
	logFailureToKibana(glid, gst, apiURL, "Successfully published GST details", nil, true, gstapiurl, datamessage)
	fmt.Printf("Published GST details for GLID: %s\n", glid)
	return nil
}

func logFailureToKibana(glid, gst, url, pubApiResponse string, err error, isSuccess bool, gstapiurl string, datamessage string) {
	status := 400
	if isSuccess {
		status = 200
	}

	var logg PubApiLogg

	logg.ServiceName = "publishToPubAPI"
	logg.RequestStart = utils.GetTimeStampCurrent()
	logg.ServiceURL = url
	logg.DataMessage = datamessage
	logg.GstPubApiUrl = gstapiurl
	logg.STATUS = status
	logg.PubApiResponse = pubApiResponse
	logg.Gst = gst
	// logg.Glid = glid

	if status == 400 {
		if err == nil {
			logg.PubApiError = "InvalidResponseOrTimeout"
		} else {
			logg.PubApiError = err.Error()
		}
	}

	WritePubApiLogForKibana(logg)
}

// Function to process a single GLID
func ProcessSingleGLIDPubapilogging(glid string, gstapiurl string) (err error) {
	uniqueID := generateUniqueLoggingID()

	// Fetch GST from the API
	data, err := fetchGST(glid)
	if err != nil {
		// log.Printf("Error fetching data for GLID %s: %v", glid, err)
		return err
	}

	gst, gstOk := data["GST"].(string)
	verificationSrcID, srcIDOk := data["FK_GST_VERIFICATION_SRC_ID"].(string)

	// Validate GST length and FK_GST_VERIFICATION_SRC_ID
	if !gstOk || len(gst) != 15 || !srcIDOk || (verificationSrcID != "1" && verificationSrcID != "2" && verificationSrcID != "3" && verificationSrcID != "4" && verificationSrcID != "5") {
		// log.Printf("Invalid GST or Verification Source ID for GLID %s", glid)
		return nil
	}

	// Process GST records
	details, err := processGSTRecords(gst)
	if err != nil {
		// log.Printf("Error processing GST records for GLID %s: %v", glid, err)
		return err
	}

	// Publish to Pub API
	err = publishToPubAPILogging(details, glid, gst, uniqueID, gstapiurl)
	if err != nil {
		// log.Printf("Error publishing for GLID %s: %v", glid, err)
		return err
	}

	return nil
	// fmt.Printf("Successfully processed GLID: %s\n", glid)
}

// GetGstFromGlid ...
/*
func GetGlidFromGst(database string, gst string) (map[string]interface{}, error) {

	// gstInt, _ := strconv.Atoi(gst)

	if conn2 == nil {
		stmtGlid2 = nil
		var err error
		if conn2, err = db.GetDatabaseConnection(database); err != nil {
			return nil, err
		}
	}

	if stmtGlid2 == nil {
		var err error
		query := `
			SELECT
							FK_GLUSR_USR_ID::text
							FROM
			GLUSR_USR_COMP_REGISTRATIONS
			WHERE GST=$1
			;
	`
		stmtGlid2, err = conn2.Prepare(query)
		if err != nil {
			return nil, err
		}
	}

	var params []interface{}
	params = append(params, gst)

	callrecords, err := selectWithStmt(stmtGlid2, params)

	if err != nil {
		conn2 = nil
		return nil, err
	}

	res := make(map[string]interface{})
	for _, v := range callrecords["queryData"].([]interface{}) {

		for k, v1 := range v.(map[string]interface{}) {
			res[k] = v1
		}
	}

	return res, nil
}
*/
// GetGlidFromGst queries the database to retrieve the GLIDs for a given GST.
func GetGlidFromGstM(database string, gst string) ([]string, error) {
	if conn2 == nil {
		stmtGlid2 = nil
		var err error
		if conn2, err = db.GetDatabaseConnection(database); err != nil {
			return nil, err
		}
	}

	// Ensure the connection is still valid
	if conn2 != nil {
		if err := conn2.Ping(); err != nil {
			conn2 = nil
			stmtGlid2 = nil
			return nil, fmt.Errorf("database connection is invalid: %w", err)
		}
	}

	if stmtGlid2 == nil {
		query := `SELECT FK_GLUSR_USR_ID::text FROM GLUSR_USR_COMP_REGISTRATIONS WHERE GST=$1;`
		var err error
		stmtGlid2, err = conn2.Prepare(query)
		if err != nil {
			return nil, err
		}
	}

	params := []interface{}{gst}
	rows, err := stmtGlid2.Query(params...)
	if err != nil {
		// conn2 = nil
		// stmtGlid2 = nil
		return nil, err
	}
	defer rows.Close()

	var glids []string
	for rows.Next() {
		var glid sql.NullString
		if err := rows.Scan(&glid); err != nil {
			return nil, err
		}
		if glid.Valid {
			glids = append(glids, glid.String)
		}
	}

	// Check for errors during row iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return glids, nil
}

//code added for passig gst to pubapi for sync the data from approval to mesh 

// Publish data to Pub API
func publishingDataToPubAPI(details map[string]string, gst string, uniqueID string, gstapiurl string) error {
	apiURL := "http://prod-soa-rmq-api-mkp-messaging-imutils.imbi.prod/rmq/publish"
	timestamp := time.Now().Format("02-01-2006 03:04:05pm")

	msg := map[string]interface{}{
		"SERVICENAME":       "GLUSR_GST_DETAILS",
		"TIMESTAMP":         timestamp,
		"UNIQUE_LOGGING_ID": uniqueID,
		"GST":               gst,
		"HISTORY_COLUMNS": map[string]string{
			"GLUSR_USR_UPDATEDBY":        "GST Auto Approval",
			"GLUSR_USR_IP":               "43.205.40.1",
			"GLUSR_USR_IP_COUNTRY":       "India",
			"GLUSR_USR_UPDATEDBY_ID":     "",
			"GLUSR_USR_UPDATEDBY_AGENCY": "",
			"GLUSR_USR_UPDATESCREEN":     "GST Auto Approval Process",
			"GLUSR_USR_UPDATEDBY_URL":    "",
		},
		"PROPRIETER_NAME":               details["proprieter_name"],
		"REGISTRATION_DATE":             details["registration_date"],
		"BUSINESS_CONSTITUTION":         details["business_constitution"],
		"CORE_BUSINESS_ACTIVITY_NATURE": details["core_business_activity_nature"],
		"ANNUAL_TURNOVER_SLAB":          details["annual_turnover_slab"],
		"BUSINESS_ACTIVITY_NATURE":      details["business_activity_nature"],
	}

	// Marshal the message into JSON
	msgBytes, _ := json.Marshal(msg)
	fmt.Println("msg json: ", string(msgBytes))
	data := url.Values{}
	data.Set("qname", "USER_GST_DETAILS")
	data.Set("rservice", "USER_GST_DETAILS")
	data.Set("rid", uniqueID)
	data.Set("msg", string(msgBytes))

	datamessage := string(msgBytes)

	client := &http.Client{Timeout: 4 * time.Second}
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		logFailureToKibana2(gst, apiURL, "Request creation failed", err, false, gstapiurl, datamessage)
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Make the API call and capture the response
	resp, err := client.Do(req)
	if err != nil {
		logFailureToKibana2(gst, apiURL, "API call failed", err, false, gstapiurl, datamessage)
		return err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logFailureToKibana2(gst, apiURL, "Response reading failed", err, false, gstapiurl, datamessage)
		return err
	}

	// Log the Pub API response for debugging
	log.Printf("Pub API Response for  %s: %s\n", gst, string(body))

	if resp.StatusCode != http.StatusOK {
		logFailureToKibana2(gst, apiURL, string(body), nil, false, gstapiurl, datamessage)
		return fmt.Errorf("failed to publish: %s", body)
	}

	logFailureToKibana2(gst, apiURL, "Successfully published GST details", nil, true, gstapiurl, datamessage)
	fmt.Printf("Published GST details for %s\n", gst)
	return nil
}

// Function to process a gst
func ProcessingGST(gst string, gstapiurl string) (err error) {
	uniqueID := generateUniqueLoggingID()
	// Process GST records ...
	details, err := processGSTRecords(gst)
	if err != nil {
		// log.Printf("Error processing GST records for GLID %s: %v", glid, err)
		return err
	}

	// Publish to Pub API
	err1 := publishingDataToPubAPI(details,gst, uniqueID,gstapiurl)
	if err1 != nil {
		// log.Printf("Error publishing for GLID %s: %v", glid, err)
		return err1
	}

	return nil
	// fmt.Printf("Successfully processed GLID: %s\n", glid)
}

func logFailureToKibana2(gst string, url, pubApiResponse string, err error, isSuccess bool, gstapiurl string, datamessage string) {
	status := 400
	if isSuccess {
		status = 200
	}

	var logg PubApiLogg

	logg.ServiceName = "publishToPubAPI"
	logg.RequestStart = utils.GetTimeStampCurrent()
	logg.ServiceURL = url
	logg.DataMessage = datamessage
	logg.GstPubApiUrl = gstapiurl
	logg.STATUS = status
	logg.PubApiResponse = pubApiResponse
	logg.Gst = gst

	if status == 400 {
		if err == nil {
			logg.PubApiError = "InvalidResponseOrTimeout"
		} else {
			logg.PubApiError = err.Error()
		}
	}

	WritePubApiLogForKibana(logg)
}
