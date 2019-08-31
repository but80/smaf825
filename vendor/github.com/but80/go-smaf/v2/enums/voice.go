package enums

import (
	"encoding/json"
	"fmt"
)

// VoiceType は、音色の種類を表す列挙子です。
type VoiceType int

const (
	VoiceTypeFM VoiceType = iota
	VoiceTypePCM
	VoiceTypeAL
)

func (t VoiceType) String() string {
	var s string
	switch t {
	case 0:
		s = "FM"
	case 1:
		s = "PCM"
	case 2:
		s = "AL"
	default:
		s = fmt.Sprintf("undefined (0x%02X)", int(t))
	}
	return s
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
	}
	return 4
}

type BasicOctave int

const (
	BasicOctaveNormal BasicOctave = 1
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
	PanpotCenter Panpot = 15
)

func (p Panpot) String() string {
	v := int(p)
	if v == 15 {
		return "C"
	}
	if 0 <= v && v < 15 {
		return fmt.Sprintf("L%d", 15-v)
	}
	if 15 < v && v < 32 {
		return fmt.Sprintf("R%d", v-15)
	}
	return "undefined"
}

type Multiplier int

func (m Multiplier) String() string {
	if m == 0 {
		return "1/2"
	}
	return fmt.Sprintf("%d", m)
}
