package gstmmcontrols

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var states map[string][]string = map[string][]string{
	"delhi":                        []string{"delhi"},
	"pondicherry":                  []string{"pondicherry", "puducherry"},
	"andaman":                      []string{"andaman", "nicobar"},
	"jammu":                        []string{"jammu", "kashmir"},
	"dadra nagar haveli Daman diu": []string{"dadra and nagar haveli and daman and diu", "dadra & nagar haveli & daman and diu"},
}

// IsSateSame ...
func IsSateSame(glState, gstState string) bool {

	arr := [2]string{glState, gstState}

	for i, v := range arr {
		arr[i] = strings.Join(strings.Fields(strings.ToLower(v)), " ")
	}

	if arr[0] == arr[1] {
		return true
	}

	for i, v := range arr {
		done := false

		for key, aliases := range states {
			for _, alias := range aliases {
				if strings.Contains(v, alias) {
					arr[i] = key
					done = true
					break
				}
			}

			if done {
				break
			}
		}
	}

	return arr[0] == arr[1]
}

// Function to check if the pincodes are the same and are exactly 6 characters long
func IsPincodeSame(glidPincode, gstPincode string) bool {
	// Trim spaces
	glidPincode = strings.TrimSpace(glidPincode)
	gstPincode = strings.TrimSpace(gstPincode)

	// Ensure both pincodes are exactly 6 characters long
	if len(glidPincode) != 6 || len(gstPincode) != 6 {
		return false
	}

	// Compare the pincodes for equality
	return glidPincode == gstPincode
}

func eachWordAtleast2charAfterClean(str string) bool {

	strArr := strings.Fields(str)

	//fmt.Println(strArr)
	if len(strArr) > 2 {
		return false
	}

	reg, err := regexp.Compile("[^a-zA-Z]+")
	if err != nil {
		fmt.Println(err)
		return false
	}

	//fmt.Println(strArr)
	for i, v := range strArr {
		strArr[i] = reg.ReplaceAllString(v, "")
	}

	//fmt.Println(strArr)

	for _, v := range strArr {
		if len(v) < 2 {
			return false
		}
	}
	return true
}

func findOwnerNameScore(glid string, gst string) float64 {

	glidArr := strings.Fields(glid)
	gstArr := strings.Fields(gst)
	if len(glidArr) == 0 || len(gstArr) == 0 {
		return 0
	}

	den := len(glidArr)
	if len(glidArr) == 1 {
		den = len(gstArr)
	}

	cnt := 0
	mp := make(map[string]int)

	for _, v := range gstArr {
		mp[v]++
	}

	for _, v := range glidArr {
		if mp[v] > 0 {
			mp[v]--
			cnt++
		}
	}
	result := float64(cnt) / float64(den)
	return float64(int(result*100)) / 100
}

func findAddrScore(glid string, gst string) float64 {

	glidArr := strings.Fields(glid)
	gstArr := strings.Fields(gst)

	if len(glidArr) == 0 || len(gstArr) == 0 {
		return 0
	}

	cnt := 0
	mp := make(map[string]bool)

	for _, v := range gstArr {
		mp[v] = true
	}
	for _, v := range glidArr {
		if mp[v] {
			cnt++
		}
	}
	result := float64(cnt) / float64(len(glidArr))
	return float64(int(result*100)) / 100
}

func remDuplicates(str string) string {

	arr := strings.Fields(str)
	mp := make(map[string]bool)
	res := []string{}
	for _, v := range arr {
		if !mp[v] {
			mp[v] = true
			res = append(res, v)
		}
	}
	return strings.Join(res, " ")
}

func remWords(addr string, rem ...string) string {

	mp := make(map[string]bool)

	for _, v := range rem {
		arr := strings.Fields(v)
		for _, v1 := range arr {
			mp[v1] = true
		}
	}

	res := []string{}
	for _, v := range strings.Fields(addr) {
		if !mp[v] {
			res = append(res, v)
		}
	}

	return strings.Join(res, " ")
}

func cleanStr(str string) string {
	return strings.Join(strings.Fields(strings.ToLower(str)), " ")
}

