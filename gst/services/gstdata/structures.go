package gstdata

// Req ...
type Req struct {
	Glid          string `json:"glid"`
	Gst           string `json:"gst"`
	ModID         string `json:"modid"`
	Validationkey string `json:"validationkey"`
	Flag          string `json:"flag,omitempty"`
}

// Res ...
type Res struct {
	Code       int    `json:"code"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
	Body       Data   `json:"data,omitempty"`
	UniqID     string `json:"uniq_id,omitempty"`
	UpdateFlag int    `json:"updateflag"`
	GstNum     string `json:"gstnum"`
	HSNcode    string `json:"hsncodes"`
}

// Data ...
type Data struct {
	Values map[string]interface{} `json:"values,omitempty"`
}

// Logg ...
type Logg struct {
	RequestStart       string                 `json:"RequestStart,omitempty"`
	RequestStartValue  float64                `json:"request_start_value,omitempty"`
	RequestEndValue    float64                `json:"request_end_value,omitempty"`
	ResponseTime       float64                `json:"response_time,omitempty"`
	ResponseTime_Float float64                `json:"ResponseTime_Float,omitempty"` // float64 type
	ServiceName        string                 `json:"ServiceName,omitempty"`
	ServiceURL         string                 `json:"ServiceURL,omitempty"`
	RemoteAddress      string                 `json:"RemoteAddress,omitempty"`
	Request            Req                    `json:"request,omitempty"`
	Response           Res                    `json:"response,omitempty"`
	DbBeforeData       map[string]interface{} `json:"db_before_data,omitempty"`
	DbData             map[string]interface{} `json:"dbdata,omitempty"`
	DbBeforeFlag2      map[string]interface{} `json:"dbbeforeflag2,omitempty"`
	DbDataFlag2        map[string]interface{} `json:"dbdataflag2,omitempty"`
	AnyError           map[string]string      `json:"any_error,omitempty"`
	StackTrace         string                 `json:"stack_trace,omitempty"`
	ExecTime           map[string]float64     `json:"exec_time,omitempty"`
	MasterIndia        MasterIndia            `json:"master_india,omitempty"`
	QueueMsgID         string                 `json:"QueueMsgID"`
}

// MasterIndia ...
type MasterIndia struct {
	Hit  bool   `json:"hit,omitempty"`
	User string `json:"user,omitempty"`
}

// WorkRequest ...
type WorkRequest struct {
	APIName     string `json:"APIName,omitempty"`
	APIUserName string `json:"APIUserName,omitempty"`
	GstPan      string `json:"GstPan,omitempty"`
	Modid       string `json:"Modid,omitempty"`
	RqstTime    string `json:"RqstTime,omitempty"`
}

type VendorResp struct {
	Vendor string `json:"Vendor"`
	Mobile string `json:"Mobile"`
	Email  string `json:"Email"`
}

type OtpResponse struct {
	Code   int          `json:"code"`
	Status string       `json:"status"`
	Error  string       `json:"error,omitempty"`
	Data   []VendorResp `json:"data"`
	UniqID string       `json:"uniq_id,omitempty"`
}
