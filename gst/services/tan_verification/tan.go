package tan_verification

import (
        "bytes"
        "encoding/json"
        "errors"
        "fmt"
        "io/ioutil"
        db "mm/components/database"
        "mm/utils"
        "net/http"
        "net/url"
        "regexp"
        "strconv"
        "strings"
        "time"
        //"database/sql"
)

//var(
//      conn2     *sql.DB
//)

type PaymentData struct {
        Code           int
        Timestamp      int
        Transaction_id string
        Data           map[string]interface{}
}
type TanData struct {
        GLID         string
        COMPANY_NAME string
        PINCODE      string
        ADDRESS      string
        STATECD      string
        TANNO        string
        FIRSTNAME    string
        LASTNAME     string
        MIDNAME      string
}

// Map of state codes to state names
var stateCodeToStateName = map[int]string{
	1:  "Andaman & Nicobar",
	2:  "Andhra Pradesh",
	3:  "Arunachal Pradesh",
	4:  "Assam",
	5:  "Bihar",
	6:  "Chandigarh",
	7:  "Dadra and Nagar Haveli and Daman and Diu",
	9:  "Delhi",
	10: "Goa",
	11: "Gujarat",
	12: "Haryana",
	13: "Himachal Pradesh",
	14: "Jammu & Kashmir",
	15: "Karnataka",
	16: "Kerala",
	17: "Lakshadweep",
	18: "Madhya Pradesh",
	19: "Maharashtra",
	20: "Manipur",
	21: "Meghalaya",
	22: "Mizoram",
	23: "Nagaland",
	24: "Odisha",
	25: "Pondicherry",
	26: "Punjab",
	27: "Rajasthan",
	28: "Sikkim",
	29: "Tamil Nadu",
	30: "Tripura",
	31: "Uttar Pradesh",
	32: "West Bengal",
	33: "Chhattisgarh",
	34: "Uttarakhand",
	35: "Jharkhand",
	36: "Telangana",
	37: "Ladakh",
}

func removeKeywords(name string, keywords []string) string {
	upperName := strings.ToUpper(name)
	for _, keyword := range keywords {
		// Replace the keyword with an empty string in the upper case version
		upperName = strings.ReplaceAll(upperName, keyword, "")
	}
	return upperName
}

func normalizeString(input string) string {
	// Convert to lower case
	input = strings.ToLower(input)
	// Trim spaces from the beginning and end
	input = strings.TrimSpace(input)
	// Replace multiple spaces with a single space
	space := regexp.MustCompile(`\s+`)
	input = space.ReplaceAllString(input, " ")
	return input
}

func isCompanyNameContained(companyName1 string, companyName2 string) string {
	// Normalize the company names
	normalizedCompanyName1 := normalizeString(companyName1)
	normalizedCompanyName2 := normalizeString(companyName2)

	// Check if one company name is contained within the other
	if strings.Contains(normalizedCompanyName2, normalizedCompanyName1) || strings.Contains(normalizedCompanyName1, normalizedCompanyName2) {
		return "yes"
	} else {
		return "no"
	}
}

func Getawttoken() (string,error) {
        startt := utils.GetTimeInNanoSeconds()
        //S3log.ExtApihit = "Y"
        url := "https://api.sandbox.co.in/authenticate"
        client := &http.Client{Timeout: 15 * time.Second}
        req, _ := http.NewRequest("POST", url, nil)

        req.Header.Add("Accept", "application/json")
        //req.Header.Add("x-api-key", "key_live_aR9RPIV24FjMdF7elcTgMJASe7pchIaK")
        //req.Header.Add("x-api-secret", "secret_live_A9AItGKgaTXxDKQLlxLayx2AYP9nBKUp")
	req.Header.Add("x-api-key", "key_live_PS44cPs9Inoql3B0lESQXUb8F7v1xbsR")
	req.Header.Add("x-api-secret", "secret_live_1xQ2m90ajq5o9WdtBG4snMipj3G1JVoc")
        req.Header.Add("x-api-version", "1.0")
        res, err := client.Do(req)

        // S3log.Result["Rensponse from sandbox gettoken api"] = res
        // Write2S3(&S3log)
        if err != nil {
                //      fmt.Println(err)
                // S3log.TicketCreated = "Y"
                // CreateTicket(Tanidd, "", "external gettoken sandbox api error ", Companyid)
                S3log.Result["external sandbox api error"] = err.Error()
                Write2S3(&S3log)
                return "",err
        }
        endd := utils.GetTimeInNanoSeconds()
        S3log.ApiResponsetime["API_RESPONSE_GETWTTOKEN "] = (endd - startt) / 1000000
        defer res.Body.Close()
        body, err := ioutil.ReadAll(res.Body)
        if err != nil {
                //fmt.Println(err)
        // S3log.TicketCreated = "Y"
                // CreateTicket(Tanidd, "", "external gettoken sandbox api error ", Companyid)
                S3log.Result["external sandbox api error"] = err.Error()
                Write2S3(&S3log)
                return "",err
        }
        var result map[string]interface{}
        err = json.Unmarshal(body, &result)
        if err != nil {
                //fmt.Println(err)
                S3log.Result["json sandbox api error"] = err.Error()
                Write2S3(&S3log)
                return "",err
        }

        return (result["access_token"]).(string) , nil

}

