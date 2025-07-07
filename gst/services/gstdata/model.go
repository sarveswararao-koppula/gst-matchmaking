package gstdata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	db "mm/components/database"
	"strconv"
	"strings"
	"time"
)

var (
	conn     *sql.DB
	stmtGlid *sql.Stmt
	stmtGST  *sql.Stmt
)

var (
	conn2    *sql.DB
	stmtGST2 *sql.Stmt
)

var (
	conn3     *sql.DB
	stmtGlid3 *sql.Stmt
)

var (
	conn4    *sql.DB
	stmtGST4 *sql.Stmt
)

// GetGSTRecords ...
func GetGSTRecords(database string, gstinNumber string) (map[string]interface{}, error) {
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
		
		GSTIN_NUMBER::text
		,TO_CHAR(GST_INSERTION_DATE,'DD-MM-YYYY') GST_INSERTION_DATE
		,STATE_NAME::text
		,PINCODE::text
		,bussiness_fields_add::text
		,BUSSINESS_FIELDS_ADD_DISTRICT::text
		,gstin_status
		,trade_name
		,trade_name_replaced
		,business_name
		,business_name_replaced
		,taxpayer_type::text
		,annual_turnover_slab::text
		,TO_CHAR(registration_date,'DD-MM-YYYY') registration_date
		,COALESCE(mobile_number::text, '')	AS mobile_number
    ,COALESCE(email_id::text,        '')	AS email_id
    ,COALESCE(gst_challan_mobile_by_befisc::text, '')	AS gst_challan_mobile_by_befisc
    ,COALESCE(gst_challan_email_by_befisc::text, '') AS gst_challan_email_by_befisc
        FROM
        GST_GLUSER_MASTERDATA
		WHERE GSTIN_NUMBER=$1;
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

	res := make(map[string]interface{})
	for _, v := range callrecords["queryData"].([]interface{}) {

		for k, v1 := range v.(map[string]interface{}) {
			res[k] = v1
		}
	}

	return res, nil
}

