package gstchallandata

import(
        "fmt"
        "encoding/json"
        model "mm/model/masterindiamodel"
        "runtime/debug"
        "time"
        "mm/utils"
        "os"
        "strings"
)

type Qlog struct{
        APIName   string                 `json:"APIName,omitempty"`
        APIUserID string                 `json:"APIUserID,omitempty"`
        APIHit    string                 `json:"APIHit,omitempty"`
        Gst       string                 `json:"Gst,omitempty"`
        Modid     string                 `json:"Modid,omitempty"`
        RqstTime  string                 `json:"RqstTime,omitempty"`
        Result    map[string]interface{} `json:"Result,omitempty"`
}

//SubcriberHandler ...
func SubcriberHandler(data string) {
        s3log := S3Log{}
        err := json.Unmarshal([]byte(data), &s3log)
        s3log.AnyError = make(map[string]string)
        if err != nil {
                s3log.AnyError["SubcriberHandler Unmarshal"] = err.Error()
                writeSLog2(s3log)
                return
        }
        postwork(s3log)
}

func postwork(logg S3Log){
        defer func() {
                if panicCheck := recover(); panicCheck != nil {
                        stack := string(debug.Stack())
                        logg.StackTrace = stack
                        writeSLog2(logg)
                        return
                }
        }()
        //apiDataClosure := make(map[string]interface{})
        //apiDataClosure["i_u_d"] = "I"
        //to be send on workerlogic
        var qlog Qlog
        qlog.APIName = logg.APIName
        qlog.APIUserID = logg.APIUserID
        qlog.APIHit = logg.APIHit
        qlog.Gst = logg.Gst
        qlog.Modid = logg.Modid
        qlog.RqstTime = utils.GetTimeStampCurrent()
        qlog.Result = make(map[string]interface{})
        gstin := logg.Gst
        res := strings.Split(gstin, "#")
        gstinNumber := ""
        for _, v := range res {
                gstinNumber = v
                break
        }
        chllandataArr,_ := logg.Result["apiData"].([]interface{})
        var challandataArr2 []interface{}
        for _,v := range chllandataArr{
                //fetching the value of v and passing it here in chllan data at eachb case
                //var chllandata map[string]interface{}
                apiDataClosure := make(map[string]interface{})
                chllandata1,ok := v.(map[string]interface{})
                //chllandata,ok := v["data"]
                chllandata, _ := chllandata1["data"].(map[string]interface{})
                //dofStr,m := chllandata["dof"].(string)
                apiDataClosure["i_u_d"] = "I"
                if ok{
                        _, err := model.InsertChallanDetails(database, gstinNumber, chllandata)
                        if err != nil {
                                apiDataClosure["i_u_d_error"] = err.Error()
                        }

                }
                apiDataClosure["data"] = chllandata
                challandataArr2 = append(challandataArr2, apiDataClosure)
        }
        qlog.Result["apiData"] = challandataArr2
        writeQLog2(qlog)
        fyLatestFiling := logg.Result["max_dof"].(string)
        fyLatestFilingOrig := logg.Result["max_dof_orig"].(string)
        if flDate, err1 := time.Parse("02-01-2006", fyLatestFiling); err1 == nil {

                flDateOrig, err2 := time.Parse("02-01-2006", fyLatestFilingOrig)

                if err2 != nil || (err2 == nil && flDate.Sub(flDateOrig) > 0) {

                        fyLatestFiling = flDate.Format("2006-01-02")

                        var params []interface{}
                        now := time.Now().In(loc).Format("2006-01-02 15:04:05")

                        params = append(params, gstinNumber, fyLatestFiling, now)
                        //fmt.Println(gstinNumber,"Dev-Update")
                        errU := model.UpdateMasterDataFilingDate(database, params)
                        fmt.Println("err_u", errU)
                }

        }
}


//writeLog2 ...
func writeQLog2(logg Qlog) {

        logsDir := serviceLogPath + utils.TodayDir()

        if _, err := os.Stat(logsDir); os.IsNotExist(err) {
                        e := os.MkdirAll(logsDir, os.ModePerm)
                        if e != nil {
                                        fmt.Println(e)
                                        return
                        }
        }

        logsDir += "/" + log2FileName

        jsonLog, _ := json.Marshal(logg)

        f, _ := os.OpenFile(logsDir, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
        defer f.Close()

        f.WriteString("\n" + string(jsonLog))

        fmt.Println("\n" + string(jsonLog))
        return
}

func writeSLog2(logg S3Log) {

        logsDir := serviceLogPath + utils.TodayDir()

        if _, err := os.Stat(logsDir); os.IsNotExist(err) {
                        e := os.MkdirAll(logsDir, os.ModePerm)
                        if e != nil {
                                        fmt.Println(e)
                                        return
                        }
        }

        logsDir += "/" + log2FileName

        jsonLog, _ := json.Marshal(logg)

        f, _ := os.OpenFile(logsDir, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
        defer f.Close()

        f.WriteString("\n" + string(jsonLog))

        fmt.Println("\n" + string(jsonLog))
        return
}