func tanVerification(TAN_ID string, GLUSR_ID int64, count int64, companyidd string) (map[string]interface{}, string, string) {

        awt_token, err:= Getawttoken()
        if err!=nil{
                CreateTicket(Tanidd, "", "Token Retrieval Error:  " + err.Error(), Companyid)  
                return make(map[string]interface{}), "XXXX", ""
        }

        var pay PaymentData
        tandata := ""
        url := "https://api.sandbox.co.in/itd/portal/public/tans/" + TAN_ID + "?consent=y&reason=For%20KYC%20of%20the%20organization"
        method := "GET"
        startt := utils.GetTimeInNanoSeconds()
        S3log.ExtApihit = "Y"
        client := &http.Client{Timeout: 15 * time.Second }
        req, err := http.NewRequest(method, url, nil)

        if err != nil {
                //fmt.Println(err)
                S3log.Result["tan data api error "] = err.Error()
                Write2S3(&S3log)
                return pay.Data, "XXXX", tandata
        }
        req.Header.Add("Authorization", awt_token)
       // req.Header.Add("x-api-key", "key_live_aR9RPIV24FjMdF7elcTgMJASe7pchIaK")
	req.Header.Add("x-api-key", "key_live_PS44cPs9Inoql3B0lESQXUb8F7v1xbsR")
        req.Header.Add("x-api-version", "1.0")
        req.Header.Add("Accept", "application/json")

        res, err := client.Do(req)
        // S3log.Result["Response of sandbox api in tanverifyapi"] = res
        if err != nil {
                //fmt.Println(err)
                S3log.Result["tan data api error"] = err.Error()
                Write2S3(&S3log)
                return pay.Data, "XXXX", tandata
        }
        endd := utils.GetTimeInNanoSeconds()
        S3log.ApiResponsetime["API_RESPONSE_TANDATA "] = (endd - startt) / 1000000
        defer res.Body.Close()
        body, err := ioutil.ReadAll(res.Body)
        if err != nil {
                //fmt.Println(err)
                S3log.Result["tan data api error"] = err.Error()
                Write2S3(&S3log)
                return pay.Data, "XXXX", tandata
        }
        tandata = string(body)
        var result map[string]interface{}
        err = json.Unmarshal(body, &result)
        if err != nil {
                //fmt.Println(err)
                S3log.Result["tan data json api error"] = err.Error()
                Write2S3(&S3log)
                return pay.Data, "XXXX", tandata
        }
        var c float64
         c = result["code"].(float64)
        if c == 200 {
                _ = json.Unmarshal(body, &pay)
                //      return pay.Data, "", tandata
                d := pay.Data
                //fmt.Println("data", d)
                var t interface{}
                t = d["messages"]
                //fmt.Println("data", t)
                v := t.([]interface{})
                if len(v) != 0 {
                        g := v[0]
                        j := g.(map[string]interface{})
                        fmt.Println(j["code"])
                        if strings.Compare(j["type"].(string), "ERROR") == 0 {
                                CreateTicket(Tanidd, j["code"].(string), j["desc"].(string), Companyid)
                                // fmt.Println("ok")
                                  S3log.Result["statuscode_tanapi"] =  j["code"].(string)
                                   S3log.Result["description_tanapi"] =  j["desc"].(string)
                                return pay.Data, "XXXX", tandata
                }
                } else {
                        return pay.Data, "", tandata
                }
        } else {
                str := ""
		fmt.Println(str)
                if count < 0 {
                        pay.Data, str, tandata = tanVerification(TAN_ID, GLUSR_ID, 1,companyidd)
                } else {
                        statuscode := fmt.Sprintf("%f", c)
                        errmessage := result["message"].(string)
                        S3log.Result["statuscode"] = statuscode
                        S3log.Result["errmessage"] = errmessage
                        CreateTicket(TAN_ID, statuscode, errmessage, companyidd)
                        S3log.Result["error_ticket"]="done"
                        Write2S3(&S3log)
                }
                return pay.Data, "XXXX", tandata
        }
        S3log.Result["tandata from external tan api"] = tandata
        //      Write2S3(&S3log)
        return pay.Data, "", tandata
}

func InsertInDB(tan TanData) error {
        mainConnection, err := db.GetDatabaseConnection("approvalPG")
        if err != nil {
                //fmt.Println(err.Error())
                CreateTicket(Tanidd, "", "error while connecting to db in insertdb", Companyid)
                S3log.Result["error while connecting to db in insertdb"] = err.Error()
                Write2S3(&S3log)
                return errors.New("error while connecting to DB")
        }
        if tan.STATECD == "" {
                tan.STATECD = "0"
        }
        if tan.PINCODE == "" {
                tan.PINCODE = "0"
        }
        query := `INSERT INTO iil_tan_master_data (tan_no,company_name,namefirst,namelast,namemid,address,state_code,pincode,master_data_insertion_date) VALUES ('` + tan.TANNO + `','` + tan.COMPANY_NAME + `','` + tan.FIRSTNAME + `','` + tan.LASTNAME + `','` + tan.MIDNAME + `','` + tan.ADDRESS + `',` + tan.STATECD + `,` + tan.PINCODE + `,current_date) `
        // fmt.Println("hi", query)
        params := make([]interface{}, 0)
        _, err = db.ExecuteQuerySql(mainConnection, query, params, false)
        // fmt.Println("hi", err)
        if err != nil {
                CreateTicket(Tanidd, "", "error while updating payment data", Companyid)
                S3log.Result["error while updating payment data"] = err.Error()
                Write2S3(&S3log)
                return errors.New("error while updating payment data")
        }
        return nil
}

