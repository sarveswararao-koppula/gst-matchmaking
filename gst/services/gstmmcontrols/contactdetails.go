package gstmmcontrols

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mm/components/constants"
	"mm/properties"
	"net/http"
	"strings"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Reverse mapping of keys to attributes for easy lookup
var keyToAttribute = map[string]int{
	"M1":  121,
	"L2":  156,
	"M2":  48,
	"L1":  120,
	"Ma1": 1293,
	"La1": 1294,
	"T1":  2074,
}

// Define User struct
type User struct {
	glusr_usr_email     string
	glusr_usr_email_alt string
}

func Contactdetails(glid string) ([]string, error) {

	// Step 1: Fetch mobile numbers from API

	mobileNumbers, attributes, err := Pnsapicall(glid, properties.Prop.SERVICES_ENV)
	if err != nil {
		fmt.Println("Error fetching numbers:", err)
		return nil,err
	}
	// Step 2: Store numbers in an array
	var numbersArray []string
	for a, number := range mobileNumbers {
		if a != "GSM" {
			numbersArray = append(numbersArray, number)
		}
	}

	// Step 3: Generate SQL query
	query := GenerateSQLQuery(numbersArray)

	// Step 4: Fetch data from database
	dbResults, err := FetchMobileDataFromDB(query)
	if err != nil {
		fmt.Println("Database Query Error:", err)
		return nil,err
	}

	//Store fetched database values in a variable
	var storedDBValues []map[string]string
	for _, record := range dbResults {
		filteredRecord := make(map[string]string)
		for key, value := range record {
			if val, ok := value.(string); ok && val != "" {
				filteredRecord[key] = val
			}
		}
		storedDBValues = append(storedDBValues, filteredRecord)
	}
	// Slice to store all GST numbers
	var gstNumbers []string

	for _, record := range storedDBValues {
		if gst, exists := record["gstin_number"]; exists {
			gstNumbers = append(gstNumbers, gst)
		}
	}

	// Step 5: Matching and verifying mobile numbers
	type VerifiedMapping struct {
		AttributeMob  string
		MobileNumber  string
		GstinNumber   string
		MobileMatched bool
		TacticalFlag  int
	}

	var verifiedMappings []VerifiedMapping

	var gstNumbersMobile []string

	var tacticalflag int
	tacticalflag = 0

	var flag int
	flag = 0

	// Iterate over stored DB values
	for _, record := range storedDBValues {
		mobileToMatch := record["mobile_number"]
		gstToMatch := record["gstin_number"]

		if mobileToMatch == "" {
			continue 
		}
		// Check if mobile number matches any attribute
		attribute, mobileMatched, found := matchMobileWithAttributes(attributes, mobileToMatch)
		if found {
			flag = 1
			muserverified, err1 := IsAttributeAlreadyVerified(glid, attribute)
			if err1 != nil {
				fmt.Println("Error verifying attribute:", err1.Error())
				continue
			}

			if muserverified {
				tacticalflag = 1
				verifiedMappings = append(verifiedMappings, VerifiedMapping{
					AttributeMob:  attribute,
					MobileNumber:  mobileMatched,
					GstinNumber:   gstToMatch,
					MobileMatched: true,
					TacticalFlag:  tacticalflag,
				})

				gstNumbersMobile = append(gstNumbersMobile, gstToMatch)
				// Debug: Print mapped GST and mobile with attribute
				fmt.Println("Matched & Verified:", "Attribute:", attribute, "Mobile:", mobileMatched, "GST:", gstToMatch)
			}
		}
	}

	// Store final verified mappings
	fmt.Println("Final Verified Mappings:", verifiedMappings)

	var gstNumbersEmail []string

	if flag == 0 {

		//glidemail
		query3 := GenerateGlidEmail(glid)

		// Step 4: Fetch data from database
		dbResults3, err := GlidEmails(query3)
		if err != nil {
			fmt.Println("Database Query Error:", err)
			return nil,err
		}
		// [map[glusr_usr_email:shreesadgurupackaging13@gmail.com glusr_usr_email_alt:]]

		var emailsArray []string

		// Extract valid emails
		for _, record := range dbResults3 {
			if email, exists := record["glusr_usr_email"].(string); exists && email != "" {
				emailsArray = append(emailsArray, email)
			}
			if altEmail, exists := record["glusr_usr_email_alt"].(string); exists && altEmail != "" {
				emailsArray = append(emailsArray, altEmail)
			}
		}

		// Generate SQL query for email
		query2 := GenerateSQLQueryEmail(emailsArray)

		// Step 4: Fetch data from database
		dbResults2, err := FetchEmailDataFromDB(query2)
		if err != nil {
			fmt.Println("Database Query Error:", err)
			return nil,err
		}
		// dbResults2: [map[email_id:shreesadgurupackaging13@gmail.com gst_challan_email_by_befisc: gstin_number:27BCPPS5552R1ZF]]

		var user User

		for _, record := range dbResults2 {

			// Extract values safely
			if email, ok := record["email_id"].(string); ok && email != "" {
				user.glusr_usr_email = email
			}
			if altEmail, ok := record["gst_challan_email_by_befisc"].(string); ok && altEmail != "" {
				user.glusr_usr_email_alt = altEmail
			}

			// users = append(users, user)
			user = User{
				glusr_usr_email:     user.glusr_usr_email,
				glusr_usr_email_alt: user.glusr_usr_email_alt,
			}

			if user.glusr_usr_email != "" || user.glusr_usr_email_alt != "" {

				// "157" // Attribute ID for glusr_usr_email_alt
				// "109" // Attribute ID for glusr_usr_email
				if user.glusr_usr_email != "" {
					a, err1 := IsAttributeAlreadyVerified(glid, "109")
					if err1 != nil {
						// logg.AnyError["userverifiedattributeemail"] = err1.Error()
						return nil,err1
					} else {
						if a {
							flag = 2
							if gst, ok := record["gstin_number"].(string); ok {
								gstNumbersEmail = append(gstNumbersEmail, gst)
							}
						}
					}
				} else {
					if user.glusr_usr_email_alt != "" {
						a, err2 := IsAttributeAlreadyVerified(glid, "157")
						if err2 != nil {
							// logg.AnyError["userverifiedattributeemail"] = err1.Error()
							return nil,err2
						} else {
							if a {
								flag = 2
								if gst, ok := record["gstin_number"].(string); ok {
									gstNumbersEmail = append(gstNumbersEmail, gst)
								}
							}
						}
					}
				}

			}
		}
	}
	fmt.Println("268GST numbers Mobile:", gstNumbersMobile)
	fmt.Println("269GST numbers Email:", gstNumbersEmail)
	fmt.Println("270GST numbers:", gstNumbers)

	if flag == 1 {
		return gstNumbersMobile,nil
	} else if flag == 2 {
		return gstNumbersEmail,nil
	}
	return gstNumbers,nil
}

