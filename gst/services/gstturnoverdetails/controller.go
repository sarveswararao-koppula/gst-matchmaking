package gstturnoverdetails

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"mm/properties"
	"mm/utils"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"context"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

var mutex = &sync.Mutex{}

type turnoverapicontrols struct {
	RemoteAddress     string  `json:"RemoteAddress,omitempty"`
	RequestStart      string  `json:"RequestStart,omitempty"`
	RequestStartValue float64 `json:"RequestStartValue,omitempty"`
	RequestEnd        string  `json:"RequestEnd,omitempty"`
	RequestEndValue   float64 `json:"RequestEndValue,omitempty"`
	ResponseTime      string  `json:"ResponseTime,omitempty"`
	ResponseTime_Float float64 `json:"ResponseTime_Float,omitempty"` // float64 type
	ServiceName       string  `json:"ServiceName,omitempty"`
	ServicePath       string  `json:"ServicePath,omitempty"`
	ServiceURL        string  `json:"ServiceURL,omitempty"`
	Response          ResBody `json:"Response,omitempty"`
	RequestData       Rqst    `json:"RequestData,omitempty"`
}

// Rqst ...
type Rqst struct {
	Modid         string `json:"modid,omitempty"`
	Validationkey string `json:"validationkey,omitempty"`
	Glusr_ids     string `json:"glusr_ids,omitempty"`
}

