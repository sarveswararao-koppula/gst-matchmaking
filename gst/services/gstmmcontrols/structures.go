package gstmmcontrols

//Req ...
type Req struct {
	Glid     string `json:"glid"`
	UniqueID string `json:"UniqueId,omitempty"`
}

//Res ...
type Res struct {
	Code int      `json:"status,omitempty"`
	Err  string   `json:"message,omitempty"`
	Body GSTMatch `json:"data,omitempty"`
}

//Body ...
// type Body struct {
// 	Gstin            string `json:"Gstin,omitempty"`
// 	BucketType       string `json:"BucketType,omitempty"`
// 	BucketName       string `json:"BucketName,omitempty"`
// 	GstStatus        string `json:"GstStatus,omitempty"`
// 	GstInsertionDate string `json:"GstInsertionDate,omitempty"`
// }

//Logg ...
type Logg struct {
	RequestStart      string             `json:"RequestStart,omitempty"`
	RequestStartValue float64            `json:"RequestStartValue,omitempty"`
	RequestEndValue   float64            `json:"RequestEndValue,omitempty"`
	ResponseTime      float64            `json:"ResponseTime,omitempty"`
	ServiceName       string             `json:"ServiceName,omitempty"`
	ServiceURL        string             `json:"ServiceURL,omitempty"`
	RemoteAddress     string             `json:"RemoteAddress,omitempty"`
	Request           Req                `json:"Request,omitempty"`
	Response          Res                `json:"Response,omitempty"`
	AnyError          map[string]string  `json:"AnyError,omitempty"`
	StackTrace        string             `json:"StackTrace,omitempty"`
	ExecTime          map[string]float64 `json:"ExecTime,omitempty"`
	MasterIndia       MasterIndia        `json:"MasterIndia,omitempty"`
	ApprovalDone      string             `json:"ApprovalDone,omitempty"`
	ScoreDetails      Score              `json:"ScoreDetails,omitempty"`
	ScoreDetailsStg1  Score1             `json:"ScoreDetailsStg1,omitempty"`
	UpdateFlags       map[string]bool    `json:"UpdateFlags,omitempty"`
	KeyWordFlag       string             `json:"KeyWordFlag,omitempty"`
	CustTypeFlag      string             `json:"CustTypeFlag,omitempty"`
	VerifyParams      map[string]string  `json:"VerifyParams,omitempty"`
	ContactSource     string              `json:"ContactSource,omitempty"`
}

type Logkibana struct {
	RequestStart       string  `json:"RequestStart,omitempty"`
	ResponseTime_Float float64 `json:"ResponseTime_Float,omitempty"`
	ServiceName        string  `json:"ServiceName,omitempty"`
	ServiceURL         string  `json:"ServiceURL,omitempty"`
	RemoteAddress      string  `json:"RemoteAddress,omitempty"`
	Request_Data       string  `json:"Request_Data,omitempty"`
	Response_Body      string  `json:"Response_Body,omitempty"`
	Any_Error          string  `json:"Any_Error,omitempty"`
	StackTrace         string  `json:"StackTrace,omitempty"`
	CustTypeFlag       string  `json:"CustTypeFlag,omitempty"`
	ContactSource      string  `json:"ContactSource,omitempty"`
}

type LogKibanaWorker struct {
	LogType                string `json:"Consumer,omitempty"`
	RequestStart           string `json:"RequestStart,omitempty"`
	ServiceName            string `json:"ServiceName,omitempty"`
	ServiceURL             string `json:"ServiceURL,omitempty"`
	RemoteAddress          string `json:"RemoteAddress,omitempty"`
	Request_Data           string `json:"Request_Data,omitempty"`
	Response_Body          string `json:"Response_Body,omitempty"`
	Any_Error              string `json:"Any_Error,omitempty"`
	MasterIndia_Hit_Status string `json:"MasterIndia_Hit_Status,omitempty"`
	Gstin                  string `json:"Gstin,omitempty"`
	BucketType             string `json:"BucketType,omitempty"`
	BucketName             string `json:"BucketName,omitempty"`
}

//MasterIndia ...
type MasterIndia struct {
	Hit  bool   `json:"Hit,omitempty"`
	User string `json:"User,omitempty"`
	Err  string `json:"Err,omitempty"`
}

//Score ... glid gst params matching scores details
type Score struct {
	PinLen                        int     `json:"PinLen,omitempty"`
	OwnerNameScore                float64 `json:"OwnerNameScore,omitempty"`
	CeoNameScore                  float64 `json:"CeoNameScore,omitempty"`
	AddressScore                  float64 `json:"AddressScore,omitempty"`
	AddressScoreWoSCP             float64 `json:"AddressScoreWoSCP,omitempty"`
	AddressScoreSecond            float64 `json:"AddressScoreSecond,omitempty"`
	OwnerNameLen                  int     `json:"OwnerNameLen,omitempty"`
	CeoNameLen                    int     `json:"CeoNameLen,omitempty"`
	AddressLen                    int     `json:"AddressLen,omitempty"`
	AddressLenWoSCP               int     `json:"AddressLenWoSCP,omitempty"`
	AddressLenWoSCP2EachWordLenG2 bool    `json:"AddressLenWoSCP2EachWordLenG2,omitempty"`
	PincodeScore                  int     `json:"PincodeScore,omitempty"`
	IsTradeBizSame                bool    `json:"IsTradeBizSame,omitempty"`
	IsME                          bool    `json:"IsME,omitempty"`
	IsSateSame                    bool    `json:"IsSateSame,omitempty"`
}

//Stage 1 Score
type Score1 struct {
	CompanyNameScore float64
	StateScore       float64
	CityScore        float64
	AddressScore     float64
	PinScore         int
	OwnerNameScore   float64
	GlOwnerNameLen   int
	GlPinLen         int
	GlAddressLen     int
	IsuniqInState    bool
	IsuniqInIndia    bool
}

//GSTMatch ...
type GSTMatch struct {
	Gstin            string `json:"Gstin,omitempty"`
	BucketType       string `json:"BucketType,omitempty"`
	BucketName       string `json:"BucketName,omitempty"`
	scores           Score
	scoresStage1     Score1 `json:"ScoresStage1,omitempty"`
	GstStatus        string `json:"GstStatus,omitempty"`
	GstInsertionDate string `json:"GstInsertionDate,omitempty"`
	GstPincode       string `json:"GstPincode,omitempty"`
	GstAddress       string `json:"GstAddress,omitempty"`
	GstState         string `json:"GstState,omitempty"`
	TradeName        string `json:"TradeName,omitempty"`
}