func matchMobileWithAttributes(attributes map[string]map[string]string, mobile string) (string, string, bool) {
	for _, details := range attributes {
		if details["number"] == mobile {
			return details["attribute"], details["number"], true
		}
	}
	return "", "", false
}

// IsAttributeAlreadyVerified ...checking if gst is already verified for glid
func IsAttributeAlreadyVerified(glid string, attributeid string) (bool, error) {
	jsonStr, err := AttributeVerified(glid, attributeid)

	if err != nil {
		return false, err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return false, err
	}

	res := make(map[string]interface{})
	for k, v := range data {
		res[strings.ToLower(k)] = v
	}

	k, _ := res["response"].(map[string]interface{})
	k, _ = k["Data"].(map[string]interface{})
	k, _ = k[attributeid].(map[string]interface{})
	status, _ := k["Status"].(string)

	if strings.ToLower(status) == "verified" {
		return true, nil
	}

	if strings.ToLower(status) == "not verified" {
		return false, nil
	}

	return false, errors.New(jsonStr)
}

// AttributeVerified ... check if gst already verified
func AttributeVerified(glid string, attributeid string) (string, error) {

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	token := "imobile@15061981"
	modid := "BI"
	attrID := attributeid
	url := "http://users.imutils.com/wservce/users/verifiedDetail/?token=" + token + "&modid=" + modid + "&glusrid=" + glid + "&attribute_id=" + attrID + "&AK=" + constants.ServerAK

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	bodyString := string(bodyBytes)
	//fmt.Println(bodyString)
	return bodyString, nil
}
func Pnsapicall(glid string, env string) (map[string]string, map[string]map[string]string, error) {
	var AK string
	AK = "eyJ0eXAiOiJKV1QiLCJhbGciOiJzaGEyNTYifQ.eyJpc3MiOiJDUk9OIiwiYXVkIjoiNDMuMjA1LjQzLjgxLDEwLjEwLjEwLjIwIiwiZXhwIjoxODMzMTA1NDg2LCJpYXQiOjE2NzU0MDU0ODYsInN1YiI6ImJpLXV0aWxzLmludGVybWVzaC5uZXQifQ.qzIQ8EVfX9bNz52oLL3btBha1FANr8grVyLr6o1DzMs"
	urlAPI := ""

	if env == "DEV" {
		urlAPI = "http://34.93.67.39/wservce/users/pnssetting/"
	} else if env == "PROD" {
		urlAPI = "http://users.imutils.com/wservce/users/pnssetting/"
	}

	glusr_usr_id := glid

	req, err := http.NewRequest("POST", urlAPI, strings.NewReader("token=imobile@15061981&AK="+AK+"&modid=BI&glusrid="+glusr_usr_id))
	if err != nil {
		return nil, nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading response: %v", err)
	}

	var pnssettingresult map[string]interface{}
	err = json.Unmarshal(body, &pnssettingresult)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	// Check if the status is Success
	if status, ok := pnssettingresult["Response"].(map[string]interface{})["Status"].(string); ok && status == "Success" {
		numbers := extractPhoneNumbers(pnssettingresult)
		attributes := extractAttributes(numbers)
		end := time.Now()
		fmt.Printf("Total time taken: %v\n", end.Sub(start))
		return numbers, attributes, nil
	} else {
		return nil, nil, errors.New("API error: Status is not Success")
	}
}

