package gstmmcontrols

import (
        "errors"
        "mm/utils"
        "strings"
        //        "fmt"
)

type bucket struct {
        Pr    int
        Dispo string
}

//for bucket's priority and its disposition
var buckets map[string]bucket = map[string]bucket{
        "N17A": bucket{1, "Trade Name + City/Zip Matching"},
        "N10A": bucket{2, "Trade Name + City/Zip Matching"},
        "AA1":  bucket{3, "Trade Name + City/Zip Matching"},
        "6A":   bucket{4, "Trade Name + City/Zip Matching"},
        "AA2":  bucket{5, "Trade Name Legal Name Matching"},
        "10A":  bucket{6, "Trade Name Legal Name Matching"},
        "5T":   bucket{7, "Trade Name + City/Zip Matching"},
        "6T":   bucket{8, "Trade Name Legal Name Matching"},
        "7T":   bucket{9, "Trade Name + City/Zip Matching"},
        "N1A":  bucket{13, "Trade Name + City/Zip Matching"},
        "N1B":  bucket{11, "Trade Name and Legal Name Matched"},
        "N1C":  bucket{12, "Trade Name and Legal Name Matched"},
        "N2A":  bucket{10, "Trade Name + City/Zip Matching"},
        "N2B":  bucket{14, "Trade Name + City/Zip Matching"},
        "6":    bucket{16, ""},
        "7":    bucket{17, ""},
        "10":   bucket{18, ""},
        "5":    bucket{19, ""},
        "N3":   bucket{23, ""},
        "CR1":  bucket{24, ""},
        // "9":    bucket{18, ""},
        // "11":   bucket{20, ""},
        // "N1":   bucket{21, ""},
        // "N2":   bucket{22, ""},
        // "N4":   bucket{24, ""},
        // "N5":   bucket{25, ""},
}

//FindGst ...
func FindGst(glid int) (GSTMatch, Score, Score1, map[string]float64, []string, error) {

        exectime := make(map[string]float64)

        _, st := utils.GetExecTime()
        arr := []string{}
        var result GSTMatch
        glidGstData, err := GetGlidRecords(Database, glid)
        exectime["query GetGlidRecords"], st = utils.GetExecTime(st)
        if err != nil {
                return result, Score{}, Score1{}, exectime, arr, err
        }

        if len(glidGstData) == 0 {
                return result, Score{}, Score1{}, exectime, arr, errors.New("no records found")
        }

        for _, row := range glidGstData {

                if strings.ToUpper(row["glusr_usr_companyname"]) == "NA" {
                        return result, Score{}, Score1{}, exectime, arr, errors.New("gl company name is NA")
                }

                if len(row["gstin_number"]) < 6 {
                        continue
                }

                res, _ := MatchMakingScore(row)
                res.BucketType, res.BucketName = LogicAUTO(res.scores)
                //AA1
                /*
                   1.If bucket name is not decided in the first LOGICAUTO then we have to go
                   for Stage 1 logic auto
                   2. Again if its not there then we have to go for LOGIC MAN
                   3. if its not in LOGIC MAN then we have to go for LOGIC MAN APPROVAL.
                */

                //Stage-1 AUTO
                if res.BucketName == "" {
                        res, _ = MatchMakingScoreStage1(row, res.scores)
                        res.BucketType, res.BucketName = GetBucketAUTO(res.scoresStage1, res.Gstin, row["glusr_usr_companyname"])
                        //N1
                        //Priorities should be placed here for Stage 1 buckets
                }
                //whenever it is not assigned to these two above buckets then res.scores will be empty
                if res.BucketName == "" {
                        //1
                        res.BucketType, res.BucketName = LogicMAN(res.scores)
                }

                if res.BucketName == "" {
                        //N1
                        res.BucketType, res.BucketName = getBucket(res.scoresStage1)
                }

                //no gst
                if res.BucketType == "" {
                        continue
                }

                //first gst
                if result.BucketName == "" {
                        result = res
                        //fmt.Println(result,"First Result")
                        continue
                }

                //different bucket priority
                //21 < 12
                if buckets[res.BucketName].Pr < buckets[result.BucketName].Pr {
                        result = res
                        //fmt.Println(result,"Priority Wise")
                        //fmt.Println(res.BucketName,"Res")
                        //fmt.Println(result.BucketName,"Result")
                        continue
                }

                //same priority / check pin count
                if buckets[res.BucketName].Pr == buckets[result.BucketName].Pr {

                        if res.scores.PincodeScore > result.scores.PincodeScore || res.scoresStage1.PinScore > result.scoresStage1.PinScore {
                                result = res
                        }
                }
        }

        exectime["logic Match making"], _ = utils.GetExecTime(st)

        if result.BucketType == "" {
                return result, Score{}, Score1{}, exectime, arr, errors.New("no gst found in match making")
        }

        var ret Score = result.scores
        var ret2 Score1
        arr = []string{"N1A", "N1B", "N1C", "N2A", "N2B", "N3", "CR1"}
        for i := 0; i < len(arr); i++ {
                if result.BucketName == arr[i] {
                        ret2 = result.scoresStage1
                        return result, ret, ret2, exectime, arr, nil
                }
        }
        return result, ret, ret2, exectime, arr, nil
}

