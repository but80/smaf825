package enums

import (
	"encoding/json"
	"fmt"
)

type KeyControlStatus int

const (
	KeyControlStatus_NonSpecified = iota
	KeyControlStatus_Off
	KeyControlStatus_On
)

func (t KeyControlStatus) String() string {
	s := "undefined"
	switch t {
	case KeyControlStatus_NonSpecified:
		s = "NonSpecified"
	case KeyControlStatus_Off:
		s = "Off"
	case KeyControlStatus_On:
		s = "On"
	}
	return fmt.Sprintf("%s(0x%02X)", s, int(t))
}

func (t KeyControlStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}
