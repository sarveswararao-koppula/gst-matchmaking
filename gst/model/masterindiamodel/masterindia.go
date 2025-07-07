package masterindiamodel

import (
	"errors"
	"fmt"
	db "mm/components/database"
	"mm/utils"
	"strconv"
	"strings"
	"time"
)

var loc *time.Location = utils.GetLocalTime()

func InsertChallanDetails(database string, gstin_number string, chllandata map[string]interface{}, opt ...string) (bool, error) {

	pgConnection, err := db.GetDatabaseConnection(database)
	if err != nil {
		return false, err
	}

	var dof_insert interface{} = nil

	if len(opt) == 0 {
		dof_str, _ := chllandata["dof"].(string)
		dof, er := time.Parse("02-01-2006", dof_str)
		if er != nil {
			return false, er
		}
		dof_str = dof.Format("2006-01-02")
		dof_insert = dof_str
	}

	var params []interface{}
	params = append(params, gstin_number)
	params = append(params, chllandata["arn"])
	params = append(params, dof_insert)
	params = append(params, chllandata["mof"])
	params = append(params, chllandata["ret_prd"])
	params = append(params, chllandata["rtntype"])
	params = append(params, chllandata["status"])
	params = append(params, chllandata["valid"])
	entered_on := time.Now().In(loc).Format("2006-01-02 15:04:05")
	params = append(params, entered_on)

	query := `INSERT INTO GST_CHALLAN_DETAILS
                 (
                    gstin_number,
                    arn_number,
                    date_of_filing,
                    mode_of_filing,
                    return_period,
                    return_type,
                    status,
                    is_valid,
                    entered_on
                 )
                 values($1,$2,$3,$4,$5,$6,$7,$8,$9)
                 `

	_, err = db.ExecuteQuerySql(pgConnection, query, params, false)

	if err != nil {
		return false, err
	}

	return true, nil
}

func CheckChallanDetails_dupl(database string, gstin_number string, chllandata map[string]interface{}) (int, error) {

	pgConnection, err := db.GetDatabaseConnection(database)

	cnt_str := "0"
	cnt := 0

	if err != nil {
		return 0, err
	}

	dof_str, _ := chllandata["dof"].(string)
	dof, er := time.Parse("02-01-2006", dof_str)
	if er != nil {
		return 0, er
	}

	dof_str = dof.Format("2006-01-02")

	query := `SELECT count(1)::text as cnt FROM GST_CHALLAN_DETAILS
    WHERE GSTIN_NUMBER=$1 AND date(DATE_OF_FILING)=$2 AND RETURN_PERIOD=$3 AND RETURN_TYPE=$4`

	var params_2 []interface{}
	params_2 = append(params_2, gstin_number, dof_str, chllandata["ret_prd"], chllandata["rtntype"])
	challanrecords, err := db.SelectQuerySql(pgConnection, query, params_2)

	if err != nil {
		return 0, err
	}

	for _, v := range challanrecords["queryData"].([]interface{}) {

		for k, v1 := range v.(map[string]interface{}) {
			if k == "cnt" && v1 != nil {
				cnt_str, _ = v1.(string)
			}
		}
	}

	cnt, _ = strconv.Atoi(cnt_str)

	return cnt, nil
}