//FindGst ...
func FindGstContactdetails(glid int, gstnumbers []string) (GSTMatch, Score, Score1, map[string]float64, []string, error) {

        exectime := make(map[string]float64)

        _, st := utils.GetExecTime()
        arr := []string{}
        var result GSTMatch
        glidGstData, err := GetGlidRecordsContactdetails(Database, glid, gstnumbers)
        exectime["query GetGlidRecords"], st = utils.GetExecTime(st)
        if err != nil {
                return result, Score{}, Score1{}, exectime, arr, err
        }

        if len(glidGstData) == 0 {
                return result, Score{}, Score1{}, exectime, arr, errors.New("no records found")
        }

        for _, row := range glidGstData {

                if strings.ToUpper(row["glusr_usr_companyname"]) == "NA" {
                        return result, Score{}, Score1{}, exectime, arr, errors.New("gl company name is NA")
                }

                if len(row["gstin_number"]) < 6 {
                        continue
                }

                res, _ := MatchMakingScore(row)
                res.BucketType, res.BucketName = LogicAUTO(res.scores)
                //AA1
                /*
                   1.If bucket name is not decided in the first LOGICAUTO then we have to go
                   for Stage 1 logic auto
                   2. Again if its not there then we have to go for LOGIC MAN
                   3. if its not in LOGIC MAN then we have to go for LOGIC MAN APPROVAL.
                */

                //Stage-1 AUTO
                if res.BucketName == "" {
                        res, _ = MatchMakingScoreStage1(row, res.scores)
                        res.BucketType, res.BucketName = GetBucketAUTO(res.scoresStage1, res.Gstin, row["glusr_usr_companyname"])
                        //N1
                        //Priorities should be placed here for Stage 1 buckets
                }
                //whenever it is not assigned to these two above buckets then res.scores will be empty
                if res.BucketName == "" {
                        //1
                        res.BucketType, res.BucketName = LogicMAN(res.scores)
                }

                if res.BucketName == "" {
                        //N1
                        res.BucketType, res.BucketName = getBucket(res.scoresStage1)
                }

                //no gst
                if res.BucketType == "" {
                        continue
                }

                //first gst
                if result.BucketName == "" {
                        result = res
                        //fmt.Println(result,"First Result")
                        continue
                }

                //different bucket priority
                //21 < 12
                if buckets[res.BucketName].Pr < buckets[result.BucketName].Pr {
                        result = res
                        //fmt.Println(result,"Priority Wise")
                        //fmt.Println(res.BucketName,"Res")
                        //fmt.Println(result.BucketName,"Result")
                        continue
                }

                //same priority / check pin count
                if buckets[res.BucketName].Pr == buckets[result.BucketName].Pr {

                        if res.scores.PincodeScore > result.scores.PincodeScore || res.scoresStage1.PinScore > result.scoresStage1.PinScore {
                                result = res
                        }
                }
        }

        exectime["logic Match making"], _ = utils.GetExecTime(st)

        if result.BucketType == "" {
                return result, Score{}, Score1{}, exectime, arr, errors.New("no gst found in match making")
        }

        var ret Score = result.scores
        var ret2 Score1
        arr = []string{"N1A", "N1B", "N1C", "N2A", "N2B", "N3", "CR1"}
        for i := 0; i < len(arr); i++ {
                if result.BucketName == arr[i] {
                        ret2 = result.scoresStage1
                        return result, ret, ret2, exectime, arr, nil
                }
        }
        return result, ret, ret2, exectime, arr, nil
}


