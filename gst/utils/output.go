package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/smtp"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"errors"
)

func GetLocalTime() *time.Location {
	loc, _ := time.LoadLocation("Asia/Kolkata")
	return loc
}

func GetIPAdress(r *http.Request) string {

	var ipAddress string
	for _, h := range []string{"X-Forwarded-For", "X-Real-Ip"} {
		for _, ip := range strings.Split(r.Header.Get(h), ",") {
			// header can contain spaces too, strip those out.
			realIP := net.ParseIP(strings.Replace(ip, " ", "", -1))
			ipAddress = string(realIP)
		}
	}
	if len(ipAddress) == 0 {
		ipAddress = r.RemoteAddr
	}
	return ipAddress
}

func GetTimeStampCurrent() string {
	return time.Now().Local().Format("02-01-2006 15:04:05")
}

func GetTimeInNanoSeconds() float64 {
	return float64(time.Now().UnixNano())
}

// GetExecTime ... exec time , time in nanosec
func GetExecTime(stArr ...float64) (float64, float64) {

	if len(stArr) == 0 {
		return 0, float64(time.Now().UnixNano())
	}
	exec := (float64(time.Now().UnixNano()) - stArr[0]) / 1000000

	return float64(int(exec*100)) / 100, float64(time.Now().UnixNano())
}

// DiffDaysddmmyyyy ...
func DiffDaysddmmyyyy(str string) (float64, error) {

	gstInsertionDate, err := time.Parse("02-01-2006", str)
	if err != nil {
		return 0, err
	}
	nowDate, _ := time.Parse("02-01-2006", time.Now().Format("02-01-2006"))
	days := nowDate.Sub(gstInsertionDate).Hours() / 24
	return days, nil
}

// TodayDir ...
func TodayDir() string {
	year, month, day := time.Now().Local().Date()
	return "/" + fmt.Sprint(year) + "/" + fmt.Sprint(int(month)) + "/" + fmt.Sprint(day)
}