func TanVerifyApi(tan TanData) int64 {
        param := `{
                "VALIDATION_KEY":      "af7f0273997b9b290bd7c57aa19f36c2",
                "action_flag":         "SP_VERIFY_ATTRIBUTE",
                "GLUSR_USR_ID":         "` + tan.GLID + `",
                "ATTRIBUTE_ID":        "347",
                "ATTRIBUTE_VALUE":     "` + tan.TANNO + `",
                "VERIFIED_BY_ID":      "32789",
                "VERIFIED_BY_NAME":    "WEBERP",
                "VERIFIED_BY_AGENCY":  "Mobile",
                "VERIFIED_BY_SCREEN":  "Tan Verification Scheduler",
                "VERIFIED_URL":        "",
                "VERIFIED_IP":         "107.22.229.251",
                "VERIFIED_IP_COUNTRY": "India",
                "VERIFIED_COMMENTS":   "Tan",
                "VERIFIED_AUTHCODE":   "PHONE"
                }`
        startt := utils.GetTimeInNanoSeconds()
        //S3log.ExtApihit = "Y"
		// S3log.Verified = "Y"
        resp, err := http.NewRequest("POST", "http://service.intermesh.net/user/verification", bytes.NewBufferString(param))
        resp.Header.Set("Content-Type", "text/plain")
        client := &http.Client{Timeout: 4 * time.Second}
        response, err := client.Do(resp)
        if err != nil {
                //fmt.Println("failure1", err)
                        CreateTicket(Tanidd, "", "failure1 in tanverifyapi", Companyid)
                S3log.Result["failure1 in tanverifyapi"] = err.Error()
                Write2S3(&S3log)
                return 0
        }
        endd := utils.GetTimeInNanoSeconds()
        S3log.ApiResponsetime["API_RESPONSE_TANVERIFY "] = (endd - startt) / 1000000
        defer response.Body.Close()
        body, _ := ioutil.ReadAll(response.Body)
        //S3log.Result["response of tanverifyapi that verifying tan "] = string(body)
        //fmt.Println(string(body))
        var res1 map[string]interface{}
        err = json.Unmarshal(body, &res1)
        if err != nil {
                CreateTicket(Tanidd, res1["STATUS"].(string), "tan_ver_api_status", Companyid)
                S3log.Result["VerificationApiError "] = err.Error()
                Write2S3(&S3log)
                return 0
                //       fmt.Println(err)
        }
	fmt.Println("TAN Verification api response: ",res1)
        //S3log.Verification="Y"
        S3log.Result["tan_ver_api_status"] = res1["STATUS"].(string)
        S3log.Result["tan_ver_api_message"] = res1["MESSAGE"].(string)
        S3log.Result["tan_ver_api_code"] = res1["CODE"].(string)
        //fmt.Println("s3 log pointer in tan verifiction",S3log)
        // Write2S3(&S3log)
        S3log.Verified = "Y"
        
        return 1
}


func StringToInteger(stringValue string) int {
        intValue := 0
        var err error
        if len(stringValue) > 0 {
                intValue, err = strconv.Atoi(stringValue)
                if err != nil {
                        fmt.Println(err)
                        intValue = 0
                }
        } else {
                intValue = 0
        }
        return intValue
}

func Similarity(glid string, account_name string) (float64, error) {
        company_name := ""

        mainConnection, err := db.GetDatabaseConnection("approvalPG")
        //fmt.Println("mainConnection  similarity is ",mainConnection,"error in similarity",err)
        if err != nil {
                //fmt.Println(err.Error())
                //fmt.Println("error while connecting to DB")
                CreateTicket(Tanidd, "", "error while connecting to DB in similarity", Companyid)
                S3log.Result["error while connecting to DB in similarity"] = err.Error()
                Write2S3(&S3log)
                return 0.0,err
        }

        GLID := int64(StringToInteger(glid))
        // fmt.Println("SIMILARITY ERROR ", GLID)
        //query := `SELECT COMPANY FROM sts_company WHERE sts_fk_glusr_id = :GLID`
        query := `SELECT GLUSR_USR_COMPANYNAME FROM glusr_usr WHERE GLUSR_USR_ID = $1`
        params := make([]interface{}, 1)
        params[0] = GLID
        result, err := db.SelectQuerySql(mainConnection, query, params)
        if err != nil {
                //fmt.Println(err)
                //fmt.Println("error while executing query")
                CreateTicket(Tanidd, "", "error while executing query", Companyid)
                S3log.Result["error while executing query"] = err.Error()
                Write2S3(&S3log)
                return 0.0,err
        }
        if result != nil {
                if queryData, dataExists := result["queryData"]; dataExists {
                        if queryData != nil {
                                resultSet := queryData.([]interface{})
                                if len(resultSet) == 1 {
                                        for _, payRecord := range resultSet {

                                                payRecordResult := payRecord.(map[string]interface{})
                                                //name, nameExists := payRecordResult["COMPANY"]
                                                name, nameExists := payRecordResult["glusr_usr_companyname"]
                                                //fmt.Println("count found from query", count)
                                                if nameExists && name != nil {
                                                        //fmt.Println("hi****************")
                                                        company_name = name.(string)
                                                        //fmt.Println("SIMILARITY FUNCTION INSIDE COMPANY NAME ",company_name)
                                                }

                                        }
                                }
                        }
                }
        }
        //fmt.Println("SIMILARITY ERROR company ", company_name)
        S3log.Result["GLUSRCompanyName"] = company_name
        //       Write2S3(&S3log)

        flag := isCompanyNameContained(company_name, account_name)
        if flag == "yes" {
             percentmatch := 100.0
             return percentmatch, err
        }

        // Keywords to exclude
	keywords := []string{"ENTERPRISE", "ENTERPRISES", "TRADER", "TRADERS", "PVT. LTD.", "Private", "Limited"}

	// Remove keywords from company name and account name
	company_name = removeKeywords(strings.ToUpper(company_name), keywords)
	account_name = removeKeywords(strings.ToUpper(account_name), keywords)

        
        reg, _ := regexp.Compile("[^ A-Za-z]+")
        company_name = strings.ToUpper(reg.ReplaceAllString(company_name, ""))
        account_name = strings.ToUpper(reg.ReplaceAllString(account_name, ""))
        if strings.Contains(company_name, "PVT") || strings.Contains(company_name, "PRIV") {
                var delimiter string
                if strings.Contains(company_name, "PVT") {
                        delimiter = "PVT"
                } else {
                        delimiter = "PRIV"
                }
                //fmt.Println("PVT/PRIV found in company name", company_name)
                arr := strings.Split(company_name, delimiter)
                company_name = arr[0]
        }
        if strings.Contains(account_name, "PVT") || strings.Contains(account_name, "PRIV") {
                var delimiter string
                if strings.Contains(account_name, "PVT") {
                        delimiter = "PVT"
                } else {
                        delimiter = "PRIV"
                }
                arr := strings.Split(account_name, delimiter)
                account_name = arr[0]
        }

        company_name = strings.Replace(company_name, " AND", "", 1)
        account_name = strings.Replace(account_name, " AND", "", 1)
        space := regexp.MustCompile(`\s+`)
        company_name = space.ReplaceAllString(company_name, " ")
        account_name = space.ReplaceAllString(account_name, " ")
        m := len(company_name)
        n := len(account_name)
        var LCStuff [100][100]int
        common := 0
        for i := 0; i <= m; i++ {
                for j := 0; j <= n; j++ {
                        if i == 0 || j == 0 {
                                LCStuff[i][j] = 0
                        } else if company_name[i-1] == account_name[j-1] {
                                LCStuff[i][j] = LCStuff[i-1][j-1] + 1
                                // if LCStuff[i][j] > common {
                                        // common = LCStuff[i][j]
                                // }

                        } else {
                                if LCStuff[i-1][j] > LCStuff[i][j-1]{
                                        LCStuff[i][j]=LCStuff[i-1][j]
                                }else{
                                        LCStuff[i][j]=LCStuff[i][j-1] 
                                }
                        }
                }
        }
        common=LCStuff[m][n]
        avg_length := float64(m+n) / float64(2)
        if avg_length == 0 {
                avg_length = 1
        }
        percent := float64(common*100) / float64(avg_length)
        S3log.Result["SimilarityPercentage"] = percent
        //       Write2S3(&S3log)
        return percent, err
}