//MatchMakingScore ...
func MatchMakingScore(data map[string]string) (GSTMatch, Score) {
        var scStage1 Score1
        glState := cleanStr(data["glusr_usr_state"])
        glCity := data["glusr_usr_city"]
        glPincode := data["glusr_usr_zip"]
        glOwnerName := data["glusr_usr_firstname"] + " " + data["glusr_usr_middlename"] + " " + data["glusr_usr_lastname"]

        glCeoName := data["glusr_usr_cfirstname"] + " " + data["glusr_usr_clastname"]
        glCeoName = cleanOwnerName(glCeoName)

        gstState := data["state_name"]
        //fmt.Println(gstState,"Dev-Testing-gstState")
        glAddress := data["glusr_usr_add1"] + " " + data["glusr_usr_add2"] + " " + data["glusr_usr_locality"] + " " + data["glusr_usr_landmark"] + " " + glCity
        glAddress = remDuplicates(glAddress)
        //fmt.Println(glAddress,"GlAddressStage0")
        glAddressWoSCP := remWords(glAddress, glState, glCity, glPincode)

        gstPincode := data["pincode"]
        gstOwnerName := data["business_name_replaced"]

        //primary cleaned address
        gstAddress := data["business_fields_add_replaced"]

        //gst original address without cleaning
        gstOrgAddress := data["bussiness_fields_add"]
        //fmt.Println(gstOrgAddress,"Dev-Testing-gstOrgAddress")
        //secondary/additional cleaned address
        gstAddressSecond := data["business_address_add_replaced"]

        gstin := data["gstin_number"]

        sc := Score{
                PinLen:                        len(glPincode),
                OwnerNameScore:                findOwnerNameScore(glOwnerName, gstOwnerName),
                CeoNameScore:                  findOwnerNameScore(glCeoName, gstOwnerName),
                AddressScore:                  findAddrScore(glAddress, gstAddress),
                AddressScoreSecond:            findAddrScore(glAddress, gstAddressSecond),
                AddressScoreWoSCP:             findAddrScore(glAddressWoSCP, gstAddress),
                OwnerNameLen:                  stringLength(glOwnerName),
                CeoNameLen:                    stringLength(glCeoName),
                AddressLen:                    stringLength(glAddress),
                AddressLenWoSCP:               stringLength(glAddressWoSCP),
                PincodeScore:                  prefixMatchCnt(glPincode, gstPincode),
                IsTradeBizSame:                tradeBizSame(data["trade_name_replaced"], gstOwnerName),
                IsME:                          gstCaseME(gstin),
                AddressLenWoSCP2EachWordLenG2: eachWordAtleast2charAfterClean(glAddressWoSCP),
                IsSateSame:                    IsSateSame(glState, gstState),
        }

        return GSTMatch{
                Gstin:            gstin,
                BucketType:       "",
                BucketName:       "",
                GstStatus:        data["gstin_status"],
                GstInsertionDate: data["gst_insertion_date"],
                GstPincode:       gstPincode,
                TradeName:        gstOwnerName,
                GstAddress:       gstOrgAddress,
                GstState:         gstState,
                scores:           sc,
                scoresStage1:     scStage1,
        }, sc
}

