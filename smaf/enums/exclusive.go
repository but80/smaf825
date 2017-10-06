package enums

import (
	"encoding/json"
	"fmt"
)

type ExclusiveType int

const (
	ExclusiveType_Unknown ExclusiveType = iota
	ExclusiveType_VMAVoice
	ExclusiveType_VM5Voice
)

func (t ExclusiveType) String() string {
	s := "Unknown"
	switch t {
	case ExclusiveType_VMAVoice:
		s = "VMAVoice"
	case ExclusiveType_VM5Voice:
		s = "VM5Voice"
	}
	return fmt.Sprintf("%s(0x%02X)", s, int(t))
}

func (t ExclusiveType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}
