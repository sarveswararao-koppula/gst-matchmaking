package constants

//Property ...
type Property struct {
        ValidaionKey string
        AllowedAPIS  map[string]string
}

//Propertytan ...
type Propertytan struct {
        ValidaionKey string
}

//Propertiestan ...
var Propertiestan map[string]Propertytan = map[string]Propertytan{
        "weberp": Propertytan{
                ValidaionKey: "bWVycF9zY3JlWER=", //bWVycF9zY3JlZW4=
        },
        "soa": Propertytan{
                ValidaionKey: "bWVycF9zY3JlWER=", //bWVycF9zY3JlZW4=
        },
}

//Properties ...
var Properties map[string]Property = map[string]Property{
        "merp": Property{
                ValidaionKey: "bWVycF9zY3JlZW4=",
                AllowedAPIS: map[string]string{
                        "masterindia": "puneetkochale",
                        "challan":     "puneetkochale",
                },
        },
        "merpinstant": Property{
                ValidaionKey: "bWVycF9zY3JlZW4=",
                AllowedAPIS: map[string]string{
                        "challan":     "amisha",
                },
        },
        "gladmin": Property{
                ValidaionKey: "Z2xhZG1pbl9zY3JlZW4=",
                AllowedAPIS: map[string]string{
                        "masterindia": "Credentials5",
                        "challan":     "puneetsingh",
                        "befisc":      "Gladminbefisc",
                },
        },
        "seller": Property{
                ValidaionKey: "c2VsbGVyX3NjcmVlbg==",
                AllowedAPIS: map[string]string{
                        "masterindia": "credentials6",
                },
        },
        "weberp": Property{
                ValidaionKey: "d2ViZXJwX3NjcmVlbg==",
                AllowedAPIS: map[string]string{
                        "masterindia": "vivek",
                        "gstchallanscreen" : "Nohit",
                },
        },
        "weberp2": Property{
                ValidaionKey: "d2ViZXJwX3NjcmVlbg==",
                AllowedAPIS: map[string]string{
                        "masterindia": "credentials7",
                },
        },
        "buyermy": Property{
                ValidaionKey: "e2WjAYKxY4OkdnWmch==",
                AllowedAPIS: map[string]string{
                        "masterindia": "credentials8",
                },
        },
        "PAY": Property{
		ValidaionKey: "d2WjAYKxY4OkdnWmch==",
		AllowedAPIS: map[string]string{
			"masterindia": "Credentials4",
		},
	},
        "GlAdmin": Property{
                ValidaionKey: "Z2xhZG1pbl9zY3JlZW4=",
                AllowedAPIS: map[string]string{
                        "masterindia": "Credentials15",
                },
        },
        "loans": Property{
		ValidaionKey: "d2WjAYKxY4OkdnWmch==",
		AllowedAPIS: map[string]string{
			"masterindia": "Credentials4",
		},
	},
        "merpcsd": Property{
                ValidaionKey: "bWVycF9zY3JlZW4=",
                AllowedAPIS: map[string]string{
                        "masterindia": "puneetkochale",
                        "challan":     "puneetkochale",
                },
        },
        "merpnsd": Property{
                ValidaionKey: "bWVycF9zY3JlZW4=",
                AllowedAPIS: map[string]string{
                        "masterindia": "puneetkochale",
                        "challan":     "puneetkochale",
                },
        },
        "gst_otp": Property{
                ValidaionKey: "d2ViZXJwX3NjcmVlbg==",
                AllowedAPIS: map[string]string{
                        "masterindia": "vivek",
                },
        },
}

