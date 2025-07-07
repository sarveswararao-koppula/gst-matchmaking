package livematchmakingapi

import (
	"encoding/json"
	"errors"
	"fmt"
	servapi "mm/api/servapi"
	"mm/components/constants"
	model "mm/model/masterindiamodel"
	"mm/properties"
	masterindia "mm/services/gstapis/masterindia"
	"mm/services/gstmmcontrols"
	workerlogic "mm/services/gstmmcontrols"
	"mm/utils"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"sync"

	// servapi "mm/api/servapi"

	"github.com/rs/xid"
)

type bucket struct {
	Pr    int
	Dispo string
}

var database string = properties.Prop.DATABASE
var MannualTactical string
var logFileMutex sync.Mutex

// Column / key  ➜  Attribute-ID expected by VerifyGlidAllAttr
var attrID = map[string]string{
	"COMPANYNAME":   "111",
	"ADD1":          "112",
	"ADD2":          "113",
	"CFIRSTNAME":    "141",
	"CLASTNAME":     "142",
	"FIRSTNAME":     "106",
	"LASTNAME":      "108",
	"ZIP":           "117",
	"FK_GL_CITY_ID": "152",
	"CITY":          "114",
	"STATE":         "115",
}

// Match ...
func Match(w http.ResponseWriter, r *http.Request) {

	uniqID := xid.New().String()
	var logg Logg
	logg.RequestStart = utils.GetTimeStampCurrent()
	logg.RequestStartValue = utils.GetTimeInNanoSeconds()
	logg.ServiceName = serviceName
	logg.ServiceURL = r.RequestURI
	logg.UpdateFlags = make(map[string]bool)
	logg.AnyError = make(map[string]string)
	logg.ExecTime = make(map[string]float64)

	if parts := strings.Split(r.RemoteAddr, ":"); len(parts) == 2 {
		logg.RemoteAddress = parts[0]
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&logg.Request)

	if err != nil {
		sendResponse(uniqID, w, 400, failure, errParam, err, Data{}, logg)
		return
	}

	// Determine identifier type and value
	var idType, idValue string
	switch {
	case strings.TrimSpace(logg.Request.GST) != "":
		idType, idValue = "gst", logg.Request.GST
	case strings.TrimSpace(logg.Request.PAN) != "":
		idType, idValue = "pan", logg.Request.PAN
	case strings.TrimSpace(logg.Request.Udyam) != "":
		idType, idValue = "udyam", logg.Request.Udyam
	case strings.TrimSpace(logg.Request.IEC) != "":
		idType, idValue = "iec", logg.Request.IEC
	}

	// --- START: Hardcoded DEV Responses ---
	if strings.EqualFold(properties.Prop.SERVICES_ENV, "DEV") {
		fmt.Printf("DEV ENV: checking hardcoded match for %s | %s\n", logg.Request.Glid, idValue)
		if data, ok := HardcodedResponse(logg.Request.Glid, idType, idValue); ok {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(data)
			return
		}
		fmt.Println("HARD MATCH NOT FOUND for:", logg.Request.Glid, idValue)
	}

	// ... (remaining logic unchanged)
	// --- END: Hardcoded DEV Responses ---
	const BIValidationKeyFromSOA = "af7f0273997b9b290bd7c57aa19f36c2"
	screenName := "Auto Approval GST Process"
	fmt.Println("ScreenName: ", screenName)
	const RemoteHost = "107.22.229.251"
	glid := logg.Request.Glid
	fmt.Println("glid: ", glid)
	apiName := "masterindia"
	APIUserID, err := masterindia.ValidateProp(logg.Request.ModID, logg.Request.ValidationKey, apiName)

	if err != nil {
		sendResponse(uniqID, w, 400, failure, errValidationKey, err, Data{}, logg)
		return
	}

	defer func() {
		if panicCheck := recover(); panicCheck != nil {
			stack := string(debug.Stack())
			logg.StackTrace = stack
			sendResponse(uniqID, w, 500, failure, errPanic, nil, Data{}, logg)
			return
		}
	}()

	_, st := utils.GetExecTime()
	glidRecords, err := GetGlidRecords(database, logg.Request.Glid)
	logg.ExecTime["GetGlidRecords"], st = utils.GetExecTime(st)

	if err != nil {
		sendResponse(uniqID, w, 200, failure, errFetchDB, err, Data{}, logg)
		return
	}

	if len(glidRecords) == 0 {
		sendResponse(uniqID, w, 200, failure, errDnfDB, errors.New(errDnfDB), Data{}, logg)
		return
	}

	gstRecords, err := GetGSTRecords(database, logg.Request.GST)
	logg.ExecTime["GetGSTRecords_1"], st = utils.GetExecTime(st)

	if err != nil {
		sendResponse(uniqID, w, 200, failure, errFetchDB, err, Data{}, logg)
		return
	}

	//checking ..api hit required? if yes then udate or insert data ?
	iu := ""
	if len(gstRecords) == 0 {
		logg.MasterIndia.Hit = true
		logg.MasterIndia.User = APIUserID
		iu = "I"
	} else if days, err := utils.DiffDaysddmmyyyy(gstRecords["gst_insertion_date"]); err != nil || days > 30 || strings.Trim(strings.ToLower(gstRecords["gstin_status"]), " ") != "active" {
		logg.MasterIndia.Hit = true
		logg.MasterIndia.User = APIUserID
		iu = "U"
	}

	if logg.MasterIndia.Hit {
		wr := masterindia.Work{
			APIName:   apiName,
			APIUserID: logg.MasterIndia.User,
			GST:       logg.Request.GST,
			Modid:     logg.Request.ModID,
			UniqID:    uniqID,
		}

		_, st = utils.GetExecTime()
		_, params, err := wr.FetchGSTData(masterindiaAPILogs, 3300)
		logg.ExecTime["FetchGSTData"], st = utils.GetExecTime(st)

		if err != nil {
			sendResponse(uniqID, w, 200, failure, errFetchAPI, err, Data{}, logg)
			return
		}

		if strings.ToUpper(iu) == "U" {
			_, err = model.UpdateGSTMasterData(database, params)
		} else if strings.ToUpper(iu) == "I" {
			_, err = model.InsertGSTMasterData(database, params)
		}
		logg.ExecTime["iu_gst_master_tab"], st = utils.GetExecTime(st)

		if err != nil {
			sendResponse(uniqID, w, 200, failure, errUpdateDB, err, Data{}, logg)
			return
		}

		gstRecords, err = GetGSTRecords(database, logg.Request.GST)
		logg.ExecTime["GetGSTRecords_2"], st = utils.GetExecTime(st)

		if err != nil {
			sendResponse(uniqID, w, 200, failure, errFetchDB, err, Data{}, logg)
			return
		}
	}

	records := make(map[string]string)
	for k, v := range glidRecords {
		records[k] = v
	}
	for k, v := range gstRecords {
		records[k] = v
	}

	responses := make(map[string]string)

	// matchmaking := "1"
	// tactical := "2"
	// unverified := "4"

	// trade_name := records["trade_name"]
	business_activity_nature := records["business_activity_nature"]

	legalStatusID := strings.TrimSpace(records["business_constitution_group_id"])
	legalStatusValue := GetLegalStatus(legalStatusID)

	var tacticalflag int
	tacticalflag = 0

	glusr_usr_email1 := records["glusr_usr_email"]
	glusr_usr_email_alt1 := records["glusr_usr_email_alt"]

	gst_mobile_number := records["mobile_number"]
	gst_email_id := records["email_id"]

	befisc_mobile := records["gst_challan_mobile_by_befisc"]
	befisc_email := records["gst_challan_email_by_befisc"]

	user := User{
		glusr_usr_email:     glusr_usr_email1,
		glusr_usr_email_alt: glusr_usr_email_alt1,
	}

	var emailattributeID string
	var emailmatch string
	emailMatched := false
	if gst_email_id != "" && gst_email_id != " " {
		emailmatch, emailattributeID = matchEmailID(gst_email_id, user)
		fmt.Println("emailmatched=", emailmatch)
		if emailattributeID == "-1" {
			emailMatched = false
		} else {
			a, err1 := IsAttributeAlreadyVerified(glid, emailattributeID)
			if err != nil {
				logg.AnyError["userverifiedattributeemail"] = err1.Error()
			} else {
				if a {
					emailMatched = true
					apiResponse, err := servapi.UserVerifiedDetailsMM(glid, emailattributeID)
					if err != nil {
						logg.AnyError["UserVerifiedDetailsapierror"] = err1.Error()
					}

					// Extract verification date dynamically
					verificationDate, status := servapi.ExtractUserVerificationDate(apiResponse)
					if status == "Verified" {
						tacticalflag = 1
						logg.TacticalAttributeSource = "email_by_authbridge"
						logg.User_verification_date = verificationDate
					}
				}
			}
		}
	}

	mobileMatched := false
	var attribute_mob string
	var mobilenumber string

	attribute_mob = ""
	mobilenumber = ""

	attributes, err := Pnsapicall(glid, properties.Prop.SERVICES_ENV)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Attributes:", attributes)

		// Example mobile number to match
		mobileToMatch := gst_mobile_number
		if mobileToMatch != "" && mobileToMatch != " " {
			attribute, mobile, found := matchMobileWithAttributes(attributes, mobileToMatch)
			if found {
				muserverified, err1 := IsAttributeAlreadyVerified(glid, attribute)
				if err != nil {
					logg.AnyError["userverifiedattributemobile"] = err1.Error()
				} else {
					if muserverified {
						apiResponse, err := servapi.UserVerifiedDetailsMM(glid, attribute)
						if err != nil {
							logg.AnyError["UserVerifiedDetailsapierror"] = err1.Error()
						}

						// Extract verification date dynamically
						verificationDate, status := servapi.ExtractUserVerificationDate(apiResponse)
						if status == "Verified" {
							tacticalflag = 1
							mobileMatched = true
							attribute_mob = attribute
							mobilenumber = mobile
							logg.TacticalAttributeSource = "mobile_by_authbridge"
							logg.User_verification_date = verificationDate
						}
					}
				}
				// fmt.Printf("Match found: Attribute: %s, Mobile: %s\n", attribute, mobile)
			} else {
				mobileMatched = false
				// fmt.Println("No match found")
			}
		}
	}

	if tacticalflag != 1 && (befisc_mobile != "" || befisc_email != "") {
		if befisc_email != "" {
			emailmatch, emailattributeID = matchEmailID(befisc_email, user)
			fmt.Println("befiscemailmatched=", emailmatch)
			if emailattributeID == "-1" {
				emailMatched = false
			} else {
				a, err1 := IsAttributeAlreadyVerified(glid, emailattributeID)
				if err != nil {
					logg.AnyError["befiscuserverifiedattributeemail"] = err1.Error()
				} else {
					if a {
						emailMatched = true
						apiResponse, err := servapi.UserVerifiedDetailsMM(glid, emailattributeID)
						if err != nil {
							logg.AnyError["UserVerifiedDetailsapierror"] = err1.Error()
						}

						// Extract verification date dynamically
						verificationDate, status := servapi.ExtractUserVerificationDate(apiResponse)
						if status == "Verified" {
							tacticalflag = 1
							logg.TacticalAttributeSource = "email_by_befisc"
							logg.User_verification_date = verificationDate
						}
					}
				}
			}
		}

		if befisc_mobile != "" {
			attribute, mobile, found := matchMobileWithAttributes(attributes, befisc_mobile)
			if found {
				muserverified, err1 := IsAttributeAlreadyVerified(glid, attribute)
				if err != nil {
					logg.AnyError["befiscuserverifiedattributemobile"] = err1.Error()
				} else {
					if muserverified {
						apiResponse, err := servapi.UserVerifiedDetailsMM(glid, attribute)
						if err != nil {
							logg.AnyError["UserVerifiedDetailsapierror"] = err1.Error()
						}

						// Extract verification date dynamically
						verificationDate, status := servapi.ExtractUserVerificationDate(apiResponse)
						if status == "Verified" {
							tacticalflag = 1
							mobileMatched = true
							attribute_mob = attribute
							mobilenumber = mobile
							logg.TacticalAttributeSource = "mobile_by_befisc"
							logg.User_verification_date = verificationDate
						}
					}
				}
			} else {
				mobileMatched = false
			}
		}

	}

	fmt.Println("attribute,mobile,mobilematch,emailmatch,tacticalflag=", attribute_mob, mobilenumber, mobileMatched, emailMatched, tacticalflag)

	if mobileMatched {
		responses[attribute_mob] = mobilenumber
	}
	if emailMatched {
		responses[emailattributeID] = emailmatch
	}

	registration_date := records["registration_date"]

	turnover, _ := formatTurnover(records["annual_turnover_slab"])
	annual_turnover_slab := getShortName(turnover)

	proprieter_name := records["proprieter_name"]
	partner_name := proprieter_name
	core_business_activity_nature := records["core_business_activity_nature"]
	//business_constitution := records["business_constitution"]

	logg.Data = records

	//**********************REJECTION CODE************************

	gstinStatus := strings.Trim(strings.ToUpper(records["gstin_status"]), " ")

	if gstinStatus != "ACTIVE" {

		var rID int = Others
		switch gstinStatus {
		case "CANCELLED":
			rID = CancelledGST
		case "SUSPENDED":
			rID = SuspendedGST
		case "INACTIVE":
			rID = InactiveGST
		case "INVALID":
			rID = InvalidGST
		default:
			rID = Others
		}

		gstVerification := GstVerification{
			Flag:          4,
			Attribute_src: responses,
		}

		data := Data{
			Flag:     autoRejected,
			ReasonID: rID,
			Reason:   reasonsMap[rID],
			Gstdetails: map[string]string{
				"business_constitution":         legalStatusValue,
				"core_business_activity_nature": core_business_activity_nature,
				"proprieter_name":               partner_name,
				"annual_turnover_slab":          annual_turnover_slab,
				"registration_date":             registration_date,
				"business_activity_nature":      business_activity_nature,
			},
			GstVerificationSrc: gstVerification,
		}
		sendResponse(uniqID, w, 200, success, "", nil, data, logg)
		return
	}

	gstState := strings.Trim(records["state_name"], " ")
	glidState := strings.Trim(records["glusr_usr_state"], " ")
	// gstinNumber := records["gstin_number"]

	gstPin := strings.Trim(records["pincode"], " ")
	glidZip := strings.Trim(records["glusr_usr_zip"], " ")

	var stateMismatch bool
	var pincodeMismatch bool

	stateMismatch = !gstmmcontrols.IsSateSame(glidState, gstState)
	pincodeMismatch = !gstmmcontrols.IsPincodeSame(glidZip, gstPin)

	glComp := strings.ToLower(strings.Trim(records["glusr_usr_companyname"], " "))
	gstComp := strings.ToLower(strings.Trim(records["trade_name_replaced"], " "))

	if gstComp == "" || gstComp == "na" || gstComp == "NA" {
		gstComp = strings.ToLower(strings.Trim(records["business_name_replaced"], " "))
	}

	// trade_name_replaced_new := strings.Trim(records["trade_name_replaced"], " ")
	if gstComp == "na" || len(gstComp) == 0 {
		var manresponse GstVerification

		manresponse = GstVerification{
			Flag:          5,
			Attribute_src: responses,
		}

		data := Data{
			Flag: manVerify,
			Gstdetails: map[string]string{
				"business_constitution":         legalStatusValue,
				"core_business_activity_nature": core_business_activity_nature,
				"proprieter_name":               partner_name,
				"annual_turnover_slab":          annual_turnover_slab,
				"registration_date":             registration_date,
				"business_activity_nature":      business_activity_nature,
			},
			GstVerificationSrc: manresponse,
		}
		sendResponse(uniqID, w, 200, success, "", nil, data, logg)
		return
	}
	/*
		if len(gstState) == 0 {
			var manresponse GstVerification
			// Approved:= "AUTO APPROVED"
			Previous_Flag := ""
			// if tacticalflag == 1 {
			// 	manresponse = GstVerification{
			// 		Flag:          2,
			// 		Attribute_src: responses,
			// 	}
			// 	Previous_Flag = "AUTO APPROVED"
			// 	MannualTactical = "AUTO"
			// } else {
			manresponse = GstVerification{
				Flag:          5,
				Attribute_src: responses,
			}
			Previous_Flag = "MANUAL VERIFICATION"
			// }
			data := Data{
				Flag: Previous_Flag,
				Gstdetails: map[string]string{
					"business_constitution":         legalStatusValue,
					"core_business_activity_nature": core_business_activity_nature,
					"proprieter_name":               proprieter_name,
					"annual_turnover_slab":          annual_turnover_slab,
					"registration_date":             registration_date,
					"business_activity_nature":      business_activity_nature,
				},
				GstVerificationSrc: manresponse,
			}
			sendResponse(uniqID, w, 200, success, "", nil, data, logg)
			return
		}
	*/
	// gst6thChar := string(gstinNumber[5])

	//fmt.Println(gstinNumber, ",", gst6thChar, ",", glidState, ",", gstState)
	/*
		if strings.ToUpper(gst6thChar) != "C" && ((strings.Trim(glidState, " ") == "") && (strings.Trim(gstState, " ") != "")) {
			var manresponse GstVerification
			Previous_Flag2 := ""
			// if tacticalflag == 1 {
			// 	manresponse = GstVerification{
			// 		Flag:          2,
			// 		Attribute_src: responses,
			// 	}
			// 	Previous_Flag2 = "AUTO APPROVED"
			// 	MannualTactical = "AUTO"
			// } else {
			manresponse = GstVerification{
				Flag:          5,
				Attribute_src: responses,
			}
			Previous_Flag2 = "MANUAL VERIFICATION"
			// }
			data := Data{
				Flag: Previous_Flag2,
				Gstdetails: map[string]string{
					"business_constitution":         legalStatusValue,
					"core_business_activity_nature": core_business_activity_nature,
					"proprieter_name":               proprieter_name,
					"annual_turnover_slab":          annual_turnover_slab,
					"registration_date":             registration_date,
					"business_activity_nature":      business_activity_nature,
				},
				GstVerificationSrc: manresponse,
			}
			sendResponse(uniqID, w, 200, success, "", nil, data, logg)
			return
		}
	*/

	if stateMismatch && pincodeMismatch {

		rID := StateMismatch

		gstVerification := GstVerification{
			Flag:          4,
			Attribute_src: responses,
		}
		data := Data{
			Flag:     autoRejected,
			ReasonID: rID,
			Reason:   reasonsMap[rID],
			Gstdetails: map[string]string{
				"business_constitution":         legalStatusValue,
				"core_business_activity_nature": core_business_activity_nature,
				"proprieter_name":               partner_name,
				"annual_turnover_slab":          annual_turnover_slab,
				"registration_date":             registration_date,
				"business_activity_nature":      business_activity_nature,
			},
			GstVerificationSrc: gstVerification,
		}
		sendResponse(uniqID, w, 200, success, "", nil, data, logg)
		return
	}

	score := jaroWinkler(glComp, gstComp)

	if score < 0.70 {
		rID := Others

		gstVerification := GstVerification{
			Flag:          4,
			Attribute_src: responses,
		}
		data := Data{
			Flag:     autoRejected,
			ReasonID: rID,
			Reason:   reasonsMap[rID],
			Gstdetails: map[string]string{
				"business_constitution":         legalStatusValue,
				"core_business_activity_nature": core_business_activity_nature,
				"proprieter_name":               partner_name,
				"annual_turnover_slab":          annual_turnover_slab,
				"registration_date":             registration_date,
				"business_activity_nature":      business_activity_nature,
			},
			GstVerificationSrc: gstVerification,
		}
		sendResponse(uniqID, w, 200, success, "", nil, data, logg)
		return

	}

	if glComp == "" || glComp == "na" || gstComp == "" || gstComp == "na" || (score >= 0.70 && score < 1.0 && tacticalflag != 1) {
		var manresponse GstVerification
		Previous_Flag3 := ""
		// if tacticalflag == 1 {
		// 	manresponse = GstVerification{
		// 		Flag:          2,
		// 		Attribute_src: responses,
		// 	}
		// 	Previous_Flag3 = "AUTO APPROVED"
		// 	MannualTactical = "AUTO"
		// } else {
		manresponse = GstVerification{
			Flag:          5,
			Attribute_src: responses,
		}
		Previous_Flag3 = "MANUAL VERIFICATION"
		// }
		data := Data{
			Flag: Previous_Flag3,
			Gstdetails: map[string]string{
				"business_constitution":         legalStatusValue,
				"core_business_activity_nature": core_business_activity_nature,
				"proprieter_name":               partner_name,
				"annual_turnover_slab":          annual_turnover_slab,
				"registration_date":             registration_date,
				"business_activity_nature":      business_activity_nature,
			},
			GstVerificationSrc: manresponse,
		}
		sendResponse(uniqID, w, 200, success, "", nil, data, logg)
		return
	}

	if score >= 0.7 && tacticalflag == 1 {
		MannualTactical = "AUTO"
	}

	//**********************AUTO APPROVAL CODE************************

	var buckType, bucketName string
	_, logg.ScoreDetails = gstmmcontrols.MatchMakingScore(records)
	buckType, bucketName = gstmmcontrols.LogicAUTO(logg.ScoreDetails)

	var bucketflag string
	bucketflag = "stage0"
	//if bucketname is not assigned then I will calculate the stage-1 core and stage-1 bucket
	if bucketName == "" {
		_, logg.ScoreDetailsStage1 = gstmmcontrols.MatchMakingScoreStage1(records, logg.ScoreDetails)
		buckType, bucketName = gstmmcontrols.GetBucketAUTO(logg.ScoreDetailsStage1, logg.Request.GST, records["glusr_usr_companyname"])
		bucketflag = "stage1"
	}
	if buckType == "AUTO" || MannualTactical == "AUTO" {
		// rID := DispoWiseReson[bucketName]

		var rID int
		// Check if MannualTactical is AUTO, set hardcoded values
		if MannualTactical == "AUTO" {
			bucketName = "AA1"
		}

		rID = DispoWiseReson[bucketName]

		// Started
		//commented for dev-users-apis data not working on dev

		m := map[string]string{
			"USR_ID":                  glid,
			"VALIDATION_KEY":          BIValidationKeyFromSOA,
			"UPDATEDBY":               screenName + "(BI)",
			"UPDATEDUSING":            screenName,
			"HIST_COMMENTS":           reasonsMap[rID],
			"IP":                      RemoteHost,
			"IP_COUNTRY":              "INDIA",
			"AK":                      constants.ServerAK,
			"DISABLE_GST_RESTRICTION": "1",
		}
		//add the changes from here

		// slices that will feed VerifyGlidAllAttr
		ids, vals := []string{}, []string{}

		//  ⬇︎  NEW helper — only stores Attribute-ID + value
		addVerify := func(col, val string) {
			if val == "" {
				return
			}
			ids = append(ids, attrID[col])
			vals = append(vals, val)
		}

		gst := records["gstin_number"]

		tradeName := strings.ToLower(strings.TrimSpace(records["trade_name"]))

		tradeName_replaced := strings.ToLower(strings.Trim(records["trade_name_replaced"], " "))

		if tradeName_replaced == "" || tradeName_replaced == "na" {
			tradeName = strings.ToLower(strings.TrimSpace(records["business_name"]))
		}

		businessName := strings.ToLower(strings.TrimSpace(records["business_name"]))

		proprieter_name := strings.ToLower(strings.TrimSpace(records["proprieter_name"]))
		// Find the first occurrence of a comma
		commaIndex := strings.Index(proprieter_name, ",")
		// Check if a comma is present
		if commaIndex != -1 {
			// Get the substring before the first comma
			proprieter_name = strings.TrimSpace(proprieter_name[:commaIndex])
		} else {
			// No comma found; keep the whole string
			proprieter_name = strings.TrimSpace(proprieter_name)
		}

		//address related columns
		dno := strings.TrimSpace(records["door_number"])
		bn := strings.TrimSpace(records["building_name"])
		streetname := strings.TrimSpace(records["street"])
		loc := strings.TrimSpace(records["location"])
		flno := strings.TrimSpace(records["floor_number"])

		gstState := strings.ToLower(strings.TrimSpace(records["state_name"]))
		gstPincode := strings.TrimSpace(records["pincode"])

		landmark := strings.TrimSpace(records["landmark"])
		locality := strings.TrimSpace(records["locality"])

		_, compName, add1, add2, state, pincode, cFirstName, _, first_name, _, listing_status, err := workerlogic.IsGlidFree(glid)
		if err != nil {
			logg.AnyError["DetailsError"] = err.Error()
		}
		state = strings.ToLower(strings.TrimSpace(state))
		fmt.Println("GST State: ", gstState)
		fmt.Println("Gl State: ", state)
		modifiedCompName := utils.TradeNameNewFormattingLogic(tradeName, compName)

		citymap, e := workerlogic.CallcityFetch(gstPincode)

		if e != nil {
			logg.AnyError["CityFetch"] = e.Error()
			citymap = map[string]interface{}{}
		}

		getStr := func(m map[string]interface{}, key string) string {
			if v, ok := m[key]; ok {
				if s, ok := v.(string); ok {
					return s
				}
			}
			return ""
		}

		glCity := getStr(citymap, "city_name")
		glDist := getStr(citymap, "district_name")
		cityID := getStr(citymap, "city_id")
		state_Loc := getStr(citymap, "state_name")

		//fmt.Println(cFirstName,cLastName,"Dev-Name",string(gst[5]))
		// var flag0, flag1, flag2, flag3, flag4a, flag5a, flag6, flag7, flag8 bool
		// flag0 = false
		// flag1 = false
		// flag2 = false
		// flag3 = false
		// flag4a = false
		// flag5a = false
		// flag6 = false
		// flag7 = false
		// flag8 = false

		if modifiedCompName != compName {
			// flag0 = true
			m["COMPANYNAME"] = modifiedCompName
		}

		addVerify("COMPANYNAME", modifiedCompName)

		// flag1 = true

		// Primaryaddress := dno + "," + flno + "," + bn + "," + streetname + "," + loc

		// Extract city_name from citymap and store it in Gluser_City_Name
		// Gluser_City_Name := citymap["city_name"].(string)
		// Gluser_Dist_Name := citymap["district_name"].(string)

		Primaryaddress := CreatePrimaryAddress(dno, flno, bn, streetname, loc, glCity, glDist, locality, landmark)
		Secondaryaddress := CreateSecondaryAddress(add1, add2, pincode, records["glusr_usr_city"])

		var addscore float64

		if bucketflag == "stage0" {
			addscore = logg.ScoreDetails.AddressScore
		} else {
			addscore = logg.ScoreDetailsStage1.AddressScore
		}

		var gluser_address_Line2 string
		var gluser_address_Line1 string
		gluser_address_Line1, gluser_address_Line2 = SplitAddress(Primaryaddress, 100)

		if addscore <= 0.4 {
			var section string
			section = "Corporate Office"
			var secondaryaddressadd1 string
			var secondaryaddressadd2 string
			//if secondary > 200 abreak it two and passed to details api
			secondaryaddressadd1, secondaryaddressadd2 = SplitAddress(Secondaryaddress, 200)

			if len(secondaryaddressadd1) >= 1 {
				//called details api and pass secondaryaddress parameter
				err := UpdateAdditionAddressDetails("PROD", glid, section, secondaryaddressadd1, secondaryaddressadd2, "GST Auto Approval")
				if err != nil {
					fmt.Println("Error updating address details:", err)
					logg.AnyError["DetailsServiceError"] = err.Error()
				}

				// gluser_address_Line1, gluser_address_Line2 = SplitAddress(Primaryaddress, 100)
			}
		}

		if len(gluser_address_Line1) >= 1 {
			m["ADD1"] = gluser_address_Line1
			m["ADD2"] = gluser_address_Line2
			// flag4a = true
		}

		addVerify("ADD1", gluser_address_Line1)
		addVerify("ADD2", gluser_address_Line2)

		if string(gst[5]) == "P" {
			if businessName != "" {
				// flag3 = true
				firstName, lastName := utils.Convert(businessName)
				m["CFIRSTNAME"] = firstName
				m["CLASTNAME"] = lastName
				addVerify("CFIRSTNAME", firstName)
				addVerify("CLASTNAME", lastName)

			}

			if businessName != "" && first_name == "" {
				// flag6 = true
				firstName, lastName := utils.Convert(businessName)
				m["FIRSTNAME"] = firstName
				m["LASTNAME"] = lastName

				addVerify("FIRSTNAME", firstName)
				addVerify("LASTNAME", lastName)
			}
		}

		if string(gst[5]) != "P" {
			if proprieter_name != "" && cFirstName == "" {
				// flag3 = true
				firstName, lastName := utils.Convert(proprieter_name)
				m["CFIRSTNAME"] = firstName
				m["CLASTNAME"] = lastName

				addVerify("CFIRSTNAME", firstName)
				addVerify("CLASTNAME", lastName)
			}

			if proprieter_name != "" && first_name == "" {
				// flag6 = true
				firstName, lastName := utils.Convert(proprieter_name)
				m["FIRSTNAME"] = firstName
				m["LASTNAME"] = lastName

				addVerify("CFIRSTNAME", firstName)
				addVerify("CLASTNAME", lastName)
			}
		}

		listing_status = strings.ToLower(strings.TrimSpace(listing_status))

		if strings.ToUpper(listing_status) == "NFL" {
			// flag7 = true
			m["LISTING_STATUS"] = "TFL"
			m["LISTING_REASON"] = "User Activity"
		}
		//logs

		if len(gstPincode) == 6 {
			// flag8 = true

			if gstPincode != pincode {
				m["ZIP"] = gstPincode
			}

			addVerify("ZIP", gstPincode)

			if cityID != "" || glCity != "" { // ← any one is present
				m["FK_GL_CITY_ID"] = cityID // (may be empty, that’s OK)
				m["CITY"] = glCity

				addVerify("FK_GL_CITY_ID", cityID) // verify both
				addVerify("CITY", glCity)
			}

			if state_Loc != "" {
				m["STATE"] = state_Loc
				addVerify("STATE", state_Loc)
			}
		}

		if len(ids) != 0 { // we have something to verify

			// update only if m has extra keys beyond the 9 fixed ones
			if len(m) > 9 {
				if err := workerlogic.UpdateComp(properties.Prop.SERVICES_ENV, m); err != nil {
					logg.AnyError["UpdateComp"] = err.Error()
					// optional: return here if you don’t want to verify on update failure
				}
			}

			// verify every attribute we collected
			if err := VerifyGlidAllAttr(
				properties.Prop.SERVICES_ENV,
				glid,
				strings.Join(ids, ","),
				strings.Join(vals, "##"),
				"Trade Name + City/Zip Matching",
			); err != nil {
				logg.AnyError["VerifyGlidAllAttr"] = err.Error()
			}
		}

		// testlog := make(map[string]bool)
		// testlog["flag0"] = flag0
		// testlog["flag1"] = flag1
		// testlog["flag2"] = flag2
		// testlog["flag3"] = flag3
		// testlog["flag4a"] = flag4a
		// testlog["flag5a"] = flag5a
		// testlog["flag6"] = flag6
		// testlog["flag7"] = flag7
		// testlog["flag8"] = flag8
		// for key, value := range testlog {
		// 	logg.UpdateFlags[key] = value
		// }

		// if flag0 || flag1 || flag2 || flag3 || flag6 || flag7 || flag8 {

		// 	if !flag0 {
		// 		//check verified or not
		// 		// if not verified , we have to verify it
		// 		// else leave it
		// 		verified_status, err_compVerify := workerlogic.IsCompanyAlreadyVerified(glid)
		// 		if err_compVerify != nil {
		// 			err_verify := workerlogic.VerifyGlidAttr(properties.Prop.SERVICES_ENV, glid, "111", modifiedCompName, "Trade Name + City/Zip Matching")
		// 			if err_verify != nil {
		// 				logg.AnyError["UserVerifyServiceError"] = err_verify.Error()
		// 			}
		// 		} else {
		// 			if !verified_status {
		// 				err_verify := workerlogic.VerifyGlidAttr(properties.Prop.SERVICES_ENV, glid, "111", modifiedCompName, "Trade Name + City/Zip Matching")
		// 				if err_verify != nil {
		// 					logg.AnyError["UserVerifyServiceError"] = err_verify.Error()
		// 				}
		// 			}
		// 		}

		// 		//update
		// 		err_update := workerlogic.UpdateComp(properties.Prop.SERVICES_ENV, m)
		// 		if err_update != nil {
		// 			logg.AnyError["UserUpdationServiceError"] = err_update.Error()
		// 		}

		// 	} else {
		// 		// if already verified or not
		// 		verified_status, err_compVerify := workerlogic.IsCompanyAlreadyVerified(glid)
		// 		if err_compVerify != nil {
		// 			//unverify
		// 			err_unverify := workerlogic.UnVerifyGlidAttr(properties.Prop.SERVICES_ENV, glid, "111", compName, "Trade Name + City/Zip Matching")
		// 			if err_unverify != nil {
		// 				logg.AnyError["UserUnverifyServiceError"] = err_unverify.Error()
		// 			}
		// 			//update
		// 			err_update := workerlogic.UpdateComp(properties.Prop.SERVICES_ENV, m)
		// 			if err_update != nil {
		// 				logg.AnyError["UserUpdationServiceError"] = err_update.Error()
		// 			} else {
		// 				//verify
		// 				err_verify := workerlogic.VerifyGlidAttr(properties.Prop.SERVICES_ENV, glid, "111", modifiedCompName, "Trade Name + City/Zip Matching")
		// 				if err_verify != nil {
		// 					logg.AnyError["UserVerifyServiceError"] = err_verify.Error()
		// 				}
		// 			}
		// 		} else {
		// 			if verified_status {
		// 				//unverify
		// 				err_unverify := workerlogic.UnVerifyGlidAttr(properties.Prop.SERVICES_ENV, glid, "111", compName, "Trade Name + City/Zip Matching")
		// 				if err_unverify != nil {
		// 					logg.AnyError["UserUnverifyServiceError"] = err_unverify.Error()
		// 				}
		// 				//update
		// 				err_update := workerlogic.UpdateComp(properties.Prop.SERVICES_ENV, m)
		// 				if err_update != nil {
		// 					logg.AnyError["UserUpdationServiceError"] = err_update.Error()
		// 				} else {
		// 					//verify
		// 					err_verify := workerlogic.VerifyGlidAttr(properties.Prop.SERVICES_ENV, glid, "111", modifiedCompName, "Trade Name + City/Zip Matching")
		// 					if err_verify != nil {
		// 						logg.AnyError["UserVerifyServiceError"] = err_verify.Error()
		// 					}
		// 				}
		// 			} else {
		// 				//update
		// 				err_update := workerlogic.UpdateComp(properties.Prop.SERVICES_ENV, m)
		// 				if err_update != nil {
		// 					logg.AnyError["UserUpdationServiceError"] = err_update.Error()
		// 				} else {
		// 					//verify
		// 					err_verify := workerlogic.VerifyGlidAttr(properties.Prop.SERVICES_ENV, glid, "111", modifiedCompName, "Trade Name + City/Zip Matching")
		// 					if err_verify != nil {
		// 						logg.AnyError["UserVerifyServiceError"] = err_verify.Error()
		// 					}
		// 				}
		// 			}
		// 		}

		// 	}
		// 	// err_unverify := workerlogic.UnVerifyGlidAttr(properties.Prop.SERVICES_ENV, glid, "111", compName, "Trade Name + City/Zip Matching")
		// 	// if err_unverify != nil {
		// 	// 	logg.AnyError["UserUnverifyServiceError"] = err_unverify.Error()
		// 	// }

		// 	// err = workerlogic.UpdateComp(properties.Prop.SERVICES_ENV, m)
		// 	// if err != nil {
		// 	// 	logg.AnyError["UserUpdationServiceError"] = err.Error()
		// 	// }

		// 	// err_verify := workerlogic.VerifyGlidAttr(properties.Prop.SERVICES_ENV, glid, "111", modifiedCompName, "Trade Name + City/Zip Matching")
		// 	// if err_verify != nil {
		// 	// 	logg.AnyError["UserVerifyServiceError"] = err_verify.Error()
		// 	// }
		// }

		// ENDED
		//commented for dev-users-apis data not working on dev

		//to here
		var autoresponse GstVerification
		if tacticalflag == 1 {
			autoresponse = GstVerification{
				Flag:          2,
				Attribute_src: responses,
			}
		} else {
			autoresponse = GstVerification{
				Flag:          1,
				Attribute_src: responses,
			}
		}

		data := Data{
			Flag:       autoApproved,
			ReasonID:   rID,
			Reason:     reasonsMap[rID],
			BucketName: bucketName,
			Gstdetails: map[string]string{
				"business_constitution":         legalStatusValue,
				"core_business_activity_nature": core_business_activity_nature,
				"proprieter_name":               partner_name,
				"annual_turnover_slab":          annual_turnover_slab,
				"registration_date":             registration_date,
				"business_activity_nature":      business_activity_nature,
			},
			GstVerificationSrc: autoresponse,
		}

		//fmt.Println("a")
		sendResponse(uniqID, w, 200, success, "", nil, data, logg)
		return
	}

	//**********************MAN VERIFY************************
	var manresponse GstVerification
	Previous_Flag4 := ""
	// if tacticalflag == 1 {
	// 	manresponse = GstVerification{
	// 		Flag:          2,
	// 		Attribute_src: responses,
	// 	}
	// 	Previous_Flag4 = "AUTO APPROVED"
	// 	MannualTactical = "AUTO"
	// } else {
	manresponse = GstVerification{
		Flag:          5,
		Attribute_src: responses,
	}
	Previous_Flag4 = "MANUAL VERIFICATION"
	// }
	data := Data{
		Flag: Previous_Flag4,
		Gstdetails: map[string]string{
			"business_constitution":         legalStatusValue,
			"core_business_activity_nature": core_business_activity_nature,
			"proprieter_name":               partner_name,
			"annual_turnover_slab":          annual_turnover_slab,
			"registration_date":             registration_date,
			"business_activity_nature":      business_activity_nature,
		},
		GstVerificationSrc: manresponse,
	}
	sendResponse(uniqID, w, 200, success, "", nil, data, logg)
	return

}