func prefixMatchCnt(gl, gst string) int {

	l := len(gl)

	if len(gst) < l {
		l = len(gst)
	}

	cnt := 0
	for i := 0; i < l; i++ {
		if gl[i] == gst[i] {
			cnt++
		} else {
			break
		}
	}
	return cnt
}

func gstCaseME(gst string) bool {

	if len(gst) < 6 {
		return false
	}

	gst = strings.ToUpper(gst)

	if gst[5] == 'C' || gst[5] == 'F' {
		return true
	}

	return false
}

func stringMatchScore(glid, gst []string) float64 {

	if len(glid) == 0 {
		return 0
	}

	cnt := 0
	hash := make(map[string]bool)

	for _, e := range gst {
		hash[e] = true
	}
	for _, e := range glid {
		if hash[e] {
			cnt++
		}
	}

	result := float64(cnt) / float64(len(glid))

	return float64(int(result*100)) / 100

}

func stringLength(str string) int {
	return len(strings.Fields(str))
}

func tradeBizSame(tradeName, ownerName string) bool {

	tradeNameArr := strings.Fields(strings.ToLower(tradeName))
	ownerNameArr := strings.Fields(strings.ToLower(ownerName))

	if len(tradeNameArr) != len(ownerNameArr) {
		return false
	}

	for i, v := range tradeNameArr {
		if v != ownerNameArr[i] {
			return false
		}
	}

	return true
}

func uniqStringLen(str string) int {

	arr := strings.Fields(strings.ToLower(str))
	mp := make(map[string]bool)

	for _, v := range arr {
		mp[v] = true
	}
	return len(mp)
}

func cleanOwnerName(ownerName string) string {

	reg, err := regexp.Compile("[^a-zA-Z]+")
	if err != nil {
		fmt.Println(err)
		return ""
	}

	//MUST BE LOWER CASE
	result := strings.ToLower(ownerName)

	result = reg.ReplaceAllString(result, " ")

	fields := strings.Fields(result)

	//removing non - alpha numeric
	for i, v := range fields {
		fields[i] = reg.ReplaceAllString(v, "")
	}

	fields = splitWordOwnerName(fields)

	fields = removeWordsOwnerName(fields)

	fields = removeSpaceFromFields(fields)

	if len(fields) > 2 {
		fields = removeLenOwnerName(fields)
	}

	fields = removeSpaceFromFields(fields)

	return strings.Join(fields, " ")
}

// removing "mr", "ms", "smt", "shree", "shri", "dr"
func removeWordsOwnerName(fields []string) []string {

	toBeRemoved := []string{"mr", "ms", "smt", "shree", "shri", "dr", "mrs", "er"}

	fields = removeSpaceFromFields(fields)
	result := []string{}

	for _, v := range fields {

		for _, rem := range toBeRemoved {
			if rem == v {
				v = ""
				break
			}
		}
		if v != "" {
			result = append(result, v)
		}
	}

	return result
}

// spliting words eg shivkumar to shiv kumar
func splitWordOwnerName(fields []string) []string {

	fields = removeSpaceFromFields(fields)

	words := []string{"kumar"}
	result := []string{}

	for _, v := range fields {

		for _, word := range words {
			if strings.HasSuffix(v, word) {
				v = v[:len(v)-len(word)] + " " + word
				break
			}
		}

		result = append(result, v)
	}

	return result
}

// removing 1 longth word
func removeLenOwnerName(fields []string) []string {

	fields = removeSpaceFromFields(fields)
	result := []string{}

	for _, v := range fields {

		if len(v) == 1 {
			v = ""
		}

		if v != "" {
			result = append(result, v)
		}
	}
	return result
}

func removeSpaceFromFields(fields []string) []string {
	return strings.Fields(strings.Join(fields, " "))
}

//Stage 1 Functions

func replaceCity(str string) string {

	str = strings.Trim(strings.ToLower(str), " ")

	values := []string{"new delhi", "delhi", "bangalore", "bangaluru"}

	for i := 0; i < len(values); i = i + 2 {
		str = strings.ReplaceAll(str, values[i], values[i+1])
	}

	return str
}

