package structures

type Logs_gstmmcontroller struct {
	RequestStart       string                   `json:"RequestStart,omitempty"`
	RequestStartValue  float64                  `json:"RequestStartValue,omitempty"`
	RequestEnd         string                   `json:"RequestEnd,omitempty"`
	RequestEndValue    float64                  `json:"RequestEndValue,omitempty"`
	ResponseTime       string                   `json:"ResponseTime,omitempty"`
	ServiceName        string                   `json:"ServiceName,omitempty"`
	ServicePath        string                   `json:"ServicePath,omitempty"`
	ServiceUrl         string                   `json:"ServiceUrl,omitempty"`
	RemoteAddress      string                   `json:"RemoteAddress,omitempty"`
	UniqueId           string                   `json:"UniqueId,omitempty"`
	QueryParams        map[string]interface{}   `json:"QueryParams,omitempty"`
	Response           Response_gstmmcontroller `json:"Response,omitempty"`
	PostServicesCalled map[string]string        `json:"PostServicesCalled,omitempty"`
	MasterIndiaApiHit  map[string]interface{}   `json:"MasterIndiaApiHit,omitempty"`
	ApprovalDone       string                   `json:"ApprovalDone,omitempty"`
	StackTrace         string                   `json:"StackTrace,omitempty"`
	ScoreDetails       string                   `json:"ScoreDetails,omitempty"`
}

type Logs_masterindiacontrols struct {
	RemoteAddress     string                            `json:"RemoteAddress,omitempty"`
	RequestStart      string                            `json:"RequestStart,omitempty"`
	RequestStartValue float64                           `json:"RequestStartValue,omitempty"`
	RequestEnd        string                            `json:"RequestEnd,omitempty"`
	RequestEndValue   float64                           `json:"RequestEndValue,omitempty"`
	ResponseTime      string                            `json:"ResponseTime,omitempty"`
	ServiceName       string                            `json:"ServiceName,omitempty"`
	ServicePath       string                            `json:"ServicePath,omitempty"`
	ServiceUrl        string                            `json:"ServiceUrl,omitempty"`
	UniqueId          string                            `json:"UniqueId,omitempty"`
	RequestFormData   map[string]string                 `json:"RequestFormData,omitempty"`
	Response          Response_masterindiacontrols      `json:"Response,omitempty"`
	StackTrace        string                            `json:"StackTrace,omitempty"`
	ApiHit            map[string]map[string]interface{} `json:"ApiHit,omitempty"`
	TableUpdated      map[string]interface{}            `json:"TableUpdated,omitempty"`
}

type Masterindiacontrols struct {
	RemoteAddress     string                       `json:"RemoteAddress,omitempty"`
	RequestStart      string                       `json:"RequestStart,omitempty"`
	RequestStartValue float64                      `json:"RequestStartValue,omitempty"`
	RequestEnd        string                       `json:"RequestEnd,omitempty"`
	RequestEndValue   float64                      `json:"RequestEndValue,omitempty"`
	ResponseTime      string                       `json:"ResponseTime,omitempty"`
	ServiceName       string                       `json:"ServiceName,omitempty"`
	ServicePath       string                       `json:"ServicePath,omitempty"`
	ServiceUrl        string                       `json:"ServiceUrl,omitempty"`
	UniqueId          string                       `json:"UniqueId,omitempty"`
	RequestFormData   map[string]string            `json:"RequestFormData,omitempty"`
	Response          Response_masterindiacontrols `json:"Response,omitempty"`
	StackTrace        string                       `json:"StackTrace,omitempty"`
	Api               interface{}                  `json:"Api,omitempty"`
}

