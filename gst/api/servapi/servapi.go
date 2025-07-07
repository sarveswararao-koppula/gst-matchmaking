package servapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mm/components/constants"
	"net/http"
	"time"
)

// RemoteHost ... (gst server) bi-utils ip
// const RemoteHost = "107.22.229.251"
const RemoteHost = "65.0.217.127"

// BIValidationKeyFromSOA ... validaiton key for modid:BI from SOA
const BIValidationKeyFromSOA = "af7f0273997b9b290bd7c57aa19f36c2"

// UserDetails ... check glid paid/free
func UserDetails(glid string) (string, error) {

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	token := "imobile@15061981"
	modid := "BI"
	debug := "1"
	url := "http://users.imutils.com/wservce/users/detail/?token=" + token + "&modid=" + modid + "&glusrid=" + glid + "&AK=" + constants.ServerAK + "&debug=" + debug

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	bodyString := string(bodyBytes)
	//fmt.Println(bodyString)
	return bodyString, nil
}

//merpakgeneration
func GetTAkFromMerpLogin(empid string) (string, error) {
	
	url := fmt.Sprintf("https://merp.intermesh.net/index.php/login/loginotpgeneration?usertype=999&display=-1&empid=%s", empid)

	
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error making GET request: %w", err)
	}
	defer resp.Body.Close()

	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("error parsing JSON: %w", err)
	}

	
	if msg, ok := result["Msg"].(string); ok && msg == "Success" {
		if t, ok := result["t"].(string); ok {
			return t, nil
		}
		return "", fmt.Errorf("key 't' not found or not a string")
	}
	return "", fmt.Errorf("message is not 'Success', got: %v", result["Msg"])
}


// UserVerifiedDetails ... check if gst already verified
func UserVerifiedDetails(glid string) (string, error) {

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	token := "imobile@15061981"
	modid := "BI"
	attrID := "2106"
	url := "http://users.imutils.com/wservce/users/verifiedDetail/?token=" + token + "&modid=" + modid + "&glusrid=" + glid + "&attribute_id=" + attrID + "&AK=" + constants.ServerAK

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	bodyString := string(bodyBytes)
	//fmt.Println(bodyString)
	return bodyString, nil
}

// UserVerifiedDetails calls the API and returns the raw response as a map
func UserVerifiedDetailsMM(glid string, matchedattrid string) (map[string]interface{}, error) {
	client := &http.Client{Timeout: 2 * time.Second}

	token := "imobile@15061981"
	modid := "BI"
	attrID := matchedattrid 
	url := fmt.Sprintf(
			"http://users.imutils.com/wservce/users/verifiedDetail/?token=%s&modid=%s&glusrid=%s&attribute_id=%s&AK=%s",
			token, modid, glid, attrID, constants.ServerAK,
	)

	resp, err := client.Get(url)
	if err != nil {
			return nil, fmt.Errorf("failed to fetch user details: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return result, nil
}

// ExtractUserVerificationDate dynamically finds User_verification_date
func ExtractUserVerificationDate(response map[string]interface{}) (string, string) {
	// Navigate to "Response" -> "Data"
	respData, ok := response["Response"].(map[string]interface{})
	if !ok {
			return "NULL", "Not Verified"
	}

	data, ok := respData["Data"].(map[string]interface{})
	if !ok {
			return "NULL", "Not Verified"
	}

	// Loop over dynamic keys inside "Data"
	for _, details := range data {
			if detailsMap, ok := details.(map[string]interface{}); ok {
					if date, found := detailsMap["User_verification_date"].(string); found && date != "" {
							return date, "Verified"
					}
			}
	}

	return "NULL", "Not Verified"
}

// CompanyVerifiedDetails ... check if CompanyName already verified
func CompanyVerifiedDetails(glid string) (string, error) {

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	token := "imobile@15061981"
	modid := "BI"
	attrID := "111"
	url := "http://users.imutils.com/wservce/users/verifiedDetail/?token=" + token + "&modid=" + modid + "&glusrid=" + glid + "&attribute_id=" + attrID + "&AK=" + constants.ServerAK

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	bodyString := string(bodyBytes)
	//fmt.Println(bodyString)
	return bodyString, nil
}

// Details ... insert into glusr_comp
func Details(env, glid, gst, screenName string) (string, error) {

	client := &http.Client{
		Timeout: 4 * time.Second,
	}
	m := make(map[string]string)

	m["glusridval"] = glid
	m["GST"] = gst
	m["updatedby"] = "GST Tech Process"
	m["updatedbyId"] = "85344"
	m["updatedbyScreen"] = screenName
	m["userIp"] = RemoteHost
	m["userIpCoun"] = "INDIA"
	m["VALIDATION_KEY"] = BIValidationKeyFromSOA
	m["type"] = "CompRgst"
	m["histComment"] = "By " + screenName + " Process"
	m["AK"] = constants.ServerAK

	url := ""
	if env == "DEV" {
		url = "http://dev-service.intermesh.net/details"
	} else if env == "PROD" {
		url = "http://service.intermesh.net/details"
	}

	reqBody, _ := json.Marshal(m)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))

	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	bodyString := string(bodyBytes)

	return bodyString, nil
}

