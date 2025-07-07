package gstchallandata

//Req ...
type Req struct {
	Glid          string `json:"glid,omitempty"`
	Gst           string `json:"gst"`
	ModID         string `json:"modid"`
	Validationkey string `json:"validationkey"`
	Flag          string `json:"flag,omitempty"`
}

type Res struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
	Body    Data   `json:"data,omitempty"`
	Max_dof string `json:"max_dof"`
	UniqID  string `json:"uniq_id,omitempty"`
}

//Data ...
type Data struct {
	Values []map[string]string `json:"values,omitempty"`
}

//Logg ...
type Logg struct {
	RequestStart      string             `json:"request_start,omitempty"`
	RequestStartValue float64            `json:"request_start_value,omitempty"`
	RequestEndValue   float64            `json:"request_end_value,omitempty"`
	ResponseTime      float64            `json:"response_time,omitempty"`
	ServiceName       string             `json:"service_name,omitempty"`
	ServiceURL        string             `json:"service_url,omitempty"`
	RemoteAddress     string             `json:"RemoteAddress,omitempty"`
	Request           Req                `json:"request,omitempty"`
	Response          Res                `json:"response,omitempty"`
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

//work
// type Work struct {
// 	credential   string `json:"credent,omitempty"`
// 	APIUserID string `json:"api_user_id,omitempty"`
// 	GST       string `json:"gst,omitempty"`
// 	Modid     string `json:"modid,omitempty"`
// 	UniqID    string `json:"uniq_id,omitempty"`
// }