func sendResponse(uniqID string, w http.ResponseWriter, httpcode int, status string, errorMsg string, err error, body Data, logg Logg) {

	w.Header().Set("Content-Type", "application/json")

	logg.Response = Res{
		Code:   httpcode,
		Error:  errorMsg,
		Status: status,
		Body:   body,
		UniqID: uniqID,
	}

	if err != nil {
		logg.AnyError[errorMsg] = err.Error()
	}

	json.NewEncoder(w).Encode(logg.Response)
	//fmt.Println("b")
	logg.RequestEndValue = utils.GetTimeInNanoSeconds()
	logg.ResponseTime = (logg.RequestEndValue - logg.RequestStartValue) / 1000000
	logg.RequestStartValue = 0
	logg.RequestEndValue = 0

	writeLog2(logg)

	LogToNewFile(logg)

	return
}

func getShortName(turnover string) string {
	shortNameMap := map[string]string{
		"0 to 40 lakhs":       "0 - 40 L",
		"40 lakhs to 1.5 Cr.": "40 L - 1.5 Cr",
		"40 lakhs to 1.5 Cr":  "40 L - 1.5 Cr",
		"1.5 Cr. to 5 Cr.":    "1.5 - 5 Cr",
		"1.5 Cr to 5 Cr":      "1.5 - 5 Cr",
		"5 Cr. to 25 Cr.":     "5 - 25 Cr",
		"5 Cr to 25 Cr":       "5 - 25 Cr",
		"25 Cr. to 100 Cr.":   "25 - 100 Cr",
		"25 Cr to 100 Cr":     "25 - 100 Cr",
		"100 Cr. to 500 Cr.":  "100 - 500 Cr",
		"100 Cr to 500 Cr":    "100 - 500 Cr",
		"500 Cr. and above":   "> 500 Cr",
		"500 Cr and above":    "> 500 Cr",
		"NA":                  "NA",
		"":                    "NA",
	}

	shortName, ok := shortNameMap[turnover]
	if !ok {
		return "NA"
	}

	return shortName
}

