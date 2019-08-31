package subtypes

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/but80/go-smaf/v2/enums"
	"github.com/but80/go-smaf/v2/internal/util"
	"github.com/but80/go-smaf/v2/log"
	"github.com/but80/go-smaf/v2/voice"
	"github.com/pkg/errors"
)

type Exclusive struct {
	variableLength bool
	Type           enums.ExclusiveType `json:"type"`
	VoiceType      enums.VoiceType     `json:"voice_type,omitempty"`
	VMAVoicePC     *voice.VMAVoicePC   `json:"vma_voice_pc,omitempty"`
	VM35VoicePC    *voice.VM35VoicePC  `json:"vm35_voice_pc,omitempty"`
	Data           []uint8             `json:"data"`
}

// NewExclusive は、新しい Exclusive を作成します。
func NewExclusive(variableLength bool) *Exclusive {
	return &Exclusive{
		variableLength: variableLength,
		Type:           enums.ExclusiveTypeUnknown,
		Data:           []uint8{},
	}
}

func (x *Exclusive) String() string {
	result := fmt.Sprintf("Exclusive %s (%d bytes)", util.Hex(x.Data), len(x.Data))
	sub := []string{}
	if x.VM35VoicePC != nil {
		sub = append(sub, x.VM35VoicePC.String())
	}
	if x.VMAVoicePC != nil {
		sub = append(sub, x.VMAVoicePC.String())
	}
	if len(sub) == 0 {
		return result
	}
	return result + "\n" + util.Indent(strings.Join(sub, "\n"), "\t")
}

// Read は、バイト列を読み取ってパースした結果をこの構造体に格納します。
func (x *Exclusive) Read(rdr io.Reader, rest *int) error {
	var err error
	var length int
	if x.variableLength {
		length, err = util.ReadVariableInt(false, rdr, rest)
	} else {
		var l uint8
		err = binary.Read(rdr, binary.BigEndian, &l)
		*rest--
		length = int(l)
	}
	if err != nil {
		return errors.WithStack(err)
	}
	length--
	log.Debugf("length = %d", length)
	x.Data = make([]uint8, length)
	n, err := rdr.Read(x.Data)
	if err != nil {
		log.Debugf("Read failed")
		return errors.WithStack(err)
	}
	log.Debugf("Read %d", n)
	*rest -= length
	var end uint8
	err = binary.Read(rdr, binary.BigEndian, &end)
	if err != nil {
		log.Debugf("Read failed")
		return errors.WithStack(err)
	}
	*rest--
	if end != 0xF7 {
		log.Warnf("Invalid end mark: 0x%02X", end)
		x.Data = append(x.Data, end)
	}
	//if 0 < len(x.Data) && x.Data[0] == 0x2C {
	//	x.Data = x.Data[1:]
	//}
	//
	if 10 <= len(x.Data) && x.Data[0] == 0x43 && x.Data[1] == 0x79 && x.Data[2] == 0x07 && x.Data[3] == 0x7F && x.Data[4] == 0x01 {
		x.Type = enums.ExclusiveTypeVM35Voice
		x.VoiceType = enums.VoiceType(x.Data[9])
		if x.VoiceType == enums.VoiceTypeFM {
			v, err := voice.NewVM35FMVoice(x.Data[10:], voice.VM35FMVoiceVersionVM5)
			if err == nil {
				x.VM35VoicePC = &voice.VM35VoicePC{
					Version:  voice.VM35FMVoiceVersionVM5,
					BankMSB:  int(x.Data[5]),
					BankLSB:  int(x.Data[6]),
					PC:       int(x.Data[7]),
					DrumNote: enums.Note(x.Data[8]),
					Voice:    v,
				}
			} else {
				log.Warnf("VM3/VM5 voice exclusive error: %s", err.Error())
			}
		} else {
			log.Warnf("Unsupported voice type: %s", x.VoiceType.String())
		}
	} else if 10 <= len(x.Data) && x.Data[0] == 0x43 && x.Data[1] == 0x79 && x.Data[2] == 0x06 && x.Data[3] == 0x7F && x.Data[4] == 0x01 {
		x.Type = enums.ExclusiveTypeVM35Voice
		x.VoiceType = enums.VoiceType(x.Data[9])
		if x.VoiceType == enums.VoiceTypeFM {
			v, err := voice.NewVM35FMVoice(x.Data[10:], voice.VM35FMVoiceVersionVM3Exclusive)
			if err == nil {
				x.VM35VoicePC = &voice.VM35VoicePC{
					Version:  voice.VM35FMVoiceVersionVM3Exclusive,
					BankMSB:  int(x.Data[5]),
					BankLSB:  int(x.Data[6]),
					PC:       int(x.Data[7]),
					DrumNote: enums.Note(x.Data[8]),
					Voice:    v,
				}
			} else {
				log.Warnf("VM3/VM5 voice exclusive error: %s", err.Error())
			}
		} else {
			log.Warnf("Unsupported voice type: %s", x.VoiceType.String())
		}
	} else if 3 <= len(x.Data) && x.Data[0] == 0x43 && x.Data[1] == 0x05 && x.Data[2] == 0x01 {
		x.Type = enums.ExclusiveTypeVM35Voice
		x.VoiceType = enums.VoiceTypeFM
		v, err := voice.NewVM35FMVoice(x.Data[5:], voice.VM35FMVoiceVersionVM5)
		if err == nil {
			x.VM35VoicePC = &voice.VM35VoicePC{
				Version:  voice.VM35FMVoiceVersionVM5,
				BankMSB:  0,
				BankLSB:  int(x.Data[3]),
				PC:       int(x.Data[4]),
				DrumNote: 0,
				Voice:    v,
			}
		} else {
			log.Warnf("VM3/VM5 voice exclusive error: %s", err.Error())
		}
	} else if 6 <= len(x.Data) && x.Data[0] == 0x43 && x.Data[1] == 0x03 {
		x.Type = enums.ExclusiveTypeVMAVoice
		x.VoiceType = enums.VoiceTypeFM
		v, err := voice.NewVMAFMVoice(x.Data[5:])
		if err == nil {
			x.VMAVoicePC = &voice.VMAVoicePC{
				Bank:  int(x.Data[3]),
				PC:    int(x.Data[4]),
				Voice: v,
			}
		} else {
			log.Warnf("VMA voice exclusive error: %s: %s", err.Error(), util.Hex(x.Data))
		}
	} else {
		log.Warnf("Unsupported exclusive type: %s", util.Hex(x.Data))
	}
	return nil
}