// GetGSTRecords consist of 46 columns ...
func GetGSTRecords36(database string, gstinNumber string) (map[string]interface{}, error) {
	if conn2 == nil {
		stmtGST2 = nil
		var err error
		if conn2, err = db.GetDatabaseConnection(database); err != nil {
			return nil, err
		}
	}

	if stmtGST2 == nil {
		var err error
		query := `
		SELECT
    gstin_number,
    business_name,
    centre_juri,
    to_char(registration_date,'DD-MM-YYYY') as registration_date,
    to_char(cancel_date,'DD-MM-YYYY') as cancel_date,
	TO_CHAR(GST_INSERTION_DATE,'DD-MM-YYYY') GST_INSERTION_DATE,
    business_constitution,
    business_activity_nature,
    gstin_status,
    to_char(last_update_date,'DD-MM-YYYY') as last_update_date,
    state_jurisdiction_code,
    state_juri,
    centre_jurisdiction_code,
    trade_name,
    bussiness_fields_add,
    bussiness_fields_pp,
    location,
    state_name,
    pincode::text as pincode,
    taxpayer_type,
    building_name,
    street,
    door_number,
    floor_number,
    longitude,
    lattitude,
    bussiness_place_add_nature,
    building_name_addl,
    street_addl,
    location_addl,
    door_number_addl,
    state_name_addl,
    floor_number_addl,    
	longitude_addl,
    lattitude_addl,
    pincode_addl::text as pincode_addl,
    nature_of_business_addl,
	mobile_number,
	email_id,
	annual_turnover_slab,
	gross_income,
	percent_of_tax_payment_in_cash,
	aadhar_authentication_status::text,
	ekyc_verification_status::text,
	core_business_activity_nature,
	proprieter_name,
	field_visit_conducted::text,
	gst_glusr_turnover::text,
	COALESCE(business_constitution_group_id,1927)::text as business_constitution_group_id
    FROM
    GST_GLUSER_MASTERDATA
    WHERE gstin_number=$1
    LIMIT 1
		`

		stmtGST2, err = conn2.Prepare(query)
		if err != nil {
			return nil, err
		}
	}

	var params []interface{}
	params = append(params, gstinNumber)

	callrecords, err := selectWithStmt(stmtGST2, params)

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

func GetGSTRecords75(database string, gstinNumber string) (map[string]interface{}, error) {
	if conn4 == nil {
		stmtGST4 = nil
		var err error
		if conn4, err = db.GetDatabaseConnection(database); err != nil {
			return nil, err
		}
	}

	if stmtGST4 == nil {
		var err error
		query := `
		SELECT
    gstin_number,
    business_name,
	business_constitution,
	businesstype,
    centre_juri,
    to_char(registration_date,'DD-MM-YYYY') as registration_date,
    to_char(cancel_date,'DD-MM-YYYY') as cancel_date,
	TO_CHAR(GST_INSERTION_DATE,'DD-MM-YYYY') GST_INSERTION_DATE,
    business_constitution,
    business_activity_nature,
    gstin_status,
    to_char(last_update_date,'DD-MM-YYYY') as last_update_date,
    state_jurisdiction_code,
    state_juri,
    centre_jurisdiction_code,
    trade_name,
    bussiness_fields_add,
    bussiness_fields_pp,
    location,
    state_name,
    pincode::text as pincode,
    taxpayer_type,
    building_name,
    street,
    door_number,
    floor_number,
    longitude,
    lattitude,
    bussiness_place_add_nature,
    building_name_addl,
    street_addl,
    location_addl,
    door_number_addl,
    state_name_addl,
    floor_number_addl,    
	longitude_addl,
    lattitude_addl,
    pincode_addl::text as pincode_addl,
    nature_of_business_addl,
	mobile_number,
	email_id,
	annual_turnover_slab,
	gross_income,
	percent_of_tax_payment_in_cash,
	aadhar_authentication_status::text,
	ekyc_verification_status::text,
	core_business_activity_nature,
	proprieter_name,
	field_visit_conducted::text,
	gst_glusr_turnover::text,
	fk_glusr_usr_id::text,
    to_char(date_of_verification,'DD-MM-YYYY') as date_of_verification,
    updated_by::text,
    bussiness_address_add,
    business_address_pp,
    business_name_pp,
    fk_gstin_turnover_id::text,
    gst_inserted_by::text,
    to_char(date_of_filing,'DD-MM-YYYY') as date_of_filing,
    to_char(filing_last_updation_date,'DD-MM-YYYY') as filing_last_updation_date,
    fk_gl_locality_id::text,
    trade_name_replaced,
    business_name_replaced,
    business_fields_add_replaced,
    business_address_add_replaced,
    building_name_replaced,
    street_replaced,
    location_replaced,
    door_number_replaced,
    floor_number_replaced,
    updated_by_screen,
    hist_comments,
    update_by_ip,
    update_by_ip_country,
    bussiness_fields_add_district,
    bussiness_fields_pp_district,
   turnover_year,
   businesstype,
   COALESCE(business_constitution_group_id,1927)::text as business_constitution_group_id
    FROM
    GST_GLUSER_MASTERDATA
    WHERE gstin_number=$1
    LIMIT 1
        `

		stmtGST4, err = conn4.Prepare(query)
		if err != nil {
			return nil, err
		}
	}

	var params []interface{}
	params = append(params, gstinNumber)

	callrecords, err := selectWithStmt(stmtGST4, params)

	if err != nil {
		conn4 = nil
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

// GetGlidRecords ...
func GetGlidRecords(database string, glid string) (map[string]interface{}, error) {

	glidInt, _ := strconv.Atoi(glid)

	if conn == nil {
		stmtGlid = nil
		var err error
		if conn, err = db.GetDatabaseConnection(database); err != nil {
			return nil, err
		}
	}

	if stmtGlid == nil {
		var err error
		query := `
                SELECT
				B.GSTIN_NUMBER::text
				,TO_CHAR(GST_INSERTION_DATE,'DD-MM-YYYY') GST_INSERTION_DATE
				,STATE_NAME::text
				,PINCODE::text
				,BUSSINESS_FIELDS_ADD_DISTRICT::text
				,bussiness_fields_add::text
				,taxpayer_type::text
				,annual_turnover_slab::text
				,TO_CHAR(registration_date,'DD-MM-YYYY') registration_date
                FROM
                GLUSR_USR_COMP_REGISTRATIONS A
                JOIN GST_GLUSER_MASTERDATA B ON(A.GST=B.GSTIN_NUMBER)
                WHERE A.FK_GLUSR_USR_ID=$1
                ;
        `
		stmtGlid, err = conn.Prepare(query)
		if err != nil {
			return nil, err
		}
	}

	var params []interface{}
	params = append(params, glidInt)

	callrecords, err := selectWithStmt(stmtGlid, params)

	if err != nil {
		conn = nil
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

// GetGstFromGlid ...
func GetGstFromGlid(database string, glid string) (map[string]interface{}, error) {

	glidInt, _ := strconv.Atoi(glid)

	if conn3 == nil {
		stmtGlid3 = nil
		var err error
		if conn3, err = db.GetDatabaseConnection(database); err != nil {
			return nil, err
		}
	}

	if stmtGlid3 == nil {
		var err error
		query := `
                SELECT
				GST::text
				FROM
                GLUSR_USR_COMP_REGISTRATIONS
                WHERE FK_GLUSR_USR_ID=$1
                ;
        `
		stmtGlid3, err = conn3.Prepare(query)
		if err != nil {
			return nil, err
		}
	}

	var params []interface{}
	params = append(params, glidInt)

	callrecords, err := selectWithStmt(stmtGlid3, params)

	if err != nil {
		conn3 = nil
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

// calculation of months
func DaysBetweenDates(date2 string) (int, error) {
	const layout = "02-01-2006" // The date format: dd-mm-yyyy

	t1 := time.Now() // Current date

	t2, err := time.Parse(layout, date2)
	if err != nil {
		return 0, err
	}

	duration := t2.Sub(t1)
	return int(duration.Hours() / 24), nil
}

// abs returns the absolute value of x.
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func GetSixthCharacter(input string) (byte, error) {
	// Check if input string is long enough
	if len(input) < 6 {
		return 0, fmt.Errorf("input string is too short")
	}

	// Return the 6th character
	return input[5], nil
}

func GetOwnershipType(char byte) string {
	switch char {
	case 'P':
		return "Individual - Proprietor"
	case 'F':
		return "Partnership Firm/Limited Liability Partnership"
	case 'C':
		return "Limited Company (Ltd/Pvt Ltd)"
	case 'H':
		return "HUF Firm (Hindu Undivided Family)"
	case 'A', 'T', 'B':
		return "Trust/Association of Person/Body of Individual"
	case 'J', 'G':
		return "Government/Local Authority/Artificial Judiciary"
	default:
		return "Unknown"
	}
}

func CheckLastWordLLP(s string) bool {
	// Trim any leading or trailing spaces
	trimmedString := strings.TrimSpace(s)

	// Check if the string is empty after trimming
	if trimmedString == "" {
		return false
	}

	// Split the string into words based on spaces
	words := strings.Split(trimmedString, " ")

	// Get the last word
	lastWord := words[len(words)-1]

	// Compare the last word with "llp"
	return strings.ToLower(lastWord) == "llp"
}
