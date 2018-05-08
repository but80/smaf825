package enums

import (
	"encoding/json"
	"fmt"
)

type ExclusiveType int

const (
	ExclusiveType_Unknown ExclusiveType = iota
	ExclusiveType_VMAVoice
	ExclusiveType_VM35Voice
)

func (t ExclusiveType) String() string {
	var s string
	switch t {
	case ExclusiveType_VMAVoice:
		s = "VMAVoice"
	case ExclusiveType_VM35Voice:
		s = "VM35Voice"
	default:
		s = fmt.Sprintf("undefined (0x%02X)", int(t))
	}
	return s
}

func (t ExclusiveType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}
