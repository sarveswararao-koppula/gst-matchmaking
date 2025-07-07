package livematchmakingapi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	db "mm/components/database"
	"strconv"
	"time"
)

var (
	conn     *sql.DB
	stmtGlid *sql.Stmt
	stmtGST  *sql.Stmt
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
		,gst_challan_email_by_befisc::text
		,gst_challan_mobile_by_befisc::text
		,COALESCE(business_constitution_group_id,1927)::text as business_constitution_group_id
		,locality::text
		,landmark::text
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

// GetGlidRecords ...
func GetGlidRecords(database string, glid string) (map[string]string, error) {

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
		A.GLUSR_USR_COMPANYNAME
		,A.GLUSR_USR_ZIP::text
		,A.GLUSR_USR_FIRSTNAME
		,A.GLUSR_USR_MIDDLENAME
		,A.GLUSR_USR_LASTNAME
		,A.GLUSR_USR_ADD1
		,A.GLUSR_USR_ADD2
		,A.GLUSR_USR_LOCALITY
		,A.GLUSR_USR_LANDMARK
		,A.GLUSR_USR_CITY
		,B.GLUSR_USR_CFIRSTNAME
		,B.GLUSR_USR_CLASTNAME
		,B.GLUSR_USR_STATE
		,B.glusr_usr_email
		,B.glusr_usr_email_alt
		FROM  GLUSR_USR_GST_CLEANUP A
		JOIN GLUSR_USR B ON(A.FK_GLUSR_USR_ID=B.GLUSR_USR_ID)
		WHERE A.FK_GLUSR_USR_ID = $1;
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
