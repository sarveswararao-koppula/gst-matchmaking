package livematchmakingapi

import (
	// "bufio"   //not needed
	"bytes"
	"fmt"
	"math"
	"mm/components/constants"

	// "os"     //not needed
	"encoding/json"
	"errors"
	"io/ioutil"
	servapi "mm/api/servapi"
	utils "mm/utils"
	"net/http"
	"strings"
	"time"
)

// const RemoteHost = "107.22.229.251"
const RemoteHost = "65.0.217.127"

// BIValidationKeyFromSOA ... validaiton key for modid:BI from SOA
const BIValidationKeyFromSOA = "af7f0273997b9b290bd7c57aa19f36c2"

// Mapping of attribute names to their respective keys
var attributeMap = map[string]string{
	"GLUSR_USR_PH_MOBILE":            "M1",
	"GLUSR_USR_PH2_NUMBER":           "L2",
	"GLUSR_USR_PH_MOBILE_ALT":        "M2",
	"GLUSR_USR_PH_NUMBER":            "L1",
	"GLUSR_USR_ADDT_PH_MOBILE":       "Ma1",
	"GLUSR_USR_ADDT_PH_NUMBER":       "La1",
	"GLUSR_USR_ADDT_TOLLFREE_NUMBER": "T1",
}

// Reverse mapping of keys to attributes for easy lookup
var keyToAttribute = map[string]int{
	"M1":  121,
	"L2":  156,
	"M2":  48,
	"L1":  120,
	"Ma1": 1293,
	"La1": 1294,
	"T1":  2074,
}

