package tan_verification

import (
	"encoding/json"
	"fmt"
	// servapi "mm/api/servapi"
	"mm/properties"
	"mm/utils"
	"os"
	"strconv"

	//"errors"
	"io/ioutil"
	"mm/components/constants"
	database "mm/components/database"
	"net/http"
	"time"
)

// S3Log ...
type S3Log struct {
	Modid              string                 `json:"Modid,omitempty"`
	APIName            string                 `json:"APIName,omitempty"`
	Tanid              string                 `json:"Tanid,omitempty"`
	Glid               string                 `json:"Glid,omitempty"`
	RqstTime           string                 `json:"RqstTime,omitempty"`
	Result             map[string]interface{} `json:"Result,omitempty"`
	ResponseTime       float64                `json:"ResponseTime,omitempty"`
	ResponseTime_Float float64                `json:"ResponseTime_Float,omitempty"` // float64 type
	ExtApihit          string                 `json:"ExtApihit,omitempty"`
	TicketCreated      string                 `json:"TicketCreated,omitempty"`
	Verified           string                 `json:"Verified,omitempty"`
	ApiResponsetime    map[string]float64     `json:"ApiResponsetime,omitempty"`
}

var S3log S3Log
var Tanidd string
var Companyid string

// S3log.Result = make(map[string]interface{})
// SubcriberHandler ...
func SubcriberHandler(data string) error {
	//      fmt.Println("SubcriberHandler started")
	wr := WorkRequest{}
	err := json.Unmarshal([]byte(data), &wr)
	if err != nil {
		return err
	}

	//fmt.Println("Worker Request", wr)
	if wr.APIName == "tanAPI" {
		tann(wr)
	}
	// fmt.Println("SubcriberHandler ended:")
	return nil
}

// tanverification ...
func tann(work WorkRequest) {
	// fmt.Println("tann  function started:")
	//      var S3log S3Log
	S3log.APIName = work.APIName
	//user := work.APIUserName
	//credential := utils.GetCred(user)

	//S3log.APIUserID = credential["username"]
	S3log.Tanid = work.Tanid
	// S3log.APIHit = ""
	S3log.Glid = work.Glid
	S3log.RqstTime = work.RqstTime
	S3log.Result = make(map[string]interface{})
	S3log.ApiResponsetime = make(map[string]float64)
	S3log.Modid = work.Modid
	S3log.ExtApihit = "N"
	S3log.TicketCreated = "N"
	S3log.Verified = "N"

	TANID := work.Tanid
	GLID := work.Glid
	//fmt.Println("calling tann verification with TANID - GLID:", TANID, " ", GLID)
	start := utils.GetTimeInNanoSeconds()
	tannverification(TANID, GLID)
	// Tanidd = ""
	// Companyid = ""
	end := utils.GetTimeInNanoSeconds()
	S3log.ResponseTime = (end - start) / 1000000
	S3log.ResponseTime_Float = (end - start) / 1000000
	//fmt.Println("tann function is ending")
	Write2S3(&S3log)
	// return
}

