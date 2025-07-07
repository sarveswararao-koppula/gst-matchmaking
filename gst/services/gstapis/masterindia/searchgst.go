package masterindia

import (
	"encoding/json"
	"errors"
	"fmt"
	"mm/utils"
	"os"
	"time"
)

//FetchGSTData ...
func (wr Work) FetchGSTData(filePath string, timeOut int) (map[string]string, []interface{}, error) {

	var params []interface{}

	logg := S3Log{
		APIHit:            false,
		RequestStart:      utils.GetTimeStampCurrent(),
		RequestStartValue: utils.GetTimeInNanoSeconds(),
		Request:           wr,
		Result:            make(map[string]string),
		AnyError:          make(map[string]string),
	}
	err := ValidateTokken(logg.Request.APIUserID)

	if err != nil {
		logg.AnyError["ValidateTokken"] = err.Error()
		write2(logg, filePath)
		return nil, params, err
	}

	logg.APIHit = true
	data, err := FetchGSTDetails(logg.Request.GST, creds[logg.Request.APIUserID].ClientID, tokens[logg.Request.APIUserID].tok, timeOut)
	//fmt.Println("GST: ",logg.Request.GST," ",data)
	if err != nil {
		logg.AnyError["FetchGSTDetails"] = err.Error()
		write2(logg, filePath)
		return nil, params, err
	}

	apiData, ok := data["data"].(map[string]interface{})
	//fmt.Println("GST: ",logg.Request.GST," ",apiData)
	dataErrorBool, _ := data["error"].(bool)
	dataErrorStr, _ := data["error"].(string)

	if !ok || dataErrorBool || dataErrorStr != "" {

		logg.AnyError["FetchGSTDetails"] = fmt.Sprint(data)

		if dataErrorStr == "invalid_grant" {
			tokens[logg.Request.APIUserID].exp = time.Now().In(loc).Add(-24 * time.Hour)
		}
		write2(logg, filePath)
		return nil, params, errors.New(fmt.Sprint(data))
	}

	logg.Result, params = utils.BusLogicOnMasterData_V2(logg.Request.GST, apiData)
	fmt.Println("BUS-LOGIC :  ",logg.Result)
	logg.RequestEndValue = utils.GetTimeInNanoSeconds()
	logg.ResponseTime = (logg.RequestEndValue - logg.RequestStartValue) / 1000000
	logg.RequestStartValue = 0
	logg.RequestEndValue = 0

	write2(logg, filePath)
	return logg.Result, params, nil
}

func write2(logg S3Log, filePath string) {

	if filePath == "" {
		filePath = serviceLogPath + utils.TodayDir()
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		e := os.MkdirAll(filePath, os.ModePerm)
		if e != nil {
			return
		}
	}

	filePath += "/S3LOG.json"
	jsonLog, _ := json.Marshal(logg)

	f, _ := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()

	f.WriteString("\n" + string(jsonLog))
}
