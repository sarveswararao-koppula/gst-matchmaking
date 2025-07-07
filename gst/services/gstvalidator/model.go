package gstvalidator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	db "mm/components/database"
	//"strconv"
	"time"
)

var (
	conn     *sql.DB
	stmtGlid *sql.Stmt
	stmtGST  *sql.Stmt
)

var (
	conn3     *sql.DB
	stmtGlid3 *sql.Stmt
)

//GetGSTRecords ...
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

//GetGstFromGlid ...
func GetGstFromGlid(database string, gst string) (map[string]interface{}, error) {

	// glidInt, _ := strconv.Atoi(glid)

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
				FK_GLUSR_USR_ID::text
				FROM
                GLUSR_USR_COMP_REGISTRATIONS
                WHERE gst=$1
                ;
        `
		stmtGlid3, err = conn3.Prepare(query)
		if err != nil {
			return nil, err
		}
	}

	var params []interface{}
	params = append(params, gst)

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

