package gstmmcontrols

import (
	"encoding/json"
	"errors"
	"fmt"
	servapi "mm/api/servapi"
	api "mm/api/thirdpartyapi"
	model "mm/model/masterindiamodel"
	"mm/properties"
	authadvance "mm/services/authbridgeadvanced"
	"mm/utils"
	"os"
	"regexp"
	"runtime/debug"
	"strings"
	"time"
)

// SubcriberHandler ...
func SubcriberHandler(data string) {
	//fmt.Println("SubcriberHandler started")
	logg := Logg{}
	err := json.Unmarshal([]byte(data), &logg)
	logg.AnyError = make(map[string]string)
	logg.UpdateFlags = make(map[string]bool)
	logg.VerifyParams = make(map[string]string)
	if err != nil {
		logg.AnyError["SubcriberHandler Unmarshal"] = err.Error()
		WriteQLog2(logg)
		return
	}
	postWork(logg)
}

func postWork(logg Logg) {
	//fmt.Println("P1")
	const BIValidationKeyFromSOA = "af7f0273997b9b290bd7c57aa19f36c2"
	// screenName := "Auto Approval GST Process"
	const RemoteHost = "65.0.217.127"
	defer func() {
		if panicCheck := recover(); panicCheck != nil {
			stack := string(debug.Stack())
			logg.StackTrace = stack
			WriteQLog2(logg)
			return
		}
	}()

	bucketType := logg.Response.Body.BucketType

	glid := logg.Request.Glid
	gst := logg.Response.Body.Gstin
	bucketName := logg.Response.Body.BucketName
	fmt.Println("BucketName :", bucketName)
	gstStatus := logg.Response.Body.GstStatus

	gstInsertionDate, err := time.Parse("02-01-2006", logg.Response.Body.GstInsertionDate)
	days := 9999.0
	if err == nil {
		nowStr := time.Now().Format("02-01-2006")
		nowDate, _ := time.Parse("02-01-2006", nowStr)
		days = nowDate.Sub(gstInsertionDate).Hours() / 24
	}

	//last insertion date is not within 30 days then fetch latest gst details  / hit masterindia api
	if days > 30.0 {

		gstStatus = ""
		logg.MasterIndia.User = "vishnu"

		//credentials needed
		cred := utils.GetCred(logg.MasterIndia.User)

		//fetching auth tokken needed to hit masterindia api
		tok, err := api.GetTokken(cred)

		if err != nil {
			logg.AnyError["GetTokken"] = err.Error()
			WriteQLog2(logg)
			LogToWorkerFile(logg)
			return
		}

		accessToken, ok := tok["access_token"].(string)

		if !ok || accessToken == "" {
			logg.AnyError["GetTokken"] = fmt.Sprint(tok)
			WriteQLog2(logg)
			LogToWorkerFile(logg)
			return
		}

		logg.MasterIndia.Hit = true
		//hit masterindia api with client ID and auth tokken
		apiData, err := api.GetMasterData(gst, map[string]string{
			"client_id":    cred["client_id"],
			"access_token": accessToken,
		})

		if err != nil {
			logg.AnyError["GetMasterData"] = err.Error()
			WriteQLog2(logg)
			LogToWorkerFile(logg)
			return
		}

		apiErr, _ := apiData["error"].(bool)
		apiErrStr, _ := apiData["error"].(string)
		apiDataData, ok := apiData["data"].(map[string]interface{})

		//Either error found or data not found
		if apiErr || apiErrStr != "" || !ok {
			logg.AnyError["GetMasterData"] = fmt.Sprint(apiData)
			WriteQLog2(logg)
			LogToWorkerFile(logg)
			return
		}

		gstStatus, _ = apiDataData["sts"].(string)

		var params []interface{}
		//converting api respose data to structured data in order to store in database
		_, params = utils.BusLogicOnMasterData_V2(gst, apiDataData)

		//updating gst latest details
		_, err = model.UpdateGSTMasterData(Database, params)
		if err != nil {
			logg.AnyError["updating gst_master_data tab"] = err.Error()
			WriteQLog2(logg)
			LogToWorkerFile(logg)
			return
		}

		// data_glid, err := authadvance.GetGlidFromGstM(Database, gst)
		//     if err != nil {
		// 	    logg.AnyError["getting error from getglidfromgst function"] = err.Error()
		//     } else {
		// 	   for _, glid := range data_glid {
		// 		if err := authadvance.ProcessSingleGLIDPubapilogging(glid, "/gstmm/v1/gst"); err != nil {
		// 			logg.AnyError["meshupdationerror_glid"] = err.Error()
		// 		}
		//         }
		//    }

		if err := authadvance.ProcessingGST(gst, "/gstmm/v1/gst"); err != nil {
			// logg.AnyError["meshupdationerror"] = err.Error()
			// fmt.Println("meshupdationerror_gst",err.Error())
			logg.AnyError["meshupdationerror_gst"] = err.Error()
		}
	}

	//gts status is not Active , exit
	if gstStatus != "Active" {
		logg.AnyError["gst_status not Active"] = gstStatus
		WriteQLog2(logg)
		LogToWorkerFile(logg)
		return
	}

	//fmt.Println("P3")
	isFree, compName, address, add2, state, pincode, cFirstName, cLastName, first_name, last_name, listing_status, err := IsGlidFree(glid)
	fmt.Println(add2)
	fmt.Println(isFree, "isFree", compName, address, state, pincode, cFirstName, cLastName, first_name, last_name, listing_status, err, "Testing IsGlidFree")
	if err != nil {
		logg.AnyError["IsGlidFree"] = err.Error()
		WriteQLog2(logg)
		LogToWorkerFile(logg)
		return
	}
	//fmt.Println("P5",isFree)
	if !isFree {
		logg.AnyError["IsGlidFree"] = "glid not free"
		WriteQLog2(logg)
		LogToWorkerFile(logg)
		return
	}

	isVerified, err := IsGSTAlreadyVerified(glid)
	if err != nil {
		logg.AnyError["IsGSTAlreadyVerified"] = err.Error()
		WriteQLog2(logg)
		LogToWorkerFile(logg)
		return
	}

	if isVerified {
		logg.AnyError["IsGSTAlreadyVerified"] = "gst already verified"
		WriteQLog2(logg)
		LogToWorkerFile(logg)
		return
	}
	//gst AUTO approval Bucket
	if bucketType == "AUTO" {
		// screenName must be "GST Auto Approval" ,it prevents insertion in iil approval pending table (GLADMIN screen)
		err := InsertCompTab(properties.Prop.SERVICES_ENV, glid, gst, "GST Auto Approval")
		if err != nil {
			logg.AnyError["InsertCompTab"] = err.Error()
			WriteQLog2(logg)
			LogToWorkerFile(logg)
			return
		}
		logg.ApprovalDone = "AUTO"
		WriteQLog2(logg)
		LogToWorkerFile(logg)
		return
	}

	//gst manual approval bucket
	if bucketType == "MAN" {

		//glid,rid,gst,host,env
		err := servapi.Publish(glid, logg.Request.UniqueID, gst, properties.Prop.SERVICES_ENV)
		if err != nil {
			logg.AnyError["manual Publish"] = err.Error()
			WriteQLog2(logg)
			LogToWorkerFile(logg)
			return
		}

		logg.ApprovalDone = "MAN"
		WriteQLog2(logg)
		LogToWorkerFile(logg)
		return
	}

}

