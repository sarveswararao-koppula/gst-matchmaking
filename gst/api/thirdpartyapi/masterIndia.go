package thirdpartyapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	//"log"
)

func GetPanData(pan string, m map[string]string) ([]map[string]interface{}, error) {

	res := make([]map[string]interface{}, 0)
	client_id := m["client_id"]
	access_token := m["access_token"]

	url := "https://commonapi.mastersindia.co/commonapis/searchpan?pan=" + pan

	client := http.Client{
		Timeout: time.Duration(20 * time.Second),
	}

	rqst, err := http.NewRequest("GET", url, nil)
	rqst.Header.Set("Content-Type", "application/json")
	rqst.Header.Set("Authorization", "Bearer "+access_token)
	rqst.Header.Set("client_id", client_id)

	if err != nil {
		return res, err
	}

	resp, err := client.Do(rqst)

	if err != nil {
		return res, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	bodyString := string(body)

	//fmt.Println(bodyString)

	var data map[string]interface{}
	err = json.Unmarshal([]byte(bodyString), &data)

	if err != nil {
		return res, err
	}

	api_err, _ := data["error"].(bool)
	str, ok := data["data"].(string)

	if api_err || !ok {

		api_err_str, _ := data["error"].(string)
		if api_err_str == "invalid_grant" {
			return append(res, data), errors.New("invalid_grant")
		}
		return append(res, data), errors.New("api_data_error")
	}

	str_decoded, err := base64.StdEncoding.DecodeString(str)

	if err != nil {
		return append(res, data), err
	}

	err = json.Unmarshal([]byte(str_decoded), &res)

	if err != nil {
		return append(res, data), err
	}

	return res, nil
}

func GetChallanData(gst string, year string, m map[string]string) (map[string]interface{}, error) {

	res := make(map[string]interface{})
	client_id := m["client_id"]
	access_token := m["access_token"]

	url := "https://commonapi.mastersindia.co/commonapis/trackReturns?gstin=" + gst + "&fy=" + year

	client := http.Client{
		Timeout: time.Duration(6 * time.Second),
	}

	rqst, err := http.NewRequest("GET", url, nil)
	rqst.Header.Set("Content-Type", "application/json")
	rqst.Header.Set("Authorization", "Bearer "+access_token)
	rqst.Header.Set("client_id", client_id)

	if err != nil {
		return res, err
	}

	resp, err := client.Do(rqst)

	if err != nil {
		return res, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	bodyString := string(body)

	//fmt.Println(bodyString)

	var data map[string]interface{}
	err = json.Unmarshal([]byte(bodyString), &data)

	if err != nil {
		return res, err
	}

	res = data

	return res, nil
}

func GetMasterData(gst string, m map[string]string) (map[string]interface{}, error) {

	res := make(map[string]interface{})
	client_id := m["client_id"]
	access_token := m["access_token"]

	url := "https://commonapi.mastersindia.co/commonapis/searchgstin?gstin=" + gst

	client := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}

	rqst, err := http.NewRequest("GET", url, nil)
	rqst.Header.Set("Content-Type", "application/json")
	rqst.Header.Set("Authorization", "Bearer "+access_token)
	rqst.Header.Set("client_id", client_id)

	if err != nil {
		return res, err
	}

	resp, err := client.Do(rqst)
	fmt.Println("Response from MasterInida API for ", gst, " : ", resp)

	if err != nil {
		return res, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	bodyString := string(body)

	fmt.Println(bodyString)

	var data map[string]interface{}
	err = json.Unmarshal([]byte(bodyString), &data)

	if err != nil {
		return res, err
	}

	res = data

	return res, nil
}

func GetTokken(m map[string]string) (map[string]interface{}, error) {

	res := make(map[string]interface{})
	reqBody, _ := json.Marshal(m)
	url := "https://commonapi.mastersindia.co/oauth/access_token"

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))

	if err != nil {
		return res, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	bodyString := string(body)

	//fmt.Println(bodyString)

	var data map[string]interface{}
	err = json.Unmarshal([]byte(bodyString), &data)

	res["error"], _ = data["error"].(string)
	res["access_token"], _ = data["access_token"].(string)

	return res, nil
}

// RequestPayload represents the request payload structure
type RequestPayload struct {
	GSTNo       string `json:"gst_no"`
	ConsentText string `json:"consent_text"`
	Consent     string `json:"consent"`
}

func GetBefiscGSTData(gstNo string, authKey string) (map[string]interface{}, string, error) {
	// Prepare the payload
	payload := RequestPayload{
		GSTNo:       gstNo,
		ConsentText: "We confirm that we have obtained the consent of the respective customer to fetch their details from authorized sources using their GST",
		Consent:     "Y",
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, "", fmt.Errorf("error marshalling payload: %v", err)
	}

	// API endpoint and headers
	apiURL := "https://kyb.befisc.com/gst-advance/v2"
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authkey", authKey)

	// Send the request
	client := &http.Client{Timeout: 8 * time.Second}
	//need to put the timeout
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("error making API call: %v", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("API call failed with status code: %d", resp.StatusCode)
	}

	// Read and parse the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("error reading response: %v", err)
	}

	var responseData map[string]interface{}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return nil, "", fmt.Errorf("error unmarshalling response: %v", err)
	}

	return responseData, string(body), nil
}