//UserVerification ...to verify gst for a glid with proper disposition
// func UserVerification(env, glid, attrID, attrVal, dispo string) (string, error) {

//      client := &http.Client{
//              Timeout: 2 * time.Second,
//      }
//      m := make(map[string]string)

//      m["VALIDATION_KEY"] = BIValidationKeyFromSOA
//      m["action_flag"] = "SP_VERIFY_ATTRIBUTE"
//      m["GLUSR_USR_ID"] = glid
//      m["ATTRIBUTE_ID"] = attrID
//      m["ATTRIBUTE_VALUE"] = attrVal
//      m["VERIFIED_BY_ID"] = "-1"
//      m["VERIFIED_BY_NAME"] = "Auto Approval GST Process"
//      m["VERIFIED_BY_AGENCY"] = "online"
//      m["VERIFIED_BY_SCREEN"] = "GST Verification Process"
//      m["VERIFIED_URL"] = ""
//      m["VERIFIED_IP"] = RemoteHost
//      m["VERIFIED_IP_COUNTRY"] = "INDIA"
//      m["VERIFIED_COMMENTS"] = dispo
//      m["VERIFIED_AUTHCODE"] = ""

//      url := ""
//      if env == "DEV" {
//              url = "http://dev-service.intermesh.net/user/verification"
//      } else if env == "PROD" {
//              url = "http://service.intermesh.net/user/verification"
//      }

//      reqBody, _ := json.Marshal(m)

//      req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))

//      if err != nil {
//              return "", err
//      }

//      resp, err := client.Do(req)

//      if err != nil {
//              return "", err
//      }

//      defer resp.Body.Close()

//      bodyBytes, err := ioutil.ReadAll(resp.Body)
//      if err != nil {
//              return "", err
//      }

//      bodyString := string(bodyBytes)

//      return bodyString, nil

// }

// UserVerification ...to verify gst for a glid with proper disposition ( multi)
func UserVerification(env, glid, attrID, attrVal, dispo, action_flag string) (string, error) {

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	m := make(map[string]string)

	m["VALIDATION_KEY"] = BIValidationKeyFromSOA
	if action_flag == "0" {
		m["action_flag"] = "SP_VERIFY_ATTRIBUTE"
	} else if action_flag == "1" {
		m["action_flag"] = "SP_BULK_VERIFY_ATTRIBUTE"
	}
	m["GLUSR_USR_ID"] = glid
	m["ATTRIBUTE_ID"] = attrID
	m["ATTRIBUTE_VALUE"] = attrVal
	m["VERIFIED_BY_ID"] = "-1"
	m["VERIFIED_BY_NAME"] = "Auto Approval GST Process"
	m["VERIFIED_BY_AGENCY"] = "online"
	m["VERIFIED_BY_SCREEN"] = "GST Verification Process"
	m["VERIFIED_URL"] = ""
	m["VERIFIED_IP"] = RemoteHost
	m["VERIFIED_IP_COUNTRY"] = "INDIA"
	m["VERIFIED_COMMENTS"] = dispo
	m["VERIFIED_AUTHCODE"] = ""

	url := ""
	if env == "DEV" {
		url = "http://dev-service.intermesh.net/user/verification"
	} else if env == "PROD" {
		url = "http://service.intermesh.net/user/verification"
	}

	reqBody, _ := json.Marshal(m)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))

	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	bodyString := string(bodyBytes)

	return bodyString, nil

}