// calling cityFetch API
func CallcityFetch(pincode string) (map[string]interface{}, error) {
	jsonStr1, err := servapi.CityFetch(pincode)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	if err := json.Unmarshal([]byte(jsonStr1), &data); err != nil {
		return nil, err
	}

	// Convert all keys in the top-level data map to lowercase
	lowercasedData := make(map[string]interface{})
	for k, v := range data {
		lowercasedData[strings.ToLower(k)] = v
	}

	// Extract and check the CODE and MESSAGE fields
	code, codeOk := lowercasedData["code"].(string)
	message, messageOk := lowercasedData["message"].(string)

	// If CODE is not 200 or MESSAGE is not "Success", return an error
	if !codeOk || code != "200" || !messageOk || message != "Success" {
		return map[string]interface{}{
			"city_id":       "",
			"city_name":     "",
			"state_id":      "",
			"state_name":    "",
			"district_id":   "",
			"district_name": "",
		}, fmt.Errorf("API error: %v", message)
	}

	// Prepare the map to store the result with lowercase keys and default values
	result := map[string]interface{}{
		"city_id":       "",
		"city_name":     "",
		"state_id":      "",
		"state_name":    "",
		"district_id":   "",
		"district_name": "",
	}

	// Handle nested fields with potential case variations in keys
	if dataField, ok := lowercasedData["data"].(map[string]interface{}); ok {
		for _, key := range []string{"city", "state", "district"} {
			// Convert key to lowercase for consistency
			for nestedKey, nestedValue := range dataField {
				if strings.ToLower(nestedKey) == key {
					if subField, ok := nestedValue.(map[string]interface{}); ok {
						for k, v := range subField {
							result[strings.ToLower(k)] = v
						}
					}
				}
			}
		}
	}

	return result, nil
}

