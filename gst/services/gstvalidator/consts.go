package gstvalidator

//errors
const (
	errParam         = "Invalid Params"
	errValidationKey = "Invalid Validation Key"
	errPanic         = "Panic ..Pls inform Dev Team"
	errFetchDB       = "error in fetching records from db"
	errDnfDB         = "no records found from db"
	errFetchAPI      = "err in fetching gst details from api"
	errUpdateDB      = "err in updating gst details in db"
	errNotAuth       = "Not Authorized"
	InvalidGst		 = "InValid GST"
	DuplicateGst	 = "GST Already Assigned to Some other GLID"
)

//others
const (
	success            = "SUCCESS"
	failure            = "FAILURE"
	serviceName        = "gstvalidator"
	logFileName        = "gst_validator_data.json"
	masterindiaAPILogs = "" //"/home/gst-user/abhay"
	serviceLogPath     = "/var/log/application/GST/GST_VALIDATOR"
)