func CheckChallanDetails(database string, params []interface{}) ([]interface{}, error) {

	//challandata := make(map[string]interface{})
	var challandata []interface{}

	pgConnection, err := db.GetDatabaseConnection(database)

	if err != nil {
		return challandata, err
	}

	query := `
SELECT FY,TO_CHAR( MAX(DATE_OF_FILING) , 'DD-MM-YYYY') DATE_OF_FILING
FROM (
    SELECT
        CASE WHEN (YY=$2 AND MM>=4) OR (YY=$1 AND MM<=3) THEN 'FY0'
             WHEN (YY=$3 AND MM>=4) OR (YY=$2 AND MM<=3) THEN 'FY1'
             WHEN (YY=$4 AND MM>=4) OR (YY=$3 AND MM<=3) THEN 'FY2'
             WHEN (YY=$5 AND MM>=4) OR (YY=$4 AND MM<=3) THEN 'FY3'
        END AS FY
        ,DATE_OF_FILING
    FROM (
            SELECT DATE_OF_FILING,DIV(RETURN_PERIOD,10000) MM,MOD(RETURN_PERIOD,10000) YY
            FROM GST_CHALLAN_DETAILS
            WHERE GSTIN_NUMBER=$6
         )A
    WHERE (YY=$1 AND MM<=3)
          OR YY=$2
          OR YY=$3
          OR YY=$4
          OR (YY=$5 AND MM>=4)
)A
GROUP BY FY
    `
	//FY: apr to mar

	challanrecords, err := db.SelectQuerySql(pgConnection, query, params)

	if err != nil {
		return challandata, err
	}

	for _, v := range challanrecords["queryData"].([]interface{}) {

		records := make(map[string]interface{})

		for k, v1 := range v.(map[string]interface{}) {
			records[k] = v1
		}

		challandata = append(challandata, records)
	}

	return challandata, nil

}

func UpdateGSTMasterData(database string, params []interface{}) (bool, error) {

	pgConnection, err := db.GetDatabaseConnection(database)

	if err != nil {
		return false, err
	}

	query := `UPDATE GST_GLUSER_MASTERDATA
                        SET
                        business_name=$2,
                centre_juri=$3,
                registration_date=$4,
                cancel_date=$5,
                business_constitution=$6,
                business_activity_nature=$7,
                gstin_status=$8,
                last_update_date=$9,
                state_jurisdiction_code=$10,
                state_juri=$11,
                centre_jurisdiction_code=$12,
                trade_name=$13,
                bussiness_fields_add=$14,
                location=$15,
                state_name=$16,
                pincode=$17,
                taxpayer_type=$18,
                building_name=$19,
                street=$20,
                door_number=$21,
                floor_number=$22,
                longitude=$23,
                lattitude=$24,
                bussiness_place_add_nature=$25,
                bussiness_address_add=$26,
                building_name_addl=$27,
                street_addl=$28,
                location_addl=$29,
                door_number_addl=$30,
                state_name_addl=$31,
                floor_number_addl=$32,
                longitude_addl=$33,
                lattitude_addl=$34,
                pincode_addl=$35,
                nature_of_business_addl=$36,
                gst_insertion_date=$37,
                gst_inserted_by=$38,
                bussiness_fields_add_district=$39,
                bussiness_fields_pp_district=$40,
                business_constitution_group_id=$41,
                landmark=$42,
                locality=$43,
                geo_code_lvl=$44,
                einvoice_status=$45
                WHERE GSTIN_NUMBER=$1`

	_, err = db.ExecuteQuerySql(pgConnection, query, params, false)

	if err != nil {
		return false, err
	}

	return true, nil
}

func UpdateMasterDataFilingDate(database string, params []interface{}) error {

	pgConnection, err := db.GetDatabaseConnection(database)

	if err != nil {
		return err
	}

	fmt.Println("database", database)

	query := `UPDATE GST_GLUSER_MASTERDATA
              SET
              DATE_OF_FILING=$2,
              FILING_LAST_UPDATION_DATE=$3
              WHERE GSTIN_NUMBER=$1`

	_, err = db.ExecuteQuerySql(pgConnection, query, params, false)

	if err != nil {
		return err
	}

	return nil
}