//Stage1MatchMakingScore
func MatchMakingScoreStage1(data map[string]string, first Score) (GSTMatch, Score1) {
        glState := data["glusr_usr_state"]
        //custtype := data["glusr_usr_custtype_id"]

        glCity := data["glusr_usr_city"]
        glCity = replaceCity(glCity)

        glPincode := data["glusr_usr_zip"]
        glOwnerName := data["glusr_usr_firstname"] + " " + data["glusr_usr_lastname"]
        glCeoName := data["glusr_usr_cfirstname"] + " " + data["glusr_usr_clastname"]
        glOwnerName = removeMrMrs(glOwnerName)
        glCeoName = removeMrMrs(glCeoName)

        //exclude city/state/pincode from addr line1 + line2 + locality + landmark
        glAddress := data["glusr_usr_add1"] + " " + data["glusr_usr_add2"] + " " + data["glusr_usr_locality"] + " " + data["glusr_usr_landmark"]
        glAddress = replaceCity(glAddress)
        glAddress = excludeFromGlAddr(glAddress, []string{glCity, glState, glPincode})
        //fmt.Println(glAddress,"GlAddress")
        gstPincode := data["pincode"]
        gstOwnerName := data["business_name_replaced"]
        gstOwnerName = removeMrMrs(gstOwnerName)

        gstAddress := data["business_fields_add_replaced"] + " " + data["business_address_add_replaced"] + " " + data["building_name_replaced"] + " " + data["street_replaced"] + " " + data["location_replaced"] + " " + data["door_number_replaced"] + " " + data["floor_number_replaced"]
        gstAddress = replaceCity(gstAddress)
        gstOrgAddressStage1 := data["bussiness_fields_add"]
        //fmt.Println(gstOrgAddressStage1,"Dev-Testing-gstOrgAddress")
        gstState := strings.Trim(strings.ToLower(data["state_name"]), " ")
        //fmt.Println(gstState,"Dev-Testing-gstStatestage1")
        gstin := data["gstin_number"]

        resSc := stringMatchScoreStage1(glOwnerName, gstOwnerName, true)
        ceoSc := stringMatchScoreStage1(glCeoName, gstOwnerName, true)

        if ceoSc > resSc {
                resSc = ceoSc
        }

        sco := Score1{
                CompanyNameScore: 1,
                StateScore:       getStateScore(glState, gstState),
                CityScore:        getCityScore(glCity, gstAddress),
                AddressScore:     stringMatchScoreStage1(glAddress, gstAddress, false),
                PinScore:         prefixMatchCnt(glPincode, gstPincode),
                OwnerNameScore:   resSc,
                GlOwnerNameLen:   stringLength(glOwnerName),
                GlPinLen:         len(glPincode),
                GlAddressLen:     stringLength(glAddress),
        }

        return GSTMatch{
                Gstin:            gstin,
                BucketType:       "",
                BucketName:       "",
                GstStatus:        data["gstin_status"],
                GstInsertionDate: data["gst_insertion_date"],
                GstPincode:       gstPincode,
                TradeName:        gstOwnerName,
                GstAddress:       gstOrgAddressStage1,
                GstState:         gstState,
                scoresStage1:     sco,
                scores:           first,
        }, sco
}

//Buckets Manual for Stage 1
func getBucket(sc Score1) (string, string) {

        if sc.CompanyNameScore == 1 && sc.StateScore == 1 {

                // if sc.CityScore == 1 {
                //         return "MAN", "N1"
                // }

                // if sc.PinScore >= 2 {
                //         return "MAN", "N2"
                // }

                if sc.OwnerNameScore == 1 {
                        return "MAN", "N3"
                }

                if (sc.CompanyNameScore == 1 && (sc.PinScore < 6 && sc.PinScore >= 2) &&  (sc.OwnerNameScore >= 0.6 && sc.OwnerNameScore <= 1) && sc.AddressScore >= 0.6 && sc.GlAddressLen >=3 ){
                        return "MAN", "CR1"
                }
         



                // if sc.OwnerNameScore > 0 && sc.OwnerNameScore < 1 && sc.AddressScore > 0 && sc.AddressScore < 0.5 {
                //         return "MAN", "N4"
                // }

                // if sc.AddressScore >= 0.5 {
                //         return "MAN", "N5"
                // }

        }

        return "", ""
}