func extractAttributes(numbers map[string]string) map[string]map[string]string {
	attributes := make(map[string]map[string]string)
	for key, number := range numbers {
		if attr, exists := keyToAttribute[key]; exists {
			attributes[key] = map[string]string{
				"number":    number,
				"attribute": fmt.Sprintf("%d", attr),
			}
		}
	}
	return attributes
}

func extractPhoneNumbers(data map[string]interface{}) map[string]string {
	numbers := make(map[string]string)
	if response, ok := data["Response"].(map[string]interface{}); ok {
		if data, ok := response["Data"].(map[string]interface{}); ok {
			for key, v := range data {
				if vMap, ok := v.(map[string]interface{}); ok {
					if number, ok := vMap["number"].(string); ok && number != "" {
						numbers[key] = number
					}
				}
			}
		}
	}
	return numbers
}

// Generate SQL Query from extracted numbers
func GenerateSQLQuery(numbers []string) string {
	if len(numbers) == 0 {
		return ""
	}

	// Convert numbers into SQL string format ('num1', 'num2', ...)
	numberList := "'" + strings.Join(numbers, "', '") + "'"

	query := fmt.Sprintf(`SELECT gstin_number, mobile_number, gst_challan_mobile_by_befisc 
		FROM gst_gluser_masterdata 
		WHERE mobile_number IN (%s) 
		OR gst_challan_mobile_by_befisc IN (%s);`, numberList, numberList)

	return query
}

// Generate SQL Query from extracted numbers
func GenerateSQLQueryEmail(emails []string) string {
	if len(emails) == 0 {
		return ""
	}

	// Convert numbers into SQL string format ('num1', 'num2', ...)
	emailList := "'" + strings.Join(emails, "', '") + "'"

	query := fmt.Sprintf(`SELECT gstin_number, email_id , gst_challan_email_by_befisc
		FROM gst_gluser_masterdata 
		WHERE email_id IN (%s) 
		OR  gst_challan_email_by_befisc IN (%s);`, emailList, emailList)

	return query
}

