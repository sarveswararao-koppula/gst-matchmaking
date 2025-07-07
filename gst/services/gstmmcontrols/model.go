package gstmmcontrols

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	db "mm/components/database"
	"strings"
	"time"
)

var (
	conn     *sql.DB
	conn2    *sql.DB
	conn3    *sql.DB
	stmt     *sql.Stmt
	stmtGlid *sql.Stmt
	stmt3    *sql.Stmt
)

//GetGlidRecords ...
func GetGlidRecords(database string, glid int) ([]map[string]string, error) {

	res := make([]map[string]string, 0)

	if conn == nil {
		stmt = nil
		var err error
		if conn, err = db.GetDatabaseConnection(database); err != nil {
			return res, err
		}
	}

	if stmt == nil {
		var err error
		if stmt, err = prepareStmt(conn); err != nil {
			return res, err
		}
	}

	var params []interface{}
	params = append(params, glid)

	callrecords, err := selectQuerySQL(stmt, params)

	if err != nil {
		conn = nil
		return res, err
	}

	for _, v := range callrecords["queryData"].([]interface{}) {

		mp := make(map[string]string)
		for k, v1 := range v.(map[string]interface{}) {
			val := ""
			if v1 != nil {
				val, _ = v1.(string)
			}
			mp[k] = val
		}
		res = append(res, mp)
	}

	return res, nil
}

func GetGlidRecordsContactdetails(database string, glid int, gstnumbers []string) ([]map[string]string, error) {

	res := make([]map[string]string, 0)

	if conn == nil {
		stmt = nil
		var err error
		if conn, err = db.GetDatabaseConnection(database); err != nil {
			return res, err
		}
	}

	// if stmt == nil {
	// 	var err error
	// 	if stmt, err = prepareStmt(conn); err != nil {
	// 		return res, err
	// 	}
	// }

	if len(gstnumbers) == 0 {
		return res, nil
	}

	// Prepare the SQL statement dynamically based on the number of GST numbers
	stmt, err := prepareStmtContactdetails(conn, len(gstnumbers))
	if err != nil {
		return res, err
	}
	defer stmt.Close()

	var params []interface{}
	params = append(params, glid)

	for _, gst := range gstnumbers {
		params = append(params, gst) // Append each GST number separately
	}

	callrecords, err := selectQuerySQL(stmt, params)

	if err != nil {
		conn = nil
		return res, err
	}

	for _, v := range callrecords["queryData"].([]interface{}) {

		mp := make(map[string]string)
		for k, v1 := range v.(map[string]interface{}) {
			val := ""
			if v1 != nil {
				val, _ = v1.(string)
			}
			mp[k] = val
		}
		res = append(res, mp)
	}

	return res, nil
}

