package subtypes

import (
	"encoding/binary"
	"io"

	"fmt"

	"github.com/mersenne-sister/smaf825/smaf/enums"
	"github.com/mersenne-sister/smaf825/smaf/util"
	"github.com/mersenne-sister/smaf825/smaf/voice"
	"github.com/pkg/errors"
)

type Exclusive struct {
	variableLength bool
	Data           []uint8 `json:"data"`
}

func NewExclusive(variableLength bool) *Exclusive {
	return &Exclusive{
		variableLength: variableLength,
		Data:           []uint8{},
	}
}

func (x *Exclusive) String() string {
	s := fmt.Sprintf("Exclusive %s (%d bytes)", util.Hex(x.Data), len(x.Data))
	//if x.IsVM5Voice() {
	//	v, err := x.ToVM5Voice()
	//	if err == nil {
	//		s = "Exclusive " + v.String()
	//	}
	//}
	return s
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
	return nil
}

func (x *Exclusive) IsVM5FMVoice() bool {
	return 5 <= len(x.Data) && x.Data[0] == 0x43 && x.Data[1] == 0x79 && x.Data[2] == 0x07 && x.Data[3] == 0x7F && x.Data[4] == 0x01 && x.Data[9] == 0x00
}

func (x *Exclusive) ToVM5FMVoice() (*voice.VM5VoicePC, error) {
	pc := &voice.VM5VoicePC{
		BankMSB:  int(x.Data[5]),
		BankLSB:  int(x.Data[6]),
		PC:       int(x.Data[7]),
		DrumNote: enums.Note(x.Data[8]),
	}
	v, err := voice.NewVM5FMVoice(x.Data[10:])
	if err != nil {
		return nil, errors.WithStack(err)
	}
	pc.Voice = v
	return pc, nil
}

func (x *Exclusive) IsVMAVoice() bool {
	return 3 <= len(x.Data) && x.Data[0] == 0x43 && x.Data[1] == 0x03
}

func (x *Exclusive) ToVMAVoice() (*voice.VMAVoicePC, error) {
	pc := &voice.VMAVoicePC{
		Bank: int(x.Data[3]),
		PC:   int(x.Data[4]),
	}
	v, err := voice.NewVMAFMVoice(x.Data[5:])
	if err != nil {
		return nil, errors.WithStack(err)
	}
	pc.Voice = v
	return pc, nil
}