func tannverification(TAN_ID string, GLUSR string) {
	// fmt.Println("calling tannverification function code base")

	// empid := "86906"

	// t, err := servapi.GetTAkFromMerpLogin(empid)

	// if err != nil {
	// 	S3log.Result["akgenerationerr"] = err.Error()
	// 	fmt.Printf("Error: %v\n", err)
	// 	return
	// }
    flagvendorapi := 0
	t := constants.GSTServerAK8s

	companyidd, comperr := companyy(GLUSR, t)
	if comperr != nil {
		//S3log.TicketCreated = "Y"
		S3log.Result["companyid api error"] = comperr.Error()
		Write2S3(&S3log)
		return
	}

	Tanidd = TAN_ID
	Companyid = companyidd

	//companyid, err := strconv.ParseInt(companyidd, 10, 64)
	//fmt.Println("companyid ",companyid)
	GLUSR_ID, err := strconv.ParseInt(GLUSR, 10, 64)
	// fmt.Println("GLUSR_ID:: ", GLUSR_ID)

	if err != nil {
		//S3log.TicketCreated = "Y"
		CreateTicket(Tanidd, "", "glid is not in correct format", Companyid)
		S3log.Result["glid is not in correct format"] = GLUSR_ID
		Write2S3(&S3log)
		return
	}
	var panicErr error
	defer func() {
		if panicCheck := recover(); panicCheck != nil {
			panicErr = panicCheck.(error)
			fmt.Println(panicErr, "Panic Error Encountered !!!")
			S3log.Result["Panic Error Encountered !!!"] = panicErr.Error()
			Write2S3(&S3log)
			return
		}
	}()

	mainConnection, err := database.GetDatabaseConnection("approvalPG")
	//fmt.Println("mainConnection",mainConnection,"error 1st",err)
	if err != nil {
		//fmt.Println("Could not start the main database because of following err " + err.Error() + " ...!!!")
		//S3log.TicketCreated = "Y"
		CreateTicket(Tanidd, "", "Could not start the approvalPG database ", Companyid)
		S3log.Result["Could not start the approvalPG  database because of following err  "] = err.Error()
		Write2S3(&S3log)
		return
	}

	// defer func() {

	// }()

	var count1 int64
	query := `Select count(*) CNT from iil_tan_master_data where tan_no = '` + TAN_ID + `'`
	params := make([]interface{}, 0)
	result, err := database.SelectQuerySql(mainConnection, query, params)
	// fmt.Println(result)
	if err != nil {
		//fmt.Println("Could not connect the approval database because of following err " + err.Error() + " ...!!!")
		//S3log.TicketCreated = "Y"
		CreateTicket(Tanidd, "", "Could not connect the approvalPG database ", Companyid)
		S3log.Result["Could not connect the approval database because of following err  "] = err.Error()
		Write2S3(&S3log)
		return
	}
	if result != nil {
		if queryData, dataExists := result["queryData"]; dataExists {
			if queryData != nil {
				resultData := queryData.([]interface{})
				if len(resultData) > 0 {
					resultFirst := resultData[0].(map[string]interface{})

					count, countExists := resultFirst["cnt"]
					if countExists && count != nil {

						count1 = count.(int64)
					}
				}
			}
		}
	}
	//fmt.Println("COUNT IS " , count1)
	if count1 == 0 && flagvendorapi == 1 {
		//fmt.Println("API BLOCK")
		var tan TanData
		readData, msg, dataTan := tanVerification(TAN_ID, GLUSR_ID, 0, companyidd)
		// fmt.Println(readData, msg, dataTan)
		if msg == "XXXX" {
			return
		}
		_ = msg
		_ = dataTan

		tan.GLID = strconv.FormatInt(GLUSR_ID, 10)
		tan.TANNO = TAN_ID
		//var ticket int64
		//verified := ""
		PAYMENT_TYPE := "NOT VALID OUTPUT"
		if readData["nameOrgn"] != nil {

			tan.COMPANY_NAME = (readData["nameOrgn"]).(string)
		}
		if readData["addLine1"] != nil {
			if readData["addLine1"] != "ADD_LINE_1" {
				tan.ADDRESS = (readData["addLine1"]).(string) + ` `
			}
		}
		if readData["addLine2"] != nil {
			tan.ADDRESS += (readData["addLine2"]).(string) + ` `
		}
		if readData["addLine3"] != nil {
			tan.ADDRESS += (readData["addLine3"]).(string) + ` `
		}
		if readData["addLine4"] != nil {
			tan.ADDRESS += (readData["addLine4"]).(string) + ` `
		}
		if readData["addLine5"] != nil {
			tan.ADDRESS += (readData["addLine5"]).(string) + ` `
		}
		if readData["nameLast"] != nil {
			tan.LASTNAME = (readData["nameLast"]).(string) + ` `
		}
		if readData["nameFirst"] != nil {
			tan.FIRSTNAME = (readData["nameFirst"]).(string) + ` `
		}
		if readData["nameMid"] != nil {
			tan.MIDNAME = (readData["nameMid"]).(string) + ` `
		}

		if readData["pin"] != nil {
			tan.PINCODE = fmt.Sprint((readData["pin"]).(float64))
		}
		if readData["stateCd"] != nil {
			tan.STATECD = fmt.Sprint((readData["stateCd"]).(float64))
		}
		S3log.Result["TAN_MASTER_TABLE_DATA"] = tan
		if tan.COMPANY_NAME != "" {
			PAYMENT_TYPE = "NOT VERIFIED"
			percent, _ := Similarity(tan.GLID, tan.COMPANY_NAME)
			if percent >= 75 {
				PAYMENT_TYPE = "VERIFIED WITH COMPANY NAME"
			}
			if percent < 75 {
				//percent, _ = SimilarityCEO(tan.GLID, tan.COMPANY_NAME)
				percent1, _ := SimilarityCEO(tan.GLID, tan.COMPANY_NAME)
				percent2, _ := SimilarityCEO(tan.GLID, tan.FIRSTNAME+" "+tan.MIDNAME+" "+tan.LASTNAME)
				if percent1 >= percent2 {
					percent = percent1
				} else {
					percent = percent2
				}
			}
			if PAYMENT_TYPE == "NOT VERIFIED" && percent >= 60 {
				percentPin, _ := SimilarityPIN(tan.GLID, tan.PINCODE)
				if percentPin == 100 {
					PAYMENT_TYPE = "VERIFIED WITH CEO NAME"
				} else {
					statecodestatus, _ := SimilarityStatecd(tan.GLID, tan.STATECD)
					if statecodestatus == true {
						PAYMENT_TYPE = "VERIFIED WITH CEO NAME"
					}
				}

			}
			if percent < 60 {
				//percent, _ = SimilarityCONTACT(tan.GLID, tan.COMPANY_NAME)
				percent1, _ := SimilarityCONTACT(tan.GLID, tan.COMPANY_NAME)
				percent2, _ := SimilarityCONTACT(tan.GLID, tan.FIRSTNAME+" "+tan.LASTNAME)
				if percent1 >= percent2 {
					percent = percent1
				} else {
					percent = percent2
				}
			}
			if PAYMENT_TYPE == "NOT VERIFIED" && percent >= 60 {
				percentPin, _ := SimilarityPIN(tan.GLID, tan.PINCODE)
				if percentPin == 100 {
					PAYMENT_TYPE = "VERIFIED WITH CEO NAME"
				} else {
					statecodestatus, _ := SimilarityStatecd(tan.GLID, tan.STATECD)
					if statecodestatus == true {
						PAYMENT_TYPE = "VERIFIED WITH CEO NAME"
					}
				}
			}

			S3log.Result[" PAYMENT_TYPE "] = PAYMENT_TYPE

			if PAYMENT_TYPE == "NOT VERIFIED" {
				CreateNewTicket(companyidd, tan)
				//fmt.Println(ticket) //checkonce
				//S3log.Result[" PAYMENT_TYPE "] = PAYMENT_TYPE
				//Write2S3(&S3log)
				//fmt.Println("PAYMENT_TYPE",PAYMENT_TYPE)
			} else {
				//S3log.Verified = "Y"
				//verified = "verified"
				status := TanVerifyApi(tan)
				S3log.Result["verificationstatus"] = status
				//fmt.Println("STATUSSSS",status)
				//fmt.Println(verified) //check once

			}
		}

		err1 := InsertInDB(tan)
		if err1 != nil {
			// fmt.Println("error fetching inserting data", err)
			S3log.Result["error fetching inserting data  "] = err.Error()
			Write2S3(&S3log)
			return
		}
		//Write2S3(&S3log)
	} else {
		// fmt.Println("IN INSERTIN BLOCK")
		var tan TanData
		tan.GLID = strconv.FormatInt(GLUSR_ID, 10)
		tan.TANNO = TAN_ID
		//var ticket int64
		//verified := ""
		PAYMENT_TYPE := "NOT VALID OUTPUT"

		//tan_no,company_name,namefirst,namelast,namemid,address,state_code,pincode

		//query = `Select company_name,namefirst,namelast,namemid,address,state_code,pincode from iil_tan_master_data where tan_no = '` + TAN_ID + `'`
		query = `Select company_name::text,namefirst::text,namelast::text,namemid::text,address::text,state_code::text,pincode::text from iil_tan_master_data where tan_no = '` + TAN_ID + `' order by master_data_insertion_date desc limit 1`

		params = make([]interface{}, 0)
		result, err = database.SelectQuerySql(mainConnection, query, params)
		//fmt.Println("TAN table result",result)
		if err != nil {
			// fmt.Println("Could not connect the approval database because of following err " + err.Error() + " ...!!!")
			//S3log.TicketCreated = "Y"
			CreateTicket(Tanidd, "", "Could not connect the approvalPG database ", Companyid)
			S3log.Result["Could not connect the approval database because of following err  "] = err.Error()
			Write2S3(&S3log)
			return
		}
		if result != nil {
			// fmt.Println("ERRRRRRRRRRRRRRRRRRRORR")
			if queryData, dataExists := result["queryData"]; dataExists {
				if queryData != nil {
					resultData := queryData.([]interface{})

					if len(resultData) > 0 {

						for _, item := range resultData {
							//fmt.Println("item")

							res := item.(map[string]interface{})
							compname, compnameExists := res["company_name"]
							if compnameExists && compname != nil {
								tan.COMPANY_NAME = compname.(string)
							}

							firstname, firstnameExists := res["namefirst"]
							if firstnameExists && firstname != nil {
								tan.FIRSTNAME = firstname.(string)
							}
							lastname, lastnameExists := res["namelast"]
							if lastnameExists && lastname != nil {
								tan.LASTNAME = lastname.(string)
							}
							midname, midnameExists := res["namemid"]
							if midnameExists && midname != nil {
								tan.MIDNAME = midname.(string)
							}
							Address, AddressExists := res["address"]
							if AddressExists && Address != nil {
								tan.ADDRESS = Address.(string)
							}
							statecode, statecodeExists := res["state_code"]
							if statecodeExists && statecode != nil {
								tan.STATECD = statecode.(string)
							}
							Pincode, PincodeExists := res["pincode"]
							if PincodeExists && Pincode != nil {
								tan.PINCODE = Pincode.(string)
							}
							// fmt.Println("EEEROR")
						}

					}
				}

			}
		}

		S3log.Result["TAN_MASTER_TABLE_DATA"] = tan
		if tan.COMPANY_NAME != "" {
			PAYMENT_TYPE = "NOT VERIFIED"
			//fmt.Println("inside not verified")
			percent, _ := Similarity(tan.GLID, tan.COMPANY_NAME)
			//fmt.Println("Similarity Percentage",percent)
			if percent >= 75 {
				PAYMENT_TYPE = "VERIFIED WITH COMPANY NAME"
			}
			// fmt.Println("inside not verified similarity")

			if percent < 75 {
				//percent, _ = SimilarityCEO(tan.GLID, tan.COMPANY_NAME)
				//fmt.Println("SimilarityCEO Percentage",percent)
				percent1, _ := SimilarityCEO(tan.GLID, tan.COMPANY_NAME)
				percent2, _ := SimilarityCEO(tan.GLID, tan.FIRSTNAME+" "+tan.MIDNAME+" "+tan.LASTNAME)
				if percent1 >= percent2 {
					percent = percent1
				} else {
					percent = percent2
				}
			}
			// fmt.Println("inside not verified ceo")

			if PAYMENT_TYPE == "NOT VERIFIED" && percent >= 60 {
				percentPin, _ := SimilarityPIN(tan.GLID, tan.PINCODE)
				//fmt.Println("SimilarityPIN Percentage",percentPin)
				if percentPin == 100 {
					PAYMENT_TYPE = "VERIFIED WITH CEO NAME"
				} else {
					// percentAddress, _ := SimilarityAddress(tan.GLID, tan.ADDRESS)
					// //fmt.Println("Similarity percentAddress  Percentage",percentAddress)
					// if percentAddress >= 50 {
					//         PAYMENT_TYPE = "VERIFIED WITH CEO NAME"
					// }
					statecodestatus, _ := SimilarityStatecd(tan.GLID, tan.STATECD)
					if statecodestatus == true {
						PAYMENT_TYPE = "VERIFIED WITH CEO NAME"
					}
				}

			}
			// fmt.Println("inside not verified PIN")

			//if percent < 60 {
			//percent, _ = SimilarityCONTACT(tan.GLID, tan.COMPANY_NAME)
			//fmt.Println("Similarity CONTACT  Percentage",percent)
			//}
			// fmt.Println("inside not verifie contact d", PAYMENT_TYPE)
			if percent < 60 {
				//percent, _ = SimilarityCONTACT(tan.GLID, tan.COMPANY_NAME)
				//fmt.Println("Similarity CONTACT  Percentage",percent)
				percent1, _ := SimilarityCONTACT(tan.GLID, tan.COMPANY_NAME)
				percent2, _ := SimilarityCONTACT(tan.GLID, tan.FIRSTNAME+" "+tan.LASTNAME)
				if percent1 >= percent2 {
					percent = percent1
				} else {
					percent = percent2
				}
			}
			if PAYMENT_TYPE == "NOT VERIFIED" && percent >= 60 {
				percentPin, _ := SimilarityPIN(tan.GLID, tan.PINCODE)
				//fmt.Println("SimilarityPIN  Percentage",percentPin)
				if percentPin == 100 {
					PAYMENT_TYPE = "VERIFIED WITH CEO NAME"
				} else {
					// percentAddress, _ := SimilarityAddress(tan.GLID, tan.ADDRESS)
					// //fmt.Println("Similarity Address  Percentage",percentAddress)
					// if percentAddress >= 50 {
					//         PAYMENT_TYPE = "VERIFIED WITH CEO NAME"
					// }
					statecodestatus, _ := SimilarityStatecd(tan.GLID, tan.STATECD)
					if statecodestatus == true {
						PAYMENT_TYPE = "VERIFIED WITH CEO NAME"
					}
				}
			}
			S3log.Result[" PAYMENT_TYPE "] = PAYMENT_TYPE
			//  Write2S3(&S3log)

			if PAYMENT_TYPE == "NOT VERIFIED" {
				CreateNewTicket(companyidd, tan)
				//fmt.Println(ticket) //checkonce
				//S3log.Result[" PAYMENT_TYPE "] = PAYMENT_TYPE
				//Write2S3(&S3log)
				fmt.Println("PAYMENT_TYPE", PAYMENT_TYPE)

			} else {
				//S3log.Verified = "Y"
				//verified = "verified"
				status := TanVerifyApi(tan)
				S3log.Result["verificationstatus"] = status

				//fmt.Println(verified) //check once

			}
		} // else {
		// 	//fmt.Println("comp empty-API hit")
		// 	//var tan TanData
		// 	readData, msg, dataTan := tanVerification(TAN_ID, GLUSR_ID, 0, companyidd)

		// 	if msg == "XXXX" {
		// 		return
		// 	}

		// 	//fmt.Println("readData",readData,"msg", msg,"dataTan", dataTan)
		// 	_ = msg
		// 	_ = dataTan

		// 	tan.GLID = strconv.FormatInt(GLUSR_ID, 10)
		// 	tan.TANNO = TAN_ID
		// 	//var ticket int64
		// 	//verified := ""
		// 	PAYMENT_TYPE = "NOT VALID OUTPUT"
		// 	if readData["nameOrgn"] != nil {

		// 		tan.COMPANY_NAME = (readData["nameOrgn"]).(string)
		// 	}
		// 	if readData["addLine1"] != nil {
		// 		if readData["addLine1"] != "ADD_LINE_1" {
		// 			tan.ADDRESS = (readData["addLine1"]).(string) + ` `
		// 		}
		// 	}
		// 	if readData["addLine2"] != nil {
		// 		tan.ADDRESS += (readData["addLine2"]).(string) + ` `
		// 	}
		// 	if readData["addLine3"] != nil {
		// 		tan.ADDRESS += (readData["addLine3"]).(string) + ` `
		// 	}
		// 	if readData["addLine4"] != nil {
		// 		tan.ADDRESS += (readData["addLine4"]).(string) + ` `
		// 	}
		// 	if readData["addLine5"] != nil {
		// 		tan.ADDRESS += (readData["addLine5"]).(string) + ` `
		// 	}
		// 	if readData["nameLast"] != nil {
		// 		tan.LASTNAME = (readData["nameLast"]).(string) + ` `
		// 	}
		// 	if readData["nameFirst"] != nil {
		// 		tan.FIRSTNAME = (readData["nameFirst"]).(string) + ` `
		// 	}
		// 	if readData["nameMid"] != nil {
		// 		tan.MIDNAME = (readData["nameMid"]).(string) + ` `
		// 	}

		// 	if readData["pin"] != nil {
		// 		tan.PINCODE = fmt.Sprint((readData["pin"]).(float64))
		// 	}
		// 	if readData["stateCd"] != nil {
		// 		tan.STATECD = fmt.Sprint((readData["stateCd"]).(float64))
		// 	}
		// 	S3log.Result["TAN_MASTER_TABLE_DATA"] = tan
		// 	if tan.COMPANY_NAME != "" {
		// 		//fmt.Println("TAN DATA",tan)
		// 		PAYMENT_TYPE = "NOT VERIFIED"
		// 		percent, _ := Similarity(tan.GLID, tan.COMPANY_NAME)
		// 		if percent >= 75 {
		// 			PAYMENT_TYPE = "VERIFIED WITH COMPANY NAME"
		// 		}
		// 		if percent < 75 {
		// 			//percent, _ = SimilarityCEO(tan.GLID, tan.COMPANY_NAME)
		// 			percent1, _ := SimilarityCEO(tan.GLID, tan.COMPANY_NAME)
		// 			percent2, _ := SimilarityCEO(tan.GLID, tan.FIRSTNAME+" "+tan.MIDNAME+" "+tan.LASTNAME)
		// 			if percent1 >= percent2 {
		// 				percent = percent1
		// 			} else {
		// 				percent = percent2
		// 			}
		// 		}
		// 		if PAYMENT_TYPE == "NOT VERIFIED" && percent >= 60 {
		// 			percentPin, _ := SimilarityPIN(tan.GLID, tan.PINCODE)
		// 			if percentPin == 100 {
		// 				PAYMENT_TYPE = "VERIFIED WITH CEO NAME"
		// 			} else {
		// 				// percentAddress, _ := SimilarityAddress(tan.GLID, tan.ADDRESS)
		// 				// if percentAddress >= 50 {
		// 				//         PAYMENT_TYPE = "VERIFIED WITH CEO NAME"
		// 				// }
		// 				statecodestatus, _ := SimilarityStatecd(tan.GLID, tan.STATECD)
		// 				if statecodestatus == true {
		// 					PAYMENT_TYPE = "VERIFIED WITH CEO NAME"
		// 				}
		// 			}

		// 		}
		// 		if percent < 60 {
		// 			// percent, _ = SimilarityCONTACT(tan.GLID, tan.COMPANY_NAME)
		// 			percent1, _ := SimilarityCONTACT(tan.GLID, tan.COMPANY_NAME)
		// 			percent2, _ := SimilarityCONTACT(tan.GLID, tan.FIRSTNAME+" "+tan.LASTNAME)
		// 			if percent1 >= percent2 {
		// 				percent = percent1
		// 			} else {
		// 				percent = percent2
		// 			}
		// 		}
		// 		if PAYMENT_TYPE == "NOT VERIFIED" && percent >= 60 {
		// 			percentPin, _ := SimilarityPIN(tan.GLID, tan.PINCODE)
		// 			if percentPin == 100 {
		// 				PAYMENT_TYPE = "VERIFIED WITH CEO NAME"
		// 			} else {
		// 				// percentAddress, _ := SimilarityAddress(tan.GLID, tan.ADDRESS)
		// 				// if percentAddress >= 50 {
		// 				//         PAYMENT_TYPE = "VERIFIED WITH CEO NAME"
		// 				// }
		// 				statecodestatus, _ := SimilarityStatecd(tan.GLID, tan.STATECD)
		// 				if statecodestatus == true {
		// 					PAYMENT_TYPE = "VERIFIED WITH CEO NAME"
		// 				}
		// 			}
		// 		}

		// 		S3log.Result[" PAYMENT_TYPE "] = PAYMENT_TYPE
		// 		//fmt.Println("PAYMENT_TYPE",PAYMENT_TYPE)

		// 		if PAYMENT_TYPE == "NOT VERIFIED" {
		// 			CreateNewTicket(companyidd, tan)
		// 			//fmt.Println(ticket) //checkonce
		// 			//S3log.Result[" PAYMENT_TYPE "] = PAYMENT_TYPE
		// 			//Write2S3(&S3log)
		// 			//fmt.Println("PAYMENT_TYPE",PAYMENT_TYPE)
		// 		} else {
		// 			//S3log.Verified = "Y"
		// 			//verified = "verified"
		// 			status := TanVerifyApi(tan)
		// 			S3log.Result["verificationstatus"] = status
		// 			//fmt.Println("STATUSSSS",status)
		// 			//fmt.Println(verified) //check once

		// 		}
		// 	}

		// 	// err1 := InsertInDB(tan)
		// 	// if err1 != nil {
		// 	//      // fmt.Println("error fetching inserting data", err)
		// 	//      S3log.Result["error fetching inserting data  "] = err.Error()
		// 	//      Write2S3(&S3log)
		// 	// }
		// 	//Write2S3(&S3log)

		// }

		// err1 := InsertInDB(tan)
		// if err1 != nil {
		// 	//fmt.Println("error fetching inserting data", err)
		// 	S3log.Result["error fetching inserting data  "] = err.Error()
		// 	Write2S3(&S3log)
		// 	return
		// }
		//Write2S3(&S3log)
	}

}