// Assuming ResBody and TurnoverData are defined in yoursubpackage
type ResBody struct {
	Code       int         `json:"code"`
	Status     string      `json:"status"`
	QueryData  string      `json:"querydata,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	ErrMessage string      `json:"err_message,omitempty"`
}

type TurnoverData struct {
	UserID                        string    `json:"glusr_usr_id"`
	AnnualTurnover                string    `json:"annual_turnover"`
	TurnoverYear                  string    `json:"turnover_year"`
	ShortNameTurnover             string    `json:"shortname_turnover"`
	Core_business_activity_nature string    `json:"core_business_activity_nature"`
	LegalStatus                   string    `json:"legal_status"`
	RegistrationDate              time.Time `json:"registration_date"`
	State_name                    string    `json:"state_name"`
	Gstin_status                  string    `json:"gstin_status"`
	Proprieter_name               string    `json:"proprieter_name"`
	Business_name                 string    `json:"business_name"`
	Trade_name                    string    `json:"trade_name"`
}

// GetGSTTurnoverDetails handles the GET request to fetch GST turnover details.
func GetGSTTurnoverDetails(w http.ResponseWriter, r *http.Request) {

	var logs turnoverapicontrols
	logs.RequestStart = utils.GetTimeStampCurrent()
	logs.RequestStartValue = utils.GetTimeInNanoSeconds()
	logs.ServicePath = r.URL.Path
	logs.ServiceURL = "/gst-turnover-details/v1/gst"
	logs.RemoteAddress = utils.GetIPAdress(r)
	logs.RequestData = Rqst{}

	requiredParams := []string{"glusr_usr_ids", "modid", "ValidationKey"}
	for _, param := range requiredParams {
		if _, ok := r.URL.Query()[param]; !ok {
			http.Error(w, fmt.Sprintf("Missing '%s' parameter", param), http.StatusBadRequest)
			return
		}
	}

	userIDsString := r.URL.Query().Get("glusr_usr_ids")
	modid := r.URL.Query().Get("modid")
	validationKey := r.URL.Query().Get("ValidationKey")

	logs.RequestData.Modid = modid
	logs.RequestData.Glusr_ids = userIDsString
	logs.RequestData.Validationkey = validationKey

	if modid != "merp" && validationKey != "d2WjAYKxY4OkdnWmch==" {
		logs.Response.ErrMessage = "Invalid modid or validationkey"
		Write2Kibana(logs)
		http.Error(w, "Invalid modid or validationkey", http.StatusUnauthorized)
		return
	}

	fmt.Println("Testing")

	// Split user IDs string into a slice
	userIDs := strings.Split(userIDsString, ",")
	for i, id := range userIDs {
		userIDs[i] = strings.TrimSpace(id)
	}

	fmt.Println("testing print for go1.19")
	if len(userIDs) == 0 {
		logs.Response.ErrMessage = "No user IDs provided"
		Write2Kibana(logs)
		http.Error(w, "No user IDs provided", http.StatusBadRequest)
		return
	}

	if len(userIDs) > 5000 {
		logs.Response.ErrMessage = "too many user IDs; maximum allowed is 3000"
		Write2Kibana(logs)
		http.Error(w, "too many user IDs; maximum allowed is 3000", http.StatusBadRequest)
		return
	}

	userIDsMap := make(map[string]bool)
	for _, userID := range userIDs {
		userIDsMap[userID] = false
	}

	data, err := fetchTurnoverData(userIDs, userIDsMap, modid)
	if err != nil {
		logs.Response.ErrMessage = err.Error()
		Write2Kibana(logs)
		http.Error(w, "Error fetching data from database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	wrappedData := map[string]interface{}{
		"values": data,
	}

	res := ResBody{
		Code:   200,
		Status: "SUCCESS",
		Data:   wrappedData,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
	logs.Response.Code = res.Code
	logs.Response.Status = res.Status
	logs.Response.QueryData = strconv.Itoa(len(data))
	logs.RequestEndValue = utils.GetTimeInNanoSeconds()
	logs.RequestEnd = utils.GetTimeStampCurrent()
	logs.ResponseTime = fmt.Sprint((logs.RequestEndValue - logs.RequestStartValue) / 1000000)
	logs.ResponseTime_Float = (logs.RequestEndValue - logs.RequestStartValue) / 1000000
	Write2Kibana(logs)
}

// fetchTurnoverData retrieves turnover data from the database.
func fetchTurnoverData(userIDs []string, userIDsMap map[string]bool, modid string) ([]TurnoverData, error) {
	// Database connection details
	var dbHost, dbName string
	var dbPort int

	if properties.Prop.SERVICES_ENV == "PROD" {
		dbHost = "34.93.67.72"
		dbPort = 5432
		dbName = "approvalpg"
	} else if properties.Prop.SERVICES_ENV == "DEV" {
		dbHost = "34.100.240.197"
		dbPort = 5432
		dbName = "mesh_glusr"
	} else {
		return nil, fmt.Errorf("unknown environment setting")
	}

	if modid == "bi-gst" {
		if properties.Prop.SERVICES_ENV == "PROD" {
			dbHost = "34.93.67.72"
			dbPort = 5432
			dbName = "approvalpg"
		}
	}

	const dbUser = "bi"
	const dbPassword = "bipass4impaypg"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Open database connection
	dbConn, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName))
	if err != nil {
		return nil, err
	}
	defer dbConn.Close()

	if err = dbConn.PingContext(ctx); err != nil {
		return nil, err
	}

	// Database query
	query := `SELECT r.fk_glusr_usr_id, COALESCE(m.annual_turnover_slab, '') as annual_turnover_slab, COALESCE(m.core_business_activity_nature, '') as core_business_activity_nature,COALESCE(m.gstin_number, '') as gstin_number,COALESCE(m.business_constitution_group_id,1927)::text as business_constitution_group_id, registration_date, COALESCE(m.state_name, '') as state_name, COALESCE(m.gstin_status, '') as  gstin_status, COALESCE(m.proprieter_name, '') as  proprieter_name,COALESCE(m.business_name, '') as  business_name, COALESCE(m.trade_name, '') as trade_name
    from GLUSR_USR_COMP_REGISTRATIONS r
    JOIN GST_GLUSER_MasterData m ON r.gst = m.gstin_number
    WHERE r.fk_glusr_usr_id = ANY($1);`

	ctxQuery, cancelQuery := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancelQuery()

	rows, err := dbConn.QueryContext(ctxQuery, query, pq.Array(userIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Define a default registration date
	defaultRegistrationDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	var results []TurnoverData
	for rows.Next() {
		var userID, turnoverSlab, core_business_activity_nature, gstin_number, business_constitution_group_id, state_name, gstin_status, proprieter_name, business_name, trade_name string
		var registration_date *time.Time
		if err := rows.Scan(&userID, &turnoverSlab, &core_business_activity_nature, &gstin_number, &business_constitution_group_id, &registration_date, &state_name, &gstin_status, &proprieter_name, &business_name, &trade_name); err != nil {
			return nil, err
		}

		turnover, year := formatTurnover(turnoverSlab)

		if registration_date == nil {
			registration_date = &defaultRegistrationDate
		}

        legalStatusValue := utils.LegalStatusRead(business_constitution_group_id)

		results = append(results, TurnoverData{
			UserID:                        userID,
			AnnualTurnover:                turnover,
			TurnoverYear:                  year,
			ShortNameTurnover:             getShortName(turnover),
			Core_business_activity_nature: core_business_activity_nature,
			LegalStatus:                   legalStatusValue,
			RegistrationDate:              *registration_date,
			State_name:                    state_name,
			Gstin_status:                  gstin_status,
			Proprieter_name:               proprieter_name,
			Business_name:                 business_name,
			Trade_name:                    trade_name,
		})

		userIDsMap[userID] = true
	}

	for userID, hasData := range userIDsMap {
		if !hasData {
			results = append(results, TurnoverData{
				UserID:                        userID,
				AnnualTurnover:                "",
				TurnoverYear:                  "",
				ShortNameTurnover:             "NA",
				Core_business_activity_nature: "",
				LegalStatus:                   "",
				RegistrationDate:              defaultRegistrationDate,
				State_name:                    "",
			    Gstin_status:                  "",
			    Proprieter_name:               "",
			    Business_name:                 "",
			    Trade_name:                    "",
			})
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// func GetOwnershipType(char byte) string {
// 	switch char {
// 	case 'P':
// 		return "Individual - Proprietor"
// 	case 'F':
// 		return "Partnership Firm/Limited Liability Partnership"
// 	case 'C':
// 		return "Limited Company (Ltd/Pvt Ltd)"
// 	case 'H':
// 		return "HUF Firm (Hindu Undivided Family)"
// 	case 'A', 'T', 'B':
// 		return "Trust/Association of Person/Body of Individual"
// 	case 'J', 'G':
// 		return "Government/Local Authority/Artificial Judiciary"
// 	default:
// 		return "Unknown"
// 	}
// }

// func CheckLastWordLLP(s string) bool {
// 	// Trim any leading or trailing spaces
// 	trimmedString := strings.TrimSpace(s)

// 	// Check if the string is empty after trimming
// 	if trimmedString == "" {
// 		return false
// 	}

// 	// Split the string into words based on spaces
// 	words := strings.Split(trimmedString, " ")

// 	// Get the last word
// 	lastWord := words[len(words)-1]

// 	// Compare the last word with "llp"
// 	return strings.ToLower(lastWord) == "llp"
// }

// func GetSixthCharacter(input string) (byte, error) {
// 	// Check if input string is long enough
// 	if len(input) < 6 {
// 		return 0, fmt.Errorf("input string is too short")
// 	}

// 	// Return the 6th character
// 	return input[5], nil
// }

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

// formatTurnover formats the turnover information.
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
	}else if  idx2 := strings.LastIndex(input, "("); idx2 != -1{
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

// Write2Kibana ...
func Write2Kibana(logs turnoverapicontrols) {

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
