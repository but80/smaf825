package enums

import (
	"encoding/json"
	"fmt"
)

type ExclusiveType int

const (
	ExclusiveTypeUnknown ExclusiveType = iota
	ExclusiveTypeVMAVoice
	ExclusiveTypeVM35Voice
)

func (t ExclusiveType) String() string {
	var s string
	switch t {
	case ExclusiveTypeVMAVoice:
		s = "VMAVoice"
	case ExclusiveTypeVM35Voice:
		s = "VM35Voice"
	default:
		s = fmt.Sprintf("undefined (0x%02X)", int(t))
	}
	return s
}

func (t ExclusiveType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}