// formatTurnover formats the turnover information.
func formatTurnover(input string) (turnoverValue, turnoverYear string) {
	// Formatting logic
	replacements := []struct {
		old string
		new string
	}{
		{"<br/>", ""},
		{"Slab:", ""},
		{"Rs.", ""},
	}

	for _, r := range replacements {
		input = strings.ReplaceAll(input, r.old, r.new)
	}
	input = strings.TrimSpace(input)

	// Extract turnover and year
	if idx := strings.Index(input, "(For FY"); idx != -1 {
		turnoverValue = strings.TrimSpace(input[:idx])
		turnoverYear = strings.TrimSpace(input[idx+len("(For FY"):])

		if len(turnoverYear) > 0 && (turnoverYear[len(turnoverYear)-1] == ')' || utils.IsSpecialOrNonNumericASCII(rune(turnoverYear[len(turnoverYear)-1]))) {
			turnoverYear = turnoverYear[:len(turnoverYear)-1]
			turnoverYear = strings.TrimSpace(turnoverYear)
		}

		if inidx := strings.Index(turnoverYear, "-"); inidx != -1 {
			turnoverYear = turnoverYear[:inidx+1] + turnoverYear[inidx+3:]
		}
	} else if idx2 := strings.LastIndex(input, "("); idx2 != -1 {
		turnoverValue = strings.TrimSpace(input[:idx2])
		turnoverYear = strings.TrimSpace(input[idx2+1:])

		if len(turnoverYear) > 0 &&
			(turnoverYear[len(turnoverYear)-1] == ')' ||
				utils.IsSpecialOrNonNumericASCII(rune(turnoverYear[len(turnoverYear)-1]))) {
			turnoverYear = strings.TrimSpace(turnoverYear[:len(turnoverYear)-1])
		}
		// normalize “2020-2021” → “2020-21”
		if inidx := strings.Index(turnoverYear, "-"); inidx != -1 {
			turnoverYear = turnoverYear[:inidx+1] + turnoverYear[inidx+3:]
		}

	} else {
		turnoverValue = input
	}

	return turnoverValue, turnoverYear
}

