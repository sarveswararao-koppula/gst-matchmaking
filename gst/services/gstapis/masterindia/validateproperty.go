package masterindia

import (
	"errors"
)

type property struct {
	validaionKey string
	allowedAPIS  map[string]string
}

var properties map[string]property = map[string]property{
	"bi": property{
		validaionKey: "Ymlfc2NyZWVu",
		allowedAPIS: map[string]string{
			"masterindia": "amrita@indiamart.com",
			"challan":     "sachin63596@indiamart.com",
		},
	},
	"soa": property{
		validaionKey: "c29hX3NjcmVlbg==",
		allowedAPIS: map[string]string{
			"masterindia": "kumar.rahul2@indiamart.com",
		},
	},
	"weberp": property{
		validaionKey: "d2ViZXJwX3NjcmVlbg==",
		allowedAPIS: map[string]string{
			"masterindia": "vivek.arya@indiamart.com",
		},
	},
	"weberp2": property{
		validaionKey: "d2ViZXJwX3NjcmVlbg==",
		allowedAPIS: map[string]string{
			"masterindia": "credentials7@indiamart.com",
		},
	},
	"buyermy": property{
                validaionKey: "e2WjAYKxY4OkdnWmch==",
                allowedAPIS: map[string]string{
                        "masterindia": "credentials8@indiamart.com",
                },
        },
	"merpinstant": property{
		validaionKey: "bWVycF9zY3JlZW4=",
		allowedAPIS: map[string]string{
			"challan":     "amisha",
		},
	},
	"bi3": property{
                validaionKey: "Ymlfc2NyZWVu",
                allowedAPIS: map[string]string{
                        "masterindia": "Credentials3@indiamart.com",
                },
        },
	"bi4": property{
                validaionKey: "Ymlfc2NyZWVu",
                allowedAPIS: map[string]string{
                        "masterindia": "Credentials4@indiamart.com",
                },
        },
		"PAY": property{
			validaionKey: "d2WjAYKxY4OkdnWmch==",
			allowedAPIS: map[string]string{
				"masterindia": "Credentials4@indiamart.com",
			},
		},
		"GlAdmin": property{
			validaionKey: "Z2xhZG1pbl9zY3JlZW4=",
			allowedAPIS: map[string]string{
					"masterindia": "Credentials15@indiamart.com",
			},
		},
		"loans": property{
			validaionKey: "d2WjAYKxY4OkdnWmch==",
			allowedAPIS: map[string]string{
				"masterindia": "Credentials4@indiamart.com",
			},
		},
}

//ValidateProp ...
func ValidateProp(modid string, validationkey string, api string) (string, error) {

	if properties[modid].validaionKey != validationkey || validationkey == "" || modid == "" || api == "" {
		return "", errors.New("Not Authorized")
	}

	for k, v := range properties[modid].allowedAPIS {
		if k == api {
			return v, nil
		}
	}

	return "", errors.New("Not Authorized")
}