func maxDistance(len1, len2 int) int {
	maxL := max(len1, len2)
	maxDist := math.Floor(float64(maxL)/2) - 1
	return int(maxDist)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func clearStr(str string) string {
	str = strings.ToLower(str)
	return strings.Join(strings.Fields(str), " ")
}

func jaroDistance(s1 string, s2 string) float64 {

	//if s1 == s2 {
	//return 1.0
	//}

	len1 := len(s1)
	len2 := len(s2)

	if len1 == 0 || len2 == 0 {
		return 0.0
	}

	if s1 == s2 {
		return 1.0
	}

	maxDist := maxDistance(len1, len2)

	var (
		match  float64 = 0
		trans  float64 = 0
		result float64 = 0
	)

	var hashS1, hashS2 map[int]bool = make(map[int]bool), make(map[int]bool)

	for i := 0; i < len1; i++ {

		for j := max(0, i-maxDist); j < min(len2, i+maxDist+1); j++ {

			val1 := string(s1[i])
			val2 := string(s2[j])

			if val1 == val2 && !hashS2[j] {
				hashS1[i] = true
				hashS2[j] = true
				match++
				break
			}

		}
	}

	if match == 0 {
		return 0.0
	}

	point := 0

	for i := 0; i < len1; i++ {
		if hashS1[i] {
			for !hashS2[point] {
				point++
			}

			if string(s1[i]) != string(s2[point]) {
				trans++
			}

			point++
		}
	}

	trans /= 2

	result = match / float64(len1)
	result += match / float64(len2)
	result += (match - trans) / match

	result /= 3.0

	return result
}

func jaroWinkler(s1 string, s2 string) float64 {
	s1 = clearStr(s1)
	s2 = clearStr(s2)
	jaroDist := jaroDistance(s1, s2)
	//return
	if jaroDist > 0.7 {

		prefix := 0
		for i := 0; i < min(len(s1), len(s2)); i++ {
			if string(s1[i]) == string(s2[i]) {
				prefix++
			} else {
				break
			}
		}

		prefix = min(prefix, 4)

		jaroDist += 0.1 * float64(prefix) * (1 - jaroDist)
	}

	jaroDist = float64(int(jaroDist*100)) / 100
	return jaroDist
}

func Pnsapicall(glid string, env string) (map[string]map[string]string, error) {
	var AK string
	AK = "eyJ0eXAiOiJKV1QiLCJhbGciOiJzaGEyNTYifQ.eyJpc3MiOiJDUk9OIiwiYXVkIjoiNDMuMjA1LjQzLjgxLDEwLjEwLjEwLjIwIiwiZXhwIjoxODMzMTA1NDg2LCJpYXQiOjE2NzU0MDU0ODYsInN1YiI6ImJpLXV0aWxzLmludGVybWVzaC5uZXQifQ.qzIQ8EVfX9bNz52oLL3btBha1FANr8grVyLr6o1DzMs"
	urlAPI := ""

	if env == "DEV" {
		urlAPI = "http://34.93.67.39/wservce/users/pnssetting/"
	} else if env == "PROD" {
		urlAPI = "http://users.imutils.com/wservce/users/pnssetting/"
	}

	glusr_usr_id := glid

	req, err := http.NewRequest("POST", urlAPI, strings.NewReader("token=imobile@15061981&AK="+AK+"&modid=BI&glusrid="+glusr_usr_id))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	var pnssettingresult map[string]interface{}
	err = json.Unmarshal(body, &pnssettingresult)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	// Check if the status is Success
	if status, ok := pnssettingresult["Response"].(map[string]interface{})["Status"].(string); ok && status == "Success" {
		numbers := extractPhoneNumbers(pnssettingresult)
		attributes := extractAttributes(numbers)
		end := time.Now()
		fmt.Printf("Total time taken: %v\n", end.Sub(start))
		return attributes, nil
	} else {
		return nil, errors.New("API error: Status is not Success")
	}
}

func extractPhoneNumbers(data map[string]interface{}) map[string]string {
	numbers := make(map[string]string)
	if response, ok := data["Response"].(map[string]interface{}); ok {
		if data, ok := response["Data"].(map[string]interface{}); ok {
			for key, v := range data {
				if vMap, ok := v.(map[string]interface{}); ok {
					if number, ok := vMap["number"].(string); ok && number != "" {
						numbers[key] = number
					}
				}
			}
		}
	}
	return numbers
}

func extractAttributes(numbers map[string]string) map[string]map[string]string {
	attributes := make(map[string]map[string]string)
	for key, number := range numbers {
		if attr, exists := keyToAttribute[key]; exists {
			attributes[key] = map[string]string{
				"number":    number,
				"attribute": fmt.Sprintf("%d", attr),
			}
		}
	}
	return attributes
}

func matchMobileWithAttributes(attributes map[string]map[string]string, mobile string) (string, string, bool) {
	for _, details := range attributes {
		if details["number"] == mobile {
			return details["attribute"], details["number"], true
		}
	}
	return "", "", false
}

func matchEmailID(GST_email_id string, user User) (string, string) {

	GST_email_id_lower := strings.ToLower(GST_email_id)
	glusr_usr_email_lower := strings.ToLower(user.glusr_usr_email)
	glusr_usr_email_alt_lower := strings.ToLower(user.glusr_usr_email_alt)

	if GST_email_id_lower == glusr_usr_email_lower {
		return GST_email_id, "109" // Attribute ID for glusr_usr_email
	} else if GST_email_id_lower == glusr_usr_email_alt_lower {
		return GST_email_id, "157" // Attribute ID for glusr_usr_email_alt
	} else {
		return "No match found", "-1"
	}
}

// AttributeVerified ... check if gst already verified
func AttributeVerified(glid string, attributeid string) (string, error) {

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	token := "imobile@15061981"
	modid := "BI"
	attrID := attributeid
	url := "http://users.imutils.com/wservce/users/verifiedDetail/?token=" + token + "&modid=" + modid + "&glusrid=" + glid + "&attribute_id=" + attrID + "&AK=" + constants.ServerAK

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	bodyString := string(bodyBytes)
	//fmt.Println(bodyString)
	return bodyString, nil
}

// IsAttributeAlreadyVerified ...checking if gst is already verified for glid
func IsAttributeAlreadyVerified(glid string, attributeid string) (bool, error) {
	jsonStr, err := AttributeVerified(glid, attributeid)

	if err != nil {
		return false, err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return false, err
	}

	res := make(map[string]interface{})
	for k, v := range data {
		res[strings.ToLower(k)] = v
	}

	k, _ := res["response"].(map[string]interface{})
	k, _ = k["Data"].(map[string]interface{})
	k, _ = k[attributeid].(map[string]interface{})
	status, _ := k["Status"].(string)

	if strings.ToLower(status) == "verified" {
		return true, nil
	}

	if strings.ToLower(status) == "not verified" {
		return false, nil
	}

	return false, errors.New(jsonStr)
}

func UpdateAdditionAddressDetails(env, glid, section, add1, add2, screenName string) error {

	jsonStr, err := DetailsApiForAdditionAddress(env, glid, section, add1, add2, screenName)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return err
	}

	res := make(map[string]interface{})
	for k, v := range data {
		res[strings.ToLower(k)] = v
	}

	status, _ := res["status"].(string)

	if strings.ToLower(status) == "successful" {
		return nil
	}

	return errors.New(jsonStr)
}

