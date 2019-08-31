package enums

import (
	"encoding/json"
	"fmt"
)

type ChannelType int

const (
	ChannelTypeNoCare = iota
	ChannelTypeMelody
	ChannelTypeNoMelody
	ChannelTypeRhythm
)

func (t ChannelType) String() string {
	var s string
	switch t {
	case ChannelTypeNoCare:
		s = "NoCare"
	case ChannelTypeMelody:
		s = "Melody"
	case ChannelTypeNoMelody:
		s = "NoMelody"
	case ChannelTypeRhythm:
		s = "Rhythm"
	default:
		s = fmt.Sprintf("undefined (0x%02X)", int(t))
	}
	return s
}

func (t ChannelType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}
