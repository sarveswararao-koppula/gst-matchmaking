package properties

import (
	"fmt"

	"github.com/magiconair/properties"
)

//properties
type prop struct {
	PORT                    string `properties:"PORT"`
	DATABASE                string `properties:"DATABASE"`
	LOG_MASTERINDIA         string `properties:"LOG_MASTERINDIA"`
	LOG_MATCH_MAKING        string `properties:"LOG_MATCH_MAKING"`
	LOG_MATCH_MAKING_KIBANA string `properties:"LOG_MATCH_MAKING_KIBANA"`
	AWS_SQS_URL             string `properties:"AWS_SQS_URL"`
	LOG_TAN                 string `properties:"LOG_TAN"`
	AWS_SQS_CRED_PATH       string `properties:"AWS_SQS_CRED_PATH"`
	SERVICES_ENV            string `properties:"SERVICES_ENV"`
}

//Prop ...
var Prop prop

func init() {

	f1 := "/go/src/mm/config.properties"

	p, err := properties.LoadFiles([]string{f1}, properties.UTF8, false)
	if err != nil {
		panic(err)
	}

	err = p.Decode(&Prop)
	if err != nil {
		panic(err)
	}

	fmt.Println(Prop)
	fmt.Println("Loaded Properties...")
}

