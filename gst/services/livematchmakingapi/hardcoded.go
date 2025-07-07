package livematchmakingapi

import (
	"github.com/rs/xid"
)

// HardcodedResponse returns a static response when running in DEV environment.
// It matches glid, idType ("gst","pan","udyam","iec"), and idValue.
func HardcodedResponse(glid, idType, idValue string) (map[string]interface{}, bool) {
	key := glid + "|" + idType + "|" + idValue
	staticResponses := map[string]map[string]interface{}{
		// GST-based lookup (glid|gst|33DZEPK8089R1Z5)
		"5688597|gst|33DZEPK8089R1Z5": responseFlag2(),
		// PAN-based lookup (glid|pan|DZEPK8089R)
		"5688597|pan|DZEPK8089R": responsePanUdyamIec(),
		// Udyam-based lookup (glid|udyam|DZEPK8089R)
		"5688597|udyam|DZEPK8089R": responseUdyamOnly(),
		// IEC-based lookup (glid|iec|DZEPK8089R)
		"5688597|iec|DZEPK8089R": responseIecOnly(),
	}
	resp, ok := staticResponses[key]
	return resp, ok
}

// responseFlag2 returns verifications for GST, PAN, Udyam, and IEC (used when GST is provided).
func responseFlag2() map[string]interface{} {
	return map[string]interface{}{
		"verifications": map[string]interface{}{
			"gst": map[string]interface{}{
				"status_code":         200,
				"status_message":      "SUCCESS",
				"gstin_number":        "33DZEPK8089R1Z5",
				"reason":              37,
				"reason_desc":         "TRADE NAME + CITY/ZIP MATCHING",
				"bucket_name":         "AA1",
				"flag":                2,
				"verification_status": "AUTO APPROVED",
				"attribute_src":       map[string]interface{}{"121": "9826805553"},
				"gstdetails": map[string]string{
					"annual_turnover_slab":          "NA",
					"business_activity_nature":      "Supplier of Services,Works Contract",
					"business_constitution":         "Proprietorship",
					"core_business_activity_nature": "",
					"proprieter_name":               "",
					"registration_date":             "01-10-2024",
				},
			},
			"pan": map[string]interface{}{
				"status_code":         200,
				"status_message":      "SUCCESS",
				"verification_status": "AUTO APPROVED",
				"attribute_src":       map[string]string{"107": "shivam@gmail.com", "121": "9826805553"},
				"pan_number":          "DZEPK8089R",
			},
			"udyam": map[string]interface{}{
				"status_code":         200,
				"status_message":      "SUCCESS",
				"verification_status": "AUTO APPROVED",
				"attribute_src":       map[string]string{"121": "9826805553"},
				"udyam_number":        "DZEPK8089R",
			},
			"iec": map[string]interface{}{
				"status_code":         200,
				"status_message":      "SUCCESS",
				"verification_status": "AUTO APPROVED",
				"attribute_src":       map[string]string{"121": "9826805553"},
				"iec_number":          "DZEPK8089R",
			},
		},
		"uniq_id": xid.New().String(),
	}
}

// responsePanUdyamIec returns verifications for PAN, Udyam, and IEC (used when PAN is provided).
func responsePanUdyamIec() map[string]interface{} {
	return map[string]interface{}{
		"verifications": map[string]interface{}{
			"pan": map[string]interface{}{
				"status_code":         200,
				"status_message":      "SUCCESS",
				"verification_status": "AUTO APPROVED",
				"attribute_src":       map[string]string{"107": "shivam@gmail.com", "121": "9826805553"},
				"pan_number":          "DZEPK8089R",
			},
			"udyam": map[string]interface{}{
				"status_code":         200,
				"status_message":      "SUCCESS",
				"verification_status": "AUTO APPROVED",
				"attribute_src":       map[string]string{"121": "9826805553"},
				"udyam_number":        "DZEPK8089R",
			},
			"iec": map[string]interface{}{
				"status_code":         200,
				"status_message":      "SUCCESS",
				"verification_status": "AUTO APPROVED",
				"attribute_src":       map[string]string{"121": "9826805553"},
				"iec_number":          "DZEPK8089R",
			},
		},
		"uniq_id": xid.New().String(),
	}
}

// responseUdyamOnly returns only Udyam verification (used when Udyam is provided).
func responseUdyamOnly() map[string]interface{} {
	return map[string]interface{}{
		"verifications": map[string]interface{}{
			"udyam": map[string]interface{}{
				"status_code":         200,
				"status_message":      "SUCCESS",
				"verification_status": "AUTO APPROVED",
				"attribute_src":       map[string]string{"121": "9826805553"},
				"udyam_number":        "DZEPK8089R",
			},
		},
		"uniq_id": xid.New().String(),
	}
}

// responseIecOnly returns only IEC verification (used when IEC is provided).
func responseIecOnly() map[string]interface{} {
	return map[string]interface{}{
		"verifications": map[string]interface{}{
			"iec": map[string]interface{}{
				"status_code":         200,
				"status_message":      "SUCCESS",
				"verification_status": "AUTO APPROVED",
				"attribute_src":       map[string]string{"121": "9826805553"},
				"iec_number":          "DZEPK8089R",
			},
		},
		"uniq_id": xid.New().String(),
	}
}