func selectQuerySQL(statement *sql.Stmt, params []interface{}, timeOutSeconds ...int) (map[string]interface{}, error) {

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

func prepareStmt(conn *sql.DB) (*sql.Stmt, error) {

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
,B.TRADE_NAME_REPLACED
,B.PINCODE::text
,B.BUSINESS_NAME_REPLACED
,B.BUSINESS_FIELDS_ADD_REPLACED
,B.bussiness_fields_add::text
,B.BUSINESS_ADDRESS_ADD_REPLACED
,B.BUILDING_NAME_REPLACED
,B.STREET_REPLACED
,B.LOCATION_REPLACED
,B.DOOR_NUMBER_REPLACED::text
,B.FLOOR_NUMBER_REPLACED::text	
,B.GSTIN_NUMBER
,B.GSTIN_STATUS
,B.STATE_NAME
,TO_CHAR(B.GST_INSERTION_DATE,'DD-MM-YYYY') GST_INSERTION_DATE 
,C.GLUSR_USR_CFIRSTNAME
,C.GLUSR_USR_CLASTNAME
,C.GLUSR_USR_STATE
,C.GLUSR_USR_CUSTTYPE_ID::text
FROM GLUSR_USR_GST_CLEANUP A 
JOIN GST_GLUSER_MASTERDATA B ON(A.GLUSR_USR_COMPANYNAME=B.TRADE_NAME_REPLACED) 
JOIN GLUSR_USR C ON(C.GLUSR_USR_ID=A.FK_GLUSR_USR_ID)
WHERE A.FK_GLUSR_USR_ID = $1  AND trim(A.GLUSR_USR_COMPANYNAME)!='';
;`

	stmt, err := conn.Prepare(query)
	if err != nil {
		return nil, err
	}

	return stmt, nil
}
// Function to prepare a SQL statement with dynamic IN clause
func prepareStmtContactdetails(conn *sql.DB, numGst int) (*sql.Stmt, error) {
	if numGst == 0 {
		return nil, errors.New("no GST numbers provided")
	}

	// Create placeholders dynamically: $2, $3, $4, ..., based on numGst
	placeholders := make([]string, numGst)
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+2) // Start from $2 since $1 is for glid
	}
	gstPlaceholder := "(" + strings.Join(placeholders, ", ") + ")"

	// Construct SQL query dynamically
	query := fmt.Sprintf(`
	SELECT 
		A.GLUSR_USR_COMPANYNAME,
		A.GLUSR_USR_ZIP::text,
		A.GLUSR_USR_FIRSTNAME,
		A.GLUSR_USR_MIDDLENAME,
		A.GLUSR_USR_LASTNAME,
		A.GLUSR_USR_ADD1,
		A.GLUSR_USR_ADD2,
		A.GLUSR_USR_LOCALITY,
		A.GLUSR_USR_LANDMARK,
		A.GLUSR_USR_CITY,
		B.TRADE_NAME_REPLACED,
		B.PINCODE::text,
		B.BUSINESS_NAME_REPLACED,
		B.BUSINESS_FIELDS_ADD_REPLACED,
		B.bussiness_fields_add::text,
		B.BUSINESS_ADDRESS_ADD_REPLACED,
		B.BUILDING_NAME_REPLACED,
		B.STREET_REPLACED,
		B.LOCATION_REPLACED,
		B.DOOR_NUMBER_REPLACED::text,
		B.FLOOR_NUMBER_REPLACED::text,
		B.GSTIN_NUMBER,
		B.GSTIN_STATUS,
		B.STATE_NAME,
		TO_CHAR(B.GST_INSERTION_DATE,'DD-MM-YYYY') GST_INSERTION_DATE,
		C.GLUSR_USR_CFIRSTNAME,
		C.GLUSR_USR_CLASTNAME,
		C.GLUSR_USR_STATE,
		C.GLUSR_USR_CUSTTYPE_ID::text
	FROM GLUSR_USR_GST_CLEANUP A 
	JOIN GST_GLUSER_MASTERDATA B ON(A.GLUSR_USR_COMPANYNAME=B.TRADE_NAME_REPLACED) 
	JOIN GLUSR_USR C ON(C.GLUSR_USR_ID=A.FK_GLUSR_USR_ID)
	WHERE A.FK_GLUSR_USR_ID = $1  
	AND trim(A.GLUSR_USR_COMPANYNAME) != '' 
	AND B.GSTIN_NUMBER IN %s;
	`, gstPlaceholder)

	// Prepare statement
	stmt, err := conn.Prepare(query)
	if err != nil {
		return nil, err
	}

	return stmt, nil
}

//GetDisabledInformation
func GetDisabledGlidRecords(database string, glidInt int) (map[string]string, error) {

	if conn2 == nil {
		stmtGlid = nil
		var err error
		if conn2, err = db.GetDatabaseConnection(database); err != nil {
			return nil, err
		}
	}

	if stmtGlid == nil {
		var err error
		query := `
		SELECT FK_GLUSR_USR_ID::text,GLUSR_DISABLE_ERRMSG_VALUE::text from GLUSR_DISABLE_ERRMSG  where FK_GLUSR_USR_ID = $1;
	`
		stmtGlid, err = conn2.Prepare(query)
		if err != nil {
			return nil, err
		}
	}

	var params []interface{}
	params = append(params, glidInt)

	callrecords, err := selectWithStmt(stmtGlid, params)

	if err != nil {
		conn2 = nil
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

//empFCP custtype_changes and company_name keyword changes

func GetCusttypeCompanyGlidRecords(database string, glidInt int) (map[string]string, error) {

	if conn3 == nil {
		stmt3 = nil
		var err error
		if conn3, err = db.GetDatabaseConnection(database); err != nil {
			return nil, err
		}
	}

	if stmt3 == nil {
		var err error
		query := `
		SELECT glusr_usr_id::text,glusr_usr_custtype_name::text,glusr_usr_companyname::text from GLUSR_USR  where glusr_usr_id = $1;
	`
		stmt3, err = conn3.Prepare(query)
		if err != nil {
			return nil, err
		}
	}

	var params []interface{}
	params = append(params, glidInt)

	callrecords, err := selectWithStmt(stmt3, params)

	if err != nil {
		conn3 = nil
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