// New User Update Service
func UserUpdate(env string, agg map[string]string) (string, error) {

	client := &http.Client{
		Timeout: 6 * time.Second,
	}

	url := ""
	if env == "DEV" {
		url = "http://stg-service.intermesh.net/user/update"
	} else if env == "PROD" {
		url = "http://service.intermesh.net/user/update"
	}

	reqBody, _ := json.Marshal(agg)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))

	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	bodyString := string(bodyBytes)

	return bodyString, nil
}

// API to fetch city and city_id from pincode
func CityFetch(pincode string) (string, error) {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	token := "immenu@7851"
	modid := "BI"
	locality := "n"
	url := "http://users.imutils.com/wservce/im/localityPincode/?token=" + token + "&modid=" + modid + "&pincode=" + pincode + "&locality=" + locality
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	bodyString := string(bodyBytes)
	return bodyString, nil
}

// HSN API
func Hsnapi(env, gst, hsnstring string) (string, error) {

	client := &http.Client{
		Timeout: 6 * time.Second,
	}
	m := make(map[string]string)
	fmt.Println(hsnstring)
	m["GST"] = gst
	m["ADDED_BY"] = "86881"
	m["VALIDATION_KEY"] = BIValidationKeyFromSOA
	m["INS_HSN"] = hsnstring
	m["DEL_HSN"] = ""
	m["AK"] = constants.ServerAK

	url := ""
	if env == "DEV" {
		// url = "http://stg-service.intermesh.net/gsthsnmapping"
		url = "http://stg-service.intermesh.net/gsthsnmapping"
	} else if env == "PROD" {
		url = "http://service.intermesh.net/gsthsnmapping"
	}

	reqBody, _ := json.Marshal(m)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))

	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	bodyString := string(bodyBytes)

	return bodyString, nil
}

// HSNReadDetails ...
func HsnReadDetails(gst string) (string, error) {

	client := &http.Client{
		Timeout: 4 * time.Second,
	}

	token := "imobile@15061981"
	modid := "BI"
	url := "http://users.imutils.com/wservce/users/gethsnfromgst/?token=" + token + "&modid=" + modid + "&gst=" + gst + "&AK=" + constants.ServerAK

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return "", err
		// fmt.Println("error: ",err)
	}

	resp, err := client.Do(req)

	if err != nil {
		return "", err
		// fmt.Println("error: ",err)
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
		// fmt.Println("error: ",err)
	}

	bodyString := string(bodyBytes)

	return bodyString, nil
	// fmt.Println("bodyString: ",bodyString)
}

// UserVerification ...to verify gst for a glid with proper disposition ( multi)
func UnVerification(env, glid, attrID, attrVal, dispo, action_flag string) (string, error) {

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	m := make(map[string]string)

	m["VALIDATION_KEY"] = BIValidationKeyFromSOA
	if action_flag == "0" {
		m["action_flag"] = "UNVERIFIED"
	} else if action_flag == "1" {
		m["action_flag"] = "UNVERIFIED"
	}
	m["GLUSR_USR_ID"] = glid
	m["ATTRIBUTE_ID"] = attrID
	m["ATTRIBUTE_VALUE"] = attrVal
	m["VERIFIED_BY_ID"] = "-1"
	m["VERIFIED_BY_NAME"] = "Auto Approval GST Process"
	m["VERIFIED_BY_AGENCY"] = "online"
	m["VERIFIED_BY_SCREEN"] = "GST Verification Process"
	m["VERIFIED_URL"] = ""
	m["VERIFIED_IP"] = RemoteHost
	m["VERIFIED_IP_COUNTRY"] = "INDIA"
	m["VERIFIED_COMMENTS"] = dispo
	m["VERIFIED_AUTHCODE"] = ""

	url := ""
	if env == "DEV" {
		url = "http://dev-service.intermesh.net/user/verification"
	} else if env == "PROD" {
		url = "http://service.intermesh.net/user/verification"
	}

	reqBody, _ := json.Marshal(m)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))

	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	bodyString := string(bodyBytes)

	return bodyString, nil

}
