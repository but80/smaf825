package enums

import (
	"encoding/json"
	"fmt"
)

type ChannelType int

const (
	ChannelType_NoCare = iota
	ChannelType_Melody
	ChannelType_NoMelody
	ChannelType_Rhythm
)

func (t ChannelType) String() string {
	s := "undefined"
	switch t {
	case ChannelType_NoCare:
		s = "NoCare"
	case ChannelType_Melody:
		s = "Melody"
	case ChannelType_NoMelody:
		s = "NoMelody"
	case ChannelType_Rhythm:
		s = "Rhythm"
	}
	return fmt.Sprintf("%s(0x%02X)", s, int(t))
}

func (t ChannelType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}
