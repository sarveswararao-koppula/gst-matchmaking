package masterindia

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var creds map[string]cred = make(map[string]cred)

func init() {

	var Allcreds []cred

	credFile, err := os.Open(credsJSONPath)
	if err != nil {
		panic(err)
	}
	raw, _ := ioutil.ReadAll(credFile)

	err = json.Unmarshal(raw, &Allcreds)
	if err != nil {
		panic(err)
	}

	for _, v := range Allcreds {
		creds[v.UserName] = v
	}

}

