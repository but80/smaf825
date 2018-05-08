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
	var s string
	switch t {
	case KeyControlStatus_NonSpecified:
		s = "NonSpecified"
	case KeyControlStatus_Off:
		s = "Off"
	case KeyControlStatus_On:
		s = "On"
	default:
		s = fmt.Sprintf("undefined (0x%02X)", int(t))
	}
	return s
}

func (t KeyControlStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}