// MaxInt ...
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Send_mail(to []string, subject string, body string) {

	from := "tyagi.prajjwal@indiamart.com"
	password := "umncbamfzhyprjjc"
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	to_id := ""
	for _, v := range to {
		to_id += v + ","
	}
	data := "To: " + to_id + "\r\n" + "Subject: " + subject + "\r\n\r\n" + body + "\r\n"

	message := []byte(data)

	auth := smtp.PlainAuth("", from, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Email Sent Successfully!")
}

func Clear(v interface{}) {
	p := reflect.ValueOf(v).Elem()
	p.Set(reflect.Zero(p.Type()))
}

// cleaning of gstAddress
func CleanAddress(addr string, state string, pincode string) string {
	var index, index2 int
	//strings.TrimSpace(addr)
	addr = strings.Trim(addr, " ")
	if strings.Contains(addr, state) {
		addr = strings.Replace(addr, state, "", -1)
		//addr = strings.Replace(addr, " ", "", -1)
	}
	if strings.Contains(addr, pincode) {
		addr = strings.Replace(addr, pincode, "", -1)
		//addr = strings.Replace(addr, " ", "", -1)
	}
	for i := 0; i < len(addr); i++ {
		if (int(addr[i]) >= 'a' && int(addr[i]) <= 'z') || (int(addr[i]) >= 'A' && int(addr[i]) <= 'Z') || (int(addr[i]) >= '0' && int(addr[i]) <= '9') {
			index2 = i
			break
		}
	}
	if index2 > 0 {
		addr = strings.TrimLeft(addr, ",-")
	}
	n := len(addr)
	for i := n - 1; i >= 0; i-- {
		if (int(addr[i]) >= 'a' && int(addr[i]) <= 'z') || (int(addr[i]) >= 'A' && int(addr[i]) <= 'Z') || (int(addr[i]) >= '0' && int(addr[i]) <= '9') {
			//fmt.Println(string(addr[i]), addr[i])
			index = i
			break
		}
	}
	if index > 0 {
		addr = strings.TrimRight(addr, ",-")
	}
	comma := regexp.MustCompile(`\,+`)
	//dash := regexp.MustCompile(`\-+`)
	s := comma.ReplaceAllString(addr, ",")
	//s= dash.ReplaceAllString(s,"-")
	s = strings.TrimSuffix(s, ",")
	s = strings.ReplaceAll(s, ".", " ")
	s = strings.ReplaceAll(s, "|", " ")
	nonalphanumeric := regexp.MustCompile(`[^a-zA-Z0-9],+`)
	s = nonalphanumeric.ReplaceAllString(s, "")
	s = strings.Replace(s, ", ", ",", -1)
	s = strings.Replace(s, ", ", ",", -1)
	s = strings.TrimSuffix(s, ",")
	//s = strings.TrimPrefix(s,"-")
	s = strings.TrimSpace(s)
	return s
}

// ModifyCompName ... title case company name
func ModifyCompName(compName string) string {
	compName = strings.ToLower(compName)
	a := strings.Fields(compName)
	for k, j := range a {
		if j == "llp" || j == "opc" || j == "m/s" {
			a[k] = strings.ToUpper(a[k])
		}
	}
	compName = strings.Join(a, " ")
	compName = strings.Title(compName)
	return compName
}

// Splitting the name into two parts
func Convert(name string) (string, string) {
	var c, finalString string
	d := ""
	b := strings.Fields(name)
	n := len(strings.Fields(name))
	if n == 1 {
		c = b[0]
		if len(c) >= 30 {
			c = c[:30]
		}
	} else {
		c = b[0]
		for i := 1; i < n; i++ {
			d = d + " " + b[i]
			finalString = strings.TrimSpace(d)
		}
		if len(finalString) >= 30 {
			finalString = finalString[:30]
		}
		if len(c) >= 30 {
			c = c[:30]
		}
	}
	// Convert to title case
	c = strings.Title(strings.ToLower(c))
	finalString = strings.Title(strings.ToLower(finalString))
	return c, finalString
}

// AuthBridge functions
// ENCRYPT FUNCTIONS
func MainFunction(txnId, docType, docNumber, token string) string {
	jsonData := fmt.Sprintf(`{"trans_id": "%s","doc_type": "%s","doc_number": "%s"}`, txnId, docType, docNumber)
	fmt.Println(jsonData)
	encryptedResult := encryption(jsonData, token)
	fmt.Println(encryptedResult)
	//format encrypted data
	jsonData = fmt.Sprintf(`{"requestData": "%s"}`, encryptedResult)
	return jsonData

	// apiResponse := pc.panToGst(jsonData)
	// temp := make(map[string]interface{})
	// json.Unmarshal([]byte(apiResponse), &temp)
	// if _, ok := temp["responseData"]; !ok {
	//      return temp, int16(501), errors.New("failed")
	// }
	// decryptionResult := Decryption(temp["responseData"].(string), token)
	// fmt.Println(decryptionResult)
	// stringMap := make(map[string]interface{})
	// json.Unmarshal([]byte(decryptionResult), &stringMap)
	// return stringMap, int16(200), nil
}

// ///////////////////CRYPTOGRAPHY///////////////////////////////////////////////////
func encryption(jsonData, token string) string {
	result := Encrypt(jsonData, token)
	///final result
	return result
}

func Decryption(responseData, token string) string {
	//////////////////////////
	block, err := aes.NewCipher([]byte(createHash(token)))
	if err != nil {
		panic(err)
	}
	slice := strings.Split(responseData, ":")
	cipherText := Decode(slice[0])
	decodedIV := Decode(slice[1])

	cbc := cipher.NewCBCDecrypter(block, decodedIV)
	paddedPlainText := make([]byte, len(cipherText))
	cbc.CryptBlocks(paddedPlainText, cipherText)

	// Now we are unpadding additional padded bytes
	plainText := PKCS5UnPadding(paddedPlainText)
	return string(plainText)
}

func getIV() []byte {
	token := make([]byte, 16)
	rand.Read(token)
	return token
}
func createHash(key string) string {

	hasher := sha512.New()
	hasher.Write([]byte(key))

	hashedKey := hex.EncodeToString(hasher.Sum(nil))

	hashedKey = hashedKey[0:16]
	return hashedKey
}

func Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(ciphertext []byte) []byte {
	length := len(ciphertext)
	unpadding := int(ciphertext[length-1])
	return ciphertext[:(length - unpadding)]
}

// Encrypt method is to encrypt or hide any classified text
func Encrypt(text, token string) string {

	block, err := aes.NewCipher([]byte(createHash(token)))
	if err != nil {
		panic(err)
	}
	// generate encodedIV
	iv := getIV()
	encodedIV := base64.StdEncoding.EncodeToString(iv)

	cbc := cipher.NewCBCEncrypter(block, iv)

	payload, _, err := prepareUpload(text)
	if err != nil {
		panic(err)
	}
	data := fmt.Sprintf("%v", payload)
	slicedata := strings.Split(data, "&{")
	slicedata1 := strings.Split(slicedata[1], "}")
	data1 := slicedata1[0] + "}"
	plainText := []byte(data1)
	// plaintext will cause panic: crypto/cipher: input not full blocks
	// to fix this issue, plaintext will be padded to full blocks
	plainText = PKCS5Padding(plainText, block.BlockSize())
	cipherText := make([]byte, len(plainText))

	// block.BlockSize()==16
	// this will work only if len(text)%block.BlockSize()==0
	cbc.CryptBlocks(cipherText, plainText)

	// concatenating encrypteddata and encryptedIV
	encryptedData := Encode(cipherText)
	encryptedPayload := encryptedData + ":" + encodedIV
	return encryptedPayload
}

func Decode(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

func prepareUpload(data string) (*strings.Reader, string, error) {

	payload := strings.NewReader(data)

	return payload, "application/json", nil
}

func BusLogicOnMasterData_V3(gstin_number string, m map[string]interface{}) (map[string]string, []interface{}) {

	/*var (
	                  business_name, centre_juri, registration_date, cancel_date, business_constitution, business_activity_nature,
	                  gstin_status, last_update_date, state_jurisdiction_code, state_juri, centre_jurisdiction_code, trade_name, bussiness_fields_add string
	                  location, state_name, pin, taxpayer_type, building_name, street, door_number, floor_number    string
	                  longitude, lattitude, bussiness_place_add_nature, bussiness_address_add, building_name_addl   string
	                  street_addl, location_addl, door_number_addl, state_name_addl, floor_number_addl
	  string
	                  longitude_addl, lattitude_addl, pincode_addl_str, nature_of_business_addl, gst_insertion_date string
	                  bussiness_fields_add_district
	  string
	                  bussiness_fields_pp_district
	  string
	          )*/

	var (
		nature_of_business_activities, door_number, building_name  string
		street, location, bussiness_fields_pp_district, state_name string
		floor_number, pin, bussiness_address, contact_details      string

		nature_of_business_activities_addl, door_number_addl string
		building_name_addl, street_addl, location_addl       string
		bussiness_fields_add_district, state_name_addl       string
		floor_number_addl, pin_addl, bussiness_address_addl  string
		contact_details_addl                                 string
		mobile, email                                        string
		gst_insertion_date                                   string
		business_constitution_group_id		int
	)

	for k, v := range m {

		if v == nil {
			m[k] = ""
		}
	}

	Gstin_number, _ := m["GSTIN/ UIN"].(string)
	fmt.Println("gstin_number::", Gstin_number)


	_, legalStatusID, err := GetLegalStatus(gstin_number)
	if err == nil {
		business_constitution_group_id = legalStatusID
	}

	Legal_Name_of_Business, _ := m["Legal Name of Business"].(string)
	fmt.Println("Legal_Name_of_Business::", Legal_Name_of_Business)

	Trade_Name, _ := m["Trade Name"].(string)
	fmt.Println("Trade_Name::", Trade_Name)

	Date_of_registration, _ := m["Date of registration"].(string)
	Date_of_registration = strings.Trim(Date_of_registration, " ")
	if strings.ToLower(Date_of_registration) == "na" || strings.ToLower(Date_of_registration) == "" {
		Date_of_registration = ""
	}
	fmt.Println("Date_of_registration::", Date_of_registration)

	ConstitutionOfBusiness, _ := m["ConstitutionOfBusiness"].(string)
	fmt.Println("ConstitutionOfBusiness::", ConstitutionOfBusiness)

	// AdministrativeOffice, _ := m["AdministrativeOffice"].(string)
	// fmt.Println("AdministrativeOffice::", AdministrativeOffice)

	// OtherOffice, _ := m["OtherOffice"].(string)
	// fmt.Println("OtherOffice::", OtherOffice)

	Taxpayer_Type, _ := m["Taxpayer Type"].(string)
	fmt.Println("Taxpayer_Type::", Taxpayer_Type)

	GSTIN_Status, _ := m["GSTIN / UIN Status"].(string)
	fmt.Println("GSTIN_STATUS::", GSTIN_Status)

	Date_of_Cancellation, _ := m["Date of Cancellation"].(string)
	Date_of_Cancellation = strings.Trim(Date_of_Cancellation, " ")
	if strings.ToLower(Date_of_Cancellation) == "na" || strings.ToLower(Date_of_Cancellation) == "" {
		Date_of_Cancellation = ""
	}

	fmt.Println("Date_of_Cancellation::", Date_of_Cancellation)

	AnnualAggregateTurnover, _ := m["AnnualAggregateTurnover"].(string)
	fmt.Println("AnnualAggregateTurnover::", AnnualAggregateTurnover)

	GrossTotalIncome, _ := m["GrossTotalIncome"].(string)
	fmt.Println("GrossTotalIncome::", GrossTotalIncome)

	PercentageOfTaxPaymentInCash, _ := m["PercentageOfTaxPaymentInCash"].(string)
	fmt.Println("PercentageOfTaxPaymentInCash::", PercentageOfTaxPaymentInCash)

	var Aadhar_authentication_status bool
	var E_KYC_verification_status bool
	var Field_visit_conducted bool
	WhetherAadhaarAuthenticated, _ := m["WhetherAadhaarAuthenticated"].(string)
	if strings.ToLower(WhetherAadhaarAuthenticated) == "no" || strings.ToLower(WhetherAadhaarAuthenticated) == "na" {
		Aadhar_authentication_status = false
	} else {
		Aadhar_authentication_status = true
	}

	WhetherE_KYCVerified, _ := m["WhetherE-KYCVerified"].(string)
	if strings.ToLower(WhetherE_KYCVerified) == "no" || strings.ToLower(WhetherE_KYCVerified) == "na" {
		E_KYC_verification_status = false
	} else {
		E_KYC_verification_status = true
	}

	field_visit_conducted, _ := m["field_visit_conducted"].(string)
	if strings.ToLower(field_visit_conducted) == "no" || strings.ToLower(field_visit_conducted) == "na" {
		Field_visit_conducted = false
	} else {
		Field_visit_conducted = true
	}

	fmt.Println("Aadhar_authentication_status: ", Aadhar_authentication_status, "E-KYC_verification_status", E_KYC_verification_status, "Field_visit_conducted", Field_visit_conducted)

	NatureOfCoreBusinessActivity, _ := m["NatureOfCoreBusinessActivity"].(string)
	NatureOfBusinessActivities, _ := m["NatureOfBusinessActivities"].(string)
	// proprietor_name, _ := m["proprietor_name"].(string)
	var proprietor_name string
	proprietorData, exists := m["proprietor_name"]

	// Check if "proprietor_name" exists and handle different types
	if exists {
		switch v := proprietorData.(type) {
		case string:
			proprietor_name = v
		case []interface{}:
			var names []string
			for _, name := range v {
				if str, ok := name.(string); ok {
					names = append(names, str)
				}
			}
			proprietor_name = strings.Join(names, ",")
		default:
			// Handle unexpected types
			proprietor_name = ""
		}
	} else {
		proprietor_name = ""
	}

	fmt.Println("NatureOfCoreBusinessActivity:", NatureOfCoreBusinessActivity)
	fmt.Println("NatureOfBusinessActivities:", NatureOfBusinessActivities)
	fmt.Println("proprietor_name:", proprietor_name)

	Centre_Juri, _ := m["Centre Juri"].(string)
	state_juri, _ := m["state_juri"].(string)
	StateJurisdiction, _ := m["StateJurisdiction"].(string)

	fmt.Println("Centre_Juri:", Centre_Juri)
	fmt.Println("state_juri:", state_juri)
	fmt.Println("StateJurisdiction:", StateJurisdiction)

	//address

	adadrArr, ok := m["placeOfBusinessData"].([]interface{})

	secaddrflag := 0
	if ok && len(adadrArr) > 0 {
		// count_addr:=0
		for _, v := range adadrArr {
			fmt.Println(v, "v value")
			v1, _ := v.(map[string]interface{})
			for k, x := range v1 {

				if x == nil {
					v1[k] = ""
				}
			}

			fmt.Println(v1, "v1 str interface")
			//type, okkk := v1["type"].(string)
			type_low := strings.ToLower(v1["type"].(string))

			// changes

			if type_low == "principal" {
				nature_of_business_activities, _ = v1["nature_of_business_activities"].(string)
				door_number, _ = v1["door_number"].(string)
				building_name, _ = v1["building_name"].(string)
				street, _ = v1["street"].(string)
				location, _ = v1["location"].(string)
				bussiness_fields_pp_district, _ = v1["bussiness_fields_pp_district"].(string)
				state_name, _ = v1["state_name"].(string)
				floor_number, _ = v1["floor_number"].(string)
				pin, _ = v1["pincode"].(string)
				bussiness_address, _ = v1["bussiness_address"].(string)
				contact_details, _ = v1["contact_details"].(string)

				contact_details = strings.Trim(contact_details, " ")

				if strings.ToLower(contact_details) == "na" || strings.ToLower(contact_details) == "" {
					_ = mobile
					_ = email
				} else {
					mobemail := SplitMobileEmail(contact_details)
					mobile = strings.Trim(mobemail[0], " ")
					email = strings.Trim(mobemail[1], " ")
				}
			}

			if type_low == "additional" {
				// count_addr=count_addr+1
				if secaddrflag == 0 {
					secaddrflag = 1
					nature_of_business_activities_addl, _ = v1["nature_of_business_activities"].(string)
					door_number_addl, _ = v1["door_number_addl"].(string)
					building_name_addl, _ = v1["building_name_addl"].(string)
					street_addl, _ = v1["street_addl"].(string)
					location_addl, _ = v1["location_addl"].(string)
					bussiness_fields_add_district, _ = v1["bussiness_fields_add_district"].(string)
					state_name_addl, _ = v1["state_name_addl"].(string)
					floor_number_addl, _ = v1["floor_number_addl"].(string)
					pin_addl, _ = v1["pincode_addl"].(string)
					bussiness_address_addl, _ = v1["bussiness_address_addl"].(string)

				}

				if len(mobile) == 0 || len(email) == 0 {
					contact_details_addl, _ = v1["contact_details"].(string)

					contact_details_addl = strings.Trim(contact_details_addl, " ")

					if strings.ToLower(contact_details_addl) == "na" || strings.ToLower(contact_details_addl) == "" {
						_ = mobile
						_ = email

					} else {
						mobemail := SplitMobileEmail(contact_details_addl)

						if len(mobile) == 0 {
							mobile = strings.Trim(mobemail[0], " ")
						}

						if len(email) == 0 {
							email = strings.Trim(mobemail[1], " ")
						}

					}

				}

			}
		}

	}

	fmt.Println(":")
	fmt.Println("nature_of_business_activities:", nature_of_business_activities)
	fmt.Println("door_number:", door_number)
	fmt.Println("building_name:", building_name)
	fmt.Println("street:", street)
	fmt.Println("location:", location)
	fmt.Println("bussiness_fields_pp_district:", bussiness_fields_pp_district)
	fmt.Println("state_name:", state_name)
	fmt.Println("floor_number:", floor_number)
	fmt.Println("pincode:", pin)
	fmt.Println("bussiness_address:", bussiness_address)
	fmt.Println("contact_details:", contact_details)
	fmt.Println("mobile: ", mobile)
	fmt.Println("email: ", email)
	fmt.Println(":")
	fmt.Println("nature_of_business_activities_addl:", nature_of_business_activities_addl)
	fmt.Println("door_number_addl:", door_number_addl)
	fmt.Println("building_name_addl:", building_name_addl)
	fmt.Println("street_addl:", street_addl)
	fmt.Println("location_addl:", location_addl)
	fmt.Println("bussiness_fields_add_district:", bussiness_fields_add_district)
	fmt.Println("state_name_addl:", state_name_addl)
	fmt.Println("floor_number_addl:", floor_number_addl)
	fmt.Println("pincode_addl:", pin_addl)
	fmt.Println("bussiness_address_addl:", bussiness_address_addl)
	fmt.Println("contact_details_addl:", contact_details_addl)

	gst_inserted_by := 111

	res := make(map[string]string)

	res["gstin_number"] = Gstin_number
	res["business_name"] = Legal_Name_of_Business
	res["centre_juri"] = Centre_Juri
	res["registration_date"] = Date_of_registration
	res["cancel_date"] = Date_of_Cancellation
	res["business_constitution"] = ConstitutionOfBusiness
	res["business_activity_nature"] = NatureOfBusinessActivities
	res["gstin_status"] = GSTIN_Status
	//res["last_update_date"] = last_update_date
	res["state_jurisdiction_code"] = StateJurisdiction
	res["state_juri"] = state_juri
	//res["centre_jurisdiction_code"] = centre_jurisdiction_code
	res["trade_name"] = Trade_Name
	res["bussiness_fields_add"] = bussiness_address
	res["location"] = location
	res["state_name"] = state_name
	res["pincode"] = pin
	res["taxpayer_type"] = Taxpayer_Type
	res["building_name"] = building_name
	res["street"] = street
	res["door_number"] = door_number
	res["floor_number"] = floor_number
	// res["longitude"] = longitude
	// res["lattitude"] = lattitude
	res["bussiness_place_add_nature"] = nature_of_business_activities
	res["bussiness_address_add"] = bussiness_address_addl
	res["building_name_addl"] = building_name_addl

	res["street_addl"] = street_addl
	res["location_addl"] = location_addl
	res["door_number_addl"] = door_number_addl
	res["state_name_addl"] = state_name_addl
	res["floor_number_addl"] = floor_number_addl
	// res["longitude_addl"] = longitude_addl
	// res["lattitude_addl"] = lattitude_addl
	res["pincode_addl"] = pin_addl
	res["nature_of_business_addl"] = nature_of_business_activities_addl
	gst_insertion_date = time.Now().Format("2006-01-02 15:04:05")
	res["bussiness_fields_add_district"] = bussiness_fields_pp_district
	res["bussiness_fields_pp_district"] = bussiness_fields_add_district
	res["mobile_number"] = mobile
	res["email_id"] = email
	res["annual_turnover_slab"] = AnnualAggregateTurnover
	res["gross_income"] = GrossTotalIncome
	res["percent_of_tax_payment_in_cash"] = PercentageOfTaxPaymentInCash

	res["core_business_activity_nature"] = NatureOfCoreBusinessActivity
	res["proprieter_name"] = proprietor_name

	res["gst_insertion_date"] = gst_insertion_date
	res["gst_inserted_by"] = strconv.Itoa(gst_inserted_by)

	for i, v := range res {
		v = strings.ReplaceAll(v, "'", " ")
		v = strings.ReplaceAll(v, "-", " ")
		v = strings.ReplaceAll(v, ",", " ")
		v = strings.Trim(v, " ")
		res[i] = v
	}

	//res["aadhar_authentication_status"]=Aadhar_authentication_status.(string)
	//res["ekyc_verification_status"]=E_KYC_verification_status.(string)
	//res["field_visit_conducted"]=Field_visit_conducted.(string)

	var registration_date_nil, cancel_date_nil, pincode, pincode_addl interface{}

	regis_date_Date, err := time.Parse("02/01/2006", res["registration_date"])
	if err == nil {
		registration_date_nil = regis_date_Date.Format("2006-01-02")
	}

	cancel_date_Date, err := time.Parse("02/01/2006", res["cancel_date"])
	if err == nil {
		cancel_date_nil = cancel_date_Date.Format("2006-01-02")
	}

	if res["pincode"] != "" {
		val, err := strconv.Atoi(res["pincode"])
		if err == nil {
			pincode = val
		}
	}
	if res["pincode_addl"] != "" {
		val, err := strconv.Atoi(res["pincode_addl"])
		if err == nil {
			pincode_addl = val
		}
	}

	var params []interface{}

	fmt.Println("=====Values in sidel params")
	fmt.Println("1.Gstin_number", Gstin_number)
	fmt.Println("2.Legal_Name_of_Business", Legal_Name_of_Business)
	fmt.Println("3.Centre_Juri", Centre_Juri)
	fmt.Println("4.registration_date_nil", registration_date_nil)
	fmt.Println("5.cancel_date_nil", cancel_date_nil)
	fmt.Println("6.ConstitutionOfBusiness", ConstitutionOfBusiness)
	fmt.Println("7.NatureOfBusinessActivities", NatureOfBusinessActivities)
	fmt.Println("8.GSTIN_Status", GSTIN_Status)

	fmt.Println("9.mobile", mobile)
	fmt.Println("10.StateJurisdiction", StateJurisdiction)
	fmt.Println("11.state_juri", state_juri)
	fmt.Println("12.email", email)
	fmt.Println("13.Trade_Name", Trade_Name)
	fmt.Println("14.bussiness_address", bussiness_address)
	fmt.Println("15.location", location)
	fmt.Println("16.state_name", state_name)
	fmt.Println("17.pincode", pincode)
	fmt.Println("18.Taxpayer_Type", Taxpayer_Type)
	fmt.Println("19.building_name", building_name)
	fmt.Println("20.street", street)
	fmt.Println("21.door_number", door_number)
	fmt.Println("22.floor_number", floor_number)
	fmt.Println("23.AnnualAggregateTurnover", AnnualAggregateTurnover)
	fmt.Println("24.GrossTotalIncome", GrossTotalIncome)
	fmt.Println("25.nature_of_business_activities", nature_of_business_activities)
	fmt.Println("26.bussiness_address_addl", bussiness_address_addl)
	fmt.Println("27.building_name_addl", building_name_addl)
	fmt.Println("28.street_addl", street_addl)
	fmt.Println("29.location_addl", location_addl)
	fmt.Println("30.door_number_addl", door_number_addl)
	fmt.Println("31.state_name_addl", state_name_addl)
	fmt.Println("32.floor_number_addl", floor_number_addl)
	fmt.Println("33.PercentageOfTaxPaymentInCash", PercentageOfTaxPaymentInCash)
	fmt.Println("34.Aadhar_authentication_status", Aadhar_authentication_status)
	fmt.Println("35.pincode_addl", pincode_addl)
	fmt.Println("36.nature_of_business_activities_addl", nature_of_business_activities_addl)
	fmt.Println("37.gst_insertion_date", gst_insertion_date)
	fmt.Println("38.gst_inserted_by", bussiness_fields_pp_district)
	fmt.Println("39.bussiness_fields_add_district", bussiness_fields_add_district)
	fmt.Println("40.E_KYC_verification_status", E_KYC_verification_status)
	fmt.Println("41.NatureOfCoreBusinessActivity", NatureOfCoreBusinessActivity)
	fmt.Println("42.proprietor_name", proprietor_name)
	fmt.Println("43.Field_visit_conducted", Field_visit_conducted)

	fmt.Println("=====issue identified=========")
	params = append(params, Gstin_number, Legal_Name_of_Business, Centre_Juri, registration_date_nil,
		cancel_date_nil, ConstitutionOfBusiness, NatureOfBusinessActivities, GSTIN_Status,
		mobile, StateJurisdiction, state_juri, email,
		Trade_Name, bussiness_address, location, state_name, pincode, Taxpayer_Type,
		building_name,
		street, door_number, floor_number, AnnualAggregateTurnover, GrossTotalIncome, nature_of_business_activities, bussiness_address_addl, building_name_addl, street_addl, location_addl, door_number_addl,
		state_name_addl, floor_number_addl, PercentageOfTaxPaymentInCash, Aadhar_authentication_status, pincode_addl,
		nature_of_business_activities_addl,
		gst_insertion_date, gst_inserted_by,
		bussiness_fields_pp_district, bussiness_fields_add_district,
		E_KYC_verification_status, NatureOfCoreBusinessActivity, proprietor_name, Field_visit_conducted,business_constitution_group_id)

	fmt.Println("res", res)
	fmt.Println("params", params)
	return res, params

	// fmt.Println("res",res)
	// fmt.Println("params",params)
}

// HSN Information Befisc BusLOogic
func BusLogicOnBefiscHSN_V2(gstin_number string, m map[string]interface{}) (map[string]string, string) {

	var (
		hsncodearray  []string
		hsncodestring string
	)

	res := make(map[string]string)

	fmt.Println(m, "dev-fifthI")

	befischsnkey,ok := m["bzgddtls"].([]interface{})
	if !ok {
		return res,""
	}
	//fmt.Println(reflect.TypeOf(goods), "dev-fifthII")
	fmt.Println(befischsnkey, "Dev-Fifth")
	if len(befischsnkey) > 0 {
		//Fetch the values and put the HSN code onto string
		for _, v := range befischsnkey {
			v1, _ := v.(map[string]interface{})
			hsncode, ok := v1["hsncd"].(string)
			if ok {
				hsncodearray = append(hsncodearray, hsncode)
			}
		}
		hsncodestring = strings.Join(hsncodearray[:], ",")
	} else {
		return res, hsncodestring
	}

	res["hsnstring"] = hsncodestring
	return res, hsncodestring
}


// HSN Information BusLOogic
func BusLogicOnAuthbridgeHSN_V1(gstin_number string, m map[string]interface{}) (map[string]string, string) {

	var (
		hsncodearray  []string
		hsncodestring string
	)

	res := make(map[string]string)

	fmt.Println(m, "dev-fifthI")

	goods := m["goods"].([]interface{})
	//fmt.Println(reflect.TypeOf(goods), "dev-fifthII")
	fmt.Println(goods, "Dev-Fifth")
	if len(goods) > 0 {
		//Fetch the values and put the HSN code onto string
		for _, v := range goods {
			fmt.Println(v)
			v1, _ := v.(map[string]interface{})
			fmt.Println(v1, "Deviiii")
			hsncode, ok := v1["hsn_code"].(string)
			fmt.Println(hsncode, "End")
			if ok {
				hsncodearray = append(hsncodearray, hsncode)
			}
		}
		hsncodestring = strings.Join(hsncodearray[:], ",")
		fmt.Println(hsncodestring, "Dev-Test")
	} else {
		return res, hsncodestring
	}

	res["hsnstring"] = hsncodestring
	return res, hsncodestring
}

func SplitMobileEmail(contact_details string) []string {
	split := strings.Split(contact_details, "<br/>")
	return split
}

// Get Error Params
func GetErrorParams(gst_in string, master_error string) []interface{} {

	var Emap = map[string]int{
		"Client.Timeout exceeded while awaiting headers":       101,
		"The GSTIN passed in the request is invalid":           102,
		"Account plan limit exceeded":                          103,
		"No Record found for the provided Inputs":              104,
		"API Under Maintenance in DC1":                         105,
		"API Under Maintenance in DC2":                         106,
		"No records found":                                     107,
		"GST API DID'NT RESPONDED":                             108,
		"Broken response from GST":                             109,
		"API Authorization Failed":                             110,
		"Access denied.Unauthorize access to GSP":              111,
		"Incorrect Data Format":                                112,
		"invalid character '<' looking for beginning of value": 113,
		"Invalid Auth Token":                                   114,
		"The access token provided is invalid":                 115,
		"Account Validity is finished":                         116,
	}

	var fin_err string
	var fin_code int

	var paramserr []interface{}
	flag := 0
	for key_error, err_code := range Emap {
		if strings.Contains(master_error, key_error) {
			fin_err = key_error
			fin_code = err_code
			flag = 1
			break
		}
		// if err contains the key->
		// Insert 1.gst 2.error_code 3.error_text 4.added_date
	}
	if flag == 0 {
		fin_err = master_error
		fin_code = 117
	}
	added_date_final := time.Now().Format("2006-01-02 15:04:05")
	paramserr = append(paramserr, gst_in, fin_code, fin_err, added_date_final)
	return paramserr
}

func CompanyNameFormatting(name string) string {

	fileds := strings.Fields(name)
	n := len(strings.Fields(name))

	if n == 2 {
		FirstWord := fileds[0]
		SecondWord := fileds[1]
		FinalFirstWord := CheckForFirstWord(FirstWord)
		FinalSecondWord := CheckForSecondThirdWord(SecondWord)
		fileds[0] = FinalFirstWord
		fileds[1] = FinalSecondWord
		// return strings.Join(fileds," ")
	} else if n > 2 {
		FirstWord := fileds[0]
		SecondWord := fileds[1]
		ThirdWord := fileds[2]
		//fmt.Println("Checking length>=3")
		FinalFirstWord := CheckForFirstWord(FirstWord)
		FinalSecondWord := CheckForSecondThirdWord(SecondWord)
		FinalThirdWord := CheckForSecondThirdWord(ThirdWord)
		fileds[0] = FinalFirstWord
		fileds[1] = FinalSecondWord
		fileds[2] = FinalThirdWord
		// return strings.Join(fileds," ")
	}

	return strings.Join(fileds, " ")

}

func CheckForSecondThirdWord(x string) string {
	dup_word := strings.ToLower(x)
	flag := 0
	if len(dup_word) <= 3 {
		flag = 2
		check := []string{"pet", "sky", "art", "of", "and", "sri", "sai", "to", "toy", "raj", "ram", "for", "avi", "uma", "ji", "wen", "hub", "sun", "oil", "air", "bio", "gem", "the", "maa", "you", "eye", "tex", "ka", "new", "sah", "sha", "way", "web", "ads", "tea", "dev", "ply", "new", "go", "fab", "by", "two", "fly", "lab", "sen", "pan", "dye", "pee", "one", "rao", "bee", "big", "das", "car", "box", "di", "bag", "era", "die", "gym", "loo", "mod", "son", "ali", "lal", "one", "pal", "das", "all", "gas", "son", "lab", "rai", "jay", "dal", "lal", "bag", "by", "two", "fly", "lab", "sen", "pan", "dye", "pee", "one", "rao", "bee", "big", "das", "car", "box", "di", "bag", "era", "die", "gym", "loo", "mod", "son", "ali", "lal", "pet", "sky", "art", "of", "and", "sri", "sai", "to", "toy", "raj", "ram", "for", "avi", "uma", "ji", "wen", "hub", "sun", "oil", "air", "bio", "gem", "the", "maa", "you", "eye", "tex", "ka", "new", "sah", "sha", "way", "web", "ads", "tea", "dev", "ply", "new", "go", "fab", "co", "co.", "pre", "pvt", "ltd", "jee", "man", "lok", "saw", "tuk", "mal", "ray", "cap", "fan", "pen", "key", "yes", "out", "age", "ice", "ten", "vet", "net.", "net", "roy", "bar", "zam", "jam", "on", "ki", "ka", "gun"}

		for _, v := range check {

			if v == dup_word {
				flag = 1
				break
			}
		}

	}

	if flag == 2 {
		return strings.ToUpper(x)
	} else {
		return x
	}
}

func CheckForFirstWord(x string) string {
	// dup_word := strings.ToLower(x)
	flag := 0
	if len(x) <= 3 && len(x) >= 2 {

		if len(x) == 2 {
			Ch2_Vowel_or_Not := CheckCharIsVowel(rune(x[1]))

			if !Ch2_Vowel_or_Not {
				flag = 1
			}

		} else if len(x) == 3 {
			Ch2 := rune(x[1])
			Ch3 := rune(x[2])
			Ch2_Vowel_or_Not := CheckCharIsVowel(Ch2)
			Ch3_Vowel_or_Not := CheckCharIsVowel(Ch3)

			if Ch2_Vowel_or_Not || Ch3_Vowel_or_Not {
				flag = 0
			} else {
				flag = 1
			}
		}

	}

	if flag == 1 {
		return strings.ToUpper(x)
	} else {
		return x
	}
}

func CheckCharIsVowel(r rune) bool {
	chars := []rune{'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U'}
	flag := 0
	for _, v := range chars {
		if r == v {
			flag = 1
			break
		}
	}

	if flag == 1 {
		return true
	} else {
		return false
	}
}

func ConvertFirst3LettersToCapital(name string) string {
	// var c, finalString string
	d := ""
	b := strings.Fields(name)
	n := len(strings.Fields(name))

	if n >= 1 {
		if len(b[0]) > 3 {
			return name
		} else {
			b[0] = strings.ToUpper(b[0])
			if n == 1 {
				return b[0]
			} else {
				d = b[0]
				for i := 1; i < n; i++ {
					d = d + " " + b[i]
				}
				finalString := strings.TrimSpace(d)
				return finalString
			}

		}
	} else {
		return ""
	}
}

func NewCompanyNameFormatting(name string) string {

	name = strings.ToLower(name)
	name = strings.TrimSpace(name)

	if len(name) < 1 {
		return name
	}
	name = strings.Title(name)
	fields := strings.Fields(name)
	// n := len(strings.Fields(name))

	// first_word_ignore_flag="0"
	// actual_word:=fileds[0]
	// if strings.ToLower(actual_word)=="m/s"{
	//      first_word_ignore_flag="1"
	// }

	ignoreFirstWord := strings.ToLower(fields[0]) == "m/s"
	startIndex := 0
	if ignoreFirstWord {
		startIndex = 1
	}

	for i := startIndex; i < len(fields); i++ {
		if i == startIndex {
			fields[i] = CheckForFirstWord(fields[i])
		} else {
			fields[i] = CheckForSecondThirdWord(fields[i])
		}
	}

	// if first_word_ignore_flag=="0"{
	//      for i := 0; i < len(fileds); i++ {
	//              if i==0 {
	//                      fileds[i] =CheckForFirstWord(fileds[i])
	//              }else{
	//                      fileds[i] = CheckForSecondThirdWord(fileds[i])
	//              }

	//      }
	// }else{

	//      for i := 1; i < len(fileds); i++ {
	//              if i==1 {
	//                      fileds[i] =CheckForFirstWord(fileds[i])
	//              }else{
	//                      fileds[i] = CheckForSecondThirdWord(fileds[i])
	//              }

	//      }
	// }

	joined_string := strings.Join(fields, " ")

	a := strings.Fields(joined_string)
	for k, j := range a {

		x := ""
		x = j
		x = strings.ToLower(x)
		if x == "llp" || x == "opc" || x == "m/s" || x == "ac" {
			a[k] = strings.ToUpper(a[k])
		}
	}
	return strings.Join(a, " ")

}

func TradeNameNewFormattingLogic(trade_name string, glusr_usr_companyname string) string {
	glusr_usr_companyname = strings.Trim(glusr_usr_companyname, " ")
	lower_glusr_usr_companyname := strings.ToLower(glusr_usr_companyname)
	trade_name = strings.Trim(trade_name, " ")
	outputString := RemoveMSfollowedbbyNonAlpha(trade_name)

	fmt.Println("After Removing MS: ", outputString)

	if len(outputString) == 0 {
		fmt.Println("String has Length 0")
		// return
		return strings.Title(lower_glusr_usr_companyname)
	}

	outputString = strings.ToLower(outputString)
	outputString = strings.TrimSpace(outputString)

	outputString = strings.Title(outputString)
	fields := strings.Split(outputString, " ")

	for i, f := range fields {
		fields[i] = strings.Trim(f, " ")
	}
	initalword := strings.TrimSpace(fields[0])
	initalword = strings.ToLower(initalword)
	// Check if the first field matches any of the given words
	var startIndex int
	switch initalword {
	case "new", "the", "sri", "shri", "shree":
		// If the first word matches, start looping from the second field
		if len(fields) > 1 {
			startIndex = 1
		}
	default:
		// If the first word does not match, start looping from the first field
		startIndex = 0
	}

	for i := startIndex; i < len(fields); i++ {
		if i == startIndex {

			//This is the first word.

			switch len(fields[i]) {
			case 1:
				fields[i] = strings.ToUpper(fields[i])
			case 2:
				// Do something for fields with length 2
				if !containsIgnoreCase([]string{"hi", "we", "om", "at", "oh", "my", "of", "to", "go", "by", "do", "no", "on", "in", "co", "ma", "re", "me", "ji", "up", "us", "ku", "sh", "da", "de", "ye", "di", "ke", "ki", "be", "jo", "ok", "yo", "ni", "ka"}, fields[i]) {
					// Convert the field to uppercase
					fields[i] = strings.ToUpper(fields[i])
				}
				fmt.Println("Processing field", i, "with length 2:", fields[i])
			case 3:
				// Do something for fields with length 3
				if isVowel(fields[i][1]) || isVowel(fields[i][2]) {
					// Convert the first word to title case (except for exception words)

					exceptionWords := map[string]bool{
						"TDI": true, "H2O": true, "MOI": true, "KMI": true, "VIP": true,
						"MNU": true, "DNO": true, "JMA": true, "PVA": true, "MBA": true,
						"MIC": true, "RVA": true, "HSA": true, "ISO": true, "SAS": true,
						"PPI": true, "MGA": true, "TBA": true, "TCI": true, "BTI": true,
						"ATI": true, "DDI": true, "UAS": true, "LII": true, "VSI": true,
						"RFA": true, "AIC": true, "SIC": true, "CPA": true, "EZE": true,
						"SSI": true, "ZOE": true, "MSA": true, "AFO": true, "AIT": true,
						"AIZ": true, "BSA": true, "DBA": true,
						"EBI": true, "EEZ": true, "HBA": true, "HRA": true, "HNU": true,
						"HGI": true, "IAS": true, "IOS": true, "ITO": true, "KSA": true,
						"MDI": true, "MEC": true, "MTI": true, "MPI": true, "MTA": true,
						"NEC": true, "NEG": true, "PDA": true, "PHI": true, "PKA": true,
						"PMI": true, "QAK": true, "RMI": true, "STE": true, "STI": true,
						"TAS": true, "TES": true, "XTO": true, "IFA": true, "IHA": true,
						"KRA": true, "MEP": true, "NGE": true,
						// Add the new words below
						"SAP": true, "VCA": true, "VBE": true, "UTI": true, "TJI": true, "TJE": true, "MIE": true, "MDA": true, "MCI": true, "OPC": true,
					}

					if exceptionWords[strings.ToUpper(fields[i])] {
						fields[i] = strings.ToUpper(fields[i])
					} else {
						fields[i] = strings.ToLower(fields[i])
						fields[i] = strings.Title(fields[i])
					}
				} else {

					exceptionWords := map[string]bool{
						"ALL": true, "SKY": true, "ASK": true, "PLY": true, "OXY": true,
						"DRY": true, "CRY": true, "TRY": true, "ART": true, "OLD": true,
						"FLY": true, "ADD": true, "AMY": true, "ANY": true, "ARC": true,
						"EZY": true, "EXP": true, "WHY": true, "GYM": true, "SPY": true,
						"SHY": true, "FRY": true,
						"OHM": true, "ASH": true, "IND": true, "IVY": true,
					}

					if exceptionWords[strings.ToUpper(fields[i])] {
						fields[i] = strings.ToLower(fields[i])
						fields[i] = strings.Title(fields[i])
					} else {
						fields[i] = strings.ToUpper(fields[i])
					}
				}
			case 4:
				// Do something for fields with length 4
				fmt.Println("Length: ", 4)
				if !containsVowel(fields[i]) && !containsIgnoreCase([]string{"myth", "hymn", "lynx"}, fields[i]) {
					fields[i] = strings.ToUpper(fields[i])
				}

				if containsIgnoreCase([]string{"a2rs", "aace", "aaco", "abcd", "abgk", "absj", "adcl", "hbax", "icti", "iedp", "iiwa", "ndpi", "ncoc", "necs", "nogm", "nsli", "smca", "snra"}, fields[i]) {
					fields[i] = strings.ToUpper(fields[i])
				}

				fmt.Println("Processing field", i, "with length 4:", fields[i])
			default:
				// Do something for fields with length greater than 4
				break
			}
		} else {
			// Do something with the current field
			fmt.Println("Processing field", i, ":", fields[i])

			if len(fields[i]) == 2 {
				if !containsIgnoreCase([]string{"hi", "we", "om", "at", "oh", "my", "of", "to", "go", "by", "do", "no", "on", "in", "co", "ma", "re", "me", "ji", "up", "us", "ku", "sh", "da", "de", "ye", "di", "ke", "ki", "be", "jo", "ok", "yo", "ni", "ka"}, fields[i]) {
					// Convert the field to uppercase
					fields[i] = strings.ToUpper(fields[i])
				}

				if containsIgnoreCase([]string{"co"}, fields[i]) {
					fields[i] = "Co."
				}
			}

			if len(fields[i]) == 3 {

				if !containsVowel(fields[i]) {

					exceptionWords := map[string]bool{
						"SKY": true, "PLY": true, "DRY": true, "CRY": true,
						"TRY": true, "FLY": true, "PVT": true, "MFG": true,
						"BBQ": true, "LTD": true,
					}
					if exceptionWords[strings.ToUpper(fields[i])] {
						fields[i] = strings.ToLower(fields[i])
						fields[i] = strings.Title(fields[i])
					} else {
						fields[i] = strings.ToUpper(fields[i])
					}
				}

				if containsIgnoreCase([]string{"opc", "acp", "vip", "huf"}, fields[i]) {
					fields[i] = strings.ToUpper(fields[i])
				}

				if containsIgnoreCase([]string{"co "}, fields[i]) {
					fields[i] = "Co."
				}

			}

			if len(fields[i]) == 4 {

				if containsIgnoreCase([]string{"a2rs", "aace", "aaco", "abcd", "abgk", "absj", "adcl", "hbax", "icti", "iedp", "iiwa", "ndpi", "ncoc", "necs", "nogm", "nsli", "smca", "snra", "cctv"}, fields[i]) {
					fields[i] = strings.ToUpper(fields[i])
				}

			}

			if len(fields[i]) == 5 {

				if containsIgnoreCase([]string{"(opc)", "(huf)"}, fields[i]) {
					fields[i] = strings.ToUpper(fields[i])
				}

			}

		}
	}

	outputString = strings.Join(fields, " ")

	fmt.Println(outputString)

	modifiedCompName := outputString

	modifiedCompName = removeSpecialCharacters(modifiedCompName)

	fmt.Println("after removing special chars: ", modifiedCompName)

	modifiedCompName = strings.TrimSpace(modifiedCompName)

	CasePvtGLID := 0
	CasePrivateGLID := 0
	CasePvtGST := 0
	CasePrivateGST := 0

	Pvtsubstrings := []string{"Pvt Ltd", "Pvt Ltd.", "Pvt Limited", "Pvt Limited.", "Pvt. Ltd", "Pvt. Ltd.", "Pvt. Limited", "Pvt. Limited.",
		"P Ltd", "P Ltd.", "P Limited", "P Limited.", "P. Ltd", "P. Ltd.", "P. Limited", "P. Limited.",
		"(P) Ltd", "(P) Ltd.", "(P) Limited", "(P) Limited.", "(P). Ltd", "(P). Ltd.", "(P). Limited", "(P). Limited.",
		"PvtLtd", "PvtLtd.", "PvtLimited", "PvtLimited.", "Pvt.Ltd", "Pvt.Ltd.", "Pvt.Limited", "Pvt.Limited.",
		"PLtd", "PLtd.", "PLimited", "PLimited.", "P.Ltd", "P.Ltd.", "P.Limited", "P.Limited.",
		"(P)Ltd", "(P)Ltd.", "(P)Limited", "(P)Limited.", "(P).Ltd", "(P).Ltd.", "(P).Limited", "(P).Limited.",
		"Pvtltd", "Pvtltd.", "Pvtlimited", "Pvtlimited.", "Pvt.Ltd", "Pvt.Ltd.", "Pvt.Limited", "Pvt.Limited.",
		"Pltd", "Pltd.", "Plimited", "Plimited.", "P.Ltd", "P.Ltd.", "P.Limited", "P.Limited.",
		"(P)Ltd", "(P)Ltd.", "(P)Limited", "(P)Limited.", "(P).Ltd", "(P).Ltd.", "(P).Limited", "(P).Limited.",
		"Pvt Ltd", "PvtLtd", "Pvtltd",
		"pvt ltd", "pvt ltd.", "pvt limited", "pvt limited.", "pvt. ltd", "pvt. ltd.", "pvt. limited", "pvt. limited.",
		"p ltd", "p ltd.", "p limited", "p limited.", "p. ltd", "p. ltd.", "p. limited", "p. limited.",
		"(p) ltd", "(p) ltd.", "(p) limited", "(p) limited.", "(p). ltd", "(p). ltd.", "(p). limited", "(p). limited.",
		"pvtltd", "pvtltd.", "pvtlimited", "pvtlimited.", "pvt.ltd", "pvt.ltd.", "pvt.limited", "pvt.limited.",
		"pltd", "pltd.", "plimited", "plimited.", "p.ltd", "p.ltd.", "p.limited", "p.limited.",
		"(p)ltd", "(p)ltd.", "(p)limited", "(p)limited.", "(p).ltd", "(p).ltd.", "(p).limited", "(p).limited.",
		"pvtltd", "pvtltd.", "pvtlimited", "pvtlimited.", "pvt.ltd", "pvt.ltd.", "pvt.limited", "pvt.limited.",
		"pltd", "pltd.", "plimited", "plimited.", "p.ltd", "p.ltd.", "p.limited", "p.limited.",
		"(p)ltd", "(p)ltd.", "(p)limited", "(p)limited.", "(p).ltd", "(p).ltd.", "(p).limited", "(p).limited.",
		"pvt ltd", "pvtltd", "pvtltd",
	}

	Privatesubstrings := []string{"Private Ltd",
		"Private Ltd.",
		"Private Limited",
		"Private. Ltd",
		"Private. Ltd.",
		"Private. Limited",
		"Private. Limited.",
		"PrivateLtd",
		"PrivateLtd.",
		"PrivateLimited",
		"PrivateLimited.",
		"Private.Ltd",
		"Private.Ltd.",
		"Private.Limited",
		"Private.Limited.",
		"Privateltd",
		"Privateltd.",
		"Privatelimited",
		"Privatelimited.",
		"Private.Ltd",
		"Private.Ltd.",
		"Private.Limited",
		"Private.Limited.",
		"private ltd",
		"private ltd.",
		"private limited",
		"private. ltd",
		"private. ltd.",
		"private. limited",
		"private. limited.",
		"privateltd",
		"privateltd.",
		"privatelimited",
		"privatelimited.",
		"private.ltd",
		"private.ltd.",
		"private.limited",
		"private.limited.",
	}

	fmt.Println("Before:")

	fmt.Println("CasePvtGLID: ", CasePvtGLID)
	fmt.Println("CasePrivateGLID: ", CasePrivateGLID)
	fmt.Println("CasePvtGST: ", CasePvtGST)
	fmt.Println("CasePrivateGST: ", CasePrivateGST)
	// fmt.Println("GLID: ", glid)

	for _, substring := range Pvtsubstrings {
		if strings.Contains(lower_glusr_usr_companyname, substring) {
			CasePvtGLID = 1
		}

		if strings.Contains(modifiedCompName, substring) {
			CasePvtGST = 1
		}
	}

	for _, substring := range Privatesubstrings {
		if strings.Contains(lower_glusr_usr_companyname, substring) {
			CasePrivateGLID = 1
		}
		if strings.Contains(modifiedCompName, substring) {
			CasePrivateGST = 1
		}
	}

	fmt.Println("After:")

	fmt.Println("CasePvtGLID: ", CasePvtGLID)
	fmt.Println("CasePrivateGLID: ", CasePrivateGLID)
	fmt.Println("CasePvtGST: ", CasePvtGST)
	fmt.Println("CasePrivateGST: ", CasePrivateGST)
	// fmt.Println("GLID: ", glid)

	if CasePvtGLID == 1 {
		for _, substring := range Pvtsubstrings {
			if strings.Contains(modifiedCompName, substring) {
				index := strings.Index(modifiedCompName, substring)
				modifiedCompName = modifiedCompName[:index] + "Pvt. Ltd." + modifiedCompName[index+len(substring):]
				fmt.Println("modifiedCompanyName while changing CasePvtGLID : ", modifiedCompName)
			}
		}
		for _, substring := range Privatesubstrings {
			if strings.Contains(modifiedCompName, substring) {
				index := strings.Index(modifiedCompName, substring)
				modifiedCompName = modifiedCompName[:index] + "Pvt. Ltd." + modifiedCompName[index+len(substring):]
				fmt.Println("modifiedCompanyName while changing CasePvtGLID : ", modifiedCompName)
			}
		}

	} else if CasePrivateGLID == 1 {
		for _, substring := range Privatesubstrings {
			if strings.Contains(modifiedCompName, substring) {
				index := strings.Index(modifiedCompName, substring)
				modifiedCompName = modifiedCompName[:index] + "Private Limited" + modifiedCompName[index+len(substring):]
				fmt.Println("modifiedCompanyName while changing CasePrivateGLID : ", modifiedCompName)
			}
		}
		for _, substring := range Pvtsubstrings {
			if strings.Contains(modifiedCompName, substring) {
				index := strings.Index(modifiedCompName, substring)
				modifiedCompName = modifiedCompName[:index] + "Private Limited" + modifiedCompName[index+len(substring):]
				fmt.Println("modifiedCompanyName while changing CasePrivateGLID : ", modifiedCompName)
			}
		}

	} else if CasePrivateGST == 1 {

		for _, substring := range Privatesubstrings {
			if strings.Contains(modifiedCompName, substring) {
				index := strings.Index(modifiedCompName, substring)
				modifiedCompName = modifiedCompName[:index] + "Private Limited" + modifiedCompName[index+len(substring):]
				fmt.Println("modifiedCompanyName while changing CasePrivateGST : ", modifiedCompName)
			}
		}
		for _, substring := range Pvtsubstrings {
			if strings.Contains(modifiedCompName, substring) {
				index := strings.Index(modifiedCompName, substring)
				modifiedCompName = modifiedCompName[:index] + "Private Limited" + modifiedCompName[index+len(substring):]
				fmt.Println("modifiedCompanyName while changing CasePrivateGST : ", modifiedCompName)
			}
		}
	} else if CasePvtGST == 1 {
		for _, substring := range Pvtsubstrings {
			if strings.Contains(modifiedCompName, substring) {
				index := strings.Index(modifiedCompName, substring)
				modifiedCompName = modifiedCompName[:index] + "Pvt. Ltd." + modifiedCompName[index+len(substring):]
				fmt.Println("modifiedCompanyName while changing CasePvtGST : ", modifiedCompName)
			}

		}
		for _, substring := range Privatesubstrings {
			if strings.Contains(modifiedCompName, substring) {
				index := strings.Index(modifiedCompName, substring)
				modifiedCompName = modifiedCompName[:index] + "Pvt. Ltd." + modifiedCompName[index+len(substring):]
				fmt.Println("modifiedCompanyName while changing CasePvtGST : ", modifiedCompName)
			}
		}

	}

	if strings.Contains(modifiedCompName, "Pvt. Ltd..") {
		index := strings.Index(modifiedCompName, "Pvt. Ltd..")
		modifiedCompName = modifiedCompName[:index] + "Pvt. Ltd." + modifiedCompName[index+len("Pvt. Ltd.."):]
	}

	if strings.Contains(modifiedCompName, "Private Limited.") {
		index := strings.Index(modifiedCompName, "Private Limited.")
		modifiedCompName = modifiedCompName[:index] + "Private Limited" + modifiedCompName[index+len("Private Limited."):]
	}

	// modifiedCompName = replaceOpcIgnoreCase(modifiedCompName)

	modifiedCompName = strings.TrimSpace(modifiedCompName)

	// //

	// Split the string into individual words
	words := strings.Split(modifiedCompName, " ")

	// Get the last word from the slice
	lastWord := words[len(words)-1]

	// Check if the last word has 2 or 3 characters
	if len(lastWord) == 2 || len(lastWord) == 3 {
		// Check if the last word contains "co" or "ltd" (case-insensitive)
		lowerLastWord := strings.ToLower(lastWord)
		if strings.Contains(lowerLastWord, "co") {
			lastWord = "Co."
		} else if strings.Contains(lowerLastWord, "ltd") {
			lastWord = "Ltd."
		}
	}

	// Update the modified last word in the words slice
	words[len(words)-1] = lastWord

	// Join the modified words slice back into a string
	modifiedCompName = strings.Join(words, " ")

	fmt.Println("FInal Modified CompnayName: ", modifiedCompName)

	return modifiedCompName

}
func containsIgnoreCase(slice []string, s string) bool {
	for _, v := range slice {
		if strings.EqualFold(v, s) {
			return true
		}
	}
	return false
}

func containsVowel(s string) bool {
	vowels := "aeiouAEIOU"
	for _, c := range s {
		if strings.ContainsRune(vowels, c) {
			return true
		}
	}
	return false
}

func isVowel(c byte) bool {
	switch c {
	case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
		return true
	default:
		return false
	}
}
func RemoveMSfollowedbbyNonAlpha(inputString string) string {
	// Return the input string if it is empty
	inputString = strings.TrimSpace(inputString)
	inputString = strings.ToLower(inputString)
	if len(inputString) == 0 {
		return inputString
	}

	// Compile the regex pattern
	pattern := regexp.MustCompile(`^m\/s[^a-zA-Z0-9]*`)

	// Split the input string into words
	words := strings.Fields(inputString)

	// Return the input string if there are no words
	if len(words) == 0 {
		return inputString
	}

	// Apply the regex pattern to the first word and replace it with an empty string
	words[0] = pattern.ReplaceAllString(words[0], "")

	// Join the words back together
	outputString := strings.Join(words, " ")

	// Trim any leading and trailing spaces
	outputString = strings.TrimSpace(outputString)

	return outputString
}

func removeSpecialCharacters(s string) string {
	i := len(s)

	for i > 0 {
		r := rune(s[i-1])
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == ')' || r == ']' || r == '}' || r == '>' {
			break
		}
		i--
	}

	return s[:i]
}

func replaceOpcIgnoreCase(s string) string {
	opcPattern := "(?i)opc"
	re := regexp.MustCompile(opcPattern)

	return re.ReplaceAllStringFunc(s, func(match string) string {
		return strings.ToUpper(match)
	})
}

func IsSpecialOrNonNumericASCII(c rune) bool {
	// Check if character is not a digit and not a letter (ASCII)
	return !(c >= '0' && c <= '9') && !(c >= 'A' && c <= 'Z') && !(c >= 'a' && c <= 'z')
}

func AddressNewFormattingLogic(glusr_usr_add string) string {
	// Split the address into words and separators using a regex that captures both
	glusr_usr_add = strings.Title(strings.ToLower(glusr_usr_add))
	words, separators := SplitWithSeparators(glusr_usr_add)

	if len(words) == 0 {
		fmt.Println("String has Length 0")
		return ""
	}

	// Apply formatting logic to each word
	var startIndex int
	initialWord := strings.ToLower(strings.TrimSpace(words[0]))
	switch initialWord {
	case "new", "the", "sri", "shri", "shree":
		// If the first word matches specific prefixes, start from the second word
		if len(words) > 1 {
			startIndex = 1
		}
	default:
		// Start from the first word if no prefix match
		startIndex = 0
	}

	// Apply formatting to each word
	for i := startIndex; i < len(words); i++ {

		if unicode.IsDigit(rune(words[i][0])) {
			words[i] = strings.Title(strings.ToLower(words[i]))
			continue
		}

		if i == startIndex {
			// Special formatting for the first significant word

			switch len(words[i]) {
			case 1:
				words[i] = strings.ToUpper(words[i])
			case 2:
				if !containsIgnoreCase([]string{"hi", "we", "om", "at", "oh", "my", "of", "to", "go", "by", "do", "no", "on", "in", "co", "ma", "re", "me", "ji", "up", "us", "ku", "sh", "da", "de", "ye", "di", "ke", "ki", "be", "jo", "ok", "yo", "ni", "ka", "sy"}, words[i]) {
					// Convert the field to uppercase
					words[i] = strings.ToUpper(words[i])
				}
			case 3:
				if isVowel(words[i][1]) || isVowel(words[i][2]) {
					// Convert the first word to title case (except for exception words)
					exceptionWords := map[string]bool{
						"TDI": true, "H2O": true, "MOI": true, "KMI": true, "VIP": true,
						"MNU": true, "DNO": true, "JMA": true, "PVA": true, "MBA": true,
						"MIC": true, "RVA": true, "HSA": true, "ISO": true, "SAS": true,
						"PPI": true, "MGA": true, "TBA": true, "TCI": true, "BTI": true,
						"ATI": true, "DDI": true, "UAS": true, "LII": true, "VSI": true,
						"RFA": true, "AIC": true, "SIC": true, "CPA": true, "EZE": true,
						"SSI": true, "ZOE": true, "MSA": true, "AFO": true, "AIT": true,
						"AIZ": true, "BSA": true, "DBA": true,
						"EBI": true, "EEZ": true, "HBA": true, "HRA": true, "HNU": true,
						"HGI": true, "IAS": true, "IOS": true, "ITO": true, "KSA": true,
						"MDI": true, "MEC": true, "MTI": true, "MPI": true, "MTA": true,
						"NEC": true, "NEG": true, "PDA": true, "PHI": true, "PKA": true,
						"PMI": true, "QAK": true, "RMI": true, "STE": true, "STI": true,
						"TAS": true, "TES": true, "XTO": true, "IFA": true, "IHA": true,
						"KRA": true, "MEP": true, "NGE": true,
						// Add the new words below
						"SAP": true, "VCA": true, "VBE": true, "UTI": true, "TJI": true, "TJE": true, "MIE": true, "MDA": true, "MCI": true, "OPC": true,
					}

					if exceptionWords[strings.ToUpper(words[i])] {
						words[i] = strings.ToUpper(words[i])
					} else {
						words[i] = strings.ToLower(words[i])
						words[i] = strings.Title(words[i])
					}

				} else {

					exceptionWords := map[string]bool{
						"ALL": true, "SKY": true, "ASK": true, "PLY": true, "OXY": true,
						"DRY": true, "CRY": true, "TRY": true, "ART": true, "OLD": true,
						"FLY": true, "ADD": true, "AMY": true, "ANY": true, "ARC": true,
						"EZY": true, "EXP": true, "WHY": true, "GYM": true, "SPY": true,
						"SHY": true, "FRY": true,
						"OHM": true, "ASH": true, "IND": true, "IVY": true,
					}

					if exceptionWords[strings.ToUpper(words[i])] {
						words[i] = strings.ToLower(words[i])
						words[i] = strings.Title(words[i])
					} else {
						words[i] = strings.ToUpper(words[i])
					}
				}
			case 4:
				if !containsVowel(words[i]) && !containsIgnoreCase([]string{"myth", "hymn", "lynx"}, words[i]) {
					words[i] = strings.ToUpper(words[i])
				}

				if containsIgnoreCase([]string{"a2rs", "aace", "aaco", "abcd", "abgk", "absj", "adcl", "hbax", "icti", "iedp", "iiwa", "ndpi", "ncoc", "necs", "nogm", "nsli", "smca", "snra"}, words[i]) {
					words[i] = strings.ToUpper(words[i])
				}
			default:
				// Do something for fields with length greater than 4
				break

			}

		} else {

			if len(words[i]) == 2 {
				if !containsIgnoreCase([]string{"hi", "we", "om", "at", "oh", "my", "of", "to", "go", "by", "do", "no", "on", "in", "co", "ma", "re", "me", "ji", "up", "us", "ku", "sh", "da", "de", "ye", "di", "ke", "ki", "be", "jo", "ok", "yo", "ni", "ka", "sy"}, words[i]) {
					// Convert the field to uppercase
					words[i] = strings.ToUpper(words[i])
				}

				if containsIgnoreCase([]string{"co"}, words[i]) {
					words[i] = "Co."
				}
			}

			if len(words[i]) == 3 {

				if !containsVowel(words[i]) {

					exceptionWords := map[string]bool{
						"SKY": true, "PLY": true, "DRY": true, "CRY": true,
						"TRY": true, "FLY": true, "PVT": true, "MFG": true,
						"BBQ": true, "LTD": true,
					}
					if exceptionWords[strings.ToUpper(words[i])] {
						words[i] = strings.ToLower(words[i])
						words[i] = strings.Title(words[i])
					} else {
						words[i] = strings.ToUpper(words[i])
					}
				}

				if containsIgnoreCase([]string{"opc", "acp", "vip", "huf"}, words[i]) {
					words[i] = strings.ToUpper(words[i])
				}

				if containsIgnoreCase([]string{"co "}, words[i]) {
					words[i] = "Co"
				}

			}

			if len(words[i]) == 4 {

				if containsIgnoreCase([]string{"a2rs", "aace", "aaco", "abcd", "abgk", "absj", "adcl", "hbax", "icti", "iedp", "iiwa", "ndpi", "ncoc", "necs", "nogm", "nsli", "smca", "snra", "cctv"}, words[i]) {
					words[i] = strings.ToUpper(words[i])
				}

			}

			if len(words[i]) == 5 {

				if containsIgnoreCase([]string{"(opc)", "(huf)"}, words[i]) {
					words[i] = strings.ToUpper(words[i])
				}

			}

		}
	}

	// Reassemble the formatted words with the original separators.
	formattedAddress := ReassembleWithSeparators(words, separators)
	//fmt.Println("Formatted Address:", formattedAddress)
	return formattedAddress

}

// Reassemble the words with the captured separators to maintain the original format.
func ReassembleWithSeparators(words []string, separators []string) string {
	var combined strings.Builder
	wordsLen := len(words)
	separatorsLen := len(separators)

	// Combine words with their corresponding separators.
	for i := 0; i < wordsLen; i++ {
		combined.WriteString(words[i])
		if i < separatorsLen {
			combined.WriteString(separators[i])
		}
	}

	return combined.String()
}

// Split the address into words and separators, capturing both for precise reassembly.
func SplitWithSeparators(address string) ([]string, []string) {
	// Regex pattern to match spaces, commas, and comma-space combinations.
	re := regexp.MustCompile(`(\s*,\s*|\s+|,)`)
	// Split the address and capture separators.
	parts := re.Split(address, -1)
	separators := re.FindAllString(address, -1)

	// Filter out empty strings from words.
	words := FilterEmpty(parts)
	return words, separators
}

// Filter out empty strings from the slice of words.
func FilterEmpty(words []string) []string {
	var filtered []string
	for _, word := range words {
		if word != "" {
			filtered = append(filtered, word)
		}
	}
	return filtered
}

// Function to get the 6th character from the GST number
func extractSixthChar(gst string) (rune, error) {
	if len(gst) < 6 {
		return 0, errors.New("GST number is too short")
	}
	return rune(gst[5]), nil // Index 5 corresponds to the 6th character
}

// Function to determine the legal status and legal status ID
func GetLegalStatus(gst string) (string, int, error) {
	// Extract the 6th character from the GST
	sixthChar, err := extractSixthChar(gst)
	if err != nil {
		return "Others", 1927, err
	}

	// Determine legal status based on the 6th character
	switch sixthChar {
	case 'P', 'H':
		return "Proprietorship", 1924, nil
	case 'F':
		return "Partnership", 1925, nil
	case 'C':
		return "Limited Company", 1926, nil
	default:
		return "Others", 1927, nil
	}
}


// LegalStatus returns the legal status string based on the provided legalStatusID
func LegalStatusRead(legalStatusID string) string {
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
	return "Others"
}