package gstmmmodel

import (
	db "mm/components/database"
)

func GetGlidRecords(database string, glid int) (map[string]string, error) {

	res := make(map[string]string)

	pgConnection, err := db.GetDatabaseConnection(database)

	if err != nil {
		return res, err
	}

	query := `SELECT
        GLUSR_USR_COMPANYNAME
        ,GLUSR_USR_ZIP::text
        ,GLUSR_USR_FIRSTNAME
        ,GLUSR_USR_MIDDLENAME
        ,GLUSR_USR_LASTNAME
        ,GLUSR_USR_ADD1
        ,GLUSR_USR_ADD2
        ,GLUSR_USR_LOCALITY
        ,GLUSR_USR_LANDMARK
        ,GLUSR_USR_CITY
         FROM  GLUSR_USR_GST_CLEANUP
         WHERE FK_GLUSR_USR_ID = $1`

	var params []interface{}
	params = append(params, glid)

	callrecords, err := db.SelectQuerySql(pgConnection, query, params)

	if err != nil {
		return res, err
	}

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

func GetGSTRecords(database string, tradeNameReplaced string) ([]map[string]string, error) {

	res := make([]map[string]string, 0)

	pgConnection, err := db.GetDatabaseConnection(database)

	if err != nil {
		return res, err
	}

	query := `SELECT
                TRADE_NAME_REPLACED
                ,PINCODE::text
                ,BUSINESS_NAME_REPLACED
                ,BUSINESS_FIELDS_ADD_REPLACED
                ,BUSINESS_ADDRESS_ADD_REPLACED
                ,BUILDING_NAME_REPLACED
                ,STREET_REPLACED
                ,LOCATION_REPLACED
                ,DOOR_NUMBER_REPLACED::text
                ,FLOOR_NUMBER_REPLACED::text
                ,GSTIN_NUMBER
                ,GSTIN_STATUS
                ,TO_CHAR(GST_INSERTION_DATE,'DD-MM-YYYY') GST_INSERTION_DATE
                FROM
                GST_GLUSER_MASTERDATA
                WHERE TRADE_NAME_REPLACED=$1
                `

	var params []interface{}
	params = append(params, tradeNameReplaced)

	callrecords, err := db.SelectQuerySql(pgConnection, query, params)

	if err != nil {
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

func GetGSTRecords2(database string, tradeNameReplaced string, gstin_number string) ([]map[string]string, error) {

	res := make([]map[string]string, 0)

	pgConnection, err := db.GetDatabaseConnection(database)

	if err != nil {
		return res, err
	}

	query := `SELECT
                TRADE_NAME_REPLACED
                ,PINCODE::text
                ,BUSINESS_NAME_REPLACED
                ,BUSINESS_FIELDS_ADD_REPLACED
                ,BUSINESS_ADDRESS_ADD_REPLACED
                ,BUILDING_NAME_REPLACED
                ,STREET_REPLACED
                ,LOCATION_REPLACED
                ,DOOR_NUMBER_REPLACED::text
                ,FLOOR_NUMBER_REPLACED::text
                ,GSTIN_NUMBER
                ,GSTIN_STATUS
                ,TO_CHAR(GST_INSERTION_DATE,'DD-MM-YYYY') GST_INSERTION_DATE
                FROM
                GST_GLUSER_MASTERDATA
                WHERE GSTIN_NUMBER=$1
                AND TRADE_NAME_REPLACED=$2
                `

	var params []interface{}
	params = append(params, gstin_number, tradeNameReplaced)

	callrecords, err := db.SelectQuerySql(pgConnection, query, params)

	if err != nil {
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