func SimilarityCEO(glid string, account_name string) (float64, error) {
        company_name := ""

        //mainConnection, err := db.GetDatabaseConnection("main")
        mainConnection, err := db.GetDatabaseConnection("approvalPG")
        //fmt.Println("mainConnection  similarityCEO is ",mainConnection,"error in similarityCEO",err)
        if err != nil {
                //fmt.Println(err.Error())
                //fmt.Println("error while connecting to DB")
                CreateTicket(Tanidd, "", "error while connecting to DB in similarityceo", Companyid)
                S3log.Result["error while connecting to DB in similarityceo"] = err.Error()
                Write2S3(&S3log)
                return 0.0,err
        }

        GLID := int64(StringToInteger(glid))
        // fmt.Println("ceo glid", GLID)
        //query := `SELECT COMPANYCEO FROM company WHERE FK_GLUSR_USR_ID = :GLID`
        query := `SELECT GLUSR_USR_CFIRSTNAME,GLUSR_USR_CLASTNAME FROM glusr_usr WHERE GLUSR_USR_ID = $1`
        params := make([]interface{}, 1)
        params[0] = GLID
        result, err := db.SelectQuerySql(mainConnection, query, params)
        //fmt.Println("SIMILARITY CEO Result:", result)
        if err != nil {
                //fmt.Println(err)
                //fmt.Println("error while executing query")
                        CreateTicket(Tanidd, "", "error while executing  similarityceo query", Companyid)
                S3log.Result["error while executing query"] = err.Error()
                Write2S3(&S3log)
                return 0.0,err

        }
        if result != nil {
                if queryData, dataExists := result["queryData"]; dataExists {
                        if queryData != nil {
                                resultSet := queryData.([]interface{})
                                if len(resultSet) > 0 {
                                        for _, payRecord := range resultSet {

                                                payRecordResult := payRecord.(map[string]interface{})
                                                //fmt.Println("payRecordResult",payRecordResult)
                                                name, nameExists := payRecordResult["glusr_usr_cfirstname"]
                                                //fmt.Println("count found from query", count)
                                                if nameExists && name != nil {
                                                        //fmt.Println("hi****************")
                                                        company_name += name.(string)
                                                        //fmt.Println( "cfirstname ",company_name)
                                                }
                                                name, nameExists = payRecordResult["glusr_usr_clastname"]
                                                //fmt.Println("count found from query", count)
                                                if nameExists && name != nil {
                                                        //fmt.Println("hi****************")
                                                        company_name += " "
                                                        company_name += name.(string)
                                                        //fmt.Println(company_name)
                                                }

                                        }
                                }
                        }
                }
        }
        //fmt.Println("ceo company name  glid", company_name)
        S3log.Result["GLUSRCeoName"] = company_name
        //       Write2S3(&S3log)
        flag := isCompanyNameContained(company_name, account_name)
        if flag == "yes" {
             percentmatch := 100.0
             return percentmatch, err
        }

        // Keywords to exclude
	keywords := []string{"ENTERPRISE", "ENTERPRISES", "TRADER", "TRADERS", "PVT. LTD.", "Private", "Limited"}

	// Remove keywords from company name and account name
	company_name = removeKeywords(strings.ToUpper(company_name), keywords)
	account_name = removeKeywords(strings.ToUpper(account_name), keywords)

        reg, _ := regexp.Compile("[^ A-Za-z]+")
        company_name = strings.ToUpper(reg.ReplaceAllString(company_name, ""))
        account_name = strings.ToUpper(reg.ReplaceAllString(account_name, ""))
        if strings.Contains(company_name, "PVT") || strings.Contains(company_name, "PRIV") {
                var delimiter string
                if strings.Contains(company_name, "PVT") {
                        delimiter = "PVT"
                } else {
                        delimiter = "PRIV"
                }
                //fmt.Println("PVT/PRIV found in company name", company_name)
                arr := strings.Split(company_name, delimiter)
                company_name = arr[0]
        }
        if strings.Contains(account_name, "PVT") || strings.Contains(account_name, "PRIV") {
                var delimiter string
                if strings.Contains(account_name, "PVT") {
                        delimiter = "PVT"
                } else {
                        delimiter = "PRIV"
                }
                arr := strings.Split(account_name, delimiter)
                account_name = arr[0]
        }
        arr1 := strings.Split(company_name, " ")
        arr2 := strings.Split(account_name, " ")
        company_name = arr1[0] + " " + arr1[len(arr1)-1]
        account_name = arr2[0] + " " + arr2[len(arr2)-1]
        // fmt.Println(company_name)
        // fmt.Println(account_name)
        m := len(company_name)
        n := len(account_name)
        var LCStuff [100][100]int
        common := 0
        for i := 0; i <= m; i++ {
                for j := 0; j <= n; j++ {
                        if i == 0 || j == 0 {
                                LCStuff[i][j] = 0
                        } else if company_name[i-1] == account_name[j-1] {
                                LCStuff[i][j] = LCStuff[i-1][j-1] + 1
                                // if LCStuff[i][j] > common {
                                        // common = LCStuff[i][j]
                                // }

                        } else {
                                if LCStuff[i-1][j] > LCStuff[i][j-1] {
                                        LCStuff[i][j]=LCStuff[i-1][j]
                                }else{
                                        LCStuff[i][j]=LCStuff[i][j-1] 
                                }
                                
                        }
                }
        }
        common=LCStuff[m][n]
        avg_length := float64(m+n) / float64(2)
        if avg_length == 0 {
                avg_length = 1
        }
        percent := float64(common*100) / float64(avg_length)
        S3log.Result["SimilarityCeoPercentage"] = percent
        //       Write2S3(&S3log)
        return percent, err
}

