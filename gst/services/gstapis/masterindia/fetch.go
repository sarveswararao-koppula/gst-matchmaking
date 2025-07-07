package masterindia

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
	"fmt"
)

//FetchGSTDetails ...
func FetchGSTDetails(gst string, clientID string, accessToken string, timeout int) (map[string]interface{}, error) {

	url := "https://commonapi.mastersindia.co/commonapis/searchgstin?gstin=" + gst

	if timeout < 1000 {
		timeout = 1000
	} else if timeout > 2000 {
		timeout = 4000
	}

	client := http.Client{
		Timeout: time.Duration(time.Duration(timeout) * time.Millisecond),
	}

	rqst, err := http.NewRequest("GET", url, nil)
	rqst.Header.Set("Content-Type", "application/json")
	rqst.Header.Set("Authorization", "Bearer "+accessToken)
	rqst.Header.Set("client_id", clientID)

	if err != nil {
		return nil, err
	}

	resp, err := client.Do(rqst)
	
	fmt.Println("GST : ",gst)
	//fmt.Println("Resp : ",resp)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	bodyString := string(body)

	fmt.Println(bodyString)

	var data map[string]interface{}
	err = json.Unmarshal([]byte(bodyString), &data)

	return data, err
}

//GetTokken ...
func GetTokken(cred interface{}) (map[string]interface{}, error) {

	reqBody, _ := json.Marshal(cred)
	url := "https://commonapi.mastersindia.co/oauth/access_token"

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	bodyString := string(body)

	//fmt.Println(bodyString)

	var data map[string]interface{}
	err = json.Unmarshal([]byte(bodyString), &data)

	return data, err
}