func companyy(glid string, ak string) (string, error) {

	client := &http.Client{
		Timeout: 6 * time.Second,
	}

	//token := "imobile@15061981"
	// empid := "86906"
	//glid:="12808473"
	modid := "gst"
	startt := utils.GetTimeInNanoSeconds()
	//S3log.ExtApihit = "Y"
	url := "https://merp.intermesh.net/index.php/userlist/GlidDetails?modid=" + modid + "&glid=" + glid + "&AK=" + ak

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		//fmt.Println("error", err)
		S3log.Result["error_in_companyapi_http_req"] = err.Error()
		Write2S3(&S3log)
		return "", err
	}

	resp, err := client.Do(req)
	endd := utils.GetTimeInNanoSeconds()
	S3log.ApiResponsetime["API_RESPONSE_COMPANYID "] = (endd - startt) / 1000000
	// S3log.Result["response from companyid fetching api"] = resp
	if err != nil {
		S3log.Result["error_in_companyapi_http_Client"] = err.Error()
		Write2S3(&S3log)
		return "", err
		//fmt.Println("error", err)
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
		//fmt.Println("error", err)
	}
	bodyString := string(bodyBytes)
	//fmt.Println(bodyString)
	var data []interface{}
	err = json.Unmarshal([]byte(bodyString), &data)
	if err != nil {
		S3log.Result["error_in_companyapi_unmarshalling"] = err.Error()
		Write2S3(&S3log)
		return "", err
		//fmt.Println(err)
	}
	//fmt.Println("data", data)

	//fmt.Println("data 0", data[0])
	x, ok := data[0].(map[string]interface{})
	if !ok {
		S3log.Result["error_in_companyapi_array"] = err.Error()
		Write2S3(&S3log)
		return "", fmt.Errorf("error_in_companyapi")
	}
	//fmt.Printf("%s", (x["STSID"].(string)))
	dataArray, ok := x["data"].([]interface{})
	if !ok || len(dataArray) == 0 {
		S3log.Result["error_in_companyapi_array"] = err.Error()
		Write2S3(&S3log)
		return "", fmt.Errorf("error_in_companyapi_unmarshalling")
	}

	itemMap, ok := dataArray[0].(map[string]interface{})
	if !ok {
		S3log.Result["error_in_companyapi_array"] = err.Error()
		Write2S3(&S3log)
		return "", fmt.Errorf("unexpected JSON structure")
	}

	v, ok := itemMap["STSID"].(string)
	if !ok {
		S3log.Result["not_getting_stsid"] = err.Error()
		Write2S3(&S3log)
		return "", fmt.Errorf("STSID not found or not a string")
	}

	// v := x["STSID"].(string)
	S3log.Result["companyid from api"] = v
	//Write2S3(&S3log)
	return v, nil
}

// Write2S3 ...
func Write2S3(logs *S3Log) {

	logsDir := properties.Prop.LOG_TAN + utils.TodayDir()

	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		e := os.MkdirAll(logsDir, os.ModePerm)
		fmt.Println(e)
	}

	logsDir += "/tan_wrapper.json"

	fmt.Println(logsDir)

	jsonLog, _ := json.Marshal(*logs)

	//fmt.Println("jsonLog after Marshalling",jsonLog)
	jsonLogString := string(jsonLog[:len(jsonLog)])
	fmt.Println("wrapper-queue", jsonLogString)
	f, _ := os.OpenFile(logsDir, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()

	mutex.Lock()
	defer mutex.Unlock()
	//fmt.Println("Write Json LOg ",jsonLogString)
	f.WriteString("\n" + jsonLogString)
	return
}
