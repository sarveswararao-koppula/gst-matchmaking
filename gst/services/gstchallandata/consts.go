package gstchallandata

//errors
const (
	errParam         = "Invalid Params"
	errValidationKey = "Invalid Validation Key"
	errPanic         = "Panic ..Pls inform Dev Team"
	errFetchDB       = "error in fetching records from db"
	errDnfDB         = "no records found from db"
	errFetchAPI      = "err in fetching challan details from api"
	errUpdateDB      = "err in updating challan details in db"
	errNotAuth       = "Not Authorized"
	errfyAPI		 = "DB Error"
)

//others
const (
	success            = "SUCCESS"
	failure            = "FAILURE"
	serviceName        = "gst_challan_data"
	logFileName        = "gst_challan_data.json"
	log2FileName	   = "gst_wrapper_queue.json"
	//need to see the path of masterindiaAPILogs
	masterindiaAPILogs = "" //"/home/gst-user/abhay"
	serviceLogPath     = "/var/log/application/GST/GST_CHALLAN_DATA"
)
