package livematchmakingapi

import "mm/services/gstmmcontrols"

// Req ...
type Req struct {
	Glid          string `json:"glid"`
	GST           string `json:"gst,omitempty"`
	PAN           string `json:"pan,omitempty"`
	IEC           string `json:"iec,omitempty"`
	Udyam         string `json:"udyam,omitempty"`
	ModID         string `json:"modid"`
	ValidationKey string `json:"validationkey"`
}

// Res ...
type Res struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Body   Data   `json:"data,omitempty"`
	UniqID string `json:"uniq_id,omitempty"`
}

// Data ...
type Data struct {
	Flag               string            `json:"flag,omitempty"`
	ReasonID           int               `json:"reason,omitempty"`
	Reason             string            `json:"reason_desc,omitempty"`
	BucketName         string            `json:"bucket_name,omitempty"`
	Gstdetails         map[string]string `json:"gstdetails,omitempty"`
	GstVerificationSrc interface{}       `json:"gst_verification_src,omitempty"`
}

type GstVerification struct {
	Flag          int               `json:"flag"`
	Attribute_src map[string]string `json:"attribute_src"`
}

type GstDetail struct {
	MobileNumber    string `json:"MobileNumber,omitempty"`
	MobileAttribute string `json:"MobileAttribute,omitempty"`
	EmailId         string `json:"EmailId,omitempty"`
	EmailAttribute  string `json:"EmailAttribute,omitempty"`
}

type User struct {
	glusr_usr_email     string
	glusr_usr_email_alt string
}

// Logg ...
type Logg struct {
	RequestStart            string               `json:"RequestStart,omitempty"`
	RequestStartValue       float64              `json:"RequestStartValue,omitempty"`
	RequestEndValue         float64              `json:"RequestEndValue,omitempty"`
	ResponseTime            float64              `json:"ResponseTime,omitempty"`
	ServiceName             string               `json:"ServiceName,omitempty"`
	ServiceURL              string               `json:"ServiceURL,omitempty"`
	RemoteAddress           string               `json:"RemoteAddress,omitempty"`
	TacticalAttributeSource string               `json:"TacticalAttributeSource,omitempty"`
	Request                 Req                  `json:"Request,omitempty"`
	Response                Res                  `json:"Response,omitempty"`
	AnyError                map[string]string    `json:"AnyError,omitempty"`
	StackTrace              string               `json:"StackTrace,omitempty"`
	ExecTime                map[string]float64   `json:"ExecTime,omitempty"`
	MasterIndia             MasterIndia          `json:"MasterIndia,omitempty"`
	ScoreDetails            gstmmcontrols.Score  `json:"ScoreDetails,omitempty"`
	ScoreDetailsStage1      gstmmcontrols.Score1 `json:"ScoreDetailsStg1,omitempty"`
	Data                    map[string]string    `json:"Data,omitempty"`
	UpdateFlags             map[string]bool      `json:"UpdateFlags,omitempty"`
	User_verification_date  string               `json:"User_verification_date,omitempty"`
}

// MasterIndia ...
type MasterIndia struct {
	Hit  bool   `json:"Hit,omitempty"`
	User string `json:"User,omitempty"`
}

type LogEntry struct {
	RequestStart            string  `json:"RequestStart,omitempty"`
	ResponseTime_Float      float64 `json:"ResponseTime_Float,omitempty"` // float64 type
	ServiceName             string  `json:"ServiceName,omitempty"`
	ServiceURL              string  `json:"ServiceURL,omitempty"`
	RemoteAddress           string  `json:"RemoteAddress,omitempty"`
	Request_Data            string  `json:"Request_Data,omitempty"`
	Response_Body           string  `json:"Response_Body,omitempty"`
	Any_Error               string  `json:"Any_Error,omitempty"`
	StackTrace              string  `json:"Stack_trace,omitempty"`
	Glusr_usr_id            string  `json:"Glusr_usr_id,omitempty"`
	Gst                     string  `json:"Gst,omitempty"`
	Modid                   string  `json:"Modid,omitempty"`
	ValidationKey           string  `json:"ValidationKey,omitempty"`
	TacticalAttributeSource string  `json:"TacticalAttributeSource,omitempty"`
	User_verification_date  string  `json:"User_verification_date,omitempty"`
}
