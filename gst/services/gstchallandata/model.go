package gstchallandata

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
        conn    *sql.DB
        stmtGST *sql.Stmt
)

var (
        conn2    *sql.DB
        stmtGST2 *sql.Stmt
)

var (
        conn3    *sql.DB
        stmtGST3 *sql.Stmt
)

var (
        conn4     *sql.DB
        stmtGlid4 *sql.Stmt
)

var (
        conn5     *sql.DB
        stmtGST5  *sql.Stmt
)
//GetGstFromGlid ...
func GetGstFromGlid(database string, glid string) (map[string]interface{}, error) {

        glidInt, _ := strconv.Atoi(glid)

        if conn4 == nil {
                stmtGlid4 = nil
                var err error
                if conn4, err = db.GetDatabaseConnection(database); err != nil {
                        return nil, err
                }
        }

        if stmtGlid4 == nil {
                var err error
                query := `
                SELECT
                                GST::text
                                FROM
                GLUSR_USR_COMP_REGISTRATIONS
                WHERE FK_GLUSR_USR_ID=$1
                ;
        `
                stmtGlid4, err = conn4.Prepare(query)
                if err != nil {
                        return nil, err
                }
        }

        var params []interface{}
        params = append(params, glidInt)

        callrecords, err := selectWithStmt(stmtGlid4, params)

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

func GetChallanDetails(database string, gstinNumber string) ([]map[string]string, error) {
        res := make([]map[string]string, 0)
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
                SELECT * FROM (SELECT gstin_number,return_period :: text,to_char(date_of_filing, 'dd-mon-yyyy' ) date_of_filing,status,
                to_char(entered_on, 'dd-mon-yyyy' ) entered_on,return_type,ROW_NUMBER()OVER(PARTITION BY return_type ORDER BY date_of_filing DESC) RN FROM (SELECT gstin_number,return_period,date_of_filing,status,entered_on,return_type
                from gst_challan_details  where gstin_number=$1) a) c WHERE rn<=10
                ;
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
        for _, v := range callrecords["queryData"].([]interface{}) {

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

func GetChallanAllDetails(database string, gstinNumber string) ([]map[string]string, error) {
        res := make([]map[string]string, 0)
        if conn5 == nil {
                stmtGST5 = nil
                var err error
                if conn5, err = db.GetDatabaseConnection(database); err != nil {
                        return nil, err
                }
        }
        if stmtGST5 == nil {
                var err error
                query := `
                SELECT * FROM (SELECT gstin_number,return_period :: text,to_char(date_of_filing, 'dd-mon-yyyy' ) date_of_filing,status,gst_challan_detail_id::text,is_valid,mode_of_filing,arn_number,
                to_char(entered_on, 'dd-mon-yyyy' ) entered_on,return_type,ROW_NUMBER()OVER(PARTITION BY return_type ORDER BY date_of_filing DESC) RN FROM (SELECT gstin_number,return_period,date_of_filing,status,entered_on,return_type,gst_challan_detail_id,is_valid,mode_of_filing,arn_number
                from gst_challan_details  where gstin_number=$1) a) c WHERE rn<=10
                ;
                `
                stmtGST5, err = conn5.Prepare(query)
                if err != nil {
                        return nil, err
                }
        }

        var params []interface{}
        params = append(params, gstinNumber)

        callrecords, err := selectWithStmt(stmtGST5, params)
        if err != nil {
                conn5 = nil
                return nil, err
        }
        for _, v := range callrecords["queryData"].([]interface{}) {

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


//GET LATEST DOF
func GetLatestDOF(database string, gst string) (map[string]string, error) {

        res := make(map[string]string)

        conn2, err := db.GetDatabaseConnection(database)

        if err != nil {
                return res, err
        }

        query := `SELECT TO_CHAR( MAX(DATE_OF_FILING) , 'dd-MM-yyyy') DATE_OF_FILING FROM gst_challan_details WHERE gstin_number=$1;`

        data, err := db.SelectQuerySql(conn2, query, []interface{}{gst})
        //fmt.Println(data, "Challan-Records")

        if err != nil {
                return res, err
        }

        for _, v := range data["queryData"].([]interface{}) {

                for k, v1 := range v.(map[string]interface{}) {
                        val := ""
                        if v1 != nil {

                                val, _ = v1.(string)
                        }
                        res[k] = val
                }
        }
        //fmt.Println(res, "Res")
        return res, nil
}


//AFTERUPDATIONDOF
func GetDOF(database string, gstinNumber string) (map[string]string, error) {
        if conn3 == nil {
                stmtGST3 = nil
                var err error
                if conn3, err = db.GetDatabaseConnection(database); err != nil {
                        return nil, err
                }
        }
        //fmt.Println(gstinNumber, "Dev-GetDof")
        if stmtGST3 == nil {
                var err error
                query := `
                select
    to_char(date_of_filing,'dd-mm-yyyy') date_of_filing
    from gst_gluser_masterdata where gstin_number=$1;
        `
                stmtGST3, err = conn3.Prepare(query)
                if err != nil {
                        return nil, err
                }
        }

        var params []interface{}
        params = append(params, gstinNumber)
        callrecords, err := selectWithStmt(stmtGST3, params)
        //fmt.Println("Dev-Challan-Records", callrecords)

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
        //fmt.Println(returnResult, "Dev-returnResult")
        //fmt.Println(finalResult, "Dev-finalResult")
        returnResult["queryData"] = finalResult
        return returnResult, err
}
