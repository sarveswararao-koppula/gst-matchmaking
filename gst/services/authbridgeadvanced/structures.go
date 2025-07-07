package authbridgeadvanced

//Req ...
// type Req struct {
//         Trans_id      string `json:"ts_trans_id"`
//         Status        int `json:"status"`
//         Msg               string `json:"msg"`
// }

//newReq ...
type Req struct {
	Responsedata string `json:"vendor_response_data"`
}

type Res struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	UniqID  string `json:"uniq_id,omitempty"`
}

//Logg ...
type Logg struct {
	RequestStart       string             `json:"RequestStart,omitempty"`
	RequestStartValue  float64            `json:"request_start_value,omitempty"`
	RequestEndValue    float64            `json:"request_end_value,omitempty"`
	ResponseTime       float64            `json:"response_time,omitempty"`
	ResponseTime_Float float64            `json:"ResponseTime_Float,omitempty"` // float64 type
	ServiceName        string             `json:"ServiceName,omitempty"`
	ServiceURL         string             `json:"ServiceURL,omitempty"`
	RemoteAddress      string             `json:"RemoteAddress,omitempty"`
	Request            Req                `json:"request,omitempty"`
	Response           Res                `json:"response,omitempty"`
	Result             map[string]string  `json:"result,omitempty"`
	ResultHSN          map[string]string  `json:"resultHSN,omitempty"`
	AnyError           map[string]string  `json:"any_error,omitempty"`
	StackTrace         string             `json:"stack_trace,omitempty"`
	ExecTime           map[string]float64 `json:"exec_time,omitempty"`
	STATUS             int                `json:"STATUS"`
}

//PubApiLogg ...
type PubApiLogg struct {
	RequestStart       string  `json:"RequestStart,omitempty"`
	ResponseTime       float64 `json:"response_time,omitempty"`
	ResponseTime_Float float64 `json:"ResponseTime_Float,omitempty"` // float64 type
	ServiceName        string  `json:"ServiceName,omitempty"`
	ServiceURL         string  `json:"ServiceURL,omitempty"`
	DataMessage        string  `json:"DataMessage,omitempty"`
	GstPubApiUrl       string  `json:"GstPubApiUrl,omitempty"`
	// Request            Req                `json:"request,omitempty"`
	// Response       Res    `json:"response,omitempty"`
	PubApiResponse string `json:"PubApiResponse,omitempty"`
	PubApiError    string `json:"PubApiError,omitempty"`
	Gst            string `json:"gst,omitempty"`
	// Glid           string `json:"glid,omitempty"`
	STATUS         int    `json:"STATUS,omitempty"`
}