// cleaning of gstAddress
func cleanAddress(addr string, state string, pincode string) string {
	var index, index2 int
	//strings.TrimSpace(addr)
	addr = strings.Trim(addr, " ")
	if strings.Contains(addr, state) {
		addr = strings.Replace(addr, state, "", -1)
		//addr = strings.Replace(addr, " ", "", -1)
	}
	if strings.Contains(addr, pincode) {
		addr = strings.Replace(addr, pincode, "", -1)
		//addr = strings.Replace(addr, " ", "", -1)
	}
	for i := 0; i < len(addr); i++ {
		if (int(addr[i]) >= 'a' && int(addr[i]) <= 'z') || (int(addr[i]) >= 'A' && int(addr[i]) <= 'Z') || (int(addr[i]) >= '0' && int(addr[i]) <= '9') {
			index2 = i
			break
		}
	}
	if index2 > 0 {
		addr = strings.TrimLeft(addr, ",-")
	}
	n := len(addr)
	for i := n - 1; i >= 0; i-- {
		if (int(addr[i]) >= 'a' && int(addr[i]) <= 'z') || (int(addr[i]) >= 'A' && int(addr[i]) <= 'Z') || (int(addr[i]) >= '0' && int(addr[i]) <= '9') {
			//fmt.Println(string(addr[i]), addr[i])
			index = i
			break
		}
	}
	if index > 0 {
		addr = strings.TrimRight(addr, ",-")
	}
	comma := regexp.MustCompile(`\,+`)
	//dash := regexp.MustCompile(`\-+`)
	s := comma.ReplaceAllString(addr, ",")
	//s= dash.ReplaceAllString(s,"-")
	s = strings.TrimSuffix(s, ",")
	s = strings.ReplaceAll(s, ".", " ")
	s = strings.ReplaceAll(s, "|", " ")
	nonalphanumeric := regexp.MustCompile(`[^a-zA-Z0-9],+`)
	s = nonalphanumeric.ReplaceAllString(s, "")
	s = strings.Replace(s, ", ", ",", -1)
	s = strings.Replace(s, ", ", ",", -1)
	s = strings.TrimSuffix(s, ",")
	//s = strings.TrimPrefix(s,"-")
	s = strings.TrimSpace(s)
	return s
}

// Splitting the name into two parts
func convert(name string) (string, string) {
	var c, finalString string
	d := ""
	b := strings.Fields(name)
	n := len(strings.Fields(name))
	if n == 1 {
		c = b[0]
		if len(c) >= 30 {
			c = c[:30]
		}
	} else {
		c = b[0]
		for i := 1; i < n; i++ {
			d = d + " " + b[i]
			finalString = strings.TrimSpace(d)
		}
		if len(finalString) >= 30 {
			finalString = finalString[:30]
		}
		if len(c) >= 30 {
			c = c[:30]
		}
	}

	// Convert to title case
	c = strings.Title(strings.ToLower(c))
	finalString = strings.Title(strings.ToLower(finalString))
	return c, finalString
}

// ModifyCompName ... title case company name
func ModifyCompName(compName string) string {
	compName = strings.ToLower(compName)
	a := strings.Fields(compName)
	for k, j := range a {
		if j == "llp" || j == "opc" || j == "m/s" {
			a[k] = strings.ToUpper(a[k])
		}
	}
	compName = strings.Join(a, " ")
	compName = strings.Title(compName)
	return compName
}

// VerifyGlidAttr ... verifying glid attr using /user/verification service with proper hist comments/disposition
func VerifyGlidAttr(env, glid, attrID, attrVal, dispo string) error {

	jsonStr, err := servapi.UserVerification(env, glid, attrID, attrVal, dispo, "0")
	// fmt.Println(jsonStr,"Inside Single Attr")
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return err
	}

	res := make(map[string]interface{})
	for k, v := range data {
		res[strings.ToLower(k)] = v
	}

	status, _ := res["status"].(string)

	if strings.ToLower(status) == "successful" {
		return nil
	}

	return errors.New(jsonStr)
}

func UnVerifyGlidAttr(env, glid, attrID, attrVal, dispo string) error {

	jsonStr, err := servapi.UnVerification(env, glid, attrID, attrVal, dispo, "0")
	// fmt.Println(jsonStr,"Inside Single Attr")
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return err
	}

	res := make(map[string]interface{})
	for k, v := range data {
		res[strings.ToLower(k)] = v
	}

	status, _ := res["status"].(string)

	if strings.ToLower(status) == "successful" {
		return nil
	}

	return errors.New(jsonStr)
}

func VerifyGlidAllAttr(env, glid, attrID, attrVal, dispo string) error {

	jsonStr, err := servapi.UserVerification(env, glid, attrID, attrVal, dispo, "1")
	// fmt.Println(jsonStr,"Inside all")
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return err
	}

	res := make(map[string]interface{})
	for k, v := range data {
		res[strings.ToLower(k)] = v
	}

	status, _ := res["status"].(string)
	// fmt.Println(status,"Inside all attributes")
	if strings.ToLower(status) == "successful" {
		return nil
	}

	return errors.New(jsonStr)
}