// Fetch data from PostgreSQL
func FetchMobileDataFromDB(query string) ([]map[string]interface{}, error) {
	var dbHost, dbName string
	var dbPort int

	// Set DB credentials based on environment
	dbHost = "34.93.67.72"
	dbPort = 5432
	dbName = "approvalpg"

	const dbUser = "bi"
	const dbPassword = "bipass4impaypg"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	dbConn, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName))
	if err != nil {
		return nil, err
	}
	defer dbConn.Close()

	if err = dbConn.PingContext(ctx); err != nil {
		return nil, err
	}

	// Execute query
	ctxQuery, cancelQuery := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancelQuery()

	rows, err := dbConn.QueryContext(ctxQuery, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}

	// Read data
	for rows.Next() {
		var gstinNumber, mobileNumber, gstChallanMobile sql.NullString
		if err := rows.Scan(&gstinNumber, &mobileNumber, &gstChallanMobile); err != nil {
			return nil, err
		}

		result := map[string]interface{}{
			"gstin_number":                 gstinNumber.String,
			"mobile_number":                mobileNumber.String,
			"gst_challan_mobile_by_befisc": gstChallanMobile.String,
		}
		results = append(results, result)
	}

	return results, nil
}

// Fetch data from PostgreSQL
func FetchEmailDataFromDB(query string) ([]map[string]interface{}, error) {
	var dbHost, dbName string
	var dbPort int

	// Set DB credentials based on environment
	dbHost = "34.93.67.72"
	dbPort = 5432
	dbName = "approvalpg"

	const dbUser = "bi"
	const dbPassword = "bipass4impaypg"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	dbConn, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName))
	if err != nil {
		return nil, err
	}
	defer dbConn.Close()

	if err = dbConn.PingContext(ctx); err != nil {
		return nil, err
	}

	// Execute query
	ctxQuery, cancelQuery := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancelQuery()

	rows, err := dbConn.QueryContext(ctxQuery, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}

	// Read data
	for rows.Next() {
		var gstinNumber, email_id, gst_challan_email_by_befisc sql.NullString
		if err := rows.Scan(&gstinNumber, &email_id, &gst_challan_email_by_befisc); err != nil {
			return nil, err
		}

		result := map[string]interface{}{
			"gstin_number":                gstinNumber.String,
			"email_id":                    email_id.String,
			"gst_challan_email_by_befisc": gst_challan_email_by_befisc.String,
		}
		results = append(results, result)
	}

	return results, nil
}

// Fetch data from PostgreSQL
func GlidEmails(query string) ([]map[string]interface{}, error) {
	var dbHost, dbName string
	var dbPort int

	// Set DB credentials based on environment
	dbHost = "34.93.67.72"
	dbPort = 5432
	dbName = "approvalpg"

	const dbUser = "bi"
	const dbPassword = "bipass4impaypg"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	dbConn, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName))
	if err != nil {
		return nil, err
	}
	defer dbConn.Close()

	if err = dbConn.PingContext(ctx); err != nil {
		return nil, err
	}

	// Execute query
	ctxQuery, cancelQuery := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancelQuery()

	rows, err := dbConn.QueryContext(ctxQuery, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}

	// Read data
	for rows.Next() {
		var glusr_usr_email, glusr_usr_email_alt sql.NullString
		if err := rows.Scan(&glusr_usr_email, &glusr_usr_email_alt); err != nil {
			return nil, err
		}

		result := map[string]interface{}{
			"glusr_usr_email":     glusr_usr_email.String,
			"glusr_usr_email_alt": glusr_usr_email_alt.String,
		}
		results = append(results, result)
	}

	return results, nil
}

// Generate SQL Query from extracted numbers
func GenerateGlidEmail(glid string) string {

	query := fmt.Sprintf(`
    SELECT 
        B.glusr_usr_email, 
        B.glusr_usr_email_alt 
    FROM GLUSR_USR_GST_CLEANUP A 
    JOIN GLUSR_USR B ON A.FK_GLUSR_USR_ID = B.GLUSR_USR_ID 
    WHERE A.FK_GLUSR_USR_ID IN (%s);`, glid)

	return query
}
