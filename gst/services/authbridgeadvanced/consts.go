package authbridgeadvanced

//errors
const (
	errParam     = "Invalid Params"
	errUnmarshal = "Error while Unmarshalling"
	errPanic     = "Panic ..Pls inform Dev Team"
	errFetchDB   = "error in fetching records from db"
	errDnfDB     = "no records found from db"
	errFetchAPI  = "Status is 0"
	errUpdateDB  = "err in updating gst details in db"
	errNotAuth   = "Not Authorized"
	errfyAPI     = "DB Error"
)

//others
const (
	success      = "SUCCESS"
	failure      = "FAILURE"
	serviceName  = "AuthBridge_Advanced-GST-Search"
	logFileName  = "Authbridge_hits.json"
	logKibanaFileName  = "Authbridge_hits_Kibana.json"
	log2FileName = "gst_wrapper_queue.json"
	//need to see the path of masterindiaAPILogs
	masterindiaAPILogs = "" //"/home/gst-user/abhay"
	serviceLogPath     = "/var/log/application/GST/AUTHBRIDGE_HITS"
)