func InsertGSTMasterData(database string, params []interface{}) (bool, error) {

	pgConnection, err := db.GetDatabaseConnection(database)

	if err != nil {
		return false, err
	}

	query := `INSERT INTO GST_GLUSER_MASTERDATA
                 (
                 gstin_number,
                business_name,
                centre_juri,
                registration_date,
                cancel_date,
                business_constitution,
                business_activity_nature,
                gstin_status,
                last_update_date,
                state_jurisdiction_code,
                state_juri,
                centre_jurisdiction_code,
                trade_name,
                bussiness_fields_add,
                location,
                state_name,
                pincode,
                taxpayer_type,
                building_name,
                street,
                door_number,
                floor_number,
                longitude,
                lattitude,
                bussiness_place_add_nature,
                bussiness_address_add,
                building_name_addl,
                street_addl,
                location_addl,
                door_number_addl,
                state_name_addl,
                floor_number_addl,
                longitude_addl,
                lattitude_addl,
                pincode_addl,
                nature_of_business_addl,
                gst_insertion_date,
                gst_inserted_by,
                bussiness_fields_add_district,
                bussiness_fields_pp_district,
                business_constitution_group_id,
                landmark,
                locality,  
                geo_code_lvl , 
                einvoice_status
                )
                values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$37,$38,$39,$40,$41,$42,$43,$44,$45)
                `

	_, err = db.ExecuteQuerySql(pgConnection, query, params, false)

	if err != nil {
		return false, err
	}

	return true, nil
}

func GetGSTRecords(database string, gstin_number string) (map[string]interface{}, error) {

	gstdata := make(map[string]interface{})

	pgConnection, err := db.GetDatabaseConnection(database)

	if err != nil {
		return gstdata, err
	}

	query := `SELECT
            gstin_number,
            business_name,
            centre_juri,
            to_char(registration_date,'DD-MM-YYYY') as registration_date,
            to_char(cancel_date,'DD-MM-YYYY') as cancel_date,
            business_constitution,
            business_activity_nature,
            gstin_status,
            to_char(last_update_date,'DD-MM-YYYY') as last_update_date,
            state_jurisdiction_code,
            state_juri,
            centre_jurisdiction_code,
            trade_name,
            bussiness_fields_add,
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
            bussiness_address_add,
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
            to_char(gst_insertion_date,'dd-mm-yyyy hh24:mi:ss') as gst_insertion_date,
            gst_inserted_by::text as  gst_inserted_by,
            COALESCE(business_constitution_group_id,1927)::text as business_constitution_group_id
        FROM
        GST_GLUSER_MASTERDATA
        WHERE gstin_number=$1
        LIMIT 1`

	var params []interface{}
	params = append(params, gstin_number)

	gstrecords, err := db.SelectQuerySql(pgConnection, query, params)

	if err != nil {
		return gstdata, err
	}

	for _, v := range gstrecords["queryData"].([]interface{}) {

		for k, v1 := range v.(map[string]interface{}) {
			gstdata[k] = v1
		}
	}

	return gstdata, nil
}

// merpcsdgetgstdata
func GetGSTRecordsNew(database string, gstin_number string) (map[string]interface{}, error) {

	gstdata := make(map[string]interface{})

	pgConnection, err := db.GetDatabaseConnection(database)

	if err != nil {
		return gstdata, err
	}

	query := `SELECT
            gstin_number,
            business_name,
            centre_juri,
            to_char(registration_date,'DD-MM-YYYY') as registration_date,
            to_char(cancel_date,'DD-MM-YYYY') as cancel_date,
            business_constitution,
            business_activity_nature,
            gstin_status,
            to_char(last_update_date,'DD-MM-YYYY') as last_update_date,
            state_jurisdiction_code,
            state_juri,
            centre_jurisdiction_code,
            trade_name,
            bussiness_fields_add,
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
            bussiness_address_add,
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
            to_char(gst_insertion_date,'dd-mm-yyyy hh24:mi:ss') as gst_insertion_date,
            gst_inserted_by::text as  gst_inserted_by,
            COALESCE(business_constitution_group_id,1927)::text as business_constitution_group_id
        FROM
        GST_GLUSER_MASTERDATA
        WHERE gstin_number=$1
        LIMIT 1`

	var params []interface{}
	params = append(params, gstin_number)

	gstrecords, err := db.SelectQuerySql(pgConnection, query, params)

	if err != nil {
		return gstdata, err
	}

	for _, v := range gstrecords["queryData"].([]interface{}) {

		for k, v1 := range v.(map[string]interface{}) {
			gstdata[k] = v1
		}
	}

	return gstdata, nil
}