func SimilarityCONTACT(glid string, account_name string) (float64, error) {
        company_name := ""

        mainConnection, err := db.GetDatabaseConnection("approvalPG")
        //fmt.Println("mainConnection  similarityCONTACT is ",mainConnection,"error in similarityCONTACT",err)
        if err != nil {
                //fmt.Println(err.Error())
                //fmt.Println("error while connecting to DB")
               CreateTicket(Tanidd, "", "error while connecting to DB in similaritycontact", Companyid)
                S3log.Result["error while connecting to DB in similaritycontact"] = err.Error()
                Write2S3(&S3log)
                return 0.0,err
        }

        GLID := int64(StringToInteger(glid))
        //query := `SELECT CONTACTPERSON FROM company WHERE FK_GLUSR_USR_ID = :GLID`     GLUSR_USR_FIRSTNAME
        query := `SELECT GLUSR_USR_FIRSTNAME,GLUSR_USR_LASTNAME FROM glusr_usr WHERE GLUSR_USR_ID = $1`
        params := make([]interface{}, 1)
        params[0] = GLID
        result, err := db.SelectQuerySql(mainConnection, query, params)
        if err != nil {
                //fmt.Println(err)
                //fmt.Println("error while executing query")
                CreateTicket(Tanidd, "", "error while executing  similaritycontact query", Companyid)
                S3log.Result["error while executing  similaritycontact query"] = err.Error()
                Write2S3(&S3log)
                return 0.0,err
        }
        if result != nil {
                if queryData, dataExists := result["queryData"]; dataExists {
                        if queryData != nil {
                                resultSet := queryData.([]interface{})
                                if len(resultSet) == 1 {
                                        for _, payRecord := range resultSet {

                                                payRecordResult := payRecord.(map[string]interface{})
                                                name, nameExists := payRecordResult["glusr_usr_firstname"]
                                                //fmt.Println("count found from query", count)
                                                if nameExists && name != nil {
                                                        //fmt.Println("hi****************")
                                                        company_name += name.(string)
                                                        // fmt.Println("CONTACT PERSON FIRSTNAME INSIDE",company_name)
                                                }
                                                name, nameExists = payRecordResult["glusr_usr_lastname"]
                                                //fmt.Println("count found from query", count)
                                                if nameExists && name != nil {
                                                        //fmt.Println("hi****************")
                                                        company_name += " "
                                                        company_name += name.(string)
                                                        //fmt.Println("CONTACT PERSON LASTNAME INSIDE",company_name)
                                                }

                                        }
                                }
                        }
                }
        }
        //  fmt.Println("CONTACT PERSON OUTSIDE",company_name)
        S3log.Result["GLUSRContact"] = company_name
        //       Write2S3(&S3log)
        flag := isCompanyNameContained(company_name, account_name)
        if flag == "yes" {
             percentmatch := 100.0
             return percentmatch, err
        }

        // Keywords to exclude
	keywords := []string{"ENTERPRISE", "ENTERPRISES", "TRADER", "TRADERS", "PVT. LTD.", "Private", "Limited"}

	// Remove keywords from company name and account name
	company_name = removeKeywords(strings.ToUpper(company_name), keywords)
	account_name = removeKeywords(strings.ToUpper(account_name), keywords)
        
        reg, _ := regexp.Compile("[^ A-Za-z]+")
        company_name = strings.ToUpper(reg.ReplaceAllString(company_name, ""))
        account_name = strings.ToUpper(reg.ReplaceAllString(account_name, ""))
        if strings.Contains(company_name, "PVT") || strings.Contains(company_name, "PRIV") {
                var delimiter string
                if strings.Contains(company_name, "PVT") {
                        delimiter = "PVT"
                } else {
                        delimiter = "PRIV"
                }
                //fmt.Println("PVT/PRIV found in company name", company_name)
                arr := strings.Split(company_name, delimiter)
                company_name = arr[0]
        }
        if strings.Contains(account_name, "PVT") || strings.Contains(account_name, "PRIV") {
                var delimiter string
                if strings.Contains(account_name, "PVT") {
                        delimiter = "PVT"
                } else {
                        delimiter = "PRIV"
                }
                arr := strings.Split(account_name, delimiter)
                account_name = arr[0]
        }
        arr1 := strings.Split(company_name, " ")
        arr2 := strings.Split(account_name, " ")
        company_name = arr1[0] + " " + arr1[len(arr1)-1]
        account_name = arr2[0] + " " + arr2[len(arr2)-1]
        // fmt.Println(company_name)
        // fmt.Println(account_name)
        m := len(company_name)
        n := len(account_name)
        var LCStuff [100][100]int
        common := 0
        for i := 0; i <= m; i++ {
                for j := 0; j <= n; j++ {
                        if i == 0 || j == 0 {
                                LCStuff[i][j] = 0
                        } else if company_name[i-1] == account_name[j-1] {
                                LCStuff[i][j] = LCStuff[i-1][j-1] + 1
                                if LCStuff[i][j] > common {
                                        common = LCStuff[i][j]
                                }

                        } else {
                                LCStuff[i][j] = 0
                        }
                }
        }
        avg_length := float64(m+n) / float64(2)
        if avg_length == 0 {
                avg_length = 1
        }
        percent := float64(common*100) / float64(avg_length)
        S3log.Result["SimilarityContactPercentage"] = percent
        //       Write2S3(&S3log)
        return percent, err
}

