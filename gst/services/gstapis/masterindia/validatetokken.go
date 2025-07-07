package masterindia

import (
	"errors"
	"mm/utils"
	"time"
)

var loc *time.Location = utils.GetLocalTime()

type token struct {
	tok string
	exp time.Time
}

var tokens map[string]*token = make(map[string]*token)

//ValidateTokken ...
func ValidateTokken(APIUserID string) error {

	var generateTokken bool = false

	if tokens[APIUserID] == nil || tokens[APIUserID].exp.IsZero() || int(time.Now().In(loc).Sub(tokens[APIUserID].exp).Hours()) > 4 {
		generateTokken = true
	}

	if !generateTokken {
		return nil
	}

	tok, err := GetTokken(creds[APIUserID])
	if err != nil {
		return err
	}

	accessToken, ok := tok["access_token"].(string)
	if !ok {
		return errors.New("err in generating access_token")
	}

	tokens[APIUserID] = &token{
		tok: accessToken,
		exp: time.Now().In(loc),
	}

	return nil
}