// GetExistsMasterdata ...
func GetExistsMasterdata(database string, gst string) (map[string]string, error) {

	res := make(map[string]string)

	conn, err := db.GetDatabaseConnection(database)

	if err != nil {
		return res, err
	}

	query := `select
    gstin_number
    ,to_char(gst_insertion_date,'dd-mm-yyyy') gst_insertion_date
    ,to_char(date_of_filing,'dd-mm-yyyy') date_of_filing
    ,to_char(registration_date,'dd-mm-yyyy') registration_date
    from gst_gluser_masterdata where gstin_number=$1`

	data, err := db.SelectQuerySql(conn, query, []interface{}{gst})

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

	return res, nil
}

func InsertAuthBridgeGSTMasterData(database string, params []interface{}) (bool, error) {

	pgConnection, err := db.GetDatabaseConnection(database)

	if err != nil {
		return false, err
	}

	query := `INSERT INTO GST_GLUSER_MASTERDATA
                 (
                 gstin_number,
                business_name,
                centre_juri,
                registration_date,
                cancel_date,
                business_constitution,
                business_activity_nature,
                gstin_status,
                mobile_number,
                state_jurisdiction_code,
                state_juri,
                email_id,
                trade_name,
                bussiness_fields_add,
                location,
                state_name,
                pincode,
                taxpayer_type,
                building_name,
                street,
                door_number,
                floor_number,
                annual_turnover_slab,
                gross_income,
                bussiness_place_add_nature,
                bussiness_address_add,
                building_name_addl,
                street_addl,
                location_addl,
                door_number_addl,
                state_name_addl,
                floor_number_addl,
                percent_of_tax_payment_in_cash,
                aadhar_authentication_status,
                pincode_addl,
                nature_of_business_addl,
                gst_insertion_date,
                gst_inserted_by,
                bussiness_fields_add_district,
                bussiness_fields_pp_district,
                ekyc_verification_status,
                core_business_activity_nature,
                proprieter_name,
                field_visit_conducted,
                business_constitution_group_id
                )
                values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$37,$38,$39,$40,$41,$42,$43,$44,$45)
                `

	_, err = db.ExecuteQuerySql(pgConnection, query, params, false)

	if err != nil {
		return false, err
	}

	return true, nil
}

func UpdateAuthBridgeGSTMasterData(database string, params []interface{}) (bool, error) {

	pgConnection, err := db.GetDatabaseConnection(database)

	if err != nil {
		return false, err
	}

	query := `UPDATE GST_GLUSER_MASTERDATA
                                        SET
                                        business_name=$2,
                        centre_juri=$3,
                        registration_date=$4,
                        cancel_date=$5,
                        business_constitution=$6,
                        business_activity_nature=$7,
                        gstin_status=$8,
                        mobile_number=$9,
                        state_jurisdiction_code=$10,
                        state_juri=$11,
                        email_id=$12,
                        trade_name=$13,
                        bussiness_fields_add=$14,
                        location=$15,
                        state_name=$16,
                        pincode=$17,
                        taxpayer_type=$18,
                        building_name=$19,
                        street=$20,
                        door_number=$21,
                        floor_number=$22,
                        annual_turnover_slab=$23,
                        gross_income=$24,
                        bussiness_place_add_nature=$25,
                        bussiness_address_add=$26,
                        building_name_addl=$27,
                        street_addl=$28,
                        location_addl=$29,
                        door_number_addl=$30,
                        state_name_addl=$31,
                        floor_number_addl=$32,
                        percent_of_tax_payment_in_cash=$33,
                        aadhar_authentication_status=$34,
                        pincode_addl=$35,
                        nature_of_business_addl=$36,
                        gst_insertion_date=$37,
                        gst_inserted_by=$38,
                        bussiness_fields_add_district=$39,
                        bussiness_fields_pp_district=$40,
                        ekyc_verification_status=$41,
            core_business_activity_nature=$42,
            proprieter_name=$43,
            field_visit_conducted=$44,
            business_constitution_group_id=$45
                                        WHERE GSTIN_NUMBER=$1`

	_, err = db.ExecuteQuerySql(pgConnection, query, params, false)

	if err != nil {
		return false, err
	}

	return true, nil
}