func SimilarityPIN(glid string, Pin string) (float64, error) {
        var pincode int64

        //mainConnection, err := db.GetDatabaseConnection("main")
        mainConnection, err := db.GetDatabaseConnection("approvalPG")
        //fmt.Println("mainConnection  similarityPIN is ",mainConnection,"error in similarityPIN",err)
        if err != nil {
                //fmt.Println(err.Error())
                //fmt.Println("error while connecting to DB")
                        CreateTicket(Tanidd, "", "error while connecting to DB in similaritypin", Companyid)
                S3log.Result["error while connecting to DB in similaritypin"] = err.Error()
                Write2S3(&S3log)
                return 0.0,err
        }

        GLID := int64(StringToInteger(glid))
        space := regexp.MustCompile(`\s+`)
        Pin = space.ReplaceAllString(Pin, "")
        reg, _ := regexp.Compile("[^0-9]+")
        Pin = reg.ReplaceAllString(Pin, "")
        pin := int64(StringToInteger(Pin))
        //query := `SELECT PIN FROM company WHERE FK_GLUSR_USR_ID = :GLID`
        query := `SELECT GLUSR_USR_ZIP::text  FROM glusr_usr WHERE GLUSR_USR_ID = $1`
        params := make([]interface{}, 1)
        params[0] = GLID
        result, err := db.SelectQuerySql(mainConnection, query, params)
        if err != nil {
                //fmt.Println(err)
                //fmt.Println("error while executing query")
        CreateTicket(Tanidd, "", "error while executing query in similarity PIN", Companyid)
        S3log.Result["error while executing query in similarity PIN"] = err.Error()
                Write2S3(&S3log)
                return 0.0,err
        }
        pinstring := ""
        if result != nil {
                if queryData, dataExists := result["queryData"]; dataExists {
                        if queryData != nil {
                                resultSet := queryData.([]interface{})
                                if len(resultSet) == 1 {
                                        for _, payRecord := range resultSet {

                                                payRecordResult := payRecord.(map[string]interface{})
                                                name, nameExists := payRecordResult["glusr_usr_zip"]
                                                //fmt.Println("count found from query", count)
                                                if nameExists && name != nil {
                                                        //fmt.Println("hi****************")
                                                        pinstring = name.(string)
                                                }

                                        }
                                }
                        }
                }
        }
        S3log.Result["GLUSRPin"] = pinstring
        //       Write2S3(&S3log)
        space = regexp.MustCompile(`\s+`)
        pinstring = space.ReplaceAllString(pinstring, "")
        reg, _ = regexp.Compile("[^0-9]+")
        pinstring = reg.ReplaceAllString(pinstring, "")
        pincode = int64(StringToInteger(pinstring))
        percent := 0.0
        if pincode == pin {
                percent = 100
        }
        // fmt.Println(pincode, pin)
        S3log.Result["SimilarityPinPercentage"] = percent
        //       Write2S3(&S3log)
        return percent, err
}