// Details ... Update addition address
func DetailsApiForAdditionAddress(env, glid, section, add1, add2, screenName string) (string, error) {

	client := &http.Client{
		Timeout: 4 * time.Second,
	}
	m := make(map[string]string)

	m["glusridval"] = glid
	m["section"] = section
	if len(add1) > 0 {
		m["add1"] = add1
	}
	if len(add2) > 0 {
		m["add2"] = add2
	}
	m["flag"] = "I"
	m["updatedby"] = "GST Tech Process"
	m["updatedbyId"] = "85344"
	m["updatedbyScreen"] = screenName
	m["userIp"] = RemoteHost
	m["userIpCoun"] = "INDIA"
	m["VALIDATION_KEY"] = BIValidationKeyFromSOA
	m["type"] = "ContactDetails"
	m["histComment"] = "By " + screenName + " Process"
	m["AK"] = constants.ServerAK

	url := ""
	if env == "DEV" {
		url = "http://dev-service.intermesh.net/details"
	} else if env == "PROD" {
		url = "http://service.intermesh.net/details"
	}

	reqBody, _ := json.Marshal(m)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))

	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	bodyString := string(bodyBytes)

	return bodyString, nil
}

// TitleCase converts a string to title case, ensuring the first letter of each word is capitalized.
func TitleCase(s string) string {
	return strings.Title(strings.ToLower(s))
}

// SplitAddress splits the given address into two parts (Line1 and Line2) without breaking words. The maxCount parameter specifies the maximum number of characters allowed for each line.
// The function returns Line1 and Line2 in title case format.
func SplitAddress(address string, maxCount int) (string, string) {
	// Trim the address and return it if it fits within maxCount.
	if len(address) <= maxCount {
		return utils.AddressNewFormattingLogic(strings.TrimSpace(address)), ""
	}

	// Function to find the last space or comma within the given limit.
	findSplitIndex := func(text string, limit int) int {
		splitIndex := limit
		for i := splitIndex; i >= 0; i-- {
			if text[i] == ' ' || text[i] == ',' {
				return i
			}
		}
		return splitIndex
	}

	// Determine where to split the address for Line1.
	splitIndex := findSplitIndex(address, maxCount)
	Line1 := address[:splitIndex]

	// Trim spaces and commas from Line1 and the remaining address.
	if splitIndex < len(address) {
		remainingAddress := address[splitIndex:]
		remainingAddress = strings.TrimSpace(remainingAddress)

		// Trim spaces and remove a trailing comma from Line1.
		Line1 = strings.TrimSpace(Line1)
		if strings.HasSuffix(Line1, ",") {
			Line1 = strings.TrimSuffix(Line1, ",")
		}

		// Remove a leading comma from the remaining address if present.
		if strings.HasPrefix(remainingAddress, ",") {
			remainingAddress = strings.TrimPrefix(remainingAddress, ",")
		}

		// If the remaining address fits within maxCount, set it as Line2.
		if len(remainingAddress) <= maxCount {
			return utils.AddressNewFormattingLogic(Line1), utils.AddressNewFormattingLogic(remainingAddress)
		}

		// Otherwise, determine where to split the remaining address for Line2.
		splitIndex = findSplitIndex(remainingAddress, maxCount)
		Line2 := remainingAddress[:splitIndex]
		Line2 = strings.TrimSpace(Line2)

		// Remove a leading comma from Line2 if present.
		if strings.HasPrefix(Line2, ",") {
			Line2 = strings.TrimPrefix(Line2, ",")
		}

		// Return both lines in title case.
		return utils.AddressNewFormattingLogic(Line1), utils.AddressNewFormattingLogic(Line2)
	}

	// If there's no remaining address, return Line1 after trimming.
	Line1 = strings.TrimSpace(Line1)
	if strings.HasSuffix(Line1, ",") {
		Line1 = strings.TrimSuffix(Line1, ",")
	}

	return utils.AddressNewFormattingLogic(Line1), ""
}

// marshalToString converts an interface to a JSON string, returning "{}" on error or nil value.
func MarshalToString(v interface{}) string {
	if v == nil {
		return "{}"
	}
	jsonData, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(jsonData)
}

// CreateLogEntry creates a LogEntry from the provided Logg struct.
func CreateLogEntry(logg Logg) LogEntry {
	return LogEntry{
		RequestStart:            logg.RequestStart,
		ResponseTime_Float:      logg.ResponseTime,
		ServiceName:             logg.ServiceName,
		ServiceURL:              logg.ServiceURL,
		RemoteAddress:           logg.RemoteAddress,
		Request_Data:            MarshalToString(logg.Request),
		Response_Body:           MarshalToString(logg.Response),
		Any_Error:               MarshalToString(logg.AnyError),
		StackTrace:              logg.StackTrace,
		Glusr_usr_id:            logg.Request.Glid,
		Gst:                     logg.Request.GST,
		Modid:                   logg.Request.ModID,
		ValidationKey:           logg.Request.ValidationKey,
		TacticalAttributeSource: logg.TacticalAttributeSource,
		User_verification_date:  logg.User_verification_date,
	}
}

