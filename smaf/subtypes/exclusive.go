package subtypes

import (
	"encoding/binary"
	"io"

	"fmt"

	"strings"

	"github.com/mersenne-sister/smaf825/smaf/enums"
	"github.com/mersenne-sister/smaf825/smaf/util"
	"github.com/mersenne-sister/smaf825/smaf/voice"
	"github.com/pkg/errors"
)

type Exclusive struct {
	variableLength bool
	Type           enums.ExclusiveType `json:"type"`
	VoiceType      enums.VoiceType     `json:"voice_type,omitempty"`
	VMAVoicePC     *voice.VMAVoicePC   `json:"vma_voice_pc,omitempty"`
	VM5VoicePC     *voice.VM5VoicePC   `json:"vm5_voice_pc,omitempty"`
	Data           []uint8             `json:"data"`
}

func NewExclusive(variableLength bool) *Exclusive {
	return &Exclusive{
		variableLength: variableLength,
		Type:           enums.ExclusiveType_Unknown,
		Data:           []uint8{},
	}
}

func (x *Exclusive) String() string {
	result := fmt.Sprintf("Exclusive %s (%d bytes)", util.Hex(x.Data), len(x.Data))
	sub := []string{}
	if x.VM5VoicePC != nil {
		sub = append(sub, x.VM5VoicePC.String())
	}
	if x.VMAVoicePC != nil {
		sub = append(sub, x.VMAVoicePC.String())
	}
	if len(sub) == 0 {
		return result
	}
	return result + "\n" + util.Indent(strings.Join(sub, "\n"), "\t")
}

func (x *Exclusive) Read(rdr io.Reader, rest *int) error {
	var err error
	var length int
	if x.variableLength {
		length, err = util.ReadVariableInt(false, rdr, rest)
	} else {
		var l uint8
		err = binary.Read(rdr, binary.BigEndian, &l)
		length = int(l)
	}
	if err != nil {
		return errors.WithStack(err)
	}
	length--
	*rest--
	x.Data = make([]uint8, length)
	err = binary.Read(rdr, binary.BigEndian, &x.Data)
	if err != nil {
		return errors.WithStack(err)
	}
	*rest -= len(x.Data)
	var end uint8
	err = binary.Read(rdr, binary.BigEndian, &end)
	if err != nil {
		return errors.WithStack(err)
	}
	if end != 0xF7 {
		return errors.Errorf("Invalid end mark: 0x%02X", end)
	}
	*rest--
	//
	if 10 <= len(x.Data) && x.Data[0] == 0x43 && x.Data[1] == 0x79 && x.Data[2] == 0x07 && x.Data[3] == 0x7F && x.Data[4] == 0x01 {
		x.Type = enums.ExclusiveType_VM5Voice
		x.VoiceType = enums.VoiceType(x.Data[9])
		if x.VoiceType == enums.VoiceType_FM {
			v, err := voice.NewVM5FMVoice(x.Data[10:])
			if err == nil {
				x.VM5VoicePC = &voice.VM5VoicePC{
					BankMSB:  int(x.Data[5]),
					BankLSB:  int(x.Data[6]),
					PC:       int(x.Data[7]),
					DrumNote: enums.Note(x.Data[8]),
					Voice:    v,
				}
			}
		}
	} else if 3 <= len(x.Data) && x.Data[0] == 0x43 && x.Data[1] == 0x05 && x.Data[2] == 0x01 {
		x.Type = enums.ExclusiveType_VM5Voice
		x.VoiceType = enums.VoiceType_FM
		v, err := voice.NewVM5FMVoice(x.Data[5:])
		if err == nil {
			x.VM5VoicePC = &voice.VM5VoicePC{
				BankMSB:  0,
				BankLSB:  0,
				PC:       int(x.Data[4]),
				DrumNote: 0,
				Voice:    v,
			}
		}
	} else if 3 <= len(x.Data) && x.Data[0] == 0x43 && x.Data[1] == 0x03 {
		x.Type = enums.ExclusiveType_VMAVoice
		x.VoiceType = enums.VoiceType_FM
		v, err := voice.NewVMAFMVoice(x.Data[5:])
		if err == nil {
			x.VMAVoicePC = &voice.VMAVoicePC{
				Bank:  int(x.Data[3]),
				PC:    int(x.Data[4]),
				Voice: v,
			}
		}
	}
	return nil
}