func SimilarityAddress(glid string, address string) (float64, error) {
        company_address := ""

        //mainConnection, err := db.GetDatabaseConnection("main")
        mainConnection, err := db.GetDatabaseConnection("approvalPG")
        //fmt.Println("mainConnection  similarityAddress is ",mainConnection,"error in similarityAddress",err)
        if err != nil {
                //fmt.Println(err.Error())
                //fmt.Println("error while connecting to DB")
                CreateTicket(Tanidd, "", "error while connecting to DB in similarityaddress", Companyid)
                S3log.Result["error while connecting to DB in similarityaddress"] = err.Error()
                Write2S3(&S3log)
                return 0.0,err
        }

        GLID := int64(StringToInteger(glid))
        //query := `SELECT ADDRESS, CITY, STATE FROM company WHERE FK_GLUSR_USR_ID = $1`
        //GLUSR_USR_CITY   GLUSR_USR_STATE   GLUSR_USR_ADD1   GLUSR_USR_ADD2
        query := `SELECT GLUSR_USR_ADD1,GLUSR_USR_ADD2, GLUSR_USR_CITY, GLUSR_USR_STATE FROM glusr_usr WHERE GLUSR_USR_ID = $1`
        params := make([]interface{}, 1)
        params[0] = GLID
        result, err := db.SelectQuerySql(mainConnection, query, params)
        if err != nil {
                //fmt.Println(err)
                //fmt.Println("error while executing query")
                CreateTicket(Tanidd, "", "error while executing  similarityAddress query ", Companyid)
                S3log.Result["error while executing  similarityAddress query"] = err.Error()
                Write2S3(&S3log)
                return 0.0,err
        }
        if result != nil {
                if queryData, dataExists := result["queryData"]; dataExists {
                        if queryData != nil {
                                resultSet := queryData.([]interface{})
                                if len(resultSet) == 1 {
                                        for _, payRecord := range resultSet {

                                                payRecordResult := payRecord.(map[string]interface{})
                                                name, nameExists := payRecordResult["glusr_usr_add1"]
                                                //fmt.Println("count found from query", count)
                                                if nameExists && name != nil {
                                                        //fmt.Println("hi****************")
                                                        company_address += name.(string)
                                                }
                                                name, nameExists = payRecordResult["glusr_usr_add2"]
                                                //fmt.Println("count found from query", count)
                                                if nameExists && name != nil {
                                                        //fmt.Println("hi****************")
                                                        company_address += name.(string)
                                                }
                                                name, nameExists = payRecordResult["glusr_usr_city"]
                                                //fmt.Println("count found from query", count)
                                                if nameExists && name != nil {
                                                        //fmt.Println("hi****************")
                                                        company_address += name.(string)
                                                }
                                                name, nameExists = payRecordResult["glusr_usr_state"]
                                                //fmt.Println("count found from query", count)
                                                if nameExists && name != nil {
                                                        //fmt.Println("hi****************")
                                                        company_address += name.(string)
                                                }

                                        }
                                }
                        }
                }
        }
        S3log.Result["GLUSRCompanyAdress"] = company_address
        //     Write2S3(&S3log)
        reg, _ := regexp.Compile("[^ A-Za-z0-9]+")
        company_address = strings.ToUpper(reg.ReplaceAllString(company_address, ""))
        address = strings.ToUpper(reg.ReplaceAllString(address, ""))
        // fmt.Println(company_address)
        // fmt.Println(address)
        m := len(company_address)
        n := len(address)
        var LCStuff [10000][10000]int
        common := 0
        for i := 0; i <= m; i++ {
                for j := 0; j <= n; j++ {
                        if i == 0 || j == 0 {
                                LCStuff[i][j] = 0
                        } else if company_address[i-1] == address[j-1] {
                                LCStuff[i][j] = LCStuff[i-1][j-1] + 1
                                if LCStuff[i][j] > common {
                                        common = LCStuff[i][j]
                                }

                        } else {
                                LCStuff[i][j] = 0
                        }
                }
        }
        avg_length := float64(m+n) / float64(2)
        if avg_length == 0 {
                avg_length = 1
        }
        percent := float64(common*100) / float64(avg_length)
        S3log.Result["Similarityadresspercentage"] = percent
        //       Write2S3(&S3log)
        return percent, err
}

func SimilarityStatecd(glid string, statecode string) (bool, error) {
        statename := ""

        //mainConnection, err := db.GetDatabaseConnection("main")
        mainConnection, err := db.GetDatabaseConnection("approvalPG")
        //fmt.Println("mainConnection  similarityAddress is ",mainConnection,"error in similarityAddress",err)
        if err != nil {
                //fmt.Println(err.Error())
                //fmt.Println("error while connecting to DB")
                CreateTicket(Tanidd, "", "error while connecting to DB in similarityaddress", Companyid)
                S3log.Result["error while connecting to DB in similarityaddress"] = err.Error()
                Write2S3(&S3log)
                return false,err
        }

        GLID := int64(StringToInteger(glid))
        //query := `SELECT ADDRESS, CITY, STATE FROM company WHERE FK_GLUSR_USR_ID = $1`
        //GLUSR_USR_CITY   GLUSR_USR_STATE   GLUSR_USR_ADD1   GLUSR_USR_ADD2
        query := `SELECT GLUSR_USR_STATE FROM glusr_usr WHERE GLUSR_USR_ID = $1`
        params := make([]interface{}, 1)
        params[0] = GLID
        result, err := db.SelectQuerySql(mainConnection, query, params)
        if err != nil {
                //fmt.Println(err)
                //fmt.Println("error while executing query")
                CreateTicket(Tanidd, "", "error while executing  similarityAddress query ", Companyid)
                S3log.Result["error while executing  similarityAddress query"] = err.Error()
                Write2S3(&S3log)
                return false,err
        }
        if result != nil {
                if queryData, dataExists := result["queryData"]; dataExists {
                        if queryData != nil {
                                resultSet := queryData.([]interface{})
                                if len(resultSet) == 1 {
                                        for _, payRecord := range resultSet {

                                                payRecordResult := payRecord.(map[string]interface{})
                                                name, nameExists := payRecordResult["glusr_usr_state"]
                                                //fmt.Println("count found from query", count)
                                                if nameExists && name != nil {
                                                        //fmt.Println("hi****************")
                                                        statename += name.(string)
                                                }

                                        }
                                }
                        }
                }
        }
        S3log.Result["statename"] = statename
        if compareStates(statename, statecode) {
		return true, err
	} else {
		return false, err
	}
        
}

// Helper function to normalize state names for comparison
func normalizeStateName(stateName string) string {
	// Convert to lowercase and replace "&" with "and"
	normalized := strings.ToLower(stateName)
	normalized = strings.ReplaceAll(normalized, "&", "and")
	normalized = strings.ReplaceAll(normalized, "  ", " ") // Remove extra spaces
	return strings.TrimSpace(normalized)
}

// Function to compare two state names (one from query, one from map)
func compareStates(queryState string, stateCodeStr string) bool {
	stateCode, err := strconv.Atoi(stateCodeStr)
	if err != nil {
		fmt.Println("Invalid state code:", stateCodeStr)
		return false
	}

	// Get the state name from the map using the state code
	apiState, exists := stateCodeToStateName[stateCode]
	if !exists {
		fmt.Println("State code not found in the map")
		return false
	}

	// Normalize both state names
	normalizedQueryState := normalizeStateName(queryState)
	normalizedApiState := normalizeStateName(apiState)

	// Compare the normalized names
	return normalizedQueryState == normalizedApiState
}