// GetLegalStatus returns the legal status string based on the provided legalStatusID
func GetLegalStatus(legalStatusID string) string {
	legalStatusMap := map[string]string{
		"1924": "Proprietorship",
		"1925": "Partnership",
		"1926": "Limited Company",
		"1927": "Others",
	}

	if legalStatusID == "" {
		return "Others"
	}

	if status, exists := legalStatusMap[legalStatusID]; exists {
		return status
	}
	return "Unknown"
}

// writeLog2 ...
func writeLog2(logg Logg) {

	logsDir := serviceLogPath + utils.TodayDir()
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		e := os.MkdirAll(logsDir, os.ModePerm)
		if e != nil {
			fmt.Println(e)
			return
		}
	}

	logsDir += "/" + logFileName

	jsonLog, _ := json.Marshal(logg)
	//fmt.Println(string(jsonLog))
	f, _ := os.OpenFile(logsDir, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()

	f.WriteString("\n" + string(jsonLog))
	return
}

func LogToNewFile(logg Logg) {
	logEntry := CreateLogEntry(logg)

	logsDir := properties.Prop.LOG_MASTERINDIA + utils.TodayDir()
	// logsDir := serviceLogPath + utils.TodayDir()
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		e := os.MkdirAll(logsDir, os.ModePerm)
		if e != nil {
			fmt.Println(e)
			return
		}
	}

	logsDir += "/masterindia_wrapper_worker_logs.json"

	fmt.Println(logsDir)

	// Convert log entry to JSON
	jsonLog, err := json.Marshal(logEntry)
	if err != nil {
		fmt.Println("Error marshaling log entry:", err)
		return
	}
	// Lock the mutex before writing to the file
	logFileMutex.Lock()
	defer logFileMutex.Unlock()

	// Open the log file and append the new log entry
	f, err := os.OpenFile(logsDir, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer f.Close()

	_, err = f.WriteString("\n" + string(jsonLog))
	if err != nil {
		fmt.Println("Error writing to log file:", err)
	}
}