func removeMrMrs(name string) string {

	name = strings.ToLower(name)
	remove := []string{"dr", "er", "mr", "ms", "smt", "shree", "shri", "mrs"}

	arr := strings.Fields(name)
	for i, v := range arr {
		for _, rem := range remove {
			if v == rem {
				arr[i] = ""
			}
		}
	}

	arr = strings.Fields(strings.Join(arr, " "))

	str := strings.Join(arr, " ")

	return str
}

func stringMatchScoreStage1(glidStr, gstStr string, isOwnerName bool) float64 {

	glid := strings.Fields(strings.ToLower(glidStr))
	gst := strings.Fields(strings.ToLower(gstStr))

	if len(glid) == 0 {
		return 0
	}
	cnt := 0
	hash := make(map[string]bool)

	for _, e := range gst {
		hash[e] = true
	}
	for _, e := range glid {
		if hash[e] {
			cnt++
		}
	}
	result := float64(cnt) / float64(len(glid))

	if isOwnerName && len(glid) == 1 {
		result = float64(cnt) / float64(len(gst))
	}
	return float64(int(result*100)) / 100
}

func excludeFromGlAddr(glAddr string, excludes []string) string {

	glAddr = strings.Trim(strings.ToLower(glAddr), " ")

	for _, v := range excludes {
		v = strings.Trim(strings.ToLower(v), " ")
		glAddr = strings.ReplaceAll(glAddr, v, " ")
	}

	return strings.Join(strings.Fields(glAddr), " ")
}

func getStateScore(gl string, gst string) float64 {

	if strings.Trim(strings.ToLower(gl), " ") == strings.Trim(strings.ToLower(gst), " ") {
		return 1
	}

	return 0
}

func getCityScore(glCity string, gstAddress string) float64 {
	if strings.Contains(gstAddress, glCity) {
		return 1
	}
	return 0
}

func countWordsCompanyName(comanyname string) int {
	return len(strings.Fields(comanyname))
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
		check := []string{"pet", "sky", "art", "of", "and", "sri", "sai", "to", "toy", "raj", "ram", "for", "avi", "uma", "ji", "wen", "hub", "sun", "oil", "air", "bio", "gem", "the", "maa", "you", "eye", "tex", "ka", "new", "sah", "sha", "way", "web", "ads", "tea", "dev", "ply", "new", "go", "fab"}

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

			if !Ch2_Vowel_or_Not || !Ch3_Vowel_or_Not {
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

// CreateLogEntry creates a LogEntry from the provided Logg struct.
func CreateLog(logg Logg) Logkibana {
	return Logkibana{
		RequestStart:       logg.RequestStart,
		ResponseTime_Float: logg.ResponseTime,
		ServiceName:        logg.ServiceName,
		ServiceURL:         CheckAndReturnPath(logg.ServiceURL),
		RemoteAddress:      logg.RemoteAddress,
		Request_Data:       MarshalToString(logg.Request),
		Response_Body:      MarshalToString(logg.Response),
		Any_Error:          MarshalToString(logg.AnyError),
		StackTrace:         logg.StackTrace,
		CustTypeFlag:       logg.CustTypeFlag,
		ContactSource:      logg.ContactSource,
	}
}

func CreateWorkerLog(logg Logg) LogKibanaWorker {
	return LogKibanaWorker{
		LogType:                "Consumer",
		RequestStart:           logg.RequestStart,
		ServiceName:            logg.ServiceName,
		ServiceURL:             CheckAndReturnPath(logg.ServiceURL),
		RemoteAddress:          logg.RemoteAddress,
		Request_Data:           MarshalToString(logg.Request),
		Response_Body:          MarshalToString(logg.Response),
		Any_Error:              MarshalToString(logg.AnyError),
		MasterIndia_Hit_Status: MarshalToString(logg.MasterIndia),
		Gstin:                  logg.Response.Body.Gstin,
		BucketType:             logg.Response.Body.BucketType,
		BucketName:             logg.Response.Body.BucketName,
	}
}

func CheckAndReturnPath(input string) string {
	if strings.Contains(input, "/gstmm/v1/gst") {
		return "/gstmm/v1/gst"
	}
	return input
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