func UpdateGstBefiscData(database string, params []interface{}) (bool, error) {

	pgConnection, err := db.GetDatabaseConnection(database)

	if err != nil {
		return false, err
	}

	query := `UPDATE GST_GLUSER_MASTERDATA
                                        SET
                                        core_business_activity_nature=$2,
                        annual_turnover_slab=$3,
                        business_constitution=$4,
                        gstin_status=$5,
                        business_name=$6,
                        registration_date=$7,
                        state_juri=$8,
                        taxpayer_type=$9,
                        centre_juri=$10,
                        trade_name=$11,
                        gst_challan_email_by_befisc=$12,
                        gst_challan_mobile_by_befisc=$13,
                        gross_income=$14,
                        cancel_date=$15,
                        einvoice_status=$16,
                        field_visit_conducted=$17,
                        proprieter_name=$18,
                        gst_refresh_date_advance_challan=$19,
                        business_constitution_group_id=$20,
                        business_activity_nature=$21,
                        bussiness_fields_add=$22,
                        gst_insertion_date=$23
                                        WHERE GSTIN_NUMBER=$1`

	_, err = db.ExecuteQuerySql(pgConnection, query, params, false)

	if err != nil {
		return false, err
	}

	return true, nil
}

func InsertGstBefiscData(database string, params []interface{}) (bool, error) {

	pgConnection, err := db.GetDatabaseConnection(database)

	if err != nil {
		return false, err
	}

	query := `INSERT INTO GST_GLUSER_MASTERDATA
                (
                   gstin_number,
                   core_business_activity_nature,
                   annual_turnover_slab,
                   business_constitution,
                   gstin_status,
                   business_name,
                   registration_date,
                   state_juri,
                   taxpayer_type,
                   centre_juri,
                   trade_name,
                   gst_challan_email_by_befisc,
                   gst_challan_mobile_by_befisc,
                   gross_income,
                   cancel_date,
                   einvoice_status,
                   field_visit_conducted,
                   proprieter_name,
                   gst_refresh_date_advance_challan,
                   business_constitution_group_id,
                   business_activity_nature,
                   bussiness_fields_add,
                   gst_insertion_date
                )
                values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)
                `

	_, err = db.ExecuteQuerySql(pgConnection, query, params, false)

	if err != nil {
		return false, err
	}

	return true, nil
}

func InsertGSTMasterErrorData(database string, params []interface{}) (bool, error) {

	pgConnection, err := db.GetDatabaseConnection(database)

	if err != nil {
		return false, err
	}

	query := `INSERT INTO GST_PROCESS_ERROR_DETAILS
                 (
                 gst,
                                 error_code,
                                 error_text,
                                 added_date
                )
                values($1,$2,$3,$4)
                `

	_, err = db.ExecuteQuerySql(pgConnection, query, params, false)

	if err != nil {
		return false, err
	}

	return true, nil
}

func UpdatePincode(database string, params []interface{}) (bool, error) {

	pgConnection, err := db.GetDatabaseConnection(database)

	if err != nil {
		return false, err
	}

	query := `UPDATE GST_GLUSER_MASTERDATA
									SET
									bussiness_fields_add_district=$2
				WHERE GSTIN_NUMBER=$1`

	_, err = db.ExecuteQuerySql(pgConnection, query, params, false)

	if err != nil {
		return false, err
	}

	return true, nil
}

