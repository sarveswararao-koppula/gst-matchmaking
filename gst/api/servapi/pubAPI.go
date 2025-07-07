package servapi

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type data struct {
	Flag           string `json:"flag"`
	AttributeID    int    `json:"attribute_id"`
	AttributeValue string `json:"attribute_value"`
	IsManual       string `json:"isManual"`
}

type formdata struct {
	Glusridval      string `json:"glusridval"`
	Qname           string `json:"qname"`
	RService        string `json:"rservice"`
	Rid             string `json:"rid"`
	Msg             []data `json:"data"`
	Host            string `json:"host"`
	ServiceName     string `json:"SERVICENAME"`
	UniqueLoggingID string `json:"UNIQUE_LOGGING_ID"`
	Modid           string `json:"modid"`
}

//Publish ...
func Publish(glid string, rid string, gst string, host string) error {

	qName := "USER_APPROVAL_ATTRIBUTE"

	formData := formdata{
		Glusridval: glid,
		Qname:      qName,
		RService:   "BI_GST_MANUAL_APPROVAL",
		Rid:        rid,
		Msg: []data{
			data{
				Flag:           "I",
				AttributeID:    2106,
				AttributeValue: gst,
				IsManual:       "1",
			},
		},
		Host:            strings.ToLower(host),
		ServiceName:     "APPROVAL_ATTRIBUTE",
		UniqueLoggingID: rid,
		Modid:           "BI",
	}

	raw, err := json.Marshal(formData)
	if err != nil {
		return err
	}

	v := url.Values{}

	v.Set("qname", qName)
	v.Set("rid", rid)
	v.Set("rservice", "BI_GST_MANUAL_APPROVAL")

	//Form Data
	v.Set("msg", string(raw))

	return hitPubAPI(v, host)
}

func hitPubAPI(payLoad url.Values, env string) error {

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := ""
	if strings.ToUpper(env) == "PROD" {
		
		url = `http://prod-soa-rmq-api-mkp-messaging-imutils.imbi.prod/rmq/publish`

	} else if strings.ToUpper(env) == "DEV" {
		url = `http://162.217.96.117:8082/rmq/publish`
	}

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payLoad.Encode()))

	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	bodyString := string(body)
	//fmt.Println(bodyString)

	result := make(map[string]interface{})
	err = json.Unmarshal([]byte(bodyString), &result)

	if err != nil {
		return err
	}

	status, _ := result["status"].(string)

	if status != "Success" {
		return errors.New(bodyString)
	}

	return nil
}
