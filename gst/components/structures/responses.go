package structures

type Response_masterindiacontrols struct {
	Code       int                    `json:"code,omitempty"`
	Status     string                 `json:"status,omitempty"`
	ErrMessage string                 `json:"err_message,omitempty"`
	Body       map[string]interface{} `json:"body,omitempty"`
}

type ErrorLogData struct {
	ErrorTitle   string                 `json:"ErrorTitle,omitempty"`
	ErrorMessage interface{}            `json:"ErrorMessage,omitempty"`
	StackTrace   string                 `json:"StackTrace,omitempty"`
	ErrorData    map[string]interface{} `json:"ErrorData,omitempty"`
}

type Response_gstmmcontroller struct {
	Code int                    `json:"status,omitempty"`
	Err  string                 `json:"message,omitempty"`
	Data map[string]interface{} `json:"data,omitempty"`
}