//Buckets Auto for Stage-1
func GetBucketAUTO(sc Score1, gstin string, glidcomp string) (string, string) {
        approvedBucket := ""
        addrScore := sc.AddressScore
        ownerScore := sc.OwnerNameScore
        pinScore := sc.PinScore
        // if sc.CityScore == 1 && sc.CompanyNameScore == 1 && sc.StateScore == 1{
        //         if pinScore == 6 && addrScore >= 0.5 && sc.GlAddressLen > 4 {
        //                 approvedBucket = "N1A"
        //         } else if string(gstin[5]) != "C" && ownerScore == 1 {
        //                 approvedBucket = "N1B"
        //         } else if string(gstin[5]) != "C" && ownerScore >= 0.5 && ownerScore < 1 {
        //                 approvedBucket = "N1C"
        //         }
        // } else if pinScore >= 2 && sc.CompanyNameScore == 1 && sc.StateScore == 1 {
        //         if pinScore == 6 && addrScore >= 0.5 && sc.GlAddressLen > 4 {
        //                 approvedBucket = "N2A"
        //         } else if string(gstin[5]) != "C" && pinScore == 6 && ownerScore >= 0.5 && ownerScore < 1 {
        //                 approvedBucket = "N2B"
        //         }
        // }

        if pinScore == 6 && sc.CityScore == 1 && sc.CompanyNameScore == 1 && sc.StateScore == 1 {
                if addrScore >= 0.5 && sc.GlAddressLen > 4 && ownerScore >= 0.5 && ownerScore < 1{
                        approvedBucket = "N2A"

                } else if string(gstin[5]) != "C" && ownerScore == 1 && countWordsCompanyName(glidcomp) == 1 {
                        //single word

                        flag_keyword := 0
                        Keywords := []string{"garments", "tirupati", "electrical", "gadjet", "trading", "marketing", "enterprise", "traders", "engineering", "store", "brothers", "associate", "trader", "enterprises", "stores", "electricals"}
                        for _, s := range Keywords {
                                if strings.ToLower(glidcomp) == s {
                                        flag_keyword = 1
                                        break
                                }
                        }
                        if flag_keyword == 0 {
                                approvedBucket = "N1B"
                        }

                } else if string(gstin[5]) != "C" && ownerScore == 1 && countWordsCompanyName(glidcomp) > 1 {
                        //more than 1 word and generic word
                        flag_keyword := 0
                        Keywords := []string{"garments", "tirupati", "electrical", "gadjet", "trading", "marketing", "enterprise", "traders", "engineering", "store", "brothers", "associate", "trader", "enterprises", "stores", "electricals"}
                        for _, s := range Keywords {
                                if strings.Contains(strings.ToLower(glidcomp), s) {
                                        flag_keyword = 1
                                        break
                                }
                        }

                        if flag_keyword == 1 {
                                approvedBucket = "N1C"
                        }

                } else if addrScore >= 0.5 && sc.GlAddressLen > 4  {
                        approvedBucket = "N1A"
                } else if string(gstin[5]) != "C" && ownerScore >= 0.5 && ownerScore < 1 {
                        approvedBucket = "N2B"
                }

        }

        if approvedBucket != "" {
                return "AUTO", approvedBucket
        }
        return "", ""
}

