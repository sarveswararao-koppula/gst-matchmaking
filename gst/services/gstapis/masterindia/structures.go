package masterindia

//Work ...
type Work struct {
	APIName   string `json:"api_name,omitempty"`
	APIUserID string `json:"api_user_id,omitempty"`
	GST       string `json:"gst,omitempty"`
	Modid     string `json:"modid,omitempty"`
	UniqID    string `json:"uniq_id,omitempty"`
}

//S3Log ...
type S3Log struct {
	RequestStart      string            `json:"RequestStart,omitempty"`
	RequestStartValue float64           `json:"request_start_value,omitempty"`
	RequestEndValue   float64           `json:"request_end_value,omitempty"`
	ResponseTime      float64           `json:"response_time,omitempty"`
	Request           Work              `json:"request,omitempty"`
	APIHit            bool              `json:"api_hit,omitempty"`
	Result            map[string]string `json:"result,omitempty"`
	AnyError          map[string]string `json:"any_error,omitempty"`
}

type cred struct {
	UserName     string `json:"username"`
	Password     string `json:"password"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
}
