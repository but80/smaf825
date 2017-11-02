package enums

import (
	"encoding/json"
	"fmt"
)

type VoiceType int

const (
	VoiceType_FM VoiceType = iota
	VoiceType_PCM
	VoiceType_AL
)

func (t VoiceType) String() string {
	s := "unknown"
	switch t {
	case 0:
		s = "FM"
	case 1:
		s = "PCM"
	case 2:
		s = "AL"
	}
	return fmt.Sprintf("%s(%d)", s, t)
}

func (t VoiceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

type Algorithm int

func (a Algorithm) String() string {
	s := "unknown"
	switch a {
	case 0:
		s = "FB(1)->2"
	case 1:
		s = "FB(1) + 2"
	case 2:
		s = "FB(1) + 2 + FB(3) + 4"
	case 3:
		s = "(FB(1) + 2->3) -> 4"
	case 4:
		s = "FB(1)->2->3->4"
	case 5:
		s = "FB(1)->2 + FB(3)->4"
	case 6:
		s = "FB(1) + 2->3->4"
	case 7:
		s = "FB(1) + 2->3 + 4"
	}
	return fmt.Sprintf("%d[ %s ]", int(a), s)
}
func (a Algorithm) OperatorCount() int {
	if int(a) < 2 {
		return 2
	} else {
		return 4
	}
}

type BasicOctave int

const (
	BasicOctave_Normal BasicOctave = 1
)

func (o BasicOctave) String() string {
	switch o {
	case 0:
		return "1"
	case 1:
		return "0"
	case 2:
		return "-1"
	case 3:
		return "-2"
	}
	return "undefined"
}

func (o BasicOctave) NoteDiff() Note {
	switch o {
	case 0:
		return Note(1 * 12)
	case 2:
		return Note(-1 * 12)
	case 3:
		return Note(-2 * 12)
	default:
		return Note(0 * 12)
	}
}

type Panpot int

const (
	Panpot_Center Panpot = 15
)

func (p Panpot) String() string {
	v := int(p)
	if v == 15 {
		return "C"
	} else if 0 <= v && v < 15 {
		return fmt.Sprintf("L%d", 15-v)
	} else if 15 < v && v < 32 {
		return fmt.Sprintf("R%d", v-15)
	} else {
		return "undefined"
	}
}

type Multiplier int

func (m Multiplier) String() string {
	if m == 0 {
		return "1/2"
	} else {
		return fmt.Sprintf("%d", m)
	}
}