//LogicAUTO ...
func LogicAUTO(sc Score) (string, string) {

        if !sc.IsSateSame {
                return "", ""
        }

        nameScore := sc.OwnerNameScore
        nameLen := sc.OwnerNameLen
        if nameScore < sc.CeoNameScore {
                nameScore = sc.CeoNameScore
                nameLen = sc.CeoNameLen
        }

        addrScore := sc.AddressScore
        addrLen := sc.AddressLen
        if addrScore < sc.AddressScoreSecond {
                addrScore = sc.AddressScoreSecond
        }

        approvalBucket := ""
        if nameLen > 1 {
                //full pincode match
                if sc.PincodeScore == 6 {
                        if nameScore < 0.5 {
                                if addrScore > 0.6 && addrLen > 4 && sc.IsTradeBizSame && sc.IsME {
                                        approvalBucket = "N17A"
                                }
                        } else if nameScore >= 0.5 && nameScore < 1 {
                                if addrScore > 0.6 && addrLen > 4 && !sc.IsTradeBizSame {
                                        approvalBucket = "N10A"
                                }
                        } else if nameScore == 1 {
                                if addrScore >= 0.5 && addrLen > 4 {
                                        approvalBucket = "AA1"
                                } else if addrScore > 0 && addrScore < 0.5 {
                                        approvalBucket = "6A"
                                }
                        }

                } else if sc.PincodeScore >= 4 { //4 digit match
                        if nameScore == 1 && addrLen > 4 {
                                if addrScore >= 0.5 {
                                        approvalBucket = "AA2"
                                } else if addrScore >= 0.35 && addrScore < 0.5 {
                                        approvalBucket = "10A"
                                }
                        }
                }
        }

        if approvalBucket != "" {
                return "AUTO", approvalBucket
        }

        if sc.PincodeScore == 6 && sc.AddressLenWoSCP >= 2 {

                if (sc.AddressLenWoSCP == 2 && sc.AddressLenWoSCP2EachWordLenG2 && sc.AddressScoreWoSCP == 1) || (sc.AddressLenWoSCP > 2 && sc.AddressScoreWoSCP >= 0.7) {
                        approvalBucket = "5T"

                } else if nameScore == 1 && sc.AddressScoreWoSCP > 0 && sc.AddressScoreWoSCP < 0.5 {
                        approvalBucket = "6T"

                } else if (sc.AddressLenWoSCP == 2 && sc.AddressLenWoSCP2EachWordLenG2 && sc.AddressScoreWoSCP == 1) || (sc.AddressLenWoSCP > 2 && nameScore >= 0.5 && sc.AddressScoreWoSCP >= 0.5 && sc.AddressScoreWoSCP < 0.7) {
                        approvalBucket = "7T"

                }
        }

        if approvalBucket != "" {
                return "AUTO", approvalBucket
        }

        return "", ""
}

//LogicMAN ...
func LogicMAN(sc Score) (string, string) {

        nameScore := sc.OwnerNameScore
        if nameScore < sc.CeoNameScore {
                nameScore = sc.CeoNameScore
        }

        addrScore := sc.AddressScore
        addrLen := sc.AddressLen
        if addrScore < sc.AddressScoreSecond {
                addrScore = sc.AddressScoreSecond
        }

        matchingCriteria := ""

        if sc.PincodeScore == 6 {
                if addrScore >= 0.7 {
                        matchingCriteria = "5"
                } else if nameScore == 1 && addrScore < 0.5 {
                        matchingCriteria = "6"
                } else if (nameScore >= 0.5 && nameScore <= 1) && (addrScore >= 0.5 && addrScore < 0.7) {
                        matchingCriteria = "7"
                }
        } else if sc.PincodeScore >= 2 && addrLen >= 3 {
                // if addrScore >= 0.7 {
                //         matchingCriteria = "9"
                // } else 
                if nameScore == 1 && addrScore > 0 && addrScore < 0.5 {
                        matchingCriteria = "10"
                } 
                // else if (nameScore >= 0.5 && nameScore <= 1) && (addrScore >= 0.5 && addrScore < 0.7) {
                //         matchingCriteria = "11"
                // }
        }

        if matchingCriteria == "" {
                return "", ""
        }

        return "MAN", matchingCriteria
}

