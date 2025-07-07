package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// SelectSingleRow function to execute sql queries
func SelectSingleRow(sqlConnection *sql.DB, query string, params []interface{}, timeOutSeconds ...int) (*sql.Row, error) {
	timeOut := 3
	if len(timeOutSeconds) > 0 {
		timeOut = timeOutSeconds[0]
	}
	statement, err := sqlConnection.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer statement.Close()
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, time.Second*time.Duration(timeOut))
	//defer cancel()
	result := statement.QueryRowContext(ctx, params...)
	return result, err
}

// SelectMultipleRows function to execute sql queries
func SelectMultipleRows(sqlConnection *sql.DB, query string, params []interface{}) (*sql.Rows, error) {
	statement, err := sqlConnection.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer statement.Close()
	result, err := statement.Query()
	return result, err
}

// SelectQuerySql function to execute sql queries
func SelectQuerySql(sqlConnection *sql.DB, query string, params []interface{}, timeOutSeconds ...int) (map[string]interface{}, error) {
	timeOut :=6
	if len(timeOutSeconds) > 0 {
		timeOut = timeOutSeconds[0]
	}
	statement, err := sqlConnection.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer statement.Close()
	//result, err := statement.Query(params...)
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(timeOut))
	defer cancel()
	result, err := statement.QueryContext(ctx, params...)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer result.Close()
	cols, err := result.Columns()
	if err != nil {
		fmt.Println(err.Error())
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

// ExecuteQuerySql function to execute sql queries
func ExecuteQuerySql(sqlConnection *sql.DB, query string, params []interface{}, returningIdRequired bool) (int64, error) {
	statement, err := sqlConnection.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer statement.Close()
	result, err := statement.Exec(params...)
	if err != nil {
		return 0, err
	}
	if !returningIdRequired {
		return 0, nil
	} else {
		returningId, err := result.LastInsertId()
		return returningId, err
	}
}

// ExecuteTransactionSql function to execute transaction sql
func ExecuteTransactionSql(sqlTransaction *sql.Tx, query string, params []interface{}, returningIdRequired bool) (int64, error) {
	statement, err := sqlTransaction.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer statement.Close()
	result, err := statement.Exec(params...)
	if err != nil {
		return 0, err
	}
	if !returningIdRequired {
		return 0, nil
	} else {
		returningId, err := result.LastInsertId()
		return returningId, err
	}
}

