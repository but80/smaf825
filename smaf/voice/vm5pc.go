package voice

import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"

	"github.com/mersenne-sister/smaf825/smaf/enums"
	"github.com/mersenne-sister/smaf825/smaf/util"
	"github.com/pkg/errors"
)

type VM5Voice interface {
	fmt.Stringer
	Read(rdr io.Reader, rest *int) error
	ReadUnusedRest(rdr io.Reader, rest *int) error
}

type VM5VoicePC struct {
	Name      string          `json:"name"`
	Flag      int             `json:"-"` // = 0x24
	BankMSB   int             `json:"bank_msb"`
	BankLSB   int             `json:"bank_lsb"`
	PC        int             `json:"pc"`
	DrumNote  enums.Note      `json:"drum_note"`
	Enigma1   int             `json:"-"`
	VoiceType enums.VoiceType `json:"voice_type"`
	Voice     VM5Voice        `json:"voice"`
}

type vm5VoicePCHeaderRawData struct {
	Enigma1   uint16
	Name      [16]byte
	Flag      uint8
	BankMSB   uint8
	BankLSB   uint8
	PC        uint8
	DrumNote  uint8
	VoiceType uint8 // bit0: 0=Type=FM 1=Type=PCM
}

func (p *VM5VoicePC) Read(rdr io.Reader, rest *int) error {
	var data vm5VoicePCHeaderRawData
	err := binary.Read(rdr, binary.BigEndian, &data)
	if err != nil {
		return errors.WithStack(err)
	}
	*rest -= int(unsafe.Sizeof(data))
	p.Name = util.ZeroPadSliceToString(data.Name[:])
	p.Flag = int(data.Flag)
	p.BankMSB = int(data.BankMSB)
	p.BankLSB = int(data.BankLSB)
	p.PC = int(data.PC)
	p.DrumNote = enums.Note(data.DrumNote)
	p.Enigma1 = int(data.Enigma1)
	p.VoiceType = enums.VoiceType(data.VoiceType)
	switch p.VoiceType {
	case enums.VoiceType_FM:
		p.Voice = &VM5FMVoice{}
		p.Voice.Read(rdr, rest)
		p.Voice.ReadUnusedRest(rdr, rest)
	case enums.VoiceType_PCM:
		p.Voice = &VM5PCMVoice{}
		p.Voice.Read(rdr, rest)
	//case enums.VoiceType_AL:
	default:
		return fmt.Errorf("contains unsupported type of voice")
	}
	return nil
}

func (p *VM5VoicePC) IsForDrum() bool {
	return p.DrumNote != 0
}

func (p *VM5VoicePC) String() string {
	// fmt.Sprintf("Flag: 0x%02X", v.Flag)
	// fmt.Sprintf("Enigma1: 0x%04X", v.Enigma1)
	// fmt.Sprintf("Enigma2: 0x%08X", v.Enigma2)
	s := fmt.Sprintf("Bank %d-%d @%d %s", p.BankMSB, p.BankLSB, p.PC, p.VoiceType.String())
	if p.IsForDrum() {
		s += fmt.Sprintf(" DrumNote=%s\n", p.DrumNote.String())
	}
	s += fmt.Sprintf(": [%s]\n", p.Name)
	return s + util.Indent(p.Voice.String(), "\t")
}