// CreatePrimaryAddress creates a primary address by combining various address components.
// It omits the Duplicate parts , as per the business rules.
func CreatePrimaryAddress(dno, flno, bn, streetname, loc, gl_city_name, gl_district_name, landmark, locality string) string {
	// Trim spaces and convert all inputs to lowercase
	dno = strings.TrimSpace(strings.ToLower(dno))
	flno = strings.TrimSpace(strings.ToLower(flno))
	bn = strings.TrimSpace(strings.ToLower(bn))
	streetname = strings.TrimSpace(strings.ToLower(streetname))
	loc = strings.TrimSpace(strings.ToLower(loc))
	gl_city_name = strings.TrimSpace(strings.ToLower(gl_city_name))
	gl_district_name = strings.TrimSpace(strings.ToLower(gl_district_name))
	locality = strings.TrimSpace(strings.ToLower(locality))
	landmark = strings.TrimSpace(strings.ToLower(landmark))
	// Normalize city names
	// if gl_city_name == "delhi" || gl_city_name == "new delhi" {
	// 	gl_city_name = "delhi"
	// }
	if gl_city_name == "puducherry" || gl_city_name == "pondicherry" {
		gl_city_name = "puducherry"
	}
	if gl_city_name == "bengaluru" || gl_city_name == "bangalore" {
		gl_city_name = "bangalore"
	}

	// if gl_district_name == "delhi" || gl_district_name == "new delhi" {
	// 	gl_district_name = "delhi"
	// }
	if gl_district_name == "puducherry" || gl_district_name == "pondicherry" {
		gl_district_name = "puducherry"
	}
	if gl_district_name == "bengaluru" || gl_district_name == "bangalore" {
		gl_district_name = "bangalore"
	}

	// if loc == "delhi" || loc == "new delhi" {
	// 	loc = "delhi"
	// }
	if loc == "puducherry" || loc == "pondicherry" {
		loc = "puducherry"
	}
	if loc == "bengaluru" || loc == "bangalore" {
		loc = "bangalore"
	}

	var addressParts []string

	// Function to check if a part should be added
	shouldAdd := func(part string) bool {
		if part == "" {
			return false
		}
		// Compare against existing addressParts
		for _, existing := range addressParts {
			if strings.EqualFold(existing, part) {
				return false
			}
		}
		// Compare against city and district
		if strings.EqualFold(part, gl_city_name) || strings.EqualFold(part, gl_district_name) {
			return false
		}
		return true
	}

	// Add in order with checks
	if shouldAdd(flno) {
		addressParts = append(addressParts, flno)
	}
	if shouldAdd(dno) {
		addressParts = append(addressParts, dno)
	}
	if shouldAdd(bn) {
		addressParts = append(addressParts, bn)
	}
	if shouldAdd(streetname) {
		addressParts = append(addressParts, streetname)
	}
	if shouldAdd(landmark) {
		addressParts = append(addressParts, landmark)
	}
	if shouldAdd(locality) {
		addressParts = append(addressParts, locality)
	}
	if shouldAdd(loc) {
		addressParts = append(addressParts, loc)
	}

	// Concatenate the final address with commas
	AddressLine1 := strings.Join(addressParts, ", ")

	return AddressLine1

}

func CreateSecondaryAddress(add1, add2, pincode, city string) string {
	// Create a slice of all the address parts
	parts := []string{add1, add2, pincode, city}

	// Filter out empty values
	filteredParts := []string{}
	for _, part := range parts {
		if part != "" {
			filteredParts = append(filteredParts, part)
		}
	}

	// Join the non-empty parts with a comma
	return strings.Join(filteredParts, ", ")
}

func FormatNonVowelWords(text string) string {
	vowels := "aeiouAEIOU"
	words := strings.Fields(text)
	for i, word := range words {
		containsVowel := false
		for _, char := range word {
			if strings.ContainsRune(vowels, char) {
				containsVowel = true
				break
			}
		}
		if !containsVowel {
			words[i] = strings.ToUpper(word) // Capitalize the word if no vowels are found.
		}
	}
	return strings.Join(words, " ")
}

func VerifyGlidAllAttr(env, glid, attrID, attrVal, dispo string) error {

	jsonStr, err := servapi.UserVerification(env, glid, attrID, attrVal, dispo, "1")
	// fmt.Println(jsonStr,"Inside all")
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return err
	}

	res := make(map[string]interface{})
	for k, v := range data {
		res[strings.ToLower(k)] = v
	}

	status, _ := res["status"].(string)
	// fmt.Println(status,"Inside all attributes")
	if strings.ToLower(status) == "successful" {
		return nil
	}

	return errors.New(jsonStr)
}
