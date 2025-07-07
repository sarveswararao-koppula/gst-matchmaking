package gstvalidator

//Req ...
type Req struct {
	// Glid          string `json:"glid"`
	Gst           string `json:"gst"`
	ModID         string `json:"modid"`
	Validationkey string `json:"validationkey"`
}

//Res ...
type Res struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Body   Data   `json:"data,omitempty"`
	UniqID string `json:"uniq_id,omitempty"`
}

//Data ...
type Data struct {
	Values map[string]interface{} `json:"values,omitempty"`
}

//Logg ...
type Logg struct {
	RequestStart      string             `json:"RequestStart,omitempty"`
	RequestStartValue float64            `json:"request_start_value,omitempty"`
	RequestEndValue   float64            `json:"request_end_value,omitempty"`
	ResponseTime      float64            `json:"response_time,omitempty"`
	ResponseTime_Float float64 `json:"ResponseTime_Float,omitempty"` // float64 type
	ServiceName       string             `json:"ServiceName,omitempty"`
	ServiceURL        string             `json:"ServiceURL,omitempty"`
	RemoteAddress     string             `json:"RemoteAddress,omitempty"`
	Request           Req                `json:"Request,omitempty"`
	Response          Res                `json:"Response,omitempty"`
	AnyError          map[string]string  `json:"any_error,omitempty"`
	StackTrace        string             `json:"stack_trace,omitempty"`
	ExecTime          map[string]float64 `json:"exec_time,omitempty"`
	MasterIndia       MasterIndia        `json:"master_india,omitempty"`
}

//MasterIndia ...
type MasterIndia struct {
	Hit  bool   `json:"hit,omitempty"`
	User string `json:"user,omitempty"`
}