func CreateNewTicket(companyidd string, tan TanData) {

        comm := `Tan number = '` + tan.TANNO + `', Tan Company name = '` + tan.COMPANY_NAME + `', Tan First name = '` + tan.FIRSTNAME + `', Tan Last name = '` + tan.LASTNAME + `', Tan Mid name = '` + tan.MIDNAME + `', Address = '` + tan.ADDRESS + `', STATECD = ` + tan.STATECD + `, PINCODE = ` + tan.PINCODE
        // comm := ""
        // S3log.TicketCreated = "Y"
        S3log.Result["Ticket issuance tan data "] = comm
        param := url.Values{
                "StsId":        {companyidd},
                "UserId":       {"32789"},
                "LoginName":    {"WebERP AutoProcess"},
                "Comments":     {"TAN Ticket issuance " + comm},
                "SourceName":   {"WebERP"},
                "SourceType":   {"1"},
                "FlagType":     {"issuetagging"},
                "TicketType":   {"General"},
                "LoginEmailId": {"weberp@indiamart.com"},
                "TypeId":       {"276"},
                "GroupId":      {"15"},
                "SelfTicket":   {"1"},
        }
        startt := utils.GetTimeInNanoSeconds()
        //S3log.ExtApihit = "Y"
        resp, err := http.PostForm("https://weberp.intermesh.net/Services/TicketIssueService.aspx", param)
        S3log.Result["response from createnewticket api "] = resp
        if err != nil {
                S3log.Result["response from createnewticket api error "] = err.Error()
                Write2S3(&S3log)
                return
                // fmt.Println("failure1", err)
        }
        endd := utils.GetTimeInNanoSeconds()
        S3log.ApiResponsetime["API_RESPONSE_TICKET "] = (endd - startt) / 1000000

       // fmt.Println(resp)
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
                S3log.Result["response from createnewticket api error "] = err.Error()
                Write2S3(&S3log)
                return
                //       fmt.Println(err)
        }
        //        tandata := string(body)
        //        fmt.Println(tandata)
        var result map[string]interface{}
        err = json.Unmarshal(body, &result)
        if err != nil {
                S3log.Result["response from createnewticket api error "] = err.Error()
                Write2S3(&S3log)
                return
                //       fmt.Println(err)
        }
	fmt.Println("create ticket response: ",result)
        	if result == nil {
		S3log.Result["map is empty "] = "Y"
	} else if result["300"].(string) != "" {
		S3log.Result["Ticketid "] = result["300"].(string)
                S3log.TicketCreated = "Y"
	} else {
		S3log.Result["tickid is empty "] = "Y"
	}
 //       S3log.Result["Ticketid "] = result["300"].(string)
       // fmt.Println("res new ticket",result)
        // Write2S3(&S3log)
}

func CreateTicket(tanid string, statuscode string, errormessage string, companyidd string) {

        //comm := `Tan number = '` + tan.TANNO + `', Tan Company name = '` + tan.COMPANY_NAME + `',Tan First name = '` + tan.FIRSTNAME + `',Tan Last name = '` + tan.LASTNAME + `',Tan Mid name = '` + tan.MIDNAME + `',Address = '` + tan.ADDRESS + `',STATECD = ` + tan.STATECD + `,PINCODE = ` + tan.PINCODE
        comm := `Tan number = '` + tanid + ` Api statuscode  = '` + statuscode + ` errormessage  = '` + errormessage
        // S3log.TicketCreated = "Y"
        S3log.Result["Ticket issuance tan data "] = comm
        param := url.Values{
                "StsId":        {companyidd},
                "UserId":       {"32789"},
                "LoginName":    {"WebERP AutoProcess"},
                "Comments":     {"TAN Ticket issuance " + comm},
                "SourceName":   {"WebERP"},
                "SourceType":   {"1"},
                "FlagType":     {"issuetagging"},
                "TicketType":   {"General"},
                "LoginEmailId": {"weberp@indiamart.com"},
                "TypeId":       {"276"},
                "GroupId":      {"15"},
                "SelfTicket":   {"1"},
        }
        startt := utils.GetTimeInNanoSeconds()
        //S3log.ExtApihit = "Y"
        resp, err := http.PostForm("https://weberp.intermesh.net/Services/TicketIssueService.aspx", param)

        if err != nil {
                S3log.Result["response from createnewticket api error "] = err.Error()
                Write2S3(&S3log)
                // fmt.Println("failure1", err)
                return
        }
        endd := utils.GetTimeInNanoSeconds()
        S3log.ApiResponsetime["API_RESPONSE_ERRORTICKET "] = (endd - startt) / 1000000
 //       fmt.Println(resp)
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
                S3log.Result["response from createnewticket api error "] = err.Error()
                Write2S3(&S3log)
                //       fmt.Println(err)
                return
        }
        //      tandata := string(body)
        //      fmt.Println(tandata)
        var result map[string]interface{}
        err = json.Unmarshal(body, &result)
        if err != nil {
                S3log.Result["response from createnewticket api error "] = err.Error()
                Write2S3(&S3log)
                //       fmt.Println(err)
                return
        }
	fmt.Println("create ticket response: ",result)
        	if result == nil {
		S3log.Result["map is empty "] = "Y"
	} else if result["300"].(string) != "" {
		S3log.Result["Ticketid "] = result["300"].(string)
                S3log.TicketCreated = "Y"
	} else {
		S3log.Result["ticketid is empty "] = "Y"
	}
       // S3log.Result["Ticketid "] = result["300"].(string)
 //       fmt.Println("res ticket",result)
        // Write2S3(&S3log)
}