// UpdateComp ... update glid company name using /user/update service
func UpdateComp(env string, m map[string]string) error {

	jsonStr, err := servapi.UserUpdate(env, m)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return err
	}

	res := make(map[string]interface{})
	for k, v := range data {
		res[strings.ToLower(k)] = v
	}

	status, _ := res["status"].(string)

	if strings.ToLower(status) == "successful" {
		return nil
	}

	return errors.New(jsonStr)
}

// IsGlidFree ... checking glid is paid or free and also getting company name
func IsGlidFree(glid string) (bool, string, string, string, string, string, string, string, string, string, string, error) {

	compName := ""
	address := ""
	state := ""
	pincode := ""
	cfirstname := ""
	clastname := ""
	first_name := ""
	last_name := ""
	listing_status := ""
	add2 := ""

	jsonStr, err := servapi.UserDetails(glid)
	if err != nil {
		return false, compName, address, add2, state, pincode, cfirstname, clastname, first_name, last_name, listing_status, err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return false, compName, address, add2, state, pincode, cfirstname, clastname, first_name, last_name, listing_status, err
	}

	res := make(map[string]interface{})
	for k, v := range data {
		res[strings.ToLower(k)] = v
	}

	isPaid, _ := res["is_paid"].(string)
	compName, _ = res["company_name"].(string)
	address, _ = res["add1"].(string)
	add2, _ = res["add2"].(string)
	state, _ = res["state"].(string)
	pincode, _ = res["zip"].(string)
	cfirstname, _ = res["ceo_fname"].(string)
	clastname, _ = res["ceo_lname"].(string)
	first_name, _ = res["first_name"].(string)
	last_name, _ = res["last_name"].(string)
	listing_status, _ = res["glusr_usr_listing_status"].(string)
	if strings.ToLower(isPaid) == "free" {
		return true, compName, address, add2, state, pincode, cfirstname, clastname, first_name, last_name, listing_status, nil
	}

	if strings.ToLower(isPaid) == "paid" {
		return false, compName, address, add2, state, pincode, cfirstname, clastname, first_name, last_name, listing_status, nil
	}

	return false, compName, address, add2, state, pincode, cfirstname, clastname, first_name, last_name, listing_status, errors.New(jsonStr)
}

// IsGSTAlreadyVerified ...checking if gst is already verified for glid
func IsGSTAlreadyVerified(glid string) (bool, error) {
	jsonStr, err := servapi.UserVerifiedDetails(glid)

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
	k, _ = k["2106"].(map[string]interface{})
	status, _ := k["Status"].(string)

	if strings.ToLower(status) == "verified" {
		return true, nil
	}

	if strings.ToLower(status) == "not verified" {
		return false, nil
	}

	return false, errors.New(jsonStr)
}

// IsGSTAlreadyVerified ...checking if gst is already verified for glid
func IsCompanyAlreadyVerified(glid string) (bool, error) {
	jsonStr, err := servapi.CompanyVerifiedDetails(glid)

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
	k, _ = k["111"].(map[string]interface{})
	status, _ := k["Status"].(string)

	if strings.ToLower(status) == "verified" {
		return true, nil
	}

	if strings.ToLower(status) == "not verified" {
		return false, nil
	}

	return false, errors.New(jsonStr)
}

// InsertCompTab ...inserting gst in comp registration table
func InsertCompTab(env, glid, gst, screenName string) error {

	jsonStr, err := servapi.Details(env, glid, gst, screenName)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return err
	}

	res := make(map[string]interface{})
	for k, v := range data {
		res[strings.ToLower(k)] = v
	}

	status, _ := res["status"].(string)

	if strings.ToLower(status) == "successful" {
		return nil
	}

	return errors.New(jsonStr)
}

// WriteQLog2 ... after processing data from queue .. writing logs
func WriteQLog2(logg Logg) {

	//2009 November 10
	year, month, day := time.Now().Date()

	logsDir := properties.Prop.LOG_MATCH_MAKING + "/" + fmt.Sprint(year) + "/" + fmt.Sprint(int(month)) + "/" + fmt.Sprint(day)

	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		e := os.MkdirAll(logsDir, os.ModePerm)
		if e != nil {
			fmt.Println(e)
			return
		}
	}

	logsDir += "/match_making_queue.json"

	jsonLog, _ := json.Marshal(logg)
	jsonLogString := string(jsonLog)

	f, err := os.OpenFile(logsDir, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	f.WriteString("\n" + jsonLogString)

	fmt.Println(logsDir)
	return
}
