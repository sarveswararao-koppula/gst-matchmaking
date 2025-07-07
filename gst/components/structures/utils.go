package structures

import (
	"time"
)

type Tokken struct {
	Tok string    `json:"Tok,omitempty"`
	Exp time.Time `json:"Exp,omitempty"`
}

