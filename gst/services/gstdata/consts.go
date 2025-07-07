package gstdata

// errors
const (
	errParam         = "Invalid request parameters. Please verify your input and try again."
	errValidationKey = "Validation key is invalid or expired. Please retry after sometime or please email to bi-support@indiamart.com."
	errPanic         = "Your request is taking longer to complete. Please retry after sometime or please email to bi-support@indiamart.com."
	errFetchDB       = "Error in fetching records from db."
	errDnfDB         = "Your request is taking longer to complete. Please retry after sometime or please email to bi-support@indiamart.com."
	errFetchAPI      = "Your request is taking longer to complete. Please retry after sometime or please email to bi-support@indiamart.com."
	errUpdateDB      = "err in updating gst details in db."
	errNotAuth       = "Not Authorized."
)

// errPanic         = "Panic ..Pls inform Dev Team"
// others
const (
	success            = "SUCCESS"
	failure            = "FAILURE"
	serviceName        = "gst_data"
	logFileName        = "gst_data.json"
	logKibanaFileName  = "gst_data_kibana.json"
	masterindiaAPILogs = "" //"/home/gst-user/abhay"
	serviceLogPath     = "/var/log/application/GST/GST_DATA"
)
