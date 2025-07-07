package livematchmakingapi

//response ...
const autoApproved = "AUTO APPROVED"
const autoRejected = "AUTO REJECTED"
const manVerify = "MANUAL VERIFICATION"

//rejected and approved reasons
const (
        TradeLegalNameMatch int = 1
        TradeCityZipMatch   int = 37
        InactiveGST         int = 23
        CancelledGST        int = 24
        SuspendedGST        int = 67
        Others              int = 68
        InvalidGST          int = 22
        StateMismatch       int = 71
)

var reasonsMap map[int]string = map[int]string{
        1:  "Trade Name Legal Name Matching",
        37: "Trade Name + City/ZIP matching",
        23: "Inactive GST",
        24: "Cancelled GST",
        67: "GST is Suspended",
        68: "Others",
        22: "Invalid GST",
        71: "State Mismatch",
}


//change all these three structures according to the stage-1 buckets.

//DispoWiseReson ... PSEUDO CONSTS
var DispoWiseReson map[string]int = map[string]int{
        "N17A": TradeCityZipMatch,
        "N10A": TradeCityZipMatch,
        "AA1":  TradeCityZipMatch,
        "6A":   TradeCityZipMatch,
        "AA2":  TradeLegalNameMatch,
        "10A":  TradeLegalNameMatch,
        "5T":   TradeCityZipMatch,
        "6T":   TradeLegalNameMatch,
        "7T":   TradeCityZipMatch,
        "N1A":  TradeCityZipMatch,
        "N1B":  TradeLegalNameMatch,
        "N1C":  TradeLegalNameMatch,
        "N2A":  TradeCityZipMatch,
        "N2B":  TradeCityZipMatch,
}

//errors
const (
        errParam         = "Invalid Params"
        errValidationKey = "Invalid Validation Key"
        errPanic         = "Panic ..Pls inform Dev Team"
        errFetchDB       = "error in fetching records from db"
        errDnfDB         = "no records found from db"
        errFetchAPI      = "err in fetching gst details from api"
        errUpdateDB      = "err in updating gst details in db"
)

//others
const (
        success            = "SUCCESS"
        failure            = "FAILURE"
        serviceName        = "live_match_making_api"
        logFileName        = "live_match_making_api.json"
        masterindiaAPILogs = ""
        serviceLogPath     = "/var/log/application/GST/LIVE_MATCH_MAKING"
)