func InsertChallanDetailsBefisc(database string, gstin_number string, chllandata map[string]interface{}, opt ...string) (bool, error) {

	pgConnection, err := db.GetDatabaseConnection(database)
	if err != nil {
		return false, err
	}

	var dof_insert interface{} = nil

	if len(opt) == 0 {
		dof_str, _ := chllandata["dof"].(string)
		dof, er := time.Parse("02/01/2006", dof_str)
		if er != nil {
			return false, er
		}
		dof_str = dof.Format("2006/01/02")
		dof_insert = dof_str
	}

	var params []interface{}

	fy := chllandata["fy"].(string)
	taxp := chllandata["taxp"].(string)

	returnPeriod, _ := CalculateReturnPeriod(fy, taxp)
	// returnPeriodStr1 := strconv.Itoa(returnPeriod)

	params = append(params, gstin_number)
	params = append(params, chllandata["arn"])
	params = append(params, dof_insert)
	params = append(params, chllandata["mof"])
	params = append(params, returnPeriod)
	params = append(params, chllandata["rtntype"])
	params = append(params, chllandata["status"])
	// params = append(params, chllandata["valid"])
	entered_on := time.Now().In(loc).Format("2006-01-02 15:04:05")
	params = append(params, entered_on)

	query := `INSERT INTO GST_CHALLAN_DETAILS
                 (
                    gstin_number,
                    arn_number,
                    date_of_filing,
                    mode_of_filing,
                    return_period,
                    return_type,
                    status,
                    entered_on
                 )
                 values($1,$2,$3,$4,$5,$6,$7,$8)
                 `

	_, err = db.ExecuteQuerySql(pgConnection, query, params, false)

	if err != nil {
		return false, err
	}

	return true, nil
}

func CheckChallanDetails_dupl_befisc(database string, gstin_number string, chllandata map[string]interface{}) (int, error) {

	pgConnection, err := db.GetDatabaseConnection(database)

	cnt_str := "0"
	cnt := 0

	if err != nil {
		return 0, err
	}

	dof_str, _ := chllandata["dof"].(string)
	dof, er := time.Parse("02/01/2006", dof_str)
	if er != nil {
		return 0, er
	}

	dof_str = dof.Format("2006/01/02")

	query := `SELECT count(1)::text as cnt FROM GST_CHALLAN_DETAILS
    WHERE GSTIN_NUMBER=$1 AND date(DATE_OF_FILING)=$2 AND RETURN_PERIOD=$3 AND RETURN_TYPE=$4`

	var params_2 []interface{}

	fy := chllandata["fy"].(string)
	taxp := chllandata["taxp"].(string)

	returnPeriod, _ := CalculateReturnPeriod(fy, taxp)

	params_2 = append(params_2, gstin_number, dof_str, returnPeriod, chllandata["rtntype"])
	challanrecords, err := db.SelectQuerySql(pgConnection, query, params_2)

	if err != nil {
		return 0, err
	}

	for _, v := range challanrecords["queryData"].([]interface{}) {

		for k, v1 := range v.(map[string]interface{}) {
			if k == "cnt" && v1 != nil {
				cnt_str, _ = v1.(string)
			}
		}
	}

	cnt, _ = strconv.Atoi(cnt_str)

	return cnt, nil
}

// CalculateReturnPeriod calculates the return period in MMYYYY format without leading zeros for the month
func CalculateReturnPeriod(fy, taxp string) (int, error) {
	// Map of lowercase months to their numeric values
	months := map[string]int{
		"january": 1, "february": 2, "march": 3,
		"april": 4, "may": 5, "june": 6,
		"july": 7, "august": 8, "september": 9,
		"october": 10, "november": 11, "december": 12,
	}

	// Convert fy and taxp to lowercase
	fy = strings.ToLower(fy)
	taxp = strings.ToLower(taxp)

	// Split fiscal year
	years := strings.Split(fy, "-")
	if len(years) != 2 {
		return 0, errors.New("invalid fiscal year format")
	}

	// Parse the years
	firstYear, err1 := strconv.Atoi(years[0])
	secondYear, err2 := strconv.Atoi(years[1])
	if err1 != nil || err2 != nil {
		return 0, errors.New("invalid fiscal year values")
	}

	// Find the month number
	month, exists := months[taxp]
	if !exists {
		return 0, errors.New("invalid tax period month")
	}

	// Determine the year based on the month
	year := firstYear
	if month == 1 || month == 2 || month == 3 { // Jan, Feb, March -> 2nd year
		year = secondYear
	}

	// Concatenate month and year without leading zeros
	returnPeriod, err := strconv.Atoi(fmt.Sprintf("%d%d", month, year))
	if err != nil {
		return 0, errors.New("error generating return period")
	}

	return returnPeriod, nil
}
